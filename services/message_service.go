package services

import (
	"time"

	"messaging-app/domain"
	"messaging-app/repositories"
)

// MessageService handles business logic for messaging operations
type MessageService struct {
	chatRepo repositories.ChatRepository
}

// NewMessageService creates a new message service
func NewMessageService(chatRepo repositories.ChatRepository) *MessageService {
	return &MessageService{
		chatRepo: chatRepo,
	}
}

// SendMessage sends a message between users with idempotency support
func (s *MessageService) SendMessage(senderID, recipientID, content, idempotencyKey string) (*domain.Message, error) {
	// Validate users exist (in production, this would check user repository)
	if senderID == "" || recipientID == "" {
		return nil, domain.ErrInvalidUser
	}

	if senderID == recipientID {
		return nil, domain.ErrCannotMessageSelf
	}

	if content == "" {
		return nil, domain.ErrEmptyMessage
	}

	// Find or create chat
	chat, err := s.chatRepo.FindByParticipants(senderID, recipientID)
	if err != nil {
		if err == domain.ErrChatNotFound {
			// Create new chat
			chat = &domain.Chat{
				Participant1: senderID,
				Participant2: recipientID,
			}
			if err := s.chatRepo.Create(chat); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Check for duplicate message using idempotency key
	if idempotencyKey != "" {
		existingMsg, err := s.chatRepo.FindMessageByKey(chat.ID, idempotencyKey)
		if err == nil && existingMsg != nil {
			return existingMsg, nil // Return existing message for idempotency
		}
	}

	// Create message
	message := &domain.Message{
		ChatID:         chat.ID,
		SenderID:       senderID,
		Content:        content,
		Status:         domain.StatusSent,
		Timestamp:      time.Now(),
		IdempotencyKey: idempotencyKey,
	}

	if err := s.chatRepo.AddMessage(message); err != nil {
		return nil, err
	}

	return message, nil
}
