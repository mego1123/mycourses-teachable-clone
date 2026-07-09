package models

import (
	"github.com/google/uuid"
	"time"

)

// NavItem represents a navigation item in the app sidebar.
type NavItem struct {
	ID              string `json:"id"`
	Label           string `json:"label"`
	Icon            string `json:"icon"`
	Target          string `json:"target"`                                       // internal route or "/p/slug"
	EntitlementGate string `json:"entitlementGate,omitempty"` // entitlement key required
	IsBuiltIn       bool   `json:"isBuiltIn"`
	Visible         bool   `json:"visible"`
	SortOrder       int    `json:"sortOrder"`
}

// BrandingConfig stores the global branding settings.
type BrandingConfig struct {
	ID uuid.UUID `json:"id"`

	// Identity
	AppName  string `json:"appName" validate:"required,min=1,max=200"`
	Tagline  string `json:"tagline"`
	LogoMode string `json:"logoMode" validate:"omitempty,valid_logo_mode"` // "text", "image", "both"

	// Theme
	PrimaryColor    string `json:"primaryColor"`
	AccentColor     string `json:"accentColor"`
	BackgroundColor string `json:"backgroundColor"`
	SurfaceColor    string `json:"surfaceColor"`
	TextColor       string `json:"textColor"`
	FontFamily      string `json:"fontFamily"`
	HeadingFont     string `json:"headingFont"`

	// Landing page
	LandingEnabled     bool   `json:"landingEnabled"`
	LandingTitle       string `json:"landingTitle"`
	LandingMeta        string `json:"landingMeta"`
	LandingHTML        string `json:"landingHtml"`

	// Dashboard
	DashboardHTML string `json:"dashboardHtml"`

	// Auth pages
	LoginHeading   string `json:"loginHeading"`
	LoginSubtext   string `json:"loginSubtext"`
	SignupHeading  string `json:"signupHeading"`
	SignupSubtext  string `json:"signupSubtext"`

	// Custom head/CSS
	CustomCSS string `json:"customCss"`
	HeadHTML  string `json:"headHtml"`

	// Social
	OgImageURL string `json:"ogImageUrl"`

	// Navigation
	NavItems []NavItem `json:"navItems"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BrandingAsset stores binary assets (logo, favicon, etc.).
type BrandingAsset struct {
	ID          uuid.UUID `json:"id"`
	Key         string             `json:"key"`                 // "logo", "favicon", or unique media ID
	Filename    string             `json:"filename"`
	ContentType string             `json:"contentType"`
	Data        []byte             `json:"-"`
	Size        int64              `json:"size"`
	CreatedAt   time.Time          `json:"createdAt"`
}

// CustomPage stores user-created pages with arbitrary HTML content.
type CustomPage struct {
	ID              uuid.UUID `json:"id"`
	Slug            string             `json:"slug" validate:"required,min=1,max=200"`
	Title           string             `json:"title" validate:"required,min=1,max=200"`
	HTMLBody        string             `json:"htmlBody"`
	MetaDescription string             `json:"metaDescription"`
	OgImage         string             `json:"ogImage"`
	IsPublished     bool               `json:"isPublished"`
	SortOrder       int                `json:"sortOrder"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
}
