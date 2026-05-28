package gamification

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

// Handler exposes gamification HTTP endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new gamification handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GetProgress returns the aggregated gamification overview for the authenticated user.
func (h *Handler) GetProgress(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	progress, err := h.svc.GetProgress(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, progress)
}

// ListBadges returns all badge definitions with per-user earned status.
func (h *Handler) ListBadges(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	badges, err := h.svc.ListBadgesForUser(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if badges == nil {
		badges = []Badge{}
	}
	response.OK(c, badges)
}

// GetHeatmap returns daily activity counts for the past N days (default 365).
func (h *Handler) GetHeatmap(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	days := 365
	if d, err := strconv.Atoi(c.Query("days")); err == nil && d > 0 {
		days = d
	}

	entries, err := h.svc.GetHeatmap(c.Request.Context(), userID, days)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if entries == nil {
		entries = []HeatmapEntry{}
	}
	response.OK(c, entries)
}

// GetStreak returns the current streak details for the authenticated user.
func (h *Handler) GetStreak(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	streak, err := h.svc.GetStreak(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, streak)
}

// ---------------------------------------------------------------------------
// Admin badge management
// ---------------------------------------------------------------------------

var validConditionTypes = map[string]bool{
	"count":  true,
	"streak": true,
	"action": true,
}

var validTiers = map[string]bool{
	"bronze": true,
	"silver": true,
	"gold":   true,
}

type adminBadgeRequest struct {
	Name            string          `json:"name" binding:"required"`
	Description     *string         `json:"description"`
	Icon            string          `json:"icon" binding:"required"`
	Colour          string          `json:"colour" binding:"required"`
	Tier            string          `json:"tier" binding:"required"`
	ConditionType   string          `json:"condition_type" binding:"required"`
	ConditionConfig json.RawMessage `json:"condition_config" binding:"required"`
}

// AdminListBadges returns all badge definitions (admin view).
func (h *Handler) AdminListBadges(c *gin.Context) {
	badges, err := h.svc.repo.ListBadges(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if badges == nil {
		badges = []Badge{}
	}
	response.OK(c, badges)
}

// AdminCreateBadge creates a new badge definition.
func (h *Handler) AdminCreateBadge(c *gin.Context) {
	var req adminBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if !validConditionTypes[req.ConditionType] {
		response.BadRequest(c, "invalid condition_type: must be one of count, streak, action")
		return
	}
	if !validTiers[strings.ToLower(req.Tier)] {
		response.BadRequest(c, "invalid tier: must be one of bronze, silver, gold")
		return
	}

	badge := Badge{
		Name:            req.Name,
		Description:     req.Description,
		Icon:            req.Icon,
		Colour:          req.Colour,
		Tier:            strings.ToLower(req.Tier),
		ConditionType:   req.ConditionType,
		ConditionConfig: req.ConditionConfig,
		IsDefault:       false,
	}

	if err := h.svc.repo.CreateBadge(c.Request.Context(), &badge); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Created(c, badge)
}

// AdminUpdateBadge updates an existing badge definition.
func (h *Handler) AdminUpdateBadge(c *gin.Context) {
	badgeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid badge ID")
		return
	}

	existing, err := h.svc.repo.GetBadgeByID(c.Request.Context(), badgeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "badge not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	var req adminBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if !validConditionTypes[req.ConditionType] {
		response.BadRequest(c, "invalid condition_type: must be one of count, streak, action")
		return
	}
	if !validTiers[strings.ToLower(req.Tier)] {
		response.BadRequest(c, "invalid tier: must be one of bronze, silver, gold")
		return
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Icon = req.Icon
	existing.Colour = req.Colour
	existing.Tier = strings.ToLower(req.Tier)
	existing.ConditionType = req.ConditionType
	existing.ConditionConfig = req.ConditionConfig

	if err := h.svc.repo.UpdateBadge(c.Request.Context(), &existing); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "badge not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, existing)
}

// AdminDeleteBadge deletes a badge definition (only non-default badges).
func (h *Handler) AdminDeleteBadge(c *gin.Context) {
	badgeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid badge ID")
		return
	}

	existing, err := h.svc.repo.GetBadgeByID(c.Request.Context(), badgeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "badge not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	if existing.IsDefault {
		response.BadRequest(c, "cannot delete a default badge")
		return
	}

	if err := h.svc.repo.DeleteBadge(c.Request.Context(), badgeID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "badge not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}
