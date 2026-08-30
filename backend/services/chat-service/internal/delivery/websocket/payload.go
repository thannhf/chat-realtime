package websocket

import (
	"chat-service/internal/domain"
	"time"
)

type WSMessage struct {
	Type string `json:"type"`
	ConversationID uint64 `json:"conversation_id,omitempty"`
	MessageID uint64 `json:"message_id,omitempty"`
	ClientMsgID    string `json:"client_msg_id,omitempty"`
	Content        string `json:"content,omitempty"`
	MessageType    string `json:"message_type,omitempty"`
}

type ConversationUpdatedPayload struct {
	ConversationID uint64 `json:"conversation_id"`
	LastMessage    *domain.Message `json:"last_message"`
	UpdatedAt time.Time `json:"updated_at"`
}