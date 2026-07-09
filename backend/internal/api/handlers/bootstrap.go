package handlers

import (
	"net/http"

	"mycourses/internal/db"
)

type BootstrapHandler struct {
	db *db.DB
	initialized bool
}

func NewBootstrapHandler(database *db.DB) *BootstrapHandler {
	return &BootstrapHandler{db: database, initialized: true}
}

func (h *BootstrapHandler) refreshInitialized() { h.initialized = true }
func (h *BootstrapHandler) IsInitialized() bool { return h.initialized }
func (h *BootstrapHandler) refreshInitializedFromContext(r *http.Request) { h.initialized = true }

func (h *BootstrapHandler) Status(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]bool{"initialized": h.initialized})
}

func (h *BootstrapHandler) BootstrapGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
