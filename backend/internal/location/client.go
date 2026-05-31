package location

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultPhotonBaseURL = "https://photon.komoot.io"

// photonResponse represents the GeoJSON FeatureCollection returned by the
// Photon geocoder API.
type photonResponse struct {
	Features []photonFeature `json:"features"`
}

type photonFeature struct {
	Properties photonProperties `json:"properties"`
}

type photonProperties struct {
	Name        string `json:"name"`
	City        string `json:"city"`
	State       string `json:"state"`
	Country     string `json:"country"`
	CountryCode string `json:"countrycode"`
	OSMKey      string `json:"osm_key"`
	OSMValue    string `json:"osm_value"`
	Type        string `json:"type"`
}

// PhotonClient calls the Photon geocoder API (free OpenStreetMap geocoder).
// No API key is required.
type PhotonClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPhotonClient creates a client for the Photon geocoder API.
// If baseURL is empty, the default public endpoint is used.
func NewPhotonClient(baseURL string) *PhotonClient {
	if baseURL == "" {
		baseURL = defaultPhotonBaseURL
	}
	return &PhotonClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// Search queries the Photon API for locations matching the given query.
// Results are formatted as clean location strings (e.g. "London, United Kingdom").
func (c *PhotonClient) Search(ctx context.Context, query string) ([]LocationResult, error) {
	endpoint := fmt.Sprintf(
		"%s/api?q=%s&limit=5&lang=en",
		c.baseURL, url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from Photon API", resp.StatusCode)
	}

	var pResp photonResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&pResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Format and deduplicate results.
	seen := make(map[string]bool)
	results := make([]LocationResult, 0, len(pResp.Features))
	for _, f := range pResp.Features {
		label := formatLocation(f.Properties)
		if label == "" {
			continue
		}
		key := strings.ToLower(label)
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, LocationResult{
			Label:  label,
			Source: "photon",
		})
	}

	return results, nil
}

// formatLocation builds a human-readable location string from Photon properties.
// The format is "City, State, Country" where state is omitted if it matches the
// primary name or is empty.
func formatLocation(p photonProperties) string {
	// Use city as the primary component, falling back to name.
	primary := p.City
	if primary == "" {
		primary = p.Name
	}
	if primary == "" {
		return ""
	}

	parts := []string{primary}

	// Append state if it differs from the primary component.
	if p.State != "" && !strings.EqualFold(p.State, primary) {
		parts = append(parts, p.State)
	}

	// Append country.
	if p.Country != "" {
		parts = append(parts, p.Country)
	}

	return strings.Join(parts, ", ")
}
