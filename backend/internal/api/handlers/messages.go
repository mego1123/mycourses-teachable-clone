package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type MessageHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewMessageHandler(database *db.DB, sysLogger *syslog.Logger) *MessageHandler {
	return &MessageHandler{db: database, syslog: sysLogger}
}
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented) }
func (h *MessageHandler) UnreadCount(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented) }
func (h *MessageHandler) MarkRead(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented) }
