package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

type SectionHandler struct{ db *db.DB }

func NewSectionHandler(database *db.DB) *SectionHandler { return &SectionHandler{db: database} }

func (h *SectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }

	var req struct{ Title, Description string; SortOrder, DripOffsetDays int }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
	if req.Title == "" { respondWithError(w, http.StatusBadRequest, "Title required"); return }

	section, err := h.db.Queries.CreateSection(r.Context(), gen.CreateSectionParams{
		CourseID: *courseID, Title: req.Title, Description: req.Description,
		SortOrder: int32(req.SortOrder), DripOffsetDays: int32(req.DripOffsetDays),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }
	respondWithJSON(w, http.StatusCreated, section)
}

func (h *SectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	sectionID := parseUUID(mux.Vars(r)["id"])
	if sectionID == nil { respondWithError(w, http.StatusBadRequest, "Invalid section ID"); return }

	var req struct{ Title, Description string; SortOrder, DripOffsetDays int }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }

	section, err := h.db.Queries.UpdateSection(r.Context(), gen.UpdateSectionParams{
		ID: *sectionID, Title: req.Title, Description: req.Description,
		SortOrder: int32(req.SortOrder), DripOffsetDays: int32(req.DripOffsetDays),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, section)
}

func (h *SectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sectionID := parseUUID(mux.Vars(r)["id"])
	if sectionID == nil { respondWithError(w, http.StatusBadRequest, "Invalid section ID"); return }
	if err := h.db.Queries.DeleteSection(r.Context(), *sectionID); err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *SectionHandler) ListByCourse(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }
	sections, err := h.db.Queries.ListSectionsByCourse(r.Context(), *courseID)
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, sections)
}
