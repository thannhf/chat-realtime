package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	UserStatusPending = "pending"
	UserStatusActive  = "active"
	UserStatusBanned  = "banned"
)

const (
	RoleUser      = "user"
	RoleModerator = "moderator"
	RoleAdmin     = "admin"
)

type OTPPurpose string

const (
	OTPPurposeForgotPassword        OTPPurpose = "forgot_password"
	OTPPurposeChangePassword        OTPPurpose = "change_password"
	OTPPurposeConfirmChangePassword OTPPurpose = OTPPurposeChangePassword
	OTPPurposeResetPassword         OTPPurpose = OTPPurposeForgotPassword
	OTPPurposeDeleteUser            OTPPurpose = "delete_user"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Username     string         `json:"username" gorm:"type:varchar(50);uniqueIndex;not null"`
	Email        string         `json:"email" gorm:"type:varchar(50);null"`
	PasswordHash string         `json:"-" gorm:"type:varchar(255);not null"`
	DisplayName  string         `json:"display_name" gorm:"type:varchar(100)"`
	AvatarURL    string         `json:"avatar_url" gorm:"type:varchar(255)"`
	Status       string         `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Role         string         `json:"role" gorm:"type:varchar(20);not null;default:'user'"`
	CreatedAt    time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type UserRepository interface {
	Create(user *User) error
	GetByUsername(username string) (*User, error)
	GetByID(id uuid.UUID) (*User, error)
	UpdatePassword(userID uuid.UUID, newHashedPassword string) error
	Update(user *User) error
	Delete(user *User) error
	GetAllUser(ctx context.Context) ([]*User, error)
	UpdateAvatar(ctx context.Context, userID string, avatarURL string) error
	CheckFriendship(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (bool, error)

	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type UserUseCase interface {
	Register(username, password string) error
	Login(username, password string) (string, string, *User, error)
	Logout(ctx context.Context, username string) error
	ChangePassword(userID uuid.UUID, oldPassword, newPassWord string, inputEmail string) error
	ConfirmChangePassword(username string, userID uuid.UUID, otpCode string, newPassword string) error
	ResetPassword(username string, userID uuid.UUID, otpCode string, newPassword string) error
	GetProfile(ctx context.Context, id string) (*User, error)
	UpdateProfile(ctx context.Context, user *User) error
	ForgotPassword(ctx context.Context, username string) error
	RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	GetAllUser(ctx context.Context) ([]*User, error)
	UploadAvatar(ctx context.Context, userID string, avatarURL string) error
	IsFriend(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (bool, error)

	GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
}

type UserCache interface {
	SetUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	DeleteUser(ctx context.Context, id string) error

	SetOTP(ctx context.Context, userID uuid.UUID, otpCode string, OTPPurpose string, expiration time.Duration) error
	GetOTP(ctx context.Context, userID uuid.UUID, OTPPurpose string) (string, error)
	DeleteOTP(ctx context.Context, userID uuid.UUID, OTPPurpose string) error

	SetRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiration time.Duration) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID) (string, error)
	DeleteRefreshToken(ctx context.Context, userID uuid.UUID) error

	IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}
