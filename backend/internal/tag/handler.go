package tag

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	tags, err := h.repo.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if tags == nil {
		tags = []Tag{}
	}

	response.OK(c, tags)
}

type createTagRequest struct {
	Name   string `json:"name" binding:"required"`
	Colour string `json:"colour"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	colour := strings.TrimSpace(req.Colour)
	if colour == "" {
		colour = "#6366f1"
	}

	t := Tag{
		ID:     uuid.New(),
		UserID: userID,
		Name:   req.Name,
		Colour: colour,
	}

	if err := h.repo.Create(c.Request.Context(), &t); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			response.BadRequest(c, "a tag with that name already exists")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, t)
}

type updateTagRequest struct {
	Name   *string `json:"name"`
	Colour *string `json:"colour"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tag ID")
		return
	}

	existing, err := h.repo.GetByID(c.Request.Context(), userID, tagID)
	if err != nil {
		response.NotFound(c, "tag not found")
		return
	}

	var req updateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			response.BadRequest(c, "name cannot be empty")
			return
		}
		existing.Name = name
	}
	if req.Colour != nil {
		existing.Colour = strings.TrimSpace(*req.Colour)
	}

	if err := h.repo.Update(c.Request.Context(), &existing); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			response.BadRequest(c, "a tag with that name already exists")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, existing)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	tagID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid tag ID")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), userID, tagID); err != nil {
		response.NotFound(c, "tag not found")
		return
	}

	response.NoContent(c)
}
