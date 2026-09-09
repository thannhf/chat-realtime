package http

import (
	"media-service/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type InternalMediaHandler struct {
	usecase usecase.MediaUsecase
}

func NewInternalMediaHandler(g *gin.RouterGroup, u usecase.MediaUsecase) {
	handler := &InternalMediaHandler{usecase: u}

	g.GET("/media/:id", handler.VerifyMedia)
}

func (h *InternalMediaHandler) VerifyMedia(c *gin.Context) {
	mediaIDStr := c.Param("id")
	mediaID, err := uuid.Parse(mediaIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID file không đúng định dạng UUID"})
		return 
	}

	media, err := h.usecase.GetMediaByID(c.Request.Context(), mediaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Không tìm thấy file hoặc file đã bị xóa"})
		return 
	}

	c.JSON(http.StatusOK, gin.H{
		"id": media.ID,
		"uploader_id": media.UploaderID,
		"file_url": media.FileURL,
		"file_type": media.FileType,
	})
}