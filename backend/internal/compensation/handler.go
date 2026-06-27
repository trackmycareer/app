package compensation

import (
	"log/slog"
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

	entries, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if entries == nil {
		entries = []Compensation{}
	}

	response.OK(c, entries)
}

type createCompensationRequest struct {
	JobID         uuid.UUID  `json:"job_id"`
	EffectiveDate types.Date `json:"effective_date" binding:"required"`
	Currency      string     `json:"currency" binding:"max=3"`
	PayBasis      string     `json:"pay_basis" binding:"max=20"`
	Amounts       Amounts    `json:"amounts"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createCompensationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}
	if req.JobID == uuid.Nil {
		response.BadRequest(c, "job_id is required")
		return
	}

	comp := Compensation{
		UserID:        userID,
		JobID:         req.JobID,
		EffectiveDate: req.EffectiveDate,
		Currency:      req.Currency,
		PayBasis:      req.PayBasis,
		Amounts:       req.Amounts,
	}

	if err := h.svc.Create(c.Request.Context(), &comp); err != nil {
		writeServiceError(c, err)
		return
	}

	created, err := h.svc.GetByID(c.Request.Context(), userID, comp.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "compensation_logged", &comp.ID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "compensation_logged", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, created, awards)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid compensation ID")
		return
	}

	comp, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "compensation not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.OK(c, comp)
}

type updateCompensationRequest struct {
	JobID         *uuid.UUID  `json:"job_id"`
	EffectiveDate *types.Date `json:"effective_date"`
	Currency      *string     `json:"currency"`
	PayBasis      *string     `json:"pay_basis"`
	Amounts       *Amounts    `json:"amounts"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid compensation ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "compensation not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	var req updateCompensationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if req.JobID != nil {
		if *req.JobID == uuid.Nil {
			response.BadRequest(c, "job_id cannot be empty")
			return
		}
		existing.JobID = *req.JobID
	}
	if req.EffectiveDate != nil {
		existing.EffectiveDate = *req.EffectiveDate
	}
	// Treat an empty currency/pay_basis as "leave unchanged" rather than letting
	// it silently reset to the GBP/annual defaults during a partial update.
	if req.Currency != nil && strings.TrimSpace(*req.Currency) != "" {
		existing.Currency = *req.Currency
	}
	if req.PayBasis != nil && strings.TrimSpace(*req.PayBasis) != "" {
		existing.PayBasis = *req.PayBasis
	}
	if req.Amounts != nil {
		existing.Amounts = *req.Amounts
	}

	if err := h.svc.Update(c.Request.Context(), &existing); err != nil {
		writeServiceError(c, err)
		return
	}

	updated, err := h.svc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "compensation_updated", &id)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "compensation_updated", "user_id", userID.String())
	}
	response.OKWithAwards(c, updated, awards)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid compensation ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, id); err != nil {
		response.NotFound(c, "compensation not found")
		return
	}

	response.NoContent(c)
}

// writeServiceError maps a service-layer error to an HTTP response. Validation
// and ownership failures are client errors; anything else (encryption, database)
// is logged and returned as a 500.
func writeServiceError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		response.NotFound(c, msg)
	case strings.Contains(msg, "invalid"),
		strings.Contains(msg, "must be"),
		strings.Contains(msg, "required"):
		response.BadRequest(c, msg)
	default:
		response.InternalError(c, err)
	}
}
