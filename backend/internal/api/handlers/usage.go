package handlers

import (
	"net/http"

	"mycourses/internal/db"
)

type UsageHandler struct{ db *db.DB }

func NewUsageHandler(database *db.DB) *UsageHandler { return &UsageHandler{db: database} }

func (h *UsageHandler) RecordUsage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *UsageHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
