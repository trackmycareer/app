package skill

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

// Search handles GET /skills/search?q=<query>.
// Returns skill name and category suggestions drawn from the user's
// existing data and a curated list of common skills.
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
