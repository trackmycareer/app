package win

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/pkg/response"
	"github.com/trackmycareer/app/pkg/types"
)

type Handler struct {
	svc          *Service
	gamification *gamification.Service
}

func NewHandler(svc *Service, gamificationSvc *gamification.Service) *Handler {
	return &Handler{svc: svc, gamification: gamificationSvc}
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	params := ListParams{
		Search:   c.Query("search"),
		Category: c.Query("category"),
		TagID:    c.Query("tag_id"),
		From:     c.Query("from"),
		To:       c.Query("to"),
		Limit:    limit,
		Offset:   offset,
	}

	wins, total, err := h.svc.List(c.Request.Context(), userID, params)
	if err != nil {
		if strings.Contains(err.Error(), "invalid category") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	if wins == nil {
		wins = []Win{}
	}

	response.OK(c, gin.H{
		"wins":  wins,
		"total": total,
	})
}

type createWinRequest struct {
	Title       string     `json:"title" binding:"required,max=255"`
	Description *string    `json:"description"`
	OccurredOn  types.Date `json:"occurred_on"`
	Category    string     `json:"category" binding:"max=100"`
	TagIDs      []string   `json:"tag_ids"`
}

const (
	maxWinTitleLen       = 255
	maxWinDescriptionLen = 10000
)

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createWinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		response.BadRequest(c, "title is required")
		return
	}

	if req.Description != nil && len(*req.Description) > maxWinDescriptionLen {
		response.BadRequest(c, "description must be at most 10000 characters")
		return
	}

	tagIDs, err := parseUUIDs(req.TagIDs)
	if err != nil {
		response.BadRequest(c, "invalid tag_ids: "+err.Error())
		return
	}

	w := Win{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		OccurredOn:  req.OccurredOn,
		Category:    req.Category,
	}

	if err := h.svc.Create(c.Request.Context(), &w, tagIDs); err != nil {
		if strings.Contains(err.Error(), "invalid category") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to include tags
	created, err := h.svc.GetByID(c.Request.Context(), userID, w.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "win_created", &w.ID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "win_created", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, created, awards)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	winID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid win ID")
		return
	}

	w, err := h.svc.GetByID(c.Request.Context(), userID, winID)
	if err != nil {
		response.NotFound(c, "win not found")
		return
	}

	response.OK(c, w)
}

type updateWinRequest struct {
	Title       *string     `json:"title"`
	Description *string     `json:"description"`
	OccurredOn  *types.Date `json:"occurred_on"`
	Category    *string     `json:"category"`
	TagIDs      []string    `json:"tag_ids"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	winID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid win ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, winID)
	if err != nil {
		response.NotFound(c, "win not found")
		return
	}

	var req updateWinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			response.BadRequest(c, "title cannot be empty")
			return
		}
		if len(title) > maxWinTitleLen {
			response.BadRequest(c, "title must be at most 255 characters")
			return
		}
		existing.Title = title
	}
	if req.Description != nil {
		if len(*req.Description) > maxWinDescriptionLen {
			response.BadRequest(c, "description must be at most 10000 characters")
			return
		}
		existing.Description = req.Description
	}
	if req.OccurredOn != nil {
		existing.OccurredOn = *req.OccurredOn
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}

	// Collect tag IDs: use provided tag_ids, or keep existing
	var tagIDs []uuid.UUID
	if req.TagIDs != nil {
		tagIDs, err = parseUUIDs(req.TagIDs)
		if err != nil {
			response.BadRequest(c, "invalid tag_ids: "+err.Error())
			return
		}
	} else {
		for _, t := range existing.Tags {
			tagIDs = append(tagIDs, t.ID)
		}
	}

	if err := h.svc.Update(c.Request.Context(), &existing, tagIDs); err != nil {
		if strings.Contains(err.Error(), "invalid category") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to include updated tags
	updated, err := h.svc.GetByID(c.Request.Context(), userID, winID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	winID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid win ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, winID); err != nil {
		response.NotFound(c, "win not found")
		return
	}

	response.NoContent(c)
}

func parseUUIDs(strs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
