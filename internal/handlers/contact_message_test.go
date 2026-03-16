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
	var createdEmail *models.Email
	var published interface{}
	publishCalled := false

	mockRepo := &mockRepository{
		createEmailFunc: func(_ context.Context, email *models.Email) error {
			createdEmail = email
			email.ID = 1
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

	if createdEmail == nil {
		t.Fatal("expected email to be created")
	}
	if createdEmail.Name == nil || *createdEmail.Name != "John Doe" {
		t.Errorf("expected name 'John Doe', got %v", createdEmail.Name)
	}
	if createdEmail.SenderEmail == nil || *createdEmail.SenderEmail != "john@example.com" {
		t.Errorf("expected sender email 'john@example.com', got %v", createdEmail.SenderEmail)
	}
	if createdEmail.Type != models.EmailTypeContactForm {
		t.Errorf("expected type %q, got %q", models.EmailTypeContactForm, createdEmail.Type)
	}
	if createdEmail.Status != models.EmailStatusPending {
		t.Errorf("expected status 'pending', got %s", createdEmail.Status)
	}

	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message, got %s", w.Body.String())
	}

	if !publishCalled {
		t.Fatal("expected publisher.Publish to be called")
	}
	evt, ok := published.(models.EmailEvent)
	if !ok {
		t.Fatalf("expected EmailEvent, got %T", published)
	}
	if evt.EmailID != createdEmail.ID {
		t.Errorf("expected EmailID %d, got %d", createdEmail.ID, evt.EmailID)
	}
}

