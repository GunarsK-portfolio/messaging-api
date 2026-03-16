package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	commonhandlers "github.com/GunarsK-portfolio/portfolio-common/handlers"
	"github.com/GunarsK-portfolio/portfolio-common/logger"
	"github.com/GunarsK-portfolio/portfolio-common/models"

	"github.com/GunarsK-portfolio/messaging-api/internal/renderer"
)

// SendEmailRequest is the DTO for the S2S email endpoint
type SendEmailRequest struct {
	Type           string            `json:"type" binding:"required"`
	RecipientEmail string            `json:"recipient_email" binding:"required,email"`
	Data           map[string]string `json:"data" binding:"required"`
}

// SendEmail godoc
// @Summary Send a templated email (S2S)
// @Description Renders a template and queues an email for delivery. Requires emails:edit scope.
// @Tags Emails
// @Accept json
// @Produce json
// @Param email body SendEmailRequest true "Email request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /emails [post]
func (h *Handler) SendEmail(c *gin.Context) {
	var req SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	subject := renderer.SubjectForType(req.Type)
	if subject == "" {
		commonhandlers.RespondError(c, http.StatusBadRequest, "unsupported email type: "+req.Type)
		return
	}

	html, err := renderer.Render(req.Type, req.Data)
	if err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to render template")
		return
	}

	email := &models.Email{
		Type:           req.Type,
		RecipientEmail: &req.RecipientEmail,
		Subject:        subject,
		Message:        html,
		Status:         models.EmailStatusPending,
	}

	if err := h.repo.CreateEmail(c.Request.Context(), email); err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to create email")
		return
	}

	event := models.EmailEvent{EmailID: email.ID}
	if err := h.publisher.Publish(c.Request.Context(), event); err != nil {
		logger.GetLogger(c).Error("Failed to publish email to queue", "error", err, "emailId", email.ID)
	}

	c.JSON(http.StatusCreated, gin.H{"id": email.ID, "message": "Email queued"})
}
