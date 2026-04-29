package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	helper "swupdate/bindings/golang/helpers"
	log "swupdate/bindings/golang/server/log"
	t "swupdate/bindings/golang/server/models"
	wsmanager "swupdate/bindings/golang/server/websocket"

	"time"
)

// WS /ws
func WsHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check for existing cookie
	cookie, err := r.Cookie("ws-connexion-id")
	var id string

	if err == nil && cookie.Value != "" {
		log.Logger.Info("Found cookie", "cookie", cookie.Value)
		id = cookie.Value
	} else {
		id = helper.GenerateUUID()
		log.Logger.Info("No cookie, generating new id", "id", id)
	}

	// Send cookie back to client
	http.SetCookie(w, &http.Cookie{
		Name:     "ws-connexion-id",
		Value:    id,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode, // important for cross-origin WS
		Expires:  time.Now().Add(24 * time.Hour),
	})

	manager, err := wsmanager.InitConnection(w, r, wsmanager.GlobalHub, id)
	if err != nil {
		log.Logger.Error("Failed to init WS manager", "err", err)
		return
	}

	log.Logger.Info("WebSocket connected", "id", id)

	go manager.Read()
	go manager.WriteLoop()

	wsmanager.GlobalHub.Register <- manager

	select {}
}

// GET /
func RootHandler(w http.ResponseWriter, r *http.Request) {
	SetEssentialHeaders(w, r)

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Logger.Info("GET /")

	SendJSON(w, http.StatusOK, "hello from backend")
}

// GET /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	SetEssentialHeaders(w, r)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodGet {
		SendJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	log.Logger.Info("GET /health")

	SendJSON(w, http.StatusOK, "backend is up")
}

// GET /getUploadedFiles
func GetUploadedFiles(w http.ResponseWriter, r *http.Request) {
	SetEssentialHeaders(w, r)

	// Handle CORS preflight
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Only GET allowed
	if r.Method != http.MethodGet {
		SendJSON(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	dir := "/tmp"
	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Logger.Error("Failed to read directory", "error", err)
		SendJSON(w, http.StatusInternalServerError, "failed to read directory")
		return
	}

	// Collect .swu files
	files := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".swu" {
			files = append(files, entry.Name())
		}
	}

	// Respond with JSON
	if err := json.NewEncoder(w).Encode(files); err != nil {
		log.Logger.Error("Failed to encode JSON", "error", err)
		SendJSON(w, http.StatusInternalServerError, "failed to encode response")
		return
	}

	log.Logger.Info("GET /uploadedFiles", "status", http.StatusOK)
}

// POST /uploadFile
func FileUploadHandler(globalChannel chan *t.UpdateChanel) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		SetEssentialHeaders(w, r)

		// Only POST allowed
		if r.Method != http.MethodPost {
			log.Logger.Error("Method Not Allowed", "method", r.Method)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract WebSocket ID from cookie
		cookie, err := r.Cookie("ws-connexion-id")
		if err != nil || cookie.Value == "" {
			log.Logger.Error("Missing ws-connexion-id cookie")
			http.Error(w, "No id received, try reconnecting", http.StatusBadRequest)
			return
		}
		webSocketId := cookie.Value
		log.Logger.Info("Client ws id received", "id", webSocketId)

		// Timeout context for the upload + notification
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		// Parse multipart form (max 10MB)
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			log.Logger.Error("Bad multipart form", "error", err)
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}

		// Extract file
		file, header, err := r.FormFile("file")
		if err != nil {
			log.Logger.Error("File missing in form", "error", err)
			http.Error(w, "file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Create destination file
		filename := helper.SanitizeFilename(header.Filename)
		dstPath := "/tmp/" + filename

		dst, err := os.Create(dstPath)
		if err != nil {
			log.Logger.Error("Cannot save file", "path", dstPath, "error", err)
			http.Error(w, "cannot save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Track upload progress
		pw := &t.ProgressWriter{
			Id:    webSocketId,
			Total: header.Size,
		}

		reader := io.TeeReader(file, pw)

		// Copy file to disk
		if _, err := io.Copy(dst, reader); err != nil {
			log.Logger.Error("Failed to write file", "error", err)
			http.Error(w, "write failed", http.StatusInternalServerError)
			return
		}

		// Respond to client

		// Notify via WebSocket channel
		update := &t.UpdateChanel{
			ConnexionId: webSocketId,
			Filename:    filename,
		}

		select {
		case globalChannel <- update:
			SendJSON(w, http.StatusCreated, "file uploaded")
			log.Logger.Info("POST /file", "status", http.StatusCreated)
		case <-ctx.Done():
			SendJSON(w, http.StatusRequestTimeout, "file uploaded but swupdate is busy")
			log.Logger.Error("Upload done but swupdate busy", "error", ctx.Err())
		}
	}
}
