package handlers

import (
	"net/http"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type BrandingHandler struct {
	db            *db.DB
	store         *configstore.Store
	syslog        *syslog.Logger
	authProviders map[string]bool
}

func NewBrandingHandler(database *db.DB, store *configstore.Store, sysLogger *syslog.Logger) *BrandingHandler {
	return &BrandingHandler{db: database, store: store, syslog: sysLogger}
}

func (h *BrandingHandler) SetAuthProviders(providers map[string]bool) { h.authProviders = providers }

func (h *BrandingHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"appName": "MyCourses"})
}
func (h *BrandingHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) GetPublicPage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) ListPublicPages(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *BrandingHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) ListMedia(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *BrandingHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) AdminListPages(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, []interface{}{})
}
func (h *BrandingHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
func (h *BrandingHandler) DeletePage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated"}`, http.StatusNotImplemented)
}
