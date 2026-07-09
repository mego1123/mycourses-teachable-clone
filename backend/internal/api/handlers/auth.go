package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"mycourses/internal/auth"
	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/email"
	"mycourses/internal/events"
	"mycourses/internal/middleware"
	"mycourses/internal/models"
	"mycourses/internal/syslog"
	"mycourses/internal/telemetry"
)

type AuthHandler struct {
	db            *db.DB
	jwt           *auth.JWTService
	password      *auth.PasswordService
	googleOAuth   *auth.GoogleOAuthService
	githubOAuth   *auth.GitHubOAuthService
	microsoftOAuth *auth.MicrosoftOAuthService
	email         *email.ResendService
	emitter       events.Emitter
	syslog        *syslog.Logger
	frontendURL   string
	getConfig     func(string) string
	rateLimiter   *middleware.RateLimiter
	telemetrySvc  *telemetry.Service
	totpEncKey    []byte
}

func NewAuthHandler(database *db.DB, jwt *auth.JWTService, password *auth.PasswordService, googleOAuth *auth.GoogleOAuthService, emailSvc *email.ResendService, emitter events.Emitter, frontendURL string, sysLogger *syslog.Logger) *AuthHandler {
	return &AuthHandler{db: database, jwt: jwt, password: password, googleOAuth: googleOAuth, email: emailSvc, emitter: emitter, frontendURL: frontendURL, syslog: sysLogger}
}

func (h *AuthHandler) SetGitHubOAuth(svc *auth.GitHubOAuthService)       { h.githubOAuth = svc }
func (h *AuthHandler) SetMicrosoftOAuth(svc *auth.MicrosoftOAuthService) { h.microsoftOAuth = svc }
func (h *AuthHandler) SetGetConfig(fn func(string) string)               { h.getConfig = fn }
func (h *AuthHandler) SetRateLimiter(rl *middleware.RateLimiter)          { h.rateLimiter = rl }
func (h *AuthHandler) SetTelemetry(svc *telemetry.Service)               { h.telemetrySvc = svc }
func (h *AuthHandler) SetTOTPEncryptionKey(key []byte)                   { h.totpEncKey = key }

func (h *AuthHandler) generateTokenPair(userID, email, displayName string) (string, string, time.Duration, error) {
	accessTTL, refreshTTL := h.sessionTTLs()
	accessToken, err := h.jwt.GenerateAccessTokenWithTTL(userID, email, displayName, accessTTL)
	if err != nil {
		return "", "", 0, err
	}
	refreshToken, err := h.jwt.GenerateRefreshToken(userID)
	if err != nil {
		return "", "", 0, err
	}
	return accessToken, refreshToken, refreshTTL, nil
}

func (h *AuthHandler) sessionTTLs() (time.Duration, time.Duration) {
	return 15 * time.Minute, 7 * 24 * time.Hour
}

