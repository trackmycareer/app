package notification

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	params := ListParams{
		UnreadOnly: c.Query("unread") == "true",
		Limit:      limit,
		Offset:     offset,
	}

	items, total, err := h.svc.List(c.Request.Context(), userID, params)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if items == nil {
		items = []Notification{}
	}

	response.OK(c, gin.H{
		"notifications": items,
		"total":         total,
	})
}

func (h *Handler) UnreadCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	count, err := h.svc.UnreadCount(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"unread": count})
}

func (h *Handler) MarkRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid notification id")
		return
	}

	if err := h.svc.MarkRead(c.Request.Context(), userID, id); err != nil {
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *Handler) MarkAllRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	count, err := h.svc.MarkAllRead(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"marked_read": count})
}

func (h *Handler) GetPreferences(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	prefs, err := h.svc.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, prefs)
}

// updatePreferencesRequest uses pointers so a partial update leaves unspecified
// fields untouched.
type updatePreferencesRequest struct {
	Enabled      *bool `json:"enabled"`
	Remind90     *bool `json:"remind_90"`
	Remind30     *bool `json:"remind_30"`
	Remind7      *bool `json:"remind_7"`
	ChannelEmail *bool `json:"channel_email"`
	ChannelInApp *bool `json:"channel_in_app"`
}

func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	prefs, err := h.svc.GetPreferences(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	var req updatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if req.Enabled != nil {
		prefs.Enabled = *req.Enabled
	}
	if req.Remind90 != nil {
		prefs.Remind90 = *req.Remind90
	}
	if req.Remind30 != nil {
		prefs.Remind30 = *req.Remind30
	}
	if req.Remind7 != nil {
		prefs.Remind7 = *req.Remind7
	}
	if req.ChannelEmail != nil {
		prefs.ChannelEmail = *req.ChannelEmail
	}
	if req.ChannelInApp != nil {
		prefs.ChannelInApp = *req.ChannelInApp
	}

	if err := h.svc.UpdatePreferences(c.Request.Context(), userID, prefs); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, prefs)
}
