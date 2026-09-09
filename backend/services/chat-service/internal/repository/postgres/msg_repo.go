package postgres

import (
	"chat-service/internal/domain"
	"context"
	"gorm.io/gorm"
)

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) SaveMessage(ctx context.Context, msg *domain.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

func (r *messageRepository) GetChatHistory(ctx context.Context, conversationID uint64, lastMessageID *uint64, limit int) ([]domain.Message, error) {
	var messages []domain.Message

	query := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID)

	if lastMessageID != nil {
		query = query.Where("id < ?", *lastMessageID)
	}

	err := query.Order("created_at DESC, id DESC").Limit(limit).Find(&messages).Error

	if err != nil {
		return nil, err 
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i + 1, j - 1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (r *messageRepository) UpdateMessageStatus(ctx context.Context, messageID uint64, status domain.MessageStatus) error {
	return r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where("id = ?", messageID).
		Update("status", status).Error
}

func (r *messageRepository) GetMessageByID(ctx context.Context, messageID uint64) (*domain.Message, error) {
	var message domain.Message

	err := r.db.WithContext(ctx).Where("id = ?", messageID).First(&message).Error

	if err != nil {
		return nil, err 
	}

	return &message, nil 
}