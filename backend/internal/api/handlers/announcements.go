package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
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
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithJSON(w, http.StatusOK, []interface{}{}); return }

	page, limit := parsePagination(r)
	anns, _ := h.db.Queries.ListPublishedAnnouncements(r.Context(), gen.ListPublishedAnnouncementsParams{
		TenantID: &tenantID, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	respondWithJSON(w, http.StatusOK, anns)
}

func (h *AnnouncementsHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() { respondWithJSON(w, http.StatusOK, []interface{}{}); return }

	page, limit := parsePagination(r)
	anns, _ := h.db.Queries.ListAllAnnouncements(r.Context(), gen.ListAllAnnouncementsParams{
		TenantID: &tenantID, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	respondWithJSON(w, http.StatusOK, anns)
}

func (h *AnnouncementsHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	user, ok := middleware.GetUserFromContext(r.Context())
	if tenantID == uuidNil() || !ok {
		respondWithError(w, http.StatusForbidden, "Auth + tenant required"); return
	}

	var req struct {
		Title       string `json:"title"`
		Body        string `json:"body"`
		IsPublished bool   `json:"isPublished"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid body"); return
	}
	if req.Title == "" { respondWithError(w, http.StatusBadRequest, "Title required"); return
	}

	ann, err := h.db.Queries.CreateAnnouncement(r.Context(), gen.CreateAnnouncementParams{
		TenantID: &tenantID, Title: req.Title, Body: req.Body,
		IsPublished: req.IsPublished, CreatedBy: &user.ID,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusCreated, ann)
}

func (h *AnnouncementsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	var req struct {
		Title       string `json:"title"`
		Body        string `json:"body"`
		IsPublished bool   `json:"isPublished"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	ann, err := h.db.Queries.UpdateAnnouncement(r.Context(), gen.UpdateAnnouncementParams{
		ID: *id, Title: req.Title, Body: req.Body, IsPublished: req.IsPublished,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	respondWithJSON(w, http.StatusOK, ann)
}

func (h *AnnouncementsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := parseUUID(mux.Vars(r)["id"])
	if id == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	h.db.Queries.DeleteAnnouncement(r.Context(), *id)
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
