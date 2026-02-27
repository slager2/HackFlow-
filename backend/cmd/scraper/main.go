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
	channels := []string{"astanahub", "uppertunity", "nuris_nu", "terriconvalley", "bluescreenkz"}

	for _, channel := range channels {
		slog.Info("Парсинг канала", "channel", channel)
		msgs, err := scrapeChannel(channel)
		if err != nil {
			slog.Error("Ошибка парсинга", "channel", channel, "error", err)
			continue
		}

		slog.Info("Найдено потенциальных хакатонов", "count", len(msgs), "channel", channel)

		for _, msg := range msgs {
			hackathon := parseWithAI(msg, apiKey)
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

// scrapeChannel парсит Telegram Web Preview и извлекает тексты постов
func scrapeChannel(channelName string) ([]string, error) {
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

	var messages []string

	doc.Find(".tgme_widget_message_text").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		lowerText := strings.ToLower(text)

		if strings.Contains(lowerText, "хакатон") || strings.Contains(lowerText, "hackathon") {
			messages = append(messages, text)
		}
	})

	return messages, nil
}

// parseWithAI использует Gemini для извлечения структурированных данных из текста
func parseWithAI(text, apiKey string) *models.Hackathon {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		slog.Error("Ошибка инициализации Gemini", "error", err)
		return nil
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash-lite") // Use latest flash model
	model.SetTemperature(0.1)
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`
	Проанализируй текст анонса IT-мероприятия. 
	Найди в нем детали хакатона и верни СТРОГО валидный JSON-объект без маркдауна (без блоков форматирования markdown) со следующими ключами:
	- "title" (строка, название хакатона)
	- "date" (строка, даты проведения)
	- "format" (строка, строго "ОФЛАЙН" или "ОНЛАЙН", или "ОФЛАЙН/ОНЛАЙН")
	- "city" (строка, город, если не указан - null)
	- "ageLimit" (строка, возрастное ограничение, если нет - "Нет ограничений")
	- "link" (строка, ссылка для регистрации, если нет - null)
	
	Текст анонса:
	---
	%s
	---
	Только чистый JSON.`, text)

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

	var hackathon models.Hackathon
	if err := json.Unmarshal([]byte(jsonText), &hackathon); err != nil {
		slog.Error("Парсинг JSON провален", "error", err, "raw_json", jsonText)
		return nil
	}

	if hackathon.Title == "" || hackathon.Title == "null" {
		slog.Warn("ИИ вернул пустой Title, пропускаем пост")
		return nil
	}

	return &hackathon
}
