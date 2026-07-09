#!/bin/bash
cd /home/z/my-project/mycourses/backend/internal/api/handlers

# section.go
cat > section.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
GOEOF

# lesson.go
cat > lesson.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	gen "mycourses/internal/db/gen"
)

type LessonHandler struct{ db *db.DB }

func NewLessonHandler(database *db.DB) *LessonHandler { return &LessonHandler{db: database} }

func (h *LessonHandler) Create(w http.ResponseWriter, r *http.Request) {
	sectionID := parseUUID(mux.Vars(r)["sectionId"])
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if sectionID == nil || courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid IDs"); return }

	var req struct{ Title, Type, Content string; SortOrder, DurationSec int; IsPreview bool }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
	if req.Title == "" { respondWithError(w, http.StatusBadRequest, "Title required"); return }
	if req.Type == "" { req.Type = "video" }

	lesson, err := h.db.Queries.CreateLesson(r.Context(), gen.CreateLessonParams{
		SectionID: *sectionID, CourseID: *courseID, Title: req.Title, Type: req.Type,
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
GOEOF

# enrollment.go
cat > enrollment.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
		UserID: *userID, Status: []string{"active", "completed"},
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
		TenantID: *tenantID, Status: &status, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"enrollments": enrollments, "page": page, "limit": limit})
}
GOEOF

# course_progress.go
cat > course_progress.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	gen "mycourses/internal/db/gen"
)

type ProgressHandler struct{ db *db.DB }

func NewProgressHandler(database *db.DB) *ProgressHandler { return &ProgressHandler{db: database} }

// Update upserts lesson progress (called by video player every 15 seconds).
func (h *ProgressHandler) Update(w http.ResponseWriter, r *http.Request) {
	lessonID := parseUUID(mux.Vars(r)["lessonId"])
	if lessonID == nil { respondWithError(w, http.StatusBadRequest, "Invalid lesson ID"); return }

	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Authentication required"); return }

	var req struct{ VideoPositionSec int; Completed bool }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }

	// Find the lesson to get course_id
	lesson, err := h.db.Queries.GetLessonByID(r.Context(), *lessonID)
	if err != nil { respondWithError(w, http.StatusNotFound, "Lesson not found"); return }

	// Find enrollment
	enrollment, err := h.db.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
		CourseID: lesson.CourseID, UserID: *userID,
	})
	if err != nil || enrollment.ID == uuidNil() {
		respondWithError(w, http.StatusForbidden, "Not enrolled in this course")
		return
	}

	progress, err := h.db.Queries.UpsertProgress(r.Context(), gen.UpsertProgressParams{
		EnrollmentID: enrollment.ID, LessonID: *lessonID, UserID: *userID,
		Completed: req.Completed, VideoPositionSec: int32(req.VideoPositionSec),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }

	// If completed, check if all lessons are done → complete enrollment + issue certificate
	if req.Completed {
		completed, _ := h.db.Queries.CountCompletedLessonsByEnrollment(r.Context(), enrollment.ID)
		total, _ := h.db.Queries.CountLessonsByCourse(r.Context(), lesson.CourseID)
		if completed == total && total > 0 {
			h.db.Queries.CompleteEnrollment(r.Context(), enrollment.ID)
			// Certificate issuance would go here
		}
	}

	respondWithJSON(w, http.StatusOK, progress)
}

// GetByCourse returns all progress for a learner in a course.
func (h *ProgressHandler) GetByCourse(w http.ResponseWriter, r *http.Request) {
	courseID := parseUUID(mux.Vars(r)["courseId"])
	if courseID == nil { respondWithError(w, http.StatusBadRequest, "Invalid course ID"); return }

	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Authentication required"); return }

	enrollment, err := h.db.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
		CourseID: *courseID, UserID: *userID,
	})
	if err != nil || enrollment.ID == uuidNil() {
		respondWithError(w, http.StatusForbidden, "Not enrolled")
		return
	}

	progress, err := h.db.Queries.ListProgressByEnrollment(r.Context(), enrollment.ID)
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }

	pct, _ := h.db.Queries.GetCourseCompletionPercentage(r.Context(), enrollment.ID)
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"progress": progress, "completionPercentage": pct,
	})
}
GOEOF

# coupon.go
cat > coupon.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	gen "mycourses/internal/db/gen"
)

type CouponHandler struct{ db *db.DB }

func NewCouponHandler(database *db.DB) *CouponHandler { return &CouponHandler{db: database} }

