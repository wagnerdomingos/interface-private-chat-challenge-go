package repositories

import "messaging-app/domain"

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(user *domain.User) error
	FindByID(id string) (*domain.User, error)
	FindByUsername(username string) (*domain.User, error)
	Exists(username string) bool
}

// ChatRepository defines the interface for chat data operations
type ChatRepository interface {
	Create(chat *domain.Chat) error
	FindByID(id string) (*domain.Chat, error)
	FindByParticipants(user1ID, user2ID string) (*domain.Chat, error)
	FindUserChats(userID string, pagination domain.PaginationParams) ([]*domain.Chat, int, error)
	FindChatMessages(chatID string, pagination domain.PaginationParams) ([]*domain.Message, int, error)
	AddMessage(message *domain.Message) error
	UpdateMessageStatus(messageID string, status domain.MessageStatus) error
	FindMessageByID(id string) (*domain.Message, error)
	FindMessageByKey(chatID, idempotencyKey string) (*domain.Message, error)
}
