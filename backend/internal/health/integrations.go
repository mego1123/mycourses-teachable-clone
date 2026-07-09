package health

import (
	"time"
)

// NewIntegration creates a health integration check.
func NewIntegration(name string, checker func() error) Integration {
	return Integration{Name: name, Checker: checker}
}

// RunIntegrationChecks runs all registered integration checks.
func (s *Service) RunIntegrationChecks() []IntegrationCheck {
	checks := make([]IntegrationCheck, 0, len(s.integrations))
	for _, integ := range s.integrations {
		err := integ.Checker()
		check := IntegrationCheck{
			Name:      integ.Name,
			Healthy:   err == nil,
			Timestamp: time.Now(),
		}
		if err != nil {
			check.Error = err.Error()
		}
		checks = append(checks, check)
	}
	return checks
}
