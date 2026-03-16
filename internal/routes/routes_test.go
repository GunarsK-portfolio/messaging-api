package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GunarsK-portfolio/messaging-api/internal/handlers"
	common "github.com/GunarsK-portfolio/portfolio-common/middleware"
	"github.com/GunarsK-portfolio/portfolio-common/models"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func init() {
	gin.SetMode(gin.TestMode)
}

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
	return []models.Email{}, nil
}

func (m *mockRepository) GetEmailByID(ctx context.Context, id int64) (*models.Email, error) {
	if m.getEmailByIDFunc != nil {
		return m.getEmailByIDFunc(ctx, id)
	}
	return &models.Email{ID: id}, nil
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
	return []models.Recipient{}, nil
}

func (m *mockRepository) GetActiveRecipients(ctx context.Context) ([]models.Recipient, error) {
	if m.getActiveRecipientsFunc != nil {
		return m.getActiveRecipientsFunc(ctx)
	}
	return []models.Recipient{}, nil
}

func (m *mockRepository) GetRecipientByID(ctx context.Context, id int64) (*models.Recipient, error) {
	if m.getRecipientByIDFunc != nil {
		return m.getRecipientByIDFunc(ctx, id)
	}
	return &models.Recipient{ID: id}, nil
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

// =============================================================================
// Mock Publisher
// =============================================================================

type mockPublisher struct{}

func (m *mockPublisher) Publish(_ context.Context, _ interface{}) error {
	return nil
}

func (m *mockPublisher) PublishToRetry(_ context.Context, _ int, _ []byte, _ string, _ amqp.Table) error {
	return nil
}

func (m *mockPublisher) PublishToDLQ(_ context.Context, _ []byte, _ string) error {
	return nil
}

func (m *mockPublisher) MaxRetries() int {
	return 3
}

func (m *mockPublisher) Close() error {
	return nil
}

// =============================================================================
// Test Helpers
// =============================================================================

func injectScopes(scopes map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("scopes", scopes)
		c.Next()
	}
}

func setupRouterWithScopes(t *testing.T, scopes map[string]string) *gin.Engine {
	t.Helper()

	router := gin.New()
	handler := handlers.New(&mockRepository{}, &mockPublisher{})

	v1 := router.Group("/api/v1")
	v1.Use(injectScopes(scopes))
	{
		// Emails
		emails := v1.Group("/emails")
		{
			emails.POST("", common.RequirePermission(common.ResourceEmails, common.LevelEdit), handler.SendEmail)
			emails.GET("", common.RequirePermission(common.ResourceEmails, common.LevelRead), handler.GetEmails)
			emails.GET("/:id", common.RequirePermission(common.ResourceEmails, common.LevelRead), handler.GetEmail)
		}

		// Legacy messages route
		messages := v1.Group("/messages")
		{
			messages.GET("", common.RequirePermission(common.ResourceMessages, common.LevelRead), handler.GetEmails)
			messages.GET("/:id", common.RequirePermission(common.ResourceMessages, common.LevelRead), handler.GetEmail)
		}

		// Recipients (full CRUD)
		recipients := v1.Group("/recipients")
		{
			recipients.GET("", common.RequirePermission(common.ResourceRecipients, common.LevelRead), handler.GetRecipients)
			recipients.GET("/:id", common.RequirePermission(common.ResourceRecipients, common.LevelRead), handler.GetRecipient)
			recipients.POST("", common.RequirePermission(common.ResourceRecipients, common.LevelEdit), handler.CreateRecipient)
			recipients.PUT("/:id", common.RequirePermission(common.ResourceRecipients, common.LevelEdit), handler.UpdateRecipient)
			recipients.DELETE("/:id", common.RequirePermission(common.ResourceRecipients, common.LevelDelete), handler.DeleteRecipient)
		}
	}

	return router
}

