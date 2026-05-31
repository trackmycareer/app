package importer

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	maxDecompressedFileSize    = 50 << 20  // 50 MB per file within the ZIP
	maxTotalDecompressedSize   = 100 << 20 // 100 MB total across all files
	maxRowsPerFile             = 10_000
	maxFilesInZip              = 200
)

// ParseLinkedIn parses a LinkedIn data export ZIP and returns an ImportPreview.
// LinkedIn ZIPs contain CSV files for Positions, Certifications, Skills, and Profile.
// Missing files are tolerated; warnings are added for each absent file.
func ParseLinkedIn(zipData []byte) (*ImportPreview, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, fmt.Errorf("opening ZIP archive: %w", err)
	}

	if len(reader.File) > maxFilesInZip {
		return nil, fmt.Errorf("ZIP contains too many files (%d), maximum is %d", len(reader.File), maxFilesInZip)
	}

	fileMap := make(map[string]*zip.File)
	for _, f := range reader.File {
		// LinkedIn ZIPs sometimes nest files in subdirectories. Use the base name.
		name := f.Name
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		fileMap[strings.ToLower(name)] = f
	}

	preview := &ImportPreview{
		Source: "linkedin",
		Jobs:   []JobPreview{},
		Certs:  []CertPreview{},
		Skills: []SkillPreview{},
		Wins:   []WinPreview{},
	}

	// Track total decompressed bytes across all files to prevent zip bombs
	var totalDecompressed int64

	// Parse Profile.csv
	if err := parseLinkedInProfile(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Profile.csv: %v", err))
	}

	// Parse Positions.csv
	if err := parseLinkedInPositions(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Positions.csv: %v", err))
	}

	// Parse Certifications.csv
	if err := parseLinkedInCertifications(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Certifications.csv: %v", err))
	}

	// Parse Skills.csv
	if err := parseLinkedInSkills(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Skills.csv: %v", err))
	}

	// Parse Education.csv
	if err := parseLinkedInEducation(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Education.csv: %v", err))
	}

	// Parse Volunteering.csv
	if err := parseLinkedInVolunteering(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Volunteering.csv: %v", err))
	}

	// Parse Projects.csv
	if err := parseLinkedInProjects(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Projects.csv: %v", err))
	}

	// Parse Publications.csv
	if err := parseLinkedInPublications(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Publications.csv: %v", err))
	}

	// Parse Organizations.csv
	if err := parseLinkedInOrganizations(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Organizations.csv: %v", err))
	}

	// Parse Recommendations Received.csv
	if err := parseLinkedInRecommendations(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Recommendations_Received.csv: %v", err))
	}

	// Parse Endorsement Received Info.csv (must come after Skills)
	if err := parseLinkedInEndorsements(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Endorsement_Received_Info.csv: %v", err))
	}

	// Parse Learning.csv
	if err := parseLinkedInLearning(fileMap, preview, &totalDecompressed); err != nil {
		preview.Warnings = append(preview.Warnings, fmt.Sprintf("Learning.csv: %v", err))
	}

	return preview, nil
}

func readCSVFromZip(f *zip.File, totalDecompressed *int64) ([][]string, error) {
	if f.UncompressedSize64 > maxDecompressedFileSize {
		return nil, fmt.Errorf("file %s exceeds maximum decompressed size (%d bytes)", f.Name, f.UncompressedSize64)
	}

	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("opening file in ZIP: %w", err)
	}
	defer rc.Close()

	limited := io.LimitReader(rc, int64(maxDecompressedFileSize)+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("reading file from ZIP: %w", err)
	}

	*totalDecompressed += int64(len(data))
	if *totalDecompressed > maxTotalDecompressedSize {
		return nil, fmt.Errorf("total decompressed size exceeds maximum of %d MB", maxTotalDecompressedSize/(1<<20))
	}

	r := csv.NewReader(bytes.NewReader(data))
	r.LazyQuotes = true
	r.TrimLeadingSpace = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}
	if len(records) > maxRowsPerFile+1 {
		return nil, fmt.Errorf("file %s has too many rows (%d), maximum is %d", f.Name, len(records)-1, maxRowsPerFile)
	}
	return records, nil
}

