package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/pkg/response"
)

// Handler provides HTTP handlers for admin endpoints.
type Handler struct {
	svc      *Service
	userRepo *user.Repository
}

// NewHandler creates a new admin handler.
func NewHandler(svc *Service, userRepo *user.Repository) *Handler {
	return &Handler{svc: svc, userRepo: userRepo}
}

// ListUsers returns a paginated list of all users.
func (h *Handler) ListUsers(c *gin.Context) {
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	users, total, err := h.userRepo.List(c.Request.Context(), search, limit, offset)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if users == nil {
		users = []user.User{}
	}

	response.OK(c, gin.H{
		"users": users,
		"total": total,
	})
}

// GetUser returns a single user by ID.
func (h *Handler) GetUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	detail, err := h.svc.GetUserDetail(c.Request.Context(), u)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, detail)
}

// DeleteUser removes a user by ID.
func (h *Handler) DeleteUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	// Prevent self-deletion
	callerID := c.MustGet("user_id").(uuid.UUID)
	if callerID == userID {
		response.BadRequest(c, "you cannot delete your own account")
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}

type toggleAdminRequest struct {
	IsAdmin bool `json:"is_admin"`
}

// ToggleAdmin sets the admin flag on a user.
func (h *Handler) ToggleAdmin(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid user ID")
		return
	}

	// Prevent removing your own admin rights
	callerID := c.MustGet("user_id").(uuid.UUID)
	if callerID == userID {
		response.BadRequest(c, "you cannot change your own admin status")
		return
	}

	var req toggleAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if err := h.userRepo.SetAdmin(c.Request.Context(), userID, req.IsAdmin); err != nil {
		response.InternalError(c, err)
		return
	}

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, u)
}

// GetStats returns aggregate instance statistics.
func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, stats)
}
