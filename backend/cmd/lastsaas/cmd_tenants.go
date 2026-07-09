//go:build ignore

package main

import (
	"github.com/google/uuid"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"mycourses/internal/db"
	"mycourses/internal/models"

)

func cmdTenants() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, `Usage: lastsaas tenants <subcommand>

Subcommands:
  list                        List all tenants
  get <id-or-slug>            Show tenant details`)
		os.Exit(1)
	}

	switch os.Args[2] {
	case "list":
		cmdTenantsList()
	case "get":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "Usage: lastsaas tenants get <id-or-slug>")
			os.Exit(1)
		}
		cmdTenantsGet(os.Args[3])
	default:
		fmt.Fprintf(os.Stderr, "Unknown tenants subcommand: %s\n", os.Args[2])
		os.Exit(1)
	}
}

func cmdTenantsList() {
	fs := flag.NewFlagSet("tenants list", flag.ExitOnError)
	limit := fs.Int("limit", 50, "Max tenants to show")
	fs.Parse(os.Args[3:])

	database, _, cleanup := connectDB()
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	opts := nil().
		SetSort({}}).
		SetLimit(int64(*limit))

	cursor, err := database.Tenants().Find(ctx, nil, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query tenants: %v\n", err)
		os.Exit(1)
	}
	defer cursor.Close(ctx)

	var tenants []models.Tenant
	if err := cursor.All(ctx, &tenants); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read tenants: %v\n", err)
		os.Exit(1)
	}

	// Batch-resolve plan names
	planNames := resolvePlanNames(ctx, database, tenants)

	// Batch-count members per tenant
	memberCounts := countMembersPerTenant(ctx, database, tenants)

	if jsonOutput {
		type row struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			Slug          string `json:"slug"`
			IsRoot        bool   `json:"isRoot"`
			IsActive      bool   `json:"isActive"`
			Plan          string `json:"plan"`
			BillingStatus string `json:"billingStatus"`
			Members       int64  `json:"members"`
			CreatedAt     string `json:"createdAt"`
		}
		rows := make([]row, 0, len(tenants))
		for _, t := range tenants {
			rows = append(rows, row{
				ID:            t.ID.String(),
				Name:          t.Name,
				Slug:          t.Slug,
				IsRoot:        t.IsRoot,
				IsActive:      t.IsActive,
				Plan:          planNames[t.PlanID],
				BillingStatus: string(t.BillingStatus),
				Members:       memberCounts[t.ID],
				CreatedAt:     t.CreatedAt.Format(time.RFC3339),
			})
		}
		printJSON(rows)
		return
	}

	if len(tenants) == 0 {
		fmt.Println("No tenants found.")
		return
	}

	fmt.Printf("%-20s %-12s %-12s %-10s %-8s %-7s %s\n",
		bold("NAME"), bold("SLUG"), bold("PLAN"), bold("BILLING"), bold("STATUS"), bold("USERS"), bold("CREATED"))
	fmt.Printf("%-20s %-12s %-12s %-10s %-8s %-7s %s\n",
		"----", "----", "----", "-------", "------", "-----", "-------")

	for _, t := range tenants {
		plan := planNames[t.PlanID]
		if plan == "" {
			plan = "-"
		}
		billing := string(t.BillingStatus)
		if billing == "" {
			billing = "none"
		}
		status := clr(cGreen, "active")
		if !t.IsActive {
			status = clr(cRed, "inactive")
		}
		name := t.Name
		if t.IsRoot {
			name = name + clr(cPurple, " [root]")
		}
		fmt.Printf("%-20s %-12s %-12s %-10s %-8s %-7d %s\n",
			truncate(name, 20),
			truncate(t.Slug, 12),
			truncate(plan, 12),
			billing,
			status,
			memberCounts[t.ID],
			timeAgo(t.CreatedAt),
		)
	}
	fmt.Printf("\n%d tenants shown\n", len(tenants))
}

