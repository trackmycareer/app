package importer

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	pdflib "github.com/ledongthuc/pdf"
)

// ParseDocument attempts heuristic extraction of career data from PDF or DOCX files.
// Detection is based on the filename extension.
func ParseDocument(data []byte, filename string) (*ImportPreview, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	var text string
	var err error

	switch ext {
	case ".pdf":
		text, err = extractTextFromPDF(data)
	case ".docx":
		text, err = extractTextFromDOCX(data)
	case ".doc":
		return nil, fmt.Errorf("the legacy .doc format is not supported; please convert to .docx or .pdf")
	default:
		return nil, fmt.Errorf("unsupported file extension: %s (expected .pdf or .docx)", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("extracting text from %s: %w", ext, err)
	}

	return parseDocumentText(text), nil
}

// extractTextFromPDF reads text from PDF data using github.com/ledongthuc/pdf.
func extractTextFromPDF(data []byte) (string, error) {
	reader, err := pdflib.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("opening PDF: %w", err)
	}

	var sb strings.Builder
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		text, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		sb.WriteString(text)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// extractTextFromDOCX parses a DOCX file (ZIP containing XML) and extracts plain text
// from word/document.xml by stripping XML tags.
func extractTextFromDOCX(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("opening DOCX ZIP: %w", err)
	}

	for _, f := range reader.File {
		if strings.EqualFold(f.Name, "word/document.xml") {
			rc, openErr := f.Open()
			if openErr != nil {
				return "", fmt.Errorf("opening document.xml: %w", openErr)
			}
			defer rc.Close()

			xmlData, readErr := io.ReadAll(rc)
			if readErr != nil {
				return "", fmt.Errorf("reading document.xml: %w", readErr)
			}

			return extractTextFromWordXML(xmlData)
		}
	}

	return "", fmt.Errorf("word/document.xml not found in DOCX archive")
}

// extractTextFromWordXML extracts plain text from Word XML by walking the XML tokens
// and collecting text within <w:t> elements. Paragraph boundaries become newlines.
func extractTextFromWordXML(data []byte) (string, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var sb strings.Builder
	inText := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				inText = true
			}
			if t.Name.Local == "p" {
				sb.WriteString("\n")
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				sb.Write(t)
			}
		}
	}

	return sb.String(), nil
}

// Section header patterns used to detect common CV sections.
var sectionPatterns = map[string]*regexp.Regexp{
	"experience":     regexp.MustCompile(`(?im)^[\s]*(?:work\s+)?experience|employment\s+history|professional\s+experience|work\s+history`),
	"certifications": regexp.MustCompile(`(?im)^[\s]*certifications?|certificates?|licen[cs]es?`),
	"skills":         regexp.MustCompile(`(?im)^[\s]*(?:technical\s+)?skills|core\s+competencies|technologies`),
	"education":      regexp.MustCompile(`(?im)^[\s]*education`),
}

// Date range patterns to detect employment periods.
var dateRangePatterns = []*regexp.Regexp{
	// "Jan 2020 - Present", "January 2020 – Dec 2023"
	regexp.MustCompile(`(?i)((?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+\d{4})\s*[-–—]+\s*(Present|(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|Jun(?:e)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\s+\d{4})`),
	// "2020 - Present", "2020 - 2023"
	regexp.MustCompile(`(\d{4})\s*[-–—]+\s*(Present|\d{4})`),
}

// parseDocumentText applies heuristics to extract structured data from plain text.
func parseDocumentText(text string) *ImportPreview {
	preview := &ImportPreview{
		Source: "document",
		Jobs:   []JobPreview{},
		Certs:  []CertPreview{},
		Skills: []SkillPreview{},
		Wins:   []WinPreview{},
		Warnings: []string{
			"Data extracted from document using heuristic parsing. Please review all entries carefully.",
		},
	}

	lines := strings.Split(text, "\n")
	sections := splitIntoSections(lines)

	// Extract jobs from experience section
	if expLines, ok := sections["experience"]; ok {
		preview.Jobs = extractJobsFromSection(expLines)
		if len(preview.Jobs) == 0 {
			preview.Warnings = append(preview.Warnings, "Experience section found but no job entries could be extracted.")
		}
	} else {
		// Try to extract jobs from the full text
		preview.Jobs = extractJobsFromSection(lines)
		if len(preview.Jobs) > 0 {
			preview.Warnings = append(preview.Warnings, "No explicit experience section header found. Jobs extracted from general text.")
		}
	}

	// Extract certifications
	if certLines, ok := sections["certifications"]; ok {
		preview.Certs = extractCertsFromSection(certLines)
	}

	// Extract skills
	if skillLines, ok := sections["skills"]; ok {
		preview.Skills = extractSkillsFromSection(skillLines)
	}

	return preview
}

// splitIntoSections divides lines into named sections based on header patterns.
func splitIntoSections(lines []string) map[string][]string {
	sections := make(map[string][]string)

	type sectionRange struct {
		name  string
		start int
	}

	var detected []sectionRange

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		for name, pat := range sectionPatterns {
			if pat.MatchString(trimmed) && len(trimmed) < 60 {
				detected = append(detected, sectionRange{name: name, start: i + 1})
				break
			}
		}
	}

	for i, d := range detected {
		end := len(lines)
		if i+1 < len(detected) {
			end = detected[i+1].start - 1
		}
		if d.start < end {
			sections[d.name] = lines[d.start:end]
		}
	}

	return sections
}

