package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"common/auth"
	"common/email"
	"user-service/internal/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase struct {
	repo  domain.UserRepository
	cache domain.UserCache
}

func NewUserUsecase(r domain.UserRepository, c domain.UserCache) domain.UserUseCase {
	return &UserUsecase{repo: r, cache: c}
}

// write
func (u *UserUsecase) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("không thể mã hóa mật khẩu: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	return u.repo.Create(user)
}

func (u *UserUsecase) ChangePassword(userID uuid.UUID, oldPassword, newPassWord string, inputEmail string) error {
	user, err := u.repo.GetByID(userID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword))
	if err != nil {
		return errors.New("mật khẩu cũ không chính xác")
	}

	targetEmail := user.Email

	if targetEmail == "" {
		if inputEmail == "" {
			return errors.New("tài khoản của bạn chưa được cấu hình email, vui lòng nhập email xác nhận")
		}

		targetEmail = inputEmail
		user.Email = inputEmail

		if err := u.repo.Update(user); err != nil {
			return err
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otpCode := fmt.Sprintf("%06d", r.Intn(1000000))

	ctx := context.Background()
	err = u.cache.SetOTP(ctx, user.ID, otpCode, string(domain.OTPPurposeChangePassword), 5*time.Minute)
	if err != nil {
		return errors.New("Không thể khởi tạo mã xác thực ngầm")
	}

	go func() {
		_ = email.SendOTPMail(targetEmail, user.Username, otpCode)
	}()

	return nil
}

func (u *UserUsecase) ConfirmChangePassword(username string, userID uuid.UUID, otpCode string, newPassword string) error {
	ctx := context.Background()
	savedOTP, err := u.cache.GetOTP(ctx, userID, string(domain.OTPPurposeConfirmChangePassword))
	if err != nil || savedOTP == "" {
		return errors.New("mã xác thực đã hết hạn hoặc không tồn tại")
	}

	if savedOTP != otpCode {
		return errors.New("Mã xác thực OTP không chính xác")
	}

	user, err := u.repo.GetByUsername(username)
	if err != nil {
		return errors.New("người dùng không tồn tại")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := u.repo.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
		return err
	}

	_ = u.cache.DeleteOTP(ctx, userID, string(domain.OTPPurposeConfirmChangePassword))
	_ = u.cache.DeleteUser(ctx, user.ID.String())

	return nil
}

func (u *UserUsecase) ResetPassword(username string, userID uuid.UUID, otpCode string, newPassword string) error {
	ctx := context.Background()
	savedOTP, err := u.cache.GetOTP(ctx, userID, string(domain.OTPPurposeResetPassword))
	if err != nil || savedOTP == "" {
		return errors.New("mã xác nhận đã hết hạn hoặc không tồn tại")
	}

	if savedOTP != otpCode {
		return errors.New("mã xác nhận OTP không chính xác")
	}

	user, err := u.repo.GetByUsername(username)
	if err != nil {
		return errors.New("người dùng không tồn tại")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := u.repo.UpdatePassword(user.ID, string(hashedPassword)); err != nil {
		return err
	}

	_ = u.cache.DeleteOTP(ctx, userID, string(domain.OTPPurposeResetPassword))
	go func(id string) {
		_ = u.cache.DeleteUser(context.Background(), id)
	}(user.ID.String())

	return nil
}

func (u *UserUsecase) UpdateProfile(ctx context.Context, user *domain.User) error {
	if err := u.repo.Update(user); err != nil {
		return err
	}

	go func(id string) {
		err := u.cache.DeleteUser(context.Background(), id)
		if err != nil {
			log.Printf("[Error] Lỗi xóa cache ngầm cho user %s: %v", id, err)
		}
	}(user.ID.String())

	return nil
}

// read
func (u *UserUsecase) Login(username, password string) (string, string, *domain.User, error) {
	user, err := u.repo.GetByUsername(username)
	
	if err != nil {
		return "", "", nil, errors.New("sai tài khoản hoặc mật khẩu")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", "", nil, errors.New("sai tài khoản hoặc mật khẩu")
	}

	if user.Status == domain.UserStatusBanned {
		return "", "", nil, errors.New("tài khoản của bạn đã bị khóa, vui lòng liên hệ Admin")
	}

	accessToken, err := auth.GenerateToken(user.ID, 15 * time.Minute, user.Role)
    if err != nil {
        return "", "", nil, err 
    }

	refreshToken, err := auth.GenerateToken(user.ID, 7*24*time.Hour, user.Role)
	if err != nil {
		return "", "", nil, err 
	}

	ctx := context.Background()
	err = u.cache.SetRefreshToken(ctx, user.ID, refreshToken, 7*24*time.Hour)

	if err != nil {
		return "", "", nil, errors.New("Không thể khởi tạo phiên làm việc")
	}

    return accessToken, refreshToken, user, nil
}

func (u *UserUsecase) RefreshToken(ctx context.Context, oldRefreshToken string) (string, string, error) {
	claims, err := auth.ValidateToken(oldRefreshToken)
	if err != nil {
		return "", "", errors.New("phiên đăng nhập không hợp lệ hoặc đã hết hạn")
	}

	user, err := u.repo.GetByID(claims.UserID)
	if err != nil {
		return "", "", errors.New("người dùng không tồn tại")
	}

	if user.Status == domain.UserStatusBanned {
		return "", "", errors.New("tài khoản đã bị khóa, không thể gia hạn phiên")
	}

	savedRefreshToken, err := u.cache.GetRefreshToken(ctx, user.ID)
	if err != nil || savedRefreshToken != oldRefreshToken {
		return "", "", errors.New("phiên làm việc không hợp lệ, vui lòng đăng nhập lại")
	}

	newAccessToken, err := auth.GenerateToken(user.ID, 15*time.Minute, user.Role)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := auth.GenerateToken(user.ID, 7*24*time.Hour, user.Role)
	if err != nil {
		return "", "", err
	}

	err = u.cache.SetRefreshToken(ctx, user.ID, newRefreshToken, 7*24*time.Hour)
	if err != nil {
		return "", "", errors.New("không thể cập nhật phiên làm việc mới")
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *UserUsecase) GetProfile(ctx context.Context, id string) (*domain.User, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("định dạng ID không hợp lệ: %v", err)
	}

	cacheCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	user, err := u.cache.GetUser(cacheCtx, id)
	if err == nil && user != nil {
		return user, nil
	}

	if err != nil {
		log.Printf("[Warning] Khởi động chế độ Fallback, lỗi Redis: %v", err)
	}

	user, err = u.repo.GetByID(parsedID)
	if err != nil {
		return nil, err
	}

	go func(uData domain.User) {
		err := u.cache.SetUser(context.Background(), &uData)
		if err != nil {
			log.Printf("[Error] Không thể nạp lại cache ngầm cho user %s: %v", uData.ID, err)
		}
	}(*user)

	return user, nil
}

func (u *UserUsecase) ForgotPassword(ctx context.Context, username string) error {
	user, err := u.repo.GetByUsername(username)
	if err != nil {
		return errors.New("người dùng không tồn tại")
	}

	if user.Email == "" {
		return errors.New("This account do not set config email verify")
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	otpCode := fmt.Sprintf("%06d", r.Intn(1000000))

	err = u.cache.SetOTP(ctx, user.ID, otpCode, string(domain.OTPPurposeForgotPassword), 5*time.Minute)
	if err != nil {
		return errors.New("Không thể tạo mã xác thực ngầm")
	}

	go func() {
		errMail := email.SendOTPMail(user.Email, user.Username, otpCode)
		if errMail != nil {
			fmt.Printf("Lỗi gửi email cho %s: %v\n", user.Email, errMail)
		} else {
			fmt.Printf("Đã gửi thành công mã OTP về hòm thư: %s\n", user.Email)
		}
	}()

	return nil
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := u.repo.GetByID(id)
	if err != nil {
		return errors.New("người dùng không tồn tại")
	}

	err = u.repo.Delete(user)
	if err != nil {
		return errors.New("không thể xóa user khỏi DB")
	}

	_ = u.cache.DeleteUser(ctx, user.ID.String())
	_ = u.cache.DeleteRefreshToken(ctx, user.ID)
	_ = u.cache.DeleteOTP(ctx, user.ID, string(domain.OTPPurposeDeleteUser))

	return nil
}

func (u *UserUsecase) GetAllUser(ctx context.Context) ([]*domain.User, error) {
	users, err := u.repo.GetAllUser(ctx)
	if err != nil {
		return nil, errors.New("không thể lấy danh sách user")
	}
	return users, nil
}

// upload avatar
func (u *UserUsecase) UploadAvatar(ctx context.Context, userID string, avatarURL string) error {
	err := u.repo.UpdateAvatar(ctx, userID, avatarURL)
	if err != nil {
		return errors.New("không thể update face id vào db")
	}

	_ = u.cache.DeleteUser(ctx, userID)

	return nil
}

// logout
func (u *UserUsecase) Logout(ctx context.Context, userID uuid.UUID) error {
	err := u.cache.DeleteRefreshToken(ctx, userID)
	if err != nil {
		return errors.New("Không thể xóa phiên đăng nhập, vui lòng thử lại")
	}
	return nil
}

// friendship
func (u *UserUsecase) IsFriend(ctx context.Context, userID uuid.UUID, friendID uuid.UUID) (bool, error) {
	if userID == friendID {
		return false, nil
	}
	return u.repo.CheckFriendship(ctx, userID, friendID)
}

func (u *UserUsecase) GetFriendIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return u.repo.GetFriendIDs(ctx, userID)
}
