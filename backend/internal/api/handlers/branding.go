package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/configstore"
	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
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
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		// Return default platform branding
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"appName": "MyCourses",
		})
		return
	}

	bc, err := h.db.Queries.GetBrandingConfigByTenant(r.Context(), tenantID)
	if err != nil {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"appName": "MyCourses"})
		return
	}

	var config map[string]interface{}
	json.Unmarshal(bc.Config, &config)
	respondWithJSON(w, http.StatusOK, config)
}

func (h *BrandingHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithError(w, http.StatusForbidden, "Tenant required")
		return
	}

	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	configJSON, _ := json.Marshal(config)
	bc, err := h.db.Queries.UpsertBrandingConfig(r.Context(), gen.UpsertBrandingConfigParams{
		TenantID: tenantID,
		Config:   configJSON,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update branding")
		return
	}

	respondWithJSON(w, http.StatusOK, bc)
}

func (h *BrandingHandler) ServeAsset(w http.ResponseWriter, r *http.Request) {
	assetID := parseUUID(mux.Vars(r)["key"])
	if assetID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid asset ID")
		return
	}

	asset, err := h.db.Queries.GetBrandingAssetByID(r.Context(), *assetID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Asset not found")
		return
	}

	w.Header().Set("Content-Type", asset.MimeType)
	w.Write(asset.Data)
}

func (h *BrandingHandler) ServeMedia(w http.ResponseWriter, r *http.Request) {
	h.ServeAsset(w, r)
}

func (h *BrandingHandler) GetPublicPage(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	slug := mux.Vars(r)["slug"]

	if tenantID == uuidNil() {
		respondWithError(w, http.StatusNotFound, "Page not found")
		return
	}

	page, err := h.db.Queries.GetCustomPageBySlug(r.Context(), gen.GetCustomPageBySlugParams{
		TenantID: tenantID,
		Slug:     slug,
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Page not found")
		return
	}

	respondWithJSON(w, http.StatusOK, page)
}

func (h *BrandingHandler) ListPublicPages(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	pages, err := h.db.Queries.ListCustomPagesByTenant(r.Context(), tenantID)
	if err != nil {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	respondWithJSON(w, http.StatusOK, pages)
}

func (h *BrandingHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithError(w, http.StatusForbidden, "Tenant required")
		return
	}

	r.ParseMultipartForm(10 << 20) // 10MB max
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "File required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	assetType := "custom"
	if r.FormValue("type") != "" {
		assetType = r.FormValue("type")
	}

	asset, err := h.db.Queries.CreateBrandingAsset(r.Context(), gen.CreateBrandingAssetParams{
		TenantID:  tenantID,
		Name:      header.Filename,
		Type:      assetType,
		MimeType:  header.Header.Get("Content-Type"),
		SizeBytes: int64(len(data)),
		Data:      data,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload asset")
		return
	}

	respondWithJSON(w, http.StatusCreated, asset)
}

func (h *BrandingHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	assetID := parseUUID(mux.Vars(r)["key"])
	if assetID == nil || tenantID == uuidNil() {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.db.Queries.DeleteBrandingAsset(r.Context(), gen.DeleteBrandingAssetParams{
		ID:       *assetID,
		TenantID: tenantID,
	})

	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *BrandingHandler) ListMedia(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	assets, _ := h.db.Queries.ListBrandingAssetsByTenant(r.Context(), tenantID)
	respondWithJSON(w, http.StatusOK, assets)
}

func (h *BrandingHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	h.UploadAsset(w, r)
}

func (h *BrandingHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	h.DeleteAsset(w, r)
}

func (h *BrandingHandler) AdminListPages(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithJSON(w, http.StatusOK, []interface{}{})
		return
	}

	pages, _ := h.db.Queries.ListAllCustomPages(r.Context(), tenantID)
	respondWithJSON(w, http.StatusOK, pages)
}

func (h *BrandingHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithError(w, http.StatusForbidden, "Tenant required")
		return
	}

	var req struct {
		Slug            string `json:"slug"`
		Title           string `json:"title"`
		HtmlBody        string `json:"htmlBody"`
		MetaDescription string `json:"metaDescription"`
		IsPublished     bool   `json:"isPublished"`
		SortOrder       int    `json:"sortOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var metaDesc *string
	if req.MetaDescription != "" {
		metaDesc = &req.MetaDescription
	}

	page, err := h.db.Queries.CreateCustomPage(r.Context(), gen.CreateCustomPageParams{
		TenantID:        tenantID,
		Slug:            req.Slug,
		Title:           req.Title,
		HtmlBody:        req.HtmlBody,
		MetaDescription: metaDesc,
		IsPublished:     req.IsPublished,
		SortOrder:       int32(req.SortOrder),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create page")
		return
	}

	respondWithJSON(w, http.StatusCreated, page)
}

func (h *BrandingHandler) UpdatePage(w http.ResponseWriter, r *http.Request) {
	pageID := parseUUID(mux.Vars(r)["id"])
	if pageID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid page ID")
		return
	}

	var req struct {
		Slug            string `json:"slug"`
		Title           string `json:"title"`
		HtmlBody        string `json:"htmlBody"`
		MetaDescription string `json:"metaDescription"`
		IsPublished     bool   `json:"isPublished"`
		SortOrder       int    `json:"sortOrder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var metaDesc *string
	if req.MetaDescription != "" {
		metaDesc = &req.MetaDescription
	}

	page, err := h.db.Queries.UpdateCustomPage(r.Context(), gen.UpdateCustomPageParams{
		ID:              *pageID,
		Slug:            req.Slug,
		Title:           req.Title,
		HtmlBody:        req.HtmlBody,
		MetaDescription: metaDesc,
		IsPublished:     req.IsPublished,
		SortOrder:       int32(req.SortOrder),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update page")
		return
	}

	respondWithJSON(w, http.StatusOK, page)
}

func (h *BrandingHandler) DeletePage(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	pageID := parseUUID(mux.Vars(r)["id"])
	if pageID == nil || tenantID == uuidNil() {
		respondWithError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	h.db.Queries.DeleteCustomPage(r.Context(), gen.DeleteCustomPageParams{
		ID:       *pageID,
		TenantID: tenantID,
	})

	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
