package repository

import (
	"context"
	"encoding/json"
	"time"

	"user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type userCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) domain.UserCache {
	return &userCache{client: client}
}

const otpKeyPrefix = "user:otp:"

func (c *userCache) SetOTP(ctx context.Context, userID uuid.UUID, otpCode string, OTPPurpose string, expiration time.Duration) error {
	if c.client == nil { return nil }
	
	key := otpKeyPrefix + userID.String() + ":" + OTPPurpose
	return c.client.Set(ctx, key, otpCode, expiration).Err()
}

func (c *userCache) GetOTP(ctx context.Context, userID uuid.UUID, OTPPurpose string) (string, error) {
	if c.client == nil { 
		return "", nil 
	}

	key := otpKeyPrefix + userID.String() + ":" + OTPPurpose
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil 
		}
		return "", err 
	}
	return val, nil
}

func (c *userCache) DeleteOTP(ctx context.Context, userID uuid.UUID, OTPPurpose string) error {
	if c.client == nil { 
		return nil 
	}
	
	key := otpKeyPrefix + userID.String() + ":" + OTPPurpose
	return c.client.Del(ctx, key).Err()
}

func (c *userCache) SetUser(ctx context.Context, user *domain.User) error {
	if c.client == nil { 
		return nil 
	}
	
	key := "user:profile:" + user.ID.String()
	data, _ := json.Marshal(user)
	return c.client.Set(ctx, key, data, time.Hour).Err()
}

func (c *userCache) GetUser(ctx context.Context, id string) (*domain.User, error) {
	if c.client == nil { return nil, nil }
	key := "user:profile:" + id
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err 
	}
	var user domain.User
	json.Unmarshal([]byte(val), &user)
	return &user, nil
}

func (c *userCache) DeleteUser(ctx context.Context, id string) error {
	if c.client == nil { return nil }
	return c.client.Del(ctx, "user:profile:"+id).Err()
}

const refreshKeyPrefix = "user:refresh:"

func (c *userCache) SetRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiration time.Duration) error {
	if c.client == nil {
		return nil
	}
	
	key := refreshKeyPrefix + userID.String()
	return c.client.Set(ctx, key, token, expiration).Err()
}

func (c *userCache) GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	if c.client == nil {
		return "", nil 
	}

	key := refreshKeyPrefix + userID.String()
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil 
		}
		return "", err 
	}
	return val, nil 
}

func (c *userCache) DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error {
	if c.client == nil {
		return nil 
	}

	key := refreshKeyPrefix + userID.String()
	return c.client.Del(ctx, key).Err()
}

// rate limit
const rateLimitPrefix = "rate_limit:"

func (c *userCache) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if c.client == nil {
		return false, nil 
	}

	redisKey := rateLimitPrefix + key 

	count, err := c.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err 
	}

	if count == 1 {
		c.client.Expire(ctx, redisKey, window)
	}

	if int(count) > limit {
		return true, nil 
	}

	return false, nil 
}