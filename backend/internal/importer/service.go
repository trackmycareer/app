package importer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/certification"
	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/internal/job"
	"github.com/trackmycareer/app/internal/skill"
	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/internal/win"
	"github.com/trackmycareer/app/pkg/types"
)

// Service orchestrates CV import parsing and record creation.
type Service struct {
	jobRepo      *job.Repository
	certRepo     *certification.Repository
	skillRepo    *skill.Repository
	winRepo      *win.Repository
	userRepo     *user.Repository
	gamification *gamification.Service
}

// NewService creates a new importer Service with the required dependencies.
func NewService(
	jobRepo *job.Repository,
	certRepo *certification.Repository,
	skillRepo *skill.Repository,
	winRepo *win.Repository,
	userRepo *user.Repository,
	gamificationSvc *gamification.Service,
) *Service {
	return &Service{
		jobRepo:      jobRepo,
		certRepo:     certRepo,
		skillRepo:    skillRepo,
		winRepo:      winRepo,
		userRepo:     userRepo,
		gamification: gamificationSvc,
	}
}

// Preview parses file data and returns a preview of what would be imported.
// The source parameter selects the parser: "linkedin", "jsonresume", "csv", or "document".
// If source is empty, it is auto-detected from the filename.
func (s *Service) Preview(fileData []byte, filename, source string) (*ImportPreview, error) {
	switch source {
	case "linkedin":
		return ParseLinkedIn(fileData)
	case "jsonresume":
		return ParseJSONResume(fileData)
	case "csv":
		// For CSV, auto-detect entity type from filename or default to "jobs"
		entityType := detectCSVEntityType(filename)
		return ParseCSV(fileData, entityType)
	case "document":
		return ParseDocument(fileData, filename)
	default:
		return nil, fmt.Errorf("unsupported import source: %s", source)
	}
}

