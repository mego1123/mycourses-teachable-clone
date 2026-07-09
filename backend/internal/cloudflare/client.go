// Package cloudflare — Cloudflare API client for custom domain management.
// Uses Cloudflare for SaaS to provision custom domains with automatic SSL.
package cloudflare

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Client manages Cloudflare for SaaS custom hostnames.
type Client struct {
	apiToken   string
	accountID  string
	zoneID     string
	httpClient *http.Client
}

// New creates a Cloudflare client from environment variables.
func New() *Client {
	return &Client{
		apiToken:   os.Getenv("CLOUDFLARE_API_TOKEN"),
		accountID:  os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
		zoneID:     os.Getenv("CLOUDFLARE_ZONE_ID"),
		httpClient: &http.Client{},
	}
}

// IsConfigured returns true if the Cloudflare credentials are set.
func (c *Client) IsConfigured() bool {
	return c.apiToken != "" && c.zoneID != ""
}

// CustomHostname represents a Cloudflare for SaaS custom hostname.
type CustomHostname struct {
	ID          string `json:"id"`
	Hostname    string `json:"hostname"`
	SSLStatus   string `json:"ssl_status"`
	Status      string `json:"status"`
	Verification struct {
		Name  string `json:"name"`
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"verification"`
	SSL struct {
		Status      string `json:"status"`
		Method      string `json:"method"`
		Validator   string `json:"validator"`
		CnameTarget string `json:"cname_target"`
	} `json:"ssl"`
}

// CreateCustomHostname provisions a custom hostname via Cloudflare for SaaS.
func (c *Client) CreateCustomHostname(hostname string) (*CustomHostname, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("cloudflare not configured")
	}

	body := map[string]interface{}{
		"hostname": hostname,
		"ssl": map[string]interface{}{
			"method": "http",
		},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST",
		fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/custom_hostnames", c.zoneID),
		bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create custom hostname: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool             `json:"success"`
		Errors  []map[string]any `json:"errors"`
		Result  *CustomHostname  `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("cloudflare API error: %v", result.Errors)
	}

	return result.Result, nil
}

// GetCustomHostname retrieves the status of a custom hostname.
func (c *Client) GetCustomHostname(hostnameID string) (*CustomHostname, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("cloudflare not configured")
	}

	req, _ := http.NewRequest("GET",
		fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/custom_hostnames/%s", c.zoneID, hostnameID),
		nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom hostname: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success bool            `json:"success"`
		Result  *CustomHostname `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}

// DeleteCustomHostname removes a custom hostname.
func (c *Client) DeleteCustomHostname(hostnameID string) error {
	if !c.IsConfigured() {
		return fmt.Errorf("cloudflare not configured")
	}

	req, _ := http.NewRequest("DELETE",
		fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/custom_hostnames/%s", c.zoneID, hostnameID),
		nil)
	req.Header.Set("Authorization", "Bearer "+c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete custom hostname: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
