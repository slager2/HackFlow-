package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"hackflow-api/internal/config"
	"hackflow-api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// SearchAIHandler содержит зависимости для агента (ключи)
type SearchAIHandler struct {
	Config *config.Config
}

func NewSearchAIHandler(cfg *config.Config) *SearchAIHandler {
	return &SearchAIHandler{
		Config: cfg,
	}
}

// TavilyRequest описывает тело запроса к API Tavily
type TavilyRequest struct {
	APIKey        string `json:"api_key"`
	Query         string `json:"query"`
	SearchDepth   string `json:"search_depth"`
	IncludeAnswer bool   `json:"include_answer"`
	MaxResults    int    `json:"max_results"`
}

// TavilyResponse описывает ответ от API Tavily
type TavilyResponse struct {
	Results []struct {
		Content string `json:"content"`
	} `json:"results"`
}

// AIHackathon — облегченная структура для ответов от ИИ (без time.Time)
type AIHackathon struct {
	Title    string  `json:"title"`
	Date     string  `json:"date"`
	Deadline *string `json:"deadline"`
	Format   string  `json:"format"`
	City     string  `json:"city"`
	AgeLimit string  `json:"ageLimit"`
	Link     *string `json:"link"`
	Status   string  `json:"status"`
}

// SearchAI выполняет поиск в реальном времени через Tavily + Gemini
func (h *SearchAIHandler) SearchAI(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	slog.Info("Starting Web-Browsing RAG Search", "query", query)

	if h.Config.TavilyAPIKey == "" || h.Config.GeminiAPIKey == "" {
		slog.Error("Missing API keys for AI Search")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server misconfigured: missing API keys"})
		return
	}

	// 1. Поиск в интернете через Tavily
	reqBody := TavilyRequest{
		APIKey:        h.Config.TavilyAPIKey,
		Query:         fmt.Sprintf("Hackathons IT events in Kazakhstan %s", query),
		SearchDepth:   "advanced",
		IncludeAnswer: false,
		MaxResults:    5,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		slog.Error("Failed to marshal Tavily request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error formatting search via Tavily"})
		return
	}

	resp, err := http.Post("https://api.tavily.com/search", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		slog.Error("Failed to reach Tavily API", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to search the web"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		slog.Error("Tavily returned error", "status", resp.StatusCode, "body", string(body))
		c.JSON(http.StatusBadGateway, gin.H{"error": "Web search provider returned an error"})
		return
	}

	var tavilyResp TavilyResponse
	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		slog.Error("Failed to decode Tavily response", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing web search results"})
		return
	}

	if len(tavilyResp.Results) == 0 {
		slog.Info("No search results found from Tavily for prompt", "query", query)
		c.JSON(http.StatusOK, []models.Hackathon{})
		return
	}

	var webContextBuilder strings.Builder
	for i, res := range tavilyResp.Results {
		webContextBuilder.WriteString(fmt.Sprintf("\n--- РЕЗУЛЬТАТ %d ---\n%s", i+1, res.Content))
	}
	webContext := webContextBuilder.String()

	// 2. Анализ данных через Gemini 2.5 Flash
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(h.Config.GeminiAPIKey))
	if err != nil {
		slog.Error("Failed to initialize Gemini client", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize AI processor"})
		return
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.5-flash")
	model.SetTemperature(0.2)
	model.ResponseMIMEType = "application/json"

	currentDate := time.Now().Format("2006-01-02")
	prompt := fmt.Sprintf(`Сегодняшняя дата: %s.
Пользователь ищет: '%s'.
Вот сырые тексты из интернета (они могут быть на английском или русском):
%s

Твоя задача — извлечь IT-мероприятия. ВАЖНЫЕ ПРАВИЛА ДЛЯ КАЗАХСТАНА:
- Если ивент имеет статус 'National' (Национальный) или проходит в '20+ cities' (например, Decentrathon), АВТОМАТИЧЕСКИ считай, что он проходит в Астане и Алматы. Обязательно добавляй его в ответ!
- Переводи названия городов на русский (Astana -> Астана, Almaty -> Алматы).
- Если точных дат нет, пиши 'Даты уточняются'.
- Если текст на английском, переведи суть и верни JSON на русском.
- Если ничего не найдено, верни пустой массив [].

Верни массив JSON. Структура одного объекта:
- title (строка, на русском)
- date (строка, на русском)
- deadline (строка формата YYYY-MM-DD или null)
- format (строка: строго ОФЛАЙН или ОНЛАЙН, или ОФЛАЙН/ОНЛАЙН)
- city (строка на русском или null)
- ageLimit (строка, например "Нет ограничений")
- link (строка URL или null)
- status (строка: LIVE если дедлайн не прошел относительно сегодняшней даты, иначе DEAD)

Только чистый JSON массив.`, currentDate, query, webContext)

	slog.Debug("Sending aggregated results to Gemini...")
	aiResp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		slog.Error("Failed to generate content via Gemini", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process search results via AI"})
		return
	}

	if len(aiResp.Candidates) == 0 || len(aiResp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusOK, []models.Hackathon{})
		return
	}

	rawString := fmt.Sprintf("%v", aiResp.Candidates[0].Content.Parts[0])
	slog.Info("Raw Gemini Response", "data", rawString)

	// Очистка от Markdown
	jsonText := strings.TrimSpace(rawString)
	jsonText = strings.ReplaceAll(jsonText, "```json", "")
	jsonText = strings.ReplaceAll(jsonText, "```", "")
	jsonText = strings.TrimSpace(jsonText)

	// Дополнительная перестраховка: берем только содержимое между [ и ]
	if start := strings.Index(jsonText, "["); start != -1 {
		if end := strings.LastIndex(jsonText, "]"); end != -1 && end >= start {
			jsonText = jsonText[start : end+1]
		}
	}

	slog.Debug("Tavily context sent to Gemini", "webContext", webContext)
	slog.Debug("Cleaned JSON ready for parsing", "jsonText", jsonText)

	// Декодируем в облегченную структуру (без time.Time для deadline)
	var hackathons []AIHackathon
	if err := json.Unmarshal([]byte(jsonText), &hackathons); err != nil {
		slog.Error("JSON Unmarshal failed", "error", err, "raw_json", jsonText)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse AI response"})
		return
	}

	slog.Info("AI Search completed successfully", "results_count", len(hackathons))

	// Возвращаем результаты (кодом 200). В БД не сохраняем!
	c.JSON(http.StatusOK, hackathons)
}
