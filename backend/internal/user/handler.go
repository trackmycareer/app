package user

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, u)
}

type UpdateProfileRequest struct {
	Name      string  `json:"name" binding:"required"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	u.Name = req.Name
	u.AvatarURL = req.AvatarURL

	if err := h.repo.Update(c.Request.Context(), &u); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, u)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if u.PasswordHash == "" {
		response.BadRequest(c, "account uses OAuth authentication; password change is not available")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		response.Unauthorised(c, "current password is incorrect")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if err := h.repo.UpdatePassword(c.Request.Context(), userID, string(hash)); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "password updated"})
}
