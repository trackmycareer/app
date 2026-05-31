package importer

import (
	"archive/zip"
	"bytes"
	"os"
	"strings"
	"testing"
)

// createTestZip builds an in-memory ZIP archive from a filename-to-content map.
func createTestZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestParseLinkedInDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "abbreviated month and year", input: "Jan 2020", want: "2020-01-01"},
		{name: "full month and year", input: "January 2020", want: "2020-01-01"},
		{name: "year only", input: "2020", want: "2020-01-01"},
		{name: "empty string", input: "", want: ""},
		{name: "invalid format", input: "invalid", want: ""},
		{name: "December abbreviated", input: "Dec 2014", want: "2014-12-01"},
		{name: "whitespace only", input: "   ", want: ""},
		{name: "leading and trailing whitespace", input: "  Mar 2022  ", want: "2022-03-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLinkedInDate(tt.input)
			if got != tt.want {
				t.Errorf("parseLinkedInDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseLinkedInRecommendationDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "standard format PM", input: "08/30/22, 05:51 PM", want: "2022-08-30"},
		{name: "standard format PM 2", input: "01/16/19, 07:04 PM", want: "2019-01-16"},
		{name: "empty string", input: "", want: ""},
		{name: "invalid format", input: "invalid", want: ""},
		{name: "wrong date format", input: "2022-08-30", want: ""},
		{name: "AM time", input: "06/05/18, 06:49 AM", want: "2018-06-05"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLinkedInRecommendationDate(tt.input)
			if got != tt.want {
				t.Errorf("parseLinkedInRecommendationDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseLinkedInLearningDate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "standard UTC format", input: "2020-04-15 00:07 UTC", want: "2020-04-15"},
		{name: "fallback to parseLinkedInDate", input: "Jan 2020", want: "2020-01-01"},
		{name: "empty string", input: "", want: ""},
		{name: "invalid format", input: "invalid", want: ""},
		{name: "year only fallback", input: "2021", want: "2021-01-01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLinkedInLearningDate(tt.input)
			if got != tt.want {
				t.Errorf("parseLinkedInLearningDate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEndorsementProficiency(t *testing.T) {
	tests := []struct {
		name  string
		count int
		want  int
	}{
		{name: "zero endorsements", count: 0, want: 1},
		{name: "one endorsement", count: 1, want: 2},
		{name: "three endorsements", count: 3, want: 2},
		{name: "four endorsements", count: 4, want: 3},
		{name: "ten endorsements", count: 10, want: 3},
		{name: "eleven endorsements", count: 11, want: 4},
		{name: "fifty endorsements", count: 50, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := endorsementProficiency(tt.count)
			if got != tt.want {
				t.Errorf("endorsementProficiency(%d) = %d, want %d", tt.count, got, tt.want)
			}
		})
	}
}

func TestParseLinkedInEducation(t *testing.T) {
	csvContent := "School Name,Start Date,End Date,Notes,Degree Name,Activities\n" +
		"MIT,Sep 2019,Jul 2021,Some modules,BSc Computer Science,Club activities\n" +
		",,,,, \n"

	zipData := createTestZip(t, map[string]string{
		"Education.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	// Filter education entries from jobs.
	var eduJobs []JobPreview
	for _, j := range preview.Jobs {
		if j.EmploymentType == "education" {
			eduJobs = append(eduJobs, j)
		}
	}

	if len(eduJobs) != 1 {
		t.Fatalf("expected 1 education job, got %d", len(eduJobs))
	}

	job := eduJobs[0]

	if job.Company != "MIT" {
		t.Errorf("Company = %q, want %q", job.Company, "MIT")
	}
	if job.Title != "BSc Computer Science" {
		t.Errorf("Title = %q, want %q", job.Title, "BSc Computer Science")
	}
	if job.StartDate != "2019-09-01" {
		t.Errorf("StartDate = %q, want %q", job.StartDate, "2019-09-01")
	}
	if job.EndDate != "2021-07-01" {
		t.Errorf("EndDate = %q, want %q", job.EndDate, "2021-07-01")
	}
	if job.Responsibilities != "Some modules" {
		t.Errorf("Responsibilities = %q, want %q", job.Responsibilities, "Some modules")
	}
	if job.Notes != "Club activities" {
		t.Errorf("Notes = %q, want %q", job.Notes, "Club activities")
	}
	if job.EmploymentType != "education" {
		t.Errorf("EmploymentType = %q, want %q", job.EmploymentType, "education")
	}
}

func TestParseLinkedInVolunteering(t *testing.T) {
	csvContent := "Company Name,Role,Cause,Started On,Finished On,Description\n" +
		"Red Cross,Volunteer Coordinator,humanitarian,May 2015,Jan 2019,Helped with relief efforts\n"

	zipData := createTestZip(t, map[string]string{
		"Volunteering.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	var volJobs []JobPreview
	for _, j := range preview.Jobs {
		if j.EmploymentType == "volunteer" {
			volJobs = append(volJobs, j)
		}
	}

	if len(volJobs) != 1 {
		t.Fatalf("expected 1 volunteer job, got %d", len(volJobs))
	}

	job := volJobs[0]

	if job.Company != "Red Cross" {
		t.Errorf("Company = %q, want %q", job.Company, "Red Cross")
	}
	if job.Title != "Volunteer Coordinator" {
		t.Errorf("Title = %q, want %q", job.Title, "Volunteer Coordinator")
	}
	if job.StartDate != "2015-05-01" {
		t.Errorf("StartDate = %q, want %q", job.StartDate, "2015-05-01")
	}
	if job.EndDate != "2019-01-01" {
		t.Errorf("EndDate = %q, want %q", job.EndDate, "2019-01-01")
	}
	if job.Notes != "Cause: humanitarian" {
		t.Errorf("Notes = %q, want %q", job.Notes, "Cause: humanitarian")
	}
	if job.Responsibilities != "Helped with relief efforts" {
		t.Errorf("Responsibilities = %q, want %q", job.Responsibilities, "Helped with relief efforts")
	}
	if job.EmploymentType != "volunteer" {
		t.Errorf("EmploymentType = %q, want %q", job.EmploymentType, "volunteer")
	}
}

func TestParseLinkedInProjects(t *testing.T) {
	csvContent := "Title,Description,Url,Started On,Finished On\n" +
		"Cool Project,Built something great,https://example.com,Jan 2019,Mar 2019\n" +
		"URL Only,,https://example.com/2,Feb 2020,\n" +
		"No URL,Just a description,,Mar 2021,\n"

	zipData := createTestZip(t, map[string]string{
		"Projects.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Wins) != 3 {
		t.Fatalf("expected 3 wins, got %d", len(preview.Wins))
	}

	tests := []struct {
		name        string
		win         WinPreview
		wantTitle   string
		wantDesc    string
		wantCat     string
		wantDate    string
	}{
		{
			name:     "project with description and URL",
			win:      preview.Wins[0],
			wantTitle: "Cool Project",
			wantDesc: "Built something great\n\nhttps://example.com",
			wantCat:  "project",
			wantDate: "2019-01-01",
		},
		{
			name:     "project with URL only",
			win:      preview.Wins[1],
			wantTitle: "URL Only",
			wantDesc: "https://example.com/2",
			wantCat:  "project",
			wantDate: "2020-02-01",
		},
		{
			name:     "project with description only",
			win:      preview.Wins[2],
			wantTitle: "No URL",
			wantDesc: "Just a description",
			wantCat:  "project",
			wantDate: "2021-03-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.win.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", tt.win.Title, tt.wantTitle)
			}
			if tt.win.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", tt.win.Description, tt.wantDesc)
			}
			if tt.win.Category != tt.wantCat {
				t.Errorf("Category = %q, want %q", tt.win.Category, tt.wantCat)
			}
			if tt.win.OccurredOn != tt.wantDate {
				t.Errorf("OccurredOn = %q, want %q", tt.win.OccurredOn, tt.wantDate)
			}
		})
	}
}

func TestParseLinkedInPublications(t *testing.T) {
	csvContent := "Name,Published On,Description,Publisher,Url\n" +
		"My Article,Aug 2015,A great read,Tech Blog,https://example.com/article\n" +
		"Minimal,Jan 2020,,,\n"

	zipData := createTestZip(t, map[string]string{
		"Publications.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Wins) != 2 {
		t.Fatalf("expected 2 wins, got %d", len(preview.Wins))
	}

	// First publication: full details.
	w0 := preview.Wins[0]
	if w0.Title != "My Article" {
		t.Errorf("Win[0].Title = %q, want %q", w0.Title, "My Article")
	}
	wantDesc := "Published by Tech Blog\n\nA great read\n\nhttps://example.com/article"
	if w0.Description != wantDesc {
		t.Errorf("Win[0].Description = %q, want %q", w0.Description, wantDesc)
	}
	if w0.Category != "publication" {
		t.Errorf("Win[0].Category = %q, want %q", w0.Category, "publication")
	}
	if w0.OccurredOn != "2015-08-01" {
		t.Errorf("Win[0].OccurredOn = %q, want %q", w0.OccurredOn, "2015-08-01")
	}

	// Second publication: minimal, no publisher/description/URL.
	w1 := preview.Wins[1]
	if w1.Title != "Minimal" {
		t.Errorf("Win[1].Title = %q, want %q", w1.Title, "Minimal")
	}
	if w1.Description != "" {
		t.Errorf("Win[1].Description = %q, want %q", w1.Description, "")
	}
	if w1.Category != "publication" {
		t.Errorf("Win[1].Category = %q, want %q", w1.Category, "publication")
	}
}

func TestParseLinkedInOrganizations(t *testing.T) {
	csvContent := "Name,Description,Position,Started On,Finished On\n" +
		"BCS,,MBCS,Mar 2026,\n" +
		"IEEE,Professional body,,Jan 2020,\n"

	zipData := createTestZip(t, map[string]string{
		"Organizations.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Wins) != 2 {
		t.Fatalf("expected 2 wins, got %d", len(preview.Wins))
	}

	// First: position + name.
	w0 := preview.Wins[0]
	if w0.Title != "MBCS, BCS" {
		t.Errorf("Win[0].Title = %q, want %q", w0.Title, "MBCS, BCS")
	}
	if w0.Category != "membership" {
		t.Errorf("Win[0].Category = %q, want %q", w0.Category, "membership")
	}
	if w0.OccurredOn != "2026-03-01" {
		t.Errorf("Win[0].OccurredOn = %q, want %q", w0.OccurredOn, "2026-03-01")
	}

	// Second: name only, no position.
	w1 := preview.Wins[1]
	if w1.Title != "IEEE" {
		t.Errorf("Win[1].Title = %q, want %q", w1.Title, "IEEE")
	}
	if w1.Description != "Professional body" {
		t.Errorf("Win[1].Description = %q, want %q", w1.Description, "Professional body")
	}
	if w1.Category != "membership" {
		t.Errorf("Win[1].Category = %q, want %q", w1.Category, "membership")
	}
}

func TestParseLinkedInRecommendations(t *testing.T) {
	csvContent := "First Name,Last Name,Company,Job Title,Text,Creation Date,Status\n" +
		"Alice,Smith,Acme Inc,CTO,Great colleague,\"08/30/22, 05:51 PM\",VISIBLE\n" +
		"Bob,Jones,,,Hidden rec,\"01/16/19, 07:04 PM\",HIDDEN\n" +
		"Charlie,Brown,Big Corp,,Visible no title,\"06/05/18, 06:49 PM\",VISIBLE\n"

	zipData := createTestZip(t, map[string]string{
		"Recommendations_Received.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Wins) != 2 {
		t.Fatalf("expected 2 wins (HIDDEN filtered out), got %d", len(preview.Wins))
	}

	// First: full details with job title and company.
	w0 := preview.Wins[0]
	wantTitle0 := "Recommendation from Alice Smith, CTO at Acme Inc"
	if w0.Title != wantTitle0 {
		t.Errorf("Win[0].Title = %q, want %q", w0.Title, wantTitle0)
	}
	if w0.Category != "positive_feedback" {
		t.Errorf("Win[0].Category = %q, want %q", w0.Category, "positive_feedback")
	}
	if w0.Description != "Great colleague" {
		t.Errorf("Win[0].Description = %q, want %q", w0.Description, "Great colleague")
	}
	if w0.OccurredOn != "2022-08-30" {
		t.Errorf("Win[0].OccurredOn = %q, want %q", w0.OccurredOn, "2022-08-30")
	}

	// Second: company but no job title.
	w1 := preview.Wins[1]
	wantTitle1 := "Recommendation from Charlie Brown, Big Corp"
	if w1.Title != wantTitle1 {
		t.Errorf("Win[1].Title = %q, want %q", w1.Title, wantTitle1)
	}
	if w1.OccurredOn != "2018-06-05" {
		t.Errorf("Win[1].OccurredOn = %q, want %q", w1.OccurredOn, "2018-06-05")
	}
}

func TestParseLinkedInEndorsements(t *testing.T) {
	skillsCSV := "Name\nGo\nPython\n"

	endorsementsCSV := "Endorsement Date,Skill Name,Endorser First Name,Endorser Last Name,Endorser Public Url,Endorsement Status\n" +
		"2023/06/18 14:24:54 UTC,Go,Alice,Smith,url1,ACCEPTED\n" +
		"2023/06/18 14:24:53 UTC,Go,Bob,Jones,url2,ACCEPTED\n" +
		"2023/06/18 14:24:46 UTC,Go,Charlie,Brown,url3,ACCEPTED\n" +
		"2023/06/18 14:24:44 UTC,Go,Dave,Wilson,url4,ACCEPTED\n" +
		"2023/06/18 14:24:42 UTC,Python,Eve,Davis,url5,ACCEPTED\n" +
		"2023/06/18 14:24:41 UTC,JavaScript,Frank,Miller,url6,ACCEPTED\n" +
		"2023/06/18 14:24:33 UTC,Go,Grace,Lee,url7,REJECTED\n"

	zipData := createTestZip(t, map[string]string{
		"Skills.csv":                   skillsCSV,
		"Endorsement_Received_Info.csv": endorsementsCSV,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Skills) != 3 {
		t.Fatalf("expected 3 skills, got %d: %+v", len(preview.Skills), preview.Skills)
	}

	// Build a map for easy lookup.
	skillMap := make(map[string]SkillPreview)
	for _, s := range preview.Skills {
		skillMap[s.Name] = s
	}

	// Go: 4 accepted endorsements -> proficiency 3.
	if go_, ok := skillMap["Go"]; !ok {
		t.Error("missing skill: Go")
	} else if go_.Proficiency != 3 {
		t.Errorf("Go proficiency = %d, want 3", go_.Proficiency)
	}

	// Python: 1 accepted endorsement -> proficiency 2.
	if py, ok := skillMap["Python"]; !ok {
		t.Error("missing skill: Python")
	} else if py.Proficiency != 2 {
		t.Errorf("Python proficiency = %d, want 2", py.Proficiency)
	}

	// JavaScript: not in Skills.csv but has 1 accepted endorsement -> added with proficiency 2.
	if js, ok := skillMap["JavaScript"]; !ok {
		t.Error("missing skill: JavaScript (should be added from endorsements)")
	} else if js.Proficiency != 2 {
		t.Errorf("JavaScript proficiency = %d, want 2", js.Proficiency)
	}
}

func TestParseLinkedInLearning(t *testing.T) {
	csvContent := "Content Title,Content Description,Content Type,Content Last Watched Date (if viewed),Content Completed At (if completed),Content Saved,Notes taken on videos (if taken)\n" +
		"Completed Course,Some description,Course,2020-04-15 00:07 UTC,2020-04-20 10:00 UTC,false,N/A\n" +
		"Not Completed,Another desc,Course,2021-06-14 12:16 UTC,N/A,false,N/A\n"

	zipData := createTestZip(t, map[string]string{
		"Learning.csv": csvContent,
	})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Certs) != 1 {
		t.Fatalf("expected 1 cert, got %d", len(preview.Certs))
	}

	cert := preview.Certs[0]

	if cert.Name != "Completed Course" {
		t.Errorf("Name = %q, want %q", cert.Name, "Completed Course")
	}
	if cert.Provider != "LinkedIn Learning" {
		t.Errorf("Provider = %q, want %q", cert.Provider, "LinkedIn Learning")
	}
	if cert.Status != "passed" {
		t.Errorf("Status = %q, want %q", cert.Status, "passed")
	}
	if cert.EarnedDate != "2020-04-20" {
		t.Errorf("EarnedDate = %q, want %q", cert.EarnedDate, "2020-04-20")
	}
}

func TestParseLinkedInMissingFiles(t *testing.T) {
	zipData := createTestZip(t, map[string]string{})

	preview, err := ParseLinkedIn(zipData)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(preview.Jobs))
	}
	if len(preview.Certs) != 0 {
		t.Errorf("expected 0 certs, got %d", len(preview.Certs))
	}
	if len(preview.Skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(preview.Skills))
	}
	if len(preview.Wins) != 0 {
		t.Errorf("expected 0 wins, got %d", len(preview.Wins))
	}
	if preview.Profile != nil {
		t.Error("expected nil profile")
	}

	expectedFiles := []string{
		"Profile.csv",
		"Positions.csv",
		"Certifications.csv",
		"Skills.csv",
		"Education.csv",
		"Volunteering.csv",
		"Projects.csv",
		"Publications.csv",
		"Organizations.csv",
		"Recommendations_Received.csv",
		"Endorsement_Received_Info.csv",
		"Learning.csv",
	}

	if len(preview.Warnings) != len(expectedFiles) {
		t.Errorf("expected %d warnings, got %d: %v", len(expectedFiles), len(preview.Warnings), preview.Warnings)
	}

	for _, file := range expectedFiles {
		found := false
		for _, w := range preview.Warnings {
			if strings.Contains(w, file) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected warning containing %q, not found in %v", file, preview.Warnings)
		}
	}
}

func TestParseLinkedInFullExport(t *testing.T) {
	const zipPath = "/Users/ben/Documents/bhcloudlabs/trackmy.career/Complete_LinkedInDataExport_05-29-2026.zip.zip"

	data, err := os.ReadFile(zipPath)
	if err != nil {
		t.Skipf("skipping full export test: %v", err)
	}

	preview, err := ParseLinkedIn(data)
	if err != nil {
		t.Fatalf("ParseLinkedIn() error = %v", err)
	}

	if len(preview.Jobs) == 0 {
		t.Error("expected at least one job")
	}
	if len(preview.Certs) == 0 {
		t.Error("expected at least one certification")
	}
	if len(preview.Skills) == 0 {
		t.Error("expected at least one skill")
	}
	if len(preview.Wins) == 0 {
		t.Error("expected at least one win")
	}

	if preview.Profile == nil {
		t.Fatal("expected non-nil profile")
	}
	if preview.Profile.Name == "" {
		t.Error("expected non-empty profile name")
	}

	// Check for a specific company.
	foundSugarAI := false
	for _, j := range preview.Jobs {
		if j.Company == "SugarAI" {
			foundSugarAI = true
			break
		}
	}
	if !foundSugarAI {
		t.Error("expected at least one job with Company = \"SugarAI\"")
	}

	// Check that education entries exist.
	foundEducation := false
	for _, j := range preview.Jobs {
		if j.EmploymentType == "education" {
			foundEducation = true
			break
		}
	}
	if !foundEducation {
		t.Error("expected at least one job with EmploymentType = \"education\"")
	}

	// Check that volunteer entries exist.
	foundVolunteer := false
	for _, j := range preview.Jobs {
		if j.EmploymentType == "volunteer" {
			foundVolunteer = true
			break
		}
	}
	if !foundVolunteer {
		t.Error("expected at least one job with EmploymentType = \"volunteer\"")
	}

	// Core files should all be present, so no warnings for them.
	coreFiles := []string{
		"Profile.csv",
		"Positions.csv",
		"Certifications.csv",
		"Skills.csv",
		"Education.csv",
	}
	for _, file := range coreFiles {
		for _, w := range preview.Warnings {
			if strings.Contains(w, file) {
				t.Errorf("unexpected warning for core file %q: %s", file, w)
			}
		}
	}
}
