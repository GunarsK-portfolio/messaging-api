package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	commonhandlers "github.com/GunarsK-portfolio/portfolio-common/handlers"
	"github.com/GunarsK-portfolio/portfolio-common/logger"
	"github.com/GunarsK-portfolio/portfolio-common/models"
)

// CreateContactMessage godoc
// @Summary Submit a contact message
// @Description Creates a new contact message (public endpoint)
// @Tags Contact
// @Accept json
// @Produce json
// @Param message body models.ContactMessageCreate true "Contact message"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /contact [post]
func (h *Handler) CreateContactMessage(c *gin.Context) {
	var req models.ContactMessageCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.IsSpam() {
		c.JSON(http.StatusCreated, gin.H{"message": "Thank you for your message"})
		return
	}

	email := &models.Email{
		Type:        models.EmailTypeContactForm,
		Name:        &req.Name,
		SenderEmail: &req.Email,
		Subject:     req.Subject,
		Message:     req.Message,
		Status:      models.EmailStatusPending,
	}

	if err := h.repo.CreateEmail(c.Request.Context(), email); err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to submit message")
		return
	}

	event := models.EmailEvent{EmailID: email.ID}
	if err := h.publisher.Publish(c.Request.Context(), event); err != nil {
		logger.GetLogger(c).Error("Failed to publish message to queue", "error", err, "emailId", email.ID)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Thank you for your message"})
}

// GetEmails godoc
// @Summary Get all emails
// @Description Returns all emails (admin only)
// @Tags Emails
// @Produce json
// @Success 200 {array} models.Email
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /emails [get]
func (h *Handler) GetEmails(c *gin.Context) {
	emails, err := h.repo.GetEmails(c.Request.Context())
	if err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to retrieve emails")
		return
	}

	c.JSON(http.StatusOK, emails)
}

// GetEmail godoc
// @Summary Get email by ID
// @Description Returns a single email (admin only)
// @Tags Emails
// @Produce json
// @Param id path int true "Email ID"
// @Success 200 {object} models.Email
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /emails/{id} [get]
func (h *Handler) GetEmail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	email, err := h.repo.GetEmailByID(c.Request.Context(), id)
	if err != nil {
		commonhandlers.HandleRepositoryError(c, err, "Email not found", "Failed to retrieve email")
		return
	}

	c.JSON(http.StatusOK, email)
}
