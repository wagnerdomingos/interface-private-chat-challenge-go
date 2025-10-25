package sockets

import (
	"encoding/json"
	"log"
	"time"

	"messaging-app/domain"

	"github.com/gorilla/websocket"
)

func (c *Client) StartWriter() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Channel closed - send close message
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			// Send ping to keep connection alive
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) StartReader(hub *ConnectionHub) {
	defer func() {
		hub.UnregisterClient(c)
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming WebSocket messages
		var msg struct {
			Type      string `json:"type"`
			MessageID string `json:"message_id"`
		}

		if err := json.Unmarshal(message, &msg); err == nil {
			if msg.Type == "mark_read" {
				hub.MessageSvc.UpdateMessageStatus(msg.MessageID, domain.StatusRead)
			}
		}
	}
}
