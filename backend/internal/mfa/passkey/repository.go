package passkey

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trackmycareer/app/internal/mfa"
)

// Repository handles persistence for passkeys and WebAuthn challenges.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new passkey Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new passkey.
func (r *Repository) Create(ctx context.Context, p *mfa.Passkey) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mfa_passkeys (id, user_id, credential_id, public_key, attestation_type, transport, sign_count, name, aaguid)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.UserID, p.CredentialID, p.PublicKey, p.AttestationType, p.Transport, p.SignCount, p.Name, p.AAGUID,
	)
	if err != nil {
		return fmt.Errorf("creating passkey: %w", err)
	}
	return nil
}

// ListByUserID retrieves all passkeys for a given user.
func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]mfa.Passkey, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, credential_id, public_key, attestation_type, transport, sign_count, name, aaguid, created_at, last_used_at
		FROM mfa_passkeys
		WHERE user_id = $1
		ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing passkeys: %w", err)
	}
	defer rows.Close()

	var passkeys []mfa.Passkey
	for rows.Next() {
		var p mfa.Passkey
		if err := rows.Scan(
			&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
			&p.AttestationType, &p.Transport, &p.SignCount,
			&p.Name, &p.AAGUID, &p.CreatedAt, &p.LastUsedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning passkey: %w", err)
		}
		passkeys = append(passkeys, p)
	}
	return passkeys, rows.Err()
}

// GetByCredentialID retrieves a passkey by its credential ID.
func (r *Repository) GetByCredentialID(ctx context.Context, credentialID []byte) (mfa.Passkey, error) {
	var p mfa.Passkey
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, credential_id, public_key, attestation_type, transport, sign_count, name, aaguid, created_at, last_used_at
		FROM mfa_passkeys
		WHERE credential_id = $1`,
		credentialID,
	).Scan(
		&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
		&p.AttestationType, &p.Transport, &p.SignCount,
		&p.Name, &p.AAGUID, &p.CreatedAt, &p.LastUsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mfa.Passkey{}, fmt.Errorf("passkey not found")
		}
		return mfa.Passkey{}, fmt.Errorf("querying passkey by credential ID: %w", err)
	}
	return p, nil
}

// GetByID retrieves a passkey by its ID and user ID.
func (r *Repository) GetByID(ctx context.Context, id, userID uuid.UUID) (mfa.Passkey, error) {
	var p mfa.Passkey
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, credential_id, public_key, attestation_type, transport, sign_count, name, aaguid, created_at, last_used_at
		FROM mfa_passkeys
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(
		&p.ID, &p.UserID, &p.CredentialID, &p.PublicKey,
		&p.AttestationType, &p.Transport, &p.SignCount,
		&p.Name, &p.AAGUID, &p.CreatedAt, &p.LastUsedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mfa.Passkey{}, fmt.Errorf("passkey not found")
		}
		return mfa.Passkey{}, fmt.Errorf("querying passkey: %w", err)
	}
	return p, nil
}

// UpdateSignCount updates the sign count and sets last_used_at to now.
func (r *Repository) UpdateSignCount(ctx context.Context, id uuid.UUID, signCount uint32) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE mfa_passkeys SET sign_count = $2, last_used_at = NOW() WHERE id = $1`,
		id, signCount,
	)
	if err != nil {
		return fmt.Errorf("updating passkey sign count: %w", err)
	}
	return nil
}

// Rename updates the display name of a passkey.
func (r *Repository) Rename(ctx context.Context, id, userID uuid.UUID, name string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE mfa_passkeys SET name = $3 WHERE id = $1 AND user_id = $2`,
		id, userID, name,
	)
	if err != nil {
		return fmt.Errorf("renaming passkey: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("passkey not found")
	}
	return nil
}

// Delete removes a passkey.
func (r *Repository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM mfa_passkeys WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting passkey: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("passkey not found")
	}
	return nil
}

// Count returns the number of passkeys for a given user.
func (r *Repository) Count(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM mfa_passkeys WHERE user_id = $1`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting passkeys: %w", err)
	}
	return count, nil
}

// StoreChallenge inserts a WebAuthn challenge for the given user and operation.
func (r *Repository) StoreChallenge(ctx context.Context, c *mfa.WebAuthnChallenge) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO webauthn_challenges (id, user_id, challenge, session_data, operation, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		c.ID, c.UserID, c.Challenge, c.SessionData, c.Operation, c.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("storing WebAuthn challenge: %w", err)
	}
	return nil
}

// GetChallenge retrieves the latest unexpired challenge for the given user and operation.
func (r *Repository) GetChallenge(ctx context.Context, userID uuid.UUID, operation string) (mfa.WebAuthnChallenge, error) {
	var c mfa.WebAuthnChallenge
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, challenge, session_data, operation, expires_at, created_at
		FROM webauthn_challenges
		WHERE user_id = $1 AND operation = $2 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`,
		userID, operation,
	).Scan(&c.ID, &c.UserID, &c.Challenge, &c.SessionData, &c.Operation, &c.ExpiresAt, &c.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mfa.WebAuthnChallenge{}, fmt.Errorf("WebAuthn challenge not found or expired")
		}
		return mfa.WebAuthnChallenge{}, fmt.Errorf("querying WebAuthn challenge: %w", err)
	}
	return c, nil
}

// DeleteChallenge removes a specific WebAuthn challenge.
func (r *Repository) DeleteChallenge(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM webauthn_challenges WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("deleting WebAuthn challenge: %w", err)
	}
	return nil
}

// PurgeExpiredChallenges removes all expired WebAuthn challenges.
func (r *Repository) PurgeExpiredChallenges(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM webauthn_challenges WHERE expires_at <= NOW()`,
	)
	if err != nil {
		return fmt.Errorf("purging expired WebAuthn challenges: %w", err)
	}
	return nil
}
