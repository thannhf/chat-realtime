package repository

import (
	"context"
	"errors"
	"fmt"

	"user-service/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type pgUserRepository struct {
	writeDB *gorm.DB
	readDB *gorm.DB 
}

func NewPostgresUserRepository(write, read *gorm.DB) domain.UserRepository {
	return &pgUserRepository{writeDB: write, readDB: read}
}

func (r *pgUserRepository) Create(user *domain.User) error {
	return r.writeDB.Create(user).Error 
}

// read flow
func (r *pgUserRepository) GetByUsername(username string) (*domain.User, error) {
	var user domain.User 
	err := r.readDB.Where("username = ?", username).First(&user).Error 
	return &user, err 
}

func (r *pgUserRepository) GetByID(id uuid.UUID) (*domain.User, error) {
    var user domain.User  
    
    err := r.readDB.Where("id = ?", id).First(&user).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("người dùng không tồn tại") 
        }
        return nil, err
    } 
    
    return &user, nil 
}

// write flow
func (r *pgUserRepository) UpdatePassword(userID uuid.UUID, newHashedPassword string) error {
	result := r.writeDB.Model(&domain.User{}).Where("id = ?", userID).Update("password_hash", newHashedPassword)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("không tìm thấy người dùng để cập nhật mật khẩu")
	}
	return nil 
}

func (r *pgUserRepository) Update(user *domain.User) error {
    return r.writeDB.Model(&domain.User{}).
        Where("id = ?", user.ID).
        Select("Username", "DisplayName", "AvatarURL", "Email"). 
        Updates(user).Error
}

func (r *pgUserRepository) Delete(user *domain.User) error {
	return r.writeDB.Model(&domain.User{}).
		Where("username = ?", user.Username). 
		Delete(&domain.User{}).Error
}

func (r *pgUserRepository) GetAllUser(ctx context.Context) ([]*domain.User, error) {
	var users []*domain.User

	err := r.readDB.WithContext(ctx).Model(&domain.User{}).Find(&users).Error

	if err != nil {
		return nil, err 
	}

	return users, nil 
}

// upload avatar
func (r *pgUserRepository) UpdateAvatar(ctx context.Context, userID string, avatarURL string) error{
	return r.writeDB.WithContext(ctx).Model(&domain.User{}).
			Where("id = ?", userID). 
			Update("avatar_url", avatarURL).Error
}

// check friendship
func (r *pgUserRepository) CheckFriendship(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (bool, error) {
	var count int64

	err := r.readDB.WithContext(ctx).
			Table("friendships").
			Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID).
			Count(&count).Error

	if err != nil {
		return false, err 
	}

	return count > 0, nil 
}

func (r *pgUserRepository) GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	var friendIDs []uuid.UUID

	err := r.readDB.WithContext(ctx).Table("friendships").Where("user_id = ?", userID).Where("status = ?", "accepted").Pluck("friend_id", &friendIDs).Error

	if err != nil {
		return nil, err 
	}

	return friendIDs, nil 
}