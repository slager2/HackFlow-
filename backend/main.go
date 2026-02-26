package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Hackathon представляет структуру данных события.
// JSON-теги написаны в camelCase для удобной интеграции с фронтендом.
type Hackathon struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Date     string `json:"date"`
	Format   string `json:"format"`
	City     string `json:"city"`
	AgeLimit string `json:"ageLimit"`
	Link     string `json:"link"`
}

// in-memory база данных (срез) реальных хакатонов
var hackathons = []Hackathon{
	{
		ID:       "1",
		Title:    "Decentrathon",
		Date:     "15 Марта 2026",
		Format:   "Офлайн",
		City:     "Астана",
		AgeLimit: "18+",
		Link:     "https://decentrathon.io",
	},
	{
		ID:       "2",
		Title:    "AITUCAP Hackathon",
		Date:     "10 Апреля 2026",
		Format:   "Офлайн",
		City:     "Алматы",
		AgeLimit: "16+",
		Link:     "https://aitucap.kz",
	},
	{
		ID:       "3",
		Title:    "NASA Space Apps Challenge",
		Date:     "5 Октября 2026",
		Format:   "Онлайн",
		City:     "Global",
		AgeLimit: "Нет ограничений",
		Link:     "https://spaceappschallenge.org",
	},
	{
		ID:       "4",
		Title:    "HackFlow Beta Launch",
		Date:     "20 Мая 2026",
		Format:   "Офлайн / Онлайн",
		City:     "Астана",
		AgeLimit: "16+",
		Link:     "https://hackflow.dev",
	},
}

func main() {
	// Инициализируем маршрутизатор Gin
	r := gin.Default()

	// Настройка CORS middleware
	// Разрешаем запросы только с фронтенда на Next.js (localhost:3000)
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type"}
	r.Use(cors.New(config))

	// Эндпоинт для получения списка хакатонов с поддержкой поиска
	r.GET("/api/hackathons", getHackathons)

	// Запускаем сервер на порту 8080
	log.Println("Сервер запущен на http://localhost:8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Ошибка при запуске сервера: %v", err)
	}
}

// getHackathons обрабатывает GET-запросы на поиск хакатонов
func getHackathons(c *gin.Context) {
	// Получаем query-параметр "q" и убираем лишние пробелы
	query := strings.TrimSpace(c.Query("q"))

	// Если параметр поиска пустой, возвращаем весь список
	if query == "" {
		c.JSON(http.StatusOK, hackathons)
		return
	}

	// Переводим запрос в нижний регистр для case-insensitive поиска
	lowerQuery := strings.ToLower(query)

	// Создаем пустой срез для хранения результатов поиска.
	// Инициализируем его сразу, чтобы возвращать [] вместо null, если совпадений нет.
	filtered := make([]Hackathon, 0)

	// Проходим циклом по всем хакатонам в нашей in-memory базе
	for _, h := range hackathons {
		// Переводим Title и City текущего хакатона в нижний регистр
		lowerTitle := strings.ToLower(h.Title)
		lowerCity := strings.ToLower(h.City)

		// Проверяем частичное совпадение (подстроки) в названии ИЛИ городе
		if strings.Contains(lowerTitle, lowerQuery) || strings.Contains(lowerCity, lowerQuery) {
			filtered = append(filtered, h)
		}
	}

	// Отправляем отфильтрованный список (с кодом 200 OK)
	c.JSON(http.StatusOK, filtered)
}
