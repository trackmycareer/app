package export

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

// Handler provides HTTP handlers for export endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new export handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ExportJSON exports all user data as a JSON file download.
func (h *Handler) ExportJSON(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	data, err := h.svc.ExportJSON(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		response.InternalError(c, err)
		return
	}

	filename := fmt.Sprintf("career-export-%s.json", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/json", jsonBytes)
}

// ExportMarkdown exports all user data as a Markdown file download.
func (h *Handler) ExportMarkdown(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	md, err := h.svc.ExportMarkdown(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	filename := fmt.Sprintf("career-export-%s.md", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(md))
}
