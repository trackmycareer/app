package settings

import (
	"github.com/gin-gonic/gin"

	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetSettings(c *gin.Context) {
	all, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, all)
}

type UpdateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

func (h *Handler) UpdateSettings(c *gin.Context) {
	var req UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	for key, value := range req.Settings {
		if err := h.repo.Set(c.Request.Context(), key, value); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	all, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, all)
}
