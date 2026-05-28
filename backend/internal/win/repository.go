package win

import (
	"context"
	"errors"
	"fmt"

	"github.com/bhcloudlabs/trackmy-career/internal/tag"
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

func (r *Repository) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Win, int, error) {
	args := []any{userID}
	where := `WHERE w.user_id = $1`
	argIdx := 2

	if params.Search != "" {
		where += fmt.Sprintf(` AND (w.title ILIKE $%d OR w.description ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	if params.Category != "" {
		where += fmt.Sprintf(` AND w.category = $%d`, argIdx)
		args = append(args, params.Category)
		argIdx++
	}

	if params.TagID != "" {
		tagUUID, err := uuid.Parse(params.TagID)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid tag_id: %w", err)
		}
		where += fmt.Sprintf(` AND EXISTS (SELECT 1 FROM win_tags wt WHERE wt.win_id = w.id AND wt.tag_id = $%d)`, argIdx)
		args = append(args, tagUUID)
		argIdx++
	}

	if params.From != "" {
		where += fmt.Sprintf(` AND w.occurred_on >= $%d`, argIdx)
		args = append(args, params.From)
		argIdx++
	}

	if params.To != "" {
		where += fmt.Sprintf(` AND w.occurred_on <= $%d`, argIdx)
		args = append(args, params.To)
		argIdx++
	}

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM wins w ` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting wins: %w", err)
	}

	// Apply pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT w.id, w.user_id, w.title, w.description, w.occurred_on, w.category, w.created_at, w.updated_at
		FROM wins w %s ORDER BY w.occurred_on DESC, w.created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing wins: %w", err)
	}
	defer rows.Close()

	var wins []Win
	var winIDs []uuid.UUID
	for rows.Next() {
		var w Win
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.Title, &w.Description,
			&w.OccurredOn, &w.Category, &w.CreatedAt, &w.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning win: %w", err)
		}
		w.Tags = []tag.Tag{}
		wins = append(wins, w)
		winIDs = append(winIDs, w.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating wins: %w", err)
	}

	// Load tags for all wins in one query
	if len(winIDs) > 0 {
		if err := r.loadTagsForWins(ctx, wins, winIDs); err != nil {
			return nil, 0, err
		}
	}

	return wins, total, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, winID uuid.UUID) (Win, error) {
	var w Win
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, title, description, occurred_on, category, created_at, updated_at
		FROM wins WHERE id = $1 AND user_id = $2`,
		winID, userID,
	).Scan(&w.ID, &w.UserID, &w.Title, &w.Description, &w.OccurredOn, &w.Category, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Win{}, fmt.Errorf("win not found")
		}
		return Win{}, fmt.Errorf("querying win: %w", err)
	}

	w.Tags = []tag.Tag{}
	if err := r.loadTagsForWins(ctx, []Win{w}, []uuid.UUID{w.ID}); err != nil {
		return Win{}, err
	}
	// loadTagsForWins operates on a copy; we need to use it directly
	tags, err := r.getTagsForWin(ctx, winID)
	if err != nil {
		return Win{}, err
	}
	w.Tags = tags

	return w, nil
}

func (r *Repository) Create(ctx context.Context, w *Win, tagIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`INSERT INTO wins (id, user_id, title, description, occurred_on, category)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		w.ID, w.UserID, w.Title, w.Description, w.OccurredOn, w.Category,
	).Scan(&w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting win: %w", err)
	}

	if err := r.replaceWinTags(ctx, tx, w.ID, tagIDs); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) Update(ctx context.Context, w *Win, tagIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	result, err := tx.Exec(ctx,
		`UPDATE wins SET title = $3, description = $4, occurred_on = $5, category = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		w.ID, w.UserID, w.Title, w.Description, w.OccurredOn, w.Category,
	)
	if err != nil {
		return fmt.Errorf("updating win: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("win not found")
	}

	if err := r.replaceWinTags(ctx, tx, w.ID, tagIDs); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) Delete(ctx context.Context, userID, winID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM wins WHERE id = $1 AND user_id = $2`,
		winID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting win: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("win not found")
	}
	return nil
}

// replaceWinTags deletes existing win_tags for the win, then inserts the new set.
func (r *Repository) replaceWinTags(ctx context.Context, tx pgx.Tx, winID uuid.UUID, tagIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM win_tags WHERE win_id = $1`, winID); err != nil {
		return fmt.Errorf("clearing win tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO win_tags (win_id, tag_id) VALUES ($1, $2)`,
			winID, tagID,
		); err != nil {
			return fmt.Errorf("inserting win tag: %w", err)
		}
	}

	return nil
}

// loadTagsForWins loads tags for a slice of wins in a single query and assigns them in place.
func (r *Repository) loadTagsForWins(ctx context.Context, wins []Win, winIDs []uuid.UUID) error {
	rows, err := r.pool.Query(ctx,
		`SELECT wt.win_id, t.id, t.user_id, t.name, t.colour
		FROM win_tags wt
		JOIN tags t ON t.id = wt.tag_id
		WHERE wt.win_id = ANY($1)
		ORDER BY t.name`,
		winIDs,
	)
	if err != nil {
		return fmt.Errorf("loading win tags: %w", err)
	}
	defer rows.Close()

	tagMap := make(map[uuid.UUID][]tag.Tag)
	for rows.Next() {
		var winID uuid.UUID
		var t tag.Tag
		if err := rows.Scan(&winID, &t.ID, &t.UserID, &t.Name, &t.Colour); err != nil {
			return fmt.Errorf("scanning win tag: %w", err)
		}
		tagMap[winID] = append(tagMap[winID], t)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating win tags: %w", err)
	}

	for i := range wins {
		if tags, ok := tagMap[wins[i].ID]; ok {
			wins[i].Tags = tags
		}
	}

	return nil
}

// getTagsForWin loads tags for a single win.
func (r *Repository) getTagsForWin(ctx context.Context, winID uuid.UUID) ([]tag.Tag, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT t.id, t.user_id, t.name, t.colour
		FROM win_tags wt
		JOIN tags t ON t.id = wt.tag_id
		WHERE wt.win_id = $1
		ORDER BY t.name`,
		winID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading tags for win: %w", err)
	}
	defer rows.Close()

	var tags []tag.Tag
	for rows.Next() {
		var t tag.Tag
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Colour); err != nil {
			return nil, fmt.Errorf("scanning tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if tags == nil {
		tags = []tag.Tag{}
	}
	return tags, nil
}
