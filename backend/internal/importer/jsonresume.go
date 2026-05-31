package importer

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// jsonResumeSchema represents the top-level structure of a JSON Resume file.
// See https://jsonresume.org/schema for the full specification.
type jsonResumeSchema struct {
	Basics       *jsonResumeBasics        `json:"basics"`
	Work         []jsonResumeWork         `json:"work"`
	Certificates []jsonResumeCertificate  `json:"certificates"`
	Skills       []jsonResumeSkill        `json:"skills"`
	Education    []jsonResumeEducation    `json:"education"`
}

type jsonResumeBasics struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

type jsonResumeWork struct {
	Company   string `json:"company"`
	Name      string `json:"name"` // some versions use "name" instead of "company"
	Position  string `json:"position"`
	URL       string `json:"url"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Summary   string `json:"summary"`
	Location  string `json:"location"`
}

type jsonResumeCertificate struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
	URL    string `json:"url"`
}

type jsonResumeSkill struct {
	Name     string   `json:"name"`
	Level    string   `json:"level"`
	Keywords []string `json:"keywords"`
}

type jsonResumeEducation struct {
	Institution string `json:"institution"`
	Area        string `json:"area"`
	StudyType   string `json:"studyType"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

// ParseJSONResume parses JSON Resume format data and returns an ImportPreview.
func ParseJSONResume(data []byte) (*ImportPreview, error) {
	var resume jsonResumeSchema
	if err := json.Unmarshal(data, &resume); err != nil {
		return nil, fmt.Errorf("parsing JSON Resume: %w", err)
	}

	preview := &ImportPreview{
		Source: "jsonresume",
		Jobs:   []JobPreview{},
		Certs:  []CertPreview{},
		Skills: []SkillPreview{},
		Wins:   []WinPreview{},
	}

	// Profile
	if resume.Basics != nil && (resume.Basics.Name != "" || resume.Basics.Summary != "") {
		preview.Profile = &ProfilePreview{
			Name: resume.Basics.Name,
			Bio:  resume.Basics.Summary,
		}
	}

	// Jobs
	for _, w := range resume.Work {
		company := w.Company
		if company == "" {
			company = w.Name
		}
		if company == "" && w.Position == "" {
			continue
		}

		jp := JobPreview{
			Company:          company,
			Title:            w.Position,
			StartDate:        normaliseJSONResumeDate(w.StartDate),
			EndDate:          normaliseJSONResumeDate(w.EndDate),
			Location:         w.Location,
			Responsibilities: w.Summary,
			Notes:            w.URL,
		}
		preview.Jobs = append(preview.Jobs, jp)
	}

	// Certifications
	for _, cert := range resume.Certificates {
		if cert.Name == "" {
			continue
		}

		earnedDate := normaliseJSONResumeDate(cert.Date)
		status := "planning"
		if earnedDate != "" {
			status = "passed"
		}

		cp := CertPreview{
			Name:          cert.Name,
			Provider:      cert.Issuer,
			Status:        status,
			EarnedDate:    earnedDate,
			CredentialURL: cert.URL,
		}
		preview.Certs = append(preview.Certs, cp)
	}

	// Skills
	for _, sk := range resume.Skills {
		if sk.Name == "" {
			continue
		}

		proficiency := mapJSONResumeProficiency(sk.Level)

		sp := SkillPreview{
			Name:        sk.Name,
			Proficiency: proficiency,
		}

		// Store keywords as category if present
		if len(sk.Keywords) > 0 {
			sp.Category = strings.Join(sk.Keywords, ", ")
		}

		preview.Skills = append(preview.Skills, sp)
	}

	return preview, nil
}

// normaliseJSONResumeDate handles "YYYY-MM-DD", "YYYY-MM", and "YYYY" formats,
// returning a normalised "YYYY-MM-DD" string.
func normaliseJSONResumeDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Try full date first: YYYY-MM-DD
	if _, err := time.Parse("2006-01-02", raw); err == nil {
		return raw
	}

	// Try YYYY-MM
	if t, err := time.Parse("2006-01", raw); err == nil {
		return t.Format("2006-01-02")
	}

	// Try year only
	if t, err := time.Parse("2006", raw); err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}

// mapJSONResumeProficiency converts JSON Resume level strings to numeric proficiency (1-4).
func mapJSONResumeProficiency(level string) int {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "master", "expert":
		return 4
	case "advanced":
		return 3
	case "intermediate":
		return 2
	case "beginner", "novice":
		return 1
	default:
		return 2 // default to intermediate
	}
}
