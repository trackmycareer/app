package export

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trackmycareer/app/pkg/types"
)

// ExportData represents the full user data export.
type ExportData struct {
	ExportedAt     time.Time     `json:"exported_at"`
	User           exportUser    `json:"user"`
	Wins           []exportWin   `json:"wins"`
	Jobs           []exportJob   `json:"jobs"`
	Certifications []exportCert  `json:"certifications"`
	Skills         []exportSkill `json:"skills"`
	Tags           []exportTag   `json:"tags"`
}

type exportUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type exportTag struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Colour string    `json:"colour"`
}

type exportWin struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	OccurredOn  types.Date `json:"occurred_on"`
	Category    string     `json:"category"`
	Tags        []string   `json:"tags"`
}

type exportJob struct {
	Company          string      `json:"company"`
	Title            string      `json:"title"`
	StartDate        types.Date  `json:"start_date"`
	EndDate          *types.Date `json:"end_date,omitempty"`
	EmploymentType   string      `json:"employment_type"`
	Location         *string     `json:"location,omitempty"`
	WorkMode         string      `json:"work_mode"`
	Responsibilities *string     `json:"responsibilities,omitempty"`
}

type exportCert struct {
	Name          string      `json:"name"`
	Provider      string      `json:"provider"`
	Status        string      `json:"status"`
	EarnedDate    *types.Date `json:"earned_date,omitempty"`
	ExpiryDate    *types.Date `json:"expiry_date,omitempty"`
	CredentialURL *string     `json:"credential_url,omitempty"`
}

type exportSkill struct {
	Name        string  `json:"name"`
	Category    *string `json:"category,omitempty"`
	Proficiency int     `json:"proficiency"`
	Notes       *string `json:"notes,omitempty"`
}

