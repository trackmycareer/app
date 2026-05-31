package company

// CompanyResult represents a single company search result, returned to the frontend
// for autocomplete suggestions. Source indicates where the result came from.
type CompanyResult struct {
	Name          string `json:"name"`
	Source        string `json:"source"`
	CompanyNumber string `json:"company_number,omitempty"`
	Status        string `json:"status,omitempty"`
}

// SearchResponse wraps the list of company search results.
type SearchResponse struct {
	Results []CompanyResult `json:"results"`
}

