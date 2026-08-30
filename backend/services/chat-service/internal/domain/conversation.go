package domain

import (
	"time"

	"github.com/google/uuid"
)

type ConversationType int16

const (
	DirectChat ConversationType = 1
	GroupChat  ConversationType = 2
)

type Conversation struct {
	ID        uint64               `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      *string              `json:"name,omitempty" gorm:"type:varchar(255);default:null"`
	Type      ConversationType     `json:"type" gorm:"type:smallint;not null;default:1"`
	CreatedAt time.Time            `json:"created_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
	UpdatedAt time.Time            `json:"updated_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
	Members   []ConversationMember `json:"members,omitempty" gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE;"`
	Messages  []Message            `json:"messages,omitempty" gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE;"`
	LastMessage *Message `json:"last_message,omitempty" gorm:"-"`
	UnreadCount int64 `json:"unread_count" gorm:"-"`
}

type ConversationMember struct {
	ID             uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	ConversationID uint64    `json:"conversation_id" gorm:"type:bigint;not null;uniqueIndex:idx_conv_user"`
	UserID         uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_conv_user;index:idx_members_user_id"`
	Role           string    `json:"role" gorm:"type:varchar(50);default:'member'"`
	JoinedAt       time.Time `json:"joined_at" gorm:"type:timestamp with time zone;default:current_timestamp"`
}

type CreateConversationInput struct {
	Type      ConversationType `json:"type" binding:"required"`
	Name      *string          `json:"name,omitempty"`
	MemberIDs []uuid.UUID      `json:"member_ids" binding:"required,gt=0"`
}
