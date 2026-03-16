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
	amqp "github.com/rabbitmq/amqp091-go"
)

// =============================================================================
// Mock Repository
// =============================================================================

type mockRepository struct {
	createEmailFunc         func(ctx context.Context, email *models.Email) error
	getEmailsFunc           func(ctx context.Context) ([]models.Email, error)
	getEmailByIDFunc        func(ctx context.Context, id int64) (*models.Email, error)
	updateEmailStatusFunc   func(ctx context.Context, id int64, status string, lastError *string) error
	getAllRecipientsFunc    func(ctx context.Context) ([]models.Recipient, error)
	getActiveRecipientsFunc func(ctx context.Context) ([]models.Recipient, error)
	getRecipientByIDFunc    func(ctx context.Context, id int64) (*models.Recipient, error)
	createRecipientFunc     func(ctx context.Context, recipient *models.Recipient) error
	updateRecipientFunc     func(ctx context.Context, recipient *models.Recipient) error
	deleteRecipientFunc     func(ctx context.Context, id int64) error
}

func (m *mockRepository) CreateEmail(ctx context.Context, email *models.Email) error {
	if m.createEmailFunc != nil {
		return m.createEmailFunc(ctx, email)
	}
	return nil
}

func (m *mockRepository) GetEmails(ctx context.Context) ([]models.Email, error) {
	if m.getEmailsFunc != nil {
		return m.getEmailsFunc(ctx)
	}
	return nil, nil
}

func (m *mockRepository) GetEmailByID(ctx context.Context, id int64) (*models.Email, error) {
	if m.getEmailByIDFunc != nil {
		return m.getEmailByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) UpdateEmailStatus(ctx context.Context, id int64, status string, lastError *string) error {
	if m.updateEmailStatusFunc != nil {
		return m.updateEmailStatusFunc(ctx, id, status, lastError)
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
	publishToRetryFunc func(ctx context.Context, retryIndex int, body []byte, correlationId string, headers amqp.Table) error
	publishToDLQFunc   func(ctx context.Context, body []byte, correlationId string) error
	maxRetries         int
}

func (m *mockPublisher) Publish(ctx context.Context, message interface{}) error {
	if m.publishFunc != nil {
		return m.publishFunc(ctx, message)
	}
	return nil
}

func (m *mockPublisher) PublishToRetry(ctx context.Context, retryIndex int, body []byte, correlationId string, headers amqp.Table) error {
	if m.publishToRetryFunc != nil {
		return m.publishToRetryFunc(ctx, retryIndex, body, correlationId, headers)
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

func strPtr(s string) *string {
	return &s
}

func createTestEmail() *models.Email {
	return &models.Email{
		ID:          1,
		Type:        models.EmailTypeContactForm,
		Name:        strPtr("John Doe"),
		SenderEmail: strPtr("john@example.com"),
		Subject:     "Test Subject",
		Message:     "Test message content",
		Status:      models.EmailStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func createTestEmails() []models.Email {
	return []models.Email{
		{
			ID:          1,
			Type:        models.EmailTypeContactForm,
			Name:        strPtr("John Doe"),
			SenderEmail: strPtr("john@example.com"),
			Subject:     "Test Subject 1",
			Message:     "Test message 1",
			Status:      models.EmailStatusPending,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          2,
			Type:        models.EmailTypeContactForm,
			Name:        strPtr("Jane Smith"),
			SenderEmail: strPtr("jane@example.com"),
			Subject:     "Test Subject 2",
			Message:     "Test message 2",
			Status:      models.EmailStatusSent,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
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
