package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// RespondError sends an error response without logging (for expected errors like validation)
func RespondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{Error: message})
}

// LogAndRespondError logs the error and sends an error response
func LogAndRespondError(c *gin.Context, statusCode int, err error, userMessage string) {
	slog.Error("Request error",
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"status", statusCode,
		"error", err.Error(),
	)
	c.JSON(statusCode, ErrorResponse{Error: userMessage})
}

// HandleRepositoryError handles repository errors with appropriate responses
// - Returns 404 for gorm.ErrRecordNotFound
// - Returns 500 and logs for other errors
func HandleRepositoryError(c *gin.Context, err error, notFoundMsg, internalMsg string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		RespondError(c, http.StatusNotFound, notFoundMsg)
		return
	}
	LogAndRespondError(c, http.StatusInternalServerError, err, internalMsg)
}
