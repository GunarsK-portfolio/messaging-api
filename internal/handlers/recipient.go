package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	commonhandlers "github.com/GunarsK-portfolio/portfolio-common/handlers"
	"github.com/GunarsK-portfolio/portfolio-common/models"
)

// GetRecipients godoc
// @Summary Get all recipients
// @Description Returns a list of all email recipients (admin only)
// @Tags Recipients
// @Produce json
// @Success 200 {array} models.Recipient
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /recipients [get]
func (h *Handler) GetRecipients(c *gin.Context) {
	recipients, err := h.repo.GetAllRecipients(c.Request.Context())
	if err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to retrieve recipients")
		return
	}
	c.JSON(http.StatusOK, recipients)
}

// GetRecipient godoc
// @Summary Get recipient by ID
// @Description Returns a single recipient by ID (admin only)
// @Tags Recipients
// @Produce json
// @Param id path int true "Recipient ID"
// @Success 200 {object} models.Recipient
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /recipients/{id} [get]
func (h *Handler) GetRecipient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	recipient, err := h.repo.GetRecipientByID(c.Request.Context(), id)
	if err != nil {
		commonhandlers.HandleRepositoryError(c, err, "Recipient not found", "Failed to retrieve recipient")
		return
	}
	c.JSON(http.StatusOK, recipient)
}

// CreateRecipient godoc
// @Summary Create a new recipient
// @Description Creates a new email recipient (admin only)
// @Tags Recipients
// @Accept json
// @Produce json
// @Param recipient body models.RecipientCreate true "Recipient data"
// @Success 201 {object} models.Recipient
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /recipients [post]
func (h *Handler) CreateRecipient(c *gin.Context) {
	var req models.RecipientCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	recipient := &models.Recipient{
		Email:    req.Email,
		Name:     req.Name,
		IsActive: true,
	}
	if req.IsActive != nil {
		recipient.IsActive = *req.IsActive
	}

	if err := h.repo.CreateRecipient(c.Request.Context(), recipient); err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to create recipient")
		return
	}

	setLocationHeader(c, recipient.ID)
	c.JSON(http.StatusCreated, recipient)
}

// UpdateRecipient godoc
// @Summary Update a recipient
// @Description Updates an existing recipient (admin only)
// @Tags Recipients
// @Accept json
// @Produce json
// @Param id path int true "Recipient ID"
// @Param recipient body models.RecipientUpdate true "Recipient data"
// @Success 200 {object} models.Recipient
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /recipients/{id} [put]
func (h *Handler) UpdateRecipient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	// Get existing recipient
	existing, err := h.repo.GetRecipientByID(c.Request.Context(), id)
	if err != nil {
		commonhandlers.HandleRepositoryError(c, err, "Recipient not found", "Failed to retrieve recipient")
		return
	}

	var req models.RecipientUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Apply updates
	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := h.repo.UpdateRecipient(c.Request.Context(), existing); err != nil {
		commonhandlers.LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to update recipient")
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteRecipient godoc
// @Summary Delete a recipient
// @Description Deletes a recipient by ID (admin only)
// @Tags Recipients
// @Param id path int true "Recipient ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security BearerAuth
// @Router /recipients/{id} [delete]
func (h *Handler) DeleteRecipient(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		commonhandlers.RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if err := h.repo.DeleteRecipient(c.Request.Context(), id); err != nil {
		commonhandlers.HandleRepositoryError(c, err, "Recipient not found", "Failed to delete recipient")
		return
	}

	c.Status(http.StatusNoContent)
}