func TestCreateContactMessage_SpamDetected(t *testing.T) {
	createCalled := false
	publishCalled := false
	mockRepo := &mockRepository{
		createEmailFunc: func(_ context.Context, _ *models.Email) error {
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

	body := `{"name":"Bot","email":"bot@spam.com","subject":"Spam","message":"Buy now!","website":"http://spam.com"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	if createCalled {
		t.Error("expected message NOT to be saved for spam")
	}
	if publishCalled {
		t.Error("expected publisher.Publish NOT to be called for spam")
	}
	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message even for spam, got %s", w.Body.String())
	}
}

func TestCreateContactMessage_PublishError_DoesNotFailRequest(t *testing.T) {
	mockRepo := &mockRepository{
		createEmailFunc: func(_ context.Context, m *models.Email) error {
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

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d even when publish fails, got %d", http.StatusCreated, w.Code)
	}
	if !strings.Contains(w.Body.String(), "Thank you for your message") {
		t.Errorf("expected success message, got %s", w.Body.String())
	}
}

func TestCreateContactMessage_ValidationError(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateContactMessage_InvalidEmail(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

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
		createEmailFunc: func(_ context.Context, _ *models.Email) error {
			return errors.New("database error")
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"John Doe","email":"john@example.com","subject":"Test","message":"Hello world"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestCreateContactMessage_InvalidJSON(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{invalid json}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateContactMessage_UnicodeAndSpecialCharacters(t *testing.T) {
	var createdEmail *models.Email
	mockRepo := &mockRepository{
		createEmailFunc: func(_ context.Context, email *models.Email) error {
			createdEmail = email
			email.ID = 1
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ any) error { return nil },
	}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"日本語 Ñoño 🎉","email":"test@example.com","subject":"Ümläut & Special <chars>","message":"Hello 世界! Here are some special chars: <html> tags & ampersands"}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	if createdEmail == nil {
		t.Fatal("expected email to be created")
	}
	if createdEmail.Name == nil || *createdEmail.Name != "日本語 Ñoño 🎉" {
		t.Errorf("expected Unicode name preserved, got %v", createdEmail.Name)
	}
}

func TestCreateContactMessage_MaxLengthInputs(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

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
			createCalled := false
			testRepo := &mockRepository{
				createEmailFunc: func(_ context.Context, m *models.Email) error {
					createCalled = true
					m.ID = 1
					return nil
				},
			}
			testPub := &mockPublisher{
				publishFunc: func(_ context.Context, _ any) error { return nil },
			}
			testHandler := New(testRepo, testPub)
			testRouter := setupTestRouter()
			testRouter.POST("/api/v1/contact", testHandler.CreateContactMessage)

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
		createEmailFunc: func(_ context.Context, _ *models.Email) error {
			createCalled = true
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ any) error { return nil },
	}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/contact", handler.CreateContactMessage)

	body := `{"name":"Real User","email":"real@example.com","subject":"Real Subject","message":"Real message","website":""}`
	w := performRequest(router, http.MethodPost, "/api/v1/contact", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	if !createCalled {
		t.Error("expected message to be saved for empty honeypot")
	}
}

func TestCreateContactMessage_MaliciousContent(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var createdEmail *models.Email
			mockRepo := &mockRepository{
				createEmailFunc: func(_ context.Context, email *models.Email) error {
					createdEmail = email
					email.ID = 1
					return nil
				},
			}
			mockPub := &mockPublisher{
				publishFunc: func(_ context.Context, _ any) error { return nil },
			}
			handler := New(mockRepo, mockPub)

			router := setupTestRouter()
			router.POST("/api/v1/contact", handler.CreateContactMessage)

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

			if createdEmail == nil {
				t.Fatal("expected email to be created")
			}
			if createdEmail.Name == nil || *createdEmail.Name != tt.expectedName {
				t.Errorf("name: expected %q, got %v", tt.expectedName, createdEmail.Name)
			}
			if createdEmail.Subject != tt.expectedSubject {
				t.Errorf("subject: expected %q, got %q", tt.expectedSubject, createdEmail.Subject)
			}
			if createdEmail.Message != tt.expectedMessage {
				t.Errorf("message: expected %q, got %q", tt.expectedMessage, createdEmail.Message)
			}
		})
	}
}

// =============================================================================
// GetEmails Tests
// =============================================================================

func TestGetEmails_Success(t *testing.T) {
	expected := createTestEmails()
	mockRepo := &mockRepository{
		getEmailsFunc: func(_ context.Context) ([]models.Email, error) {
			return expected, nil
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails", handler.GetEmails)

	w := performRequest(router, http.MethodGet, "/api/v1/emails", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.Email
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(result) != len(expected) {
		t.Errorf("expected %d emails, got %d", len(expected), len(result))
	}
}

func TestGetEmails_Empty(t *testing.T) {
	mockRepo := &mockRepository{
		getEmailsFunc: func(_ context.Context) ([]models.Email, error) {
			return []models.Email{}, nil
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails", handler.GetEmails)

	w := performRequest(router, http.MethodGet, "/api/v1/emails", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result []models.Email
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 emails, got %d", len(result))
	}
}

func TestGetEmails_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		getEmailsFunc: func(_ context.Context) ([]models.Email, error) {
			return nil, errors.New("database error")
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails", handler.GetEmails)

	w := performRequest(router, http.MethodGet, "/api/v1/emails", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// GetEmail Tests
// =============================================================================

func TestGetEmail_Success(t *testing.T) {
	expected := createTestEmail()
	mockRepo := &mockRepository{
		getEmailByIDFunc: func(_ context.Context, id int64) (*models.Email, error) {
			if id != 1 {
				t.Errorf("expected id 1, got %d", id)
			}
			return expected, nil
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails/:id", handler.GetEmail)

	w := performRequest(router, http.MethodGet, "/api/v1/emails/1", nil)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result models.Email
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result.ID != expected.ID {
		t.Errorf("expected ID %d, got %d", expected.ID, result.ID)
	}
}

func TestGetEmail_InvalidID(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails/:id", handler.GetEmail)

	w := performRequest(router, http.MethodGet, "/api/v1/emails/invalid", nil)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "Invalid ID format") {
		t.Errorf("expected 'Invalid ID format' error, got %s", w.Body.String())
	}
}

func TestGetEmail_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getEmailByIDFunc: func(_ context.Context, _ int64) (*models.Email, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails/:id", handler.GetEmail)

	w := performRequest(router, http.MethodGet, "/api/v1/emails/999", nil)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetEmail_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		getEmailByIDFunc: func(_ context.Context, _ int64) (*models.Email, error) {
			return nil, errors.New("database error")
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails/:id", handler.GetEmail)

	w := performRequest(router, http.MethodGet, "/api/v1/emails/1", nil)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// SendEmail Tests
// =============================================================================

func TestSendEmail_Success(t *testing.T) {
	var createdEmail *models.Email
	publishCalled := false

	mockRepo := &mockRepository{
		createEmailFunc: func(_ context.Context, email *models.Email) error {
			createdEmail = email
			email.ID = 10
			return nil
		},
	}
	mockPub := &mockPublisher{
		publishFunc: func(_ context.Context, _ interface{}) error {
			publishCalled = true
			return nil
		},
	}
	handler := New(mockRepo, mockPub)

	router := setupTestRouter()
	router.POST("/api/v1/emails", handler.SendEmail)

	body := `{"type":"email_verification","recipient_email":"user@example.com","data":{"username":"testuser","verify_url":"https://example.com/verify?token=abc"}}`
	w := performRequest(router, http.MethodPost, "/api/v1/emails", strings.NewReader(body))

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	if createdEmail == nil {
		t.Fatal("expected email to be created")
	}
	if createdEmail.Type != models.EmailTypeEmailVerification {
		t.Errorf("expected type %q, got %q", models.EmailTypeEmailVerification, createdEmail.Type)
	}
	if createdEmail.RecipientEmail == nil || *createdEmail.RecipientEmail != "user@example.com" {
		t.Errorf("expected recipient user@example.com, got %v", createdEmail.RecipientEmail)
	}
	if createdEmail.Subject != "Verify your email address" {
		t.Errorf("expected subject 'Verify your email address', got %q", createdEmail.Subject)
	}
	if !strings.Contains(createdEmail.Message, "testuser") {
		t.Error("expected rendered template to contain username")
	}
	if !strings.Contains(createdEmail.Message, "https://example.com/verify?token=abc") {
		t.Error("expected rendered template to contain verify_url")
	}
	if !publishCalled {
		t.Error("expected publisher.Publish to be called")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["id"] != float64(10) {
		t.Errorf("expected id 10 in response, got %v", resp["id"])
	}
}

func TestSendEmail_UnsupportedType(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

	router := setupTestRouter()
	router.POST("/api/v1/emails", handler.SendEmail)

	body := `{"type":"unknown_type","recipient_email":"user@example.com","data":{"key":"val"}}`
	w := performRequest(router, http.MethodPost, "/api/v1/emails", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "unsupported email type") {
		t.Errorf("expected unsupported type error, got %s", w.Body.String())
	}
}

func TestSendEmail_MissingFields(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

	router := setupTestRouter()
	router.POST("/api/v1/emails", handler.SendEmail)

	body := `{"type":"email_verification"}`
	w := performRequest(router, http.MethodPost, "/api/v1/emails", strings.NewReader(body))

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSendEmail_RepositoryError(t *testing.T) {
	mockRepo := &mockRepository{
		createEmailFunc: func(_ context.Context, _ *models.Email) error {
			return errors.New("database error")
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	router := setupTestRouter()
	router.POST("/api/v1/emails", handler.SendEmail)

	body := `{"type":"email_verification","recipient_email":"user@example.com","data":{"username":"test","verify_url":"https://example.com/verify"}}`
	w := performRequest(router, http.MethodPost, "/api/v1/emails", strings.NewReader(body))

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// =============================================================================
// Context Propagation Tests
// =============================================================================

func TestCreateContactMessage_ContextPropagation(t *testing.T) {
	var capturedCtx context.Context
	mockRepo := &mockRepository{
		createEmailFunc: func(ctx context.Context, _ *models.Email) error {
			capturedCtx = ctx
			return nil
		},
	}
	handler := New(mockRepo, &mockPublisher{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
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

// =============================================================================
// Invalid ID Format Tests
// =============================================================================

func TestInvalidIDFormats_Emails(t *testing.T) {
	handler := New(&mockRepository{}, &mockPublisher{})

	router := setupTestRouter()
	router.GET("/api/v1/emails/:id", handler.GetEmail)

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
			w := performRequest(router, http.MethodGet, "/api/v1/emails/"+tt.id, nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("GetEmail(%q) status = %d, want %d", tt.id, w.Code, http.StatusBadRequest)
			}
		})
	}
}
