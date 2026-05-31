package linkedaccount

import (
	"context"
	"log/slog"
	"time"
)

func StartReverification(ctx context.Context, repo *Repository) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				recheckWebsites(ctx, repo)
			}
		}
	}()
}

func recheckWebsites(ctx context.Context, repo *Repository) {
	accounts, err := repo.ListVerifiedWebsitesDueForRecheck(ctx)
	if err != nil {
		slog.Error("failed to list websites for recheck", "error", err)
		return
	}

	if len(accounts) == 0 {
		return
	}

	slog.Info("starting DNS re-verification", "count", len(accounts))

	for _, la := range accounts {
		domain, err := ExtractDomain(la.ProfileURL)
		if err != nil {
			slog.Error("invalid URL during recheck", "id", la.ID, "url", la.ProfileURL, "error", err)
			continue
		}

		if la.VerifyToken == nil {
			slog.Warn("no verify token for website", "id", la.ID)
			continue
		}

		found, err := VerifyDNSTXT(domain, *la.VerifyToken)
		if err != nil {
			slog.Error("DNS lookup failed during recheck", "id", la.ID, "domain", domain, "error", err)
			if updateErr := repo.UpdateVerificationStatus(ctx, la.ID, la.Verified, la.VerifiedAt); updateErr != nil {
				slog.Error("failed to update last_checked", "id", la.ID, "error", updateErr)
			}
			continue
		}

		if found {
			if updateErr := repo.UpdateVerificationStatus(ctx, la.ID, true, la.VerifiedAt); updateErr != nil {
				slog.Error("failed to update last_checked", "id", la.ID, "error", updateErr)
			}
		} else {
			slog.Warn("TXT record missing, revoking verification", "id", la.ID, "domain", domain)
			if updateErr := repo.UpdateVerificationStatus(ctx, la.ID, false, nil); updateErr != nil {
				slog.Error("failed to revoke verification", "id", la.ID, "error", updateErr)
			}
		}
	}

	slog.Info("DNS re-verification complete", "checked", len(accounts))
}
