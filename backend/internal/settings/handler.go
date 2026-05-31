package settings

import (
	"github.com/gin-gonic/gin"

	"github.com/trackmycareer/app/pkg/response"
)

var validSettingKeys = map[string]bool{
	"registration_enabled": true,
}

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
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	for key := range req.Settings {
		if !validSettingKeys[key] {
			response.BadRequest(c, "unknown setting: "+key)
			return
		}
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
