package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNotChatMember = errors.New("bạn không phải là thành viên của cuộc hội thoại này!")
)

type SendMessageInput struct {
	ConversationID uint64 `json:"conversation_id"`
	SenderID	uuid.UUID `json:"sender_id"`
	MessageType string `json:"message_type"`
	Content 	string `json:"content"`
	ClientMsgID string `json:"client_msg_id"`
}

type MediaVerifyResponse struct {
	ID         uuid.UUID `json:"id"`
	UploaderID uuid.UUID `json:"uploader_id"`
	FileURL    string    `json:"file_url"`
	FileType   string    `json:"file_type"`
}

type ChatUsecase interface {
	SendMessage(ctx context.Context, input SendMessageInput) (*Message, error)
	GetChatHistory(ctx context.Context, conversationID uint64, userID uuid.UUID, lastMessageID *uint64, limit int) ([]MessageResponse, error)
	CreateConversation(ctx context.Context, senderID uuid.UUID, input CreateConversationInput) (*Conversation, error)
	GetUserConversations(ctx context.Context, userID uuid.UUID) ([]Conversation, error)
	GetConversationMembers(ctx context.Context, conversationID uint64,) ([]ConversationMember, error)
	GetReadStatus(ctx context.Context, conversationID uint64, userID uuid.UUID) (*ConversationRead, error)
	MarkAsRead(ctx context.Context, conversationID uint64, userID uuid.UUID, messageID uint64) error
	CreateReadStatus(ctx context.Context, conversationID uint64, userIDs []uuid.UUID) error
	GetConversationDetail(ctx context.Context, converationID uint64, userID uuid.UUID) (*Conversation, error)
	SetReadReceiptPublisher(publisher ReadReceiptPublisher)
	MarkMessageAsDelivered(ctx context.Context, messageID uint64, userID uuid.UUID) error 
	GetPresenceRecipients(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetUserConversationIDs(ctx context.Context, userID uuid.UUID) ([]uint64, error)
}