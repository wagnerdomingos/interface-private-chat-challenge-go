package sockets

import (
	"encoding/json"
	"log"
	"sync"

	"messaging-app/domain"
	"messaging-app/services"

	"github.com/gorilla/websocket"
)

// Client represents a WebSocket connection for a user
type Client struct {
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
}

// ConnectionHub manages WebSocket connections and message broadcasting
type ConnectionHub struct {
	Clients    map[string]*Client // userID -> Client
	Broadcast  chan *BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	Mutex      sync.RWMutex
	MessageSvc *services.MessageService
}

// BroadcastMessage contains both the message and recipient information
type BroadcastMessage struct {
	Message     *domain.Message
	RecipientID string
}

// NewConnectionHub creates a new connection hub
func NewConnectionHub(messageSvc *services.MessageService) *ConnectionHub {
	return &ConnectionHub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *BroadcastMessage, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		MessageSvc: messageSvc,
	}
}

// Run starts the hub's main loop
func (h *ConnectionHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			// Disconnect existing client for the same user
			if existing, exists := h.Clients[client.UserID]; exists {
				close(existing.Send)
				delete(h.Clients, client.UserID)
			}
			h.Clients[client.UserID] = client
			h.Mutex.Unlock()

			log.Printf("Client registered: %s", client.UserID)

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if existing, exists := h.Clients[client.UserID]; exists && existing == client {
				close(client.Send)
				delete(h.Clients, client.UserID)
				log.Printf("Client unregistered: %s", client.UserID)
			}
			h.Mutex.Unlock()

		case broadcastMsg := <-h.Broadcast:
			h.broadcastMessage(broadcastMsg)
		}
	}
}

// broadcastMessage sends a message to the intended recipient if they're connected
func (h *ConnectionHub) broadcastMessage(broadcastMsg *BroadcastMessage) {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	// Send to recipient if connected
	if client, exists := h.Clients[broadcastMsg.RecipientID]; exists {
		messageJSON, err := json.Marshal(broadcastMsg.Message)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			return
		}

		select {
		case client.Send <- messageJSON:
			// Update message status to delivered
			h.MessageSvc.UpdateMessageStatus(broadcastMsg.Message.ID, domain.StatusDelivered)
		default:
			close(client.Send)
			delete(h.Clients, broadcastMsg.RecipientID)
		}
	}
}

// RegisterClient registers a new WebSocket client
func (h *ConnectionHub) RegisterClient(client *Client) {
	h.Register <- client
}

// UnregisterClient unregisters a WebSocket client
func (h *ConnectionHub) UnregisterClient(client *Client) {
	h.Unregister <- client
}

// BroadcastMessage broadcasts a message to a specific recipient
func (h *ConnectionHub) BroadcastMessage(message *domain.Message, recipientID string) {
	broadcastMsg := &BroadcastMessage{
		Message:     message,
		RecipientID: recipientID,
	}
	h.Broadcast <- broadcastMsg
}
