package jobtitle

// JobTitleResult represents a single job title search result, returned to the
// frontend for autocomplete suggestions. Source indicates where the result came
// from: "user" for the user's own previous titles, "common" for the curated
// static list.
type JobTitleResult struct {
	Title  string `json:"title"`
	Source string `json:"source"`
}

// SearchResponse wraps the list of job title search results.
type SearchResponse struct {
	Results []JobTitleResult `json:"results"`
}