// Confirm creates records in the database from a confirmed ImportPreview.
// It skips duplicates and returns a summary of what was created.
func (s *Service) Confirm(ctx context.Context, userID uuid.UUID, preview *ImportPreview) (*ImportResult, error) {
	result := &ImportResult{}

	// Load existing records to check for duplicates
	existingJobs, err := s.jobRepo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listing existing jobs: %w", err)
	}

	existingCerts, _, err := s.certRepo.List(ctx, userID, certification.ListParams{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("listing existing certifications: %w", err)
	}

	existingSkills, _, err := s.skillRepo.List(ctx, userID, skill.ListParams{Limit: 1000})
	if err != nil {
		return nil, fmt.Errorf("listing existing skills: %w", err)
	}

	// Create jobs
	for _, jp := range preview.Jobs {
		if isDuplicateJob(jp, existingJobs) {
			continue
		}

		j := job.Job{
			ID:             uuid.New(),
			UserID:         userID,
			Company:        jp.Company,
			Title:          jp.Title,
			EmploymentType: jp.EmploymentType,
			WorkMode:       jp.WorkMode,
		}

		if jp.StartDate != "" {
			d, parseErr := types.ParseDate(jp.StartDate)
			if parseErr == nil {
				j.StartDate = d
			}
		}
		if jp.EndDate != "" {
			d, parseErr := types.ParseDate(jp.EndDate)
			if parseErr == nil {
				j.EndDate = &d
			}
		}
		if j.EmploymentType == "" {
			j.EmploymentType = "full_time"
		}
		if jp.Location != "" {
			j.Location = &jp.Location
		}
		if jp.Responsibilities != "" {
			j.Responsibilities = &jp.Responsibilities
		}
		if jp.Notes != "" {
			j.Notes = &jp.Notes
		}

		if createErr := s.jobRepo.Create(ctx, &j); createErr != nil {
			slog.Error("import: failed to create job", "error", createErr, "company", jp.Company, "title", jp.Title)
			continue
		}

		result.JobsCreated++
		s.recordActivity(ctx, userID, "job_created", &j.ID)
	}

	// Create certifications
	for _, cp := range preview.Certs {
		if isDuplicateCert(cp, existingCerts) {
			continue
		}

		c := certification.Certification{
			ID:       uuid.New(),
			UserID:   userID,
			Name:     cp.Name,
			Provider: cp.Provider,
			Status:   cp.Status,
			Currency: "GBP",
		}

		if c.Status == "" {
			if cp.EarnedDate != "" {
				c.Status = "passed"
			} else {
				c.Status = "planning"
			}
		}

		if cp.EarnedDate != "" {
			d, parseErr := types.ParseDate(cp.EarnedDate)
			if parseErr == nil {
				c.EarnedDate = &d
			}
		}
		if cp.ExpiryDate != "" {
			d, parseErr := types.ParseDate(cp.ExpiryDate)
			if parseErr == nil {
				c.ExpiryDate = &d
			}
		}
		if cp.CredentialURL != "" {
			c.CredentialURL = &cp.CredentialURL
		}

		if createErr := s.certRepo.Create(ctx, &c); createErr != nil {
			slog.Error("import: failed to create certification", "error", createErr, "name", cp.Name)
			continue
		}

		result.CertsCreated++
		action := "cert_created"
		if c.Status == "passed" {
			action = "cert_passed"
		}
		s.recordActivity(ctx, userID, action, &c.ID)
	}

	// Create skills
	for _, sp := range preview.Skills {
		if isDuplicateSkill(sp, existingSkills) {
			continue
		}

		sk := skill.Skill{
			ID:          uuid.New(),
			UserID:      userID,
			Name:        sp.Name,
			Proficiency: sp.Proficiency,
		}

		if sk.Proficiency < 1 || sk.Proficiency > 4 {
			sk.Proficiency = 2
		}

		if sp.Category != "" {
			sk.Category = &sp.Category
		}

		if createErr := s.skillRepo.Create(ctx, &sk); createErr != nil {
			slog.Error("import: failed to create skill", "error", createErr, "name", sp.Name)
			continue
		}

		result.SkillsCreated++
		s.recordActivity(ctx, userID, "skill_created", &sk.ID)
	}

	// Create wins
	for _, wp := range preview.Wins {
		w := win.Win{
			ID:       uuid.New(),
			UserID:   userID,
			Title:    wp.Title,
			Category: wp.Category,
		}

		if w.Category == "" {
			w.Category = "general"
		}

		if wp.OccurredOn != "" {
			d, parseErr := types.ParseDate(wp.OccurredOn)
			if parseErr == nil {
				w.OccurredOn = d
			} else {
				w.OccurredOn = types.Today()
			}
		} else {
			w.OccurredOn = types.Today()
		}

		if wp.Description != "" {
			w.Description = &wp.Description
		}

		if createErr := s.winRepo.Create(ctx, &w, nil); createErr != nil {
			slog.Error("import: failed to create win", "error", createErr, "title", wp.Title)
			continue
		}

		result.WinsCreated++
		s.recordActivity(ctx, userID, "win_created", &w.ID)
	}

	// Update profile if data is present
	if preview.Profile != nil && (preview.Profile.Name != "" || preview.Profile.Bio != "") {
		existingUser, userErr := s.userRepo.GetByID(ctx, userID)
		if userErr == nil {
			updated := false

			if preview.Profile.Name != "" && existingUser.Name == "" {
				existingUser.Name = preview.Profile.Name
				updated = true
			}
			if preview.Profile.Bio != "" && (existingUser.Bio == nil || *existingUser.Bio == "") {
				existingUser.Bio = &preview.Profile.Bio
				updated = true
			}

			if updated {
				if updateErr := s.userRepo.Update(ctx, &existingUser); updateErr != nil {
					slog.Error("import: failed to update profile", "error", updateErr, "user_id", userID)
				} else {
					result.ProfileUpdated = true
				}
			}
		}
	}

	return result, nil
}

// recordActivity calls gamification and logs any error without failing the import.
func (s *Service) recordActivity(ctx context.Context, userID uuid.UUID, action string, entityID *uuid.UUID) {
	if _, err := s.gamification.RecordActivity(ctx, userID, action, entityID); err != nil {
		slog.Error("import: gamification recording failed",
			"error", err, "action", action, "user_id", userID,
		)
	}
}

// isDuplicateJob checks whether a job with the same company, title, and start date already exists.
func isDuplicateJob(jp JobPreview, existing []job.Job) bool {
	for _, j := range existing {
		if strings.EqualFold(j.Company, jp.Company) &&
			strings.EqualFold(j.Title, jp.Title) &&
			j.StartDate.String() == jp.StartDate {
			return true
		}
	}
	return false
}

// isDuplicateCert checks whether a certification with the same name and provider already exists.
func isDuplicateCert(cp CertPreview, existing []certification.Certification) bool {
	for _, c := range existing {
		if strings.EqualFold(c.Name, cp.Name) && strings.EqualFold(c.Provider, cp.Provider) {
			return true
		}
	}
	return false
}

// isDuplicateSkill checks whether a skill with the same name already exists.
func isDuplicateSkill(sp SkillPreview, existing []skill.Skill) bool {
	for _, s := range existing {
		if strings.EqualFold(s.Name, sp.Name) {
			return true
		}
	}
	return false
}

// detectCSVEntityType attempts to determine the entity type from the filename.
func detectCSVEntityType(filename string) string {
	lower := strings.ToLower(filename)
	switch {
	case strings.Contains(lower, "cert"):
		return "certifications"
	case strings.Contains(lower, "skill"):
		return "skills"
	case strings.Contains(lower, "win"):
		return "wins"
	default:
		return "jobs"
	}
}
