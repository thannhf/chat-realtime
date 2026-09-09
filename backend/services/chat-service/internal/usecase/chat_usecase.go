package usecase

import (
	"chat-service/internal/domain"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type chatUsecase struct {
	msgRepo        domain.MessageRepository
	convRepo       domain.ConversationRepository
	userGateway    domain.UserGateway
	readReceiptPub domain.ReadReceiptPublisher
}

func NewChatUsecase(msgRepo domain.MessageRepository, convRepo domain.ConversationRepository, userGateway domain.UserGateway, readReceiptPub domain.ReadReceiptPublisher) domain.ChatUsecase {
	return &chatUsecase{
		msgRepo:        msgRepo,
		convRepo:       convRepo,
		userGateway:    userGateway,
		readReceiptPub: readReceiptPub,
	}
}

func (u *chatUsecase) GetUserConversations(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
	conversations, err := u.convRepo.GetUserConversations(ctx, userID)
	if err != nil {
		return nil, err
	}

	for i := range conversations {
		conversation := &conversations[i]

		lastMessage, err := u.convRepo.GetLastMessage(ctx, conversation.ID)
		if err != nil {
			return nil, err
		}

		conversation.LastMessage = lastMessage

		readStatus, err := u.convRepo.GetReadStatus(ctx, conversation.ID, userID)
		if err != nil {
			return nil, err
		}

		unreadCount, err := u.convRepo.GetUnreadCount(ctx, conversation.ID, readStatus.LastReadMessageID)
		if err != nil {
			return nil, err
		}
		conversation.UnreadCount = unreadCount
	}
	return conversations, nil
}

func (u *chatUsecase) CreateConversation(ctx context.Context, senderID uuid.UUID, input domain.CreateConversationInput) (*domain.Conversation, error) {
	var conv *domain.Conversation
	var allMembers []uuid.UUID

	if input.Type == domain.DirectChat {
		if len(input.MemberIDs) != 1 {
			return nil, errors.New(
				"Chat đơn chỉ được truyền đúng 1 người nhận",
			)
		}
		receiverID := input.MemberIDs[0]

		isFriend, err := u.userGateway.VerifyFriendship(ctx, senderID, receiverID)
		if err != nil || !isFriend {
			return nil, errors.New(
				"không thể tạo phòng chat với người chưa kết bạn",
			)
		}

		existingConv, err := u.convRepo.GetDirectConversation(ctx, senderID, receiverID)
		if err == nil && existingConv != nil {
			return existingConv, nil
		}

		allMembers = []uuid.UUID{
			senderID,
			receiverID,
		}

		conv = &domain.Conversation{
			Type: domain.DirectChat,
		}
	}

	if input.Type == domain.GroupChat {
		if input.Name == nil || *input.Name == "" {
			return nil, errors.New(
				"chat nhóm bắt buộc phải có tên nhóm",
			)
		}

		allMembers = append(input.MemberIDs, senderID)
		conv = &domain.Conversation{
			Type: domain.GroupChat,
			Name: input.Name,
		}
	}

	if conv == nil {
		return nil, errors.New("loại cuộc trò chuyện không hợp lệ")
	}

	err := u.convRepo.CreateConversation(ctx, conv, allMembers)
	if err != nil {
		return nil, err
	}

	err = u.CreateReadStatus(ctx, conv.ID, allMembers)
	if err != nil {
		return nil, err
	}

	return conv, nil
}

func (u *chatUsecase) SendMessage(ctx context.Context, input domain.SendMessageInput) (*domain.Message, error) {
	members, err := u.convRepo.GetConversationMembers(ctx, input.ConversationID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, m := range members {
		if m.UserID == input.SenderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, domain.ErrNotChatMember
	}

	contentData := input.Content
	if input.MessageType == "image" || input.MessageType == "video" {
		actualFileURL, err := u.verifyMediaWithRemote(ctx, input.Content, input.SenderID)
		if err != nil {
			return nil, fmt.Errorf("xác thực media thất bại: %w", err)
		}
		contentData = actualFileURL
	}

	msg := &domain.Message{
		ConversationID: input.ConversationID,
		SenderID:       input.SenderID,
		MessageType:    input.MessageType,
		Content:        contentData,
		Status:         domain.MessageStatusSend,
	}

	if input.ClientMsgID != "" {
		msg.ClientMsgID = &input.ClientMsgID
	}

	if err := msg.Validate(); err != nil {
		return nil, err
	}

	if err := u.msgRepo.SaveMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (u *chatUsecase) GetConversationMembers(ctx context.Context, conversationID uint64) ([]domain.ConversationMember, error) {
	return u.convRepo.GetConversationMembers(ctx, conversationID)
}

func (u *chatUsecase) GetChatHistory(ctx context.Context, conversationID uint64, userID uuid.UUID, lastMessageID *uint64, limit int) ([]domain.MessageResponse, error) {
	members, err := u.convRepo.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, m := range members {
		if m.UserID == userID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, domain.ErrNotChatMember
	}

	if limit > 50 || limit <= 0 {
		limit = 20
	}

	messages, err := u.msgRepo.GetChatHistory(ctx, conversationID, lastMessageID, limit)
	if err != nil {
		return nil, err 
	}

	readStatuses, err := u.convRepo.GetConversationReadStatuses(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.MessageResponse, 0, len(messages))
	for _, message := range messages {
		response := domain.MessageResponse{
			Message: message,
			SeenBy: []uuid.UUID{},
		}

		for _, read := range readStatuses {
			if read.LastReadMessageID == nil {
				continue 
			}

			if read.UserID == message.SenderID {
				continue
			}

			if message.ID <= *read.LastReadMessageID {
				response.SeenBy = append(response.SeenBy, read.UserID)
			}
		}
		responses = append(responses, response)
	}

	return responses, nil
}

func (u *chatUsecase) verifyMediaWithRemote(ctx context.Context, mediaIDStr string, senderID uuid.UUID) (string, error) {
	if _, err := uuid.Parse(mediaIDStr); err != nil {
		return "", errors.New("chuỗi nội dung gửi lên phải là một UUID hợp lệ của file media")
	}

	url := fmt.Sprintf("http://localhost:8083/api/internal/media/%s", mediaIDStr)

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("không thể kết nối tới media-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", errors.New("file media truyền lên không tồn tại trên hệ thống")
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("media-service phản hồi mã lỗi: %d", resp.StatusCode)
	}

	var mediaInfo domain.MediaVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&mediaInfo); err != nil {
		return "", fmt.Errorf("không thể giải mã dữ liệu xác thực: %w", err)
	}

	if mediaInfo.UploaderID != senderID {
		return "", errors.New("bảo mật: bạn không có quyền chia sẻ file media của người khác")
	}

	return mediaInfo.FileURL, nil
}

func (u *chatUsecase) MarkAsRead(ctx context.Context, conversationID uint64, userID uuid.UUID, messageID uint64) error {
	members, err := u.convRepo.GetConversationMembers(
		ctx,
		conversationID,
	)

	if err != nil {
		return err
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return domain.ErrNotChatMember
	}

	if err = u.convRepo.MarkAsRead(ctx,conversationID,userID,messageID); err != nil {
		return err 
	}

	if u.readReceiptPub != nil {
		u.readReceiptPub.PublishReadReceipt(
			domain.ReadReceiptPayload{
				ConversationID: conversationID,
				UserID: userID,
				LastReadMessageID: messageID,
			},
		)
	}

	return nil 
}

func (u *chatUsecase) GetReadStatus(ctx context.Context, conversationID uint64, userID uuid.UUID) (*domain.ConversationRead, error) {
	members, err := u.convRepo.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, domain.ErrNotChatMember
	}

	return u.convRepo.GetReadStatus(ctx, conversationID, userID)
}

func (u *chatUsecase) CreateReadStatus(
	ctx context.Context,
	conversationID uint64,
	userIDs []uuid.UUID,
) error {

	reads := make(
		[]*domain.ConversationRead,
		0,
		len(userIDs),
	)

	for _, userID := range userIDs {

		reads = append(
			reads,
			&domain.ConversationRead{
				ConversationID:        conversationID,
				UserID:                userID,
				LastReadMessageID:     nil,
				LastReceivedMessageID: nil,
				ReadAt:                time.Now(),
			},
		)
	}

	return u.convRepo.CreateReadStatus(
		ctx,
		reads,
	)
}

func (u *chatUsecase) GetConversationDetail(ctx context.Context, conversationID uint64, userID uuid.UUID) (*domain.Conversation, error) {
	members, err := u.convRepo.GetConversationMembers(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isMember := false
	for _, member := range members {
		if member.UserID == userID {
			isMember = true
			break
		}
	}

	if !isMember {
		return nil, domain.ErrNotChatMember
	}

	return u.convRepo.GetConversationDetail(ctx, conversationID)
}

func (u *chatUsecase) SetReadReceiptPublisher(publisher domain.ReadReceiptPublisher) {
	u.readReceiptPub = publisher
}

func (u *chatUsecase) MarkMessageAsDelivered(ctx context.Context, messageID uint64, userID uuid.UUID) error {
	message, err := u.msgRepo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err 
	}

	if message.SenderID == userID {
		return nil 
	}

	members, err := u.convRepo.GetConversationMembers(ctx, message.ConversationID)
	if err != nil {
		return err 
	}

	isMember := false 
	for _, member := range members {
		if member.UserID == userID {
			isMember = true 
			break 
		}
	}

	if !isMember {
		return domain.ErrNotChatMember
	}

	if message.Status >= domain.MessageStatusDelivered {
		return nil 
	}

	return u.msgRepo.UpdateMessageStatus(ctx, messageID, domain.MessageStatusDelivered)
}

func (u *chatUsecase) GetPresenceRecipients(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return u.convRepo.GetPresenceRecipients(ctx, userID)
}

func (u *chatUsecase) GetUserConversationIDs(ctx context.Context, userID uuid.UUID) ([]uint64, error) {
	return u.convRepo.GetUserConversationIDs(ctx, userID)
}