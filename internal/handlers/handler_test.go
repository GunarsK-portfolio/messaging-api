package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GunarsK-portfolio/messaging-api/internal/repository"
	"github.com/GunarsK-portfolio/portfolio-common/models"
	"github.com/GunarsK-portfolio/portfolio-common/queue"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// =============================================================================
// Mock Repository
// =============================================================================

type mockRepository struct {
	createContactMessageFunc     func(ctx context.Context, message *models.ContactMessage) error
	getContactMessagesFunc       func(ctx context.Context) ([]models.ContactMessage, error)
	getContactMessageByIDFunc    func(ctx context.Context, id int64) (*models.ContactMessage, error)
	updateContactMessageStatusFn func(ctx context.Context, id int64, status string, lastError *string) error
	getAllRecipientsFunc         func(ctx context.Context) ([]models.Recipient, error)
	getActiveRecipientsFunc      func(ctx context.Context) ([]models.Recipient, error)
	getRecipientByIDFunc         func(ctx context.Context, id int64) (*models.Recipient, error)
	createRecipientFunc          func(ctx context.Context, recipient *models.Recipient) error
	updateRecipientFunc          func(ctx context.Context, recipient *models.Recipient) error
	deleteRecipientFunc          func(ctx context.Context, id int64) error
}

func (m *mockRepository) CreateContactMessage(ctx context.Context, message *models.ContactMessage) error {
	if m.createContactMessageFunc != nil {
		return m.createContactMessageFunc(ctx, message)
	}
	return nil
}

func (m *mockRepository) GetContactMessages(ctx context.Context) ([]models.ContactMessage, error) {
	if m.getContactMessagesFunc != nil {
		return m.getContactMessagesFunc(ctx)
	}
	return nil, nil
}

func (m *mockRepository) GetContactMessageByID(ctx context.Context, id int64) (*models.ContactMessage, error) {
	if m.getContactMessageByIDFunc != nil {
		return m.getContactMessageByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) UpdateContactMessageStatus(ctx context.Context, id int64, status string, lastError *string) error {
	if m.updateContactMessageStatusFn != nil {
		return m.updateContactMessageStatusFn(ctx, id, status, lastError)
	}
	return nil
}

func (m *mockRepository) GetAllRecipients(ctx context.Context) ([]models.Recipient, error) {
	if m.getAllRecipientsFunc != nil {
		return m.getAllRecipientsFunc(ctx)
	}
	return nil, nil
}

func (m *mockRepository) GetActiveRecipients(ctx context.Context) ([]models.Recipient, error) {
	if m.getActiveRecipientsFunc != nil {
		return m.getActiveRecipientsFunc(ctx)
	}
	return nil, nil
}

func (m *mockRepository) GetRecipientByID(ctx context.Context, id int64) (*models.Recipient, error) {
	if m.getRecipientByIDFunc != nil {
		return m.getRecipientByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) CreateRecipient(ctx context.Context, recipient *models.Recipient) error {
	if m.createRecipientFunc != nil {
		return m.createRecipientFunc(ctx, recipient)
	}
	return nil
}

func (m *mockRepository) UpdateRecipient(ctx context.Context, recipient *models.Recipient) error {
	if m.updateRecipientFunc != nil {
		return m.updateRecipientFunc(ctx, recipient)
	}
	return nil
}

func (m *mockRepository) DeleteRecipient(ctx context.Context, id int64) error {
	if m.deleteRecipientFunc != nil {
		return m.deleteRecipientFunc(ctx, id)
	}
	return nil
}

// Verify mock implements Repository interface
var _ repository.Repository = (*mockRepository)(nil)

// =============================================================================
// Mock Publisher
// =============================================================================

type mockPublisher struct {
	publishFunc        func(ctx context.Context, message interface{}) error
	publishToRetryFunc func(ctx context.Context, retryIndex int, body []byte, correlationId string) error
	publishToDLQFunc   func(ctx context.Context, body []byte, correlationId string) error
	maxRetries         int
}

func (m *mockPublisher) Publish(ctx context.Context, message interface{}) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, message)
	}
	return nil
}

func (m *mockPublisher) PublishToRetry(ctx context.Context, retryIndex int, body []byte, correlationId string) error {
	if m.publishToRetryFunc != nil {
		return m.publishToRetryFunc(ctx, retryIndex, body, correlationId)
	}
	return nil
}

