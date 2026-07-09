// Package email — Course platform email templates and notification service.
// Sends transactional emails for course events: enrollment, completion, payout, certificate.
package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"

	"github.com/resend/resend-go/v2"
)

//go:embed templates/*.html
var templateFS embed.FS

// CourseEmailService sends course-related transactional emails.
type CourseEmailService struct {
	client      *resend.Client
	fromEmail   string
	fromName    string
	templates   *template.Template
}

// NewCourseEmailService creates a new email service for course notifications.
func NewCourseEmailService(apiKey, fromEmail, fromName string) *CourseEmailService {
	// Parse embedded templates
	tmpls, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		slog.Error("Failed to parse email templates", "error", err)
		tmpls = template.New("fallback")
	}

	var client *resend.Client
	if apiKey != "" {
		client = resend.NewClient(apiKey)
	}

	return &CourseEmailService{
		client:    client,
		fromEmail: fromEmail,
		fromName:  fromName,
		templates: tmpls,
	}
}

// SendEnrollmentEmail sends a course enrollment confirmation to the learner.
func (s *CourseEmailService) SendEnrollmentEmail(ctx context.Context, to, learnerName, courseTitle, creatorName, courseURL string) error {
	data := map[string]interface{}{
		"LearnerName":  learnerName,
		"CourseTitle":  courseTitle,
		"CreatorName":  creatorName,
		"CourseURL":    courseURL,
		"Subject":      fmt.Sprintf("You're enrolled in %s", courseTitle),
	}
	return s.sendTemplate(ctx, to, "You're enrolled in "+courseTitle, "enrollment_created.html", data)
}

// SendCertificateEmail sends a certificate issuance notification.
func (s *CourseEmailService) SendCertificateEmail(ctx context.Context, to, learnerName, courseTitle, creatorName, certificateURL, certificateNumber string) error {
	data := map[string]interface{}{
		"LearnerName":      learnerName,
		"CourseTitle":      courseTitle,
		"CreatorName":      creatorName,
		"CertificateURL":   certificateURL,
		"CertificateNumber": certificateNumber,
		"Subject":          fmt.Sprintf("Your certificate for %s", courseTitle),
	}
	return s.sendTemplate(ctx, to, "Your certificate is ready", "certificate_issued.html", data)
}

// SendPayoutEmail sends a payout notification to the creator.
func (s *CourseEmailService) SendPayoutEmail(ctx context.Context, to, creatorName string, amountCents int64, currency, payoutDate string) error {
	data := map[string]interface{}{
		"CreatorName": creatorName,
		"Amount":      fmt.Sprintf("$%.2f", float64(amountCents)/100),
		"Currency":    currency,
		"PayoutDate":  payoutDate,
		"Subject":     fmt.Sprintf("Payout of %s sent", fmt.Sprintf("$%.2f", float64(amountCents)/100)),
	}
	return s.sendTemplate(ctx, to, "Your payout has been sent", "payout_paid.html", data)
}

// SendRefundEmail sends a refund notification to the learner.
func (s *CourseEmailService) SendRefundEmail(ctx context.Context, to, learnerName, courseTitle string, amountCents int64, reason string) error {
	data := map[string]interface{}{
		"LearnerName": learnerName,
		"CourseTitle": courseTitle,
		"Amount":      fmt.Sprintf("$%.2f", float64(amountCents)/100),
		"Reason":      reason,
		"Subject":     fmt.Sprintf("Refund processed for %s", courseTitle),
	}
	return s.sendTemplate(ctx, to, "Refund processed", "refund_processed.html", data)
}

// sendTemplate renders an HTML template and sends the email.
func (s *CourseEmailService) sendTemplate(ctx context.Context, to, subject, templateName string, data interface{}) error {
	if s.client == nil {
		slog.Info("Email not sent (no API key configured)", "to", to, "subject", subject)
		return nil
	}

	var body bytes.Buffer
	if err := s.templates.ExecuteTemplate(&body, templateName, data); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	req := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{to},
		Subject: subject,
		Html:    body.String(),
	}

	_, err := s.client.Emails.Send(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
