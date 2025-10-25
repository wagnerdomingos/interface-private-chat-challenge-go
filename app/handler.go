package app

import (
	"encoding/json"
	"net/http"
	"strconv"

	"messaging-app/domain"
	"messaging-app/sockets"

	"github.com/gorilla/mux"
)

// HTTP handler methods for the App
func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" {
		writeError(w, http.StatusBadRequest, "Username is required")
		return
	}

	user := &domain.User{
		Username: req.Username,
	}

	if err := a.userRepo.Create(user); err != nil {
		if err == domain.ErrUsernameExists {
			writeError(w, http.StatusConflict, "Username already exists")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	writeJSON(w, http.StatusCreated, user)
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	user, err := a.userRepo.FindByID(userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			writeError(w, http.StatusNotFound, "User not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get user")
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (a *App) sendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SenderID       string `json:"sender_id"`
		RecipientID    string `json:"recipient_id"`
		Content        string `json:"content"`
		IdempotencyKey string `json:"idempotency_key,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	message, err := a.messageSvc.SendMessage(req.SenderID, req.RecipientID, req.Content, req.IdempotencyKey)
	if err != nil {
		switch err {
		case domain.ErrInvalidUser, domain.ErrCannotMessageSelf, domain.ErrEmptyMessage:
			writeError(w, http.StatusBadRequest, err.Error())
		case domain.ErrUserNotFound:
			writeError(w, http.StatusNotFound, "User not found")
		default:
			writeError(w, http.StatusInternalServerError, "Failed to send message")
		}
		return
	}

	// Broadcast to recipient
	a.hub.BroadcastMessage(message, req.RecipientID)

	writeJSON(w, http.StatusCreated, message)
}

func (a *App) listUserChats(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id parameter is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	response, err := a.messageSvc.GetUserChats(userID, page, pageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get chats")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *App) listChatMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chatId"]

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	response, err := a.messageSvc.GetChatMessages(chatID, page, pageSize)
	if err != nil {
		if err == domain.ErrChatNotFound {
			writeError(w, http.StatusNotFound, "Chat not found")
		} else {
			writeError(w, http.StatusInternalServerError, "Failed to get messages")
		}
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id parameter is required")
		return
	}

	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Log error but don't write response as Upgrade may have already written headers
		return
	}

	client := &sockets.Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	a.hub.RegisterClient(client)

	// Start goroutines for reading and writing
	go client.StartWriter()
	go client.StartReader(a.hub)
}

func (a *App) healthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
