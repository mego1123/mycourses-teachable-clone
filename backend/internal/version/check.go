package version

import (
	"context"
	"log/slog"
	"time"

	"mycourses/internal/db"
)

// CheckAndMigrate compares the VERSION file to the DB version.
func CheckAndMigrate(database *db.DB) {
	if Current == "" || Current == "unknown" {
		slog.Warn("VERSION file not found, skipping version check")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Check if system is initialized by looking for a config var with key "system.version"
	var version string
	err := database.Pool.QueryRow(ctx,
		"SELECT value->>'value' FROM config_vars WHERE key = 'system.version'").Scan(&version)
	if err != nil {
		// System not initialized yet — nothing to check
		return
	}

	if version == Current {
		slog.Info("Version up to date", "version", Current)
		return
	}

	oldVersion := version
	slog.Info("Version changed", "from", oldVersion, "to", Current)

	// Update DB version
	_, err = database.Pool.Exec(ctx,
		"UPDATE config_vars SET value = $2 WHERE key = 'system.version'",
		Current)
	if err != nil {
		slog.Warn("Failed to update version in DB", "error", err)
	} else {
		slog.Info("Database version updated", "version", Current)
	}

	// Send upgrade message to root tenant owner
	sendUpgradeMessage(ctx, database, Current)
}

func sendUpgradeMessage(ctx context.Context, database *db.DB, newVersion string) {
	// Find root tenant + owner membership via direct SQL
	var userID string
	err := database.Pool.QueryRow(ctx,
		`SELECT tm.user_id FROM tenant_memberships tm
		 JOIN tenants t ON tm.tenant_id = t.id
		 WHERE t.is_root = true AND tm.role = 'owner'
		 LIMIT 1`).Scan(&userID)
	if err != nil {
		slog.Warn("Could not find root tenant owner for upgrade message", "error", err)
		return
	}

	var tenantID string
	database.Pool.QueryRow(ctx, "SELECT id FROM tenants WHERE is_root = true LIMIT 1").Scan(&tenantID)

	// Insert message
	_, err = database.Pool.Exec(ctx,
		`INSERT INTO messages (user_id, tenant_id, type, title, body, is_read, metadata)
		 VALUES ($1, $2, 'system', $3, $4, false, '{}')`,
		userID, tenantID,
		"Welcome to MyCourses v"+newVersion,
		"Your system has been upgraded to version "+newVersion+".")
	if err != nil {
		slog.Warn("Failed to send upgrade message", "error", err)
	}
}
