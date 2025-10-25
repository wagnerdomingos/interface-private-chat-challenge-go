package repositories

import (
	"sort"
	"sync"
	"time"

	"messaging-app/domain"

	"github.com/google/uuid"
)

// MemoryChatRepository implements ChatRepository with in-memory storage
type MemoryChatRepository struct {
	chats    map[string]*domain.Chat
	messages map[string][]*domain.Message // chatID -> messages
	mutex    sync.RWMutex
}

// NewMemoryChatRepository creates a new in-memory chat repository
func NewMemoryChatRepository() *MemoryChatRepository {
	return &MemoryChatRepository{
		chats:    make(map[string]*domain.Chat),
		messages: make(map[string][]*domain.Message),
	}
}

// Create adds a new chat to the repository
func (r *MemoryChatRepository) Create(chat *domain.Chat) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if chat.ID == "" {
		chat.ID = uuid.New().String()
	}

	if chat.CreatedAt.IsZero() {
		chat.CreatedAt = time.Now()
	}

	chat.UpdatedAt = time.Now()
	r.chats[chat.ID] = chat
	r.messages[chat.ID] = []*domain.Message{}

	return nil
}

// FindByID retrieves a chat by its ID
func (r *MemoryChatRepository) FindByID(id string) (*domain.Chat, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	chat, exists := r.chats[id]
	if !exists {
		return nil, domain.ErrChatNotFound
	}

	return chat, nil
}

// FindByParticipants finds a chat between two users
func (r *MemoryChatRepository) FindByParticipants(user1ID, user2ID string) (*domain.Chat, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, chat := range r.chats {
		if (chat.Participant1 == user1ID && chat.Participant2 == user2ID) ||
			(chat.Participant1 == user2ID && chat.Participant2 == user1ID) {
			return chat, nil
		}
	}

	return nil, domain.ErrChatNotFound
}

// FindUserChats retrieves all chats for a user with pagination
func (r *MemoryChatRepository) FindUserChats(userID string, pagination domain.PaginationParams) ([]*domain.Chat, int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var userChats []*domain.Chat
	for _, chat := range r.chats {
		if chat.Participant1 == userID || chat.Participant2 == userID {
			userChats = append(userChats, chat)
		}
	}

	// Sort by updated_at (most recent first)
	sort.Slice(userChats, func(i, j int) bool {
		return userChats[i].UpdatedAt.After(userChats[j].UpdatedAt)
	})

	total := len(userChats)
	start, end := calculatePaginationBounds(pagination.Page, pagination.PageSize, total)

	if start >= total {
		return []*domain.Chat{}, total, nil
	}

	return userChats[start:end], total, nil
}

// FindChatMessages retrieves messages for a chat with pagination
func (r *MemoryChatRepository) FindChatMessages(chatID string, pagination domain.PaginationParams) ([]*domain.Message, int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	messages, exists := r.messages[chatID]
	if !exists {
		return nil, 0, domain.ErrChatNotFound
	}

	total := len(messages)
	start, end := calculatePaginationBounds(pagination.Page, pagination.PageSize, total)

	if start >= total {
		return []*domain.Message{}, total, nil
	}

	// Return messages in chronological order (oldest first)
	result := make([]*domain.Message, end-start)
	for i := start; i < end; i++ {
		result[i-start] = messages[i]
	}

	return result, total, nil
}

// AddMessage adds a message to a chat
func (r *MemoryChatRepository) AddMessage(message *domain.Message) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Update chat's updated_at timestamp
	if chat, exists := r.chats[message.ChatID]; exists {
		chat.UpdatedAt = time.Now()
	}

	r.messages[message.ChatID] = append(r.messages[message.ChatID], message)
	return nil
}

// UpdateMessageStatus updates the status of a message
func (r *MemoryChatRepository) UpdateMessageStatus(messageID string, status domain.MessageStatus) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, messages := range r.messages {
		for _, msg := range messages {
			if msg.ID == messageID {
				msg.Status = status
				return nil
			}
		}
	}

	return domain.ErrMessageNotFound
}

// FindMessageByID finds a message by its ID
func (r *MemoryChatRepository) FindMessageByID(id string) (*domain.Message, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, messages := range r.messages {
		for _, msg := range messages {
			if msg.ID == id {
				return msg, nil
			}
		}
	}

	return nil, domain.ErrMessageNotFound
}

// FindMessageByKey finds a message by idempotency key
func (r *MemoryChatRepository) FindMessageByKey(chatID, idempotencyKey string) (*domain.Message, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	messages, exists := r.messages[chatID]
	if !exists {
		return nil, domain.ErrChatNotFound
	}

	for _, msg := range messages {
		if msg.IdempotencyKey == idempotencyKey {
			return msg, nil
		}
	}

	return nil, domain.ErrMessageNotFound
}

// calculatePaginationBounds calculates start and end indices for pagination
func calculatePaginationBounds(page, pageSize, total int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}

	end := start + pageSize
	if end > total {
		end = total
	}

	return start, end
}
