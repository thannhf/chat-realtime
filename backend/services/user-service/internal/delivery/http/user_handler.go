package http

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"user-service/config"
	"user-service/internal/delivery/http/dto"
	"user-service/internal/delivery/http/middleware"
	"user-service/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	usecase domain.UserUseCase
	cfg     *config.Config
}

func NewUserHandler(r *gin.Engine, uc domain.UserUseCase, cache domain.UserCache, cfg *config.Config) {
	h := &UserHandler{
		usecase: uc,
		cfg:     cfg,
	}

	// public
	api := r.Group("/api/v1/users")
	{
		api.POST("/register", middleware.RateLimiter(cache, 5, 1*time.Minute), h.Register)
		api.POST("/login", middleware.RateLimiter(cache, 3, 1*time.Minute), h.Login)
		api.POST("/reset-password", h.ResetPassword)
		api.POST("/forgot-password", h.ForgotPassword)
		api.POST("/refresh-token", h.RefreshToken)
	}

	// login
	secured := api.Group("/")
	secured.Use(middleware.AuthMiddleware())
	{
		secured.POST("/logout", h.Logout)
		secured.GET("", h.GetUsersByIDs)
		secured.GET("/profile", h.GetProfile)
		secured.PUT("/profile", h.UpdateProfile)
		secured.POST("/change-password", h.ChangePassword)
		secured.POST("/confirm-change", h.ConfirmChangePassword)
	}

	r.Static("/uploads", "./uploads")
	secured.Use(middleware.AuthMiddleware())
	{
		secured.POST("/avatar/upload", middleware.RateLimiter(cache, 2, 1*time.Minute), h.UploadAvatar)
	}

	// super admin
	adminRoute := r.Group("/api/v1/admin")
	adminRoute.Use(middleware.AuthMiddleware())
	{
		adminRoute.DELETE("/users/:id", middleware.RequireRole(domain.RoleAdmin), h.DeleteUser)
		adminRoute.GET("/users", middleware.RequireRole(domain.RoleAdmin, domain.RoleModerator), h.GetAllUser)
	}

	// private
	internal := r.Group("/api/v1/internal")
	{
		internal.GET("/users/:user_id/friends/:friend_id", h.VerifyFriendship)
		internal.GET("/users/:user_id/friends", h.GetFriendIDs)
	}
}

func (h *UserHandler) Register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=30"`
		Password string `json:"password" binding:"required,min=8,max=72"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data không hợp lệ"})
		return
	}

	err := h.usecase.Register(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "đăng ký thành công"})
}

func (h *UserHandler) Login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required,min=3,max=30"`
		Password string `json:"password" binding:"required,min=8,max=72"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "thiếu data"})
		return
	}

	accessToken, refreshToken, user, err := h.usecase.Login(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "sai tài khoản hoặc mật khẩu"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"avatar_url": user.AvatarURL,
		},
	},
	)
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "thiếu mã refresh_token",
		})
		return
	}

	newAccessToken, newRefreshToken, err := h.usecase.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

func (h *UserHandler) ChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var input struct {
		OldPassword string `json:"oldPassword" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
		Email       string `json:"email" binding:"omitempty,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu không hợp lệ hoặc thiếu dữ liệu"})
		return
	}

	err := h.usecase.ChangePassword(userID, input.OldPassword, input.NewPassword, input.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xác thực bước 1 thành công! Mã OTP đã được gửi về Email của bạn để hoàn tất thay đổi."})
}

func (h *UserHandler) ConfirmChangePassword(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	user, _ := h.usecase.GetProfile(c.Request.Context(), userID.String())

	var input struct {
		Otp         string `json:"otp" binding:"required,len=6"`
		NewPassword string `json:"newPassword" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vui lòng nhập mã OTP xác thực"})
		return
	}

	err := h.usecase.ConfirmChangePassword(user.Username, userID, input.Otp, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Đổi password thành công!"})
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Username    string `json:"username" binding:"required"`
		Otp         string `json:"otp" binding:"required,len=6"`
		NewPassWord string `json:"newPassword" binding:"required,min=8"`
	}

	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "data không hợp lệ"})
		return
	}

	err := h.usecase.ResetPassword(input.Username, userID, input.Otp, input.NewPassWord)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mật khẩu đã được đặt lại thành công"})
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	user, err := h.usecase.GetProfile(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Lấy thông tin cấu hình thành công",
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"created_at": user.CreatedAt,
		},
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uuid.UUID)

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dữ liệu cập nhật không hợp lệ"})
		return
	}

	userUpdate := &domain.User{
		ID:       userID,
		Username: req.Username,
		Email:    req.Email,
	}

	err := h.usecase.UpdateProfile(c.Request.Context(), userUpdate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cập nhật thông tin cá nhân và Email thành công",
	})
}

