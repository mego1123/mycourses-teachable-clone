package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/syslog"

	"time"
	gen "mycourses/internal/db/gen"
)

type LogHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewLogHandler(database *db.DB, sysLogger *syslog.Logger) *LogHandler {
	return &LogHandler{db: database, syslog: sysLogger}
}

func (h *LogHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	category := r.URL.Query().Get("category")
	page, limit := parsePagination(r)

	logs, _ := h.db.Queries.ListSystemLogs(r.Context(), gen.ListSystemLogsParams{
		Column1: level, Column2: category, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	respondWithJSON(w, http.StatusOK, logs)
}

func (h *LogHandler) SeverityCounts(w http.ResponseWriter, r *http.Request) {
	counts, _ := h.db.Queries.GetSeverityCounts(r.Context(), time.Now().Add(-24*time.Hour))
	respondWithJSON(w, http.StatusOK, counts)
}

func (h *LogHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, http.StatusNotImplemented, "CSV export not yet implemented")
}
