package admin

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/internal/linkedaccount"
	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/pkg/types"
)

// Stats holds aggregate counts across all users.
type Stats struct {
	TotalUsers          int `json:"total_users"`
	TotalWins           int `json:"total_wins"`
	TotalJobs           int `json:"total_jobs"`
	TotalCertifications int `json:"total_certifications"`
	TotalSkills         int `json:"total_skills"`
}

// UserDetail is the enriched view of a single user returned to admins. It groups
// the user's account record with their activity counts, gamification summary,
// linked accounts and a handful of their most recent items.
type UserDetail struct {
	User           user.User                           `json:"user"`
	Counts         ActivityCounts                      `json:"counts"`
	Gamification   GamificationSummary                 `json:"gamification"`
	LinkedAccounts []linkedaccount.PublicLinkedAccount `json:"linked_accounts"`
	Recent         RecentItems                         `json:"recent"`
}

// ActivityCounts holds per-user totals across the tracked domains.
type ActivityCounts struct {
	Wins                 int `json:"wins"`
	Jobs                 int `json:"jobs"`
	Certifications       int `json:"certifications"`
	CertificationsPassed int `json:"certifications_passed"`
	Skills               int `json:"skills"`
	Evidence             int `json:"evidence"`
	Badges               int `json:"badges"`
}

// GamificationSummary holds the user's points, level and streak figures.
// LastActiveOn is a proxy for engagement drawn from the gamification streak; it
// reflects the last day the user recorded a tracked activity (win, job, cert,
// skill), not a login, and is nil if they have never recorded one.
type GamificationSummary struct {
	TotalPoints   int         `json:"total_points"`
	Level         int         `json:"level"`
	LevelTitle    string      `json:"level_title"`
	CurrentStreak int         `json:"current_streak"`
	LongestStreak int         `json:"longest_streak"`
	LastActiveOn  *types.Date `json:"last_active_on,omitempty"`
}

// RecentItems holds the most recent few entries per domain for a quick glance.
type RecentItems struct {
	Wins           []RecentWin   `json:"wins"`
	Jobs           []RecentJob   `json:"jobs"`
	Certifications []RecentCert  `json:"certifications"`
	Skills         []RecentSkill `json:"skills"`
}

// RecentWin is a lightweight preview of a win.
type RecentWin struct {
	ID         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Category   string     `json:"category"`
	OccurredOn types.Date `json:"occurred_on"`
}

// RecentJob is a lightweight preview of a job.
type RecentJob struct {
	ID        uuid.UUID   `json:"id"`
	Company   string      `json:"company"`
	Title     string      `json:"title"`
	StartDate types.Date  `json:"start_date"`
	EndDate   *types.Date `json:"end_date,omitempty"`
}

// RecentCert is a lightweight preview of a certification.
type RecentCert struct {
	ID         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	Provider   string      `json:"provider"`
	Status     string      `json:"status"`
	EarnedDate *types.Date `json:"earned_date,omitempty"`
}

// RecentSkill is a lightweight preview of a skill.
type RecentSkill struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Category    *string   `json:"category,omitempty"`
	Proficiency int       `json:"proficiency"`
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

// GetUserDetail aggregates a single user's activity counts, gamification figures,
// linked accounts and recent items. All queries are read-only: unlike the
// gamification repository's GetOrCreate helpers, viewing a user here never
// creates rows for them.
func (s *Service) GetUserDetail(ctx context.Context, u user.User) (UserDetail, error) {
	detail := UserDetail{
		User:           u,
		LinkedAccounts: []linkedaccount.PublicLinkedAccount{},
		Recent: RecentItems{
			Wins:           []RecentWin{},
			Jobs:           []RecentJob{},
			Certifications: []RecentCert{},
			Skills:         []RecentSkill{},
		},
	}

	if err := s.scanSummary(ctx, u.ID, &detail); err != nil {
		return UserDetail{}, err
	}
	if err := s.scanRecentWins(ctx, u.ID, &detail); err != nil {
		return UserDetail{}, err
	}
	if err := s.scanRecentJobs(ctx, u.ID, &detail); err != nil {
		return UserDetail{}, err
	}
	if err := s.scanRecentCerts(ctx, u.ID, &detail); err != nil {
		return UserDetail{}, err
	}
	if err := s.scanRecentSkills(ctx, u.ID, &detail); err != nil {
		return UserDetail{}, err
	}
	if err := s.scanLinkedAccounts(ctx, u.ID, &detail); err != nil {
		return UserDetail{}, err
	}

	return detail, nil
}

