package v1

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// context keys
type ctxKey string

const ctxKeyAPIKey ctxKey = "api_key"

// APIKey holds the authenticated key info extracted from a request.
type APIKey struct {
	ID                 int
	Name               string
	Tier               string
	DailyQuota         int
	RateLimitPerMinute int
}

// Middleware returns a chi middleware chain.
func Middleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. API Key extraction
			keyStr := extractAPIKey(r)
			if keyStr == "" {
				WriteError(w, http.StatusUnauthorized, "unauthorized", "API key required in X-API-Key header or query param")
				return
			}

			// 2. Validate key and load quota
			key, err := validateKey(r.Context(), pool, keyStr)
			if err != nil {
				WriteError(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
				return
			}

			// 3. Rate limit check
			if !checkRateLimit(key.ID, key.RateLimitPerMinute) {
				WriteError(w, http.StatusTooManyRequests, "rate_limited", "Rate limit exceeded")
				return
			}

			// 4. Quota check
			if !checkQuota(r.Context(), pool, key.ID, key.DailyQuota) {
				WriteError(w, http.StatusTooManyRequests, "quota_exceeded", "Daily quota exceeded")
				return
			}

			// Wrap response writer to capture status code for metering
			wr := &responseRecorder{ResponseWriter: w, statusCode: 200}

			start := time.Now()
			next.ServeHTTP(wr, r.WithContext(context.WithValue(r.Context(), ctxKeyAPIKey, key)))
			duration := time.Since(start)

			// 5. Metering
			recordMetering(r.Context(), pool, key.ID, r.Method, r.URL.Path, wr.statusCode, int(duration.Milliseconds()))
		})
	}
}

func extractAPIKey(r *http.Request) string {
	if v := r.Header.Get("X-API-Key"); v != "" {
		return v
	}
	if v := r.URL.Query().Get("api_key"); v != "" {
		return v
	}
	// In development, allow a default key
	if r.Header.Get("X-Env") == "development" {
		return "test-key-123"
	}
	return ""
}

func validateKey(ctx context.Context, pool *pgxpool.Pool, keyStr string) (APIKey, error) {
	var key APIKey
	err := pool.QueryRow(ctx, `
		SELECT id, name, tier, daily_quota, rate_limit_per_minute
		FROM api_keys
		WHERE key_hash = $1 AND is_active = TRUE
			AND (expires_at IS NULL OR expires_at > NOW())
	`, keyStr).Scan(&key.ID, &key.Name, &key.Tier, &key.DailyQuota, &key.RateLimitPerMinute)
	if err != nil {
		return key, err
	}
	return key, nil
}

// In-memory rate limiter (per key ID, per minute)
var (
	rateLimitMu      sync.Mutex
	rateLimitBuckets = make(map[int]*rateLimitBucket)
)

type rateLimitBucket struct {
	count     int
	resetTime time.Time
}

func checkRateLimit(keyID, limitPerMinute int) bool {
	rateLimitMu.Lock()
	defer rateLimitMu.Unlock()

	now := time.Now()
	b, ok := rateLimitBuckets[keyID]
	if !ok || now.After(b.resetTime) {
		rateLimitBuckets[keyID] = &rateLimitBucket{count: 1, resetTime: now.Add(time.Minute)}
		return true
	}
	if b.count >= limitPerMinute {
		return false
	}
	b.count++
	return true
}

func checkQuota(ctx context.Context, pool *pgxpool.Pool, keyID, dailyQuota int) bool {
	if dailyQuota <= 0 {
		return true
	}
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM metering
		WHERE api_key_id = $1 AND created_at >= CURRENT_DATE
	`, keyID).Scan(&count)
	if err != nil {
		return true // fail open on metering error
	}
	return count < dailyQuota
}

func recordMetering(ctx context.Context, pool *pgxpool.Pool, keyID int, method, endpoint string, statusCode, durationMs int) {
	_, _ = pool.Exec(ctx, `
		INSERT INTO metering (api_key_id, endpoint, method, status_code, response_time_ms)
		VALUES ($1, $2, $3, $4, $5)
	`, keyID, endpoint, method, statusCode, durationMs)
}

// responseRecorder wraps http.ResponseWriter to capture status code.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.statusCode = code
		rr.written = true
		rr.ResponseWriter.WriteHeader(code)
	}
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		rr.WriteHeader(http.StatusOK)
	}
	return rr.ResponseWriter.Write(b)
}

// GetAPIKey retrieves the authenticated API key from context.
func GetAPIKey(ctx context.Context) (APIKey, bool) {
	key, ok := ctx.Value(ctxKeyAPIKey).(APIKey)
	return key, ok
}

// OptionalAuth middleware allows requests without keys but attaches a nil key.
func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// intQueryParam extracts an integer query parameter with a default.
func intQueryParam(r *http.Request, name string, fallback int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
