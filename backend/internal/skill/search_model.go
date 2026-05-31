package skill

// SkillSearchResult represents a single skill search result, returned to
// the frontend for autocomplete suggestions. Source indicates where the result
// came from: "user" for the user's own previous skills, "common" for
// the curated static list.
type SkillSearchResult struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Source   string `json:"source"`
}

// SkillSearchResponse wraps the list of skill search results.
type SkillSearchResponse struct {
	Results []SkillSearchResult `json:"results"`
}
