package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type URL struct {
	ID          string    `json:"id"`
	URL         string    `json:"url"`
	ShortCode   string    `json:"shortCode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessCount"`
}

var (
	urlStore = make(map[string]*URL)
	mu       sync.RWMutex
)

func main() {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"}, // Daftar origin yang diizinkan
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},                 // Metode HTTP yang diizinkan
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},      // Header yang diizinkan
		ExposeHeaders:    []string{"Content-Length"},                               // Header yang dapat diakses di client
		AllowCredentials: true,                                                     // Izinkan cookie untuk cross-origin
		MaxAge:           12 * time.Hour,                                           // Cache preflight request selama 12 jam
	}))

	r.POST("/shorten", createShortURL)
	r.GET("/shorten/:shortCode", retrieveOriginalURL)
	r.PUT("/shorten/:shortCode", updateShortURL)
	r.DELETE("/shorten/:shortCode", deleteShortURL)
	r.GET("/shorten/:shortCode/stats", getURLStats)

	r.Run(":8080")
}

// Generate a random short code
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Create a new short URL
func createShortURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	shortCode := generateShortCode()
	newURL := &URL{
		ID:          uuid.NewString(),
		URL:         req.URL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AccessCount: 0,
	}
	urlStore[shortCode] = newURL

	c.JSON(http.StatusCreated, newURL)
}

// Retrieve the original URL
func retrieveOriginalURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	mu.RLock()
	defer mu.RUnlock()

	url, exists := urlStore[shortCode]
	if !exists {
		// if not found, user will redirect to fe url
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	url.AccessCount++
	// c.JSON(http.StatusOK, url)
	c.Redirect(http.StatusFound, url.URL)
}

// Update an existing short URL
func updateShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	url, exists := urlStore[shortCode]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	url.URL = req.URL
	url.UpdatedAt = time.Now()
	c.JSON(http.StatusOK, url)
}

// Delete a short URL
func deleteShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	mu.Lock()
	defer mu.Unlock()

	if _, exists := urlStore[shortCode]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	delete(urlStore, shortCode)
	c.Status(http.StatusNoContent)
}

// Get statistics for a short URL
func getURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")

	mu.RLock()
	defer mu.RUnlock()

	url, exists := urlStore[shortCode]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, url)
}
