// Package main — Plan and config seeding for the course platform.
// Seeds 3 creator plans (Free/Pro/Business) and config defaults on startup.
package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

// seedCreatorData creates the 3 creator plans and config defaults.
// Idempotent — uses ON CONFLICT DO NOTHING / DO UPDATE.
func seedCreatorData(ctx context.Context, database *db.DB) {
	seedCreatorPlans(ctx, database)
	seedConfigDefaults(ctx, database)
}

func seedCreatorPlans(ctx context.Context, database *db.DB) {
	type planSeed struct {
		name            string
		monthlyCents    int64
		yearlyCents     int64
		entitlements    map[string]interface{}
		sortOrder       int32
	}

	plans := []planSeed{
		{
			name:         "Creator Free",
			monthlyCents: 0,
			yearlyCents:  0,
			entitlements: map[string]interface{}{
				"max_courses":             map[string]interface{}{"type": "numeric", "numericValue": 1, "description": "Maximum courses"},
				"max_video_storage_mb":    map[string]interface{}{"type": "numeric", "numericValue": 100, "description": "Max video storage (MB)"},
				"custom_domain_enabled":   map[string]interface{}{"type": "bool", "boolValue": false, "description": "Custom domain support"},
				"commission_rate_bps":     map[string]interface{}{"type": "numeric", "numericValue": 2000, "description": "Platform commission (20%)"},
			},
			sortOrder: 0,
		},
		{
			name:         "Creator Pro",
			monthlyCents: 2900,
			yearlyCents:  29000,
			entitlements: map[string]interface{}{
				"max_courses":             map[string]interface{}{"type": "numeric", "numericValue": 10, "description": "Maximum courses"},
				"max_video_storage_mb":    map[string]interface{}{"type": "numeric", "numericValue": 50000, "description": "Max video storage (MB)"},
				"custom_domain_enabled":   map[string]interface{}{"type": "bool", "boolValue": true, "description": "Custom domain support"},
				"commission_rate_bps":     map[string]interface{}{"type": "numeric", "numericValue": 1000, "description": "Platform commission (10%)"},
			},
			sortOrder: 1,
		},
		{
			name:         "Creator Business",
			monthlyCents: 9900,
			yearlyCents:  99000,
			entitlements: map[string]interface{}{
				"max_courses":             map[string]interface{}{"type": "numeric", "numericValue": -1, "description": "Unlimited courses"},
				"max_video_storage_mb":    map[string]interface{}{"type": "numeric", "numericValue": 500000, "description": "Max video storage (MB)"},
				"custom_domain_enabled":   map[string]interface{}{"type": "bool", "boolValue": true, "description": "Custom domain support"},
				"commission_rate_bps":     map[string]interface{}{"type": "numeric", "numericValue": 500, "description": "Platform commission (5%)"},
			},
			sortOrder: 2,
		},
	}

	for _, p := range plans {
		entJSON, _ := json.Marshal(p.entitlements)
		_, err := database.Queries.UpsertPlan(ctx, gen.UpsertPlanParams{
			Name:               p.name,
			Description:        "Course creator plan",
			MonthlyPriceCents:  p.monthlyCents,
			YearlyPriceCents:   p.yearlyCents,
			Currency:           "usd",
			IncludedSeats:      1,
			MinSeats:           1,
			MaxSeats:           1,
			UsageCreditsPerMonth: 0,
			TrialPeriodDays:    0,
			Entitlements:       entJSON,
			IsSystem:           true,
			IsActive:           true,
			IsPublic:           true,
			SortOrder:          p.sortOrder,
		})
		if err != nil {
			slog.Error("Failed to seed plan", "name", p.name, "error", err)
		}
	}

	slog.Info("Creator plans seeded", "count", len(plans))
}

func seedConfigDefaults(ctx context.Context, database *db.DB) {
	defaults := []struct {
		key, description, category string
		value                      interface{}
	}{
		{"course.default_commission_rate_bps", "Default platform commission (10%)", "marketplace", 1000},
		{"course.payout_schedule", "Creator payout schedule", "marketplace", "weekly"},
		{"custom_domains.enabled", "Allow creators to add custom domains", "custom_domains", true},
		{"marketplace.discovery_enabled", "Show courses in marketplace", "marketplace", true},
		{"course.refund_window_days", "Refund window in days", "course", 30},
		{"i18n.default_locale", "Default locale", "i18n", "en"},
	}

	for _, d := range defaults {
		valueJSON, _ := json.Marshal(d.value)
		_, err := database.Pool.Exec(ctx,
			`INSERT INTO config_vars (key, value, description, category, is_system, is_readonly)
			 VALUES ($1, $2, $3, $4, true, false)
			 ON CONFLICT (key) DO NOTHING`,
			d.key, valueJSON, d.description, d.category)
		if err != nil {
			slog.Error("Failed to seed config", "key", d.key, "error", err)
		}
	}

	slog.Info("Config defaults seeded", "count", len(defaults))
}
