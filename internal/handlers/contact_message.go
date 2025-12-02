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

	// Check honeypot for spam
	if req.IsSpam() {
		// Silently accept but don't save - don't let bots know they've been detected
		c.JSON(http.StatusCreated, gin.H{"message": "Thank you for your message"})
		return
	}

	message := &models.ContactMessage{
		Name:    req.Name,
		Email:   req.Email,
		Subject: req.Subject,
		Message: req.Message,
		Status:  models.MessageStatusPending,
	}

	if err := h.repo.CreateContactMessage(c.Request.Context(), message); err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to submit message")
		return
	}

	// Publish to queue for async processing (email notifications)
	event := models.ContactMessageEvent{MessageID: message.ID}
	if err := h.publisher.Publish(c.Request.Context(), event); err != nil {
		// Log error but don't fail the request - message is saved, notification can be retried
		logger.GetLogger(c).Error("Failed to publish message to queue", "error", err, "messageId", message.ID)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Thank you for your message"})
}

// GetContactMessages godoc
// @Summary Get all contact messages
// @Description Returns all contact messages (admin only)
// @Tags Messages
// @Produce json
// @Success 200 {array} models.ContactMessage
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /messages [get]
func (h *Handler) GetContactMessages(c *gin.Context) {
	messages, err := h.repo.GetContactMessages(c.Request.Context())
	if err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to retrieve messages")
		return
	}

	c.JSON(http.StatusOK, messages)
}

// GetContactMessage godoc
// @Summary Get contact message by ID
// @Description Returns a single contact message (admin only)
// @Tags Messages
// @Produce json
// @Param id path int true "Message ID"
// @Success 200 {object} models.ContactMessage
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /messages/{id} [get]
func (h *Handler) GetContactMessage(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	message, err := h.repo.GetContactMessageByID(c.Request.Context(), id)
	if err != nil {
		commonhandlers.HandleRepositoryError(c, err, "Message not found", "Failed to retrieve message")
		return
	}

	c.JSON(http.StatusOK, message)
}
