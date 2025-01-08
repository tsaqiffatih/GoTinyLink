package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type URL struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	URL         string    `json:"url"`
	ShortCode   string    `json:"shortCode" gorm:"uniqueIndex"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	AccessCount int       `json:"accessCount"`
}

var (
	db    *gorm.DB
	cache *redis.Client
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatalf("DATABASE_URL is not set in the environment")
	}

	db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil || sqlDB.Ping() != nil {
		panic("failed to ping database")
	}

	log.Println("Successfully connected and pinged database")

	db.AutoMigrate(&URL{})

	// Redis connection
	cache = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_PORT"),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err = cache.Ping(context.Background()).Result()
	if err != nil {
		log.Println(err)
		panic("failed to connect Redis")
	}
	log.Println("Successfully connected to Redis")

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://example.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/shorten", createShortURL)
	r.GET("/shorten/:shortCode", retrieveOriginalURL)
	r.PUT("/shorten/:shortCode", updateShortURL)
	r.DELETE("/shorten/:shortCode", deleteShortURL)
	r.GET("/shorten/:shortCode/stats", getURLStats)

	r.Run(":8080")
}

// generate a random short code
func generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// create a new short URL
func createShortURL(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortCode := generateShortCode()
	newURL := &URL{
		ID:          uuid.NewString(),
		URL:         req.URL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AccessCount: 0,
	}

	if err := db.Create(newURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newURL)
}

// retrieve the original URL
func retrieveOriginalURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	ctx := c.Request.Context()

	// check in redis cache
	cachedURL, err := cache.Get(ctx, shortCode).Result()
	if err == nil {

		go func() {
			updateAccessCount(ctx, shortCode)
		}()
		c.Redirect(http.StatusFound, cachedURL)
		return
	}

	var url URL
	if err := db.WithContext(ctx).Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	url.AccessCount++
	db.WithContext(ctx).Save(&url)

	cache.Set(ctx, shortCode, url.URL, 10*time.Minute)

	c.Redirect(http.StatusFound, url.URL)
}

// update an existing short URL
func updateShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var url URL
	ctx := c.Request.Context()
	if err := db.WithContext(ctx).Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	url.URL = req.URL
	url.UpdatedAt = time.Now()
	db.WithContext(ctx).Save(&url)

	cache.Set(ctx, shortCode, req.URL, 10*time.Minute)

	c.JSON(http.StatusOK, url)
}

// delete a short URL
func deleteShortURL(c *gin.Context) {
	shortCode := c.Param("shortCode")
	ctx := c.Request.Context()

	if err := db.WithContext(ctx).Where("short_code = ?", shortCode).Delete(&URL{}).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	cache.Del(ctx, shortCode)

	c.Status(http.StatusNoContent)
}

// get statistics for a short URL
func getURLStats(c *gin.Context) {
	shortCode := c.Param("shortCode")

	var url URL
	if err := db.Where("short_code = ?", shortCode).First(&url).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Short URL not found"})
		return
	}

	c.JSON(http.StatusOK, url)
}

// update access count asynchronously
func updateAccessCount(ctx context.Context, shortCode string) {
	var url URL
	if err := db.WithContext(ctx).Where("short_code = ?", shortCode).First(&url).Error; err == nil {
		url.AccessCount++
		db.Save(&url)
	}
}
