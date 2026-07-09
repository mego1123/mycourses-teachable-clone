package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"mycourses/internal/auth"
	"mycourses/internal/db"
	"mycourses/internal/models"
)

type contextKey string

const (
	UserContextKey           contextKey = "user"
	APIKeyContextKey         contextKey = "apikey"
	ImpersonatedByContextKey contextKey = "impersonatedBy"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
	db         *db.DB
}

func NewAuthMiddleware(jwtService *auth.JWTService, database *db.DB) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		db:         database,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		// API key authentication
		if strings.HasPrefix(tokenString, "lsk_") {
			m.authenticateAPIKey(w, r, next, tokenString)
			return
		}

		// JWT authentication
		m.authenticateJWT(w, r, next, tokenString)
	})
}

func (m *AuthMiddleware) authenticateJWT(w http.ResponseWriter, r *http.Request, next http.Handler, tokenString string) {
	claims, err := m.jwtService.ValidateAccessToken(tokenString)
	if err != nil {
		if err == auth.ErrExpiredToken {
			http.Error(w, `{"error":"Token has expired"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
		return
	}

	if m.isTokenRevoked(r.Context(), tokenString) {
		http.Error(w, `{"error":"Token has been revoked"}`, http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"Invalid user ID"}`, http.StatusUnauthorized)
		return
	}

	// Query Postgres for user
	pgUser, err := m.db.Queries.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
		return
	}

	// Convert to domain model
	user := db.ToUser(pgUser)

	if !user.IsActive {
		http.Error(w, `{"error":"User account is inactive"}`, http.StatusUnauthorized)
		return
	}

	ctx := context.WithValue(r.Context(), UserContextKey, &user)
	if claims.ImpersonatedBy != "" {
		ctx = context.WithValue(ctx, ImpersonatedByContextKey, claims.ImpersonatedBy)
	}
	next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *AuthMiddleware) authenticateAPIKey(w http.ResponseWriter, r *http.Request, next http.Handler, rawKey string) {
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := base64.RawURLEncoding.EncodeToString(hash[:])

	// Query Postgres for API key by hash
	pgKey, err := m.db.Queries.GetAPIKeyByHash(r.Context(), keyHash)
	if err != nil {
		http.Error(w, `{"error":"Invalid API key"}`, http.StatusUnauthorized)
		return
	}

	apiKey := db.ToAPIKey(pgKey)

	// Look up key creator
	if apiKey.CreatedBy == uuid.Nil {
		http.Error(w, `{"error":"API key owner not found"}`, http.StatusUnauthorized)
		return
	}

	pgUser, err := m.db.Queries.GetUserByID(r.Context(), apiKey.CreatedBy)
	if err != nil || !pgUser.IsActive {
		http.Error(w, `{"error":"API key owner account is inactive"}`, http.StatusUnauthorized)
		return
	}

	user := db.ToUser(pgUser)

	ctx := context.WithValue(r.Context(), UserContextKey, &user)

	// Admin keys: auto-resolve root tenant + admin membership
	if apiKey.Authority == models.APIKeyAuthorityAdmin {
		pgTenant, err := m.db.Queries.GetRootTenant(r.Context())
		if err != nil {
			http.Error(w, `{"error":"System configuration error"}`, http.StatusInternalServerError)
			return
		}
		rootTenant := db.ToTenant(pgTenant)
		ctx = context.WithValue(ctx, TenantContextKey, &rootTenant)
		ctx = context.WithValue(ctx, MembershipContextKey, &models.TenantMembership{
			UserID:   user.ID,
			TenantID: rootTenant.ID,
			Role:     models.RoleAdmin,
		})
	}

	ctx = context.WithValue(ctx, APIKeyContextKey, &apiKey)

	// Update lastUsedAt asynchronously
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		m.db.Queries.UpdateAPIKeyLastUsed(ctx, apiKey.ID)
	}()

	next.ServeHTTP(w, r.WithContext(ctx))
}

func (m *AuthMiddleware) isTokenRevoked(ctx context.Context, rawToken string) bool {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := base64.StdEncoding.EncodeToString(hash[:])

	// Query Postgres for revoked token
	var count int64
	err := m.db.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM revoked_tokens WHERE jti = $1", tokenHash).Scan(&count)
	if err != nil {
		slog.Warn("revoked-token lookup failed, denying access", "error", err)
		return true
	}
	return count > 0
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*models.User)
	return user, ok
}

func GetAPIKeyFromContext(ctx context.Context) (*models.APIKey, bool) {
	key, ok := ctx.Value(APIKeyContextKey).(*models.APIKey)
	return key, ok
}

func GetImpersonatedBy(ctx context.Context) string {
	v, _ := ctx.Value(ImpersonatedByContextKey).(string)
	return v
}
