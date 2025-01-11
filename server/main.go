package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/speps/go-hashids"
)

type URL struct {
	gorm.Model
	URL         string    `json:"url"`
	ShortCode   string    `json:"shortCode" gorm:"uniqueIndex"`
	AccessCount int       `json:"accessCount"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type URLResponse struct {
	Success     bool      `json:"success"`
	URL         string    `json:"url"`
	ShortCode   string    `json:"shortCode"`
	AccessCount int       `json:"accessCount"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

var (
	db                 *gorm.DB
	cache              *redis.Client
	hashSalt           string
	tickerInterval     = 24 * time.Hour
	expirationInterval = 24 * 30 * time.Hour
)

var rateLimitStore = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_PORT"),
	Username: os.Getenv("REDIS_USERNAME"),
	Password: os.Getenv("REDIS_PASSWORD"),
	DB:       0,
})

func main() {
	if _, err := os.Stat(".env"); err == nil {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
		}
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatalf("DATABASE_URL is not set in the environment")
	}

	hashSalt = os.Getenv("HASH_SALT")
	if hashSalt == "" {
		log.Fatalf("HASH_SALT is not set in the environment")
	}

	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		log.Fatalf("ALLOW_ORIGINS is not set in the environment")
	}
	origins := strings.Split(allowOrigins, ",")

	var err error
	db, err = gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil || sqlDB.Ping() != nil {
		panic("failed to ping database")
	}

	log.Println("Successfully connected and pinged database")

	db.Migrator().DropTable(&URL{})

	if err := db.AutoMigrate(&URL{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

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

	r.Use(RateLimiterWithBlacklist(10, 1*time.Minute, 10*time.Minute))

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
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

	// ticker to delete expired URLs every 24 hours
	go func() {
		ticker := time.NewTicker(tickerInterval)
		defer ticker.Stop()
		for {
			<-ticker.C
			deleteExpiredURLs()
		}
	}()

	r.Run(":8080")
}

// ticker for deleting expired URLs
func deleteExpiredURLs() {
	now := time.Now()
	result := db.Where("expires_at <= ?", now).Delete(&URL{})
	log.Printf("Deleted %d expired URLs", result.RowsAffected)
}

// middleware Rate Limiter dengan blacklist ip
func RateLimiterWithBlacklist(limit int, window time.Duration, blacklistTTL time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := context.Background()

		isBlacklisted, err := rateLimitStore.Get(ctx, "blacklist:"+ip).Result()
		if err == nil && isBlacklisted == "1" {
			// Jika IP ada dalam blacklist, kembalikan error 403
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied. Your IP is temporarily blocked.",
			})
			return
		}

		key := "rate_limiter:" + ip
		count, err := rateLimitStore.Incr(ctx, key).Result()
		if err != nil {
			log.Println("Redis error:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		if count == 1 {
			rateLimitStore.Expire(ctx, key, window)
		}

		if int(count) > limit {
			rateLimitStore.Set(ctx, "blacklist:"+ip, "1", blacklistTTL)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded. Your IP has been temporarily blocked.",
				"limit":       limit,
				"time_window": window.String(),
			})
			return
		}

		c.Next()
	}
}

func generateShortCode(id uint) string {
	hd := hashids.NewData()
	hd.Salt = hashSalt
	hd.MinLength = 6
	h, _ := hashids.NewWithData(hd)
	hash, _ := h.Encode([]int{int(id)})
	return hash
}

// create a new short URL
func createShortURL(c *gin.Context) {
	var newURL URL

	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newURL.URL = req.URL
	newURL.AccessCount = 0
	newURL.ExpiresAt = time.Now().Add(expirationInterval)
	// newURL.ExpiresAt = time.Now().Add(30 * 24 * time.Hour)

	if err := db.Create(&newURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	shortCode := generateShortCode(newURL.ID)
	newURL.ShortCode = shortCode

	if err := db.Save(&newURL).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := URLResponse{
		Success:     true,
		URL:         newURL.URL,
		ShortCode:   newURL.ShortCode,
		AccessCount: newURL.AccessCount,
		ExpiresAt:   newURL.ExpiresAt,
	}

	c.JSON(http.StatusCreated, response)
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

	cache.Set(ctx, shortCode, url.URL, 30*24*time.Hour)

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

	cache.Set(ctx, shortCode, req.URL, expirationInterval)

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
