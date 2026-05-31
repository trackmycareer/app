package linkedaccount

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const linkedAccountColumns = `id, user_id, provider, provider_id, profile_url, verified, verify_token, verified_at, last_checked, created_at`

func scanLinkedAccount(row pgx.Row) (LinkedAccount, error) {
	var la LinkedAccount
	err := row.Scan(
		&la.ID, &la.UserID, &la.Provider, &la.ProviderID,
		&la.ProfileURL, &la.Verified, &la.VerifyToken,
		&la.VerifiedAt, &la.LastChecked, &la.CreatedAt,
	)
	return la, err
}

func scanLinkedAccountRows(rows pgx.Rows) ([]LinkedAccount, error) {
	defer rows.Close()

	var accounts []LinkedAccount
	for rows.Next() {
		var la LinkedAccount
		if err := rows.Scan(
			&la.ID, &la.UserID, &la.Provider, &la.ProviderID,
			&la.ProfileURL, &la.Verified, &la.VerifyToken,
			&la.VerifiedAt, &la.LastChecked, &la.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning linked account: %w", err)
		}
		accounts = append(accounts, la)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating linked accounts: %w", err)
	}
	return accounts, nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]LinkedAccount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+linkedAccountColumns+` FROM linked_accounts WHERE user_id = $1 ORDER BY provider`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing linked accounts: %w", err)
	}
	return scanLinkedAccountRows(rows)
}

func (r *Repository) GetByUserAndProvider(ctx context.Context, userID uuid.UUID, provider string) (LinkedAccount, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+linkedAccountColumns+` FROM linked_accounts WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)
	la, err := scanLinkedAccount(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return LinkedAccount{}, fmt.Errorf("linked account not found")
	}
	if err != nil {
		return LinkedAccount{}, fmt.Errorf("querying linked account: %w", err)
	}
	return la, nil
}

func (r *Repository) GetByProviderID(ctx context.Context, provider, providerID string) (LinkedAccount, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+linkedAccountColumns+` FROM linked_accounts WHERE provider = $1 AND provider_id = $2`,
		provider, providerID,
	)
	la, err := scanLinkedAccount(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return LinkedAccount{}, fmt.Errorf("linked account not found")
	}
	if err != nil {
		return LinkedAccount{}, fmt.Errorf("querying linked account by provider ID: %w", err)
	}
	return la, nil
}

func (r *Repository) Upsert(ctx context.Context, la *LinkedAccount) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO linked_accounts (id, user_id, provider, provider_id, profile_url, verified, verify_token, verified_at, last_checked, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (user_id, provider) DO UPDATE SET
		   provider_id  = EXCLUDED.provider_id,
		   profile_url  = EXCLUDED.profile_url,
		   verified     = EXCLUDED.verified,
		   verify_token = EXCLUDED.verify_token,
		   verified_at  = EXCLUDED.verified_at,
		   last_checked = EXCLUDED.last_checked`,
		la.ID, la.UserID, la.Provider, la.ProviderID,
		la.ProfileURL, la.Verified, la.VerifyToken,
		la.VerifiedAt, la.LastChecked, la.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upserting linked account: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID uuid.UUID, provider string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM linked_accounts WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	)
	if err != nil {
		return fmt.Errorf("deleting linked account: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("linked account not found")
	}
	return nil
}

func (r *Repository) ListVerifiedWebsitesDueForRecheck(ctx context.Context) ([]LinkedAccount, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+linkedAccountColumns+` FROM linked_accounts
		 WHERE provider = 'website' AND verified = TRUE
		   AND (last_checked IS NULL OR last_checked < NOW() - INTERVAL '30 days')`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing websites due for recheck: %w", err)
	}
	return scanLinkedAccountRows(rows)
}

func (r *Repository) UpdateVerificationStatus(ctx context.Context, id uuid.UUID, verified bool, verifiedAt *time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE linked_accounts SET verified = $1, verified_at = $2, last_checked = NOW() WHERE id = $3`,
		verified, verifiedAt, id,
	)
	if err != nil {
		return fmt.Errorf("updating verification status: %w", err)
	}
	return nil
}
