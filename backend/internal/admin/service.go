package admin

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Stats holds aggregate counts across all users.
type Stats struct {
	TotalUsers          int `json:"total_users"`
	TotalWins           int `json:"total_wins"`
	TotalJobs           int `json:"total_jobs"`
	TotalCertifications int `json:"total_certifications"`
	TotalSkills         int `json:"total_skills"`
}

// Service provides admin-level operations.
type Service struct {
	pool *pgxpool.Pool
}

// NewService creates a new admin service.
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// GetStats returns aggregate counts for the entire instance.
func (s *Service) GetStats(ctx context.Context) (Stats, error) {
	var stats Stats

	row := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM users),
			(SELECT COUNT(*) FROM wins),
			(SELECT COUNT(*) FROM jobs),
			(SELECT COUNT(*) FROM certifications),
			(SELECT COUNT(*) FROM skills)
	`)

	if err := row.Scan(
		&stats.TotalUsers,
		&stats.TotalWins,
		&stats.TotalJobs,
		&stats.TotalCertifications,
		&stats.TotalSkills,
	); err != nil {
		return Stats{}, fmt.Errorf("querying admin stats: %w", err)
	}

	return stats, nil
}
