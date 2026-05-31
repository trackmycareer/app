package importer

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

const maxUploadSize = 10 << 20 // 10 MB

// Handler provides HTTP endpoints for the import feature.
type Handler struct {
	svc *Service
}

// NewHandler creates a new import Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Preview handles POST /api/v1/import/preview.
// Accepts multipart form with "file" (required) and "source" (optional, auto-detected).
// For CSV imports, an optional "entity_type" field can specify the entity type.
func (h *Handler) Preview(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	if header.Size > maxUploadSize {
		response.BadRequest(c, fmt.Sprintf("file size exceeds maximum of %d MB", maxUploadSize/(1<<20)))
		return
	}

	// Read file data with independent size enforcement (do not trust header.Size)
	limitedReader := io.LimitReader(file, maxUploadSize+1)
	data, readErr := io.ReadAll(limitedReader)
	if readErr != nil {
		response.BadRequest(c, "failed to read uploaded file")
		return
	}
	if int64(len(data)) > maxUploadSize {
		response.BadRequest(c, fmt.Sprintf("file size exceeds maximum of %d MB", maxUploadSize/(1<<20)))
		return
	}

	filename := header.Filename
	source := strings.TrimSpace(c.PostForm("source"))

	// Auto-detect source from file extension if not provided
	if source == "" {
		source = detectSource(filename)
	}

	if source == "" {
		response.BadRequest(c, "could not determine import source; please specify the 'source' field (linkedin, jsonresume, csv, or document)")
		return
	}

	// Verify content type matches detected source
	if source != "" {
		if err := verifyContentType(data, source); err != nil {
			response.BadRequest(c, err.Error())
			return
		}
	}

	// For CSV, allow overriding entity type via form field
	if source == "csv" {
		if entityType := strings.TrimSpace(c.PostForm("entity_type")); entityType != "" {
			preview, parseErr := ParseCSV(data, entityType)
			if parseErr != nil {
				response.BadRequest(c, fmt.Sprintf("failed to parse CSV: %v", parseErr))
				return
			}
			response.OK(c, preview)
			return
		}
	}

	preview, parseErr := h.svc.Preview(data, filename, source)
	if parseErr != nil {
		response.BadRequest(c, fmt.Sprintf("failed to parse file: %v", parseErr))
		return
	}

	response.OK(c, preview)
}

// Confirm handles POST /api/v1/import/confirm.
// Accepts JSON body containing the ImportPreview (potentially edited by the user).
func (h *Handler) Confirm(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var preview ImportPreview
	if err := c.ShouldBindJSON(&preview); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	result, err := h.svc.Confirm(c.Request.Context(), userID, &preview)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.OK(c, result)
}

// Template handles GET /api/v1/import/templates/:type.
// Returns a CSV template file for the given entity type.
func (h *Handler) Template(c *gin.Context) {
	entityType := c.Param("type")
	if entityType == "" {
		response.BadRequest(c, "entity type is required")
		return
	}

	data, err := GenerateTemplate(entityType)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	filename := fmt.Sprintf("trackmy-career-%s-template.csv", entityType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(200, "text/csv; charset=utf-8", data)
}

// detectSource auto-detects the import source from the file extension.
func detectSource(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".zip":
		return "linkedin"
	case ".json":
		return "jsonresume"
	case ".csv":
		return "csv"
	case ".pdf", ".docx":
		return "document"
	default:
		return ""
	}
}

// verifyContentType checks that the file's magic bytes are consistent with the
// detected source type. This prevents attackers from uploading a malicious file
// with a spoofed extension.
func verifyContentType(data []byte, source string) error {
	if len(data) < 4 {
		return fmt.Errorf("file too small to verify")
	}
	switch source {
	case "linkedin":
		// ZIP magic bytes: PK (0x504B0304)
		if data[0] != 0x50 || data[1] != 0x4B {
			return fmt.Errorf("file does not appear to be a valid ZIP archive")
		}
	case "jsonresume":
		// JSON should start with { or [ (after whitespace)
		trimmed := bytes.TrimLeft(data, " \t\r\n")
		if len(trimmed) == 0 || (trimmed[0] != '{' && trimmed[0] != '[') {
			return fmt.Errorf("file does not appear to be valid JSON")
		}
	case "document":
		// PDF: %PDF or DOCX: PK (ZIP-based)
		contentType := http.DetectContentType(data)
		if !strings.Contains(contentType, "pdf") &&
			!strings.Contains(contentType, "zip") &&
			!strings.Contains(contentType, "octet-stream") {
			return fmt.Errorf("file does not appear to be a valid PDF or DOCX")
		}
		// CSV: no reliable magic bytes, skip verification
	}
	return nil
}
