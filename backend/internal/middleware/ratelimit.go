package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter provides in-memory rate limiting (no MongoDB dependency).
// For production with multiple instances, replace with Redis-backed implementation.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*rateBucket
}

type rateBucket struct {
	count    int
	expiresAt time.Time
}

func NewDistributedRateLimiter() *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*rateBucket)}
}

// Allow checks if the request is within rate limits.
func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.buckets[key]
	if !exists || now.After(bucket.expiresAt) {
		rl.buckets[key] = &rateBucket{count: 1, expiresAt: now.Add(window)}
		return true
	}

	if bucket.count >= limit {
		return false
	}

	bucket.count++
	return true
}

// RateLimitHandler returns middleware that enforces rate limits.
func (rl *RateLimiter) RateLimitHandler(limit int, window time.Duration, keyFn func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)
			if key == "" {
				key = GetClientIP(r)
			}

			if !rl.Allow(key, limit, window) {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP extracts the client IP from the request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// Cleanup removes expired buckets periodically.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, bucket := range rl.buckets {
		if now.After(bucket.expiresAt) {
			delete(rl.buckets, key)
		}
	}
}

// Rate limit constants (were previously in MongoDB-backed implementation)
const (
	AccountCreationLimit   = 5
	LoginAttemptLimit      = 10
	TokenRefreshLimit      = 30
	EmailVerificationLimit = 5
	ResendVerificationLimit = 3
	PasswordResetLimit     = 5
	MFACodeLimit           = 5
	MagicLinkLimit         = 5
)

const (
	ResetTokenVerifyLimit = 5
	MFAChallengeLimit     = 5
	MagicLinkVerifyLimit  = 5
	OAuthInitLimit        = 10
)

const (
	InvitationLimit   = 5
	UsageRecordLimit  = 100
	TelemetryLimit    = 100
)

const (
	TelemetryAnonymousLimit   = 100
	TelemetryAuthenticatedLimit = 200
	CSVExportLimit            = 5
)
