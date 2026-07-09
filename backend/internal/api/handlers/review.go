package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

type ReviewHandler struct{ db *db.DB }

func NewReviewHandler(database *db.DB) *ReviewHandler { return &ReviewHandler{db: database} }

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }
	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Auth required"); return }

	var req struct{ Rating int; Comment string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
	if req.Rating < 1 || req.Rating > 5 { respondWithError(w, http.StatusBadRequest, "Rating 1-5 required"); return }

	course, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil { respondWithError(w, http.StatusNotFound, "Course not found"); return }

	// Verify enrollment
	enrollment, _ := h.db.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
		CourseID: *courseID, UserID: *userID,
	})
	if enrollment.ID == uuidNil() { respondWithError(w, http.StatusForbidden, "Must be enrolled to review"); return }

	review, err := h.db.Queries.CreateReview(r.Context(), gen.CreateReviewParams{
		CourseID: *courseID, UserID: *userID, TenantID: course.TenantID,
		Rating: int32(req.Rating), Comment: req.Comment,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }
	respondWithJSON(w, http.StatusCreated, review)
}

func (h *ReviewHandler) ListPublic(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }
	page, limit := parsePagination(r)
	reviews, err := h.db.Queries.ListPublicReviewsByCourse(r.Context(), gen.ListPublicReviewsByCourseParams{
		CourseID: *courseID, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	avg, _ := h.db.Queries.GetAverageRatingByCourse(r.Context(), *courseID)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"reviews": reviews, "averageRating": avg.AvgRating, "reviewCount": avg.ReviewCount})
}

func (h *ReviewHandler) Hide(w http.ResponseWriter, r *http.Request) {
	reviewID := parseUUID(mux.Vars(r)["id"])
	if reviewID == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }
	if err := h.db.Queries.HideReview(r.Context(), *reviewID); err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]bool{"hidden": true})
}
