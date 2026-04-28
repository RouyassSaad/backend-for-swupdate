package handlers

import (
	"encoding/json"
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