// csvColumnIndex builds a case-insensitive header-to-index map from the first row.
func csvColumnIndex(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return idx
}

func csvGet(row []string, idx map[string]int, key string) string {
	i, ok := idx[strings.ToLower(key)]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseLinkedInProfile(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["profile.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return fmt.Errorf("no data rows")
	}

	idx := csvColumnIndex(records[0])
	row := records[1]

	firstName := csvGet(row, idx, "first name")
	lastName := csvGet(row, idx, "last name")
	headline := csvGet(row, idx, "headline")
	summary := csvGet(row, idx, "summary")

	name := strings.TrimSpace(firstName + " " + lastName)
	bio := headline
	if summary != "" {
		bio = summary
	}

	if name != "" || bio != "" {
		preview.Profile = &ProfilePreview{
			Name: name,
			Bio:  bio,
		}
	}

	return nil
}

func parseLinkedInPositions(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["positions.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil // headers only, no data
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		company := csvGet(row, idx, "company name")
		title := csvGet(row, idx, "title")
		if company == "" && title == "" {
			continue
		}

		jp := JobPreview{
			Company:          company,
			Title:            title,
			StartDate:        parseLinkedInDate(csvGet(row, idx, "started on")),
			EndDate:          parseLinkedInDate(csvGet(row, idx, "finished on")),
			Location:         csvGet(row, idx, "location"),
			Responsibilities: csvGet(row, idx, "description"),
		}
		preview.Jobs = append(preview.Jobs, jp)
	}

	return nil
}

func parseLinkedInCertifications(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["certifications.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		name := csvGet(row, idx, "name")
		if name == "" {
			continue
		}

		earnedDate := parseLinkedInDate(csvGet(row, idx, "started on"))
		status := "planning"
		if earnedDate != "" {
			status = "passed"
		}

		cp := CertPreview{
			Name:          name,
			Provider:      csvGet(row, idx, "authority"),
			Status:        status,
			EarnedDate:    earnedDate,
			ExpiryDate:    parseLinkedInDate(csvGet(row, idx, "finished on")),
			CredentialURL: csvGet(row, idx, "url"),
		}
		preview.Certs = append(preview.Certs, cp)
	}

	return nil
}

func parseLinkedInSkills(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["skills.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	// Skills.csv typically has a single "Name" column.
	for _, row := range records[1:] {
		if len(row) == 0 {
			continue
		}
		name := strings.TrimSpace(row[0])
		if name == "" {
			continue
		}
		preview.Skills = append(preview.Skills, SkillPreview{
			Name:        name,
			Proficiency: 2, // default: intermediate
		})
	}

	return nil
}

// parseLinkedInDate converts LinkedIn date formats to "YYYY-MM-DD".
// LinkedIn uses "Mon YYYY" (e.g., "Jan 2020") or just "YYYY".
// Returns empty string for unparseable or empty input.
func parseLinkedInDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Try "Mon YYYY" format (e.g., "Jan 2020")
	if t, err := time.Parse("Jan 2006", raw); err == nil {
		return t.Format("2006-01-02")
	}

	// Try full month name (e.g., "January 2020")
	if t, err := time.Parse("January 2006", raw); err == nil {
		return t.Format("2006-01-02")
	}

	// Try year only (e.g., "2020")
	if t, err := time.Parse("2006", raw); err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}

// parseLinkedInRecommendationDate converts the recommendation date format "MM/DD/YY, HH:MM AM/PM"
// to "YYYY-MM-DD". Returns empty string for unparseable or empty input.
func parseLinkedInRecommendationDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if t, err := time.Parse("01/02/06, 3:04 PM", raw); err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}

