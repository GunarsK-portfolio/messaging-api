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
// CreateContactMessage Tests
// =============================================================================

func TestCreateContactMessage_Success(t *testing.T) {
	var createdMessage *models.ContactMessage
	var published interface{}
	publishCalled := false

	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, message *models.ContactMessage) error {
			createdMessage = message
			message.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, msg interface{}) error {
			publishCalled = true
			published = msg
			return nil
		},
	}
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

	if !publishCalled {
		t.Fatal("expected publisher.Publish to be called")
	}
	evt, ok := published.(models.ContactMessageEvent)
	if !ok {
		t.Fatalf("expected ContactMessageEvent, got %T", published)
	}
	if evt.MessageID != createdMessage.ID {
		t.Errorf("expected MessageID %d, got %d", createdMessage.ID, evt.MessageID)
	}
}

func TestCreateContactMessage_SpamDetected(t *testing.T) {
	createCalled := false
	publishCalled := false
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, _ *models.ContactMessage) error {
			createCalled = true
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ any) error {
			publishCalled = true
			return nil
		},
	}
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

	// And should NOT publish to queue
	if publishCalled {
		t.Error("expected publisher.Publish NOT to be called for spam")
	}

	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message even for spam, got %s", w.Body.String())
	}
}

func TestCreateContactMessage_PublishError_DoesNotFailRequest(t *testing.T) {
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, m *models.ContactMessage) error {
			m.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ any) error {
			return errors.New("mq down")
		},
	}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John Doe","email":"john@example.com","subject":"Test","message":"Hello world"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	// Should still return 201 even when publish fails
	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d even when publish fails, got %d", http.StatusCreated, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message, got %s", w.Body.String())
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

func TestCreateContactMessage_UnicodeAndSpecialCharacters(t *testing.T) {
	var createdMessage *models.ContactMessage
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, message *models.ContactMessage) error {
			createdMessage = message
			message.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ any) error {
			return nil
		},
	}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	// Test with Unicode characters (Chinese, emoji, special symbols)
	body := `{"name":"日本語 Ñoño 🎉","email":"test@example.com","subject":"Ümläut & Special <chars>","message":"Hello 世界! Here are some special chars: <html> tags & ampersands"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	if createdMessage == nil {
		t.Fatal("expected message to be created")
	}
	if createdMessage.Name != "日本語 Ñoño 🎉" {
		t.Errorf("expected Unicode name preserved, got %s", createdMessage.Name)
	}
	if createdMessage.Subject != "Ümläut & Special <chars>" {
		t.Errorf("expected special chars in subject preserved, got %s", createdMessage.Subject)
	}
}

func TestCreateContactMessage_MaxLengthInputs(t *testing.T) {
	mockRepo := &mockRepository{}
	mockPub := &mockPublisher{}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	tests := []struct {
		name       string
		field      string
		length     int
		wantStatus int
	}{
		{"name_at_limit", "name", 255, http.StatusCreated},
		{"name_over_limit", "name", 256, http.StatusBadRequest},
		{"subject_at_limit", "subject", 500, http.StatusCreated},
		{"subject_over_limit", "subject", 501, http.StatusBadRequest},
		{"message_at_limit", "message", 10000, http.StatusCreated},
		{"message_over_limit", "message", 10001, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock for each test
			createCalled := false
			testRepo := &mockRepository{
				createContactMessageFunc: func(_ context.Context, m *models.ContactMessage) error {
					createCalled = true
					m.ID = 1
					return nil
				},
			}
			testPub := &mockPublisher{
				publishFunc: func(_ context.Context, _ any) error {
					return nil
				},
			}
			testHandler := New(testRepo, testPub)
			testRouter := setupTestRouter()
			testRouter.POST("/api/v1/contact", testHandler.CreateContactMessage)

			// Generate string of exact length
			longString := strings.Repeat("a", tt.length)

			var body string
			switch tt.field {
			case "name":
				body = `{"name":"` + longString + `","email":"test@example.com","subject":"Test","message":"Hello"}`
			case "subject":
				body = `{"name":"Test","email":"test@example.com","subject":"` + longString + `","message":"Hello"}`
			case "message":
				body = `{"name":"Test","email":"test@example.com","subject":"Test","message":"` + longString + `"}`
			}

			w := performRequest(testRouter, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

			if w.Code != tt.wantStatus {
				t.Errorf("%s: expected status %d, got %d", tt.name, tt.wantStatus, w.Code)
			}

			// Verify DB was called only for valid inputs
			if tt.wantStatus == http.StatusCreated && !createCalled {
				t.Errorf("%s: expected repository to be called", tt.name)
			}
			if tt.wantStatus == http.StatusBadRequest && createCalled {
				t.Errorf("%s: expected repository NOT to be called for invalid input", tt.name)
			}
		})
	}
}

func TestCreateContactMessage_EmptyHoneypotIsNotSpam(t *testing.T) {
	createCalled := false
	mockRepo := &mockRepository{
		createContactMessageFunc: func(_ context.Context, _ *models.ContactMessage) error {
			createCalled = true
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ any) error {
			return nil
		},
	}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	// Empty honeypot field should NOT be spam
	body := `{"name":"Real User","email":"real@example.com","subject":"Real Subject","message":"Real message","website":""}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Should save the message (not spam)
	if !createCalled {
		t.Error("expected message to be saved for empty honeypot")
	}
}

