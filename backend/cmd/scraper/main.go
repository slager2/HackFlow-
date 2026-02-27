package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"hackflow-api/internal/config"
	"hackflow-api/internal/database"
	"hackflow-api/internal/logger"
	"hackflow-api/internal/models"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var db *gorm.DB

// AIResponse - промежуточная структура для строгого JSON от ИИ
type AIResponse struct {
	Title    string  `json:"title"`
	DateStr  string  `json:"date_str"`
	Deadline string  `json:"deadline"`
	Format   string  `json:"format"`
	City     *string `json:"city"`
	AgeLimit string  `json:"ageLimit"`
	Link     *string `json:"link"`
	Status   string  `json:"status"`
}

// ScrapedPost структура для передачи текста и даты публикации
type ScrapedPost struct {
	Text        string
	PublishedAt time.Time
}

func main() {
	cfg := config.Load()
	logger.Setup(cfg.Env)

	slog.Info("Запуск Telegram-парсера HackFlow (Docker Mode)")

	var err error
	db, err = database.Init(cfg)
	if err != nil {
		slog.Error("Ошибка инициализации БД", "error", err)
		os.Exit(1)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		slog.Error("Не найден GEMINI_API_KEY в переменных окружения")
		os.Exit(1)
	}

	// Первый запуск сразу после старта контейнера
	runScraper(apiKey)

	// Запускаем парсер каждые 6 часов
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	slog.Info("Парсер переведен в фоновый режим. Следующий запуск через 6 часов.")
	for range ticker.C {
		runScraper(apiKey)
	}
}

func runScraper(apiKey string) {
	channels := []string{"astanahub", "uppertunity", "nuris_nu", "terriconvalley", "bluescreenkz", "kolesa_team", "tce_kz", "hackathons_ru"}

	twoMonthsAgo := time.Now().AddDate(0, -2, 0)

	for _, channel := range channels {
		slog.Info("Парсинг канала", "channel", channel)
		posts, err := scrapeChannel(channel)
		if err != nil {
			slog.Error("Ошибка парсинга", "channel", channel, "error", err)
			continue
		}

		slog.Info("Найдено потенциальных хакатонов", "count", len(posts), "channel", channel)

		for _, post := range posts {
			// Игнор старья: пропускаем посты старше 2 месяцев
			if post.PublishedAt.Before(twoMonthsAgo) {
				slog.Debug("Пропуск слишком старого поста", "published_at", post.PublishedAt)
				continue
			}

			hackathon := parseWithAI(post, apiKey)
			if hackathon == nil {
				continue
			}

			var existing models.Hackathon
			res := db.Where("title = ?", hackathon.Title).First(&existing)
			if res.Error == nil {
				slog.Info("Хакатон уже существует, пропускаем", "title", hackathon.Title)
				continue
			} else if res.Error != gorm.ErrRecordNotFound {
				slog.Error("Ошибка проверки дубликата", "error", res.Error)
				continue
			}

			if err := db.Create(hackathon).Error; err != nil {
				slog.Error("Ошибка сохранения хакатона", "title", hackathon.Title, "error", err)
			} else {
				slog.Info("✅ Успешно добавлен новый хакатон!", "title", hackathon.Title)
			}

			time.Sleep(3 * time.Second)
		}
	}
	slog.Info("Текущий цикл парсинга завершен!")
}

// scrapeChannel парсит Telegram Web Preview и извлекает тексты постов с датами
func scrapeChannel(channelName string) ([]ScrapedPost, error) {
	url := fmt.Sprintf("https://t.me/s/%s", channelName)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("статус код ошибки: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	var posts []ScrapedPost

	doc.Find(".tgme_widget_message").Each(func(i int, s *goquery.Selection) {
		textSelection := s.Find(".tgme_widget_message_text")
		if textSelection.Length() == 0 {
			return
		}

		text := strings.TrimSpace(textSelection.Text())
		lowerText := strings.ToLower(text)

		// Находим тег <time> с датой
		timeAttr, exists := s.Find("time").Attr("datetime")
		if !exists {
			return
		}

		// Парсим дату стандарта ISO (напр. 2024-02-21T15:04:05+00:00)
		publishedAt, err := time.Parse(time.RFC3339, timeAttr)
		if err != nil {
			return
		}

		// Оставляем только посты с 언급анием хакатонов
		if strings.Contains(lowerText, "хакатон") || strings.Contains(lowerText, "hackathon") {
			posts = append(posts, ScrapedPost{
				Text:        text,
				PublishedAt: publishedAt,
			})
		}
	})

	return posts, nil
}

// parseWithAI использует Gemini для извлечения структурированных данных из текста
func parseWithAI(post ScrapedPost, apiKey string) *models.Hackathon {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		slog.Error("Ошибка инициализации Gemini", "error", err)
		return nil
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash-lite")
	model.SetTemperature(0.1)
	model.ResponseMIMEType = "application/json"

	currentDate := time.Now().Format("2006-01-02")
	postDate := post.PublishedAt.Format("2006-01-02")

	prompt := fmt.Sprintf(`Сегодняшняя дата: %s. Пост был опубликован: %s. 
Проанализируй текст анонса. Если дедлайн регистрации или сам хакатон уже прошли относительно сегодняшней даты, верни статус 'DEAD'. Если он еще предстоит — 'LIVE'. Вычисли точный год, опираясь на дату публикации. 
Верни СТРОГО JSON: title (string), date_str (string, например '21-22 февраля 2024'), deadline (string 'YYYY-MM-DD', если нет - пустая строка), format (ОФЛАЙН/ОНЛАЙН), city (string/null), ageLimit (string), link (string/null), status ('LIVE' или 'DEAD').

Текст анонса:
---
%s
---
Только чистый JSON.`, currentDate, postDate, post.Text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		slog.Error("Ошибка генерации ИИ", "error", err)
		return nil
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		slog.Error("Пустой ответ от Gemini")
		return nil
	}

	jsonText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	jsonText = strings.TrimPrefix(jsonText, "```json\n")
	jsonText = strings.TrimPrefix(jsonText, "```\n")
	jsonText = strings.TrimSuffix(jsonText, "\n```")
	jsonText = strings.TrimSpace(jsonText)

	var aiResp AIResponse
	if err := json.Unmarshal([]byte(jsonText), &aiResp); err != nil {
		slog.Error("Парсинг JSON провален", "error", err, "raw_json", jsonText)
		return nil
	}

	if aiResp.Title == "" || aiResp.Title == "null" {
		slog.Warn("ИИ вернул пустой Title, пропускаем пост")
		return nil
	}

	// Конвертация из AIResponse в models.Hackathon
	hackathon := &models.Hackathon{
		Title:    aiResp.Title,
		Date:     aiResp.DateStr,
		Format:   aiResp.Format,
		AgeLimit: aiResp.AgeLimit,
		Status:   aiResp.Status,
	}

	if aiResp.City != nil && *aiResp.City != "null" {
		hackathon.City = *aiResp.City
	}
	if aiResp.Link != nil && *aiResp.Link != "null" {
		hackathon.Link = *aiResp.Link
	}

	if aiResp.Deadline != "" && aiResp.Deadline != "null" {
		parsedDeadline, err := time.Parse("2006-01-02", aiResp.Deadline)
		if err == nil {
			hackathon.Deadline = &parsedDeadline
		} else {
			slog.Warn("Ошибка парсинга Deadline от ИИ", "deadline", aiResp.Deadline, "error", err)
		}
	}

	return hackathon
}
