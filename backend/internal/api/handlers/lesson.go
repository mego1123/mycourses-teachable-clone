package handlers

import (
        "encoding/json"
        "net/http"

        "github.com/gorilla/mux"

        "mycourses/internal/db"
        gen "mycourses/internal/db/gen"
)

type LessonHandler struct{ db *db.DB }

func NewLessonHandler(database *db.DB) *LessonHandler { return &LessonHandler{db: database} }

func (h *LessonHandler) Create(w http.ResponseWriter, r *http.Request) {
        sectionID := parseUUID(mux.Vars(r)["sectionId"])
        if sectionID == nil { respondWithError(w, http.StatusBadRequest, "Invalid section ID"); return }

        var req struct{ Title, Type, Content string; SortOrder, DurationSec int; IsPreview bool }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
        if req.Title == "" { respondWithError(w, http.StatusBadRequest, "Title required"); return }
        if req.Type == "" { req.Type = "video" }

        // Look up the course ID from the section
        section, err := h.db.Queries.GetSectionByID(r.Context(), *sectionID)
        if err != nil { respondWithError(w, http.StatusNotFound, "Section not found"); return }

        lesson, err := h.db.Queries.CreateLesson(r.Context(), gen.CreateLessonParams{
                SectionID: *sectionID, CourseID: section.CourseID, Title: req.Title, Type: req.Type,
                Content: req.Content, SortOrder: int32(req.SortOrder),
                IsPreview: req.IsPreview, DurationSec: int32(req.DurationSec),
        })
        if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }
        respondWithJSON(w, http.StatusCreated, lesson)
}

func (h *LessonHandler) Update(w http.ResponseWriter, r *http.Request) {
        lessonID := parseUUID(mux.Vars(r)["id"])
        if lessonID == nil { respondWithError(w, http.StatusBadRequest, "Invalid lesson ID"); return }

        var req struct{ Title, Type, Content string; SortOrder, DurationSec int; IsPreview bool }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }

        lesson, err := h.db.Queries.UpdateLesson(r.Context(), gen.UpdateLessonParams{
                ID: *lessonID, Title: req.Title, Type: req.Type, Content: req.Content,
                SortOrder: int32(req.SortOrder), IsPreview: req.IsPreview, DurationSec: int32(req.DurationSec),
        })
        if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
        respondWithJSON(w, http.StatusOK, lesson)
}

func (h *LessonHandler) Delete(w http.ResponseWriter, r *http.Request) {
        lessonID := parseUUID(mux.Vars(r)["id"])
        if lessonID == nil { respondWithError(w, http.StatusBadRequest, "Invalid lesson ID"); return }
        if err := h.db.Queries.DeleteLesson(r.Context(), *lessonID); err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
        respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *LessonHandler) ListBySection(w http.ResponseWriter, r *http.Request) {
        sectionID := parseUUID(mux.Vars(r)["sectionId"])
        if sectionID == nil { respondWithError(w, http.StatusBadRequest, "Invalid section ID"); return }
        lessons, err := h.db.Queries.ListLessonsBySection(r.Context(), *sectionID)
        if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
        respondWithJSON(w, http.StatusOK, lessons)
}

func (h *LessonHandler) ListByCourse(w http.ResponseWriter, r *http.Request) {
        courseID := parseUUID(mux.Vars(r)["courseId"])
        if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }
        lessons, err := h.db.Queries.ListLessonsByCourse(r.Context(), *courseID)
        if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
        respondWithJSON(w, http.StatusOK, lessons)
}
