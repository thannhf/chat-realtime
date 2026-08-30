package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyMessage   = errors.New("Nội dung tin nhắn không được để trống")
	ErrMessageTooLong = errors.New("nội dung tin nhắn không được vượt quá 5000 ký tự")
)

type MessageStatus int16

const (
	MessageStatusSend      MessageStatus = 1
	MessageStatusDelivered MessageStatus = 2
	MessageStatusRead      MessageStatus = 3
)

type Message struct {
	ID             uint64        `json:"id" gorm:"primaryKey;autoIncrement"`
	ConversationID uint64        `json:"conversation_id" gorm:"type:bigint;not null;index:idx_messages_conv_created,priority:1"`
	SenderID       uuid.UUID     `json:"sender_id" gorm:"type:uuid;not null"`
	MessageType    string        `json:"message_type" gorm:"type:varchar(50);default:'text'"`
	Content        string        `json:"content" gorm:"type:text;not null"`
	Status         MessageStatus `json:"status" gorm:"type:smallint;not null;default:1"`
	CreatedAt      time.Time     `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp;index:idx_messages_conv_created,priority:2,sort:desc"`
	ClientMsgID    *string       `json:"client_msg_id,omitempty" gorm:"type:uuid;default:null"`
}

type ConversationRead struct {
	ID                    uint64 `gorm:"primaryKey"`
	ConversationID        uint64
	UserID                uuid.UUID
	LastReadMessageID     *uint64 `gorm:"column:last_read_message_id"`
	LastReceivedMessageID *uint64 `gorm:"column:last_received_message_id"`
	ReadAt                time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (m *Message) Validate() error {
	if len(m.Content) == 0 {
		return ErrEmptyMessage
	}
	if len(m.Content) > 5000 {
		return ErrMessageTooLong
	}
	return nil
}

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *Message) error
	GetChatHistory(ctx context.Context, conversationID uint64, lastMessageID *uint64, limit int) ([]Message, error)
	UpdateMessageStatus(ctx context.Context, messageID uint64, status MessageStatus) error
	GetMessageByID(ctx context.Context, messageID uint64) (*Message, error)
}

type ConversationRepository interface {
	CreateConversation(ctx context.Context, conv *Conversation, memberIDs []uuid.UUID) error
	GetConversationMembers(ctx context.Context, conversationID uint64) ([]ConversationMember, error)
	GetUserConversations(ctx context.Context, userID uuid.UUID) ([]Conversation, error)
	GetDirectConversation(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (*Conversation, error)
	GetReadStatus(ctx context.Context, conversationID uint64, userID uuid.UUID) (*ConversationRead, error)
	MarkAsRead(ctx context.Context, conversationID uint64, userID uuid.UUID, messageID uint64) error
	CreateReadStatus(ctx context.Context, read []*ConversationRead) error
	GetLastMessage(ctx context.Context, conversationID uint64) (*Message, error)
	GetUnreadCount(ctx context.Context, conversationID uint64, lastReadMessageID *uint64) (int64, error)
	GetConversationDetail(ctx context.Context, conversationID uint64) (*Conversation, error)
	GetPresenceRecipients(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetUserConversationIDs(ctx context.Context, userID uuid.UUID) ([]uint64, error)
}
