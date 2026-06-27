package application

import (
	"log/slog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/pkg/response"
	"github.com/trackmycareer/app/pkg/types"
)

const (
	maxCompanyLen  = 255
	maxTitleLen    = 255
	maxLocationLen = 255
	maxJobURLLen   = 2000
	maxSourceLen   = 100
	maxSalaryLen   = 100
	maxNotesLen    = 5000
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

	applications, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if applications == nil {
		applications = []Application{}
	}

	response.OK(c, applications)
}

type createApplicationRequest struct {
	Company     string      `json:"company" binding:"required,max=255"`
	Title       string      `json:"title" binding:"required,max=255"`
	Status      string      `json:"status" binding:"max=20"`
	Location    *string     `json:"location"`
	WorkMode    *string     `json:"work_mode"`
	JobURL      *string     `json:"job_url"`
	Source      *string     `json:"source"`
	Salary      *string     `json:"salary"`
	AppliedDate *types.Date `json:"applied_date"`
	Notes       *string     `json:"notes"`
	SortOrder   int         `json:"sort_order"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
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

	if msg, ok := validateOptionalLengths(req.Location, req.JobURL, req.Source, req.Salary, req.Notes); !ok {
		response.BadRequest(c, msg)
		return
	}

	a := Application{
		UserID:      userID,
		Company:     req.Company,
		Title:       req.Title,
		Status:      req.Status,
		Location:    req.Location,
		WorkMode:    req.WorkMode,
		JobURL:      req.JobURL,
		Source:      req.Source,
		Salary:      req.Salary,
		AppliedDate: req.AppliedDate,
		Notes:       req.Notes,
		SortOrder:   req.SortOrder,
	}

	if err := h.svc.Create(c.Request.Context(), &a); err != nil {
		if strings.Contains(err.Error(), "invalid status") ||
			strings.Contains(err.Error(), "invalid work_mode") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	created, err := h.svc.GetByID(c.Request.Context(), userID, a.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "application_created", &a.ID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "application_created", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, created, awards)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application ID")
		return
	}

	a, err := h.svc.GetByID(c.Request.Context(), userID, applicationID)
	if err != nil {
		response.NotFound(c, "application not found")
		return
	}

	response.OK(c, a)
}

type updateApplicationRequest struct {
	Company     *string     `json:"company"`
	Title       *string     `json:"title"`
	Status      *string     `json:"status"`
	Location    *string     `json:"location"`
	WorkMode    *string     `json:"work_mode"`
	JobURL      *string     `json:"job_url"`
	Source      *string     `json:"source"`
	Salary      *string     `json:"salary"`
	AppliedDate *types.Date `json:"applied_date"`
	Notes       *string     `json:"notes"`
	SortOrder   *int        `json:"sort_order"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, applicationID)
	if err != nil {
		response.NotFound(c, "application not found")
		return
	}

	var req updateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	if req.Company != nil {
		company := strings.TrimSpace(*req.Company)
		if company == "" {
			response.BadRequest(c, "company cannot be empty")
			return
		}
		if len(company) > maxCompanyLen {
			response.BadRequest(c, "company must be at most 255 characters")
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
		if len(title) > maxTitleLen {
			response.BadRequest(c, "title must be at most 255 characters")
			return
		}
		existing.Title = title
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Location != nil {
		existing.Location = req.Location
	}
	if req.WorkMode != nil {
		existing.WorkMode = req.WorkMode
	}
	if req.JobURL != nil {
		existing.JobURL = req.JobURL
	}
	if req.Source != nil {
		existing.Source = req.Source
	}
	if req.Salary != nil {
		existing.Salary = req.Salary
	}
	if req.AppliedDate != nil {
		existing.AppliedDate = req.AppliedDate
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}
	if req.SortOrder != nil {
		existing.SortOrder = *req.SortOrder
	}

	if msg, ok := validateOptionalLengths(existing.Location, existing.JobURL, existing.Source, existing.Salary, existing.Notes); !ok {
		response.BadRequest(c, msg)
		return
	}

	if err := h.svc.Update(c.Request.Context(), &existing); err != nil {
		if strings.Contains(err.Error(), "invalid status") ||
			strings.Contains(err.Error(), "invalid work_mode") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	updated, err := h.svc.GetByID(c.Request.Context(), userID, applicationID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, updated)
}

type moveApplicationRequest struct {
	Status     string      `json:"status" binding:"required,max=20"`
	OrderedIDs []uuid.UUID `json:"ordered_ids"`
}

// Move changes an application's stage. With ordered_ids it rewrites the
// destination column's order (used for drag and reorder); without it the card is
// appended to the end of the target column (used by the per-card status select).
func (h *Handler) Move(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, applicationID)
	if err != nil {
		response.NotFound(c, "application not found")
		return
	}

	var req moveApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	priorStatus := existing.Status

	if len(req.OrderedIDs) > 0 {
		err = h.svc.Reorder(c.Request.Context(), userID, req.Status, req.OrderedIDs)
	} else {
		err = h.svc.Move(c.Request.Context(), userID, applicationID, req.Status)
	}
	if err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	updated, err := h.svc.GetByID(c.Request.Context(), userID, applicationID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	// Reward the milestone only on the transition into accepted.
	if req.Status == StatusAccepted && priorStatus != StatusAccepted {
		awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "application_accepted", &applicationID)
		if awardsErr != nil {
			slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "application_accepted", "user_id", userID.String())
		}
		response.OKWithAwards(c, updated, awards)
		return
	}

	response.OK(c, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	applicationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid application ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, applicationID); err != nil {
		response.NotFound(c, "application not found")
		return
	}

	response.NoContent(c)
}

// validateOptionalLengths bounds the free-text optional fields. Returns a
// human-readable message and false when a field is too long.
func validateOptionalLengths(location, jobURL, source, salary, notes *string) (string, bool) {
	if location != nil && len(*location) > maxLocationLen {
		return "location must be at most 255 characters", false
	}
	if jobURL != nil && len(*jobURL) > maxJobURLLen {
		return "job_url must be at most 2000 characters", false
	}
	if source != nil && len(*source) > maxSourceLen {
		return "source must be at most 100 characters", false
	}
	if salary != nil && len(*salary) > maxSalaryLen {
		return "salary must be at most 100 characters", false
	}
	if notes != nil && len(*notes) > maxNotesLen {
		return "notes must be at most 5000 characters", false
	}
	return "", true
}
