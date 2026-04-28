package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"

	//"os"
	//"path/filepath"
	"testing"

	handlers "swupdate/bindings/golang/server/handlers"
	tp "swupdate/bindings/golang/server/models"
)

func createMultipartFile(t *testing.T, fieldName, fileName, content string) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, fileName)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}

	_, err = io.WriteString(part, content)
	if err != nil {
		t.Fatalf("failed to write file content: %v", err)
	}

	writer.Close()
	return body, writer.FormDataContentType()
}

func TestFileUploadHandler_MethodNotAllowed(t *testing.T) {
	handler := handlers.FileUploadHandler(make(chan *tp.UpdateChanel, 1))

	req := httptest.NewRequest(http.MethodGet, "/upload", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestFileUploadHandler_MissingCookie(t *testing.T) {
	handler := handlers.FileUploadHandler(make(chan *tp.UpdateChanel, 1))

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestFileUploadHandler_MissingFile(t *testing.T) {
	handler := handlers.FileUploadHandler(make(chan *tp.UpdateChanel, 1))

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	req.AddCookie(&http.Cookie{Name: "ws-connexion-id", Value: "abc123"})

	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
