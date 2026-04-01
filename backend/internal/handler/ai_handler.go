package handler

import (
	"encoding/json"
	"strings"

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
	LatexSource   string `json:"latex_source"`
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
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	latexQuestion := resolveSolveLatexInput(req)
	if latexQuestion == "" {
		httputil.BadRequest(c, "latex_question or latex_source required")
		return
	}

	result, err := h.aiService.GenerateSolutionByLatex(c.Request.Context(), req.Subject, req.QuestionType, latexQuestion)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}

	httputil.OK(c, result)
}

func (h *AIHandler) SolveLatexStream(c *gin.Context) {
	var req SolveLatexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	latexQuestion := resolveSolveLatexInput(req)
	if latexQuestion == "" {
		httputil.BadRequest(c, "latex_question or latex_source required")
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	_, err := h.aiService.GenerateSolutionByLatexStream(c.Request.Context(), req.Subject, req.QuestionType, latexQuestion, func(evt *service.SolveStreamEvent) error {
		c.SSEvent("progress", evt)
		c.Writer.Flush()
		return nil
	})
	if err != nil {
		c.SSEvent("progress", &service.SolveStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		c.Writer.Flush()
	}
}

func resolveSolveLatexInput(req SolveLatexRequest) string {
	if strings.TrimSpace(req.LatexQuestion) != "" {
		return strings.TrimSpace(req.LatexQuestion)
	}

	raw := strings.TrimSpace(req.LatexSource)
	if raw == "" {
		return ""
	}

	var out service.LatexOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return raw
	}

	stem := strings.TrimSpace(out.Stem)
	if stem == "" {
		return ""
	}

	if strings.TrimSpace(out.QuestionType) == "选择" && len(out.Options) > 0 {
		items := make([]string, 0, len(out.Options))
		for _, opt := range out.Options {
			clean := strings.TrimSpace(opt)
			if clean != "" {
				items = append(items, "\\item "+clean)
			}
		}
		if len(items) > 0 {
			return stem + "\n\\begin{choices}\n" + strings.Join(items, "\n") + "\n\\end{choices}"
		}
	}

	if strings.TrimSpace(out.QuestionType) == "解答" && len(out.SubQuestions) > 0 {
		items := make([]string, 0, len(out.SubQuestions))
		for _, sub := range out.SubQuestions {
			clean := strings.TrimSpace(sub)
			if clean != "" {
				items = append(items, "\\item "+clean)
			}
		}
		if len(items) > 0 {
			return stem + "\n\\begin{enumerate}\n" + strings.Join(items, "\n") + "\n\\end{enumerate}"
		}
	}

	return stem
}
