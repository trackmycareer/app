package response

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type envelopeWithAwards struct {
	Data   any `json:"data,omitempty"`
	Awards any `json:"awards,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, envelope{Data: data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, envelope{Data: data})
}

func OKWithAwards(c *gin.Context, data any, awards any) {
	c.JSON(http.StatusOK, envelopeWithAwards{Data: data, Awards: awards})
}

func CreatedWithAwards(c *gin.Context, data any, awards any) {
	c.JSON(http.StatusCreated, envelopeWithAwards{Data: data, Awards: awards})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, envelope{Message: message})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorised(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func InternalError(c *gin.Context, err error) {
	slog.Error("internal error",
		"error", err.Error(),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)
	Error(c, http.StatusInternalServerError, "an internal error occurred")
}