func (s *Service) scanSummary(ctx context.Context, userID uuid.UUID, detail *UserDetail) error {
	row := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM wins WHERE user_id = $1),
			(SELECT COUNT(*) FROM jobs WHERE user_id = $1),
			(SELECT COUNT(*) FROM certifications WHERE user_id = $1),
			(SELECT COUNT(*) FROM certifications WHERE user_id = $1 AND status = 'passed'),
			(SELECT COUNT(*) FROM skills WHERE user_id = $1),
			(SELECT COUNT(*) FROM skill_evidence se JOIN skills s ON s.id = se.skill_id WHERE s.user_id = $1),
			(SELECT COUNT(*) FROM user_badges WHERE user_id = $1),
			COALESCE(p.total_points, 0),
			COALESCE(st.current_streak, 0),
			COALESCE(st.longest_streak, 0),
			st.last_active_on
		FROM (SELECT 1) AS d
		LEFT JOIN user_points p ON p.user_id = $1
		LEFT JOIN user_streaks st ON st.user_id = $1
	`, userID)

	var lastActive *types.Date
	if err := row.Scan(
		&detail.Counts.Wins,
		&detail.Counts.Jobs,
		&detail.Counts.Certifications,
		&detail.Counts.CertificationsPassed,
		&detail.Counts.Skills,
		&detail.Counts.Evidence,
		&detail.Counts.Badges,
		&detail.Gamification.TotalPoints,
		&detail.Gamification.CurrentStreak,
		&detail.Gamification.LongestStreak,
		&lastActive,
	); err != nil {
		return fmt.Errorf("querying user detail summary: %w", err)
	}

	level, title := gamification.LevelForPoints(detail.Gamification.TotalPoints)
	detail.Gamification.Level = level
	detail.Gamification.LevelTitle = title
	detail.Gamification.LastActiveOn = lastActive

	return nil
}

func (s *Service) scanRecentWins(ctx context.Context, userID uuid.UUID, detail *UserDetail) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, title, category, occurred_on
		FROM wins WHERE user_id = $1
		ORDER BY occurred_on DESC, created_at DESC LIMIT 5
	`, userID)
	if err != nil {
		return fmt.Errorf("querying recent wins: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var w RecentWin
		if err := rows.Scan(&w.ID, &w.Title, &w.Category, &w.OccurredOn); err != nil {
			return fmt.Errorf("scanning recent win: %w", err)
		}
		detail.Recent.Wins = append(detail.Recent.Wins, w)
	}
	return rows.Err()
}

func (s *Service) scanRecentJobs(ctx context.Context, userID uuid.UUID, detail *UserDetail) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, company, title, start_date, end_date
		FROM jobs WHERE user_id = $1
		ORDER BY start_date DESC, created_at DESC LIMIT 5
	`, userID)
	if err != nil {
		return fmt.Errorf("querying recent jobs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var j RecentJob
		if err := rows.Scan(&j.ID, &j.Company, &j.Title, &j.StartDate, &j.EndDate); err != nil {
			return fmt.Errorf("scanning recent job: %w", err)
		}
		detail.Recent.Jobs = append(detail.Recent.Jobs, j)
	}
	return rows.Err()
}

func (s *Service) scanRecentCerts(ctx context.Context, userID uuid.UUID, detail *UserDetail) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, provider, status, earned_date
		FROM certifications WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 5
	`, userID)
	if err != nil {
		return fmt.Errorf("querying recent certifications: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cert RecentCert
		if err := rows.Scan(&cert.ID, &cert.Name, &cert.Provider, &cert.Status, &cert.EarnedDate); err != nil {
			return fmt.Errorf("scanning recent certification: %w", err)
		}
		detail.Recent.Certifications = append(detail.Recent.Certifications, cert)
	}
	return rows.Err()
}

func (s *Service) scanRecentSkills(ctx context.Context, userID uuid.UUID, detail *UserDetail) error {
	rows, err := s.pool.Query(ctx, `
		SELECT id, name, category, proficiency
		FROM skills WHERE user_id = $1
		ORDER BY created_at DESC LIMIT 5
	`, userID)
	if err != nil {
		return fmt.Errorf("querying recent skills: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sk RecentSkill
		if err := rows.Scan(&sk.ID, &sk.Name, &sk.Category, &sk.Proficiency); err != nil {
			return fmt.Errorf("scanning recent skill: %w", err)
		}
		detail.Recent.Skills = append(detail.Recent.Skills, sk)
	}
	return rows.Err()
}

func (s *Service) scanLinkedAccounts(ctx context.Context, userID uuid.UUID, detail *UserDetail) error {
	rows, err := s.pool.Query(ctx, `
		SELECT provider, profile_url, verified, verified_at
		FROM linked_accounts WHERE user_id = $1
		ORDER BY provider
	`, userID)
	if err != nil {
		return fmt.Errorf("querying linked accounts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var la linkedaccount.PublicLinkedAccount
		if err := rows.Scan(&la.Provider, &la.ProfileURL, &la.Verified, &la.VerifiedAt); err != nil {
			return fmt.Errorf("scanning linked account: %w", err)
		}
		detail.LinkedAccounts = append(detail.LinkedAccounts, la)
	}
	return rows.Err()
}
