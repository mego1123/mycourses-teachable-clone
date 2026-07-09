package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/email"
	"mycourses/internal/events"
	"mycourses/internal/middleware"
	"mycourses/internal/models"
	"mycourses/internal/syslog"

	"crypto/sha256"
	"encoding/base64"
	"time"
)

type TenantHandler struct {
	db     *db.DB
	email  *email.ResendService
	emitter events.Emitter
	syslog *syslog.Logger
	stripeSvc interface{}
}

func NewTenantHandler(database *db.DB, emailSvc *email.ResendService, emitter events.Emitter, sysLogger *syslog.Logger) *TenantHandler {
	return &TenantHandler{db: database, email: emailSvc, emitter: emitter, syslog: sysLogger}
}

func (h *TenantHandler) SetStripe(svc interface{}) { h.stripeSvc = svc }

func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusForbidden, "No tenant context")
		return
	}
	respondWithJSON(w, http.StatusOK, tenant)
}

func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusForbidden, "No tenant context")
		return
	}

	var req struct{ Name string `json:"name"` }
	json.NewDecoder(r.Body).Decode(&req)

	if req.Name != "" {
		h.db.Pool.Exec(r.Context(), "UPDATE tenants SET name = $2, updated_at = NOW() WHERE id = $1", tenant.ID, req.Name)
	}

	updated, _ := h.db.Queries.GetTenantByID(r.Context(), tenant.ID)
	respondWithJSON(w, http.StatusOK, db.ToTenant(updated))
}

func (h *TenantHandler) UpdateTenantSettings(w http.ResponseWriter, r *http.Request) {
	h.UpdateTenant(w, r)
}

func (h *TenantHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusForbidden, "No tenant context")
		return
	}

	members, err := h.db.Queries.ListMembershipsByTenant(r.Context(), gen.ListMembershipsByTenantParams{
		TenantID: tenant.ID, Limit: 100, Offset: 0,
	})
	if err != nil {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	result := make([]map[string]interface{}, len(members))
	for i, m := range members {
		user, _ := h.db.Queries.GetUserByID(r.Context(), m.UserID)
		result[i] = map[string]interface{}{
			"userId":   m.UserID,
			"role":     m.Role,
			"joinedAt": m.CreatedAt,
			"user":     db.ToUser(user),
		}
	}

	respondWithJSON(w, http.StatusOK, result)
}

func (h *TenantHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	user, ok2 := middleware.GetUserFromContext(r.Context())
	if !ok || !ok2 {
		respondWithError(w, http.StatusForbidden, "Authentication required")
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Role == "" { req.Role = "user" }

	token := generateRandomToken()
	tokenHash := hashTokenSHA256(token)

	_, err := h.db.Pool.Exec(r.Context(),
		`INSERT INTO invitations (email, tenant_id, invited_by, role, token_hash, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		req.Email, tenant.ID, user.ID, req.Role, tokenHash, time.Now().Add(7*24*time.Hour))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create invitation")
		return
	}

	respondWithJSON(w, http.StatusCreated, map[string]interface{}{
		"email":      req.Email,
		"role":       req.Role,
		"expiresAt":  time.Now().Add(7 * 24 * time.Hour),
	})
}

func (h *TenantHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusForbidden, "No tenant context")
		return
	}

	userID := parseUUID(mux.Vars(r)["userId"])
	if userID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	h.db.Queries.DeleteMembership(r.Context(), gen.DeleteMembershipParams{
		TenantID: tenant.ID, UserID: *userID,
	})

	respondWithJSON(w, http.StatusOK, map[string]bool{"removed": true})
}

func (h *TenantHandler) ChangeRole(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		respondWithError(w, http.StatusForbidden, "No tenant context")
		return
	}

	userID := parseUUID(mux.Vars(r)["userId"])
	if userID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req struct{ Role string `json:"role"` }
	json.NewDecoder(r.Body).Decode(&req)

	h.db.Queries.UpdateMembershipRole(r.Context(), gen.UpdateMembershipRoleParams{
		TenantID: tenant.ID, UserID: *userID, Role: req.Role,
	})

	respondWithJSON(w, http.StatusOK, map[string]bool{"updated": true})
}

func (h *TenantHandler) ChangeMemberRole(w http.ResponseWriter, r *http.Request) { h.ChangeRole(w, r) }

func (h *TenantHandler) TransferOwnership(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	user, ok2 := middleware.GetUserFromContext(r.Context())
	if !ok || !ok2 {
		respondWithError(w, http.StatusForbidden, "Authentication required")
		return
	}

	newOwnerID := parseUUID(mux.Vars(r)["userId"])
	if newOwnerID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Change current owner to admin
	h.db.Queries.UpdateMembershipRole(r.Context(), gen.UpdateMembershipRoleParams{
		TenantID: tenant.ID, UserID: user.ID, Role: "admin",
	})
	// Change new owner to owner
	h.db.Queries.UpdateMembershipRole(r.Context(), gen.UpdateMembershipRoleParams{
		TenantID: tenant.ID, UserID: *newOwnerID, Role: "owner",
	})

	respondWithJSON(w, http.StatusOK, map[string]bool{"transferred": true})
}

func (h *TenantHandler) CancelInvitation(w http.ResponseWriter, r *http.Request) {
	invitationID := parseUUID(mux.Vars(r)["invitationId"])
	if invitationID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid invitation ID")
		return
	}

	h.db.Pool.Exec(r.Context(), "UPDATE invitations SET status = 'revoked' WHERE id = $1", invitationID)
	respondWithJSON(w, http.StatusOK, map[string]bool{"cancelled": true})
}

func (h *TenantHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	tenant, ok := middleware.GetTenantFromContext(r.Context())
	if !ok {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	rows, _ := h.db.Pool.Query(r.Context(),
		"SELECT * FROM audit_log WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 50", tenant.ID)
	defer rows.Close()

	result := []map[string]interface{}{}
	for rows.Next() {
		result = append(result, map[string]interface{}{"status": "ok"})
	}

	respondWithJSON(w, http.StatusOK, result)
}

func hashTokenSHA256(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}

var _ = models.RoleAdmin
