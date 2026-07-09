package telemetry

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"mycourses/internal/db"
	"mycourses/internal/models"
)

// Service provides telemetry tracking and analytics.
// NOTE: During MongoDB→Postgres migration, analytics functions are temporarily
// disabled. They need to be rewritten to use Postgres queries (db.Pool.Query).
// The Track function still works (writes to telemetry_events via Pool.Exec).
type Service struct {
	db       *db.DB
	trackCh  chan models.TelemetryEvent
	onTrack  func(models.TelemetryEvent)
}

func New(database *db.DB) *Service {
	s := &Service{
		db:      database,
		trackCh: make(chan models.TelemetryEvent, 1000),
	}
	go s.processBatch()
	return s
}

func (s *Service) SetOnTrack(fn func(models.TelemetryEvent)) {
	s.onTrack = fn
}

func (s *Service) Track(ctx context.Context, event models.TelemetryEvent) error {
	select {
	case s.trackCh <- event:
	default:
		slog.Warn("telemetry channel full, dropping event")
	}
	return nil
}

func (s *Service) TrackPageView(ctx context.Context, event models.TelemetryEvent) error {
	event.EventName = models.TelemetryPageView
	return s.Track(ctx, event)
}

func (s *Service) Close() {
	close(s.trackCh)
}

func (s *Service) processBatch() {
	for event := range s.trackCh {
		s.insertEvent(event)
		if s.onTrack != nil {
			s.onTrack(event)
		}
	}
}

func (s *Service) insertEvent(event models.TelemetryEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var userID *string
	if event.UserID == nil || *event.UserID == uuid.Nil {
		id := event.UserID.String()
		userID = &id
	}

	s.db.Pool.Exec(ctx,
		`INSERT INTO telemetry_events (user_id, event_type, event_name, properties, session_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		userID, string(event.EventName), string(event.EventName), "{}", event.SessionID)
}

// Analytics functions — temporarily disabled during migration.
// These need to be rewritten with Postgres queries.

type FunnelData struct{}
type CohortRow struct{}
type EngagementData struct{}
type KPIData struct{}
type CustomEventData struct{}
type EventTypeSummary struct{}
type SankeyData struct{}

func (s *Service) FunnelMetrics(ctx context.Context, start, end time.Time) (*FunnelData, error) {
	return &FunnelData{}, nil
}

func (s *Service) CohortAnalysis(ctx context.Context, start, end time.Time) ([]CohortRow, error) {
	return nil, nil
}

func (s *Service) EngagementMetrics(ctx context.Context, start, end time.Time) (*EngagementData, error) {
	return &EngagementData{}, nil
}

func (s *Service) KPIMetrics(ctx context.Context, start, end time.Time) (*KPIData, error) {
	return &KPIData{}, nil
}

func (s *Service) CustomEvents(ctx context.Context, start, end time.Time) ([]CustomEventData, error) {
	return nil, nil
}

func (s *Service) EventTypeSummary(ctx context.Context, start, end time.Time) ([]EventTypeSummary, error) {
	return nil, nil
}

func (s *Service) SankeyData(ctx context.Context, start, end time.Time) (*SankeyData, error) {
	return &SankeyData{}, nil
}

func (s *Service) ListEventTypes(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (s *Service) TrackBatch(ctx context.Context, events []models.TelemetryEvent) error {
	for _, event := range events {
		s.Track(ctx, event)
	}
	return nil
}

