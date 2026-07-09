package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/syslog"
	"mycourses/internal/telemetry"
)

type PMHandler struct {
	db       *db.DB
	telemetry *telemetry.Service
}

func NewPMHandler(database *db.DB, telemetrySvc *telemetry.Service, sysLogger *syslog.Logger) *PMHandler {
	return &PMHandler{db: database, telemetry: telemetrySvc}
}

func (h *PMHandler) GetFunnel(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *PMHandler) GetCohorts(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *PMHandler) GetKPIs(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *PMHandler) GetEngagement(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *PMHandler) GetCustomEvents(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *PMHandler) GetEventTypeSummary(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *PMHandler) ListEventTypes(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []string{})
}
func (h *PMHandler) GetSankey(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}

func (h *PMHandler) GetRetention(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, []interface{}{}) }
