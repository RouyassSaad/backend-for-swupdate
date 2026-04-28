package websocket

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSConnection struct {
	id         string
	expires_at time.Time
	conn       *websocket.Conn
	outgoing   chan any
	hub        *Hub // Reference to hub for unregister
}

// Hub has all WSConnection Objects
type Hub struct {
	managers   map[*WSConnection]bool // Map of active connections
	Register   chan *WSConnection
	unregister chan *WSConnection
	broadcast  chan OutgoingMessage // Channel for broadcasting to all
	contact    chan OutgoingMessage //contact one specific webSocket
	mu         sync.Mutex           // Protects the map
}

type IncomingMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type OutgoingMessage struct {
	Type      string `json:"type"`
	TimeStamp string `json:"timestamp"`
	Error     error  `json:"error"`
	Id        string `json:"id"`
	Data      any    `json:"data"`
}

type SwupdateMessage struct {
	Time            string `json:"time"`
	Status          int    `json:"status"`
	Image           string `json:"image"`
	Handler         string `json:"handler"`
	Info            string `json:"info"`
	TotalBytes      int    `json:"total_bytes"`
	TotalSteps      int    `json:"total_steps"`
	DownloadedBytes int    `json:"downloaded_bytes"`
	CurrentStep     int    `json:"current_step"`
	CurrentPercent  int    `json:"current_percent"`
}

type JournalctlLogMessage struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Msg    string `json:"message"`
}

type Msg struct {
	Id  string          `json:"id"`
	Msg SwupdateMessage `json:"msg"`
}