func performRequest(t *testing.T, router *gin.Engine, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// =============================================================================
// Route Permission Definitions
// =============================================================================

type routePermission struct {
	method   string
	path     string
	resource string
	level    string
}

var emailsRoutes = []routePermission{
	{"GET", "/api/v1/emails", common.ResourceEmails, common.LevelRead},
	{"GET", "/api/v1/emails/1", common.ResourceEmails, common.LevelRead},
}

var messagesRoutes = []routePermission{
	{"GET", "/api/v1/messages", common.ResourceMessages, common.LevelRead},
	{"GET", "/api/v1/messages/1", common.ResourceMessages, common.LevelRead},
}

var recipientsRoutes = []routePermission{
	{"GET", "/api/v1/recipients", common.ResourceRecipients, common.LevelRead},
	{"GET", "/api/v1/recipients/1", common.ResourceRecipients, common.LevelRead},
	{"POST", "/api/v1/recipients", common.ResourceRecipients, common.LevelEdit},
	{"PUT", "/api/v1/recipients/1", common.ResourceRecipients, common.LevelEdit},
	{"DELETE", "/api/v1/recipients/1", common.ResourceRecipients, common.LevelDelete},
}

// =============================================================================
// Emails Route Permission Tests
// =============================================================================

func TestEmailsRoutes_Forbidden_WithoutPermission(t *testing.T) {
	for _, route := range emailsRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			router := setupRouterWithScopes(t, map[string]string{})
			w := performRequest(t, router, route.method, route.path)

			if w.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}
			if response["error"] != "insufficient permissions" {
				t.Errorf("error = %v, want 'insufficient permissions'", response["error"])
			}
			if response["resource"] != route.resource {
				t.Errorf("resource = %v, want %q", response["resource"], route.resource)
			}
		})
	}
}

func TestEmailsRoutes_Allowed_WithPermission(t *testing.T) {
	for _, route := range emailsRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			scopes := map[string]string{route.resource: route.level}
			router := setupRouterWithScopes(t, scopes)
			w := performRequest(t, router, route.method, route.path)

			if w.Code == http.StatusForbidden {
				t.Errorf("got 403 Forbidden with permission %s:%s", route.resource, route.level)
			}
			if w.Code == http.StatusUnauthorized {
				t.Errorf("got 401 Unauthorized - scopes not injected")
			}
		})
	}
}

// =============================================================================
// Legacy Messages Route Permission Tests
// =============================================================================

func TestMessagesRoutes_Forbidden_WithoutPermission(t *testing.T) {
	for _, route := range messagesRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			router := setupRouterWithScopes(t, map[string]string{})
			w := performRequest(t, router, route.method, route.path)

			if w.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
			}
		})
	}
}

func TestMessagesRoutes_Allowed_WithPermission(t *testing.T) {
	for _, route := range messagesRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			scopes := map[string]string{route.resource: route.level}
			router := setupRouterWithScopes(t, scopes)
			w := performRequest(t, router, route.method, route.path)

			if w.Code == http.StatusForbidden {
				t.Errorf("got 403 Forbidden with permission %s:%s", route.resource, route.level)
			}
		})
	}
}

// =============================================================================
// Recipients Route Permission Tests
// =============================================================================

func TestRecipientsRoutes_Forbidden_WithoutPermission(t *testing.T) {
	for _, route := range recipientsRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			router := setupRouterWithScopes(t, map[string]string{})
			w := performRequest(t, router, route.method, route.path)

			if w.Code != http.StatusForbidden {
				t.Errorf("status = %d, want %d", w.Code, http.StatusForbidden)
			}
		})
	}
}

func TestRecipientsRoutes_Allowed_WithPermission(t *testing.T) {
	for _, route := range recipientsRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			scopes := map[string]string{route.resource: route.level}
			router := setupRouterWithScopes(t, scopes)
			w := performRequest(t, router, route.method, route.path)

			if w.Code == http.StatusForbidden {
				t.Errorf("got 403 Forbidden with permission %s:%s", route.resource, route.level)
			}
		})
	}
}

// =============================================================================
// Permission Hierarchy Tests
// =============================================================================

