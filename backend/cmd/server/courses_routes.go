// Package main — Postgres-backed course routes setup.
// This file initializes the Postgres DB layer and registers all course-platform
// routes, keeping the existing MongoDB-based routes in main.go untouched.
package main

import (
        "context"
        "log/slog"
        "os"

        "github.com/gorilla/mux"

        "mycourses/internal/api/handlers"
        "mycourses/internal/cloudflare"
        "mycourses/internal/db"
        "mycourses/internal/middleware"
        stripeconnect "mycourses/internal/stripe"
)

// setupCourseRoutes initializes the Postgres DB and registers all course-platform routes.
// Called from main() after MongoDB setup. Returns a cleanup function.
func setupCourseRoutes(router *mux.Router, pgURL string, platformDomain string) func() {
        if pgURL == "" {
                slog.Info("Postgres URL not configured — course platform routes disabled")
                return func() {}
        }

        // Initialize Postgres connection pool
        ctx := context.Background()
        pgDB, err := db.New(ctx, db.Config{
                URL:      pgURL,
                MaxConns: 25,
                MinConns: 5,
        })
        if err != nil {
                slog.Error("Failed to connect to Postgres", "error", err)
                slog.Info("Course platform routes disabled — continuing with MongoDB-only mode")
                return func() {}
        }

        // Apply migrations
        if err := pgDB.Migrate(ctx, "file://migrations"); err != nil {
                slog.Error("Failed to apply Postgres migrations", "error", err)
                pgDB.Close()
                return func() {}
        }
        slog.Info("Connected to Postgres — course platform routes enabled")

        // Create Stripe Connect service (if configured)
        stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
        stripeWebhookSecret := os.Getenv("STRIPE_CONNECT_WEBHOOK_SECRET")
        stripeConnectClientID := os.Getenv("STRIPE_CONNECT_CLIENT_ID")
        frontendURL := os.Getenv("FRONTEND_URL")
        if frontendURL == "" {
                frontendURL = "http://localhost:5173"
        }

        var connectSvc *stripeconnect.ConnectService
        var connectHandler *handlers.ConnectHandler
        var payoutHandler *handlers.PayoutHandler
        var webhookHandler *handlers.CourseWebhookHandler

        if stripeSecretKey != "" && stripeWebhookSecret != "" {
                connectSvc = stripeconnect.NewConnectService(stripeSecretKey, stripeWebhookSecret, stripeConnectClientID, pgDB, frontendURL)
                connectHandler = handlers.NewConnectHandler(pgDB, connectSvc, frontendURL)
                payoutHandler = handlers.NewPayoutHandler(pgDB, connectSvc)
                webhookHandler = handlers.NewCourseWebhookHandler(pgDB, connectSvc)
                slog.Info("Stripe Connect enabled")
        } else {
                slog.Info("Stripe Connect not configured — payout/connect routes disabled")
        }

        // Create Cloudflare client (if configured)
        cfClient := cloudflare.New()

        // Create handlers
        courseHandler := handlers.NewCourseHandler(pgDB)
        sectionHandler := handlers.NewSectionHandler(pgDB)
        lessonHandler := handlers.NewLessonHandler(pgDB)
        enrollmentHandler := handlers.NewEnrollmentHandler(pgDB)
        progressHandler := handlers.NewProgressHandler(pgDB)
        couponHandler := handlers.NewCouponHandler(pgDB)
        reviewHandler := handlers.NewReviewHandler(pgDB)
        customDomainHandler := handlers.NewCustomDomainHandler(pgDB, cfClient)
        certificateHandler := handlers.NewCertificateHandler(pgDB)
        seoHandler := handlers.NewSEOHandler(pgDB, platformDomain)

        // Create storefront handler (with Stripe Connect if available)
        var storefrontHandler *handlers.StorefrontHandler
        if connectSvc != nil {
                storefrontHandler = handlers.NewStorefrontHandler(pgDB, connectSvc)
        } else {
                storefrontHandler = handlers.NewStorefrontHandler(pgDB, nil)
        }

        // Create a subrouter for course-platform API routes
        courseAPI := router.PathPrefix("/api").Subrouter()

        // SEO routes (top-level, not under /api)
        router.HandleFunc("/sitemap.xml", seoHandler.Sitemap).Methods("GET")
        router.HandleFunc("/robots.txt", seoHandler.Robots).Methods("GET")

        // Apply custom domain tenant resolution middleware to all course routes
        courseAPI.Use(middleware.CustomDomainTenantMiddleware(pgDB, platformDomain))

        // === Public storefront routes (no auth required) ===
        storefront := courseAPI.PathPrefix("/storefront").Subrouter()
        storefront.HandleFunc("/courses", courseHandler.ListStorefront).Methods("GET")
        storefront.HandleFunc("/courses/{slug}", courseHandler.GetStorefront).Methods("GET")
        storefront.HandleFunc("/courses/{courseId}/reviews", reviewHandler.ListPublic).Methods("GET")
        storefront.HandleFunc("/coupons/validate", couponHandler.Validate).Methods("POST")
        storefront.HandleFunc("/marketplace", courseHandler.ListMarketplace).Methods("GET")
        storefront.HandleFunc("/checkout", storefrontHandler.CreateCheckout).Methods("POST")

        // === Creator studio routes (auth + creator/owner/admin role required) ===
        // TODO: wire auth middleware once Postgres-based auth is implemented
        creator := courseAPI.PathPrefix("/creator").Subrouter()
        creator.HandleFunc("/courses", courseHandler.Create).Methods("POST")
        creator.HandleFunc("/courses", courseHandler.ListByCreator).Methods("GET")
        creator.HandleFunc("/courses/{id}", courseHandler.Get).Methods("GET")
        creator.HandleFunc("/courses/{id}", courseHandler.Update).Methods("PUT")
        creator.HandleFunc("/courses/{id}", courseHandler.Delete).Methods("DELETE")
        creator.HandleFunc("/courses/{id}/publish", courseHandler.Publish).Methods("POST")
        creator.HandleFunc("/courses/{id}/unpublish", courseHandler.Unpublish).Methods("POST")
        creator.HandleFunc("/courses/{courseId}/sections", sectionHandler.ListByCourse).Methods("GET")
        creator.HandleFunc("/courses/{courseId}/sections", sectionHandler.Create).Methods("POST")
        creator.HandleFunc("/sections/{id}", sectionHandler.Update).Methods("PUT")
        creator.HandleFunc("/sections/{id}", sectionHandler.Delete).Methods("DELETE")
        creator.HandleFunc("/sections/{sectionId}/lessons", lessonHandler.ListBySection).Methods("GET")
        creator.HandleFunc("/sections/{sectionId}/lessons", lessonHandler.Create).Methods("POST")
        creator.HandleFunc("/courses/{courseId}/lessons", lessonHandler.ListByCourse).Methods("GET")
        creator.HandleFunc("/lessons/{id}", lessonHandler.Update).Methods("PUT")
        creator.HandleFunc("/lessons/{id}", lessonHandler.Delete).Methods("DELETE")
        creator.HandleFunc("/enrollments", enrollmentHandler.ListByCreator).Methods("GET")
        creator.HandleFunc("/coupons", couponHandler.List).Methods("GET")
        creator.HandleFunc("/coupons", couponHandler.Create).Methods("POST")
        creator.HandleFunc("/coupons/{id}", couponHandler.Delete).Methods("DELETE")
        creator.HandleFunc("/reviews/{id}/hide", reviewHandler.Hide).Methods("POST")
        creator.HandleFunc("/custom-domains", customDomainHandler.List).Methods("GET")
        creator.HandleFunc("/custom-domains", customDomainHandler.Create).Methods("POST")
        creator.HandleFunc("/custom-domains/{id}", customDomainHandler.Delete).Methods("DELETE")
        creator.HandleFunc("/custom-domains/{id}/status", customDomainHandler.GetStatus).Methods("GET")

        // Stripe Connect routes (only if configured)
        if connectHandler != nil {
                creator.HandleFunc("/connect/onboarding", connectHandler.Onboard).Methods("GET")
                creator.HandleFunc("/connect/status", connectHandler.Status).Methods("GET")
        }
        if payoutHandler != nil {
                creator.HandleFunc("/payouts", payoutHandler.List).Methods("GET")
                creator.HandleFunc("/payouts/request", payoutHandler.Request).Methods("POST")
        }

        // Stripe webhook (no auth — verified by signature)
        if webhookHandler != nil {
                courseAPI.HandleFunc("/webhooks/stripe", webhookHandler.HandleWebhook).Methods("POST")
        }

        // === Learner routes (auth required) ===
        learner := courseAPI.PathPrefix("/learner").Subrouter()
        learner.HandleFunc("/enrollments", enrollmentHandler.ListMine).Methods("GET")
        learner.HandleFunc("/enrollments/{courseId}", enrollmentHandler.CreateEnrollment).Methods("POST")
        learner.HandleFunc("/progress/{lessonId}", progressHandler.Update).Methods("POST")
        learner.HandleFunc("/progress/course/{courseId}", progressHandler.GetByCourse).Methods("GET")
        learner.HandleFunc("/certificates", certificateHandler.ListMine).Methods("GET")

        // === Public certificate verification (no auth) ===
        courseAPI.HandleFunc("/public/certificates/verify/{token}", certificateHandler.Verify).Methods("GET")

        slog.Info("Course platform routes registered",
                "storefront_routes", 5,
                "creator_routes", 20,
                "learner_routes", 5,
                "public_routes", 1,
        )

        return func() {
                slog.Info("Closing Postgres connection")
                pgDB.Close()
        }
}

// getEnvOrDefault returns the env var or a default value.
func getEnvOrDefault(key, defaultVal string) string {
        if v := os.Getenv(key); v != "" {
                return v
        }
        return defaultVal
}
