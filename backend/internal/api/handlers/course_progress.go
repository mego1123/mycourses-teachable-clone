package handlers

import (
        "encoding/json"
        "net/http"

        "github.com/google/uuid"
        "github.com/gorilla/mux"

        "mycourses/internal/db"
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

                        // Auto-issue certificate if not already issued
                        existing, _ := h.db.Queries.GetCertificateByEnrollment(r.Context(), enrollment.ID)
                        if existing.ID == uuidNil() {
                                // Get course and user for certificate details
                                course, _ := h.db.Queries.GetCourseByID(r.Context(), lesson.CourseID)
                                tenant, _ := h.db.Queries.GetTenantByID(r.Context(), course.TenantID)
                                user, _ := h.db.Queries.GetUserByID(r.Context(), *userID)

                                certNum, _ := h.db.Queries.GetNextCertificateNumber(r.Context())
                                certNumStr, _ := certNum.(string)

                                learnerName := ""
                                if user.ID != uuid.Nil {
                                        learnerName = user.DisplayName
                                }
                                courseTitle := ""
                                if course.ID != uuid.Nil {
                                        courseTitle = course.Title
                                }
                                creatorName := ""
                                if tenant.ID != uuid.Nil {
                                        creatorName = tenant.Name
                                }

                                h.db.Queries.CreateCertificate(r.Context(), gen.CreateCertificateParams{
                                        EnrollmentID:      enrollment.ID,
                                        UserID:            *userID,
                                        CourseID:          lesson.CourseID,
                                        TenantID:          course.TenantID,
                                        CertificateNumber: certNumStr,
                                        VerificationToken: uuid.New().String(),
                                        LearnerName:       learnerName,
                                        CourseTitle:       courseTitle,
                                        CreatorName:       creatorName,
                                })
                        }
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
