package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

type EnrollmentHandler struct{ db *db.DB }

func NewEnrollmentHandler(database *db.DB) *EnrollmentHandler { return &EnrollmentHandler{db: database} }

// CreateEnrollment creates an enrollment (called after Stripe checkout completes, or for free courses).
func (h *EnrollmentHandler) CreateEnrollment(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }

	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Authentication required"); return }

	// Check if already enrolled
	existing, _ := h.db.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
		CourseID: *courseID, UserID: *userID,
	})
	if existing.ID != uuidNil() {
		respondWithJSON(w, http.StatusOK, existing) // idempotent — already enrolled
		return
	}

	course, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil { respondWithError(w, http.StatusNotFound, "Course not found"); return }

	enrollment, err := h.db.Queries.CreateEnrollment(r.Context(), gen.CreateEnrollmentParams{
		CourseID: *courseID, TenantID: course.TenantID, UserID: *userID,
		Status: "active", PricePaidCents: course.PriceCents, Currency: course.Currency,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed to enroll: "+err.Error()); return }

	// Record financial transaction
	h.db.Queries.CreateTransaction(r.Context(), gen.CreateTransactionParams{
		TenantID: &course.TenantID, UserID: userID, Type: "course_purchase",
		AmountCents: course.PriceCents, Currency: course.Currency,
		Description: "Course purchase: " + course.Title,
	})

	respondWithJSON(w, http.StatusCreated, enrollment)
}

// ListMine lists the authenticated learner's enrollments.
func (h *EnrollmentHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Authentication required"); return }

	page, limit := parsePagination(r)
	enrollments, err := h.db.Queries.ListEnrollmentsByUser(r.Context(), gen.ListEnrollmentsByUserParams{
		UserID: *userID, Column2: []string{"active", "completed"},
		Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"enrollments": enrollments, "page": page, "limit": limit})
}

// ListByCreator lists enrollments (sales) for the creator's tenant.
func (h *EnrollmentHandler) ListByCreator(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant context required"); return }

	page, limit := parsePagination(r)
	status := r.URL.Query().Get("status")
	enrollments, err := h.db.Queries.ListEnrollmentsByTenant(r.Context(), gen.ListEnrollmentsByTenantParams{
		TenantID: *tenantID, Column2: status, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"enrollments": enrollments, "page": page, "limit": limit})
}
