package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type AnnouncementsHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewAnnouncementsHandler(database *db.DB, sysLogger *syslog.Logger) *AnnouncementsHandler {
	return &AnnouncementsHandler{db: database, syslog: sysLogger}
}

func (h *AnnouncementsHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *AnnouncementsHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *AnnouncementsHandler) Create(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *AnnouncementsHandler) Update(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *AnnouncementsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}
