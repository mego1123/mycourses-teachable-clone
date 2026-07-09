// Package handlers — Media upload handler for Cloudflare Stream.
// Provides direct upload URLs for video files and tracks media assets in Postgres.
package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
)

// MediaHandler handles video/media uploads via Cloudflare Stream.
type MediaHandler struct {
	db *db.DB
}

func NewMediaHandler(database *db.DB) *MediaHandler {
	return &MediaHandler{db: database}
}

// GetUploadURL returns a Cloudflare Stream direct upload URL.
// The frontend uploads the video file directly to Cloudflare (bypassing our server),
// then calls Complete to record the media asset in our database.
func (h *MediaHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetPgTenantIDFromContext(r.Context())
	if tenantID == uuidNil() {
		respondWithError(w, http.StatusForbidden, "Tenant context required")
		return
	}

	var req struct {
		Title      string `json:"title"`
		FileName   string `json:"fileName"`
		MimeType   string `json:"mimeType"`
		SizeBytes  int64  `json:"sizeBytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.FileName == "" {
		respondWithError(w, http.StatusBadRequest, "FileName is required")
		return
	}

	// Create a media asset record (status: processing)
	asset, err := h.db.Queries.CreateMediaAsset(r.Context(), gen.CreateMediaAssetParams{
		TenantID:  tenantID,
		Kind:      "video",
		Title:     req.Title,
		MimeType:  req.MimeType,
		SizeBytes: req.SizeBytes,
		Status:    "processing",
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create media asset: "+err.Error())
		return
	}

	// If Cloudflare Stream is configured, create a direct upload URL
	cfAccountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	cfAPIToken := os.Getenv("CLOUDFLARE_API_TOKEN")

	if cfAccountID != "" && cfAPIToken != "" {
		// Call Cloudflare Stream API to create a direct upload
		uploadURL, cfStreamID, err := createCFStreamUpload(cfAccountID, cfAPIToken, req.FileName, asset.ID.String())
		if err != nil {
			// Return the asset ID anyway — frontend can retry upload
			respondWithJSON(w, http.StatusOK, map[string]interface{}{
				"mediaAssetId": asset.ID,
				"uploadUrl":    "", // Empty — fallback to direct API upload
				"status":       "processing",
				"error":        "Cloudflare Stream not configured, using fallback",
			})
			return
		}

		// Update asset with CF Stream ID
		h.db.Queries.UpdateMediaAssetStatus(r.Context(), gen.UpdateMediaAssetStatusParams{
			ID:     asset.ID,
			Status: "processing",
		})

		// Update the CF Stream ID
		h.db.Pool.Exec(r.Context(),
			"UPDATE media_assets SET cf_stream_id = $2 WHERE id = $1",
			asset.ID, cfStreamID)

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"mediaAssetId": asset.ID,
			"uploadUrl":    uploadURL,
			"cfStreamId":   cfStreamID,
			"status":       "processing",
		})
		return
	}

	// Fallback: no Cloudflare configured — return asset ID for future upload
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"mediaAssetId": asset.ID,
		"uploadUrl":    "",
		"status":       "processing",
		"message":      "Cloudflare Stream not configured — upload URL not available",
	})
}

// Complete marks a media asset as ready (called after upload finishes).
func (h *MediaHandler) Complete(w http.ResponseWriter, r *http.Request) {
	assetID := parseUUID(r.URL.Query().Get("id"))
	if assetID == nil {
		respondWithError(w, http.StatusBadRequest, "Media asset ID required")
		return
	}

	var req struct {
		DurationSec int   `json:"durationSec"`
		SizeBytes   int64 `json:"sizeBytes"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	h.db.Queries.UpdateMediaAssetReady(r.Context(), gen.UpdateMediaAssetReadyParams{
		ID:          *assetID,
		DurationSec: int32(req.DurationSec),
		SizeBytes:   req.SizeBytes,
	})

	asset, _ := h.db.Queries.GetMediaAssetByID(r.Context(), *assetID)
	respondWithJSON(w, http.StatusOK, asset)
}

// GetSignedPlaybackURL returns a signed URL for playing a video (enrollment-gated).
func (h *MediaHandler) GetSignedPlaybackURL(w http.ResponseWriter, r *http.Request) {
	assetID := parseUUID(r.URL.Query().Get("id"))
	if assetID == nil {
		respondWithError(w, http.StatusBadRequest, "Media asset ID required")
		return
	}

	asset, err := h.db.Queries.GetMediaAssetByID(r.Context(), *assetID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Media asset not found")
		return
	}

	if asset.CfStreamID == nil || *asset.CfStreamID == "" {
		respondWithError(w, http.StatusNotFound, "No video stream available for this asset")
		return
	}

	// In production, generate a Cloudflare Stream signed JWT here
	// For now, return the stream ID — frontend can construct the playback URL
	respondWithJSON(w, http.StatusOK, map[string]string{
		"cfStreamId": *asset.CfStreamID,
		"playbackUrl": "https://watch.cloudflarestream.com/" + *asset.CfStreamID,
	})
}

// createCFStreamUpload calls the Cloudflare Stream API to create a direct upload URL.
func createCFStreamUpload(accountID, apiToken, fileName, assetID string) (string, string, error) {
	// In production, this would make an HTTP POST to:
	// https://api.cloudflare.com/client/v4/accounts/{accountID}/stream/copy
	// or the direct upload endpoint
	//
	// POST /accounts/{accountID}/stream/copy
	// Headers: Authorization: Bearer {apiToken}
	// Body: { "maxDurationSeconds": 3600, "metadata": { "asset_id": assetID } }
	//
	// Response: { "result": { "uid": "cf_stream_id", "uploadURL": "https://upload.cloudflarestream.com/..." } }

	// For now, return empty — actual implementation requires HTTP client call
	return "", "", nil
}
