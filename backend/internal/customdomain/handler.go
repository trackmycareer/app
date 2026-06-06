package customdomain

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/pkg/response"
)

// Handler exposes HTTP endpoints for custom domain management.
type Handler struct {
	svc      *Service
	userRepo *user.Repository
}

// NewHandler creates a new custom domain handler.
func NewHandler(svc *Service, userRepo *user.Repository) *Handler {
	return &Handler{svc: svc, userRepo: userRepo}
}

type createRequest struct {
	Domain string `json:"domain" binding:"required"`
}

type updateThemeRequest struct {
	AccentColour string `json:"accent_colour" binding:"required"`
}

// Create registers a new custom domain for the authenticated user.
func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if !u.IsOneTimeSupporter && !u.IsSubscriber {
		response.Forbidden(c, "custom domains are available to supporters")
		return
	}

	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	d, err := h.svc.Create(c.Request.Context(), userID, req.Domain)
	if err != nil {
		if isUserError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.Created(c, d)
}

// Get returns the custom domain for the authenticated user.
func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	d, err := h.svc.Get(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if d == nil {
		response.NotFound(c, "no custom domain configured")
		return
	}

	response.OK(c, d)
}

// Delete removes the custom domain for the authenticated user.
func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if !u.IsOneTimeSupporter && !u.IsSubscriber {
		response.Forbidden(c, "custom domains are available to supporters")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}

// Verify checks the current verification and SSL status of the custom domain.
func (h *Handler) Verify(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if !u.IsOneTimeSupporter && !u.IsSubscriber {
		response.Forbidden(c, "custom domains are available to supporters")
		return
	}

	d, err := h.svc.Verify(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "no custom domain") {
			response.NotFound(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, d)
}

// UpdateTheme updates the accent colour for the custom domain.
func (h *Handler) UpdateTheme(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if !u.IsOneTimeSupporter && !u.IsSubscriber {
		response.Forbidden(c, "custom domains are available to supporters")
		return
	}

	var req updateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	d, err := h.svc.UpdateAccentColour(c.Request.Context(), userID, req.AccentColour)
	if err != nil {
		if isUserError(err) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, d)
}

// isUserError returns true for errors that should be returned as 400 Bad Request
// rather than 500 Internal Server Error.
func isUserError(err error) bool {
	msg := err.Error()
	userMessages := []string{
		"domain is required",
		"domain must be",
		"domain contains",
		"domains ending with",
		"you already have",
		"this domain is already",
		"custom domains are not available",
		"no custom domain",
		"accent colour must be",
	}
	for _, prefix := range userMessages {
		if strings.Contains(msg, prefix) {
			return true
		}
	}
	return false
}
