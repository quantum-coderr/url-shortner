package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	analyticsMemory "url-shortner/internal/analytics/memory"
	repoMemory "url-shortner/internal/repository/memory"
	"url-shortner/internal/service"
)

type fixedGenerator struct {
	key string
}

func (g fixedGenerator) NextKey() string {
	return g.key
}

func setupTestServer() (*gin.Engine, *service.ShortenerService) {
	gin.SetMode(gin.TestMode)

	repo := repoMemory.NewURLRepository()
	svc := service.NewShortenerService(repo, fixedGenerator{key: "abc123"}, analyticsMemory.NewClickCounter())

	engine := gin.New()
	h := NewURLHandler(svc, "http://localhost:8080")
	h.RegisterRoutes(engine)

	return engine, svc
}

func TestShortenHandler(t *testing.T) {
	engine, _ := setupTestServer()

	body := `{"url":"https://example.com/page"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, res.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if got["key"] != "abc123" {
		t.Fatalf("expected key abc123, got %q", got["key"])
	}
}

func TestShortenHandlerRejectsInvalidURL(t *testing.T) {
	engine, _ := setupTestServer()

	body := `{"url":"not-a-url"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/shorten", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	engine.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, res.Code)
	}
}

func TestRedirectHandler(t *testing.T) {
	engine, svc := setupTestServer()

	_, err := svc.Shorten(context.Background(), "https://example.com/a")
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	res := httptest.NewRecorder()
	engine.ServeHTTP(res, req)

	if res.Code != http.StatusFound {
		t.Fatalf("expected %d, got %d", http.StatusFound, res.Code)
	}

	location := res.Header().Get("Location")
	if location != "https://example.com/a" {
		t.Fatalf("expected redirect location %q, got %q", "https://example.com/a", location)
	}

	clicks, err := svc.Clicks(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("clicks: %v", err)
	}

	if clicks != 1 {
		t.Fatalf("expected clicks=1, got %d", clicks)
	}
}

func TestRedirectHandlerNotFound(t *testing.T) {
	engine, _ := setupTestServer()

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	res := httptest.NewRecorder()
	engine.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestAnalyticsHandler(t *testing.T) {
	engine, svc := setupTestServer()

	_, err := svc.Shorten(context.Background(), "https://example.com/a")
	if err != nil {
		t.Fatalf("shorten: %v", err)
	}
	if err := svc.RegisterClick(context.Background(), "abc123"); err != nil {
		t.Fatalf("register click: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/abc123", nil)
	res := httptest.NewRecorder()
	engine.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, res.Code)
	}

	var got struct {
		Key    string `json:"key"`
		Clicks uint64 `json:"clicks"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got.Key != "abc123" {
		t.Fatalf("expected key abc123, got %q", got.Key)
	}
	if got.Clicks != 1 {
		t.Fatalf("expected clicks=1, got %d", got.Clicks)
	}
}
