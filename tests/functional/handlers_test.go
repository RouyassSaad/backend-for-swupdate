package tests

import (
	"net/http"
	"net/http/httptest"
	"strings"
	handlers "swupdate/bindings/golang/server/handlers"
	"testing"
)

func TestRootHandlerGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handlers.RootHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "hello from backend") {
		t.Fatalf("expected body to contain 'hello from backend', got: %s", body)
	}

	// Ensure JSON content type
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}
}

func TestRootHandlerOPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	w := httptest.NewRecorder()

	handlers.RootHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204 for OPTIONS, got %d", resp.StatusCode)
	}

	if w.Body.Len() != 0 {
		t.Fatalf("expected empty body for OPTIONS, got: %s", w.Body.String())
	}
}

func TestHealthHandler_OPTIONS(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", resp.StatusCode)
	}

	if w.Body.Len() != 0 {
		t.Fatalf("expected empty body for OPTIONS, got: %s", w.Body.String())
	}
}

func TestHealthHandler_GET(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "backend is up") {
		t.Fatalf("expected body to contain 'backend is up', got: %s", body)
	}

	// Ensure JSON content type
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("expected Content-Type application/json, got %s", ct)
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthHandler(w, req)

	resp := w.Result()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 Method Not Allowed, got %d", resp.StatusCode)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Method not allowed") {
		t.Fatalf("expected error message in body, got: %s", body)
	}
}
