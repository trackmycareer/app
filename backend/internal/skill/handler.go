package skill

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhcloudlabs/trackmy-career/internal/gamification"
	"github.com/bhcloudlabs/trackmy-career/pkg/response"
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
		Search:   c.Query("search"),
		Category: c.Query("category"),
		Limit:    limit,
		Offset:   offset,
	}

	skills, total, err := h.svc.List(c.Request.Context(), userID, params)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if skills == nil {
		skills = []Skill{}
	}

	response.OK(c, gin.H{
		"skills": skills,
		"total":  total,
	})
}

type createSkillRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    *string `json:"category"`
	Proficiency int     `json:"proficiency"`
	Notes       *string `json:"notes"`
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req createSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	if req.Proficiency == 0 {
		req.Proficiency = 1
	}

	sk := Skill{
		UserID:      userID,
		Name:        req.Name,
		Category:    req.Category,
		Proficiency: req.Proficiency,
		Notes:       req.Notes,
	}

	if err := h.svc.Create(c.Request.Context(), &sk); err != nil {
		if strings.Contains(err.Error(), "invalid proficiency") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to get full record
	created, err := h.svc.GetByID(c.Request.Context(), userID, sk.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "skill_created", &sk.ID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "skill_created", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, created, awards)
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	skillID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid skill ID")
		return
	}

	sk, err := h.svc.GetByID(c.Request.Context(), userID, skillID)
	if err != nil {
		response.NotFound(c, "skill not found")
		return
	}

	response.OK(c, sk)
}

type updateSkillRequest struct {
	Name        *string `json:"name"`
	Category    *string `json:"category"`
	Proficiency *int    `json:"proficiency"`
	Notes       *string `json:"notes"`
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	skillID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid skill ID")
		return
	}

	existing, err := h.svc.GetByID(c.Request.Context(), userID, skillID)
	if err != nil {
		response.NotFound(c, "skill not found")
		return
	}

	var req updateSkillRequest
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
	if req.Category != nil {
		existing.Category = req.Category
	}
	if req.Proficiency != nil {
		existing.Proficiency = *req.Proficiency
	}
	if req.Notes != nil {
		existing.Notes = req.Notes
	}

	if err := h.svc.Update(c.Request.Context(), &existing); err != nil {
		if strings.Contains(err.Error(), "invalid proficiency") {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, err)
		return
	}

	// Reload to get updated timestamps
	updated, err := h.svc.GetByID(c.Request.Context(), userID, skillID)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, updated)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	skillID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid skill ID")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), userID, skillID); err != nil {
		response.NotFound(c, "skill not found")
		return
	}

	response.NoContent(c)
}

type addEvidenceRequest struct {
	EvidenceType string `json:"evidence_type" binding:"required"`
	EvidenceID   string `json:"evidence_id" binding:"required"`
}

func (h *Handler) AddEvidence(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	skillID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid skill ID")
		return
	}

	var req addEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	evidenceID, err := uuid.Parse(req.EvidenceID)
	if err != nil {
		response.BadRequest(c, "invalid evidence_id")
		return
	}

	evidence, err := h.svc.AddEvidence(c.Request.Context(), userID, skillID, req.EvidenceType, evidenceID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid evidence_type") {
			response.BadRequest(c, err.Error())
			return
		}
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "skill not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	awards, awardsErr := h.gamification.RecordActivity(c.Request.Context(), userID, "evidence_linked", &skillID)
	if awardsErr != nil {
		slog.Error("gamification recording failed", "error", awardsErr.Error(), "action", "evidence_linked", "user_id", userID.String())
	}
	response.CreatedWithAwards(c, evidence, awards)
}

func (h *Handler) RemoveEvidence(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	skillID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid skill ID")
		return
	}

	evidenceID, err := uuid.Parse(c.Param("evidenceId"))
	if err != nil {
		response.BadRequest(c, "invalid evidence ID")
		return
	}

	if err := h.svc.RemoveEvidence(c.Request.Context(), userID, skillID, evidenceID); err != nil {
		if strings.Contains(err.Error(), "skill not found") {
			response.NotFound(c, "skill not found")
			return
		}
		if strings.Contains(err.Error(), "evidence not found") {
			response.NotFound(c, "evidence not found")
			return
		}
		response.InternalError(c, err)
		return
	}

	response.NoContent(c)
}