func TestCreateContactMessage_MaliciousContent(t *testing.T) {
	// Test that potentially malicious content is stored as-is
	// GORM parameterizes queries (prevents SQL injection)
	// HTML escaping should happen at render time, not storage
	tests := []struct {
		name            string
		inputName       string
		inputSubject    string
		inputMessage    string
		expectedName    string
		expectedSubject string
		expectedMessage string
	}{
		{
			name:            "sql_injection_attempt",
			inputName:       "Robert'); DROP TABLE users;--",
			inputSubject:    "Test",
			inputMessage:    "Hello",
			expectedName:    "Robert'); DROP TABLE users;--",
			expectedSubject: "Test",
			expectedMessage: "Hello",
		},
		{
			name:            "html_script_injection",
			inputName:       "<script>alert('xss')</script>",
			inputSubject:    "<img src=x onerror=alert('xss')>",
			inputMessage:    "<iframe src='evil.com'></iframe>",
			expectedName:    "<script>alert('xss')</script>",
			expectedSubject: "<img src=x onerror=alert('xss')>",
			expectedMessage: "<iframe src='evil.com'></iframe>",
		},
		{
			name:            "html_email_content",
			inputName:       "User",
			inputSubject:    "Check this <b>important</b> message",
			inputMessage:    "<a href='http://phishing.com'>Click here</a> for prize!",
			expectedName:    "User",
			expectedSubject: "Check this <b>important</b> message",
			expectedMessage: "<a href='http://phishing.com'>Click here</a> for prize!",
		},
		{
			name:            "null_bytes_and_control_chars",
			inputName:       "User\x00Name",
			inputSubject:    "Subject\nwith\nnewlines",
			inputMessage:    "Message\twith\ttabs",
			expectedName:    "User\x00Name",
			expectedSubject: "Subject\nwith\nnewlines",
			expectedMessage: "Message\twith\ttabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var createdMessage *models.ContactMessage
			mockRepo := &mockRepository{
				createContactMessageFunc: func(_ context.Context, message *models.ContactMessage) error {
					createdMessage = message
					message.ID = 1
					return nil
				},
			}
			mockPub := &mockPublisher{
				publishFunc: func(_ context.Context, _ any) error {
					return nil
				},
			}
			handler := New(mockRepo, mockPub)

			router := setupTestRouter()
			router.POST("/api/v1/contact", handler.CreateContactMessage)

			// Use proper JSON encoding to handle special characters
			reqBody := map[string]string{
				"name":    tt.inputName,
				"email":   "test@example.com",
				"subject": tt.inputSubject,
				"message": tt.inputMessage,
			}
			jsonBody, _ := json.Marshal(reqBody)
			w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(string(jsonBody)))

			if w.Code != http.StatusCreated {
				t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
				return
			}

			if createdMessage == nil {
				t.Fatal("expected message to be created")
			}

			// Verify content is stored as-is (not escaped or modified)
			if createdMessage.Name != tt.expectedName {
				t.Errorf("name: expected %q, got %q", tt.expectedName, createdMessage.Name)
			}
			if createdMessage.Subject != tt.expectedSubject {
				t.Errorf("subject: expected %q, got %q", tt.expectedSubject, createdMessage.Subject)
			}
			if createdMessage.Message != tt.expectedMessage {
				t.Errorf("message: expected %q, got %q", tt.expectedMessage, createdMessage.Message)
			}
		})
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
// Context Propagation Tests (Contact Messages)
// =============================================================================

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

// =============================================================================
// Invalid ID Format Tests (Messages)
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
