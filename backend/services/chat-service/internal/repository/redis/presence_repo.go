package redis

import (
	"chat-service/internal/domain"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	presencePrefix            = "user:presence:%s"
	lastSeenPrefix            = "user:last_seen:%s"
	multipleConnectionsPrefix = "user:connections:%s"
	connectionPrefix          = "user:connection:%s"
	presenceTTL               = 60 * time.Second
)

type presenceRepository struct {
	redisClient *redis.Client
}

func NewPresenceRepository(redisClient *redis.Client) domain.PresenceRepository {
	return &presenceRepository{redisClient: redisClient}
}

func (r *presenceRepository) SetUserOnline(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf(presencePrefix, userID.String())
	return r.redisClient.Set(ctx, key, "online", presenceTTL).Err()
}

func (r *presenceRepository) SetUserOffline(ctx context.Context, userID uuid.UUID) error {
	presenceKey := fmt.Sprintf(presencePrefix, userID.String())
	lastSeenKey := fmt.Sprintf(lastSeenPrefix, userID.String())

	now := time.Now().UTC()
	pipe := r.redisClient.TxPipeline()

	pipe.Del(ctx, presenceKey)
	pipe.Set(ctx, lastSeenKey, now.Format(time.RFC3339), 0)

	_, err := pipe.Exec(ctx)

	return err
}

func (r *presenceRepository) GetUserPresence(ctx context.Context, userID uuid.UUID) (string, error) {
	key := fmt.Sprintf(presencePrefix, userID)
	val, err := r.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "offline", nil
	} else if err != nil {
		return "offline", err
	}
	return val, nil
}

func (r *presenceRepository) GetUsersPresence(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	if len(userIDs) == 0 {
		return result, nil
	}

	keys := make([]string, 0, len(userIDs))

	for _, userID := range userIDs {
		keys = append(keys, fmt.Sprintf(presencePrefix, userID))
	}

	values, err := r.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, value := range values {
		userID := userIDs[i]

		if value == nil {
			result[userID] = "offline"
			continue
		}
		result[userID] = value.(string)
	}

	return result, nil
}

func (r *presenceRepository) GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	key := fmt.Sprintf(lastSeenPrefix, userID.String())

	value, err := r.redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	lastSeen, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, err
	}

	return &lastSeen, nil
}

func (r *presenceRepository) GetUsersLastSeen(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error) {
	result := make(map[uuid.UUID]*time.Time)

	if len(userIDs) == 0 {
		return result, nil
	}

	keys := make([]string, 0, len(userIDs))

	for _, userID := range userIDs {
		keys = append(keys, fmt.Sprintf(lastSeenPrefix, userID.String()))
	}

	values, err := r.redisClient.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	for i, value := range values {
		userID := userIDs[i]

		if value == nil {
			result[userID] = nil
			continue
		}

		lastSeen, err := time.Parse(time.RFC3339, value.(string))
		if err != nil {
			return nil, err
		}

		result[userID] = &lastSeen
	}
	return result, nil
}

func (r *presenceRepository) SetLastSeen(ctx context.Context, userID uuid.UUID, lastSeen time.Time) error {
	key := fmt.Sprintf(lastSeenPrefix, userID.String())

	return r.redisClient.Set(ctx, key, lastSeen.UTC().Format(time.RFC3339Nano), 0).Err()
}

func (r *presenceRepository) AddConnection(ctx context.Context, userID uuid.UUID, connectionID string) error {
	connectionKey := fmt.Sprintf(connectionPrefix, connectionID)

	connectionsKey := fmt.Sprintf(multipleConnectionsPrefix, userID.String())

	pipe := r.redisClient.TxPipeline()
	pipe.SAdd(ctx, connectionsKey, connectionID)
	pipe.Set(ctx, connectionKey, "online", presenceTTL)

	_, err := pipe.Exec(ctx)

	return err
}

func (r *presenceRepository) RemoveConnection(ctx context.Context, userID uuid.UUID, connectionID string) error {
	connectionsKey := fmt.Sprintf(multipleConnectionsPrefix, userID.String())

	connectionKey := fmt.Sprintf(connectionPrefix, connectionID)

	pipe := r.redisClient.TxPipeline()
	pipe.SRem(ctx, connectionsKey, connectionID)
	pipe.Del(ctx, connectionKey)

	_, err := pipe.Exec(ctx)

	return err 
}

func (r *presenceRepository) CountConnections(ctx context.Context, userID uuid.UUID) (int64, error) {
	key := fmt.Sprintf(multipleConnectionsPrefix, userID.String())

	connections, err := r.redisClient.SMembers(ctx, key).Result()
	if err != nil {
		return 0, err 
	}

	var count int64 
	for _, connectionID := range connections {
		connectionKey := fmt.Sprintf(connectionPrefix, connectionID)

		exists, err := r.redisClient.Exists(ctx, connectionKey).Result()
		if err != nil {
			return 0, err 
		}

		if exists == 0 {
			if err := r.redisClient.SRem(ctx, key, connectionID).Err(); err != nil {
				return 0, err
			}
			continue
		}
		count++
	}

	// count, err = r.redisClient.SCard(ctx, key).Result()
	// if err != nil {
	// 	return 0, err
	// }

	return count, err
}

func (r *presenceRepository) RefreshPresence(ctx context.Context, connectionID string) error {
	key := fmt.Sprintf(connectionPrefix, connectionID)

	ok, err := r.redisClient.Expire(ctx, key, presenceTTL).Result()
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf(
			"connection %s does not exist",
			connectionID,
		)
	}

	return nil
}
