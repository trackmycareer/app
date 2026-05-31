package skill

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

func (r *Repository) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Skill, int, error) {
	args := []any{userID}
	where := `WHERE s.user_id = $1`
	argIdx := 2

	if params.Search != "" {
		where += fmt.Sprintf(` AND (s.name ILIKE $%d OR s.notes ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	if params.Category != "" {
		where += fmt.Sprintf(` AND s.category = $%d`, argIdx)
		args = append(args, params.Category)
		argIdx++
	}

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM skills s ` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting skills: %w", err)
	}

	// Apply pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT s.id, s.user_id, s.name, s.category, s.proficiency, s.notes, s.created_at, s.updated_at
		FROM skills s %s ORDER BY s.name ASC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing skills: %w", err)
	}
	defer rows.Close()

	var skills []Skill
	var skillIDs []uuid.UUID
	for rows.Next() {
		var s Skill
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.Name, &s.Category,
			&s.Proficiency, &s.Notes, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning skill: %w", err)
		}
		s.Evidence = []SkillEvidence{}
		skills = append(skills, s)
		skillIDs = append(skillIDs, s.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating skills: %w", err)
	}

	// Load evidence for all skills in one query
	if len(skillIDs) > 0 {
		if err := r.loadEvidenceForSkills(ctx, skills, skillIDs); err != nil {
			return nil, 0, err
		}
	}

	return skills, total, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, skillID uuid.UUID) (Skill, error) {
	var s Skill
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, category, proficiency, notes, created_at, updated_at
		FROM skills WHERE id = $1 AND user_id = $2`,
		skillID, userID,
	).Scan(&s.ID, &s.UserID, &s.Name, &s.Category, &s.Proficiency, &s.Notes, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Skill{}, fmt.Errorf("skill not found")
		}
		return Skill{}, fmt.Errorf("querying skill: %w", err)
	}

	evidence, err := r.getEvidenceForSkill(ctx, s.ID)
	if err != nil {
		return Skill{}, err
	}
	s.Evidence = evidence

	return s, nil
}

func (r *Repository) Create(ctx context.Context, s *Skill) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO skills (id, user_id, name, category, proficiency, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`,
		s.ID, s.UserID, s.Name, s.Category, s.Proficiency, s.Notes,
	).Scan(&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting skill: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, s *Skill) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE skills SET name = $3, category = $4, proficiency = $5, notes = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		s.ID, s.UserID, s.Name, s.Category, s.Proficiency, s.Notes,
	)
	if err != nil {
		return fmt.Errorf("updating skill: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("skill not found")
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, skillID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM skills WHERE id = $1 AND user_id = $2`,
		skillID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting skill: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("skill not found")
	}
	return nil
}

func (r *Repository) AddEvidence(ctx context.Context, userID, skillID uuid.UUID, evidenceType string, evidenceID uuid.UUID) (SkillEvidence, error) {
	// Validate skill belongs to user
	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM skills WHERE id = $1 AND user_id = $2)`,
		skillID, userID,
	).Scan(&exists); err != nil {
		return SkillEvidence{}, fmt.Errorf("checking skill ownership: %w", err)
	}
	if !exists {
		return SkillEvidence{}, fmt.Errorf("skill not found")
	}

	// Validate evidence entity belongs to user
	var evidenceExists bool
	var evidenceQuery string
	switch evidenceType {
	case "job":
		evidenceQuery = `SELECT EXISTS(SELECT 1 FROM jobs WHERE id = $1 AND user_id = $2)`
	case "certification":
		evidenceQuery = `SELECT EXISTS(SELECT 1 FROM certifications WHERE id = $1 AND user_id = $2)`
	case "win":
		evidenceQuery = `SELECT EXISTS(SELECT 1 FROM wins WHERE id = $1 AND user_id = $2)`
	default:
		return SkillEvidence{}, fmt.Errorf("invalid evidence_type: %s", evidenceType)
	}
	if err := r.pool.QueryRow(ctx, evidenceQuery, evidenceID, userID).Scan(&evidenceExists); err != nil {
		return SkillEvidence{}, fmt.Errorf("checking evidence ownership: %w", err)
	}
	if !evidenceExists {
		return SkillEvidence{}, fmt.Errorf("evidence not found")
	}

	var e SkillEvidence
	e.ID = uuid.New()
	err := r.pool.QueryRow(ctx,
		`INSERT INTO skill_evidence (id, skill_id, evidence_type, evidence_id)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at`,
		e.ID, skillID, evidenceType, evidenceID,
	).Scan(&e.CreatedAt)
	if err != nil {
		return SkillEvidence{}, fmt.Errorf("inserting skill evidence: %w", err)
	}

	e.SkillID = skillID
	e.EvidenceType = evidenceType
	e.EvidenceID = evidenceID
	return e, nil
}

func (r *Repository) RemoveEvidence(ctx context.Context, userID, skillID, evidenceID uuid.UUID) error {
	// Validate skill belongs to user
	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM skills WHERE id = $1 AND user_id = $2)`,
		skillID, userID,
	).Scan(&exists); err != nil {
		return fmt.Errorf("checking skill ownership: %w", err)
	}
	if !exists {
		return fmt.Errorf("skill not found")
	}

	result, err := r.pool.Exec(ctx,
		`DELETE FROM skill_evidence WHERE id = $1 AND skill_id = $2`,
		evidenceID, skillID,
	)
	if err != nil {
		return fmt.Errorf("deleting skill evidence: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("evidence not found")
	}
	return nil
}

// loadEvidenceForSkills loads evidence for a slice of skills in a single query and assigns them in place.
func (r *Repository) loadEvidenceForSkills(ctx context.Context, skills []Skill, skillIDs []uuid.UUID) error {
	rows, err := r.pool.Query(ctx,
		`SELECT id, skill_id, evidence_type, evidence_id, created_at
		FROM skill_evidence
		WHERE skill_id = ANY($1)
		ORDER BY created_at DESC`,
		skillIDs,
	)
	if err != nil {
		return fmt.Errorf("loading skill evidence: %w", err)
	}
	defer rows.Close()

	evidenceMap := make(map[uuid.UUID][]SkillEvidence)
	for rows.Next() {
		var e SkillEvidence
		if err := rows.Scan(&e.ID, &e.SkillID, &e.EvidenceType, &e.EvidenceID, &e.CreatedAt); err != nil {
			return fmt.Errorf("scanning skill evidence: %w", err)
		}
		evidenceMap[e.SkillID] = append(evidenceMap[e.SkillID], e)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterating skill evidence: %w", err)
	}

	for i := range skills {
		if evidence, ok := evidenceMap[skills[i].ID]; ok {
			skills[i].Evidence = evidence
		}
	}

	return nil
}

// getEvidenceForSkill loads evidence for a single skill.
func (r *Repository) getEvidenceForSkill(ctx context.Context, skillID uuid.UUID) ([]SkillEvidence, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, skill_id, evidence_type, evidence_id, created_at
		FROM skill_evidence
		WHERE skill_id = $1
		ORDER BY created_at DESC`,
		skillID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading evidence for skill: %w", err)
	}
	defer rows.Close()

	var evidence []SkillEvidence
	for rows.Next() {
		var e SkillEvidence
		if err := rows.Scan(&e.ID, &e.SkillID, &e.EvidenceType, &e.EvidenceID, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning evidence: %w", err)
		}
		evidence = append(evidence, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if evidence == nil {
		evidence = []SkillEvidence{}
	}
	return evidence, nil
}