// Service provides data export operations.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a new export service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// ExportJSON fetches all user data and returns it as a structured export.
func (s *Service) ExportJSON(ctx context.Context, userID uuid.UUID) (ExportData, error) {
	data := ExportData{
		ExportedAt: time.Now().UTC(),
	}

	// Fetch user info
	err := s.pool.QueryRow(ctx,
		`SELECT name, email FROM users WHERE id = $1`, userID,
	).Scan(&data.User.Name, &data.User.Email)
	if err != nil {
		return ExportData{}, fmt.Errorf("fetching user for export: %w", err)
	}

	// Fetch tags
	tagRows, err := s.pool.Query(ctx,
		`SELECT id, name, colour FROM tags WHERE user_id = $1 ORDER BY name`, userID)
	if err != nil {
		return ExportData{}, fmt.Errorf("fetching tags for export: %w", err)
	}
	defer tagRows.Close()

	tagNames := make(map[uuid.UUID]string)
	data.Tags = []exportTag{}
	for tagRows.Next() {
		var t exportTag
		if err := tagRows.Scan(&t.ID, &t.Name, &t.Colour); err != nil {
			return ExportData{}, fmt.Errorf("scanning tag: %w", err)
		}
		tagNames[t.ID] = t.Name
		data.Tags = append(data.Tags, t)
	}
	if err := tagRows.Err(); err != nil {
		return ExportData{}, fmt.Errorf("iterating tags: %w", err)
	}

	// Fetch wins with tags
	winRows, err := s.pool.Query(ctx,
		`SELECT w.id, w.title, w.description, w.occurred_on, w.category
		FROM wins w WHERE w.user_id = $1 ORDER BY w.occurred_on DESC`, userID)
	if err != nil {
		return ExportData{}, fmt.Errorf("fetching wins for export: %w", err)
	}
	defer winRows.Close()

	type winEntry struct {
		id  uuid.UUID
		win exportWin
	}
	var winEntries []winEntry
	data.Wins = []exportWin{}
	for winRows.Next() {
		var we winEntry
		if err := winRows.Scan(&we.id, &we.win.Title, &we.win.Description, &we.win.OccurredOn, &we.win.Category); err != nil {
			return ExportData{}, fmt.Errorf("scanning win: %w", err)
		}
		we.win.Tags = []string{}
		winEntries = append(winEntries, we)
	}
	if err := winRows.Err(); err != nil {
		return ExportData{}, fmt.Errorf("iterating wins: %w", err)
	}

	// Load win tags
	if len(winEntries) > 0 {
		winIDs := make([]uuid.UUID, len(winEntries))
		for i, we := range winEntries {
			winIDs[i] = we.id
		}
		wtRows, err := s.pool.Query(ctx,
			`SELECT wt.win_id, t.name FROM win_tags wt
			JOIN tags t ON t.id = wt.tag_id
			WHERE wt.win_id = ANY($1)
			ORDER BY t.name`, winIDs)
		if err != nil {
			return ExportData{}, fmt.Errorf("fetching win tags for export: %w", err)
		}
		defer wtRows.Close()

		winTagMap := make(map[uuid.UUID][]string)
		for wtRows.Next() {
			var winID uuid.UUID
			var tagName string
			if err := wtRows.Scan(&winID, &tagName); err != nil {
				return ExportData{}, fmt.Errorf("scanning win tag: %w", err)
			}
			winTagMap[winID] = append(winTagMap[winID], tagName)
		}
		if err := wtRows.Err(); err != nil {
			return ExportData{}, fmt.Errorf("iterating win tags: %w", err)
		}

		for i, we := range winEntries {
			if tags, ok := winTagMap[we.id]; ok {
				winEntries[i].win.Tags = tags
			}
		}
	}

	for _, we := range winEntries {
		data.Wins = append(data.Wins, we.win)
	}

	// Fetch jobs
	jobRows, err := s.pool.Query(ctx,
		`SELECT company, title, start_date, end_date, employment_type, location, work_mode, responsibilities
		FROM jobs WHERE user_id = $1 ORDER BY start_date DESC`, userID)
	if err != nil {
		return ExportData{}, fmt.Errorf("fetching jobs for export: %w", err)
	}
	defer jobRows.Close()

	data.Jobs = []exportJob{}
	for jobRows.Next() {
		var j exportJob
		if err := jobRows.Scan(&j.Company, &j.Title, &j.StartDate, &j.EndDate,
			&j.EmploymentType, &j.Location, &j.WorkMode, &j.Responsibilities); err != nil {
			return ExportData{}, fmt.Errorf("scanning job: %w", err)
		}
		data.Jobs = append(data.Jobs, j)
	}
	if err := jobRows.Err(); err != nil {
		return ExportData{}, fmt.Errorf("iterating jobs: %w", err)
	}

	// Fetch certifications
	certRows, err := s.pool.Query(ctx,
		`SELECT name, provider, status, earned_date, expiry_date, credential_url
		FROM certifications WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return ExportData{}, fmt.Errorf("fetching certifications for export: %w", err)
	}
	defer certRows.Close()

	data.Certifications = []exportCert{}
	for certRows.Next() {
		var ce exportCert
		if err := certRows.Scan(&ce.Name, &ce.Provider, &ce.Status,
			&ce.EarnedDate, &ce.ExpiryDate, &ce.CredentialURL); err != nil {
			return ExportData{}, fmt.Errorf("scanning certification: %w", err)
		}
		data.Certifications = append(data.Certifications, ce)
	}
	if err := certRows.Err(); err != nil {
		return ExportData{}, fmt.Errorf("iterating certifications: %w", err)
	}

	// Fetch skills
	skillRows, err := s.pool.Query(ctx,
		`SELECT name, category, proficiency, notes
		FROM skills WHERE user_id = $1 ORDER BY name`, userID)
	if err != nil {
		return ExportData{}, fmt.Errorf("fetching skills for export: %w", err)
	}
	defer skillRows.Close()

	data.Skills = []exportSkill{}
	for skillRows.Next() {
		var sk exportSkill
		if err := skillRows.Scan(&sk.Name, &sk.Category, &sk.Proficiency, &sk.Notes); err != nil {
			return ExportData{}, fmt.Errorf("scanning skill: %w", err)
		}
		data.Skills = append(data.Skills, sk)
	}
	if err := skillRows.Err(); err != nil {
		return ExportData{}, fmt.Errorf("iterating skills: %w", err)
	}

	return data, nil
}

// ExportMarkdown fetches all user data and formats it as a Markdown document.
func (s *Service) ExportMarkdown(ctx context.Context, userID uuid.UUID) (string, error) {
	data, err := s.ExportJSON(ctx, userID)
	if err != nil {
		return "", err
	}

	var b strings.Builder

	b.WriteString("# Career Portfolio\n\n")
	b.WriteString(fmt.Sprintf("**Name:** %s\n", data.User.Name))
	b.WriteString(fmt.Sprintf("**Email:** %s\n", data.User.Email))
	b.WriteString(fmt.Sprintf("**Exported:** %s\n\n", data.ExportedAt.Format("2 January 2006")))

	// Jobs
	if len(data.Jobs) > 0 {
		b.WriteString("## Work Experience\n\n")
		for _, j := range data.Jobs {
			dateRange := j.StartDate.String()
			if j.EndDate != nil {
				dateRange += " to " + j.EndDate.String()
			} else {
				dateRange += " to Present"
			}
			b.WriteString(fmt.Sprintf("### %s at %s\n\n", j.Title, j.Company))
			b.WriteString(fmt.Sprintf("**Period:** %s\n", dateRange))
			b.WriteString(fmt.Sprintf("**Type:** %s\n", j.EmploymentType))
			if j.Location != nil && *j.Location != "" {
				b.WriteString(fmt.Sprintf("**Location:** %s\n", *j.Location))
			}
			if j.WorkMode != "" && j.WorkMode != "onsite" {
				b.WriteString(fmt.Sprintf("**Work mode:** %s\n", j.WorkMode))
			}
			if j.Responsibilities != nil && *j.Responsibilities != "" {
				b.WriteString(fmt.Sprintf("\n%s\n", *j.Responsibilities))
			}
			b.WriteString("\n")
		}
	}

	// Wins
	if len(data.Wins) > 0 {
		b.WriteString("## Achievements\n\n")
		for _, w := range data.Wins {
			b.WriteString(fmt.Sprintf("### %s\n\n", w.Title))
			b.WriteString(fmt.Sprintf("**Date:** %s\n", w.OccurredOn.String()))
			b.WriteString(fmt.Sprintf("**Category:** %s\n", w.Category))
			if len(w.Tags) > 0 {
				b.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(w.Tags, ", ")))
			}
			if w.Description != nil && *w.Description != "" {
				b.WriteString(fmt.Sprintf("\n%s\n", *w.Description))
			}
			b.WriteString("\n")
		}
	}

	// Certifications
	if len(data.Certifications) > 0 {
		b.WriteString("## Certifications\n\n")
		for _, c := range data.Certifications {
			b.WriteString(fmt.Sprintf("### %s\n\n", c.Name))
			b.WriteString(fmt.Sprintf("**Provider:** %s\n", c.Provider))
			b.WriteString(fmt.Sprintf("**Status:** %s\n", c.Status))
			if c.EarnedDate != nil && !c.EarnedDate.IsZero() {
				b.WriteString(fmt.Sprintf("**Earned:** %s\n", c.EarnedDate.String()))
			}
			if c.ExpiryDate != nil && !c.ExpiryDate.IsZero() {
				b.WriteString(fmt.Sprintf("**Expires:** %s\n", c.ExpiryDate.String()))
			}
			if c.CredentialURL != nil && *c.CredentialURL != "" {
				b.WriteString(fmt.Sprintf("**Credential:** %s\n", *c.CredentialURL))
			}
			b.WriteString("\n")
		}
	}

	// Skills
	if len(data.Skills) > 0 {
		b.WriteString("## Skills\n\n")
		b.WriteString("| Skill | Category | Proficiency |\n")
		b.WriteString("|-------|----------|-------------|\n")
		for _, sk := range data.Skills {
			cat := ""
			if sk.Category != nil {
				cat = *sk.Category
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %d/5 |\n", sk.Name, cat, sk.Proficiency))
		}
		b.WriteString("\n")
	}

	// Tags
	if len(data.Tags) > 0 {
		b.WriteString("## Tags\n\n")
		tagStrs := make([]string, len(data.Tags))
		for i, t := range data.Tags {
			tagStrs[i] = t.Name
		}
		b.WriteString(strings.Join(tagStrs, ", "))
		b.WriteString("\n")
	}

	return b.String(), nil
}
