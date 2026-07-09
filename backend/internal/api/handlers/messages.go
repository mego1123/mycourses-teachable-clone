package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"mycourses/internal/db"
	gen "mycourses/internal/db/gen"
	"mycourses/internal/middleware"
)

type MessageHandler struct{ db *db.DB }

func NewMessageHandler(database *db.DB, _ interface{}) *MessageHandler { return &MessageHandler{db: database} }

func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok { respondWithError(w, http.StatusUnauthorized, "Auth required"); return }

	page, limit := parsePagination(r)
	msgs, _ := h.db.Queries.ListMessagesByUser(r.Context(), gen.ListMessagesByUserParams{
		UserID: user.ID, Limit: int32(limit), Offset: int32((page - 1) * limit),
	})
	respondWithJSON(w, http.StatusOK, msgs)
}

func (h *MessageHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok { respondWithError(w, http.StatusUnauthorized, "Auth required"); return }

	msgID := parseUUID(mux.Vars(r)["id"])
	if msgID == nil { respondWithError(w, http.StatusBadRequest, "Invalid ID"); return }

	h.db.Queries.MarkMessageRead(r.Context(), gen.MarkMessageReadParams{ID: *msgID, UserID: user.ID})
	respondWithJSON(w, http.StatusOK, map[string]bool{"read": true})
}

func (h *MessageHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok { respondWithError(w, http.StatusUnauthorized, "Auth required"); return }

	count, _ := h.db.Queries.CountUnreadMessages(r.Context(), user.ID)
	respondWithJSON(w, http.StatusOK, map[string]int64{"count": count})
}
