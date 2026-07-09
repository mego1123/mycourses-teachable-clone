package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/email"
	"mycourses/internal/events"
	"mycourses/internal/syslog"
)

type TenantHandler struct {
	db     *db.DB
	email  *email.ResendService
	emitter events.Emitter
	syslog *syslog.Logger
}

func NewTenantHandler(database *db.DB, emailSvc *email.ResendService, emitter events.Emitter, sysLogger *syslog.Logger) *TenantHandler {
	return &TenantHandler{db: database, email: emailSvc, emitter: emitter, syslog: sysLogger}
}

func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *TenantHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *TenantHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *TenantHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *TenantHandler) ChangeMemberRole(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *TenantHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}

func (h *TenantHandler) GetActivity(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, []interface{}{}) }
func (h *TenantHandler) UpdateTenantSettings(w http.ResponseWriter, r *http.Request) {}
func (h *TenantHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {}
func (h *TenantHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {}
