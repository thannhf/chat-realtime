package websocket

import "github.com/google/uuid"

type ReadReceiptPayload struct {
	ConversationID    uint64    `json:"conversation_id"`
	UserID            uuid.UUID `json:"user_id"`
	LastReadMessageID uint64    `json:"last_read_message_id"`
}