// extractJobsFromSection attempts to find job entries with date ranges and nearby titles.
func extractJobsFromSection(lines []string) []JobPreview {
	var jobs []JobPreview

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Look for date ranges in this line
		for _, pat := range dateRangePatterns {
			matches := pat.FindStringSubmatch(trimmed)
			if len(matches) < 3 {
				continue
			}

			startDateStr := parseDocDate(matches[1])
			endDateStr := ""
			if strings.ToLower(matches[2]) != "present" {
				endDateStr = parseDocDate(matches[2])
			}

			// Try to extract title and company from surrounding lines
			title, company := extractTitleCompany(lines, i, trimmed, pat)

			if title == "" && company == "" {
				continue
			}

			jp := JobPreview{
				Company:   company,
				Title:     title,
				StartDate: startDateStr,
				EndDate:   endDateStr,
			}
			jobs = append(jobs, jp)
			break // Only match one pattern per line
		}
	}

	return jobs
}

// extractTitleCompany tries to identify the job title and company name near a date range line.
func extractTitleCompany(lines []string, dateLineIdx int, dateLine string, datePat *regexp.Regexp) (string, string) {
	// Remove the date range from the line to get remaining text
	remaining := strings.TrimSpace(datePat.ReplaceAllString(dateLine, ""))
	remaining = strings.Trim(remaining, "|-–—,. ")

	// If the remaining text contains a separator (| or -), split into title/company
	if parts := splitTitleCompany(remaining); len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	// Check the line above for title/company
	title := remaining
	company := ""

	if dateLineIdx > 0 {
		prevLine := strings.TrimSpace(lines[dateLineIdx-1])
		if prevLine != "" && len(prevLine) < 100 {
			if title == "" {
				title = prevLine
			} else {
				company = prevLine
			}
		}
	}

	return title, company
}

// splitTitleCompany attempts to split text like "Software Engineer | Acme Corp" or
// "Software Engineer at Acme Corp".
func splitTitleCompany(text string) []string {
	for _, sep := range []string{"|", " at ", " @ "} {
		if parts := strings.SplitN(text, sep, 2); len(parts) == 2 {
			a := strings.TrimSpace(parts[0])
			b := strings.TrimSpace(parts[1])
			if a != "" && b != "" {
				return []string{a, b}
			}
		}
	}
	return nil
}

// extractCertsFromSection parses certification names from section lines.
func extractCertsFromSection(lines []string) []CertPreview {
	var certs []CertPreview

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || len(trimmed) < 3 {
			continue
		}

		// Remove bullet points and list markers
		trimmed = strings.TrimLeft(trimmed, "•·-*▪► ")
		trimmed = strings.TrimSpace(trimmed)
		if trimmed == "" {
			continue
		}

		// Skip lines that look like section headers
		isHeader := false
		for _, pat := range sectionPatterns {
			if pat.MatchString(trimmed) {
				isHeader = true
				break
			}
		}
		if isHeader {
			continue
		}

		cp := CertPreview{
			Name:   trimmed,
			Status: "planning",
		}
		certs = append(certs, cp)
	}

	return certs
}

// extractSkillsFromSection splits skill lines by common delimiters.
func extractSkillsFromSection(lines []string) []SkillPreview {
	var skills []SkillPreview
	seen := make(map[string]bool)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || len(trimmed) < 2 {
			continue
		}

		// Remove bullet markers
		trimmed = strings.TrimLeft(trimmed, "•·-*▪► ")
		trimmed = strings.TrimSpace(trimmed)

		// Skip lines that look like section headers
		isHeader := false
		for _, pat := range sectionPatterns {
			if pat.MatchString(trimmed) {
				isHeader = true
				break
			}
		}
		if isHeader {
			continue
		}

		// Skills are often comma-separated, pipe-separated, or one per line
		var candidates []string
		if strings.Contains(trimmed, ",") {
			candidates = strings.Split(trimmed, ",")
		} else if strings.Contains(trimmed, "|") {
			candidates = strings.Split(trimmed, "|")
		} else if strings.Contains(trimmed, "·") {
			candidates = strings.Split(trimmed, "·")
		} else {
			candidates = []string{trimmed}
		}

		for _, c := range candidates {
			name := strings.TrimSpace(c)
			if name == "" || len(name) > 50 {
				continue
			}
			lower := strings.ToLower(name)
			if seen[lower] {
				continue
			}
			seen[lower] = true
			skills = append(skills, SkillPreview{
				Name:        name,
				Proficiency: 2,
			})
		}
	}

	return skills
}

// parseDocDate converts informal date strings from CVs to "YYYY-MM-DD".
func parseDocDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Try "Jan 2020" or "January 2020"
	for _, layout := range []string{"Jan 2006", "January 2006"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Format("2006-01-02")
		}
	}

	// Try year only
	if t, err := time.Parse("2006", raw); err == nil {
		return t.Format("2006-01-02")
	}

	return ""
}
