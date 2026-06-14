// news.go - Новостной агрегатор на Go (веб-сервер)
// Атрибуция: Данные предоставлены GNews API
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Article struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	URL         string `json:"url"`
	Image       string `json:"image"`
	PublishedAt string `json:"publishedAt"`
	Source      struct {
		Name string `json:"name"`
	} `json:"source"`
}

type GNewsResponse struct {
	TotalArticles int       `json:"totalArticles"`
	Articles      []Article `json:"articles"`
}

var apiKey string
var cache struct {
	Data      []Article
	Query     string
	Timestamp time.Time
}
var cacheDuration = 10 * time.Minute

func main() {
	apiKey = os.Getenv("GNEWS_API_KEY")
	if apiKey == "" {
		log.Println("⚠️ GNEWS_API_KEY не установлен, используйте переменную окружения")
	}
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/news", newsHandler)
	http.HandleFunc("/api/favorites", favoritesHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("🚀 Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, nil)
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query().Get("q")
	category := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		page, _ = strconv.Atoi(pageStr)
	}
	searchTerm := query
	if searchTerm == "" {
		searchTerm = category
		if searchTerm == "" {
			searchTerm = "general"
		}
	}
	// Кэширование
	if cache.Data != nil && cache.Query == searchTerm && time.Since(cache.Timestamp) < cacheDuration {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"articles": cache.Data,
			"total":    len(cache.Data),
		})
		return
	}
	articles, total := fetchNews(searchTerm, page)
	if articles != nil {
		cache.Data = articles
		cache.Query = searchTerm
		cache.Timestamp = time.Now()
	}
	json.NewEncoder(w).Encode(map[string]interface{}
