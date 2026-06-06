package user

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/password"
	"github.com/trackmycareer/app/internal/storage"
	"github.com/trackmycareer/app/pkg/response"
)

// DomainCleaner is an optional interface for cleaning up custom domain
// resources when a user account is deleted.
type DomainCleaner interface {
	Delete(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	repo          *Repository
	storage       *storage.Client
	domainCleaner DomainCleaner
}

func NewHandler(repo *Repository, storage *storage.Client, domainCleaner DomainCleaner) *Handler {
	return &Handler{repo: repo, storage: storage, domainCleaner: domainCleaner}
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
	Name string `json:"name" binding:"required,max=255"`
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	u.Name = req.Name

	if err := h.repo.Update(c.Request.Context(), &u); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, u)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=128"`
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
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

	match, err := password.Verify(req.CurrentPassword, u.PasswordHash)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if !match {
		response.Unauthorised(c, "current password is incorrect")
		return
	}

	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if err := h.repo.UpdatePassword(c.Request.Context(), userID, hash); err != nil {
		response.InternalError(c, err)
		return
	}

	// Invalidate all existing tokens by incrementing the token version
	if err := h.repo.IncrementTokenVersion(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "password updated"})
}

// maxAvatarSize is the maximum allowed avatar file size (2 MB).
const maxAvatarSize = 2 << 20

type avatarFormat struct {
	outExt         string
	format         imaging.Format
	outContentType string
}

// allowedAvatarTypes maps input MIME types to output format details.
// WebP inputs are re-encoded as PNG because imaging cannot encode WebP.
var allowedAvatarTypes = map[string]avatarFormat{
	"image/jpeg": {".jpg", imaging.JPEG, "image/jpeg"},
	"image/png":  {".png", imaging.PNG, "image/png"},
	"image/webp": {".png", imaging.PNG, "image/png"},
}

// avatarExtensions lists all possible avatar file extensions used for cleanup.
var avatarExtensions = []string{".jpg", ".png", ".webp"}

// UploadAvatar handles avatar image upload, resizes to 256x256, and stores in S3.
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if err := c.Request.ParseMultipartForm(maxAvatarSize); err != nil {
		response.BadRequest(c, "request body too large or not multipart")
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.BadRequest(c, "avatar file is required")
		return
	}
	defer file.Close()

	if header.Size > maxAvatarSize {
		response.BadRequest(c, "avatar file must be 2 MB or smaller")
		return
	}

	contentType := header.Header.Get("Content-Type")
	af, ok := allowedAvatarTypes[contentType]
	if !ok {
		response.BadRequest(c, "avatar must be a JPEG, PNG, or WebP image")
		return
	}

	// Verify the file content matches the declared type via magic bytes.
	sniff := make([]byte, 512)
	n, err := io.ReadFull(file, sniff)
	if err != nil && err != io.ErrUnexpectedEOF {
		response.BadRequest(c, "unable to read avatar file")
		return
	}
	detected := http.DetectContentType(sniff[:n])
	if _, ok := allowedAvatarTypes[detected]; !ok {
		response.BadRequest(c, "file content does not match an allowed image type")
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.InternalError(c, fmt.Errorf("seeking avatar file: %w", err))
		return
	}

	img, err := imaging.Decode(file)
	if err != nil {
		response.BadRequest(c, "unable to decode image")
		return
	}

	resized := imaging.Fill(img, 256, 256, imaging.Center, imaging.Lanczos)

	// Encode the resized image into a buffer.
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, resized, af.format); err != nil {
		response.InternalError(c, fmt.Errorf("encoding avatar: %w", err))
		return
	}

	// Remove any existing avatar files for this user (all possible extensions).
	for _, ext := range avatarExtensions {
		key := "images/" + userID.String() + ext
		if err := h.storage.Delete(c.Request.Context(), key); err != nil {
			slog.Warn("failed to delete old avatar", "key", key, "error", err)
		}
	}

	key := "images/" + userID.String() + af.outExt
	avatarURL, err := h.storage.Upload(c.Request.Context(), key, &buf, int64(buf.Len()), af.outContentType)
	if err != nil {
		response.InternalError(c, fmt.Errorf("uploading avatar: %w", err))
		return
	}

	if err := h.repo.UpdateAvatarURL(c.Request.Context(), userID, &avatarURL); err != nil {
		response.InternalError(c, fmt.Errorf("updating avatar URL: %w", err))
		return
	}

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, fmt.Errorf("fetching updated user: %w", err))
		return
	}

	response.OK(c, u)
}

type DeleteAccountRequest struct {
	Password     string `json:"password"`
	Confirmation string `json:"confirmation"`
}

func (h *Handler) DeleteAccount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	var req DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if u.Provider == "email" {
		if req.Password == "" {
			response.BadRequest(c, "password is required to delete your account")
			return
		}
		match, err := password.Verify(req.Password, u.PasswordHash)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		if !match {
			response.Unauthorised(c, "incorrect password")
			return
		}
	} else {
		if req.Confirmation != "DELETE" {
			response.BadRequest(c, "please type DELETE to confirm account deletion")
			return
		}
	}

	// Best-effort custom domain cleanup (Cloudflare hostname removal).
	if h.domainCleaner != nil {
		if cdErr := h.domainCleaner.Delete(c.Request.Context(), userID); cdErr != nil {
			slog.Warn("failed to clean up custom domain during account deletion",
				"error", cdErr.Error(), "user_id", userID.String())
		}
	}

	// Remove avatar files from S3 (all possible extensions).
	for _, ext := range avatarExtensions {
		key := "images/" + userID.String() + ext
		if err := h.storage.Delete(c.Request.Context(), key); err != nil {
			slog.Warn("failed to delete old avatar", "key", key, "error", err)
		}
	}

	if err := h.repo.Delete(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err)
		return
	}

	// Clear the refresh token cookie.
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", true, true)

	response.NoContent(c)
}

type UpdateNewsletterRequest struct {
	OptIn bool `json:"opt_in"`
}

func (h *Handler) UpdateNewsletter(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req UpdateNewsletterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if err := h.repo.UpdateNewsletterOptIn(c.Request.Context(), userID, req.OptIn); err != nil {
		response.InternalError(c, err)
		return
	}

	u, err := h.repo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, u)
}

// DeleteAvatar removes the avatar from S3 and clears the avatar_url in the database.
func (h *Handler) DeleteAvatar(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if _, err := h.repo.GetByID(c.Request.Context(), userID); err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Remove avatar files from S3 (all possible extensions).
	for _, ext := range avatarExtensions {
		key := "images/" + userID.String() + ext
		if err := h.storage.Delete(c.Request.Context(), key); err != nil {
			slog.Warn("failed to delete old avatar", "key", key, "error", err)
		}
	}

	if err := h.repo.UpdateAvatarURL(c.Request.Context(), userID, nil); err != nil {
		response.InternalError(c, fmt.Errorf("clearing avatar URL: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "avatar removed"}})
}
