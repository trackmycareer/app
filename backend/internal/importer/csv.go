package importer

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
)

// Expected CSV headers for each entity type.
var expectedHeaders = map[string][]string{
	"jobs":           {"company", "title", "start_date", "end_date", "employment_type", "location", "work_mode", "responsibilities", "notes"},
	"certifications": {"name", "provider", "status", "earned_date", "expiry_date", "cost", "currency", "credential_url"},
	"skills":         {"name", "category", "proficiency", "notes"},
	"wins":           {"title", "description", "occurred_on", "category"},
}

// ParseCSV parses a CSV file for the given entity type and returns an ImportPreview.
func ParseCSV(data []byte, entityType string) (*ImportPreview, error) {
	expected, ok := expectedHeaders[entityType]
	if !ok {
		return nil, fmt.Errorf("unsupported entity type: %s (expected one of: jobs, certifications, skills, wins)", entityType)
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	// Validate headers (case-insensitive)
	headers := make([]string, len(records[0]))
	for i, h := range records[0] {
		headers[i] = strings.ToLower(strings.TrimSpace(h))
	}

	if err := validateHeaders(headers, expected); err != nil {
		return nil, fmt.Errorf("invalid CSV headers for %s: %w", entityType, err)
	}

	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[h] = i
	}

	preview := &ImportPreview{
		Source: "csv",
		Jobs:   []JobPreview{},
		Certs:  []CertPreview{},
		Skills: []SkillPreview{},
		Wins:   []WinPreview{},
	}

	for i, row := range records[1:] {
		if isEmptyRow(row) {
			continue
		}

		switch entityType {
		case "jobs":
			jp, warn := parseCSVJobRow(row, idx, i+2)
			if warn != "" {
				preview.Warnings = append(preview.Warnings, warn)
				continue
			}
			preview.Jobs = append(preview.Jobs, jp)

		case "certifications":
			cp, warn := parseCSVCertRow(row, idx, i+2)
			if warn != "" {
				preview.Warnings = append(preview.Warnings, warn)
				continue
			}
			preview.Certs = append(preview.Certs, cp)

		case "skills":
			sp, warn := parseCSVSkillRow(row, idx, i+2)
			if warn != "" {
				preview.Warnings = append(preview.Warnings, warn)
				continue
			}
			preview.Skills = append(preview.Skills, sp)

		case "wins":
			wp, warn := parseCSVWinRow(row, idx, i+2)
			if warn != "" {
				preview.Warnings = append(preview.Warnings, warn)
				continue
			}
			preview.Wins = append(preview.Wins, wp)
		}
	}

	return preview, nil
}

// GenerateTemplate returns the CSV header row (with an example data row) for the given entity type.
func GenerateTemplate(entityType string) ([]byte, error) {
	headers, ok := expectedHeaders[entityType]
	if !ok {
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(headers); err != nil {
		return nil, fmt.Errorf("writing CSV headers: %w", err)
	}

	// Write an example row
	example := templateExampleRow(entityType)
	if len(example) > 0 {
		if err := w.Write(example); err != nil {
			return nil, fmt.Errorf("writing CSV example row: %w", err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flushing CSV writer: %w", err)
	}

	return buf.Bytes(), nil
}

func templateExampleRow(entityType string) []string {
	switch entityType {
	case "jobs":
		return []string{"Acme Corp", "Software Engineer", "2023-01-15", "2024-06-30", "full_time", "London", "false", "Built microservices architecture", "Great team"}
	case "certifications":
		return []string{"AWS Solutions Architect", "Amazon Web Services", "passed", "2023-06-15", "2026-06-15", "300", "GBP", "https://example.com/cert"}
	case "skills":
		return []string{"Go", "programming", "3", "Primary backend language"}
	case "wins":
		return []string{"Reduced API latency by 40%", "Optimised database queries and added caching layer", "2024-03-15", "process_improvement"}
	default:
		return nil
	}
}

func validateHeaders(actual, expected []string) error {
	if len(actual) < len(expected) {
		return fmt.Errorf("expected at least %d columns, got %d", len(expected), len(actual))
	}

	for i, exp := range expected {
		if actual[i] != exp {
			return fmt.Errorf("column %d: expected %q, got %q", i+1, exp, actual[i])
		}
	}
	return nil
}

func isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func csvGetIdx(row []string, idx map[string]int, key string) string {
	i, ok := idx[key]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseCSVJobRow(row []string, idx map[string]int, lineNum int) (JobPreview, string) {
	company := csvGetIdx(row, idx, "company")
	title := csvGetIdx(row, idx, "title")
	if company == "" || title == "" {
		return JobPreview{}, fmt.Sprintf("row %d: company and title are required", lineNum)
	}

	workMode := strings.ToLower(csvGetIdx(row, idx, "work_mode"))
	if workMode != "hybrid" && workMode != "remote" {
		workMode = "onsite"
	}

	return JobPreview{
		Company:          company,
		Title:            title,
		StartDate:        csvGetIdx(row, idx, "start_date"),
		EndDate:          csvGetIdx(row, idx, "end_date"),
		EmploymentType:   csvGetIdx(row, idx, "employment_type"),
		Location:         csvGetIdx(row, idx, "location"),
		WorkMode:         workMode,
		Responsibilities: csvGetIdx(row, idx, "responsibilities"),
		Notes:            csvGetIdx(row, idx, "notes"),
	}, ""
}

func parseCSVCertRow(row []string, idx map[string]int, lineNum int) (CertPreview, string) {
	name := csvGetIdx(row, idx, "name")
	if name == "" {
		return CertPreview{}, fmt.Sprintf("row %d: name is required", lineNum)
	}

	return CertPreview{
		Name:          name,
		Provider:      csvGetIdx(row, idx, "provider"),
		Status:        csvGetIdx(row, idx, "status"),
		EarnedDate:    csvGetIdx(row, idx, "earned_date"),
		ExpiryDate:    csvGetIdx(row, idx, "expiry_date"),
		CredentialURL: csvGetIdx(row, idx, "credential_url"),
	}, ""
}

func parseCSVSkillRow(row []string, idx map[string]int, lineNum int) (SkillPreview, string) {
	name := csvGetIdx(row, idx, "name")
	if name == "" {
		return SkillPreview{}, fmt.Sprintf("row %d: name is required", lineNum)
	}

	proficiency := 2
	if p := csvGetIdx(row, idx, "proficiency"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v >= 1 && v <= 4 {
			proficiency = v
		}
	}

	return SkillPreview{
		Name:        name,
		Category:    csvGetIdx(row, idx, "category"),
		Proficiency: proficiency,
	}, ""
}

func parseCSVWinRow(row []string, idx map[string]int, lineNum int) (WinPreview, string) {
	title := csvGetIdx(row, idx, "title")
	if title == "" {
		return WinPreview{}, fmt.Sprintf("row %d: title is required", lineNum)
	}

	return WinPreview{
		Title:       title,
		Description: csvGetIdx(row, idx, "description"),
		OccurredOn:  csvGetIdx(row, idx, "occurred_on"),
		Category:    csvGetIdx(row, idx, "category"),
	}, ""
}