func (h *UserHandler) ForgotPassword(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vui lòng nhập tên tài khoản"})
		return
	}

	err := h.usecase.ForgotPassword(c.Request.Context(), input.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mã xác thực OTP đã được gửi (Hãy kiểm tra terminal backend để lấy mã)"})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	var delUser struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&delUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dữ liệu yêu cầu không hợp lệ"})
		return
	}

	err := h.usecase.DeleteUser(c.Request.Context(), delUser.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Xóa user thành công"})
}

func (h *UserHandler) GetAllUser(c *gin.Context) {
	users, err := h.usecase.GetAllUser(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

// upload avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "yêu cầu không được xác thực"})
		return
	}
	userID := userIDVal.(uuid.UUID).String()

	oldUser, err := h.usecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "không tìm thấy user",
		})
		return 
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "không tìm thấy file"})
		return
	}

	var maxFileSize int64 = 5 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "kích thước file quá lớn"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "định dạng file không hợp lệ"})
		return
	}

	uniqueID := uuid.New().String()
	newFilename := fmt.Sprintf("avatar_%s_%d%s", uniqueID, time.Now().Unix(), ext)

	uploadDir := "./uploads/images/avatars"
	filePath := filepath.Join(uploadDir, newFilename)

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "không thể tạo thư mục lưu avatar",
		})
		return
	}

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không thể lưu file trên hệ thống server"})
		return
	}

	avatarURL := fmt.Sprintf("%s/uploads/images/avatars/%s", h.cfg.GatewayURL.URL, newFilename)

	err = h.usecase.UploadAvatar(c.Request.Context(), userID, avatarURL)
	if err != nil {
		_ = os.Remove(filePath)
		
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if oldUser.AvatarURL != "" {
		oldAvatarPath := strings.TrimPrefix(oldUser.AvatarURL, h.cfg.GatewayURL.URL)

		oldAvatarPath = "." + oldAvatarPath

		if err := os.Remove(oldAvatarPath); err != nil && !os.IsNotExist(err) {
			log.Printf("lỗi xóa file cũ")
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Tải ảnh đại diện thành công",
		"avatar_url": avatarURL,
	})
}

// logout
func (h *UserHandler) Logout(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "yêu cầu không được xác thực"})
		return
	}
	userID := userIDVal.(uuid.UUID)

	err := h.usecase.Logout(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Đăng xuất thành công, phiên làm việc được hủy",
	})
}

// friendship
func (h *UserHandler) VerifyFriendship(c *gin.Context) {
	userID, err1 := uuid.Parse(c.Param("user_id"))
	friendID, err2 := uuid.Parse(c.Param("friend_id"))

	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"is_friend": false, "error": "ID user không hợp lệ"})
		return
	}

	isFriend, err := h.usecase.IsFriend(c.Request.Context(), userID, friendID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"is_friend": false, "error": "Lỗi kiểm tra hệ thống"})
		return
	}

	if !isFriend {
		c.JSON(http.StatusOK, gin.H{"is_friend": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_friend": true})
}

func (h *UserHandler) GetFriendIDs(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID user không hợp lệ"})
		return
	}

	friendIDs, err := h.usecase.GetFriendIDs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Lỗi lấy danh sách bạn bè"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": friendIDs,
	})
}

func (h *UserHandler) GetUsersByIDs(c *gin.Context) {
	idsParam := c.Query("ids")

	if idsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids không được để trống"})
		return
	}

	rawIDs := strings.Split(idsParam, ",")
	userIDs := make([]uuid.UUID, 0, len(rawIDs))

	for _, rawID := range rawIDs {
		id, err := uuid.Parse(strings.TrimSpace(rawID))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id không hợp lệ"})
			return
		}

		userIDs = append(userIDs, id)
	}

	users, err := h.usecase.GetUserByIDs(c.Request.Context(), userIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Không thể lấy thông tin users",
		})
		return
	}

	responses := make([]dto.PublicUserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, dto.PublicUserResponse{
			ID:        user.ID,
			Username:  user.Username,
			AvatarURL: &user.AvatarURL,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}
