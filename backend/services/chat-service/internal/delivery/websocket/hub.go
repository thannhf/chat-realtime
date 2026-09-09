package websocket

import (
	"chat-service/internal/domain"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	Clients         map[uuid.UUID]map[*Client]bool
	Broadcast       chan domain.MessageWrapper
	Register        chan *Client
	Unregister      chan *Client
	mu              sync.RWMutex
	ChatUsecase     domain.ChatUsecase
	PresenceUsecase domain.PresenceUsecase
}

func NewHub(chatUsecase domain.ChatUsecase, presenceUsecase domain.PresenceUsecase) *Hub {
	return &Hub{
		Broadcast:       make(chan domain.MessageWrapper),
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		Clients:         make(map[uuid.UUID]map[*Client]bool),
		ChatUsecase:     chatUsecase,
		PresenceUsecase: presenceUsecase,
	}
}

func (h *Hub) Run() {
	for {
		select {
		// register
		case client := <-h.Register:
			h.mu.Lock()
			if h.Clients[client.UserID] == nil {
				h.Clients[client.UserID] = make(map[*Client]bool)
			}
			h.Clients[client.UserID][client] = true
			h.mu.Unlock()

			ctx := context.Background()
			
			if err := h.PresenceUsecase.AddConnection(
				ctx, client.UserID, client.ConnectionID); err != nil {

				log.Printf("Lỗi không thể tạo ConnectionID %s: %v",
					client.ConnectionID, err,
				)

				h.mu.Lock()
				delete(h.Clients[client.UserID], client)

				if len(h.Clients[client.UserID]) == 0 {
					delete(h.Clients, client.UserID)
				}
				h.mu.Unlock()

				continue
			}

			count, err := h.PresenceUsecase.CountConnections(ctx, client.UserID)


			if err != nil {
				log.Printf("Lỗi trong việc đếm connection của user %s: %v",client.UserID, err)

				continue
			}

			log.Printf(
				"Presence DEBUG user=%s connectionID=%s count=%d",
				client.UserID,
				client.ConnectionID,
				count,
			)

			if count != 1 {
				continue
			}

			if err := h.PresenceUsecase.SetUserOnline(ctx, client.UserID); err != nil {
				log.Printf("Lỗi ghi nhận trạng thái Online cho User %s: %v", client.UserID, err)

				continue
			}
			log.Printf("User %s vừa kết nối (Online)",client.UserID)

			recipients, err := h.ChatUsecase.GetPresenceRecipients(
				ctx, client.UserID,
			)

			if err != nil {
				log.Printf("Get presence recipients failed for user %s: %v", client.UserID, err)
				
				continue
			}

			h.PublishPresenceToUsers(
				domain.PresencePayload{
					UserID: client.UserID,
					Status: "online",
				},
				recipients,
			)

			h.SendInitialPresence(ctx, client.UserID)

			// unregister
		case client := <-h.Unregister:
			h.mu.Lock()

			connections, exists := h.Clients[client.UserID]
			if !exists {
				h.mu.Unlock()
				continue
			}

			if _, ok := connections[client]; !ok {
				h.mu.Unlock()
				continue
			}

			delete(connections, client)
			close(client.Send)

			isLastConnection := len(connections) == 0

			if isLastConnection {
				delete(h.Clients, client.UserID)
			}

			h.mu.Unlock()

			ctx := context.Background()
			if err := h.PresenceUsecase.RemoveConnection(ctx, client.UserID, client.ConnectionID); err != nil {
				log.Printf("Lỗi không thể xóa connectionID %s: %v", client.ConnectionID, err)

				continue
			}

			count, err := h.PresenceUsecase.CountConnections(ctx, client.UserID)

			if err != nil {
				log.Printf("Lỗi trong việc đếm connection của user %s: %v", client.UserID, err)

				continue
			}

			if count > 0 {
				continue
			}

			if err := h.PresenceUsecase.SetUserOffline(ctx, client.UserID); err != nil {
				log.Printf(
					"Lỗi ghi nhận trạng thái Offline + Last Seen cho User %s: %v",
					client.UserID,
					err,
				)
			} else {
				log.Printf("User %s đã ngắt toàn bộ kết nối (offline)", client.UserID)

				recipients, err := h.ChatUsecase.GetPresenceRecipients(ctx, client.UserID)

				if err != nil {
					log.Printf("Get presence recipients failed for user %s: %v", client.UserID, err)
				} else {
					h.PublishPresenceToUsers(
						domain.PresencePayload{
							UserID: client.UserID,
							Status: "offline",
						},
						recipients,
					)
				}
			}

		// broadcast and client event
		case wrapper := <-h.Broadcast:
			var clientEvent domain.WSClientEvent

			if err := json.Unmarshal(wrapper.Payload, &clientEvent); err != nil {
				log.Printf("Invalid websocket event from user %s: %v", wrapper.SenderID, err)
				continue
			}

			switch clientEvent.Type {
			// message
			case domain.EventMessage:
				var payloadMessage WSMessage

				if err := json.Unmarshal(clientEvent.Data, &payloadMessage); err != nil {
					log.Printf("Invalid message payload from user %s: %v", wrapper.SenderID, err)
					continue
				}

				ctx := context.Background()

				input := domain.SendMessageInput{
					ConversationID: payloadMessage.ConversationID,
					SenderID:       wrapper.SenderID,
					MessageType:    payloadMessage.MessageType,
					Content:        payloadMessage.Content,
					ClientMsgID:    payloadMessage.ClientMsgID,
				}

				savedMsg, err := h.ChatUsecase.SendMessage(ctx, input)
				if err != nil {
					log.Printf("Gửi tin nhắn thất bại từ User %s: %v", wrapper.SenderID, err)
					continue
				}

				// message event
				messageEvent := domain.WSEvent{
					Type: domain.EventMessage,
					Data: savedMsg,
				}

				messagePayload, err := json.Marshal(messageEvent)
				if err != nil {
					log.Printf("Lỗi mã hóa JSON tin nhắn: %v", err)
					continue
				}

				// conversation updated
				conversationEvent := domain.WSEvent{
					Type: domain.EventConversationUpdated,
					Data: ConversationUpdatedPayload{
						ConversationID: payloadMessage.ConversationID,
						LastMessage:    savedMsg,
						UpdatedAt:      savedMsg.CreatedAt,
					},
				}

				conversationPayload, err := json.Marshal(conversationEvent)
				if err != nil {
					log.Printf("Encode conversation event failed: %v", err)
					continue
				}

				// conversation members
				members, err := h.ChatUsecase.GetConversationMembers(
					ctx,
					payloadMessage.ConversationID,
				)
				if err != nil {
					log.Printf(
						"Get conversation members failed: %v",
						err,
					)
					continue
				}

				// broadcast
				h.mu.Lock()
				for _, member := range members {
					connections, ok := h.Clients[member.UserID]
					if !ok {
						continue
					}

					for client := range connections {
						// message
						if !h.sendToClient(client, messagePayload) {
							delete(connections, client)
							continue
						}

						// conversation_updated
						if !h.sendToClient(client, conversationPayload) {
							delete(connections, client)
							continue
						}
					}
				}
				h.mu.Unlock()

			// message delivered
			case domain.EventMessageDelivered:
				var payload domain.MessageDeliveredPayload

				if err := json.Unmarshal(clientEvent.Data, &payload); err != nil {
					log.Printf("Invalid message_delivered payload from user %s: %v", wrapper.SenderID, err)
					continue
				}

				ctx := context.Background()
				err := h.ChatUsecase.MarkMessageAsDelivered(ctx, payload.MessageID, wrapper.SenderID)

				if err != nil {
					log.Printf(
						"Mark message delivered failed: %v",
						err,
					)
					continue
				}

				statusEvent := domain.WSEvent{
					Type: domain.EventMessageStatus,
					Data: domain.MessageStatusPayload{
						MessageID:      payload.MessageID,
						ConversationID: payload.ConversationID,
						UserID:         wrapper.SenderID,
						Status:         domain.MessageStatusDelivered,
					},
				}

				statusData, err := json.Marshal(statusEvent)
				if err != nil {
					log.Printf("marshal message status failed: %v", err)
					continue
				}

				members, err := h.ChatUsecase.GetConversationMembers(ctx, payload.ConversationID)
				if err != nil {
					log.Printf("get conversation members failed: %v", err)
					continue
				}

				h.mu.RLock()
				for _, member := range members {
					connections, ok := h.Clients[member.UserID]
					if !ok {
						continue
					}

					for client := range connections {
						select {
						case client.Send <- statusData:
						default:
							log.Printf("client send buffer full: %s", member.UserID)
						}
					}
				}
				h.mu.RUnlock()
				continue

			// unknown
			default:
				log.Printf("Unknown websocket event type: %s", clientEvent.Type)
			}
		}
	}
}

