package handlers

import (
	"net/http"

	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type BundlesHandler struct {
	db     *db.DB
}

type bundleRequest struct {
	db     *db.DB
}

func NewBundlesHandler(database *db.DB, sysLogger *syslog.Logger) *BundlesHandler {
	return &BundlesHandler{db: database}
}

func (h *BundlesHandler) ListBundles(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *BundlesHandler) CreateBundle(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *BundlesHandler) UpdateBundle(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *BundlesHandler) DeleteBundle(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *BundlesHandler) ListBundlesPublic(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

