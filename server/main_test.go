package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestEnvironment() {
	// in-memory SQLite database for testing
	var err error
	db, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	if err := db.AutoMigrate(&URL{}); err != nil {
		panic("failed to migrate test database")
	}

	cache = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})
}

func TestCreateShortURL(t *testing.T) {
	setupTestEnvironment()

	r := gin.Default()
	r.POST("/shorten", createShortURL)

	payload := `{"url":"https://example.com"}`
	req, _ := http.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response URL
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.ShortCode)
	assert.Equal(t, "https://example.com", response.URL)
}

func TestRetrieveOriginalURL(t *testing.T) {
	setupTestEnvironment()

	testURL := URL{
		URL:       "https://example.com",
		ShortCode: "abc123",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	db.Create(&testURL)

	r := gin.Default()
	r.GET("/shorten/:shortCode", retrieveOriginalURL)

	req, _ := http.NewRequest(http.MethodGet, "/shorten/abc123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}

func TestUpdateShortURL(t *testing.T) {
	setupTestEnvironment()

	testURL := URL{
		URL:       "https://example.com",
		ShortCode: "abc123",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	db.Create(&testURL)

	r := gin.Default()
	r.PUT("/shorten/:shortCode", updateShortURL)

	payload := `{"url":"https://newexample.com"}`
	req, _ := http.NewRequest(http.MethodPut, "/shorten/abc123", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response URL
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "https://newexample.com", response.URL)
	assert.Equal(t, "abc123", response.ShortCode)
}

func TestDeleteShortURL(t *testing.T) {
	setupTestEnvironment()

	// Insert a test URL into the database
	testURL := URL{
		URL:       "https://cobaDulu.com",
		ShortCode: "abc123",
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	db.Create(&testURL)

	r := gin.Default()
	r.DELETE("/shorten/:shortCode", deleteShortURL)

	req, _ := http.NewRequest(http.MethodDelete, "/shorten/abc123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Check if the URL is deleted
	var url URL
	result := db.Where("short_code = ?", "abc123").First(&url)
	assert.Error(t, result.Error)
}

func TestGetURLStats(t *testing.T) {
	setupTestEnvironment()

	// Insert a test URL into the database
	testURL := URL{
		URL:         "https://cobaDulu.com",
		ShortCode:   "abc123",
		AccessCount: 5,
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}
	db.Create(&testURL)

	r := gin.Default()
	r.GET("/shorten/:shortCode/stats", getURLStats)

	req, _ := http.NewRequest(http.MethodGet, "/shorten/abc123/stats", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response URL
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "https://cobaDulu.com", response.URL)
	assert.Equal(t, "abc123", response.ShortCode)
	assert.Equal(t, 5, response.AccessCount)
}