func (h *CouponHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct{ Code, DiscountType string; DiscountValue int; Currency, CourseID string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
	if req.Code == "" { respondWithError(w, http.StatusBadRequest, "Code required"); return }
	if req.Currency == "" { req.Currency = "usd" }

	var courseID *uuid.UUID
	if req.CourseID != "" { courseID = parseUUID(req.CourseID) }

	coupon, err := h.db.Queries.CreateCoupon(r.Context(), gen.CreateCouponParams{
		TenantID: *tenantID, Upper: req.Code, DiscountType: req.DiscountType,
		DiscountValue: int32(req.DiscountValue), Currency: req.Currency, CourseID: courseID,
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }
	respondWithJSON(w, http.StatusCreated, coupon)
}

func (h *CouponHandler) Delete(w http.ResponseWriter, r *http.Request) {
	couponID := parseUUID(mux.Vars(r)["id"])
	if couponID == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }
	if err := h.db.Queries.DeleteCoupon(r.Context(), *couponID); err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (h *CouponHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }
	page, limit := parsePagination(r)
	coupons, err := h.db.Queries.ListCouponsByTenant(r.Context(), gen.ListCouponsByTenantParams{
		TenantID: *tenantID, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"coupons": coupons, "page": page, "limit": limit})
}

func (h *CouponHandler) Validate(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct{ Code string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }

	coupon, err := h.db.Queries.ValidateCoupon(r.Context(), gen.ValidateCouponParams{
		TenantID: *tenantID, Upper: req.Code,
	})
	if err != nil { respondWithError(w, http.StatusNotFound, "Invalid or expired coupon"); return }
	respondWithJSON(w, http.StatusOK, coupon)
}
GOEOF

# review.go
cat > review.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
GOEOF

# custom_domain.go
cat > custom_domain.go << 'GOEOF'
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	gen "mycourses/internal/db/gen"
)

type CustomDomainHandler struct{ db *db.DB }

func NewCustomDomainHandler(database *db.DB) *CustomDomainHandler { return &CustomDomainHandler{db: database} }

func (h *CustomDomainHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }

	var req struct{ Domain string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { respondWithError(w, http.StatusBadRequest, "Invalid body"); return }
	if req.Domain == "" { respondWithError(w, http.StatusBadRequest, "Domain required"); return }

	domain, err := h.db.Queries.CreateCustomDomain(r.Context(), gen.CreateCustomDomainParams{
		Domain: req.Domain, TenantID: *tenantID, VerificationRecords: []byte("{}"),
	})
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed: "+err.Error()); return }
	respondWithJSON(w, http.StatusCreated, domain)
}

func (h *CustomDomainHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := getTenantIDFromContext(r)
	if tenantID == nil { respondWithError(w, http.StatusForbidden, "Tenant required"); return }
	domains, err := h.db.Queries.ListCustomDomainsByTenant(r.Context(), *tenantID)
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, domains)
}

func (h *CustomDomainHandler) Delete(w http.ResponseWriter, r *http.Request) {
	domainID := parseUUID(mux.Vars(r)["id"])
	if domainID == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }
	if err := h.db.Queries.DeleteCustomDomain(r.Context(), *domainID); err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}
GOEOF

# certificate.go
cat > certificate.go << 'GOEOF'
package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	gen "mycourses/internal/db/gen"
)

type CertificateHandler struct{ db *db.DB }

func NewCertificateHandler(database *db.DB) *CertificateHandler { return &CertificateHandler{db: database} }

func (h *CertificateHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Auth required"); return }
	certs, err := h.db.Queries.ListCertificatesByUser(r.Context(), *userID)
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, certs)
}

func (h *CertificateHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	cert, err := h.db.Queries.GetCertificateByToken(r.Context(), token)
	if err != nil { respondWithError(w, http.StatusNotFound, "Certificate not found or revoked"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"valid": true, "certificate": cert})
}
GOEOF

# helpers.go — shared helpers for all course handlers
cat > course_helpers.go << 'GOEOF'
package handlers

import (
	"mycourses/internal/db"
	"github.com/google/uuid"
)

func getUserIDFromContext(r *http.Request) *uuid.UUID {
	return nil // TODO: wire to auth middleware
}

func uuidNil() uuid.UUID { return uuid.Nil }
GOEOF

echo "Handler files created"
ls *.go | grep -E "course|section|lesson|enrollment|progress|coupon|review|custom_domain|certificate|course_helpers" | wc -l