package usecase

import (
	"chat-service/internal/domain"
	"context"
	"fmt"
	"time"


	"github.com/google/uuid"
)

type presenceUsecase struct {
	presenceRepo domain.PresenceRepository
}

func NewPresenceUsecase(repo domain.PresenceRepository) domain.PresenceUsecase {
	return &presenceUsecase{presenceRepo: repo}
}

func (u *presenceUsecase) SetUserOnline(ctx context.Context, userID uuid.UUID) error {
	return u.presenceRepo.SetUserOnline(ctx, userID)
}

func (u *presenceUsecase) SetUserOffline(ctx context.Context, userID uuid.UUID) error {
	return u.presenceRepo.SetUserOffline(ctx, userID)
}

func (u *presenceUsecase) GetUserPresence(ctx context.Context, userID uuid.UUID) (string, error) {
	return u.presenceRepo.GetUserPresence(ctx, userID)
}

func (u *presenceUsecase) GetUsersPresence(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	return u.presenceRepo.GetUsersPresence(ctx, userIDs)
}

func (u *presenceUsecase) GetLastSeen(ctx context.Context, userID uuid.UUID) (*time.Time, error) {
	return u.presenceRepo.GetLastSeen(ctx, userID)
}

func (u *presenceUsecase) GetUsersLastSeen(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*time.Time, error) {
	return u.presenceRepo.GetUsersLastSeen(ctx, userIDs)
}

func (u *presenceUsecase) SetLastSeen(ctx context.Context, userID uuid.UUID, lastSeen time.Time) error {
	return u.presenceRepo.SetLastSeen(ctx, userID, lastSeen)
}

func (u *presenceUsecase) AddConnection(ctx context.Context, userID uuid.UUID, connectionID string) error {
	return u.presenceRepo.AddConnection(ctx, userID, connectionID)
}

func (u *presenceUsecase) RemoveConnection(ctx context.Context, userID uuid.UUID, connectionID string) error {
	return u.presenceRepo.RemoveConnection(ctx, userID, connectionID)
}

func (u *presenceUsecase) CountConnections(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := u.presenceRepo.CountConnections(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("count connections: %w", err)
	}

	return count, nil
}

func (u *presenceUsecase) RefreshPresence(ctx context.Context, connectionID string) error {
	return u.presenceRepo.RefreshPresence(ctx, connectionID)
}