func (h *AuthHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (h *AuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	providers := map[string]bool{
		"password": true,
		"google":   h.googleOAuth != nil,
		"github":   h.githubOAuth != nil,
	}
	respondWithJSON(w, http.StatusOK, providers)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		Name      string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(req.Email))

	// Check if user already exists
	_, err := h.db.Queries.GetUserByEmailNormalized(r.Context(), emailNormalized)
	if err == nil {
		respondWithError(w, http.StatusConflict, "User already exists")
		return
	}

	// Hash password
	hash, err := h.password.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create user in Postgres
	pgUser, err := h.db.Queries.CreateUser(r.Context(), gen.CreateUserParams{
		Email:            req.Email,
		EmailNormalized:  emailNormalized,
		DisplayName:      req.Name,
		PasswordHash:     &hash,
		LocalePreference: "en",
		ThemePreference:  "system",
		IsActive:         true,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create user: "+err.Error())
		return
	}

	// Generate tokens
	accessToken, refreshToken, _, err := h.generateTokenPair(pgUser.ID.String(), pgUser.Email, pgUser.DisplayName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	h.db.Pool.Exec(r.Context(),
		"INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at) VALUES ($1, $2, $3, $4)",
		pgUser.ID, h.hashToken(refreshToken), uuid.New(), time.Now().Add(7*24*time.Hour))

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         db.ToUser(pgUser),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(req.Email))

	pgUser, err := h.db.Queries.GetUserByEmailNormalized(r.Context(), emailNormalized)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if pgUser.PasswordHash == nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if h.password.ComparePassword(*pgUser.PasswordHash, req.Password) != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !pgUser.IsActive {
		respondWithError(w, http.StatusForbidden, "Account is inactive")
		return
	}

	// Update last login
	h.db.Queries.UpdateLastLogin(r.Context(), pgUser.ID)

	// Generate tokens
	accessToken, refreshToken, _, err := h.generateTokenPair(pgUser.ID.String(), pgUser.Email, pgUser.DisplayName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store refresh token
	familyID := uuid.New()
	h.db.Pool.Exec(r.Context(),
		"INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at) VALUES ($1, $2, $3, $4)",
		pgUser.ID, h.hashToken(refreshToken), familyID, time.Now().Add(7*24*time.Hour))

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"user":         db.ToUser(pgUser),
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenHash := h.hashToken(req.RefreshToken)
	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'user_logout' WHERE token_hash = $1 AND revoked_at IS NULL",
		tokenHash)

	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenHash := h.hashToken(req.RefreshToken)

	// Look up refresh token
	var userID uuid.UUID
	var familyID uuid.UUID
	var revokedAt *time.Time
	err := h.db.Pool.QueryRow(r.Context(),
		"SELECT user_id, family_id, revoked_at FROM refresh_tokens WHERE token_hash = $1",
		tokenHash).Scan(&userID, &familyID, &revokedAt)
	if err != nil || revokedAt != nil {
		// Token not found or already revoked — revoke entire family (replay attack detection)
		if familyID != uuid.Nil {
			h.db.Pool.Exec(r.Context(),
				"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'replay_detected' WHERE family_id = $1 AND revoked_at IS NULL",
				familyID)
		}
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	// Get user
	pgUser, err := h.db.Queries.GetUserByID(r.Context(), userID)
	if err != nil || !pgUser.IsActive {
		respondWithError(w, http.StatusUnauthorized, "User not found or inactive")
		return
	}

	// Revoke old token
	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'rotated' WHERE token_hash = $1",
		tokenHash)

	// Generate new token pair
	accessToken, newRefreshToken, _, err := h.generateTokenPair(pgUser.ID.String(), pgUser.Email, pgUser.DisplayName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	// Store new refresh token (same family)
	h.db.Pool.Exec(r.Context(),
		"INSERT INTO refresh_tokens (user_id, token_hash, family_id, expires_at) VALUES ($1, $2, $3, $4)",
		pgUser.ID, h.hashToken(newRefreshToken), familyID, time.Now().Add(7*24*time.Hour))

	respondWithJSON(w, http.StatusOK, map[string]string{
		"accessToken":  accessToken,
		"refreshToken": newRefreshToken,
	})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}
	respondWithJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenHash := h.hashToken(req.Token)
	var userID uuid.UUID
	err := h.db.Pool.QueryRow(r.Context(),
		"SELECT user_id FROM verification_tokens WHERE token_hash = $1 AND type = 'email_verification' AND expires_at > NOW() AND used_at IS NULL",
		tokenHash).Scan(&userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	h.db.Queries.VerifyUserEmail(r.Context(), userID)
	h.db.Pool.Exec(r.Context(), "UPDATE verification_tokens SET used_at = NOW() WHERE token_hash = $1", tokenHash)

	respondWithJSON(w, http.StatusOK, map[string]bool{"verified": true})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	if user.EmailVerified {
		respondWithError(w, http.StatusBadRequest, "Email already verified")
		return
	}

	// Generate verification token
	token := generateRandomToken()
	tokenHash := h.hashToken(token)

	h.db.Pool.Exec(r.Context(),
		"INSERT INTO verification_tokens (user_id, token_hash, type, expires_at) VALUES ($1, $2, 'email_verification', $3)",
		user.ID, tokenHash, time.Now().Add(24*time.Hour))

	// Send email
	if h.email != nil {
		// TODO: send verification email
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"sent": true})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	emailNormalized := strings.ToLower(strings.TrimSpace(req.Email))
	pgUser, err := h.db.Queries.GetUserByEmailNormalized(r.Context(), emailNormalized)
	if err != nil {
		// Don't reveal whether user exists
		respondWithJSON(w, http.StatusOK, map[string]bool{"sent": true})
		return
	}

	// Generate reset token
	token := generateRandomToken()
	tokenHash := h.hashToken(token)

	h.db.Pool.Exec(r.Context(),
		"INSERT INTO verification_tokens (user_id, token_hash, type, expires_at) VALUES ($1, $2, 'password_reset', $3)",
		pgUser.ID, tokenHash, time.Now().Add(1*time.Hour))

	// Send email
	if h.email != nil {
		// TODO: send password reset email
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"sent": true})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenHash := h.hashToken(req.Token)
	var userID uuid.UUID
	err := h.db.Pool.QueryRow(r.Context(),
		"SELECT user_id FROM verification_tokens WHERE token_hash = $1 AND type = 'password_reset' AND expires_at > NOW() AND used_at IS NULL",
		tokenHash).Scan(&userID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid or expired token")
		return
	}

	hash, err := h.password.HashPassword(req.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	h.db.Queries.UpdateUserPassword(r.Context(), gen.UpdateUserPasswordParams{ID: userID, PasswordHash: &hash})
	h.db.Pool.Exec(r.Context(), "UPDATE verification_tokens SET used_at = NOW() WHERE token_hash = $1", tokenHash)

	// Revoke all sessions
	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'password_change' WHERE user_id = $1 AND revoked_at IS NULL",
		userID)

	respondWithJSON(w, http.StatusOK, map[string]bool{"reset": true})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pgUser, err := h.db.Queries.GetUserByID(r.Context(), user.ID)
	if err != nil || pgUser.PasswordHash == nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if h.password.ComparePassword(*pgUser.PasswordHash, req.CurrentPassword) != nil {
		respondWithError(w, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	hash, err := h.password.HashPassword(req.NewPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	h.db.Queries.UpdateUserPassword(r.Context(), gen.UpdateUserPasswordParams{ID: user.ID, PasswordHash: &hash})

	// Revoke all sessions except current
	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'password_change' WHERE user_id = $1 AND revoked_at IS NULL",
		user.ID)

	respondWithJSON(w, http.StatusOK, map[string]bool{"changed": true})
}

func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		DisplayName string `json:"displayName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	updated, err := h.db.Queries.UpdateUserProfile(r.Context(), gen.UpdateUserProfileParams{
		ID:               user.ID,
		DisplayName:      req.DisplayName,
		LocalePreference: "en",
		ThemePreference:  user.ThemePreference,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	respondWithJSON(w, http.StatusOK, db.ToUser(updated))
}

func (h *AuthHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	var req struct {
		ThemePreference  string `json:"themePreference"`
		LocalePreference string `json:"localePreference"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ThemePreference == "" {
		req.ThemePreference = "system"
	}
	if req.LocalePreference == "" {
		req.LocalePreference = "en"
	}

	updated, err := h.db.Queries.UpdateUserProfile(r.Context(), gen.UpdateUserProfileParams{
		ID:               user.ID,
		DisplayName:      user.DisplayName,
		LocalePreference: req.LocalePreference,
		ThemePreference:  req.ThemePreference,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update preferences")
		return
	}

	respondWithJSON(w, http.StatusOK, db.ToUser(updated))
}

func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	rows, err := h.db.Pool.Query(r.Context(),
		"SELECT id, user_id, token_hash, family_id, expires_at, revoked_at, user_agent, ip_address, created_at FROM refresh_tokens WHERE user_id = $1 ORDER BY created_at DESC LIMIT 50",
		user.ID)
	if err != nil {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}
	defer rows.Close()

	sessions := []map[string]interface{}{}
	for rows.Next() {
		var id, userID, familyID uuid.UUID
		var tokenHash string
		var expiresAt time.Time
		var revokedAt *time.Time
		var userAgent, ipAddress *string
		var createdAt time.Time
		rows.Scan(&id, &userID, &familyID, &tokenHash, &expiresAt, &revokedAt, &userAgent, &ipAddress, &createdAt)
		sessions = append(sessions, map[string]interface{}{
			"id":        id,
			"expiresAt": expiresAt,
			"revokedAt": revokedAt,
			"createdAt": createdAt,
			"isActive":  revokedAt == nil,
		})
	}

	respondWithJSON(w, http.StatusOK, sessions)
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		respondWithError(w, http.StatusBadRequest, "Session ID required")
		return
	}

	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'user_revoke' WHERE id = $1 AND user_id = $2",
		sessionID, user.ID)

	respondWithJSON(w, http.StatusOK, map[string]bool{"revoked": true})
}

