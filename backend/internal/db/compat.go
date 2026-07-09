// Package db — MongoDB compatibility layer.
// Provides MongoDB-style collection accessor methods on *DB that delegate to
// direct SQL via Pool.Exec/Query. This is a transitional layer to allow
// existing handlers to compile during the MongoDB → Postgres migration.
// Each method returns a *CompatCollection that wraps the Pool + table name.
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CompatCollection wraps a Postgres table for MongoDB-style operations.
type CompatCollection struct {
	pool   *pgxpool.Pool
	table  string
}

// InsertOne inserts a single row. The data parameter must be a struct with json tags.
// This is a simplified compatibility method — production code should use sqlc queries.
func (c *CompatCollection) InsertOne(ctx context.Context, data interface{}) (interface{}, error) {
	return nil, fmt.Errorf("InsertOne: migrate to sqlc queries")

}

// FindOne finds a single row. Currently returns an error (not implemented).
// Existing code should be migrated to use sqlc queries instead.
func (c *CompatCollection) FindOne(ctx context.Context, filter interface{}) *CompatSingleResult {
	return &CompatSingleResult{pool: c.pool, table: c.table, err: fmt.Errorf("FindOne not supported in compatibility layer — use sqlc queries")}
}

// Find finds multiple rows. Currently returns an error.
func (c *CompatCollection) Find(ctx context.Context, filter interface{}) (*CompatCursor, error) {
	return nil, fmt.Errorf("Find not supported in compatibility layer — use sqlc queries")
}

// CountDocuments counts rows. Currently returns 0.
func (c *CompatCollection) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	return 0, fmt.Errorf("CountDocuments not supported in compatibility layer")
}

// UpdateOne updates a single row. Currently returns an error.
func (c *CompatCollection) UpdateOne(ctx context.Context, filter, update interface{}) (interface{}, error) {
	return nil, fmt.Errorf("UpdateOne: migrate to sqlc queries")
}

// CompatSingleResult wraps a FindOne result.
type CompatSingleResult struct {
	pool  *pgxpool.Pool
	table string
	err   error
}

func (r *CompatSingleResult) Decode(v interface{}) error {
	return r.err
}

func (r *CompatSingleResult) Err() error {
	return r.err
}

// CompatCursor wraps a Find result.
type CompatCursor struct {
	rows pgx.Rows
	err  error
}

func (c *CompatCursor) All(ctx context.Context, v interface{}) error {
	return c.err
}

func (c *CompatCursor) Close(ctx context.Context) {}

// Collection accessor methods — these exist on db.MongoDB and are called by
// existing handlers. We provide them on db.DB as a compatibility layer.
// Each returns a CompatCollection that will error on use, forcing migration.

func (db *DB) Users() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "users"} }
func (db *DB) Tenants() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "tenants"} }
func (db *DB) TenantMemberships() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "tenant_memberships"} }
func (db *DB) Plans() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "plans"} }
func (db *DB) FinancialTransactions() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "financial_transactions"} }
func (db *DB) RefreshTokens() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "refresh_tokens"} }
func (db *DB) VerificationTokens() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "verification_tokens"} }
func (db *DB) OAuthStates() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "oauth_states"} }
func (db *DB) RevokedTokens() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "revoked_tokens"} }
func (db *DB) APIKeys() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "api_keys"} }
func (db *DB) Webhooks() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "webhooks"} }
func (db *DB) WebhookDeliveries() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "webhook_deliveries"} }
func (db *DB) BrandingConfig() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "branding_configs"} }
func (db *DB) BrandingAssets() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "branding_assets"} }
func (db *DB) CustomPages() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "custom_pages"} }
func (db *DB) Invitations() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "invitations"} }
func (db *DB) AuditLog() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "audit_log"} }
func (db *DB) SystemLogs() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "system_logs"} }
func (db *DB) SystemConfig() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "config_vars"} }
func (db *DB) SystemMetrics() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "system_metrics"} }
func (db *DB) DailyMetrics() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "daily_metrics"} }
func (db *DB) LeaderLocks() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "leader_locks"} }
func (db *DB) EventDefinitions() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "event_definitions"} }
func (db *DB) TelemetryEvents() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "telemetry_events"} }
func (db *DB) UsageEvents() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "usage_events"} }
func (db *DB) Messages() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "messages"} }
func (db *DB) Announcements() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "announcements"} }
func (db *DB) ConfigVars() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "config_vars"} }
func (db *DB) StripeMappings() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "stripe_mappings"} }
func (db *DB) CreditBundles() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "credit_bundles"} }
func (db *DB) Promotions() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "promotions"} }
func (db *DB) PromotionCodes() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "promotion_codes"} }
func (db *DB) ImpersonationLogs() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "impersonation_logs"} }
func (db *DB) WebauthnCredentials() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "webauthn_credentials"} }
func (db *DB) WebauthnSessions() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "webauthn_sessions"} }
func (db *DB) AuthCodes() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "auth_codes"} }
func (db *DB) SSOConnections() *CompatCollection { return &CompatCollection{pool: db.Pool, table: "sso_connections"} }
