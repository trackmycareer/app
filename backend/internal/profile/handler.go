package profile

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/certification"
	"github.com/trackmycareer/app/internal/customdomain"
	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/internal/job"
	"github.com/trackmycareer/app/internal/linkedaccount"
	"github.com/trackmycareer/app/internal/skill"
	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/internal/win"
	"github.com/trackmycareer/app/pkg/response"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]{1,48}[a-zA-Z0-9]$`)

// Handler manages public profiles and profile settings.
type Handler struct {
	userRepo             *user.Repository
	gamificationRepo     *gamification.Repository
	winRepo              *win.Repository
	jobRepo              *job.Repository
	certRepo             *certification.Repository
	skillRepo            *skill.Repository
	linkedAccountService *linkedaccount.Service
	customDomainRepo     *customdomain.Repository
}

// NewHandler creates a new profile handler.
func NewHandler(
	userRepo *user.Repository,
	gamificationRepo *gamification.Repository,
	winRepo *win.Repository,
	jobRepo *job.Repository,
	certRepo *certification.Repository,
	skillRepo *skill.Repository,
	linkedAccountService *linkedaccount.Service,
	customDomainRepo *customdomain.Repository,
) *Handler {
	return &Handler{
		userRepo:             userRepo,
		gamificationRepo:     gamificationRepo,
		winRepo:              winRepo,
		jobRepo:              jobRepo,
		certRepo:             certRepo,
		skillRepo:            skillRepo,
		linkedAccountService: linkedAccountService,
		customDomainRepo:     customDomainRepo,
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
	Name              string          `json:"name"`
	Username          *string         `json:"username"`
	Bio               *string         `json:"bio"`
	Location          *string         `json:"location"`
	Headline          *string         `json:"headline"`
	OpenToWork        string          `json:"open_to_work"`
	ProfileVisibility json.RawMessage `json:"profile_visibility"`
}

// updateProfileRequest is the request body for PUT /user/me/profile.
type updateProfileRequest struct {
	Name              *string         `json:"name"`
	Username          *string         `json:"username"`
	Bio               *string         `json:"bio"`
	Location          *string         `json:"location"`
	Headline          *string         `json:"headline"`
	OpenToWork        *string         `json:"open_to_work"`
	ProfileVisibility json.RawMessage `json:"profile_visibility"`
}

// publicProfileResponse is the response for GET /profiles/:username.
type publicProfileResponse struct {
	Name           string                              `json:"name"`
	Bio            *string                             `json:"bio,omitempty"`
	AvatarURL      *string                             `json:"avatar_url,omitempty"`
	Location       *string                             `json:"location,omitempty"`
	Headline       *string                             `json:"headline,omitempty"`
	OpenToWork     string                              `json:"open_to_work"`
	IsStaff        bool                                `json:"is_staff"`
	IsSupporter    bool                                `json:"is_supporter"`
	LinkedAccounts []linkedaccount.PublicLinkedAccount `json:"linked_accounts"`
	Level          int                                 `json:"level"`
	LevelTitle     string                              `json:"level_title"`
	Badges         []gamification.Badge                `json:"badges"`
	Jobs           []job.Job                           `json:"jobs,omitempty"`
	Certifications []certification.Certification       `json:"certifications,omitempty"`
	Skills         []skill.Skill                       `json:"skills,omitempty"`
	Wins           []win.Win                           `json:"wins,omitempty"`
	IsCustomDomain bool                                `json:"is_custom_domain,omitempty"`
	AccentColour   string                              `json:"accent_colour,omitempty"`
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
		Name:              u.Name,
		Username:          u.Username,
		Bio:               u.Bio,
		Location:          u.Location,
		Headline:          u.Headline,
		OpenToWork:        u.OpenToWork,
		ProfileVisibility: u.ProfileVisibility,
	}

	response.OK(c, resp)
}

// UpdateProfileSettings updates the profile settings for the authenticated user.
func (h *Handler) UpdateProfileSettings(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	// Fetch current user to merge fields.
	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Determine final values, keeping existing if not provided.
	finalName := u.Name
	finalUsername := u.Username
	finalBio := u.Bio
	finalVisibility := u.ProfileVisibility
	finalLocation := u.Location
	finalHeadline := u.Headline
	finalOpenToWork := u.OpenToWork

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			response.BadRequest(c, "name is required")
			return
		}
		if len(trimmed) > 255 {
			response.BadRequest(c, "name must be at most 255 characters")
			return
		}
		finalName = trimmed
	}

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
		if len(*req.Bio) > 2000 {
			response.BadRequest(c, "bio must be at most 2000 characters")
			return
		}
		finalBio = req.Bio
	}

	if len(req.ProfileVisibility) > 0 {
		if len(req.ProfileVisibility) > 1024 {
			response.BadRequest(c, "profile visibility payload is too large")
			return
		}
		// Parse into the typed struct to strip unknown fields and ensure
		// only the expected boolean fields are stored.
		var vis visibilityConfig
		if err := json.Unmarshal(req.ProfileVisibility, &vis); err != nil {
			response.BadRequest(c, "profile visibility must be a valid JSON object")
			return
		}
		sanitised, err := json.Marshal(vis)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		finalVisibility = sanitised
	}

	if req.Location != nil {
		trimmed := strings.TrimSpace(*req.Location)
		if len(trimmed) > 255 {
			response.BadRequest(c, "location must be 255 characters or fewer")
			return
		}
		if trimmed == "" {
			finalLocation = nil
		} else {
			finalLocation = &trimmed
		}
	}

	if req.Headline != nil {
		trimmed := strings.TrimSpace(*req.Headline)
		if len(trimmed) > 255 {
			response.BadRequest(c, "headline must be 255 characters or fewer")
			return
		}
		if trimmed == "" {
			finalHeadline = nil
		} else {
			finalHeadline = &trimmed
		}
	}

	if req.OpenToWork != nil {
		val := strings.TrimSpace(*req.OpenToWork)
		switch val {
		case "not_looking", "open", "actively_looking":
			finalOpenToWork = val
		default:
			response.BadRequest(c, "open_to_work must be one of: not_looking, open, actively_looking")
			return
		}
	}

	if err := h.userRepo.UpdateProfile(c.Request.Context(), userID, finalName, finalUsername, finalBio, finalVisibility,
		finalLocation, finalHeadline, finalOpenToWork); err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, profileSettingsResponse{
		Name:              finalName,
		Username:          finalUsername,
		Bio:               finalBio,
		Location:          finalLocation,
		Headline:          finalHeadline,
		OpenToWork:        finalOpenToWork,
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

	u, err := h.userRepo.GetByUsername(c.Request.Context(), username)
	if err != nil {
		response.NotFound(c, "profile not found")
		return
	}

	resp, buildErr := h.buildPublicProfile(c, u)
	if buildErr != nil {
		return // buildPublicProfile already wrote the HTTP response
	}

	response.OK(c, resp)
}

// GetPublicProfileByDomain returns the public profile for a custom domain.
func (h *Handler) GetPublicProfileByDomain(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		response.BadRequest(c, "domain is required")
		return
	}

	ctx := c.Request.Context()

	cd, err := h.customDomainRepo.GetByDomain(ctx, strings.ToLower(strings.TrimSpace(domain)))
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if cd == nil || cd.Status != "active" {
		response.NotFound(c, "profile not found")
		return
	}

	u, err := h.userRepo.GetByID(ctx, cd.UserID)
	if err != nil {
		response.NotFound(c, "profile not found")
		return
	}

	// Verify the user is still a supporter.
	if !u.IsOneTimeSupporter && !u.IsSubscriber && !u.IsAdmin {
		response.NotFound(c, "profile not found")
		return
	}

	resp, buildErr := h.buildPublicProfile(c, u)
	if buildErr != nil {
		return // buildPublicProfile already wrote the HTTP response
	}

	resp.IsCustomDomain = true
	resp.AccentColour = cd.AccentColour

	response.OK(c, resp)
}

// buildPublicProfile constructs the public profile response for the given user.
// If an error occurs, it writes the HTTP response directly and returns a non-nil error
// to signal the caller should not write further.
func (h *Handler) buildPublicProfile(c *gin.Context, u user.User) (*publicProfileResponse, error) {
	ctx := c.Request.Context()

	// Determine visibility.
	vis := parseVisibility(u.ProfileVisibility)

	// Fetch gamification data (always visible).
	points, err := h.gamificationRepo.GetOrCreatePoints(ctx, u.ID)
	if err != nil {
		response.InternalError(c, err)
		return nil, err
	}
	level, levelTitle := gamification.LevelForPoints(points.TotalPoints)

	// Fetch earned badges (always visible).
	allBadges, err := h.gamificationRepo.ListBadgesWithEarnedStatus(ctx, u.ID)
	if err != nil {
		response.InternalError(c, err)
		return nil, err
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

	linkedAccounts, err := h.linkedAccountService.ListPublicByUser(ctx, u.ID)
	if err != nil {
		response.InternalError(c, err)
		return nil, err
	}

	resp := &publicProfileResponse{
		Name:           u.Name,
		Bio:            u.Bio,
		AvatarURL:      u.AvatarURL,
		Location:       u.Location,
		Headline:       u.Headline,
		OpenToWork:     u.OpenToWork,
		IsStaff:        u.IsAdmin,
		IsSupporter:    u.IsOneTimeSupporter || u.IsSubscriber,
		LinkedAccounts: linkedAccounts,
		Level:          level,
		LevelTitle:     levelTitle,
		Badges:         earnedBadges,
	}

	// Conditionally include sections based on visibility settings.
	if vis.Jobs {
		jobs, jobErr := h.jobRepo.List(ctx, u.ID)
		if jobErr != nil {
			response.InternalError(c, jobErr)
			return nil, jobErr
		}
		if jobs == nil {
			jobs = []job.Job{}
		}
		resp.Jobs = jobs
	}

	if vis.Certifications {
		certs, _, certErr := h.certRepo.List(ctx, u.ID, certification.ListParams{
			Limit:    100,
			Statuses: []string{"passed", "expired"},
		})
		if certErr != nil {
			response.InternalError(c, certErr)
			return nil, certErr
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
			return nil, skillErr
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
			return nil, winErr
		}
		if wins == nil {
			wins = []win.Win{}
		}
		resp.Wins = wins
	}

	return resp, nil
}