// send to client
func (h *Hub) sendToClient(client *Client, data []byte) bool {
	select {
	case client.Send <- data:
		return true
	default:
		log.Printf(
			"Client send buffer full, closing connection: %s",
			client.UserID,
		)

		return false
	}
}

// publish read receipt
func (h *Hub) PublishReadReceipt(payload domain.ReadReceiptPayload) {
	event := domain.WSEvent{
		Type: domain.EventReadReceipt,
		Data: payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf(
			"marshal read receipt failed: %v",
			err,
		)
		return
	}

	ctx := context.Background()
	members, err := h.ChatUsecase.GetConversationMembers(
		ctx,
		payload.ConversationID,
	)
	if err != nil {
		log.Printf("get conversation members failed: %v", err)
		return
	}

	// log.Printf("READ RECEIPT: conversation=%d reader=%s last_read=%d members=%d", payload.ConversationID, payload.UserID, payload.LastReadMessageID, len(members))

	// for _, member := range members {
	// 	log.Printf(
	// 		"READ RECEIPT MEMBER: user=%s",
	// 		member.UserID,
	// 	)
	// }

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, member := range members {
		connections, ok := h.Clients[member.UserID]
		if !ok {
			continue
		}

		for client := range connections {
			select {
			case client.Send <- data:
			default:
				log.Printf(
					"client send buffer full: %s",
					member.UserID,
				)
			}
		}
	}
}