func TestPermissionHierarchy(t *testing.T) {
	tests := []struct {
		name       string
		resource   string
		granted    string
		required   string
		method     string
		path       string
		wantAccess bool
	}{
		// Recipients hierarchy
		{"delete grants delete", common.ResourceRecipients, common.LevelDelete, common.LevelDelete, "DELETE", "/api/v1/recipients/1", true},
		{"delete grants edit", common.ResourceRecipients, common.LevelDelete, common.LevelEdit, "POST", "/api/v1/recipients", true},
		{"delete grants read", common.ResourceRecipients, common.LevelDelete, common.LevelRead, "GET", "/api/v1/recipients", true},
		{"edit grants edit", common.ResourceRecipients, common.LevelEdit, common.LevelEdit, "POST", "/api/v1/recipients", true},
		{"edit grants read", common.ResourceRecipients, common.LevelEdit, common.LevelRead, "GET", "/api/v1/recipients", true},
		{"edit denies delete", common.ResourceRecipients, common.LevelEdit, common.LevelDelete, "DELETE", "/api/v1/recipients/1", false},
		{"read grants read", common.ResourceRecipients, common.LevelRead, common.LevelRead, "GET", "/api/v1/recipients", true},
		{"read denies edit", common.ResourceRecipients, common.LevelRead, common.LevelEdit, "POST", "/api/v1/recipients", false},
		{"none denies read", common.ResourceRecipients, common.LevelNone, common.LevelRead, "GET", "/api/v1/recipients", false},
		// Emails hierarchy
		{"emails read grants read", common.ResourceEmails, common.LevelRead, common.LevelRead, "GET", "/api/v1/emails", true},
		{"emails delete grants read", common.ResourceEmails, common.LevelDelete, common.LevelRead, "GET", "/api/v1/emails", true},
		// Legacy messages
		{"messages read grants read", common.ResourceMessages, common.LevelRead, common.LevelRead, "GET", "/api/v1/messages", true},
		{"messages delete grants read", common.ResourceMessages, common.LevelDelete, common.LevelRead, "GET", "/api/v1/messages", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scopes := map[string]string{tt.resource: tt.granted}
			router := setupRouterWithScopes(t, scopes)
			w := performRequest(t, router, tt.method, tt.path)

			gotAccess := w.Code != http.StatusForbidden
			if gotAccess != tt.wantAccess {
				t.Errorf("granted=%s required=%s: gotAccess=%v wantAccess=%v (status=%d)",
					tt.granted, tt.required, gotAccess, tt.wantAccess, w.Code)
			}
		})
	}
}

// =============================================================================
// Cross-Resource Permission Tests
// =============================================================================

func TestCrossResourcePermissions_Denied(t *testing.T) {
	scopes := map[string]string{common.ResourceEmails: common.LevelDelete}
	router := setupRouterWithScopes(t, scopes)

	w := performRequest(t, router, "GET", "/api/v1/recipients")

	if w.Code != http.StatusForbidden {
		t.Errorf("emails:delete should not grant recipients:read, got status %d", w.Code)
	}
}

// =============================================================================
// Middleware Error Handling Tests
// =============================================================================

func TestRoutes_NoScopes_Unauthorized(t *testing.T) {
	router := gin.New()
	handler := handlers.New(&mockRepository{}, &mockPublisher{})

	router.GET("/api/v1/emails",
		common.RequirePermission(common.ResourceEmails, common.LevelRead),
		handler.GetEmails,
	)

	req, _ := http.NewRequest("GET", "/api/v1/emails", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d (no scopes = unauthorized)", w.Code, http.StatusUnauthorized)
	}
}

func TestRoutes_InvalidScopesFormat_InternalError(t *testing.T) {
	router := gin.New()
	handler := handlers.New(&mockRepository{}, &mockPublisher{})

	router.Use(func(c *gin.Context) {
		c.Set("scopes", "invalid-format")
		c.Next()
	})

	router.GET("/api/v1/emails",
		common.RequirePermission(common.ResourceEmails, common.LevelRead),
		handler.GetEmails,
	)

	req, _ := http.NewRequest("GET", "/api/v1/emails", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (invalid scopes = internal error)", w.Code, http.StatusInternalServerError)
	}
}
