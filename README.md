# URL Shortener 

A step-by-step URL shortener built in Go to learn layered backend design.

## Tech stack

- Go
- Gin (HTTP)
- Redis (cache + analytics counter)
- In-memory URL storage (for now)

## Architecture

`handlers -> services -> repository -> cache -> generator`

Analytics path:

`handlers -> services -> analytics counter`

### What each layer does

- **handlers**: parse HTTP requests, return JSON/redirect/status codes
- **services**: business rules (URL validation, orchestration)
- **repository**: URL storage abstraction
- **cache**: fast lookup by short key (Redis; noop fallback if Redis unavailable)
- **generator**: Base62 key generation
- **analytics**: click counter (Redis `INCR`; in-memory fallback)

## Implemented features

1. Shorten URL
2. Redirect by short key
3. Base62 key generation
4. Repository abstraction
5. Redis cache layer
6. Click analytics endpoint

## Run Redis (Docker-first)

Start Redis:

```bash
docker run --name url-shortner-redis -p 6379:6379 -d redis:7-alpine
```

Check Redis container:

```bash
docker ps --filter name=url-shortner-redis
```

Stop/remove later:

```bash
docker stop url-shortner-redis
docker rm url-shortner-redis
```

## Run the app

Use IPv4 localhost to avoid IPv6 localhost issues:

```bash
export REDIS_ADDR=127.0.0.1:6379
go run ./cmd/server
```

When Redis is healthy, startup log shows:

```text
redis cache and analytics enabled at 127.0.0.1:6379 (startup read/write probe ok)
```

If Redis is unavailable, app still runs with local fallback:

- noop cache
- in-memory analytics counter

## API examples

### 1) Shorten a URL

```bash
curl -i -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```

Example response:

```json
{
  "key": "1",
  "short_url": "http://localhost:8080/1",
  "original_url": "https://example.com"
}
```

### 2) Redirect with short key

```bash
curl -i http://localhost:8080/1
```

### 3) Get click analytics

```bash
curl -s http://localhost:8080/api/v1/analytics/1
```

Example response:

```json
{
  "key": "1",
  "clicks": 3
}
```

## Useful logs to understand behavior

- `cache warm key=...`
- `cache hit key=...`
- `cache miss key=...`
- `storage read key=...`
- `analytics increment key=... (redis)`
- `analytics read key=... clicks=... (redis)`

## Tests

```bash
go test ./...
```
