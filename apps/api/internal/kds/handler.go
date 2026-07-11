package kds

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/auth"
	"github.com/smithgh-ops/restaurant-multibranch-qris/apps/api/internal/middleware"
)

// Handler groups HTTP handlers for KDS endpoints.
type Handler struct {
	repo           *Repository
	hub            *Hub
	jwtSecret      string
	allowedOrigins map[string]struct{}
}

// NewHandler creates a new KDS handler.
func NewHandler(repo *Repository, hub *Hub, jwtSecret, corsOrigins string) *Handler {
	allowedOrigins := make(map[string]struct{})
	for _, origin := range strings.Split(corsOrigins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowedOrigins[trimmed] = struct{}{}
		}
	}
	return &Handler{
		repo:           repo,
		hub:            hub,
		jwtSecret:      jwtSecret,
		allowedOrigins: allowedOrigins,
	}
}

// ListStations handles GET /api/v1/branches/:branch_id/kds/stations.
func (h *Handler) ListStations(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok || !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	stations, err := h.repo.ListStations(c.Request.Context(), branchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stations})
}

// CreateStation handles POST /api/v1/branches/:branch_id/kds/stations.
func (h *Handler) CreateStation(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok || !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	var req CreateStationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	station, err := h.repo.CreateStation(c.Request.Context(), branchID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, station)
}

// ListTickets handles GET /api/v1/branches/:branch_id/kds/tickets.
func (h *Handler) ListTickets(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok || !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}

	filter := TicketListFilter{}
	if rawStationID := c.Query("station_id"); rawStationID != "" {
		stationID, err := strconv.ParseUint(rawStationID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "station_id tidak valid"})
			return
		}
		filter.StationID = &stationID
	}
	if rawStatus := c.Query("status"); rawStatus != "" {
		if _, ok := validTicketStatuses[rawStatus]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "status ticket tidak valid"})
			return
		}
		filter.Status = &rawStatus
	}

	tickets, err := h.repo.ListTickets(c.Request.Context(), branchID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tickets})
}

// UpdateTicketStatus handles PATCH /api/v1/branches/:branch_id/kds/tickets/:id/status.
func (h *Handler) UpdateTicketStatus(c *gin.Context) {
	branchID, orgID, ok := h.parseBranchParam(c)
	if !ok || !h.branchBelongsToOrg(c, branchID, orgID) {
		return
	}
	ticketID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id tidak valid"})
		return
	}
	var req UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data tidak valid: " + err.Error()})
		return
	}
	if _, ok := validTicketStatuses[req.Status]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status ticket tidak valid"})
		return
	}

	ticket, err := h.repo.UpdateTicketStatus(c.Request.Context(), ticketID, branchID, req.Status)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket dapur tidak ditemukan"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	_ = h.hub.Broadcast(branchID, RealtimeEvent{Type: "ticket.updated", Ticket: ticket})
	c.JSON(http.StatusOK, ticket)
}

// ServeWS handles GET /api/v1/kds/ws?branch_id=X&token=Y.
func (h *Handler) ServeWS(c *gin.Context) {
	branchID, err := strconv.ParseUint(c.Query("branch_id"), 10, 64)
	if err != nil || branchID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
		return
	}
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token tidak ditemukan"})
		return
	}
	claims, err := auth.ParseAccessToken(h.jwtSecret, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token tidak valid atau sudah kadaluarsa"})
		return
	}
	if !h.branchBelongsToOrganization(c, branchID, claims.OrganizationID) {
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				return true
			}
			_, ok := h.allowedOrigins[origin]
			return ok
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	h.hub.ServeConnection(branchID, conn)
}

// RegisterRoutes mounts KDS endpoints.
func RegisterRoutes(v1 *gin.RouterGroup, h *Handler, authMiddleware gin.HandlerFunc) {
	branches := v1.Group("/branches/:branch_id/kds", authMiddleware)
	{
		branches.GET("/stations", h.ListStations)
		branches.POST("/stations", h.CreateStation)
		branches.GET("/tickets", h.ListTickets)
		branches.PATCH("/tickets/:id/status", h.UpdateTicketStatus)
	}
	v1.GET("/kds/ws", h.ServeWS)
}

func (h *Handler) parseBranchParam(c *gin.Context) (branchID uint64, orgID uint64, ok bool) {
	orgID, ok = middleware.OrgIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tidak terautentikasi"})
		return
	}
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "branch_id tidak valid"})
		ok = false
		return
	}
	return
}

func (h *Handler) branchBelongsToOrg(c *gin.Context, branchID, orgID uint64) bool {
	return h.branchBelongsToOrganization(c, branchID, orgID)
}

func (h *Handler) branchBelongsToOrganization(c *gin.Context, branchID, orgID uint64) bool {
	var foundOrgID uint64
	err := h.repo.db.QueryRowContext(c.Request.Context(),
		`SELECT organization_id FROM branches WHERE id = ? LIMIT 1`, branchID,
	).Scan(&foundOrgID)
	if err != nil || foundOrgID != orgID {
		c.JSON(http.StatusNotFound, gin.H{"error": "cabang tidak ditemukan"})
		return false
	}
	return true
}
