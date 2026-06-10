package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

var cfHexIDRegex = regexp.MustCompile(`^[0-9a-f]{32}$`)
var cfUUIDRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

const baseURL = "https://api.cloudflare.com/client/v4"

// Client wraps the Cloudflare API for custom hostname (Cloudflare for SaaS)
// operations. All methods target a single zone.
type Client struct {
	apiToken       string
	zoneID         string
	fallbackOrigin string
	httpClient     *http.Client
}

// NewClient creates a Cloudflare API client for custom hostname management.
// It validates that zoneID matches the expected Cloudflare UUID format.
func NewClient(apiToken, zoneID, fallbackOrigin string) *Client {
	if !cfHexIDRegex.MatchString(zoneID) {
		panic(fmt.Sprintf("invalid Cloudflare zone ID format: %q", zoneID))
	}
	return &Client{
		apiToken:       apiToken,
		zoneID:         zoneID,
		fallbackOrigin: fallbackOrigin,
		httpClient:     &http.Client{Timeout: 15 * time.Second},
	}
}

// cfResponse is the common Cloudflare JSON envelope.
type cfResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Errors  []cfError       `json:"errors"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// createHostnameResult holds the fields we need from the create response.
type createHostnameResult struct {
	ID string `json:"id"`
}

// hostnameStatusResult holds status fields from the get response.
type hostnameStatusResult struct {
	Status string `json:"status"`
	SSL    struct {
		Status string `json:"status"`
	} `json:"ssl"`
}

// createHostnameRequest is the POST body for creating a custom hostname.
type createHostnameRequest struct {
	Hostname string               `json:"hostname"`
	SSL      createHostnameSSLOpt `json:"ssl"`
}

type createHostnameSSLOpt struct {
	Method string `json:"method"`
	Type   string `json:"type"`
}

// CreateCustomHostname registers a new custom hostname in Cloudflare for SaaS.
// It returns the Cloudflare-assigned hostname ID.
func (c *Client) CreateCustomHostname(ctx context.Context, domain string) (string, error) {
	body := createHostnameRequest{
		Hostname: domain,
		SSL:      createHostnameSSLOpt{Method: "http", Type: "dv"},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshalling request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/zones/%s/custom_hostnames", baseURL, c.zoneID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	var cfResp cfResponse
	if err := c.decodeResponse(resp.Body, &cfResp); err != nil {
		return "", err
	}

	if !cfResp.Success {
		return "", fmt.Errorf("cloudflare API error: %s", formatCFErrors(cfResp.Errors))
	}

	var result createHostnameResult
	if err := json.Unmarshal(cfResp.Result, &result); err != nil {
		return "", fmt.Errorf("parsing result: %w", err)
	}

	return result.ID, nil
}

// GetCustomHostnameStatus retrieves the current verification and SSL status
// of a custom hostname.
func (c *Client) GetCustomHostnameStatus(ctx context.Context, hostnameID string) (status, sslStatus string, err error) {
	if !cfUUIDRegex.MatchString(hostnameID) {
		return "", "", fmt.Errorf("invalid hostname ID format")
	}
	endpoint := fmt.Sprintf("%s/zones/%s/custom_hostnames/%s", baseURL, c.zoneID, hostnameID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	var cfResp cfResponse
	if err := c.decodeResponse(resp.Body, &cfResp); err != nil {
		return "", "", err
	}

	if !cfResp.Success {
		return "", "", fmt.Errorf("cloudflare API error: %s", formatCFErrors(cfResp.Errors))
	}

	var result hostnameStatusResult
	if err := json.Unmarshal(cfResp.Result, &result); err != nil {
		return "", "", fmt.Errorf("parsing result: %w", err)
	}

	return result.Status, result.SSL.Status, nil
}

// DeleteCustomHostname removes a custom hostname from Cloudflare.
func (c *Client) DeleteCustomHostname(ctx context.Context, hostnameID string) error {
	if !cfUUIDRegex.MatchString(hostnameID) {
		return fmt.Errorf("invalid hostname ID format")
	}
	endpoint := fmt.Sprintf("%s/zones/%s/custom_hostnames/%s", baseURL, c.zoneID, hostnameID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	var cfResp cfResponse
	if err := c.decodeResponse(resp.Body, &cfResp); err != nil {
		return err
	}

	if !cfResp.Success {
		return fmt.Errorf("cloudflare API error: %s", formatCFErrors(cfResp.Errors))
	}

	return nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) decodeResponse(body io.Reader, target *cfResponse) error {
	if err := json.NewDecoder(io.LimitReader(body, 1<<20)).Decode(target); err != nil {
		return fmt.Errorf("decoding cloudflare response: %w", err)
	}
	return nil
}

func formatCFErrors(errs []cfError) string {
	if len(errs) == 0 {
		return "unknown error"
	}
	msg := errs[0].Message
	for _, e := range errs[1:] {
		msg += "; " + e.Message
	}
	return msg
}
