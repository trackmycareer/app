package polar

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/config"
	"github.com/trackmycareer/app/pkg/response"
)

type Handler struct {
	client *Client
	cfg    config.Config
}

func NewHandler(client *Client, cfg config.Config) *Handler {
	return &Handler{
		client: client,
		cfg:    cfg,
	}
}

type configResponse struct {
	PolarEnabled bool `json:"polar_enabled"`
}

func GetConfigHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.OK(c, configResponse{
			PolarEnabled: cfg.PolarEnabled(),
		})
	}
}

func (h *Handler) CreateCheckout(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	email := c.MustGet("email").(string)

	supportType := c.DefaultQuery("type", "one_time")

	var productID string
	switch supportType {
	case "one_time":
		productID = h.cfg.PolarProductIDOneTime
	case "subscription":
		productID = h.cfg.PolarProductIDSubscription
	default:
		response.BadRequest(c, "invalid support type: must be one_time or subscription")
		return
	}

	if productID == "" {
		response.Error(c, http.StatusNotFound, "this support option is not available")
		return
	}

	successURL := h.cfg.FrontendURL + "/support/thank-you"
	checkoutURL, err := h.client.CreateCheckoutSession(c.Request.Context(), productID, userID.String(), email, successURL)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"checkout_url": checkoutURL})
}

func (h *Handler) GetPortalURL(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	returnURL := h.cfg.FrontendURL + "/support"
	portalURL, err := h.client.CreateCustomerPortalSession(c.Request.Context(), userID.String(), returnURL)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"portal_url": portalURL})
}
