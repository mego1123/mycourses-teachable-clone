package handlers

import (
	"net/http"

	"mycourses/internal/health"
)

type HealthHandler struct {
	service *health.Service
}

func (h *HealthHandler) SetEmailService(svc interface{}) {}
func NewHealthHandler(svc *health.Service) *HealthHandler {
	return &HealthHandler{service: svc}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (h *HealthHandler) GetNodes(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *HealthHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *HealthHandler) GetAggregateMetrics(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *HealthHandler) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{})
}
func (h *HealthHandler) GetIntegrationStatus(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *HealthHandler) GetIntegrationCounts24h(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]int64{"stripeCalls": 0, "resendEmails": 0})
}

func (h *HealthHandler) ListNodes(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}

func (h *HealthHandler) GetCurrent(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, map[string]interface{}{}) }
func (h *HealthHandler) GetIntegrations(w http.ResponseWriter, r *http.Request) { respondWithJSON(w, http.StatusOK, []interface{}{}) }
func (h *HealthHandler) SendTestEmail(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented) }
