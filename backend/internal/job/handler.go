package job

import (
	"log/slog"
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

	jobs, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if jobs == nil {
		jobs = []Job{}
	}

	response.OK(c, jobs)
}

type createJobRequest struct {
	Company          string      `json:"company" binding:"required"`
	Title            string      `json:"title" binding:"required"`
	StartDate        types.Date  `json:"start_date" binding:"required"`
	EndDate          *types.Date `json:"end_date"`
	EmploymentType   string      `json:"employment_type"`
	TransitionType   *string     `json:"transition_type"`
	Location         *string     `json:"location"`
	Remote           bool        `json:"remote"`
	Responsibilities *string     `json:"responsibilities"`
	Notes            *string     `json:"notes"`
	SortOrder        int         `json:"sort_order"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	req.Company = strings.TrimSpace(req.Company)
	if req.Company == "" {
		response.BadRequest(c, "company is required")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		response.BadRequest(c, "title is required")
		return
	}

	j := Job{
		UserID:           userID,
		Company:          req.Company,
		Title:            req.Title,
		StartDate:        req.StartDate,
		EndDate:          req.EndDate,
		EmploymentType:   req.EmploymentType,
		TransitionType:   req.TransitionType,
		Location:         req.Location,
		Remote:           req.Remote,
		Responsibilities: req.Responsibilities,
		Notes:            req.Notes,
		SortOrder:        req.SortOrder,
	}

	if err := h.svc.Create(c.Request.Context(), &j); err != nil {
		if strings.Contains(err.Error(), "invalid employment_type") ||
			strings.Contains(err.Error(), "invalid transition_type") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to get full record with timestamps
	created, err := h.svc.GetByID(c.Request.Context(), userID, j.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "job_created", &j.ID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "job_created", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, created, awards)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job ID")
		return
	}

	j, err := h.svc.GetByID(c.Request.Context(), userID, jobID)
	if err != nil {
		response.NotFound(c, "job not found")
		return
	}

	response.OK(c, j)
}

type updateJobRequest struct {
	Company          *string     `json:"company"`
	Title            *string     `json:"title"`
	StartDate        *types.Date `json:"start_date"`
	EndDate          *types.Date `json:"end_date"`
	EmploymentType   *string     `json:"employment_type"`
	TransitionType   *string     `json:"transition_type"`
	Location         *string     `json:"location"`
	Remote           *bool       `json:"remote"`
	Responsibilities *string     `json:"responsibilities"`
	Notes            *string     `json:"notes"`
	SortOrder        *int        `json:"sort_order"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, jobID)
	if err != nil {
		response.NotFound(c, "job not found")
		return
	}

	var req updateJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if req.Company != nil {
		company := strings.TrimSpace(*req.Company)
		if company == "" {
			response.BadRequest(c, "company cannot be empty")
			return
		}
		existing.Company = company
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			response.BadRequest(c, "title cannot be empty")
			return
		}
		existing.Title = title
	}
	if req.StartDate != nil {
		existing.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		existing.EndDate = req.EndDate
	}
	if req.EmploymentType != nil {
		existing.EmploymentType = *req.EmploymentType
	}
	if req.TransitionType != nil {
		existing.TransitionType = req.TransitionType
	}
	if req.Location != nil {
		existing.Location = req.Location
	}
	if req.Remote != nil {
		existing.Remote = *req.Remote
	}
	if req.Responsibilities != nil {
		existing.Responsibilities = req.Responsibilities
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}

	if err := h.svc.Update(c.Request.Context(), &existing); err != nil {
		if strings.Contains(err.Error(), "invalid employment_type") ||
			strings.Contains(err.Error(), "invalid transition_type") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to get updated timestamps
	updated, err := h.svc.GetByID(c.Request.Context(), userID, jobID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "job_updated", &jobID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "job_updated", "user_id", userID.String())
	}
	response.OKWithAwards(c, updated, awards)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid job ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, jobID); err != nil {
		response.NotFound(c, "job not found")
		return
	}

	response.NoContent(c)
}
