package domain

import (
	"time"
)

// User represents an application user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// Chat represents a 1:1 conversation between two users
type Chat struct {
	ID           string    `json:"id"`
	Participant1 string    `json:"participant1"` // User ID
	Participant2 string    `json:"participant2"` // User ID
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Message represents a single message in a chat
type Message struct {
	ID             string        `json:"id"`
	ChatID         string        `json:"chat_id"`
	SenderID       string        `json:"sender_id"`
	Content        string        `json:"content"`
	Status         MessageStatus `json:"status"`
	Timestamp      time.Time     `json:"timestamp"`
	IdempotencyKey string        `json:"idempotency_key,omitempty"`
}

// MessageStatus represents the delivery status of a message
type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)
