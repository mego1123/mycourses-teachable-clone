package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
)

type CertificateHandler struct{ db *db.DB }

func NewCertificateHandler(database *db.DB) *CertificateHandler { return &CertificateHandler{db: database} }

func (h *CertificateHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r)
	if userID == nil { respondWithError(w, http.StatusUnauthorized, "Auth required"); return }
	certs, err := h.db.Queries.ListCertificatesByUser(r.Context(), *userID)
	if err != nil { respondWithError(w, http.StatusInternalServerError, "Failed"); return }
	respondWithJSON(w, http.StatusOK, certs)
}

func (h *CertificateHandler) Verify(w http.ResponseWriter, r *http.Request) {
	token := mux.Vars(r)["token"]
	cert, err := h.db.Queries.GetCertificateByToken(r.Context(), token)
	if err != nil { respondWithError(w, http.StatusNotFound, "Certificate not found or revoked"); return }
	respondWithJSON(w, http.StatusOK, map[string]interface{}{"valid": true, "certificate": cert})
}
