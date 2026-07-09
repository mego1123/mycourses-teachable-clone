package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"mycourses/internal/db"
	"mycourses/internal/middleware"
)

type UsageHandler struct{ db *db.DB }

func NewUsageHandler(database *db.DB) *UsageHandler { return &UsageHandler{db: database} }

func (h *UsageHandler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct {
		Type string `json:"type"`
		Amount int64 `json:"amount"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	h.db.Pool.Exec(r.Context(),
		"INSERT INTO usage_events (tenant_id, type, amount, metadata) VALUES ($1, $2, $3, $4)",
		tenantID, req.Type, req.Amount, "{}")

	respondWithJSON(w, http.StatusCreated, map[string]bool{"recorded": true})
}

func (h *UsageHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var total int64
	h.db.Pool.QueryRow(r.Context(),
		"SELECT COALESCE(SUM(amount), 0) FROM usage_events WHERE tenant_id = $1 AND created_at >= $2",
		tenantID, monthStart).Scan(&total)

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"total":       total,
		"periodStart": monthStart,
		"periodEnd":   now,
	})
}
