package customdomain

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/cloudflare"
)

// domainRegex matches a valid hostname: only ASCII letters, digits, hyphens, and dots.
var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?$`)

// hexColourRegex matches a 7-character hex colour string like #4F46E5.
var hexColourRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Service coordinates custom domain operations between the repository and
// the Cloudflare API.
type Service struct {
	repo           *Repository
	cf             *cloudflare.Client // may be nil when Cloudflare is not configured
	fallbackOrigin string
}

// NewService creates a new custom domain service.
func NewService(repo *Repository, cf *cloudflare.Client, fallbackOrigin string) *Service {
	return &Service{
		repo:           repo,
		cf:             cf,
		fallbackOrigin: fallbackOrigin,
	}
}

// Create registers a new custom domain for the given user. It validates the
// domain, calls the Cloudflare API to create a custom hostname, and stores
// the record in the database.
func (s *Service) Create(ctx context.Context, userID uuid.UUID, domain string) (*CustomDomain, error) {
	if s.cf == nil {
		return nil, fmt.Errorf("custom domains are not available in this environment")
	}

	domain = strings.ToLower(strings.TrimSpace(domain))

	if err := validateDomain(domain); err != nil {
		return nil, err
	}

	// Ensure the user does not already have a custom domain.
	existing, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("checking existing domain: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("you already have a custom domain configured")
	}

	// Insert the DB row first to claim the domain via UNIQUE constraint,
	// preventing a race where two requests both pass the availability check
	// and both call Cloudflare.
	d := &CustomDomain{
		ID:           uuid.New(),
		UserID:       userID,
		Domain:       domain,
		Status:       "pending",
		SSLStatus:    "pending",
		AccentColour: "#4F46E5",
	}

	if err := s.repo.Create(ctx, d); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("this domain is already in use")
		}
		return nil, fmt.Errorf("saving custom domain: %w", err)
	}

	// Register with Cloudflare. If this fails, remove the DB row.
	hostnameID, err := s.cf.CreateCustomHostname(ctx, domain)
	if err != nil {
		if delErr := s.repo.Delete(ctx, d.ID); delErr != nil {
			slog.Error("failed to clean up domain after Cloudflare error",
				"error", delErr.Error(), "domain", domain)
		}
		return nil, fmt.Errorf("registering domain with Cloudflare: %w", err)
	}

	if err := s.repo.UpdateCloudflareHostnameID(ctx, d.ID, hostnameID); err != nil {
		slog.Error("failed to store Cloudflare hostname ID",
			"error", err.Error(), "domain", domain)
	}
	d.CloudflareHostnameID = &hostnameID
	d.CNAMETarget = s.fallbackOrigin
	return d, nil
}

// Get returns the custom domain for the given user, populated with the
// CNAME target from configuration.
func (s *Service) Get(ctx context.Context, userID uuid.UUID) (*CustomDomain, error) {
	d, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching custom domain: %w", err)
	}
	if d == nil {
		return nil, nil
	}

	d.CNAMETarget = s.fallbackOrigin
	return d, nil
}

// Verify checks the current status of the custom domain with Cloudflare
// and updates the local record accordingly.
func (s *Service) Verify(ctx context.Context, userID uuid.UUID) (*CustomDomain, error) {
	d, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching custom domain: %w", err)
	}
	if d == nil {
		return nil, fmt.Errorf("no custom domain configured")
	}

	if s.cf == nil || d.CloudflareHostnameID == nil {
		d.CNAMETarget = s.fallbackOrigin
		return d, nil
	}

	status, sslStatus, err := s.cf.GetCustomHostnameStatus(ctx, *d.CloudflareHostnameID)
	if err != nil {
		return nil, fmt.Errorf("checking domain status: %w", err)
	}

	var verifiedAt *time.Time
	if status == "active" && d.VerifiedAt == nil {
		now := time.Now()
		verifiedAt = &now
	} else {
		verifiedAt = d.VerifiedAt
	}

	if err := s.repo.UpdateStatus(ctx, d.ID, status, sslStatus, verifiedAt); err != nil {
		return nil, fmt.Errorf("updating domain status: %w", err)
	}

	d.Status = status
	d.SSLStatus = sslStatus
	d.VerifiedAt = verifiedAt
	d.CNAMETarget = s.fallbackOrigin

	return d, nil
}

// Delete removes the custom domain for the given user. Cloudflare deletion
// is best-effort; the database record is always removed.
func (s *Service) Delete(ctx context.Context, userID uuid.UUID) error {
	d, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching custom domain: %w", err)
	}
	if d == nil {
		return nil
	}

	// Best-effort Cloudflare cleanup.
	if s.cf != nil && d.CloudflareHostnameID != nil {
		if cfErr := s.cf.DeleteCustomHostname(ctx, *d.CloudflareHostnameID); cfErr != nil {
			slog.Error("failed to delete Cloudflare custom hostname",
				"error", cfErr.Error(),
				"hostname_id", *d.CloudflareHostnameID,
				"domain", d.Domain,
			)
		}
	}

	if err := s.repo.Delete(ctx, d.ID); err != nil {
		return fmt.Errorf("deleting custom domain: %w", err)
	}

	return nil
}

// UpdateAccentColour validates and updates the accent colour for the user's
// custom domain.
func (s *Service) UpdateAccentColour(ctx context.Context, userID uuid.UUID, colour string) (*CustomDomain, error) {
	colour = strings.TrimSpace(colour)
	if !hexColourRegex.MatchString(colour) {
		return nil, fmt.Errorf("accent colour must be a valid hex colour (e.g. #4F46E5)")
	}

	d, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("fetching custom domain: %w", err)
	}
	if d == nil {
		return nil, fmt.Errorf("no custom domain configured")
	}

	if err := s.repo.UpdateAccentColour(ctx, d.ID, colour); err != nil {
		return nil, fmt.Errorf("updating accent colour: %w", err)
	}

	d.AccentColour = colour
	d.CNAMETarget = s.fallbackOrigin
	return d, nil
}

// GetByDomain returns the custom domain for a given domain string. This is
// used by the public profile endpoint to look up profiles served on custom domains.
func (s *Service) GetByDomain(ctx context.Context, domain string) (*CustomDomain, error) {
	d, err := s.repo.GetByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("fetching custom domain: %w", err)
	}
	return d, nil
}

// validateDomain checks that a domain string is a valid, allowed hostname.
func validateDomain(domain string) error {
	if len(domain) == 0 {
		return fmt.Errorf("domain is required")
	}
	if len(domain) > 253 {
		return fmt.Errorf("domain must be 253 characters or fewer")
	}
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("domain contains invalid characters")
	}

	labels := strings.Split(domain, ".")
	if len(labels) < 2 {
		return fmt.Errorf("domain must have at least two labels (e.g. example.com)")
	}

	for _, label := range labels {
		if label == "" {
			return fmt.Errorf("domain contains empty labels")
		}
		if len(label) > 63 {
			return fmt.Errorf("each domain label must be 63 characters or fewer")
		}
		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return fmt.Errorf("domain labels must not start or end with a hyphen")
		}
	}

	lower := strings.ToLower(domain)
	if lower == "trackmy.career" || strings.HasSuffix(lower, ".trackmy.career") {
		return fmt.Errorf("domains ending with trackmy.career are not allowed")
	}

	blockedDomains := []string{"localhost", "localhost.localdomain"}
	blockedSuffixes := []string{".local", ".localhost", ".internal", ".pages.dev"}
	for _, blocked := range blockedDomains {
		if lower == blocked {
			return fmt.Errorf("this domain is not allowed")
		}
	}
	for _, suffix := range blockedSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return fmt.Errorf("this domain is not allowed")
		}
	}

	return nil
}
