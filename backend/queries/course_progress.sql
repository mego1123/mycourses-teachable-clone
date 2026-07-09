-- name: GetProgressByEnrollmentAndLesson :one
SELECT * FROM course_progress WHERE enrollment_id = $1 AND lesson_id = $2;
-- name: UpsertProgress :one
INSERT INTO course_progress (enrollment_id, lesson_id, user_id, completed, video_position_sec, last_viewed_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (enrollment_id, lesson_id) DO UPDATE SET completed = EXCLUDED.completed, completed_at = CASE WHEN EXCLUDED.completed = TRUE THEN COALESCE(course_progress.completed_at, NOW()) ELSE NULL END, video_position_sec = EXCLUDED.video_position_sec, last_viewed_at = NOW()
RETURNING *;
-- name: ListProgressByEnrollment :many
SELECT * FROM course_progress WHERE enrollment_id = $1 ORDER BY last_viewed_at DESC;
-- name: CountCompletedLessonsByEnrollment :one
SELECT COUNT(*) FROM course_progress WHERE enrollment_id = $1 AND completed = TRUE;
-- name: GetCourseCompletionPercentage :one
SELECT CASE WHEN COUNT(*) = 0 THEN 0 ELSE (COUNT(cp.id) * 100 / COUNT(*))::int END as percentage FROM lessons LEFT JOIN course_progress cp ON cp.lesson_id = lessons.id AND cp.enrollment_id = $1 AND cp.completed = TRUE WHERE lessons.course_id = (SELECT course_id FROM enrollments WHERE id = $1);
