package location

// LocationResult represents a single location search result, returned to the
// frontend for autocomplete suggestions. Source indicates where the result
// originated ("user" for the user's own data, "photon" for OpenStreetMap).
type LocationResult struct {
	Label  string `json:"label"`
	Source string `json:"source"`
}

// SearchResponse wraps the list of location search results.
type SearchResponse struct {
	Results []LocationResult `json:"results"`
}
