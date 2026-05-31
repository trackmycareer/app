package importer

// ImportPreview holds the parsed data from any import source before the user confirms saving.
type ImportPreview struct {
	Source   string          `json:"source"`
	Profile  *ProfilePreview `json:"profile,omitempty"`
	Jobs     []JobPreview    `json:"jobs"`
	Certs    []CertPreview   `json:"certifications"`
	Skills   []SkillPreview  `json:"skills"`
	Wins     []WinPreview    `json:"wins"`
	Warnings []string        `json:"warnings"`
}

// ProfilePreview contains optional profile data extracted from the import source.
type ProfilePreview struct {
	Name string `json:"name,omitempty"`
	Bio  string `json:"bio,omitempty"`
}

// JobPreview represents a single job extracted from the import source.
// Date strings are in "YYYY-MM-DD" format.
type JobPreview struct {
	Company          string `json:"company"`
	Title            string `json:"title"`
	StartDate        string `json:"start_date"`
	EndDate          string `json:"end_date,omitempty"`
	EmploymentType   string `json:"employment_type,omitempty"`
	Location         string `json:"location,omitempty"`
	WorkMode         string `json:"work_mode,omitempty"`
	Responsibilities string `json:"responsibilities,omitempty"`
	Notes            string `json:"notes,omitempty"`
}

// CertPreview represents a single certification extracted from the import source.
type CertPreview struct {
	Name          string `json:"name"`
	Provider      string `json:"provider"`
	Status        string `json:"status,omitempty"`
	EarnedDate    string `json:"earned_date,omitempty"`
	ExpiryDate    string `json:"expiry_date,omitempty"`
	CredentialURL string `json:"credential_url,omitempty"`
}

// SkillPreview represents a single skill extracted from the import source.
type SkillPreview struct {
	Name        string `json:"name"`
	Category    string `json:"category,omitempty"`
	Proficiency int    `json:"proficiency,omitempty"`
}

// WinPreview represents a single win extracted from the import source.
type WinPreview struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	OccurredOn  string `json:"occurred_on,omitempty"`
	Category    string `json:"category,omitempty"`
}

// ImportResult summarises the outcome of a confirmed import.
type ImportResult struct {
	JobsCreated    int  `json:"jobs_created"`
	CertsCreated   int  `json:"certifications_created"`
	SkillsCreated  int  `json:"skills_created"`
	WinsCreated    int  `json:"wins_created"`
	ProfileUpdated bool `json:"profile_updated"`
}
