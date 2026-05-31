package response

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, message)
}

// FormatBindingError converts a Gin binding/validation error into a
// human-readable message that does not leak struct names or internal
// validator details. For validator.ValidationErrors each field error is
// mapped to a clear sentence; for any other error type it returns a
// generic message.
func FormatBindingError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		msgs := make([]string, 0, len(ve))
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			switch fe.Tag() {
			case "required":
				msgs = append(msgs, field+" is required")
			case "min":
				msgs = append(msgs, field+" must be at least "+fe.Param()+" characters")
			case "max":
				msgs = append(msgs, field+" must be at most "+fe.Param()+" characters")
			case "email":
				msgs = append(msgs, field+" must be a valid email address")
			case "oneof":
				msgs = append(msgs, field+" must be one of: "+fe.Param())
			default:
				msgs = append(msgs, field+" is invalid")
			}
		}
		return strings.Join(msgs, "; ")
	}
	return "invalid request body"
}

func InternalError(c *gin.Context, err error) {
	slog.Error("internal error",
		"error", err.Error(),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
	)
	Error(c, http.StatusInternalServerError, "an internal error occurred")
}
