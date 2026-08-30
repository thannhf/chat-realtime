package domain

import (
	
	"github.com/google/uuid"
	"encoding/json"
)

type ReadReceiptPayload struct {
	ConversationID uint64 `json:"conversation_id"`
	UserID uuid.UUID `json:"user_id"`
	LastReadMessageID uint64 `json:"last_read_message_id"`
}

type ReadReceiptPublisher interface {
	PublishReadReceipt(payload ReadReceiptPayload)
}

type MessageWrapper struct {
	SenderID uuid.UUID
	Payload  []byte
}

type EventType string

const (
	EventMessage             EventType = "message"
	EventConversationUpdated EventType = "conversation_updated"
	EventTyping              EventType = "typing"
	EventPresence            EventType = "presence"
	EventReadReceipt         EventType = "read_receipt"
	EventMessageStatus       EventType = "message_status"
	EventMessageDelivered    EventType = "message_delivered"
)

type MessageStatusPayload struct {
	MessageID      uint64               `json:"message_id"`
	ConversationID uint64               `json:"conversation_id"`
	UserID         uuid.UUID            `json:"user_id"`
	Status         MessageStatus `json:"status"`
}

type MessageDeliveredPayload struct {
	ConversationID uint64 `json:"conversation_id"`
	MessageID      uint64 `json:"message_id"`
}

type WSEvent struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

type PresencePayload struct {
	UserID uuid.UUID `json:"user_id"`
	Status string `json:"status"`
}

type WSClientEvent struct {
	Type EventType       `json:"type"`
	Data json.RawMessage `json:"data"`
}