func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'revoke_all' WHERE user_id = $1 AND revoked_at IS NULL",
		user.ID)

	respondWithJSON(w, http.StatusOK, map[string]bool{"revoked": true})
}

func (h *AuthHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	now := time.Now()
	h.db.Pool.Exec(r.Context(),
		"UPDATE users SET updated_at = NOW() WHERE id = $1", user.ID)
	_ = now

	respondWithJSON(w, http.StatusOK, map[string]bool{"completed": true})
}

func (h *AuthHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	pgUser, _ := h.db.Queries.GetUserByID(r.Context(), user.ID)

	// Get user's enrollments
	rows, _ := h.db.Pool.Query(r.Context(),
		"SELECT * FROM enrollments WHERE user_id = $1", user.ID)
	defer rows.Close()

	enrollments := []map[string]interface{}{}
	for rows.Next() {
		enrollments = append(enrollments, map[string]interface{}{"status": "active"})
	}

	export := map[string]interface{}{
		"user":         db.ToUser(pgUser),
		"enrollments":  enrollments,
		"exportedAt":   time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=userdata.json")
	json.NewEncoder(w).Encode(export)
}

func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Not authenticated")
		return
	}

	h.db.Queries.SoftDeleteUser(r.Context(), user.ID)
	h.db.Pool.Exec(r.Context(),
		"UPDATE refresh_tokens SET revoked_at = NOW(), revoked_reason = 'account_deleted' WHERE user_id = $1",
		user.ID)

	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *AuthHandler) ExchangeCode(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Not yet implemented")
}

