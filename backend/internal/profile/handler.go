package profile

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/bhcloudlabs/trackmy-career/internal/certification"
	"github.com/bhcloudlabs/trackmy-career/internal/gamification"
	"github.com/bhcloudlabs/trackmy-career/internal/job"
	"github.com/bhcloudlabs/trackmy-career/internal/skill"
	"github.com/bhcloudlabs/trackmy-career/internal/user"
	"github.com/bhcloudlabs/trackmy-career/internal/win"
	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{1,48}[a-zA-Z0-9]$`)

// Handler manages public profiles and profile settings.
type Handler struct {
	userRepo         *user.Repository
	gamificationRepo *gamification.Repository
	winRepo          *win.Repository
	jobRepo          *job.Repository
	certRepo         *certification.Repository
	skillRepo        *skill.Repository
}

// NewHandler creates a new profile handler.
func NewHandler(
	userRepo *user.Repository,
	gamificationRepo *gamification.Repository,
	winRepo *win.Repository,
	jobRepo *job.Repository,
	certRepo *certification.Repository,
	skillRepo *skill.Repository,
) *Handler {
	return &Handler{
		userRepo:         userRepo,
		gamificationRepo: gamificationRepo,
		winRepo:          winRepo,
		jobRepo:          jobRepo,
		certRepo:         certRepo,
		skillRepo:        skillRepo,
	}
}

// visibilityConfig holds the parsed profile_visibility JSON.
type visibilityConfig struct {
	Jobs           bool `json:"jobs"`
	Certifications bool `json:"certifications"`
	Skills         bool `json:"skills"`
	Wins           bool `json:"wins"`
}

func parseVisibility(raw json.RawMessage) visibilityConfig {
	var cfg visibilityConfig
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &cfg)
	}
	return cfg
}

// profileSettingsResponse is the response for GET /user/me/profile.
type profileSettingsResponse struct {
	Username          *string         `json:"username"`
	Bio               *string         `json:"bio"`
	ProfileVisibility json.RawMessage `json:"profile_visibility"`
}

// updateProfileRequest is the request body for PUT /user/me/profile.
type updateProfileRequest struct {
	Username          *string         `json:"username"`
	Bio               *string         `json:"bio"`
	ProfileVisibility json.RawMessage `json:"profile_visibility"`
}

// publicProfileResponse is the response for GET /profiles/:username.
type publicProfileResponse struct {
	Name           string             `json:"name"`
	Bio            *string            `json:"bio,omitempty"`
	Level          int                `json:"level"`
	LevelTitle     string             `json:"level_title"`
	Badges         []gamification.Badge `json:"badges"`
	Jobs           []job.Job          `json:"jobs,omitempty"`
	Certifications []certification.Certification `json:"certifications,omitempty"`
	Skills         []skill.Skill      `json:"skills,omitempty"`
	Wins           []win.Win          `json:"wins,omitempty"`
}

// GetProfileSettings returns the profile settings for the authenticated user.
func (h *Handler) GetProfileSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	resp := profileSettingsResponse{
		Username:          u.Username,
		Bio:               u.Bio,
		ProfileVisibility: u.ProfileVisibility,
	}

	response.OK(c, resp)
}

// UpdateProfileSettings updates the profile settings for the authenticated user.
func (h *Handler) UpdateProfileSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// Fetch current user to merge fields.
	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Determine final values, keeping existing if not provided.
	finalUsername := u.Username
	finalBio := u.Bio
	finalVisibility := u.ProfileVisibility

	if req.Username != nil {
		trimmed := strings.TrimSpace(*req.Username)
		if len(trimmed) < 3 || len(trimmed) > 50 {
			response.BadRequest(c, "username must be between 3 and 50 characters")
			return
		}
		if !usernameRegex.MatchString(trimmed) {
			response.BadRequest(c, "username may only contain alphanumeric characters and hyphens")
			return
		}

		// Check uniqueness (only if changing).
		if u.Username == nil || *u.Username != trimmed {
			existing, lookupErr := h.userRepo.GetByUsername(c.Request.Context(), trimmed)
			if lookupErr == nil && existing.ID != userID {
				response.BadRequest(c, "username is already taken")
				return
			}
		}
		finalUsername = &trimmed
	}

	if req.Bio != nil {
		finalBio = req.Bio
	}

	if len(req.ProfileVisibility) > 0 {
		finalVisibility = req.ProfileVisibility
	}

	if err := h.userRepo.UpdateProfile(c.Request.Context(), userID, finalUsername, finalBio, finalVisibility); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, profileSettingsResponse{
		Username:          finalUsername,
		Bio:               finalBio,
		ProfileVisibility: finalVisibility,
	})
}

// GetPublicProfile returns the public profile for a given username.
func (h *Handler) GetPublicProfile(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		response.BadRequest(c, "username is required")
		return
	}

	ctx := c.Request.Context()

	u, err := h.userRepo.GetByUsername(ctx, username)
	if err != nil {
		response.NotFound(c, "profile not found")
		return
	}

	// Determine visibility.
	vis := parseVisibility(u.ProfileVisibility)

	// Fetch gamification data (always visible).
	points, err := h.gamificationRepo.GetOrCreatePoints(ctx, u.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	level, levelTitle := gamification.LevelForPoints(points.TotalPoints)

	// Fetch earned badges (always visible).
	allBadges, err := h.gamificationRepo.ListBadgesWithEarnedStatus(ctx, u.ID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	var earnedBadges []gamification.Badge
	for _, b := range allBadges {
		if b.Earned {
			earnedBadges = append(earnedBadges, b)
		}
	}
	if earnedBadges == nil {
		earnedBadges = []gamification.Badge{}
	}

	resp := publicProfileResponse{
		Name:       u.Name,
		Bio:        u.Bio,
		Level:      level,
		LevelTitle: levelTitle,
		Badges:     earnedBadges,
	}

	// Conditionally include sections based on visibility settings.
	if vis.Jobs {
		jobs, jobErr := h.jobRepo.List(ctx, u.ID)
		if jobErr != nil {
			response.InternalError(c, jobErr)
			return
		}
		if jobs == nil {
			jobs = []job.Job{}
		}
		resp.Jobs = jobs
	}

	if vis.Certifications {
		certs, _, certErr := h.certRepo.List(ctx, u.ID, certification.ListParams{Limit: 100})
		if certErr != nil {
			response.InternalError(c, certErr)
			return
		}
		if certs == nil {
			certs = []certification.Certification{}
		}
		resp.Certifications = certs
	}

	if vis.Skills {
		skills, _, skillErr := h.skillRepo.List(ctx, u.ID, skill.ListParams{Limit: 100})
		if skillErr != nil {
			response.InternalError(c, skillErr)
			return
		}
		if skills == nil {
			skills = []skill.Skill{}
		}
		resp.Skills = skills
	}

	if vis.Wins {
		wins, _, winErr := h.winRepo.List(ctx, u.ID, win.ListParams{Limit: 100})
		if winErr != nil {
			response.InternalError(c, winErr)
			return
		}
		if wins == nil {
			wins = []win.Win{}
		}
		resp.Wins = wins
	}

	response.OK(c, resp)
}
