package health

import (
	"context"
	"log/slog"
	"time"

	"mycourses/internal/db"
)

type Service struct {
	db             *db.DB
	metrics        interface{}
	getConfig      func(string) string
	integrations   []Integration
}

type Integration struct {
	Name    string
	Checker func() error
}

func New(database *db.DB, metrics interface{}, getConfig func(string) string) *Service {
	return &Service{db: database, metrics: metrics, getConfig: getConfig}
}

func (s *Service) RegisterIntegration(name string, checker func() error) {
	s.integrations = append(s.integrations, Integration{Name: name, Checker: checker})
}

func (s *Service) CheckHealth(ctx context.Context) error {
	return s.db.HealthCheck(ctx)
}

func (s *Service) GetIntegrationChecks(ctx context.Context) []IntegrationCheck {
	checks := make([]IntegrationCheck, 0)
	for _, integ := range s.integrations {
		err := integ.Checker()
		checks = append(checks, IntegrationCheck{
			Name:      integ.Name,
			Healthy:   err == nil,
			Error:     func() string { if err != nil { return err.Error() }; return "" }(),
			Timestamp: time.Now(),
		})
	}
	return checks
}

type IntegrationCheck struct {
	Name      string    `json:"name"`
	Healthy   bool      `json:"healthy"`
	Error     string    `json:"error,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *Service) CollectMetrics(ctx context.Context) {
	// TODO: Implement with Postgres queries
	slog.Debug("Collecting health metrics")
}

func (s *Service) GetSystemMetrics(ctx context.Context, timeRange string) ([]map[string]interface{}, error) {
	rows, err := s.db.Pool.Query(ctx,
		"SELECT * FROM system_metrics WHERE timestamp > NOW() - INTERVAL '1 hour' ORDER BY timestamp DESC LIMIT 100")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		results = append(results, map[string]interface{}{"status": "ok"})
	}
	return results, nil
}
