package handlers

import (
	"net/http"
	"strconv"

	"github.com/GunarsK-templates/template-api/internal/models"
	"github.com/gin-gonic/gin"
)

// GetItems godoc
// @Summary Get all items
// @Description Returns a list of all items
// @Tags Items
// @Produce json
// @Success 200 {array} models.Item
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/items [get]
func (h *Handler) GetItems(c *gin.Context) {
	items, err := h.repo.GetAllItems(c.Request.Context())
	if err != nil {
		LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to retrieve items")
		return
	}
	c.JSON(http.StatusOK, items)
}

// GetItem godoc
// @Summary Get item by ID
// @Description Returns a single item by ID
// @Tags Items
// @Produce json
// @Param id path int true "Item ID"
// @Success 200 {object} models.Item
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/items/{id} [get]
func (h *Handler) GetItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	item, err := h.repo.GetItemByID(c.Request.Context(), id)
	if err != nil {
		HandleRepositoryError(c, err, "Item not found", "Failed to retrieve item")
		return
	}
	c.JSON(http.StatusOK, item)
}

// CreateItem godoc
// @Summary Create a new item
// @Description Creates a new item
// @Tags Items
// @Accept json
// @Produce json
// @Param item body models.CreateItemRequest true "Item data"
// @Success 201 {object} models.Item
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/items [post]
func (h *Handler) CreateItem(c *gin.Context) {
	var req models.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	item := &models.Item{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.repo.CreateItem(c.Request.Context(), item); err != nil {
		LogAndRespondError(c, http.StatusInternalServerError, err, "Failed to create item")
		return
	}
	c.JSON(http.StatusCreated, item)
}

// UpdateItem godoc
// @Summary Update an item
// @Description Updates an existing item
// @Tags Items
// @Accept json
// @Produce json
// @Param id path int true "Item ID"
// @Param item body models.UpdateItemRequest true "Item data"
// @Success 200 {object} models.Item
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/items/{id} [put]
func (h *Handler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	var req models.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	item := &models.Item{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.repo.UpdateItem(c.Request.Context(), item); err != nil {
		HandleRepositoryError(c, err, "Item not found", "Failed to update item")
		return
	}
	c.JSON(http.StatusOK, item)
}

// DeleteItem godoc
// @Summary Delete an item
// @Description Deletes an item by ID
// @Tags Items
// @Param id path int true "Item ID"
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/items/{id} [delete]
func (h *Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "Invalid ID format")
		return
	}

	if err := h.repo.DeleteItem(c.Request.Context(), id); err != nil {
		HandleRepositoryError(c, err, "Item not found", "Failed to delete item")
		return
	}
	c.Status(http.StatusNoContent)
}