func (h *Hub) PublishPresence(userID uuid.UUID, status string) {
	event := domain.WSEvent{
		Type: domain.EventPresence,
		Data: domain.PresencePayload{
			UserID: userID,
			Status: status,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("marshal presence event failed: %v", err)
		return
	}

	ctx := context.Background()
	recipients, err := h.getPresenceRecipients(ctx, userID)
	if err != nil {
		log.Printf("get presence recipients failed for user %s: %v", userID, err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for recipientID := range recipients {
		connections, ok := h.Clients[recipientID]
		if !ok {
			continue
		}

		for client := range connections {
			select {
			case client.Send <- data:
			default:
				log.Printf("client send buffer full: %s", client.UserID)
			}
		}
	}
}

func (h *Hub) PublishPresenceToUsers(payload domain.PresencePayload, userIDs []uuid.UUID) {
	// log.Printf(
	// 	"PUBLISH PRESENCE: user=%s status=%s recipients=%v",
	// 	payload.UserID,
	// 	payload.Status,
	// 	userIDs,
	// )
	event := domain.WSEvent{
		Type: domain.EventPresence,
		Data: payload,
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("marshal presence event failed: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range userIDs {
		connections, ok := h.Clients[userID]
		if !ok {
			continue
		}

		for client := range connections {
			select {
			case client.Send <- data:
			default:
				log.Printf("client send buffer full: %s", client.UserID)
			}
		}
	}
}

func (h *Hub) getPresenceRecipients(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]bool, error) {
	conversations, err := h.ChatUsecase.GetUserConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	recipients := make(map[uuid.UUID]bool)
	for _, conversation := range conversations {
		members, err := h.ChatUsecase.GetConversationMembers(ctx, conversation.ID)
		if err != nil {
			return nil, err
		}

		for _, member := range members {
			if member.UserID == userID {
				continue
			}
			recipients[member.UserID] = true
		}
	}
	return recipients, nil
}

func (h *Hub) SendInitialPresence(ctx context.Context, userID uuid.UUID) {
	recipients, err := h.ChatUsecase.GetPresenceRecipients(ctx, userID)
	if err != nil {
		log.Printf("Get presence recipients failed for user %s: %v", userID, err)
		return
	}

	if len(recipients) == 0 {
		return
	}

	presenceMap, err := h.PresenceUsecase.GetUsersPresence(ctx, recipients)
	if err != nil {
		log.Printf("Get users presence failed for user %s: %v", userID, err)
		return
	}

	h.mu.RLock()

	connections, ok := h.Clients[userID]
	clients := make([]*Client, 0, len(connections))

	if ok {
		for client := range connections {
			clients = append(clients, client)
		}
	}

	h.mu.RUnlock()

	if !ok {
		return
	}

	for recipientID, status := range presenceMap {
		event := domain.WSEvent{
			Type: domain.EventPresence,
			Data: domain.PresencePayload{
				UserID: recipientID,
				Status: status,
			},
		}

		data, err := json.Marshal(event)
		if err != nil {
			log.Printf("marshal initial presence failed: %v", err)
			continue
		}

		for _, client := range clients {
			select {
			case client.Send <- data:
			default:
				log.Printf("client send buffer full: %s", client.UserID)
			}
		}
	}
}
