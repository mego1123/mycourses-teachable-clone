// Package handlers — SEO handler for sitemap and JSON-LD.
package handlers

import (
	"net/http"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
)

// SEOHandler handles SEO-related endpoints: sitemaps, robots.txt.
type SEOHandler struct {
	db             *db.DB
	platformDomain string
}

func NewSEOHandler(database *db.DB, platformDomain string) *SEOHandler {
	return &SEOHandler{db: database, platformDomain: platformDomain}
}

// Sitemap generates an XML sitemap listing all published courses.
func (h *SEOHandler) Sitemap(w http.ResponseWriter, r *http.Request) {
	// Get all published courses across all tenants
	courses, err := h.db.Queries.ListMarketplaceCourses(r.Context(), gen.ListMarketplaceCoursesParams{
		Limit:  1000,
		Offset: 0,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate sitemap")
		return
	}

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>`))
	w.Write([]byte(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`))

	for _, course := range courses {
		w.Write([]byte(`<url>`))
		w.Write([]byte(`<loc>https://` + h.platformDomain + `/courses/` + course.Slug + `</loc>`))
		w.Write([]byte(`<lastmod>` + course.UpdatedAt.Format("2006-01-02") + `</lastmod>`))
		w.Write([]byte(`<changefreq>weekly</changefreq>`))
		w.Write([]byte(`<priority>0.8</priority>`))
		w.Write([]byte(`</url>`))
	}

	w.Write([]byte(`</urlset>`))
}

// Robots returns a robots.txt that allows all crawlers and points to the sitemap.
func (h *SEOHandler) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\n"))
	w.Write([]byte("Allow: /\n"))
	w.Write([]byte("Disallow: /api/\n"))
	w.Write([]byte("Disallow: /studio/\n"))
	w.Write([]byte("Disallow: /learn/\n"))
	w.Write([]byte("\n"))
	w.Write([]byte("Sitemap: https://" + h.platformDomain + "/sitemap.xml\n"))
}
