package handlers

import (
	"context"
	"io"
	"net/http/httptest"
	"time"

	"github.com/GunarsK-portfolio/messaging-api/internal/repository"
	"github.com/GunarsK-portfolio/portfolio-common/models"
	"github.com/GunarsK-portfolio/portfolio-common/queue"
	"github.com/gin-gonic/gin"
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

// Context key for propagation tests
type ctxKey struct{}
