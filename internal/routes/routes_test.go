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
	createContactMessageFunc       func(ctx context.Context, message *models.ContactMessage) error
	getContactMessagesFunc         func(ctx context.Context) ([]models.ContactMessage, error)
	getContactMessageByIDFunc      func(ctx context.Context, id int64) (*models.ContactMessage, error)
	updateContactMessageStatusFunc func(ctx context.Context, id int64, status string, lastError *string) error
	getAllRecipientsFunc           func(ctx context.Context) ([]models.Recipient, error)
	getActiveRecipientsFunc        func(ctx context.Context) ([]models.Recipient, error)
	getRecipientByIDFunc           func(ctx context.Context, id int64) (*models.Recipient, error)
	createRecipientFunc            func(ctx context.Context, recipient *models.Recipient) error
	updateRecipientFunc            func(ctx context.Context, recipient *models.Recipient) error
	deleteRecipientFunc            func(ctx context.Context, id int64) error
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
	return []models.ContactMessage{}, nil
}

func (m *mockRepository) GetContactMessageByID(ctx context.Context, id int64) (*models.ContactMessage, error) {
	if m.getContactMessageByIDFunc != nil {
		return m.getContactMessageByIDFunc(ctx, id)
	}
	return &models.ContactMessage{ID: id}, nil
}

func (m *mockRepository) UpdateContactMessageStatus(ctx context.Context, id int64, status string, lastError *string) error {
	if m.updateContactMessageStatusFunc != nil {
		return m.updateContactMessageStatusFunc(ctx, id, status, lastError)
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

func (m *mockPublisher) Publish(ctx context.Context, message interface{}) error {
	return nil
}

func (m *mockPublisher) PublishToRetry(ctx context.Context, retryIndex int, body []byte, correlationId string, headers amqp.Table) error {
	return nil
}

func (m *mockPublisher) PublishToDLQ(ctx context.Context, body []byte, correlationId string) error {
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
		// Messages (read-only)
		messages := v1.Group("/messages")
		{
			messages.GET("", common.RequirePermission(common.ResourceMessages, common.LevelRead), handler.GetContactMessages)
			messages.GET("/:id", common.RequirePermission(common.ResourceMessages, common.LevelRead), handler.GetContactMessage)
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
// Messages Route Permission Tests
// =============================================================================

func TestMessagesRoutes_Forbidden_WithoutPermission(t *testing.T) {
	for _, route := range messagesRoutes {
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

func TestMessagesRoutes_Allowed_WithPermission(t *testing.T) {
	for _, route := range messagesRoutes {
		t.Run(route.method+" "+route.path, func(t *testing.T) {
			scopes := map[string]string{route.resource: route.level}
			router := setupRouterWithScopes(t, scopes)
			w := performRequest(t, router, route.method, route.path)

			// We only verify authorization passes (not 403/401).
			// Handler may return 400/404/500 due to missing body or mock defaults.
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

			// We only verify authorization passes (not 403/401).
			// Handler may return 400/404/500 due to missing body or mock defaults.
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
		// Messages hierarchy (read-only resource)
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
	// Messages permission should NOT grant recipients access
	scopes := map[string]string{common.ResourceMessages: common.LevelDelete}
	router := setupRouterWithScopes(t, scopes)

	w := performRequest(t, router, "GET", "/api/v1/recipients")

	if w.Code != http.StatusForbidden {
		t.Errorf("messages:delete should not grant recipients:read, got status %d", w.Code)
	}
}

// =============================================================================
// Middleware Error Handling Tests
// =============================================================================

func TestRoutes_NoScopes_Unauthorized(t *testing.T) {
	router := gin.New()
	handler := handlers.New(&mockRepository{}, &mockPublisher{})

	// Route without scope injection middleware
	router.GET("/api/v1/messages",
		common.RequirePermission(common.ResourceMessages, common.LevelRead),
		handler.GetContactMessages,
	)

	req, _ := http.NewRequest("GET", "/api/v1/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d (no scopes = unauthorized)", w.Code, http.StatusUnauthorized)
	}
}

func TestRoutes_InvalidScopesFormat_InternalError(t *testing.T) {
	router := gin.New()
	handler := handlers.New(&mockRepository{}, &mockPublisher{})

	// Inject invalid scopes format
	router.Use(func(c *gin.Context) {
		c.Set("scopes", "invalid-format")
		c.Next()
	})

	router.GET("/api/v1/messages",
		common.RequirePermission(common.ResourceMessages, common.LevelRead),
		handler.GetContactMessages,
	)

	req, _ := http.NewRequest("GET", "/api/v1/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d (invalid scopes = internal error)", w.Code, http.StatusInternalServerError)
	}
}
