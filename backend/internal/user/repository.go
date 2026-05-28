package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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

const userColumns = `id, email, password_hash, name, avatar_url, provider, provider_id, is_admin, username, bio, profile_visibility, created_at, updated_at`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL,
		&u.Provider, &u.ProviderID, &u.IsAdmin,
		&u.Username, &u.Bio, &u.ProfileVisibility,
		&u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	query := `
		INSERT INTO users (id, email, password_hash, name, avatar_url, provider, provider_id, username, bio, profile_visibility)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING is_admin, created_at, updated_at`

	return r.pool.QueryRow(ctx, query,
		u.ID, u.Email, u.PasswordHash, u.Name, u.AvatarURL, u.Provider, u.ProviderID,
		u.Username, u.Bio, u.ProfileVisibility,
	).Scan(&u.IsAdmin, &u.CreatedAt, &u.UpdatedAt)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE id = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("querying user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE email = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("querying user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByProvider(ctx context.Context, provider, providerID string) (User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE provider = $1 AND provider_id = $2`
	u, err := scanUser(r.pool.QueryRow(ctx, query, provider, providerID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("querying user: %w", err)
	}
	return u, nil
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (User, error) {
	query := `SELECT ` + userColumns + ` FROM users WHERE username = $1`
	u, err := scanUser(r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, fmt.Errorf("user not found")
		}
		return User{}, fmt.Errorf("querying user by username: %w", err)
	}
	return u, nil
}

func (r *Repository) Update(ctx context.Context, u *User) error {
	query := `
		UPDATE users SET name = $2, avatar_url = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`

	return r.pool.QueryRow(ctx, query, u.ID, u.Name, u.AvatarURL).Scan(&u.UpdatedAt)
}

func (r *Repository) UpdateProfile(ctx context.Context, userID uuid.UUID, username *string, bio *string, visibility json.RawMessage) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET username = $2, bio = $3, profile_visibility = $4, updated_at = NOW()
		WHERE id = $1`,
		userID, username, bio, visibility,
	)
	if err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}
	return nil
}

func (r *Repository) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`,
		id, hash)
	return err
}

func (r *Repository) List(ctx context.Context, search string, limit, offset int) ([]User, int, error) {
	args := []any{}
	where := ""
	argIdx := 1

	if search != "" {
		where = fmt.Sprintf(` WHERE (name ILIKE $%d OR email ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM users` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`SELECT %s FROM users%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		userColumns, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.AvatarURL,
			&u.Provider, &u.ProviderID, &u.IsAdmin,
			&u.Username, &u.Bio, &u.ProfileVisibility,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *Repository) SetAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET is_admin = $2, updated_at = NOW() WHERE id = $1`,
		id, isAdmin)
	return err
}
