package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"mycourses/internal/auth"
	"mycourses/internal/db"
	"mycourses/internal/email"
	"mycourses/internal/events"
	"mycourses/internal/health"
	"mycourses/internal/models"
	"mycourses/internal/syslog"
)

type AdminHandler struct {
	db          *db.DB
	emitter     events.Emitter
	syslog      *syslog.Logger
	healthSvc   *health.Service
	jwtSvc      *auth.JWTService
	emailSvc    *email.ResendService
	getConfig   func(string) string
}

func NewAdminHandler(database *db.DB, emitter events.Emitter, sysLogger *syslog.Logger) *AdminHandler {
	return &AdminHandler{db: database, emitter: emitter, syslog: sysLogger}
}

func (h *AdminHandler) SetHealthService(svc *health.Service, getConfig func(string) string) {
	h.healthSvc = svc
	h.getConfig = getConfig
}
func (h *AdminHandler) SetJWTService(svc *auth.JWTService) { h.jwtSvc = svc }
func (h *AdminHandler) SetEmailService(svc *email.ResendService) { h.emailSvc = svc }

func (h *AdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"tenants": []interface{}{}, "total": 0, "page": 1, "limit": 25})
}
func (h *AdminHandler) ExportTenantsCSV(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) UpdateTenantStatus(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"users": []interface{}{}, "total": 0, "page": 1, "limit": 25})
}
func (h *AdminHandler) ExportUsersCSV(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) PreflightDeleteUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) GetAbout(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"version": "1.0.0"})
}
func (h *AdminHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *AdminHandler) ListRootMembers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *AdminHandler) InviteRootMember(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) RemoveRootMember(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) ChangeRootMemberRole(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) CancelRootInvitation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *AdminHandler) isRootTenantOwner(ctx context.Context, userID uuid.UUID) bool { return false }
func (h *AdminHandler) getRootTenant(ctx context.Context) (*models.Tenant, error) { return nil, nil }
