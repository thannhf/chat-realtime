package postgres

import (
	"chat-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type conversationRepository struct {
	db *gorm.DB 
}

func NewConversationRepository(db *gorm.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) CreateConversation(ctx context.Context, conv *domain.Conversation, memberIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(conv).Error; err != nil {
			return err 
		}
		for _, uid := range memberIDs {
			member := domain.ConversationMember{
				ConversationID: conv.ID,
				UserID: uid,
				Role: "member",
			}
			if err := tx.Create(&member).Error; err != nil {
				return err 
			}
		}
		return nil 
	})
}

func (r *conversationRepository) GetReadStatus(ctx context.Context, conversationID uint64, userID uuid.UUID) (*domain.ConversationRead, error) {
	var read domain.ConversationRead

	err := r.db.WithContext(ctx).Where("conversation_id = ? AND user_id = ?", conversationID, userID).First(&read).Error
	if err != nil {
		return nil, err 
	}

	if read.LastReadMessageID != nil {
		fmt.Println(
			"lastRead:",
			*read.LastReadMessageID,
		)
	} else {
		fmt.Println(
			"lastRead: nil",
		)
	}

	return &read, nil 
}

func (r *conversationRepository) MarkAsRead(
	ctx context.Context,
	conversationID uint64,
	userID uuid.UUID,
	messageID uint64,
) error {
	// fmt.Println(
	// 	"UPDATE READ:",
	// 	conversationID,
	// 	userID,
	// 	messageID,
	// )
	result := r.db.WithContext(ctx).
		Model(&domain.ConversationRead{}).
		Where(
			"conversation_id = ? AND user_id = ?",
			conversationID,
			userID,
		).
		Updates(map[string]any{
			"last_read_message_id": messageID,
			"read_at":              time.Now(),
			"updated_at":           time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	// fmt.Println(
	// 	"MarkAsRead rows affected:",
	// 	result.RowsAffected,
	// )
	return nil
}

func (r *conversationRepository) CreateReadStatus(ctx context.Context, read []*domain.ConversationRead) error {
	return r.db.WithContext(ctx).Create(read).Error
}

func (r *conversationRepository) GetConversationMembers(ctx context.Context, conversationID uint64) ([]domain.ConversationMember, error) {
	var members []domain.ConversationMember
	err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Find(&members).Error
	return members, err 
}

// func (r *conversationRepository) GetUserConversations(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
// 	var conversations []domain.Conversation
// 	err := r.db.WithContext(ctx). 
// 			Joins("JOIN conversation_members ON conversation_members.conversation_id = conversations.id").
// 			Where("conversation_members.user_id = ?", userID). 
// 			Find(&conversations).Error
// 	return conversations, err 
// }
func (r *conversationRepository) GetUserConversations(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
    var conversations []domain.Conversation

    err := r.db.WithContext(ctx).
        Model(&domain.Conversation{}).
        Joins(`
            JOIN conversation_members
            ON conversation_members.conversation_id = conversations.id
        `).
        Where("conversation_members.user_id = ?", userID).
        Preload("Members").
        Order("conversations.updated_at DESC").
        Find(&conversations).Error

    if err != nil {
        return nil, err
    }

    return conversations, nil
}

func (r *conversationRepository) GetConversationDetail(ctx context.Context, conversationID uint64) (*domain.Conversation, error) {
	var conversation domain.Conversation

	err := r.db.WithContext(ctx).Preload("Members").First(&conversation, conversationID).Error

	if err != nil {
		return nil, err 
	}

	return &conversation, nil 
}

func (r *conversationRepository) GetDirectConversation(ctx context.Context, userA, userB uuid.UUID) (*domain.Conversation, error) {
	query := `
		SELECT c.id, c.type, c.name, c.created_at, c.updated_at FROM conversations c 
		JOIN conversation_members cm1 ON c.id = cm1.conversation_id AND cm1.user_id = $1
		JOIN conversation_members cm2 ON c.id = cm2.conversation_id AND cm2.user_id = $2
		WHERE c.type = 'direct'
		LIMIT 1
	`

	var conv domain.Conversation
	err := r.db.ConnPool.QueryRowContext(ctx, query, userA, userB).Scan(
		&conv.ID,
		&conv.Type,
		&conv.Name,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil 
	}

	if err != nil {
		return nil, fmt.Errorf("Lỗi khi tìm cuộc trò chuyện: %w", err)
	}

	return &conv, nil
}

func (r *conversationRepository) GetUnreadCount(
	ctx context.Context,
	conversationID uint64,
	lastReadMessageID *uint64,
) (int64, error) {

	var count int64

	var lastRead uint64

	if lastReadMessageID != nil {
		lastRead = *lastReadMessageID
	}

	err := r.db.WithContext(ctx).
		Model(&domain.Message{}).
		Where(
			"conversation_id = ? AND id > ?",
			conversationID,
			lastRead,
		).
		Count(&count).Error

	return count, err
}

func (r *conversationRepository) GetLastMessage(ctx context.Context, conversationID uint64) (*domain.Message, error) {
	var message domain.Message

	err := r.db.WithContext(ctx).Where("conversation_id = ?", conversationID).Order("id DESC").First(&message).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil 
	}

	if err != nil {
		return nil, err 
	}

	return &message, nil 
}

func (r *conversationRepository) GetPresenceRecipients(
    ctx context.Context,
    userID uuid.UUID,
) ([]uuid.UUID, error) {

    var userIDs []uuid.UUID

    err := r.db.WithContext(ctx).
        Table("conversation_members AS cm1").
        Select("DISTINCT cm2.user_id").
        Joins(`
            JOIN conversation_members AS cm2
                ON cm2.conversation_id = cm1.conversation_id
        `).
        Where("cm1.user_id = ?", userID).
        Where("cm2.user_id <> ?", userID).
        Scan(&userIDs).Error

    if err != nil {
        return nil, err
    }

    return userIDs, nil
}

func (r *conversationRepository) GetUserConversationIDs(ctx context.Context, userID uuid.UUID) ([]uint64, error) {
	var conversationIDs []uint64

	err := r.db.WithContext(ctx).Table("conversation_members").Where("user_id = ?", userID).Pluck("conversation_id", &conversationIDs).Error

	if err != nil {
		return nil, err 
	}

	return conversationIDs, nil 
}