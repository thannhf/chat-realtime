package http

import (
	"chat-service/internal/delivery/websocket"
	"chat-service/internal/domain"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
)

type ChatHandler struct {
	usecase         domain.ChatUsecase
	presenceUsecase domain.PresenceUsecase
	hub             *websocket.Hub
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewChatHandler(g *gin.Engine, usecase domain.ChatUsecase, presenceUsecase domain.PresenceUsecase, hub *websocket.Hub) {
	handler := &ChatHandler{
		usecase:         usecase,
		presenceUsecase: presenceUsecase,
		hub:             hub,
	}

	v1 := g.Group("/api/v1/chat")
	{
		v1.GET("/ws", handler.ConnectWebSocket)
		v1.POST("/conversations", handler.CreateConversation)
		v1.GET("/history/:conversation_id", handler.FetchHistory)
		v1.GET("/presence/:user_id", handler.GetPresence)
		v1.GET("/get_conversations", handler.GetConversations)
		v1.PATCH("/conversations/:id/read", handler.MarkAsRead)
		v1.GET("/conversations/:id", handler.GetConversationDetail)
		v1.GET("/presence", handler.GetUsersPresence)
		v1.GET("/last-seen/:user_id", handler.GetLastSeen)
		v1.GET("/last-seen", handler.GetUsersLastSeen)
	}
}

func (h *ChatHandler) GetConversations(c *gin.Context) {
	val, _ := c.Get("userID")
	userID := val.(uuid.UUID)

	conversation, err := h.usecase.GetUserConversations(c.Request.Context(), userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"data": conversation,
	})
}

func (h *ChatHandler) GetConversationDetail(c *gin.Context) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 
	}

	userID := userIDValue.(uuid.UUID)

	conversationID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid conversation id"})
		return 
	}

	conversation, err := h.usecase.GetConversationDetail(c.Request.Context(), conversationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": conversation,
	})
}

func (h *ChatHandler) MarkAsRead(c *gin.Context) {
	type MarkAsReadRequest struct {
		MessageID uint64 `json:"message_id"`
	}
	userIDValue, exists := c.Get("userID")
	if !exists {
        c.JSON(
            http.StatusUnauthorized,
            gin.H{"error":"unauthorized"},
        )
        return
    }

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"không tìm thấy thông tin xác thực"})
		return
	}

	var input MarkAsReadRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	conversationIDParam := c.Param("id")
	conversationID, err := strconv.ParseUint(conversationIDParam, 10, 64)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation id không hợp lệ"})
		return
	}

	err = h.hub.ChatUsecase.MarkAsRead(c.Request.Context(), conversationID, userID, input.MessageID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	c.JSON(http.StatusOK, gin.H{"message": "đã đánh dấu cuộc trò chuyện là đã đọc."})
}

func (h *ChatHandler) CreateConversation(c *gin.Context) {
	val, _ := c.Get("userID")
	senderID := val.(uuid.UUID)

	var input domain.CreateConversationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	conv, err := h.usecase.CreateConversation(c.Request.Context(), senderID, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return 
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "tạo cuộc trò chuyện thành công",
		"conversation": conv,
	})
}

func (h *ChatHandler) ConnectWebSocket(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "không tìm thấy thông tin xác thực"})
		return
	}
	userID := val.(uuid.UUID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		_ = c.Error(err)
		return
	}

	client := &websocket.Client{
		Hub:          h.hub,
		ConnectionID: uuid.NewString(),
		Conn:         conn,
		UserID:       userID,
		Send:         make(chan []byte, 256),
	}
	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *ChatHandler) FetchHistory(c *gin.Context) {
	val, _ := c.Get("userID")
	userID := val.(uuid.UUID)

	convID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID cuộc hội thoại không hợp lệ"})
		return
	}

	var lastMsgID *uint64 

	if value := c.Query("last_message_id"); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":"last_message_id không hợp lệ"})
			return 
		}
		lastMsgID = &id 
	}

	limit := 50
	if value := c.Query("limit"); value != "" {
		parsedLimit, err := strconv.Atoi(value)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	messages, err := h.usecase.GetChatHistory(c.Request.Context(), convID, userID, lastMsgID, limit)
	if err != nil {
		if err.Error() == "Bạn không phải là thành viên cuộc hội thoại này" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống"})
		return
	}

	var nextCursor *uint64 
	if len(messages) > 0 {
		id := messages[0].ID
		nextCursor = &id 
	}
	c.JSON(http.StatusOK, gin.H{
		"data": messages,
		"next_cursor": nextCursor,
	})
}

func (h *ChatHandler) GetPresence(c *gin.Context) {
	targetUserIDStr := c.Param("user_id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID người dùng không hợp lệ"})
		return 
	}

	status, err := h.presenceUsecase.GetUserPresence(c.Request.Context(), targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi hệ thống khi lấy trạng thái kết nối"})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": targetUserID.String(),
		"status": status,
	})
}

func (h *ChatHandler) GetUsersPresence(c *gin.Context) {
	userIDsParam := c.Query("user_ids")

	if userIDsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_ids không được để trống",
		})
		return 
	}

	rawUserIDs := strings.Split(userIDsParam, ",")
	userIDs := make([]uuid.UUID, 0, len(rawUserIDs))

	for _, rawID := range rawUserIDs {
		userID, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":"user_id không hợp lệ",
			})
			return
		}

		userIDs = append(userIDs, userID)
	}

	presence, err := h.presenceUsecase.GetUsersPresence(c.Request.Context(), userIDs)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "không thể lấy trạng thái presence",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"data": presence,
	})
}

func (h *ChatHandler) GetLastSeen(c *gin.Context) {
	userIDParam := c.Param("user_id")
	userID, err := uuid.Parse(userIDParam)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return 
	}

	lastSeen, err := h.presenceUsecase.GetLastSeen(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get last seen",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"last_seen": lastSeen,
	})
}

func (h *ChatHandler) GetUsersLastSeen(c *gin.Context) {
	userIDsParam := c.Query("user_ids")

	if userIDsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_ids is required",
		})
		return 
	}

	rawUserIDs := strings.Split(userIDsParam, ",")
	userIDs := make([]uuid.UUID, 0, len(rawUserIDs))

	for _, rawID := range rawUserIDs {
		userID, err := uuid.Parse(strings.TrimSpace(rawID))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid user id",
			})
			return 
		}
		userIDs = append(userIDs, userID)
	}

	lastSeen, err := h.presenceUsecase.GetUsersLastSeen(
		c.Request.Context(),
		userIDs,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get users last seen",
		})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"last_seen": lastSeen,
	})
}