// parseLinkedInLearningDate converts the LinkedIn Learning completed date format
// "2006-01-02 15:04 UTC" to "YYYY-MM-DD". Falls back to parseLinkedInDate if
// the primary format fails. Returns empty string for unparseable or empty input.
func parseLinkedInLearningDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if t, err := time.Parse("2006-01-02 15:04 UTC", raw); err == nil {
		return t.Format("2006-01-02")
	}

	return parseLinkedInDate(raw)
}

func parseLinkedInEducation(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["education.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		school := csvGet(row, idx, "School Name")
		degree := csvGet(row, idx, "Degree Name")
		if school == "" && degree == "" {
			continue
		}

		jp := JobPreview{
			Company:          school,
			Title:            degree,
			StartDate:        parseLinkedInDate(csvGet(row, idx, "Start Date")),
			EndDate:          parseLinkedInDate(csvGet(row, idx, "End Date")),
			EmploymentType:   "education",
			Responsibilities: csvGet(row, idx, "Notes"),
			Notes:            csvGet(row, idx, "Activities"),
		}
		preview.Jobs = append(preview.Jobs, jp)
	}

	return nil
}

func parseLinkedInVolunteering(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["volunteering.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		company := csvGet(row, idx, "Company Name")
		role := csvGet(row, idx, "Role")
		if company == "" && role == "" {
			continue
		}

		notes := ""
		cause := csvGet(row, idx, "Cause")
		if cause != "" {
			notes = "Cause: " + cause
		}

		jp := JobPreview{
			Company:          company,
			Title:            role,
			StartDate:        parseLinkedInDate(csvGet(row, idx, "Started On")),
			EndDate:          parseLinkedInDate(csvGet(row, idx, "Finished On")),
			EmploymentType:   "volunteer",
			Responsibilities: csvGet(row, idx, "Description"),
			Notes:            notes,
		}
		preview.Jobs = append(preview.Jobs, jp)
	}

	return nil
}

func parseLinkedInProjects(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["projects.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		title := csvGet(row, idx, "Title")
		if title == "" {
			continue
		}

		description := csvGet(row, idx, "Description")
		url := csvGet(row, idx, "Url")
		if url != "" {
			if description != "" {
				description += "\n\n" + url
			} else {
				description = url
			}
		}

		wp := WinPreview{
			Title:       title,
			Description: description,
			OccurredOn:  parseLinkedInDate(csvGet(row, idx, "Started On")),
			Category:    "project",
		}
		preview.Wins = append(preview.Wins, wp)
	}

	return nil
}

func parseLinkedInPublications(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["publications.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		name := csvGet(row, idx, "Name")
		if name == "" {
			continue
		}

		var parts []string
		publisher := csvGet(row, idx, "Publisher")
		if publisher != "" {
			parts = append(parts, "Published by "+publisher)
		}
		desc := csvGet(row, idx, "Description")
		if desc != "" {
			parts = append(parts, desc)
		}
		url := csvGet(row, idx, "Url")
		if url != "" {
			parts = append(parts, url)
		}

		description := strings.Join(parts, "\n\n")

		wp := WinPreview{
			Title:       name,
			Description: description,
			OccurredOn:  parseLinkedInDate(csvGet(row, idx, "Published On")),
			Category:    "publication",
		}
		preview.Wins = append(preview.Wins, wp)
	}

	return nil
}

func parseLinkedInOrganizations(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["organizations.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		name := csvGet(row, idx, "Name")
		if name == "" {
			continue
		}

		title := name
		position := csvGet(row, idx, "Position")
		if position != "" {
			title = position + ", " + name
		}

		wp := WinPreview{
			Title:       title,
			Description: csvGet(row, idx, "Description"),
			OccurredOn:  parseLinkedInDate(csvGet(row, idx, "Started On")),
			Category:    "membership",
		}
		preview.Wins = append(preview.Wins, wp)
	}

	return nil
}

