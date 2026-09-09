package http

import (
	"media-service/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MediaHandler struct {
	usecase usecase.MediaUsecase
}

func NewMediaHandler(g *gin.RouterGroup, u usecase.MediaUsecase) {
	handler := &MediaHandler{usecase: u}

	g.POST("/upload", handler.Upload)
}

func (h *MediaHandler) Upload(c *gin.Context) {
	val, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"không tìm thấy thông tin xác thực người dùng"})
		return 
	} 

	uploaderID, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error":"Định danh người dùng không chính xác"})
		return 
	}

	fileHeader, err := c.FormFile("file") 
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error":"không tìm thấy file tải lên"})
		return 
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "không thể mở file để xử lý luồng dữ liệu"})
	}
	defer file.Close()

	media, err := h.usecase.UploadFile(c.Request.Context(), uploaderID, fileHeader.Filename, file, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "tải file lên thành công",
		"data": media,
	})
}