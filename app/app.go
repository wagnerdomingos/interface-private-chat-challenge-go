package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"messaging-app/repositories"
	"messaging-app/services"
	"messaging-app/sockets"
)

// App represents the main application structure
type App struct {
	router     *mux.Router
	upgrader   *websocket.Upgrader
	userRepo   repositories.UserRepository
	chatRepo   repositories.ChatRepository
	messageSvc *services.MessageService
	hub        *sockets.ConnectionHub
}

// NewApp creates and initializes a new App instance
func NewApp() *App {
	app := &App{
		router: mux.NewRouter(),
		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, validate specific origins
				return true
			},
		},
	}

	// Initialize repositories and services
	app.userRepo = repositories.NewMemoryUserRepository()
	app.chatRepo = repositories.NewMemoryChatRepository()
	app.messageSvc = services.NewMessageService(app.chatRepo)
	app.hub = sockets.NewConnectionHub(app.messageSvc)

	// Setup routes
	app.setupRoutes()

	// Start WebSocket hub
	go app.hub.Run()

	return app
}

// Handler returns the HTTP handler
func (a *App) Handler() http.Handler {
	return a.router
}

// setupRoutes registers all application routes
func (a *App) setupRoutes() {
	// API routes
	api := a.router.PathPrefix("/api/v1").Subrouter()

	// User management
	api.HandleFunc("/users", a.createUser).Methods("POST")
	api.HandleFunc("/users/{id}", a.getUser).Methods("GET")

	// Chat management
	api.HandleFunc("/chats", a.listUserChats).Methods("GET")
	api.HandleFunc("/chats/{chatId}/messages", a.listChatMessages).Methods("GET")

	// Message handling
	api.HandleFunc("/messages", a.sendMessage).Methods("POST")

	// WebSocket endpoint for real-time communication
	a.router.HandleFunc("/ws", a.handleWebSocket)

	// Health check
	a.router.HandleFunc("/health", a.healthCheck).Methods("GET")
}
