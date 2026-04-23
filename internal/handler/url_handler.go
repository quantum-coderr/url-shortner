package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"url-shortner/internal/domain"
	"url-shortner/internal/service"
)

type URLHandler struct {
	service *service.ShortenerService
	baseURL string
}

func NewURLHandler(service *service.ShortenerService, baseURL string) *URLHandler {
	return &URLHandler{
		service: service,
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (h *URLHandler) RegisterRoutes(engine *gin.Engine) {
	api := engine.Group("/api/v1")
	api.POST("/shorten", h.Shorten)
	api.GET("/analytics/:key", h.Analytics)
	engine.GET("/:key", h.Redirect)
}

type shortenRequest struct {
	URL string `json:"url" binding:"required"`
}

type shortenResponse struct {
	Key         string `json:"key"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *URLHandler) Shorten(c *gin.Context) {
	var req shortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	shortURL, err := h.service.Shorten(c.Request.Context(), req.URL)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, domain.ErrDuplicateKey):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, shortenResponse{
		Key:         shortURL.Key,
		ShortURL:    fmt.Sprintf("%s/%s", h.baseURL, shortURL.Key),
		OriginalURL: shortURL.OriginalURL,
	})
}

func (h *URLHandler) Redirect(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	shortURL, err := h.service.Resolve(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.service.RegisterClick(c.Request.Context(), key); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Redirect(http.StatusFound, shortURL.OriginalURL)
}

func (h *URLHandler) Analytics(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	clicks, err := h.service.Clicks(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":    key,
		"clicks": clicks,
	})
}
