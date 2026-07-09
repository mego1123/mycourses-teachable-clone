// Package handlers — Storefront checkout handler for Stripe Checkout Session creation.
package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
	stripeconnect "mycourses/internal/stripe"
)

// StorefrontHandler handles public storefront operations.
type StorefrontHandler struct {
	db         *db.DB
	connectSvc *stripeconnect.ConnectService
}

func NewStorefrontHandler(database *db.DB, connectSvc *stripeconnect.ConnectService) *StorefrontHandler {
	return &StorefrontHandler{db: database, connectSvc: connectSvc}
}

// CreateCheckout creates a Stripe Checkout Session for a course purchase.
// For free courses (price = 0), creates enrollment directly.
func (h *StorefrontHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseID  string `json:"courseId"`
		CouponCode string `json:"couponCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get the course
	courseID := parseUUID(req.CourseID)
	if courseID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid course ID")
		return
	}

	course, err := h.db.Queries.GetCourseByID(r.Context(), *courseID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	// Verify course is published
	if course.Status != "published" {
		respondWithError(w, http.StatusBadRequest, "Course is not available")
		return
	}

	// Get tenant
	tenant, err := h.db.Queries.GetTenantByID(r.Context(), course.TenantID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Creator not found")
		return
	}

	// For free courses, create enrollment directly
	if course.PriceCents == 0 {
		// TODO: Get user ID from auth context
		// For now, return a message that auth is required
		respondWithError(w, http.StatusUnauthorized, "Authentication required for enrollment")
		return
	}

	// For paid courses, create Stripe Checkout Session
	if h.connectSvc == nil {
		respondWithError(w, http.StatusServiceUnavailable, "Payment processing not configured")
		return
	}

	// Get user info from context (placeholder — will use auth middleware)
	userID := getUserIDFromContext(r)
	if userID == nil {
		respondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Get user email
	user, err := h.db.Queries.GetUserByID(r.Context(), *userID)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "User not found")
		return
	}

	// Build success/cancel URLs
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	successURL := frontendURL + "/learn/course/" + course.ID.String() + "?status=success"
	cancelURL := frontendURL + "/courses/" + course.Slug + "?status=cancelled"

	// Create Stripe Checkout Session
	session, err := h.connectSvc.CreateCourseCheckout(
		r.Context(),
		&course,
		&tenant,
		user.Email,
		user.ID.String(),
		successURL,
		cancelURL,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create checkout session: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"checkoutUrl": session.URL,
		"sessionId":   session.ID,
	})
}

// GetCourseDetail returns a published course with its sections and lessons for the storefront.
func (h *StorefrontHandler) GetCourseDetail(w http.ResponseWriter, r *http.Request) {
	slug := mux.Vars(r)["slug"]
	tenant := middleware.GetPgTenantFromContext(r.Context())
	if tenant == nil {
		respondWithError(w, http.StatusNotFound, "Storefront not found")
		return
	}

	course, err := h.db.Queries.GetPublishedCourseBySlug(r.Context(), gen.GetPublishedCourseBySlugParams{
		TenantID: tenant.ID,
		Slug:     slug,
	})
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Course not found")
		return
	}

	respondWithJSON(w, http.StatusOK, course)
}
