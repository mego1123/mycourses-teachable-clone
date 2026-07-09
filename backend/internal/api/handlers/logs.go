package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type LogHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewLogHandler(database *db.DB, sysLogger *syslog.Logger) *LogHandler {
	return &LogHandler{db: database, syslog: sysLogger}
}

func (h *LogHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}

func (h *LogHandler) SeverityCounts(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]int64{})
}
func (h *LogHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
