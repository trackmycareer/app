package tag

import (
	"context"
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

func scanTag(row pgx.Row) (Tag, error) {
	var t Tag
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Colour)
	return t, err
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, name, colour FROM tags WHERE user_id = $1 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Colour); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *Repository) Create(ctx context.Context, t *Tag) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO tags (id, user_id, name, colour) VALUES ($1, $2, $3, $4) RETURNING id`,
		t.ID, t.UserID, t.Name, t.Colour,
	).Scan(&t.ID)
}

func (r *Repository) Update(ctx context.Context, t *Tag) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE tags SET name = $3, colour = $4 WHERE id = $1 AND user_id = $2`,
		t.ID, t.UserID, t.Name, t.Colour,
	)
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, tagID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM tags WHERE id = $1 AND user_id = $2`,
		tagID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("tag not found")
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, userID, tagID uuid.UUID) (Tag, error) {
	t, err := scanTag(r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, colour FROM tags WHERE id = $1 AND user_id = $2`,
		tagID, userID,
	))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tag{}, fmt.Errorf("tag not found")
		}
		return Tag{}, fmt.Errorf("querying tag: %w", err)
	}
	return t, nil
}
