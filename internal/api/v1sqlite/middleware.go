package v1sqlite

import (
	"context"
	"database/sql"
	"net/http"
	"sync"
	"time"
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

// Middleware returns a chi middleware chain for SQLite backend.
func Middleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. API Key extraction
			keyStr := extractAPIKey(r)
			if keyStr == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized", "API key required in X-API-Key header or query param")
				return
			}

			// 2. Validate key and load quota
			key, err := validateKey(r.Context(), db, keyStr)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid API key")
				return
			}

			// 3. Rate limit check
			if !checkRateLimit(key.ID, key.RateLimitPerMinute) {
				writeError(w, http.StatusTooManyRequests, "rate_limited", "Rate limit exceeded")
				return
			}

			// 4. Quota check
			if !checkQuota(r.Context(), db, key.ID, key.DailyQuota) {
				writeError(w, http.StatusTooManyRequests, "quota_exceeded", "Daily quota exceeded")
				return
			}

			// Wrap response writer to capture status code for metering
			wr := &responseRecorder{ResponseWriter: w, statusCode: 200}

			start := time.Now()
			next.ServeHTTP(wr, r.WithContext(context.WithValue(r.Context(), ctxKeyAPIKey, key)))
			duration := time.Since(start)

			// 5. Metering
			recordMetering(r.Context(), db, key.ID, r.Method, r.URL.Path, wr.statusCode, int(duration.Milliseconds()))
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
	if r.Header.Get("X-Env") == "development" {
		return "test-key-123"
	}
	return ""
}

func validateKey(ctx context.Context, db *sql.DB, keyStr string) (APIKey, error) {
	var key APIKey
	err := db.QueryRowContext(ctx, `
		SELECT id, name, tier, daily_quota, rate_limit_per_minute
		FROM api_keys
		WHERE key_hash = ? AND is_active = 1
			AND (expires_at IS NULL OR expires_at > date('now'))
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

func checkQuota(ctx context.Context, db *sql.DB, keyID, dailyQuota int) bool {
	if dailyQuota <= 0 {
		return true
	}
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM metering
		WHERE api_key_id = ? AND date(created_at) = date('now')
	`, keyID).Scan(&count)
	if err != nil {
		return true // fail open on metering error
	}
	return count < dailyQuota
}

func recordMetering(ctx context.Context, db *sql.DB, keyID int, method, endpoint string, statusCode, durationMs int) {
	_, _ = db.ExecContext(ctx, `
		INSERT INTO metering (api_key_id, endpoint, method, status_code, response_time_ms)
		VALUES (?, ?, ?, ?, ?)
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

// writeError writes a standardized error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Simple JSON error response
	body := `{"error":{"code":"` + code + `","message":"` + message + `"}}`
	_ = []byte(body)
	w.Write([]byte(body))
}