func parseLinkedInRecommendations(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["recommendations_received.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		status := csvGet(row, idx, "Status")
		if !strings.EqualFold(status, "VISIBLE") {
			continue
		}

		text := csvGet(row, idx, "Text")
		if text == "" {
			continue
		}

		firstName := csvGet(row, idx, "First Name")
		lastName := csvGet(row, idx, "Last Name")
		fullName := strings.TrimSpace(firstName + " " + lastName)
		title := "Recommendation from " + fullName

		jobTitle := csvGet(row, idx, "Job Title")
		company := csvGet(row, idx, "Company")
		if jobTitle != "" && company != "" {
			title += ", " + jobTitle + " at " + company
		} else if jobTitle != "" {
			title += ", " + jobTitle
		} else if company != "" {
			title += ", " + company
		}

		wp := WinPreview{
			Title:       title,
			Description: text,
			OccurredOn:  parseLinkedInRecommendationDate(csvGet(row, idx, "Creation Date")),
			Category:    "positive_feedback",
		}
		preview.Wins = append(preview.Wins, wp)
	}

	return nil
}

func parseLinkedInEndorsements(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["endorsement_received_info.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	// Count endorsements per skill (accepted only).
	endorsementCounts := make(map[string]int)
	for _, row := range records[1:] {
		status := csvGet(row, idx, "Endorsement Status")
		if !strings.EqualFold(status, "ACCEPTED") {
			continue
		}
		skillName := strings.ToLower(csvGet(row, idx, "Skill Name"))
		if skillName == "" {
			continue
		}
		endorsementCounts[skillName]++
	}

	// Build a set of existing skill names (lowercased) for quick lookup.
	existingSkills := make(map[string]bool, len(preview.Skills))
	for i := range preview.Skills {
		existingSkills[strings.ToLower(preview.Skills[i].Name)] = true
	}

	// Adjust proficiency of existing skills based on endorsement count.
	for i := range preview.Skills {
		key := strings.ToLower(preview.Skills[i].Name)
		count, ok := endorsementCounts[key]
		if !ok || count == 0 {
			continue
		}
		preview.Skills[i].Proficiency = endorsementProficiency(count)
	}

	// Add new skill entries for endorsed skills not already present.
	for skillName, count := range endorsementCounts {
		if existingSkills[skillName] {
			continue
		}
		// Use the original casing from the CSV. We stored lowercase keys, so
		// re-scan for the first matching row to recover casing.
		displayName := skillName
		for _, row := range records[1:] {
			candidate := csvGet(row, idx, "Skill Name")
			if strings.EqualFold(candidate, skillName) {
				displayName = candidate
				break
			}
		}
		preview.Skills = append(preview.Skills, SkillPreview{
			Name:        displayName,
			Proficiency: endorsementProficiency(count),
		})
	}

	return nil
}

// endorsementProficiency maps an endorsement count to a proficiency level.
func endorsementProficiency(count int) int {
	switch {
	case count >= 11:
		return 4
	case count >= 4:
		return 3
	case count >= 1:
		return 2
	default:
		return 1
	}
}

func parseLinkedInLearning(fileMap map[string]*zip.File, preview *ImportPreview, totalDecompressed *int64) error {
	f, ok := fileMap["learning.csv"]
	if !ok {
		return fmt.Errorf("file not found in ZIP")
	}

	records, err := readCSVFromZip(f, totalDecompressed)
	if err != nil {
		return err
	}
	if len(records) < 2 {
		return nil
	}

	idx := csvColumnIndex(records[0])

	for _, row := range records[1:] {
		completedAt := csvGet(row, idx, "Content Completed At (if completed)")
		if completedAt == "" || strings.EqualFold(completedAt, "N/A") {
			continue
		}

		title := csvGet(row, idx, "Content Title")
		if title == "" {
			continue
		}

		cp := CertPreview{
			Name:       title,
			Provider:   "LinkedIn Learning",
			Status:     "passed",
			EarnedDate: parseLinkedInLearningDate(completedAt),
		}
		preview.Certs = append(preview.Certs, cp)
	}

	return nil
}

