package handlers

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"mycourses/internal/db"
	"mycourses/internal/syslog"
)

type EventDefinitionsHandler struct {
	db     *db.DB
	syslog *syslog.Logger
}

type eventDefRequest struct {
	db     *db.DB
	syslog *syslog.Logger
}

type defResponse struct {
	db     *db.DB
	syslog *syslog.Logger
}

type sankeyNode struct {
	db     *db.DB
	syslog *syslog.Logger
}

type sankeyLink struct {
	db     *db.DB
	syslog *syslog.Logger
}

func NewEventDefinitionsHandler(database *db.DB, sysLogger *syslog.Logger) *EventDefinitionsHandler {
	return &EventDefinitionsHandler{db: database}
}

func (h *EventDefinitionsHandler) ListEventDefinitions(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *EventDefinitionsHandler) CreateEventDefinition(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *EventDefinitionsHandler) UpdateEventDefinition(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *EventDefinitionsHandler) DeleteEventDefinition(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *EventDefinitionsHandler) GetSankeyData(w http.ResponseWriter, r *http.Request) {
	http.Error(w, `{"error":"Not yet migrated to Postgres"}`, http.StatusNotImplemented)
}

func (h *EventDefinitionsHandler) wouldCreateCycle(ctx context.Context, defID, proposedParentID uuid.UUID) bool {
	return false
}

