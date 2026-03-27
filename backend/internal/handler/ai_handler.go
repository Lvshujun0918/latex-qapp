package handler

import (
	"latex-qapp/backend/internal/service"
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiService *service.AIService
}

type VisionLatexRequest struct {
	ImageBase64 string `json:"image_base64"`
}

func NewAIHandler(aiService *service.AIService) *AIHandler {
	return &AIHandler{aiService: aiService}
}

func (h *AIHandler) VisionLatex(c *gin.Context) {
	var req VisionLatexRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ImageBase64 == "" {
		httputil.BadRequest(c, "image_base64 required")
		return
	}

	result, err := h.aiService.GenerateLatexDraft(req.ImageBase64)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}

	httputil.OK(c, result)
}
