package domain

import (
	"context"
	"github.com/google/uuid"
)

type UserGateway interface {
	VerifyFriendship(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (bool, error)
}