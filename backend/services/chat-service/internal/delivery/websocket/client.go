package websocket

import (
	"bytes"
	"chat-service/internal/domain"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 5120
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Hub          *Hub
	ConnectionID string
	Conn         *websocket.Conn
	UserID       uuid.UUID
	Send         chan []byte
}

// readpump
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {
		log.Printf("Received Pong from user %s", c.UserID)
		_ = c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		ctx := context.Background()

		if c.UserID != uuid.Nil {
			if err := c.Hub.PresenceUsecase.RefreshPresence(ctx, c.ConnectionID); err != nil {
				log.Printf("Refresh presence failed for user %s: %v", c.UserID, err)
			}
		} else {
			log.Printf("Refresh presence skipped: UserID is nil")
		}

		// _ = c.Hub.PresenceUsecase.SetUserOnline(ctx, c.UserID)
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Lỗi kết nối đột ngột của User %s: %v", c.UserID, err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		c.Hub.Broadcast <- domain.MessageWrapper{
			SenderID: c.UserID,
			Payload:  message,
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		// c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				_ = c.Conn.WriteMessage(
					websocket.CloseMessage,
					[]byte{},
				)
				return
			}

			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				return
			}
		case <-ticker.C:
			log.Printf("Sending Ping to user %s", c.UserID)
			_ = c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
