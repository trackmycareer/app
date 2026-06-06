package customdomain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides persistence for custom domain records.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new custom domain repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const columns = `id, user_id, domain, status, cloudflare_hostname_id, ssl_status, accent_colour, created_at, verified_at`

func scanCustomDomain(row pgx.Row) (*CustomDomain, error) {
	var d CustomDomain
	err := row.Scan(
		&d.ID, &d.UserID, &d.Domain, &d.Status,
		&d.CloudflareHostnameID, &d.SSLStatus, &d.AccentColour,
		&d.CreatedAt, &d.VerifiedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

// GetByUserID returns the custom domain for a given user, or nil if none exists.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (*CustomDomain, error) {
	query := `SELECT ` + columns + ` FROM custom_domains WHERE user_id = $1`
	d, err := scanCustomDomain(r.pool.QueryRow(ctx, query, userID))
	if err != nil {
		return nil, fmt.Errorf("querying custom domain by user: %w", err)
	}
	return d, nil
}

// GetByDomain returns the custom domain record for a given domain string,
// or nil if none exists.
func (r *Repository) GetByDomain(ctx context.Context, domain string) (*CustomDomain, error) {
	query := `SELECT ` + columns + ` FROM custom_domains WHERE domain = $1`
	d, err := scanCustomDomain(r.pool.QueryRow(ctx, query, domain))
	if err != nil {
		return nil, fmt.Errorf("querying custom domain by domain: %w", err)
	}
	return d, nil
}

// Create inserts a new custom domain record.
func (r *Repository) Create(ctx context.Context, d *CustomDomain) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO custom_domains (id, user_id, domain, status, cloudflare_hostname_id, ssl_status, accent_colour)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`,
		d.ID, d.UserID, d.Domain, d.Status, d.CloudflareHostnameID, d.SSLStatus, d.AccentColour,
	).Scan(&d.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting custom domain: %w", err)
	}
	return nil
}

// UpdateStatus updates the verification and SSL status of a custom domain.
func (r *Repository) UpdateStatus(ctx context.Context, id uuid.UUID, status, sslStatus string, verifiedAt *time.Time) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE custom_domains SET status = $2, ssl_status = $3, verified_at = $4 WHERE id = $1`,
		id, status, sslStatus, verifiedAt,
	)
	if err != nil {
		return fmt.Errorf("updating custom domain status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("custom domain not found")
	}
	return nil
}

// UpdateCloudflareHostnameID stores the Cloudflare hostname ID after registration.
func (r *Repository) UpdateCloudflareHostnameID(ctx context.Context, id uuid.UUID, hostnameID string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE custom_domains SET cloudflare_hostname_id = $2 WHERE id = $1`,
		id, hostnameID,
	)
	if err != nil {
		return fmt.Errorf("updating cloudflare hostname ID: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("custom domain not found")
	}
	return nil
}

// UpdateAccentColour updates the accent colour of a custom domain.
func (r *Repository) UpdateAccentColour(ctx context.Context, id uuid.UUID, colour string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE custom_domains SET accent_colour = $2 WHERE id = $1`,
		id, colour,
	)
	if err != nil {
		return fmt.Errorf("updating accent colour: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("custom domain not found")
	}
	return nil
}

// Delete removes a custom domain record by ID.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM custom_domains WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("deleting custom domain: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("custom domain not found")
	}
	return nil
}
