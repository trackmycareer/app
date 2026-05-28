package certification

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhcloudlabs/trackmy-career/internal/gamification"
	"github.com/bhcloudlabs/trackmy-career/pkg/response"
	"github.com/bhcloudlabs/trackmy-career/pkg/types"
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

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	params := ListParams{
		Search: c.Query("search"),
		Status: c.Query("status"),
		Limit:  limit,
		Offset: offset,
	}

	certs, total, err := h.svc.List(c.Request.Context(), userID, params)
	if err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	if certs == nil {
		certs = []Certification{}
	}

	response.OK(c, gin.H{
		"certifications": certs,
		"total":          total,
	})
}

type createCertificationRequest struct {
	Name          string      `json:"name" binding:"required"`
	Provider      string      `json:"provider" binding:"required"`
	Status        string      `json:"status"`
	EarnedDate    *types.Date `json:"earned_date"`
	ExpiryDate    *types.Date `json:"expiry_date"`
	Cost          *float64    `json:"cost"`
	Currency      string      `json:"currency"`
	CredentialURL *string     `json:"credential_url"`
	StudyNotes    *string     `json:"study_notes"`
	StudyProgress int         `json:"study_progress"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	req.Provider = strings.TrimSpace(req.Provider)
	if req.Provider == "" {
		response.BadRequest(c, "provider is required")
		return
	}

	cert := Certification{
		UserID:        userID,
		Name:          req.Name,
		Provider:      req.Provider,
		Status:        req.Status,
		EarnedDate:    req.EarnedDate,
		ExpiryDate:    req.ExpiryDate,
		Cost:          req.Cost,
		Currency:      req.Currency,
		CredentialURL: req.CredentialURL,
		StudyNotes:    req.StudyNotes,
		StudyProgress: req.StudyProgress,
	}

	if err := h.svc.Create(c.Request.Context(), &cert); err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to get full record with timestamps
	created, err := h.svc.GetByID(c.Request.Context(), userID, cert.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "cert_created", &cert.ID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "cert_created", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, created, awards)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	certID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid certification ID")
		return
	}

	cert, err := h.svc.GetByID(c.Request.Context(), userID, certID)
	if err != nil {
		response.NotFound(c, "certification not found")
		return
	}

	response.OK(c, cert)
}

type updateCertificationRequest struct {
	Name          *string     `json:"name"`
	Provider      *string     `json:"provider"`
	Status        *string     `json:"status"`
	EarnedDate    *types.Date `json:"earned_date"`
	ExpiryDate    *types.Date `json:"expiry_date"`
	Cost          *float64    `json:"cost"`
	Currency      *string     `json:"currency"`
	CredentialURL *string     `json:"credential_url"`
	StudyNotes    *string     `json:"study_notes"`
	StudyProgress *int        `json:"study_progress"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	certID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid certification ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, certID)
	if err != nil {
		response.NotFound(c, "certification not found")
		return
	}

	var req updateCertificationRequest
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
	if req.Provider != nil {
		provider := strings.TrimSpace(*req.Provider)
		if provider == "" {
			response.BadRequest(c, "provider cannot be empty")
			return
		}
		existing.Provider = provider
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.EarnedDate != nil {
		existing.EarnedDate = req.EarnedDate
	}
	if req.ExpiryDate != nil {
		existing.ExpiryDate = req.ExpiryDate
	}
	if req.Cost != nil {
		existing.Cost = req.Cost
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.CredentialURL != nil {
		existing.CredentialURL = req.CredentialURL
	}
	if req.StudyNotes != nil {
		existing.StudyNotes = req.StudyNotes
	}
	if req.StudyProgress != nil {
		existing.StudyProgress = *req.StudyProgress
	}

	if err := h.svc.Update(c.Request.Context(), &existing); err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to get updated timestamps
	updated, err := h.svc.GetByID(c.Request.Context(), userID, certID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, updated)
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	certID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid certification ID")
		return
	}

	var req updateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if err := h.svc.UpdateStatus(c.Request.Context(), userID, certID, req.Status); err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			response.BadRequest(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "certification not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to return updated record
	updated, err := h.svc.GetByID(c.Request.Context(), userID, certID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.Status == "passed" {
		awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "cert_passed", &certID)
		if awardsErr != nil {
			slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "cert_passed", "user_id", userID.String())
		}
		response.OKWithAwards(c, updated, awards)
		return
	}

	response.OK(c, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	certID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid certification ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, certID); err != nil {
		response.NotFound(c, "certification not found")
		return
	}

	response.NoContent(c)
}