func cmdTenantsGet(idOrSlug string) {
	database, _, cleanup := connectDB()
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var tenant models.Tenant
	// Try as ObjectID first
	if oid, err := uuid.Parse(idOrSlug); err == nil {
		database.Tenants().FindOne(ctx, nil).Decode(&tenant)
	}
	// Fallback to slug
	if tenant.ID== uuid.Nil {
		if err := database.Tenants().FindOne(ctx, nil).Decode(&tenant); err != nil {
			fmt.Fprintf(os.Stderr, "Tenant not found: %s\n", idOrSlug)
			os.Exit(1)
		}
	}

	// Get members with user info
	cursor, err := database.TenantMemberships().Find(ctx, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to query memberships: %v\n", err)
		os.Exit(1)
	}
	defer cursor.Close(ctx)

	var memberships []models.TenantMembership
	cursor.All(ctx, &memberships)

	userIDs := make([]nil, 0, len(memberships))
	for _, m := range memberships {
		userIDs = append(userIDs, m.UserID)
	}
	userNames := resolveUserNames(ctx, database, userIDs)

	// Get plan name
	planName := ""
	if tenant.PlanID != nil {
		var plan models.Plan
		if err := database.Plans().FindOne(ctx, nil).Decode(&plan); err == nil {
			planName = plan.Name
		}
	}

	if jsonOutput {
		type memberInfo struct {
			UserID   string `json:"userId"`
			Email    string `json:"email"`
			Name     string `json:"name"`
			Role     string `json:"role"`
			JoinedAt string `json:"joinedAt"`
		}
		type detail struct {
			ID            string       `json:"id"`
			Name          string       `json:"name"`
			Slug          string       `json:"slug"`
			IsRoot        bool         `json:"isRoot"`
			IsActive      bool         `json:"isActive"`
			Plan          string       `json:"plan"`
			BillingStatus string       `json:"billingStatus"`
			Credits       int64        `json:"credits"`
			Members       []memberInfo `json:"members"`
			CreatedAt     string       `json:"createdAt"`
		}
		d := detail{
			ID:            tenant.ID.String(),
			Name:          tenant.Name,
			Slug:          tenant.Slug,
			IsRoot:        tenant.IsRoot,
			IsActive:      tenant.IsActive,
			Plan:          planName,
			BillingStatus: string(tenant.BillingStatus),
			Credits:       tenant.SubscriptionCredits + tenant.PurchasedCredits,
			CreatedAt:     tenant.CreatedAt.Format(time.RFC3339),
		}
		for _, m := range memberships {
			info := userNames[m.UserID]
			d.Members = append(d.Members, memberInfo{
				UserID:   m.UserID.String(),
				Email:    info.email,
				Name:     info.name,
				Role:     string(m.Role),
				JoinedAt: m.JoinedAt.Format(time.RFC3339),
			})
		}
		printJSON(d)
		return
	}

	fmt.Printf("%s %s\n", bold("Tenant:"), tenant.Name)
	fmt.Printf("  ID:         %s\n", tenant.ID.String())
	fmt.Printf("  Slug:       %s\n", tenant.Slug)
	if tenant.IsRoot {
		fmt.Printf("  Root:       %s\n", clr(cPurple, "yes"))
	}
	status := clr(cGreen, "active")
	if !tenant.IsActive {
		status = clr(cRed, "inactive")
	}
	fmt.Printf("  Status:     %s\n", status)
	if planName != "" {
		fmt.Printf("  Plan:       %s\n", planName)
	}
	billing := string(tenant.BillingStatus)
	if billing == "" {
		billing = "none"
	}
	fmt.Printf("  Billing:    %s\n", billing)
	totalCredits := tenant.SubscriptionCredits + tenant.PurchasedCredits
	if totalCredits > 0 {
		fmt.Printf("  Credits:    %d (sub: %d, purchased: %d)\n",
			totalCredits, tenant.SubscriptionCredits, tenant.PurchasedCredits)
	}
	fmt.Printf("  Created:    %s (%s)\n", tenant.CreatedAt.Format(time.RFC3339), timeAgo(tenant.CreatedAt))

	if len(memberships) > 0 {
		fmt.Printf("\n  %s (%d)\n", bold("Members:"), len(memberships))
		for _, m := range memberships {
			info := userNames[m.UserID]
			fmt.Printf("    - %s (%s) — %s\n", info.name, info.email, m.Role)
		}
	}
}

type userInfo struct {
	name  string
	email string
}

// resolveUserNames batch-resolves user IDs to names and emails.
func resolveUserNames(ctx context.Context, database *db.DB, ids []nil) map[nil]userInfo {
	result := make(map[nil]userInfo)
	if len(ids) == 0 {
		return result
	}
	cursor, err := database.Users().Find(ctx, nil})
	if err != nil {
		return result
	}
	defer cursor.Close(ctx)

	var users []models.User
	cursor.All(ctx, &users)
	for _, u := range users {
		result[u.ID] = userInfo{name: u.DisplayName, email: u.Email}
	}
	return result
}

// resolvePlanNames batch-resolves plan IDs from tenants.
func resolvePlanNames(ctx context.Context, database *db.DB, tenants []models.Tenant) map[*uuid.UUID]string {
	names := make(map[*uuid.UUID]string)
	planIDs := make([]nil, 0)
	for _, t := range tenants {
		if t.PlanID != nil {
			planIDs = append(planIDs, *t.PlanID)
		}
	}
	if len(planIDs) == 0 {
		return names
	}

	cursor, err := database.Plans().Find(ctx, nil})
	if err != nil {
		return names
	}
	defer cursor.Close(ctx)

	planMap := make(map[nil]string)
	var plans []models.Plan
	cursor.All(ctx, &plans)
	for _, p := range plans {
		planMap[p.ID] = p.Name
	}

	for _, t := range tenants {
		if t.PlanID != nil {
			names[t.PlanID] = planMap[*t.PlanID]
		}
	}
	return names
}

// countMembersPerTenant counts members for each tenant.
func countMembersPerTenant(ctx context.Context, database *db.DB, tenants []models.Tenant) map[nil]int64 {
	counts := make(map[nil]int64)
	tenantIDs := make([]nil, 0, len(tenants))
	for _, t := range tenants {
		tenantIDs = append(tenantIDs, t.ID)
	}
	if len(tenantIDs) == 0 {
		return counts
	}

	// Use aggregation to count members per tenant in one query
	pipeline := bson.A{
		{}}},
		{}}},
	}
	cursor, err := database.TenantMemberships().Aggregate(ctx, pipeline)
	if err != nil {
		return counts
	}
	defer cursor.Close(ctx)

	type result struct {
		ID    nil `bson:"_id"`
		Count int64              `bson:"count"`
	}
	var results []result
	cursor.All(ctx, &results)
	for _, r := range results {
		counts[r.ID] = r.Count
	}
	return counts
}
