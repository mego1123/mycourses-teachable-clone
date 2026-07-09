// Package handlers — Course CRUD handler.
// Uses the Postgres DB layer (sqlc-generated queries).
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/google/uuid"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

// CourseHandler handles course CRUD for creators and public storefront.
type CourseHandler struct {
	db *db.DB
}

func NewCourseHandler(database *db.DB) *CourseHandler {
	return &CourseHandler{db: database}
}

// Create creates a new draft course (creator only).
func (h *CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		Slug         string `json:"slug"`
		PriceCents   int64  `json:"priceCents"`
		Currency     string `json:"currency"`
		Category     string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Title == "" || req.Slug == "" {
		respondWithError(w, http.StatusBadRequest, "Title and slug are required")
		return
	}
	if req.Currency == "" {
		req.Currency = "usd"
	}

	var category *string
	if req.Category != "" {
		category = &req.Category
	}

	course, err := h.db.Queries.CreateCourse(r.Context(), gen.CreateCourseParams{
		TenantID:    *tenantID,
		Title:       req.Title,
		Description: req.Description,
		Slug:        req.Slug,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Status:      "draft",
		Category:    category,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create course: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, course)
}

// Get returns a single course by ID (creator only — any status).
func (h *CourseHandler) Get(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["id"])
	if courseID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	course, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	// Verify tenant isolation
	tenantID := getTenantIDFromContext(r)
	if tenantID != nil && course.TenantID != *tenantID {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	respondWithJSON(w, http.StatusOK, course)
}

// Update updates a course (creator only).
func (h *CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["id"])
	if courseID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	// Verify ownership
	existing, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}
	tenantID := getTenantIDFromContext(r)
	if tenantID != nil && existing.TenantID != *tenantID {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		Slug         string `json:"slug"`
		PriceCents   int64  `json:"priceCents"`
		Currency     string `json:"currency"`
		Category     string `json:"category"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var category *string
	if req.Category != "" {
		category = &req.Category
	}

	course, err := h.db.Queries.UpdateCourse(r.Context(), gen.UpdateCourseParams{
		ID:          *courseID,
		Title:       req.Title,
		Description: req.Description,
		Slug:        req.Slug,
		PriceCents:  req.PriceCents,
		Currency:    req.Currency,
		Category:    category,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update course: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, course)
}

// Publish publishes a course.
func (h *CourseHandler) Publish(w http.ResponseWriter, r *http.Request) {
	h.setCourseStatus(w, r, "publish")
}

// Unpublish reverts a course to draft.
func (h *CourseHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	h.setCourseStatus(w, r, "unpublish")
}

func (h *CourseHandler) setCourseStatus(w http.ResponseWriter, r *http.Request, action string) {
	courseID := parseUUID(mux.Vars(r)["id"])
	if courseID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	// Verify ownership
	existing, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}
	tenantID := getTenantIDFromContext(r)
	if tenantID != nil && existing.TenantID != *tenantID {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	var course gen.Course
	if action == "publish" {
		course, err = h.db.Queries.PublishCourse(r.Context(), *courseID)
	} else {
		course, err = h.db.Queries.UnpublishCourse(r.Context(), *courseID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update course status")
		return
	}

	respondWithJSON(w, http.StatusOK, course)
}

// Delete soft-deletes a course.
func (h *CourseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["id"])
	if courseID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	// Verify ownership
	existing, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}
	tenantID := getTenantIDFromContext(r)
	if tenantID != nil && existing.TenantID != *tenantID {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	if err := h.db.Queries.DeleteCourse(r.Context(), *courseID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete course")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// ListByCreator lists courses for the creator's tenant with pagination.
func (h *CourseHandler) ListByCreator(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	page, limit := parsePagination(r)
	status := r.URL.Query().Get("status")

	courses, err := h.db.Queries.ListCoursesByTenant(r.Context(), gen.ListCoursesByTenantParams{
		TenantID: *tenantID,
		Column2: status,
		Limit:    int32(limit),
		Offset:   int32((page - 1) * limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list courses")
		return
	}

	total, _ := h.db.Queries.CountCoursesByTenant(r.Context(), gen.CountCoursesByTenantParams{
		TenantID: *tenantID,
		Column2: status,
	})

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"courses": courses,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// ListStorefront lists published courses for a storefront (public, tenant resolved from middleware).
func (h *CourseHandler) ListStorefront(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil {
		respondWithError(w, http.StatusNotFound, "Storefront not found")
		return
	}

	page, limit := parsePagination(r)

	courses, err := h.db.Queries.ListCoursesByTenant(r.Context(), gen.ListCoursesByTenantParams{
		TenantID: *tenantID,
		Column2: "published",
		Limit:    int32(limit),
		Offset:   int32((page - 1) * limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list courses")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"courses": courses,
		"page":    page,
		"limit":   limit,
	})
}

// GetStorefront returns a published course for the storefront (public).
func (h *CourseHandler) GetStorefront(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil {
		respondWithError(w, http.StatusNotFound, "Storefront not found")
		return
	}

	slug := mux.Vars(r)["slug"]
	course, err := h.db.Queries.GetPublishedCourseBySlug(r.Context(), gen.GetPublishedCourseBySlugParams{
		TenantID: *tenantID,
		Slug:     slug,
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	respondWithJSON(w, http.StatusOK, course)
}

// ListMarketplace lists published courses across all tenants (platform marketplace).
func (h *CourseHandler) ListMarketplace(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePagination(r)
	category := r.URL.Query().Get("category")

	courses, err := h.db.Queries.ListMarketplaceCourses(r.Context(), gen.ListMarketplaceCoursesParams{
		Column1: category,
		Limit:    int32(limit),
		Offset:   int32((page - 1) * limit),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list marketplace courses")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"courses": courses,
		"page":    page,
		"limit":   limit,
	})
}

// =============================================================================
// Helpers (shared across all course handlers)
// =============================================================================


func parseUUID(s string) *uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return nil
	}
	return &id
}

func parsePagination(r *http.Request) (page, limit int) {
	page, _ = strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 25
	}
	return
}

