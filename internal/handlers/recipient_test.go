package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/GunarsK-portfolio/portfolio-common/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// =============================================================================
// GetRecipients Tests
// =============================================================================

func TestGetRecipients_Success(t *testing.T) {
	expected := createTestRecipients()
	mockRepo := &mockRepository{
		getAllRecipientsFunc: func(_ context.Context) ([]models.Recipient, error) {
			return expected, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients", handler.GetRecipients)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.Recipient
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf("expected %d recipients, got %d", len(expected), len(result))
	}
}

func TestGetRecipients_Empty(t *testing.T) {
	mockRepo := &mockRepository{
		getAllRecipientsFunc: func(_ context.Context) ([]models.Recipient, error) {
			return []models.Recipient{}, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients", handler.GetRecipients)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.Recipient
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 recipients, got %d", len(result))
	}
}

func TestGetRecipients_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		getAllRecipientsFunc: func(_ context.Context) ([]models.Recipient, error) {
			return nil, errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients", handler.GetRecipients)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// GetRecipient Tests
// =============================================================================

func TestGetRecipient_Success(t *testing.T) {
	expected := createTestRecipient()
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, id int64) (*models.Recipient, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return expected, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients/:id", handler.GetRecipient)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients/1", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result models.Recipient
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.ID != expected.ID {
		t.Errorf("expected ID %d, got %d", expected.ID, result.ID)
	}
	if result.Email != expected.Email {
		t.Errorf("expected email %s, got %s", expected.Email, result.Email)
	}
}

func TestGetRecipient_InvalidID(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients/:id", handler.GetRecipient)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients/invalid", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid ID format") {
		t.Errorf("expected 'Invalid ID format' error, got %s", w.Body.String())
	}
}

func TestGetRecipient_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, _ int64) (*models.Recipient, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients/:id", handler.GetRecipient)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients/999", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetRecipient_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, _ int64) (*models.Recipient, error) {
			return nil, errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients/:id", handler.GetRecipient)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients/1", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// CreateRecipient Tests
// =============================================================================

func TestCreateRecipient_Success(t *testing.T) {
	var createdRecipient *models.Recipient
	mockRepo := &mockRepository{
		createRecipientFunc: func(_ context.Context, recipient *models.Recipient) error {
			createdRecipient = recipient
			recipient.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/recipients", handler.CreateRecipient)

	body := `{"email":"new@example.com","name":"New Recipient"}`
	w := performRequest(router, http.MethodPost, "/api/v1/recipients", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if createdRecipient == nil {
		t.Fatal("expected recipient to be created")
	}
	if createdRecipient.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got %s", createdRecipient.Email)
	}
	if createdRecipient.Name != "New Recipient" {
		t.Errorf("expected name 'New Recipient', got %s", createdRecipient.Name)
	}
	if !createdRecipient.IsActive {
		t.Error("expected IsActive to be true by default")
	}

	// Check Location header
	location := w.Header().Get("Location")
	if location != "/api/v1/recipients/1" {
		t.Errorf("expected Location header '/api/v1/recipients/1', got %s", location)
	}
}

func TestCreateRecipient_WithIsActiveFalse(t *testing.T) {
	var createdRecipient *models.Recipient
	mockRepo := &mockRepository{
		createRecipientFunc: func(_ context.Context, recipient *models.Recipient) error {
			createdRecipient = recipient
			recipient.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/recipients", handler.CreateRecipient)

	body := `{"email":"new@example.com","name":"New Recipient","isActive":false}`
	w := performRequest(router, http.MethodPost, "/api/v1/recipients", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if createdRecipient == nil {
		t.Fatal("expected recipient to be created")
	}
	if createdRecipient.IsActive {
		t.Error("expected IsActive to be false")
	}
}

func TestCreateRecipient_ValidationError(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/recipients", handler.CreateRecipient)

	// Missing required email
	body := `{"name":"Test"}`
	w := performRequest(router, http.MethodPost, "/api/v1/recipients", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateRecipient_InvalidEmail(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/recipients", handler.CreateRecipient)

	body := `{"email":"not-an-email","name":"Test"}`
	w := performRequest(router, http.MethodPost, "/api/v1/recipients", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateRecipient_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		createRecipientFunc: func(_ context.Context, _ *models.Recipient) error {
			return errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/recipients", handler.CreateRecipient)

	body := `{"email":"new@example.com","name":"New Recipient"}`
	w := performRequest(router, http.MethodPost, "/api/v1/recipients", strings.NewReader(body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCreateRecipient_InvalidJSON(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/recipients", handler.CreateRecipient)

	body := `{invalid json}`
	w := performRequest(router, http.MethodPost, "/api/v1/recipients", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// =============================================================================
// UpdateRecipient Tests
// =============================================================================

func TestUpdateRecipient_Success(t *testing.T) {
	existing := createTestRecipient()
	var updatedRecipient *models.Recipient
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, id int64) (*models.Recipient, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			// Return a copy to avoid mutation
			copy := *existing
			return &copy, nil
		},
		updateRecipientFunc: func(_ context.Context, recipient *models.Recipient) error {
			updatedRecipient = recipient
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)

	body := `{"email":"updated@example.com","name":"Updated Name","isActive":false}`
	w := performRequest(router, http.MethodPut, "/api/v1/recipients/1", strings.NewReader(body))

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if updatedRecipient == nil {
		t.Fatal("expected recipient to be updated")
	}
	if updatedRecipient.Email != "updated@example.com" {
		t.Errorf("expected email 'updated@example.com', got %s", updatedRecipient.Email)
	}
	if updatedRecipient.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got %s", updatedRecipient.Name)
	}
	if updatedRecipient.IsActive {
		t.Error("expected IsActive to be false")
	}
}

func TestUpdateRecipient_PartialUpdate(t *testing.T) {
	existing := createTestRecipient()
	var updatedRecipient *models.Recipient
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, _ int64) (*models.Recipient, error) {
			copy := *existing
			return &copy, nil
		},
		updateRecipientFunc: func(_ context.Context, recipient *models.Recipient) error {
			updatedRecipient = recipient
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)

	// Only update name, keep other fields
	body := `{"name":"Only Name Updated"}`
	w := performRequest(router, http.MethodPut, "/api/v1/recipients/1", strings.NewReader(body))

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if updatedRecipient == nil {
		t.Fatal("expected recipient to be updated")
	}
	// Name should be updated
	if updatedRecipient.Name != "Only Name Updated" {
		t.Errorf("expected name 'Only Name Updated', got %s", updatedRecipient.Name)
	}
	// Email should remain unchanged
	if updatedRecipient.Email != existing.Email {
		t.Errorf("expected email to remain %s, got %s", existing.Email, updatedRecipient.Email)
	}
	// IsActive should remain unchanged
	if updatedRecipient.IsActive != existing.IsActive {
		t.Errorf("expected IsActive to remain %v, got %v", existing.IsActive, updatedRecipient.IsActive)
	}
}

func TestUpdateRecipient_InvalidID(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)

	body := `{"name":"Test"}`
	w := performRequest(router, http.MethodPut, "/api/v1/recipients/invalid", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid ID format") {
		t.Errorf("expected 'Invalid ID format' error, got %s", w.Body.String())
	}
}

func TestUpdateRecipient_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, _ int64) (*models.Recipient, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)

	body := `{"name":"Test"}`
	w := performRequest(router, http.MethodPut, "/api/v1/recipients/999", strings.NewReader(body))

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestUpdateRecipient_ValidationError(t *testing.T) {
	existing := createTestRecipient()
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, _ int64) (*models.Recipient, error) {
			copy := *existing
			return &copy, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)

	// Invalid email format
	body := `{"email":"not-an-email"}`
	w := performRequest(router, http.MethodPut, "/api/v1/recipients/1", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateRecipient_RepositoryError(t *testing.T) {
	existing := createTestRecipient()
	mockRepo := &mockRepository{
		getRecipientByIDFunc: func(_ context.Context, _ int64) (*models.Recipient, error) {
			copy := *existing
			return &copy, nil
		},
		updateRecipientFunc: func(_ context.Context, _ *models.Recipient) error {
			return errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)

	body := `{"name":"Test"}`
	w := performRequest(router, http.MethodPut, "/api/v1/recipients/1", strings.NewReader(body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// DeleteRecipient Tests
// =============================================================================

func TestDeleteRecipient_Success(t *testing.T) {
	deleteCalled := false
	mockRepo := &mockRepository{
		deleteRecipientFunc: func(_ context.Context, id int64) error {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			deleteCalled = true
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.DELETE("/api/v1/recipients/:id", handler.DeleteRecipient)

	w := performRequest(router, http.MethodDelete, "/api/v1/recipients/1", nil)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	if !deleteCalled {
		t.Error("expected delete to be called")
	}
}

func TestDeleteRecipient_InvalidID(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.DELETE("/api/v1/recipients/:id", handler.DeleteRecipient)

	w := performRequest(router, http.MethodDelete, "/api/v1/recipients/invalid", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid ID format") {
		t.Errorf("expected 'Invalid ID format' error, got %s", w.Body.String())
	}
}

func TestDeleteRecipient_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		deleteRecipientFunc: func(_ context.Context, _ int64) error {
			return gorm.ErrRecordNotFound
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.DELETE("/api/v1/recipients/:id", handler.DeleteRecipient)

	w := performRequest(router, http.MethodDelete, "/api/v1/recipients/999", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteRecipient_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		deleteRecipientFunc: func(_ context.Context, _ int64) error {
			return errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.DELETE("/api/v1/recipients/:id", handler.DeleteRecipient)

	w := performRequest(router, http.MethodDelete, "/api/v1/recipients/1", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// Context Propagation Tests (Recipients)
// =============================================================================

func TestGetRecipients_ContextPropagation(t *testing.T) {
	var capturedCtx context.Context
	mockRepo := &mockRepository{
		getAllRecipientsFunc: func(ctx context.Context) ([]models.Recipient, error) {
			capturedCtx = ctx
			return []models.Recipient{}, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ctxKey{}, "test-marker")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	router.GET("/api/v1/recipients", handler.GetRecipients)

	w := performRequest(router, http.MethodGet, "/api/v1/recipients", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if capturedCtx == nil {
		t.Error("expected context to be propagated to repository")
	}

	if capturedCtx.Value(ctxKey{}) != "test-marker" {
		t.Error("context sentinel value was not propagated to repository")
	}
}

// =============================================================================
// Invalid ID Format Tests (Recipients)
// =============================================================================

func TestInvalidIDFormats_Recipients(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/recipients/:id", handler.GetRecipient)
	router.PUT("/api/v1/recipients/:id", handler.UpdateRecipient)
	router.DELETE("/api/v1/recipients/:id", handler.DeleteRecipient)

	tests := []struct {
		name string
		id   string
	}{
		{"alphabetic", "abc"},
		{"special characters", "!@#"},
		{"float", "1.5"},
		{"overflow", "99999999999999999999"},
	}

	for _, tt := range tests {
		t.Run("GET_"+tt.name, func(t *testing.T) {
			w := performRequest(router, http.MethodGet, "/api/v1/recipients/"+tt.id, nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("GetRecipient(%q) status = %d, want %d", tt.id, w.Code, http.StatusBadRequest)
			}
		})

		t.Run("PUT_"+tt.name, func(t *testing.T) {
			w := performRequest(router, http.MethodPut, "/api/v1/recipients/"+tt.id, strings.NewReader(`{}`))
			if w.Code != http.StatusBadRequest {
				t.Errorf("UpdateRecipient(%q) status = %d, want %d", tt.id, w.Code, http.StatusBadRequest)
			}
		})

		t.Run("DELETE_"+tt.name, func(t *testing.T) {
			w := performRequest(router, http.MethodDelete, "/api/v1/recipients/"+tt.id, nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("DeleteRecipient(%q) status = %d, want %d", tt.id, w.Code, http.StatusBadRequest)
			}
		})
	}
}
