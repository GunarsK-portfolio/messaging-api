package handlers

import (
	"testing"
)

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
