// Package handlers — Custom domain handler with Cloudflare for SaaS integration.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/cloudflare"
	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
)

// CustomDomainHandler handles creator custom domain management.
type CustomDomainHandler struct {
	db     *db.DB
	cf     *cloudflare.Client
}

func NewCustomDomainHandler(database *db.DB, cfClient *cloudflare.Client) *CustomDomainHandler {
	return &CustomDomainHandler{db: database, cf: cfClient}
}

// Create adds a custom domain and provisions it via Cloudflare for SaaS.
func (h *CustomDomainHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetPgTenantFromContext(r.Context())
	if tenant == nil {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	var req struct{ Domain string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.Domain == "" {
		respondWithError(w, http.StatusBadRequest, "Domain is required")
		return
	}

	// Create record in database first
	verificationRecords := []byte(`{}`)
	domain, err := h.db.Queries.CreateCustomDomain(r.Context(), gen.CreateCustomDomainParams{
		Domain:              req.Domain,
		TenantID:            tenant.ID,
		VerificationRecords: verificationRecords,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create custom domain: "+err.Error())
		return
	}

	// If Cloudflare is configured, provision the custom hostname
	if h.cf != nil && h.cf.IsConfigured() {
		cfHostname, err := h.cf.CreateCustomHostname(req.Domain)
		if err != nil {
			// Log error but don't fail — domain is in pending status
			respondWithJSON(w, http.StatusCreated, map[string]interface{}{
				"domain":  domain,
				"warning": "Cloudflare provisioning failed: " + err.Error(),
			})
			return
		}

		// Update with Cloudflare hostname ID and verification records
		cfRecords := map[string]string{
			"cf_hostname_id": cfHostname.ID,
			"cname_target":   cfHostname.SSL.CnameTarget,
			"verification_name":  cfHostname.Verification.Name,
			"verification_value": cfHostname.Verification.Value,
		}
		cfRecordsJSON, _ := json.Marshal(cfRecords)

		h.db.Queries.UpdateCustomDomainStatus(r.Context(), gen.UpdateCustomDomainStatusParams{
			ID:        domain.ID,
			Status:    "pending",
			DnsVerified: false,
			SslStatus: cfHostname.SSL.Status,
			CfHostnameID: &cfHostname.ID,
		})

		// Update verification records
		h.db.Pool.Exec(r.Context(),
			"UPDATE custom_domains SET verification_records = $2 WHERE id = $1",
			domain.ID, cfRecordsJSON)

		domain.CfHostnameID = &cfHostname.ID
		domain.SslStatus = cfHostname.SSL.Status
	}

	respondWithJSON(w, http.StatusCreated, domain)
}

// List returns all custom domains for the creator's tenant.
func (h *CustomDomainHandler) List(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.GetPgTenantFromContext(r.Context())
	if tenant == nil {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	domains, err := h.db.Queries.ListCustomDomainsByTenant(r.Context(), tenant.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list custom domains")
		return
	}

	// If Cloudflare is configured, sync live status
	if h.cf != nil && h.cf.IsConfigured() {
		for i, d := range domains {
			if d.CfHostnameID != nil && *d.CfHostnameID != "" {
				cfHostname, err := h.cf.GetCustomHostname(*d.CfHostnameID)
				if err == nil {
					// Update status if changed
					if cfHostname.SSL.Status == "active" && d.SslStatus != "active" {
						h.db.Queries.UpdateCustomDomainStatus(r.Context(), gen.UpdateCustomDomainStatusParams{
							ID:          d.ID,
							Status:      "active",
							DnsVerified:  true,
							SslStatus:   "active",
							CfHostnameID: d.CfHostnameID,
						})
						domains[i].Status = "active"
						domains[i].DnsVerified = true
						domains[i].SslStatus = "active"
					}
				}
			}
		}
	}

	respondWithJSON(w, http.StatusOK, domains)
}

// Delete removes a custom domain and deprovisions it from Cloudflare.
func (h *CustomDomainHandler) Delete(w http.ResponseWriter, r *http.Request) {
	domainID := parseUUID(mux.Vars(r)["id"])
	if domainID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	// Get domain to find CF hostname ID
	domain, err := h.db.Queries.GetCustomDomainByID(r.Context(), *domainID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Custom domain not found")
		return
	}

	// Deprovision from Cloudflare if configured
	if h.cf != nil && h.cf.IsConfigured() && domain.CfHostnameID != nil && *domain.CfHostnameID != "" {
		h.cf.DeleteCustomHostname(*domain.CfHostnameID)
	}

	// Delete from database
	if err := h.db.Queries.DeleteCustomDomain(r.Context(), *domainID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete custom domain")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

// GetStatus returns the live status of a custom domain (polled by frontend).
func (h *CustomDomainHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	domainID := parseUUID(mux.Vars(r)["id"])
	if domainID == nil {
		respondWithError(w, http.StatusBadRequest, "Invalid domain ID")
		return
	}

	domain, err := h.db.Queries.GetCustomDomainByID(r.Context(), *domainID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Custom domain not found")
		return
	}

	// Sync with Cloudflare if configured
	if h.cf != nil && h.cf.IsConfigured() && domain.CfHostnameID != nil && *domain.CfHostnameID != "" {
		cfHostname, err := h.cf.GetCustomHostname(*domain.CfHostnameID)
		if err == nil {
			if cfHostname.SSL.Status == "active" && domain.SslStatus != "active" {
				h.db.Queries.UpdateCustomDomainStatus(r.Context(), gen.UpdateCustomDomainStatusParams{
					ID:          domain.ID,
					Status:      "active",
					DnsVerified:  true,
					SslStatus:   "active",
					CfHostnameID: domain.CfHostnameID,
				})
				domain.Status = "active"
				domain.DnsVerified = true
				domain.SslStatus = "active"
			}
		}
	}

	respondWithJSON(w, http.StatusOK, domain)
}
