package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"

	"url-shortner/internal/analytics"
	analyticsMemory "url-shortner/internal/analytics/memory"
	analyticsRedis "url-shortner/internal/analytics/redis"
	"url-shortner/internal/cache"
	redisCache "url-shortner/internal/cache/redis"
	"url-shortner/internal/generator"
	"url-shortner/internal/handler"
	"url-shortner/internal/repository"
	"url-shortner/internal/repository/cached"
	"url-shortner/internal/repository/memory"
	"url-shortner/internal/service"
)

func main() {
	baseRepo := memory.NewURLRepository()
	infra := newInfra()
	urlCache := infra.urlCache
	var repo repository.URLRepository = cached.NewURLRepository(baseRepo, urlCache)

	keyGenerator := generator.NewBase62Generator(0)
	shortenerService := service.NewShortenerService(repo, keyGenerator, infra.clickCounter)

	engine := gin.Default()
	urlHandler := handler.NewURLHandler(shortenerService, "http://localhost:8080")
	urlHandler.RegisterRoutes(engine)

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

type infra struct {
	urlCache     cache.URLCache
	clickCounter analytics.ClickCounter
}

func newInfra() infra {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	client := goredis.NewClient(&goredis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fallbackNoopInfra(client, "redis unavailable at %s, using local fallback analytics+cache: %v", addr, err)
	}

	probeKey := "url-shortner:startup-probe"
	probeValue := time.Now().UTC().Format(time.RFC3339Nano)

	if err := client.Set(ctx, probeKey, probeValue, 30*time.Second).Err(); err != nil {
		return fallbackNoopInfra(client, "redis write probe failed at %s, using local fallback analytics+cache: %v", addr, err)
	}

	cachedProbeValue, err := client.Get(ctx, probeKey).Result()
	if err != nil {
		return fallbackNoopInfra(client, "redis read probe failed at %s, using local fallback analytics+cache: %v", addr, err)
	}

	if cachedProbeValue != probeValue {
		return fallbackNoopInfra(client, "redis probe mismatch at %s, using local fallback analytics+cache", addr)
	}

	log.Printf("redis cache and analytics enabled at %s (startup read/write probe ok)", addr)
	return infra{
		urlCache:     redisCache.NewURLCache(client, 10*time.Minute),
		clickCounter: analyticsRedis.NewClickCounter(client),
	}
}

func fallbackNoopInfra(client *goredis.Client, format string, args ...any) infra {
	log.Printf(format, args...)
	if closeErr := client.Close(); closeErr != nil {
		log.Printf("redis close error: %v", closeErr)
	}
	return infra{
		urlCache:     cache.NewNoopURLCache(),
		clickCounter: analyticsMemory.NewClickCounter(),
	}
}
