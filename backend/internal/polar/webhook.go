package polar

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/gamification"
	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/pkg/response"
)

const (
	webhookTimestampTolerance = 5 * time.Minute
	supporterBadgeName        = "Supporter"
)

type WebhookHandler struct {
	webhookSecret    string
	userRepo         *user.Repository
	gamificationRepo *gamification.Repository
}

func NewWebhookHandler(webhookSecret string, userRepo *user.Repository, gamificationRepo *gamification.Repository) *WebhookHandler {
	return &WebhookHandler{
		webhookSecret:    webhookSecret,
		userRepo:         userRepo,
		gamificationRepo: gamificationRepo,
	}
}

type webhookPayload struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type webhookCustomer struct {
	ExternalID string `json:"external_id"`
	ID         string `json:"id"`
}

type orderData struct {
	Customer      webhookCustomer `json:"customer"`
	BillingReason string          `json:"billing_reason"`
}

type subscriptionData struct {
	Customer webhookCustomer `json:"customer"`
}

func (h *WebhookHandler) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
	if err != nil {
		response.BadRequest(c, "failed to read request body")
		return
	}

	msgID := c.GetHeader("webhook-id")
	timestamp := c.GetHeader("webhook-timestamp")
	signature := c.GetHeader("webhook-signature")

	if msgID == "" || timestamp == "" || signature == "" {
		response.BadRequest(c, "missing webhook headers")
		return
	}

	if err := h.verifySignature(msgID, timestamp, signature, body); err != nil {
		slog.Warn("webhook signature verification failed", "error", err)
		response.Error(c, http.StatusUnauthorized, "invalid webhook signature")
		return
	}

	var payload webhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		response.BadRequest(c, "invalid webhook payload")
		return
	}

	ctx := c.Request.Context()

	switch payload.Type {
	case "order.paid":
		h.handleOrderPaid(ctx, payload.Data)
	case "subscription.active":
		h.handleSubscriptionActive(ctx, payload.Data)
	case "subscription.canceled", "subscription.revoked":
		h.handleSubscriptionEnded(ctx, payload.Data)
	default:
		slog.Debug("ignoring unhandled webhook event", "type", payload.Type)
	}

	c.Status(http.StatusOK)
}

func (h *WebhookHandler) verifySignature(msgID, timestampStr, signatureHeader string, body []byte) error {
	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	webhookTime := time.Unix(ts, 0)
	if math.Abs(time.Since(webhookTime).Seconds()) > webhookTimestampTolerance.Seconds() {
		return fmt.Errorf("timestamp outside tolerance")
	}

	// Standard Webhooks: sign "msg_id.timestamp.body"
	toSign := fmt.Sprintf("%s.%s.%s", msgID, timestampStr, string(body))

	secret := strings.TrimPrefix(h.webhookSecret, "whsec_")
	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		secretBytes = []byte(secret)
	}

	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(toSign))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Signature header is space-delimited, each prefixed with version (e.g. "v1,base64sig")
	for _, sig := range strings.Split(signatureHeader, " ") {
		parts := strings.SplitN(sig, ",", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == "v1" && hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
			return nil
		}
	}

	return fmt.Errorf("no matching signature found")
}

func (h *WebhookHandler) handleOrderPaid(ctx context.Context, data json.RawMessage) {
	var order orderData
	if err := json.Unmarshal(data, &order); err != nil {
		slog.Error("failed to parse order data", "error", err)
		return
	}

	if order.BillingReason != "purchase" {
		return
	}

	if order.Customer.ExternalID == "" {
		slog.Warn("order.paid webhook missing external_id")
		return
	}

	userID, err := uuid.Parse(order.Customer.ExternalID)
	if err != nil {
		slog.Warn("order.paid webhook has invalid external_id", "external_id", order.Customer.ExternalID)
		return
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Warn("order.paid webhook user not found", "user_id", userID, "error", err)
		return
	}

	now := time.Now()
	supporterSince := &now
	if u.SupporterSince != nil {
		supporterSince = u.SupporterSince
	}

	if err := h.userRepo.UpdateSupporterStatus(ctx, userID, true, u.IsSubscriber, supporterSince, order.Customer.ID); err != nil {
		slog.Error("failed to update supporter status", "user_id", userID, "error", err)
		return
	}

	h.awardSupporterBadge(ctx, userID)
}

func (h *WebhookHandler) handleSubscriptionActive(ctx context.Context, data json.RawMessage) {
	var sub subscriptionData
	if err := json.Unmarshal(data, &sub); err != nil {
		slog.Error("failed to parse subscription data", "error", err)
		return
	}

	if sub.Customer.ExternalID == "" {
		slog.Warn("subscription.active webhook missing external_id")
		return
	}

	userID, err := uuid.Parse(sub.Customer.ExternalID)
	if err != nil {
		slog.Warn("subscription.active webhook has invalid external_id", "external_id", sub.Customer.ExternalID)
		return
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Warn("subscription.active webhook user not found", "user_id", userID, "error", err)
		return
	}

	now := time.Now()
	supporterSince := &now
	if u.SupporterSince != nil {
		supporterSince = u.SupporterSince
	}

	if err := h.userRepo.UpdateSupporterStatus(ctx, userID, u.IsOneTimeSupporter, true, supporterSince, sub.Customer.ID); err != nil {
		slog.Error("failed to update supporter status", "user_id", userID, "error", err)
		return
	}

	h.awardSupporterBadge(ctx, userID)
}

func (h *WebhookHandler) handleSubscriptionEnded(ctx context.Context, data json.RawMessage) {
	var sub subscriptionData
	if err := json.Unmarshal(data, &sub); err != nil {
		slog.Error("failed to parse subscription data", "error", err)
		return
	}

	if sub.Customer.ExternalID == "" {
		slog.Warn("subscription ended webhook missing external_id")
		return
	}

	userID, err := uuid.Parse(sub.Customer.ExternalID)
	if err != nil {
		slog.Warn("subscription ended webhook has invalid external_id", "external_id", sub.Customer.ExternalID)
		return
	}

	u, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		slog.Warn("subscription ended webhook user not found", "user_id", userID, "error", err)
		return
	}

	if err := h.userRepo.UpdateSupporterStatus(ctx, userID, u.IsOneTimeSupporter, false, u.SupporterSince, u.DerefPolarCustomerID()); err != nil {
		slog.Error("failed to update supporter status", "user_id", userID, "error", err)
		return
	}

	// Only revoke badge if user is not also a one-time supporter
	if !u.IsOneTimeSupporter {
		h.revokeSupporterBadge(ctx, userID)
	}
}

func (h *WebhookHandler) awardSupporterBadge(ctx context.Context, userID uuid.UUID) {
	badge, err := h.gamificationRepo.GetBadgeByName(ctx, supporterBadgeName)
	if err != nil {
		slog.Error("failed to find Supporter badge", "error", err)
		return
	}

	if err := h.gamificationRepo.AwardBadge(ctx, userID, badge.ID); err != nil {
		slog.Error("failed to award Supporter badge", "user_id", userID, "error", err)
	}
}

func (h *WebhookHandler) revokeSupporterBadge(ctx context.Context, userID uuid.UUID) {
	badge, err := h.gamificationRepo.GetBadgeByName(ctx, supporterBadgeName)
	if err != nil {
		slog.Error("failed to find Supporter badge", "error", err)
		return
	}

	if err := h.gamificationRepo.RevokeBadge(ctx, userID, badge.ID); err != nil {
		slog.Error("failed to revoke Supporter badge", "user_id", userID, "error", err)
	}
}