func (m *mockPublisher) PublishToDLQ(ctx context.Context, body []byte, correlationId string) error {
	if m.publishToDLQFunc != nil {
		return m.publishToDLQFunc(ctx, body, correlationId)
	}
	return nil
}

func (m *mockPublisher) MaxRetries() int {
	return m.maxRetries
}

func (m *mockPublisher) Close() error {
	return nil
}

// Verify mock implements Publisher interface
var _ queue.Publisher = (*mockPublisher)(nil)

// =============================================================================
// Test Helpers
// =============================================================================

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func performRequest(router *gin.Engine, method, path string, body io.Reader) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	router.ServeHTTP(w, req)
	return w
}

func createTestContactMessage() *models.ContactMessage {
	return &models.ContactMessage{
		ID:        1,
		Name:      "John Doe",
		Email:     "john@example.com",
		Subject:   "Test Subject",
		Message:   "Test message content",
		Status:    models.MessageStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestContactMessages() []models.ContactMessage {
	return []models.ContactMessage{
		{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Subject:   "Test Subject 1",
			Message:   "Test message 1",
			Status:    models.MessageStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			Subject:   "Test Subject 2",
			Message:   "Test message 2",
			Status:    models.MessageStatusSent,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

func createTestRecipient() *models.Recipient {
	return &models.Recipient{
		ID:        1,
		Email:     "admin@example.com",
		Name:      "Admin User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestRecipients() []models.Recipient {
	return []models.Recipient{
		{
			ID:        1,
			Email:     "admin@example.com",
			Name:      "Admin User",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Email:     "support@example.com",
			Name:      "Support Team",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNew_ReturnsHandler(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	if handler == nil {
		t.Fatal("expected handler to not be nil")
	}
	if handler.repo == nil {
		t.Error("expected repo to be set")
	}
	if handler.publisher == nil {
		t.Error("expected publisher to be set")
	}
}

// =============================================================================
// Health Check Tests
// =============================================================================

func TestHealthCheck_Success(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/health", handler.HealthCheck)

	w := performRequest(router, http.MethodGet, "/health", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !strings.Contains(w.Body.String(), "healthy") {
		t.Errorf("expected 'healthy' in response, got %s", w.Body.String())
	}
}

// =============================================================================
// CreateContactMessage Tests
// =============================================================================

func TestCreateContactMessage_Success(t *testing.T) {
	var createdMessage *models.ContactMessage
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, message *models.ContactMessage) error {
			createdMessage = message
			message.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John Doe","email":"john@example.com","subject":"Test","message":"Hello world"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if createdMessage == nil {
		t.Fatal("expected message to be created")
	}
	if createdMessage.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %s", createdMessage.Name)
	}
	if createdMessage.Email != "john@example.com" {
		t.Errorf("expected email 'john@example.com', got %s", createdMessage.Email)
	}
	if createdMessage.Status != models.MessageStatusPending {
		t.Errorf("expected status 'pending', got %s", createdMessage.Status)
	}

	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message, got %s", w.Body.String())
	}
}

func TestCreateContactMessage_SpamDetected(t *testing.T) {
	createCalled := false
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, _ *models.ContactMessage) error {
			createCalled = true
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	// Honeypot field (website) is filled - indicates spam
	body := `{"name":"Bot","email":"bot@spam.com","subject":"Spam","message":"Buy now!","website":"http://spam.com"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	// Should return success to not alert bots
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// But should NOT save the message
	if createCalled {
		t.Error("expected message NOT to be saved for spam")
	}

	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message even for spam, got %s", w.Body.String())
	}
}

func TestCreateContactMessage_ValidationError(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	// Missing required fields
	body := `{"name":"John"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateContactMessage_InvalidEmail(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John","email":"not-an-email","subject":"Test","message":"Hello"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateContactMessage_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, _ *models.ContactMessage) error {
			return errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John Doe","email":"john@example.com","subject":"Test","message":"Hello world"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCreateContactMessage_InvalidJSON(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{invalid json}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// =============================================================================
// GetContactMessages Tests
// =============================================================================

func TestGetContactMessages_Success(t *testing.T) {
	expected := createTestContactMessages()
	mockRepo := &mockRepository{
		getContactMessagesFunc: func(_ context.Context) ([]models.ContactMessage, error) {
			return expected, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages", handler.GetContactMessages)

	w := performRequest(router, http.MethodGet, "/api/v1/messages", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.ContactMessage
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf("expected %d messages, got %d", len(expected), len(result))
	}
}

func TestGetContactMessages_Empty(t *testing.T) {
	mockRepo := &mockRepository{
		getContactMessagesFunc: func(_ context.Context) ([]models.ContactMessage, error) {
			return []models.ContactMessage{}, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages", handler.GetContactMessages)

	w := performRequest(router, http.MethodGet, "/api/v1/messages", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.ContactMessage
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected 0 messages, got %d", len(result))
	}
}

func TestGetContactMessages_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		getContactMessagesFunc: func(_ context.Context) ([]models.ContactMessage, error) {
			return nil, errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages", handler.GetContactMessages)

	w := performRequest(router, http.MethodGet, "/api/v1/messages", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// GetContactMessage Tests
// =============================================================================

func TestGetContactMessage_Success(t *testing.T) {
	expected := createTestContactMessage()
	mockRepo := &mockRepository{
		getContactMessageByIDFunc: func(_ context.Context, id int64) (*models.ContactMessage, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return expected, nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages/:id", handler.GetContactMessage)

	w := performRequest(router, http.MethodGet, "/api/v1/messages/1", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result models.ContactMessage
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.ID != expected.ID {
		t.Errorf("expected ID %d, got %d", expected.ID, result.ID)
	}
	if result.Name != expected.Name {
		t.Errorf("expected name %s, got %s", expected.Name, result.Name)
	}
}

func TestGetContactMessage_InvalidID(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages/:id", handler.GetContactMessage)

	w := performRequest(router, http.MethodGet, "/api/v1/messages/invalid", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Invalid ID format") {
		t.Errorf("expected 'Invalid ID format' error, got %s", w.Body.String())
	}
}

func TestGetContactMessage_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getContactMessageByIDFunc: func(_ context.Context, _ int64) (*models.ContactMessage, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages/:id", handler.GetContactMessage)

	w := performRequest(router, http.MethodGet, "/api/v1/messages/999", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetContactMessage_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		getContactMessageByIDFunc: func(_ context.Context, _ int64) (*models.ContactMessage, error) {
			return nil, errors.New("database error")
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages/:id", handler.GetContactMessage)

	w := performRequest(router, http.MethodGet, "/api/v1/messages/1", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

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
// Context Propagation Tests
// =============================================================================

type ctxKey struct{}

func TestCreateContactMessage_ContextPropagation(t *testing.T) {
	var capturedCtx context.Context
	mockRepo := &mockRepository{
		createContactMessageFunc: func(ctx context.Context, _ *models.ContactMessage) error {
			capturedCtx = ctx
			return nil
		},
	}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add middleware that injects a sentinel value into the context
	router.Use(func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), ctxKey{}, "test-marker")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})

	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John Doe","email":"john@example.com","subject":"Test","message":"Hello world"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	if capturedCtx == nil {
		t.Error("expected context to be propagated to repository")
	}

	if capturedCtx.Value(ctxKey{}) != "test-marker" {
		t.Error("context sentinel value was not propagated to repository")
	}
}

func TestGetContactMessages_ContextPropagation(t *testing.T) {
	var capturedCtx context.Context
	mockRepo := &mockRepository{
		getContactMessagesFunc: func(ctx context.Context) ([]models.ContactMessage, error) {
			capturedCtx = ctx
			return []models.ContactMessage{}, nil
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

	router.GET("/api/v1/messages", handler.GetContactMessages)

	w := performRequest(router, http.MethodGet, "/api/v1/messages", nil)

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
// Invalid ID Format Tests (Table-Driven)
// =============================================================================

func TestInvalidIDFormats_Messages(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.GET("/api/v1/messages/:id", handler.GetContactMessage)

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
		t.Run(tt.name, func(t *testing.T) {
			w := performRequest(router, http.MethodGet, "/api/v1/messages/"+tt.id, nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("GetContactMessage(%q) status = %d, want %d", tt.id, w.Code, http.StatusBadRequest)
			}
		})
	}
}

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
