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
	ImageBase64     string `json:"image_base64"`
	IncludeSolution *bool  `json:"include_solution,omitempty"`
}

type SolveLatexRequest struct {
	LatexQuestion string `json:"latex_question"`
	QuestionType  string `json:"question_type"`
	Subject       string `json:"subject"`
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

	includeSolution := true
	if req.IncludeSolution != nil {
		includeSolution = *req.IncludeSolution
	}

	var (
		result *service.VisionResult
		err    error
	)
	if includeSolution {
		result, err = h.aiService.GenerateLatexDraft(req.ImageBase64)
	} else {
		result, err = h.aiService.GenerateQuestionDraft(req.ImageBase64)
	}
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}

	httputil.OK(c, result)
}

func (h *AIHandler) VisionLatexStream(c *gin.Context) {
	var req VisionLatexRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.ImageBase64 == "" {
		httputil.BadRequest(c, "image_base64 required")
		return
	}

	includeSolution := true
	if req.IncludeSolution != nil {
		includeSolution = *req.IncludeSolution
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	streamFn := h.aiService.GenerateLatexDraftStream
	if !includeSolution {
		streamFn = h.aiService.GenerateQuestionDraftStream
	}

	_, err := streamFn(c.Request.Context(), req.ImageBase64, func(evt *service.VisionStreamEvent) error {
		c.SSEvent("progress", evt)
		c.Writer.Flush()
		return nil
	})
	if err != nil {
		c.SSEvent("progress", &service.VisionStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		c.Writer.Flush()
	}
}

func (h *AIHandler) SolveLatex(c *gin.Context) {
	var req SolveLatexRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.LatexQuestion == "" {
		httputil.BadRequest(c, "latex_question required")
		return
	}

	result, err := h.aiService.GenerateSolutionByLatex(c.Request.Context(), req.Subject, req.QuestionType, req.LatexQuestion)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}

	httputil.OK(c, result)
}
