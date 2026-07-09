// Package middleware — RequireEnrollment middleware.
// Verifies that the authenticated user has an active enrollment in the course
// identified by the course_id URL parameter (or lesson_id → course_id lookup).
package middleware

import (
        "context"
        "time"
        "net/http"

        "github.com/gorilla/mux"
        "github.com/google/uuid"

        "mycourses/internal/db"
        gen "mycourses/internal/db/gen"
)

// RequireEnrollment returns middleware that checks user enrollment in the course.
// The courseIDParam specifies which URL parameter contains the course ID.
// If the lesson is a free preview (is_preview=true), access is allowed without enrollment.
func RequireEnrollment(database *db.DB, courseIDParam string) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        // Get user ID from auth context (using existing auth middleware context)
                        userID := getUserIDFromAuthContext(r.Context())
                        if userID == uuid.Nil {
                                http.Error(w, `{"error":"Authentication required"}`, http.StatusUnauthorized)
                                return
                        }

                        courseIDStr := mux.Vars(r)[courseIDParam]
                        courseID, err := uuid.Parse(courseIDStr)
                        if err != nil {
                                http.Error(w, `{"error":"Invalid course ID"}`, http.StatusBadRequest)
                                return
                        }

                        // Check enrollment
                        enrollment, err := database.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
                                CourseID: courseID,
                                UserID:   userID,
                        })
                        if err != nil || enrollment.ID == uuid.Nil {
                                // Check if this is a preview lesson — if so, allow access
                                lessonIDStr := mux.Vars(r)["lessonId"]
                                if lessonIDStr != "" {
                                        lessonID, err := uuid.Parse(lessonIDStr)
                                        if err == nil {
                                                lesson, err := database.Queries.GetLessonByID(r.Context(), lessonID)
                                                if err == nil && lesson.IsPreview {
                                                        next.ServeHTTP(w, r)
                                                        return
                                                }
                                        }
                                }
                                http.Error(w, `{"error":"Not enrolled in this course"}`, http.StatusForbidden)
                                return
                        }

                        // Check enrollment status
                        if enrollment.Status != "active" && enrollment.Status != "completed" {
                                http.Error(w, `{"error":"Enrollment is not active"}`, http.StatusForbidden)
                                return
                        }

                        next.ServeHTTP(w, r)
                })
        }
}

// RequireDripEligibility returns middleware that checks if a lesson is accessible
// based on the section's drip_offset_days and the learner's enrollment date.
func RequireDripEligibility(database *db.DB) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        lessonIDStr := mux.Vars(r)["lessonId"]
                        lessonID, err := uuid.Parse(lessonIDStr)
                        if err != nil {
                                next.ServeHTTP(w, r) // No lesson ID — let the handler deal with it
                                return
                        }

                        // Get the lesson to find its section
                        lesson, err := database.Queries.GetLessonByID(r.Context(), lessonID)
                        if err != nil {
                                next.ServeHTTP(w, r)
                                return
                        }

                        // Get the section to check drip_offset_days
                        section, err := database.Queries.GetSectionByID(r.Context(), lesson.SectionID)
                        if err != nil {
                                next.ServeHTTP(w, r)
                                return
                        }

                        // If no drip configured (0 days), allow access
                        if section.DripOffsetDays == 0 {
                                next.ServeHTTP(w, r)
                                return
                        }

                        // Get user ID
                        userID := getUserIDFromAuthContext(r.Context())
                        if userID == uuid.Nil {
                                next.ServeHTTP(w, r)
                                return
                        }

                        // Get enrollment to find enrollment date
                        enrollment, err := database.Queries.GetEnrollmentByCourseAndUser(r.Context(), gen.GetEnrollmentByCourseAndUserParams{
                                CourseID: lesson.CourseID,
                                UserID:   userID,
                        })
                        if err != nil || enrollment.ID == uuid.Nil {
                                next.ServeHTTP(w, r) // Let RequireEnrollment handle this
                                return
                        }

                        // Calculate unlock date
                        unlockDate := enrollment.EnrolledAt.AddDate(0, 0, int(section.DripOffsetDays))
                        if r.Context().Value("now") != nil {
                                // Use injected time for testing
                        }

                        // Check if current time is before unlock date
                        if time.Now().Before(unlockDate) {
                                http.Error(w, `{"error":"Lesson not yet available","available_at":"`+unlockDate.Format(time.RFC3339)+`"}`, http.StatusForbidden)
                                return
                        }

                        next.ServeHTTP(w, r)
                })
        }
}

// getUserIDFromAuthContext extracts the Postgres user UUID from context.
// Uses the AuthBridge middleware's context key.
func getUserIDFromAuthContext(ctx context.Context) uuid.UUID {
        return GetPgUserIDFromContext(ctx)
}
