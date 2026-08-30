package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PresenceRepository interface {
	SetUserOnline(ctx context.Context, userID uuid.UUID) error
	SetUserOffline(ctx context.Context, userID uuid.UUID) error
	GetUserPresence(ctx context.Context, userID uuid.UUID) (string, error)
	GetUsersPresence(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetUsersLastSeen(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error)
	GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	SetLastSeen(ctx context.Context, userID uuid.UUID, lastSeen time.Time) error
	AddConnection(ctx context.Context, userID uuid.UUID, connectionID string) error
	RemoveConnection(ctx context.Context, userID uuid.UUID, connectionID string) error
	CountConnections(ctx context.Context, userID uuid.UUID) (int64, error)
	RefreshPresence(ctx context.Context, connectionID string) error
}

type PresenceUsecase interface {
	SetUserOnline(ctx context.Context, userID uuid.UUID) error
	SetUserOffline(ctx context.Context, userID uuid.UUID) error
	GetUserPresence(ctx context.Context, userID uuid.UUID) (string, error)
	GetUsersPresence(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error)
	GetUsersLastSeen(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error)
	SetLastSeen(ctx context.Context, userID uuid.UUID, lastSeen time.Time) error
	AddConnection(ctx context.Context, userID uuid.UUID, connectionID string) error
	RemoveConnection(ctx context.Context, userID uuid.UUID, connectionID string) error
	CountConnections(ctx context.Context, userID uuid.UUID) (int64, error)
	RefreshPresence(ctx context.Context, connectionID string) error
}