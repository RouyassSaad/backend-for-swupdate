package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type JSONResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func SetEssentialHeaders(w http.ResponseWriter, r *http.Request) {
	// Content & representation
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Caching
	w.Header().Set("Cache-Control", "no-store")

	// Security headers
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

	// CORS (adjust origin for your frontend)
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
}

func SendJSON(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)

	resp := JSONResponse{
		Status:  status,
		Message: message,
	}

	json.NewEncoder(w).Encode(resp)
}

func OnlyMethodsAllowed(w http.ResponseWriter, r *http.Request, methods ...string) bool {
	for _, m := range methods {
		if r.Method == m {
			return true
		}
	}

	SendJSON(w, http.StatusMethodNotAllowed, "Method not allowed")
	return false
}

/*
cookie, err := r.Cookie("ws-connexion-id")
		if err != nil || cookie.Value == "" {
			log.Logger.Error("Missing ws-connexion-id cookie")
			http.Error(w, "No id received, try reconnecting", http.StatusBadRequest)
			return
		}
		webSocketId := cookie.Value
		log.Logger.Info("Client ws id received", "id", webSocketId)
*/

func GetConnectionId(w http.ResponseWriter, r *http.Request) (string, error) {
	cookie, err := r.Cookie("ws-connexion-id")
	if err != nil || cookie.Value == "" {
		http.Error(w, "No id received, try reconnecting", http.StatusBadRequest)
		return "", fmt.Errorf("missing websocket connection id")
	}

	return cookie.Value, nil
}
