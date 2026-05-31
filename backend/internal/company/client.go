package company

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type companiesHouseResponse struct {
	Items []companiesHouseItem `json:"items"`
}

type companiesHouseItem struct {
	Title         string `json:"title"`
	CompanyNumber string `json:"company_number"`
	CompanyStatus string `json:"company_status"`
}

// CompaniesHouseClient calls the Companies House public search API.
// If the API key is empty, all searches return an empty slice, allowing
// self-hosted deployments to run without a Companies House account.
const defaultBaseURL = "https://api.company-information.service.gov.uk"

type CompaniesHouseClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewCompaniesHouseClient creates a client for the Companies House search API.
func NewCompaniesHouseClient(apiKey, baseURL string) *CompaniesHouseClient {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &CompaniesHouseClient{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// Search queries the Companies House API for companies matching the given query.
// Returns an empty slice (not an error) when the API key is not configured.
func (c *CompaniesHouseClient) Search(ctx context.Context, query string) ([]CompanyResult, error) {
	if c.apiKey == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf(
		"%s/search/companies?q=%s&items_per_page=10",
		c.baseURL, url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Companies House uses HTTP Basic auth with the API key as the username
	// and an empty password.
	req.SetBasicAuth(c.apiKey, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from Companies House API", resp.StatusCode)
	}

	var chResp companiesHouseResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&chResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	results := make([]CompanyResult, 0, len(chResp.Items))
	for _, item := range chResp.Items {
		results = append(results, CompanyResult{
			Name:          item.Title,
			Source:        "companies_house",
			CompanyNumber: item.CompanyNumber,
			Status:        item.CompanyStatus,
		})
	}

	return results, nil
}