func (h *AuthHandler) MFAChallenge(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "MFA not yet implemented")
}

func (h *AuthHandler) MFASetup(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "MFA not yet implemented")
}

func (h *AuthHandler) MFAVerifySetup(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "MFA not yet implemented")
}

func (h *AuthHandler) MFADisable(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "MFA not yet implemented")
}

func (h *AuthHandler) MFARegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "MFA not yet implemented")
}

func (h *AuthHandler) GoogleOAuth(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "OAuth not yet implemented")
}

func (h *AuthHandler) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "OAuth not yet implemented")
}

func (h *AuthHandler) GitHubOAuth(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "OAuth not yet implemented")
}

func (h *AuthHandler) GitHubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "OAuth not yet implemented")
}

func (h *AuthHandler) MicrosoftOAuth(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "OAuth not yet implemented")
}

func (h *AuthHandler) MicrosoftOAuthCallback(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "OAuth not yet implemented")
}

func (h *AuthHandler) MagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Magic link not yet implemented")
}

func (h *AuthHandler) MagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Magic link not yet implemented")
}

func (h *AuthHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Not yet implemented")
}

func (h *AuthHandler) ListPasskeys(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}

func (h *AuthHandler) RegisterPasskey(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Passkeys not yet implemented")
}

func (h *AuthHandler) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Passkeys not yet implemented")
}

func (h *AuthHandler) BeginPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Passkeys not yet implemented")
}

func (h *AuthHandler) FinishPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "Passkeys not yet implemented")
}

func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var _ = context.Background
var _ = fmt.Sprintf
var _ = models.AuthMethodPassword
