package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"user-service/internal/domain"
	"user-service/internal/usecase"
	"user-service/internal/mocks" // Import thư mục mocks bạn vừa tạo

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Trường hợp 1: Cache Hit (Dữ liệu có sẵn trong Redis -> Trả về luôn, không xuống DB)
func TestGetProfile_CacheHit(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockCache := new(mocks.MockUserCache)

	userID := uuid.New()
	mockUser := &domain.User{ID: userID, Username: "Sunny Backend"}

	// Giả lập: Khi gọi GetUser trong cache, trả về mockUser và không có lỗi
	mockCache.On("GetUser", mock.Anything, userID.String()).Return(mockUser, nil)

	uc := usecase.NewUserUsecase(mockRepo, mockCache)
	result, err := uc.GetProfile(context.Background(), userID.String())

	assert.NoError(t, err)
	assert.Equal(t, mockUser.Username, result.Username)
	
	// Đảm bảo GetByID của DB không bao giờ bị gọi
	mockRepo.AssertNotCalled(t, "GetByID", mock.Anything)
}

// Trường hợp 2: Fallback (Redis lỗi/sập -> Hệ thống vẫn phải lấy được từ DB)
func TestGetProfile_RedisFailure_FallbackToDB(t *testing.T) {
	mockRepo := new(mocks.MockUserRepository)
	mockCache := new(mocks.MockUserCache)

	userID := uuid.New()
	mockUser := &domain.User{ID: userID, Username: "Sunny Backend"}

	// Giả lập: Redis bị lỗi kết nối (Timeout hoặc sập)
	mockCache.On("GetUser", mock.Anything, userID.String()).Return(nil, errors.New("redis connection refused"))
	
	// Giả lập: DB vẫn hoạt động tốt và trả về thông tin
	mockRepo.On("GetByID", userID).Return(mockUser, nil)
	
	// Giả lập: Việc nạp lại vào cache chạy ngầm (mình cho phép chạy qua)
	mockCache.On("SetUser", mock.Anything, mock.Anything).Return(nil)

	uc := usecase.NewUserUsecase(mockRepo, mockCache)
	result, err := uc.GetProfile(context.Background(), userID.String())

	// Kết quả mong muốn: App không bị sập, vẫn trả về dữ liệu từ DB thành công
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Sunny Backend", result.Username)

	// Đợi một chút ngắn vì SetUser chạy ngầm bằng goroutine `go func()`
	time.Sleep(10 * time.Millisecond)
	
	mockRepo.AssertExpectations(t)
}
