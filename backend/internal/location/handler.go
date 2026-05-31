package location

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

// Handler exposes HTTP endpoints for location search.
type Handler struct {
	svc *Service
}

// NewHandler creates a new location Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Search handles GET /locations/search?q=<query>.
// Returns location suggestions drawn from the user's existing data and the
// Photon geocoder.
func (h *Handler) Search(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	q := c.Query("q")
	if len(q) < 2 {
		response.BadRequest(c, "query must be at least 2 characters")
		return
	}
	if len(q) > 160 {
		response.BadRequest(c, "query must be at most 160 characters")
		return
	}

	resp, err := h.svc.Search(c.Request.Context(), userID, q)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, resp)
}
