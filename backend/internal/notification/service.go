package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mailer"
	"github.com/trackmycareer/app/pkg/types"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type Service struct {
	repo   *Repository
	mailer mailer.Mailer
}

func NewService(repo *Repository, m mailer.Mailer) *Service {
	return &Service{repo: repo, mailer: m}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Notification, int, error) {
	if params.Limit <= 0 {
		params.Limit = defaultListLimit
	}
	if params.Limit > maxListLimit {
		params.Limit = maxListLimit
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	return s.repo.ListByUser(ctx, userID, params)
}

func (s *Service) UnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.UnreadCount(ctx, userID)
}

func (s *Service) MarkRead(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.MarkRead(ctx, userID, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *Service) GetPreferences(ctx context.Context, userID uuid.UUID) (Preferences, error) {
	return s.repo.GetPreferences(ctx, userID)
}

func (s *Service) UpdatePreferences(ctx context.Context, userID uuid.UUID, p Preferences) error {
	return s.repo.UpsertPreferences(ctx, userID, p)
}

// RunDailyReminders is the core scheduled job. It finds certifications crossing
// a reminder band, writes one notification per crossing (deduped at the
// database), and sends the reminder email when the email channel is enabled.
//
// On the very first run it performs a self-suppressing pass: existing backlog
// is seeded into the dedup index as already-handled (read and email-sent) so a
// fresh deployment does not produce an email flood. Genuine future crossings
// then fire normally.
func (s *Service) RunDailyReminders(ctx context.Context) {
	initialised, err := s.repo.IsReminderEngineInitialised(ctx)
	if err != nil {
		slog.Error("reminder engine: reading init flag failed", "error", err)
		return
	}
	initialise := !initialised

	due, err := s.repo.ListDueReminders(ctx)
	if err != nil {
		slog.Error("reminder engine: listing due reminders failed", "error", err)
		return
	}

	now := time.Now().UTC()
	var created, emailed, failed int

	for _, d := range due {
		d.ApplicableType = applicableType(d.DaysRemaining, d.Remind90, d.Remind30, d.Remind7)
		if d.ApplicableType == "" {
			continue
		}

		certID := d.CertID
		forDate := d.ExpiryDate
		title, body := composeReminder(d)

		n := &Notification{
			UserID:                 d.UserID,
			Type:                   d.ApplicableType,
			Title:                  title,
			Body:                   body,
			RelatedCertificationID: &certID,
			ReminderForDate:        &forDate,
		}

		switch {
		case initialise:
			// Seed as already handled: present for dedup, but silent.
			n.ReadAt = &now
			n.EmailSentAt = &now
		case !d.ChannelInApp:
			// In-app channel off: keep the dedup anchor but do not surface it
			// as an unread notification.
			n.ReadAt = &now
		}

		wasCreated, err := s.repo.Create(ctx, n)
		if err != nil {
			slog.Error("reminder engine: creating notification failed",
				"cert_id", certID, "type", d.ApplicableType, "error", err)
			failed++
			continue
		}
		if !wasCreated {
			continue
		}
		created++

		if initialise || !d.ChannelEmail {
			continue
		}

		if err := s.mailer.SendCertificationReminder(d.Email, d.Name, toCertReminderData(d)); err != nil {
			slog.Error("reminder engine: sending email failed",
				"cert_id", certID, "to", d.Email, "error", err)
			failed++
			continue
		}
		if err := s.repo.MarkEmailSent(ctx, n.ID); err != nil {
			slog.Error("reminder engine: marking email sent failed",
				"notification_id", n.ID, "error", err)
		}
		emailed++
	}

	if initialise {
		if err := s.repo.SetReminderEngineInitialised(ctx); err != nil {
			slog.Error("reminder engine: setting init flag failed", "error", err)
		}
		slog.Info("reminder engine initialised", "seeded", created)
		return
	}

	slog.Info("reminder engine run complete", "created", created, "emailed", emailed, "failed", failed)
}

// applicableType returns the single most-urgent enabled reminder band for a
// certification that is `days` from expiry (negative means already expired), or
// "" when no enabled band applies. It checks expired first, then 7, 30, 90, so
// a certification only ever fires its most-urgent band; a disabled band falls
// through to the next wider enabled one.
func applicableType(days int, remind90, remind30, remind7 bool) string {
	switch {
	case days < 0:
		return TypeExpired
	case days <= 7 && remind7:
		return TypeExpiry7
	case days <= 30 && remind30:
		return TypeExpiry30
	case days <= 90 && remind90:
		return TypeExpiry90
	default:
		return ""
	}
}

// composeReminder builds the in-app title and body for a due reminder.
func composeReminder(d DueReminder) (string, string) {
	expiry := formatDateBritish(d.ExpiryDate)
	from := ""
	if d.Provider != "" {
		from = " from " + d.Provider
	}

	if d.ApplicableType == TypeExpired {
		title := fmt.Sprintf("%s has expired", d.CertName)
		body := fmt.Sprintf("Your %s certification%s expired on %s. Renew it to keep your credentials current.",
			d.CertName, from, expiry)
		return title, body
	}

	days := d.DaysRemaining
	unit := "days"
	if days == 1 {
		unit = "day"
	}
	title := fmt.Sprintf("%s expires in %d %s", d.CertName, days, unit)
	body := fmt.Sprintf("Your %s certification%s expires on %s, %d %s from now. Plan your renewal so it does not lapse.",
		d.CertName, from, expiry, days, unit)
	return title, body
}

func toCertReminderData(d DueReminder) mailer.CertReminderData {
	return mailer.CertReminderData{
		CertID:        d.CertID.String(),
		CertName:      d.CertName,
		Provider:      d.Provider,
		ExpiryDate:    formatDateBritish(d.ExpiryDate),
		DaysRemaining: d.DaysRemaining,
		Expired:       d.ApplicableType == TypeExpired,
		CredentialURL: d.CredentialURL,
	}
}

func formatDateBritish(d types.Date) string {
	return d.Format("2 January 2006")
}
