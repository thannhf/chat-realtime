package external

import (
	"chat-service/config"
	"chat-service/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type userServiceClient struct {
	client *http.Client
	userURL string 
}

func NewUserGateway(cfg *config.Config) domain.UserGateway {
	return &userServiceClient{
		client: &http.Client{Timeout: 3 * time.Second},
		userURL: "http://localhost:8001",
	}
}

type verifyFriendshipResponse struct {
	IsFriend bool `json:"is_friend"`
	Error string `json:"error,omitempty"`
}

func (c *userServiceClient) VerifyFriendship(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/internal/users/%s/friends/%s", c.userURL, userID.String(), friendID.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, nil 
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("không thể kết nối tới user-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("user-service trả về mã lỗi HTTP: %d", resp.StatusCode)
	}

	var res verifyFriendshipResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, fmt.Errorf("lỗi giải mã JSON phản hồi: %w", err)
	}

	return res.IsFriend, nil 
}