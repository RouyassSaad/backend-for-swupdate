package websocket

import (
	"fmt"
	"net/http"

	log "swupdate/bindings/golang/server/log"
	"time"

	"github.com/gorilla/websocket"
)

var GlobalHub = &Hub{
	managers:   make(map[*WSConnection]bool),
	Register:   make(chan *WSConnection),
	unregister: make(chan *WSConnection),
	broadcast:  make(chan OutgoingMessage),
	contact:    make(chan OutgoingMessage),
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins (i will change it later)
		return true
	},
}

func (h *Hub) Run() {
	for {
		select {
		case manager := <-h.Register:
			h.mu.Lock()

			h.managers[manager] = true
			manager.expires_at = time.Now().Add(24 * time.Hour)

			h.mu.Unlock()
			log.Logger.Info("Registered new socket connection", "total", len(h.managers))
		case manager := <-h.unregister:
			h.mu.Lock()

			if _, ok := h.managers[manager]; ok {
				delete(h.managers, manager)
				log.Logger.Info("Unregistered a socket connection", "total", len(h.managers))
			}

			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.Lock()

			for manager := range h.managers {
				// Send to each manager's outgoing channel
				manager.outgoing <- msg
			}

			h.mu.Unlock()
		case msg := <-h.contact:
			h.mu.Lock()

			for manager := range h.managers {
				if manager.id == msg.Id {
					manager.outgoing <- msg
				}
			}

			h.mu.Unlock()

		}
	}
}

// Message to all the clients
func (h *Hub) Broadcast(msg OutgoingMessage) {
	h.broadcast <- msg
}

// Message to a specific web socket
func (h *Hub) Contact(msg OutgoingMessage) {
	h.contact <- msg
}

func InitConnection(w http.ResponseWriter, r *http.Request, hub *Hub, id string) (*WSConnection, error) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Logger.Error("Failed to upgrade connection", "err", err)
		return nil, err
	}

	manager := &WSConnection{
		id:       id,
		conn:     conn,
		outgoing: make(chan any),
		hub:      hub,
	}
	log.Logger.Info("Connection Accepted", "id", manager.id)

	return manager, nil
}

func (m *WSConnection) Read() {
	defer func() {
		m.conn.Close()
		m.hub.unregister <- m // Unregister when done
		close(m.outgoing)     // Close channel to stop write loop
	}()

	for {
		var msg IncomingMessage
		if err := m.conn.ReadJSON(&msg); err != nil {
			log.Logger.Error("Error While Reading", "err", err)
			return
		}
		log.Logger.Info("JSON Message Received",
			"type", msg.Type,
			"data", msg.Data,
		)
		switch msg.Type {
		case "HELLO":
			log.Logger.Info("hello event", "name", msg.Data)
		default:
			log.Logger.Warn("unknown message type", "type", msg.Type)
		}
	}
}

func (m *WSConnection) Write(msg any) error {

	switch v := msg.(type) {
	case OutgoingMessage:
		if err := m.conn.WriteJSON(msg); err != nil {
			log.Logger.Error("Error While Writing", "err", err)
			return err
		}
		log.Logger.Info("JSON Message Sent",
			"type", v.Type)
		return nil

	case SwupdateMessage:
		if err := m.conn.WriteJSON(msg); err != nil {
			log.Logger.Error("Error While Writing", "err", err)
			return err
		}
		return nil

	case JournalctlLogMessage:
		if err := m.conn.WriteJSON(msg); err != nil {
			log.Logger.Error("Error While Writing", "err", err)
			return err
		}
		return nil
	case *Msg:
		if err := m.conn.WriteJSON(msg); err != nil {
			log.Logger.Error("Error While Writing", "err", err)
			return err
		}
		return nil
	default:
		log.Logger.Error("Unsupported message type", "type", fmt.Sprintf("%T", msg))
		return fmt.Errorf("unsupported message type: %T", msg)

	}

}

func (m *WSConnection) WriteLoop() {
	for msg := range m.outgoing {
		if err := m.Write(msg); err != nil {
			return
		}
	}
}

func (m *WSConnection) ShouldItLive() bool {
	return time.Now().Before(m.expires_at)
}

func (h *Hub) CleanUpRoutine() {
	ticker := time.NewTicker(5 * time.Hour)
	defer ticker.Stop()

	for {
		<-ticker.C

		now := time.Now()

		for conn := range h.managers {
			if now.After(conn.expires_at) {
				h.unregister <- conn
			}
		}
	}
}
