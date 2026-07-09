package handlers

import (
	"net/http"
	"time"

	"mycourses/internal/auth"
	"mycourses/internal/db"
	"mycourses/internal/email"
	"mycourses/internal/events"
	"mycourses/internal/middleware"
	"mycourses/internal/syslog"
	"mycourses/internal/telemetry"
)

// AuthHandler handles authentication operations.
// NOTE: During MongoDB→Postgres migration, auth handler methods are stubs.
// They need to be rewritten to use Postgres queries.
type AuthHandler struct {
	db          *db.DB
	jwt         *auth.JWTService
	password    *auth.PasswordService
	googleOAuth *auth.GoogleOAuthService
	githubOAuth *auth.GitHubOAuthService
	microsoftOAuth *auth.MicrosoftOAuthService
	email       *email.ResendService
	emitter     events.Emitter
	syslog      *syslog.Logger
	frontendURL string
	getConfig   func(string) string
	rateLimiter *middleware.RateLimiter
	telemetrySvc *telemetry.Service
	totpEncKey  []byte
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

func (h *AuthHandler) GetProviders(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) SetupMFA(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MFAChallenge(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MicrosoftCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Auth handler not yet migrated to Postgres"}`, http.StatusNotImplemented)
}


func (h *AuthHandler) ExchangeCode(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GoogleOAuth(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GitHubOAuth(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) GitHubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MicrosoftOAuth(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MicrosoftOAuthCallback(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}


func (h *AuthHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MFASetup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) MFAVerifySetup(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) ListPasskeys(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *AuthHandler) RegisterPasskey(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) DeletePasskey(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) BeginPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AuthHandler) FinishPasskeyLogin(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}

func (h *AuthHandler) MFADisable(w http.ResponseWriter, r *http.Request) {}
func (h *AuthHandler) MFARegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {}
func (h *AuthHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {}
var _ = time.Now
