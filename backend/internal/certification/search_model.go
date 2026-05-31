package certification

// CertSearchResult represents a single certification search result, returned to
// the frontend for autocomplete suggestions. Source indicates where the result
// came from: "user" for the user's own previous certifications, "common" for
// the curated static list.
type CertSearchResult struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Source   string `json:"source"`
}

// CertSearchResponse wraps the list of certification search results.
type CertSearchResponse struct {
	Results []CertSearchResult `json:"results"`
}
