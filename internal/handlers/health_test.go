package handlers

import (
	"net/http"
	"strings"
	"testing"
)

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
