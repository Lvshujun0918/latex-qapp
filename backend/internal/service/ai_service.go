package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	einoopenai "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type AIService struct {
	apiKey      string
	baseURL     string
	visionModel string
	textModel   string
	httpClient  *http.Client

	visionChat *einoopenai.ChatModel
	textChat   *einoopenai.ChatModel

	initOnce sync.Once
	initErr  error
}

type agentPipelineState struct {
	ImageBase64 string
	Meta        *classifyResult
	LatexOut    *latexResult
	SolveOut    *solveResult
	TagOut      *tagResult
	RawContent  string
	Trace       []string
}

type VisionResult struct {
	QuestionJSON  *LatexOutput `json:"question_json"`
	LatexSource   string       `json:"latex_source"`
	LatexAnswer   string       `json:"latex_answer"`
	LatexSolution string       `json:"latex_solution"`
	Tags          []string     `json:"tags"`
	Subject       string       `json:"subject"`
	Title         string       `json:"title"`
	QuestionType  string       `json:"question_type"`
	RawContent    string       `json:"raw_content"`
	AgentTrace    []string     `json:"agent_trace"`
}

type VisionStreamEvent struct {
	Stage         string       `json:"stage"`
	Subject       string       `json:"subject,omitempty"`
	Title         string       `json:"title,omitempty"`
	QuestionType  string       `json:"question_type,omitempty"`
	QuestionJSON  *LatexOutput `json:"question_json,omitempty"`
	LatexSource   string       `json:"latex_source,omitempty"`
	LatexAnswer   string       `json:"latex_answer,omitempty"`
	LatexSolution string       `json:"latex_solution,omitempty"`
	Tags          []string     `json:"tags,omitempty"`
	RawContent    string       `json:"raw_content,omitempty"`
	AgentTrace    []string     `json:"agent_trace,omitempty"`
	Done          bool         `json:"done,omitempty"`
	Error         string       `json:"error,omitempty"`
}

type VisionRunOptions struct {
	IncludeSolution bool
}

type SolveStreamEvent struct {
	Stage         string `json:"stage"`
	LatexAnswer   string `json:"latex_answer,omitempty"`
	LatexSolution string `json:"latex_solution,omitempty"`
	Done          bool   `json:"done,omitempty"`
	Error         string `json:"error,omitempty"`
}

type classifyResult struct {
	Subject      string `json:"subject"`
	QuestionType string `json:"question_type"`
	Title        string `json:"title"`
}

type latexResult struct {
	LatexQuestion string `json:"latex_question"`
	LatexAnswer   string `json:"latex_answer"`
}

type solveResult struct {
	LatexAnswer   string `json:"latex_answer"`
	LatexSolution string `json:"latex_solution"`
}

type tagResult struct {
	Tags []string `json:"tags"`
}

func NewAIService(apiKey, baseURL, visionModel, textModel string) *AIService {
	return &AIService{
		apiKey:      apiKey,
		baseURL:     strings.TrimRight(baseURL, "/"),
		visionModel: visionModel,
		textModel:   textModel,
		httpClient:  &http.Client{},
	}
}

func (s *AIService) GenerateLatexDraft(imageBase64 string) (*VisionResult, error) {
	return s.generateLatexDraftStreamWithOptions(context.Background(), imageBase64, nil, VisionRunOptions{IncludeSolution: true})
}

func (s *AIService) GenerateLatexDraftStream(ctx context.Context, imageBase64 string, emit func(*VisionStreamEvent) error) (*VisionResult, error) {
	return s.generateLatexDraftStreamWithOptions(ctx, imageBase64, emit, VisionRunOptions{IncludeSolution: true})
}

func (s *AIService) GenerateQuestionDraft(imageBase64 string) (*VisionResult, error) {
	return s.generateLatexDraftStreamWithOptions(context.Background(), imageBase64, nil, VisionRunOptions{IncludeSolution: false})
}

func (s *AIService) GenerateQuestionDraftStream(ctx context.Context, imageBase64 string, emit func(*VisionStreamEvent) error) (*VisionResult, error) {
	return s.generateLatexDraftStreamWithOptions(ctx, imageBase64, emit, VisionRunOptions{IncludeSolution: false})
}

func (s *AIService) generateLatexDraftStreamWithOptions(ctx context.Context, imageBase64 string, emit func(*VisionStreamEvent) error, opts VisionRunOptions) (*VisionResult, error) {
	if err := s.ensureEinoModels(); err != nil {
		return nil, err
	}

	if s.apiKey == "" {
		return nil, errors.New("QWEN_API_KEY is empty")
	}

	emitEvent := func(evt *VisionStreamEvent) error {
		if emit == nil {
			return nil
		}
		return emit(evt)
	}

	// Create initial state
	initialState := &PipelineState{
		ImageBase64: imageBase64,
		Trace:       []string{},
	}

	state, err := s.nodeClassifyImage(ctx, initialState)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		return nil, err
	}
	_ = emitEvent(&VisionStreamEvent{
		Stage:        "classify",
		Subject:      state.ClassifyOut.Subject,
		Title:        state.ClassifyOut.Title,
		QuestionType: state.ClassifyOut.QuestionType,
		AgentTrace:   append([]string{}, state.Trace...),
		Done:         true,
	})

	state, err = s.nodeGenerateLatex(ctx, state)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		return nil, err
	}
	_ = emitEvent(&VisionStreamEvent{
		Stage:        "latex",
		Subject:      state.ClassifyOut.Subject,
		Title:        state.ClassifyOut.Title,
		QuestionType: state.ClassifyOut.QuestionType,
		QuestionJSON: state.LatexOut,
		LatexSource:  func() string { b, _ := json.Marshal(state.LatexOut); return string(b) }(),
		RawContent:   state.RawContent,
		AgentTrace:   append([]string{}, state.Trace...),
		Done:         true,
	})

	state, err = s.nodeGenerateTags(ctx, state)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		return nil, err
	}
	_ = emitEvent(&VisionStreamEvent{
		Stage: "tags",
		Tags: func() []string {
			if state.TagsOut != nil {
				return state.TagsOut.Tags
			}
			return nil
		}(),
		AgentTrace: append([]string{}, state.Trace...),
		Done:       true,
	})

	if opts.IncludeSolution {
		state, err = s.nodeSolveQuestion(ctx, state)
		if err != nil {
			_ = emitEvent(&VisionStreamEvent{Stage: "error", Error: err.Error(), Done: true})
			return nil, err
		}
		_ = emitEvent(&VisionStreamEvent{
			Stage: "solve",
			LatexAnswer: func() string {
				if state.SolveOut != nil {
					return state.SolveOut.LatexAnswer
				}
				return ""
			}(),
			LatexSolution: func() string {
				if state.SolveOut != nil {
					return state.SolveOut.LatexSolution
				}
				return ""
			}(),
			AgentTrace: append([]string{}, state.Trace...),
			Done:       true,
		})
	}

	// Emit final event
	result, err := nodeMergeFinalResult(ctx, state)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		return nil, err
	}
	_ = emitEvent(&VisionStreamEvent{
		Stage:         "final",
		Subject:       result.Subject,
		Title:         result.Title,
		QuestionType:  result.QuestionType,
		QuestionJSON:  result.QuestionJSON,
		LatexSource:   result.LatexSource,
		LatexAnswer:   result.LatexAnswer,
		LatexSolution: result.LatexSolution,
		Tags:          result.Tags,
		RawContent:    result.RawContent,
		AgentTrace:    result.AgentTrace,
		Done:          true,
	})

	return result, nil
}

func (s *AIService) GenerateSolutionByLatex(ctx context.Context, subject string, questionType string, latexQuestion string) (*solveResult, error) {
	if err := s.ensureEinoModels(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(latexQuestion) == "" {
		return nil, errors.New("latex_question is empty")
	}

	meta := &classifyResult{Subject: subject, QuestionType: questionType}
	meta.Subject = normalizeSubjectLabel(meta.Subject)
	meta.QuestionType = resolveQuestionType(meta.QuestionType, latexQuestion)

	return s.solveByLatex(ctx, meta, &latexResult{LatexQuestion: latexQuestion})
}

func (s *AIService) GenerateSolutionByLatexStream(
	ctx context.Context,
	subject string,
	questionType string,
	latexQuestion string,
	emit func(*SolveStreamEvent) error,
) (*solveResult, error) {
	emitEvent := func(evt *SolveStreamEvent) error {
		if emit == nil {
			return nil
		}
		return emit(evt)
	}

	if err := emitEvent(&SolveStreamEvent{Stage: "solve_start"}); err != nil {
		return nil, err
	}

	out, err := s.GenerateSolutionByLatex(ctx, subject, questionType, latexQuestion)
	if err != nil {
		_ = emitEvent(&SolveStreamEvent{Stage: "error", Error: err.Error(), Done: true})
		return nil, err
	}

	if err := emitEvent(&SolveStreamEvent{
		Stage:         "solve_final",
		LatexAnswer:   out.LatexAnswer,
		LatexSolution: out.LatexSolution,
		Done:          true,
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *AIService) classifyImageMeta(ctx context.Context, imageBase64 string) (*classifyResult, error) {
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64)
	messages := []*schema.Message{
		schema.SystemMessage("你是一位试题分类专家。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{Type: schema.ChatMessagePartTypeText, Text: "先判断图片的学科和题目类型。学科类型为数学/物理/化学/生物，题目类型为选择题/填空题/解答题。只需要根据图片内容进行判断，不要根据其他信息猜测。"},
				{
					Type: schema.ChatMessagePartTypeImageURL,
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{URL: &imageURL},
						Detail:            schema.ImageURLDetailAuto,
					},
				},
			},
		},
	}

	msg, err := s.generateWithForcedTool(ctx, s.visionChat, messages, classifyToolInfo(), 0)
	if err != nil {
		return nil, err
	}

	out := &classifyResult{}
	if !unmarshalToolCallArguments(msg, "classify_exam_meta", out) {
		content := msg.Content
		_ = unmarshalJSONFromText(content, out)
	}

	out.QuestionType = normalizeQuestionTypeLabel(out.QuestionType)
	out.Subject = normalizeSubjectLabel(out.Subject)
	return out, nil
}

func (s *AIService) generateLatexByType(ctx context.Context, imageBase64 string, meta *classifyResult) (*latexResult, string, error) {
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64)
	messages := []*schema.Message{
		schema.SystemMessage("你是一位图片解析转Latex高手，请按照提示词和用户提供的图片工作。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{Type: schema.ChatMessagePartTypeText, Text: buildExamPrompt(meta.QuestionType, meta.Subject)},
				{
					Type: schema.ChatMessagePartTypeImageURL,
					Image: &schema.MessageInputImage{
						MessagePartCommon: schema.MessagePartCommon{URL: &imageURL},
						Detail:            schema.ImageURLDetailAuto,
					},
				},
			},
		},
	}

	msg, err := s.generateWithForcedTool(ctx, s.visionChat, messages, latexToolInfo(), 0.1)
	if err != nil {
		return nil, "", err
	}
	rawContent := msg.Content

	out := &latexResult{}
	if !unmarshalToolCallArguments(msg, "submit_exam_latex", out) {
		_ = unmarshalJSONFromText(rawContent, out)
	}
	if strings.TrimSpace(out.LatexQuestion) == "" {
		out.LatexQuestion = extractLatexBlock(rawContent)
	}

	return out, rawContent, nil
}

func (s *AIService) solveByLatex(ctx context.Context, meta *classifyResult, latexOut *latexResult) (*solveResult, error) {

	prompt := fmt.Sprintf("请对以下%s题进行解答，输出可直接展示的latex结果。必须通过 toolcall 返回两个字段：latex_answer(最终答案) 与 latex_solution(分步解答)。不要输出框架外内容。\\n\\n题目latex:\\n%s", meta.QuestionType, latexOut.LatexQuestion)
	messages := []*schema.Message{
		schema.SystemMessage("你是一位试题解答专家。"),
		schema.UserMessage(prompt),
	}

	msg, err := s.generateWithForcedTool(ctx, s.textChat, messages, solveToolInfo(), 0.2)
	if err != nil {
		return nil, err
	}
	rawContent := msg.Content

	out := &solveResult{}
	if !unmarshalToolCallArguments(msg, "submit_latex_solution", out) {
		_ = unmarshalJSONFromText(rawContent, out)
	}
	if strings.TrimSpace(out.LatexAnswer) == "" {
		out.LatexAnswer = inferAnswerFromSolution(rawContent)
	}
	if strings.TrimSpace(out.LatexSolution) == "" {
		out.LatexSolution = extractLatexBlock(rawContent)
	}

	return out, nil
}

func (s *AIService) generateTags(ctx context.Context, meta *classifyResult, latexOut *latexResult, solveOut *solveResult) (*tagResult, error) {
	solutionText := "(未生成解答)"
	if solveOut != nil && strings.TrimSpace(solveOut.LatexSolution) != "" {
		solutionText = solveOut.LatexSolution
	}

	prompt := fmt.Sprintf(
		"根据题目与（可选）解答生成标签，该题目是%s的%s，并补充2-4个知识点标签。返回 toolcall。\\n\\n题目latex:\\n%s\\n\\n解答latex:\\n%s",
		meta.Subject,
		meta.QuestionType,
		latexOut.LatexQuestion,
		solutionText,
	)

	messages := []*schema.Message{
		schema.SystemMessage("你是一位资深的教育专家，擅长根据题目内容和解答提炼知识点标签。"),
		schema.UserMessage(prompt),
	}

	msg, err := s.generateWithForcedTool(ctx, s.textChat, messages, tagsToolInfo(), 0.1)
	if err != nil {
		return nil, err
	}
	rawContent := msg.Content

	out := &tagResult{}
	if !unmarshalToolCallArguments(msg, "submit_tags", out) {
		_ = unmarshalJSONFromText(rawContent, out)
	}
	if len(out.Tags) == 0 {
		out.Tags = inferTags(latexOut.LatexQuestion)
	}

	return out, nil
}

func (s *AIService) ensureEinoModels() error {
	s.initOnce.Do(func() {
		if s.apiKey == "" {
			s.initErr = errors.New("QWEN_API_KEY is empty")
			return
		}

		ctx := context.Background()

		visionChat, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
			APIKey:     s.apiKey,
			BaseURL:    s.baseURL,
			Model:      s.visionModel,
			HTTPClient: s.httpClient,
		})
		if err != nil {
			s.initErr = fmt.Errorf("init vision model failed: %w", err)
			return
		}

		textChat, err := einoopenai.NewChatModel(ctx, &einoopenai.ChatModelConfig{
			APIKey:     s.apiKey,
			BaseURL:    s.baseURL,
			Model:      s.textModel,
			HTTPClient: s.httpClient,
		})
		if err != nil {
			s.initErr = fmt.Errorf("init text model failed: %w", err)
			return
		}

		s.visionChat = visionChat
		s.textChat = textChat
	})

	return s.initErr
}

func (s *AIService) generateWithForcedTool(ctx context.Context, chatModel *einoopenai.ChatModel, messages []*schema.Message, tool *schema.ToolInfo, temperature float32) (*schema.Message, error) {
	if chatModel == nil {
		return nil, errors.New("chat model is nil")
	}

	withTool, err := chatModel.WithTools([]*schema.ToolInfo{tool})
	if err != nil {
		return nil, err
	}

	return withTool.Generate(
		ctx,
		messages,
		model.WithToolChoice(schema.ToolChoiceForced, tool.Name),
		model.WithTemperature(temperature),
	)
}

func unmarshalToolCallArguments(msg *schema.Message, functionName string, out any) bool {
	for _, tc := range msg.ToolCalls {
		if tc.Function.Name != functionName {
			continue
		}
		argStr := tc.Function.Arguments
		if argStr == "" {
			continue
		}
		if err := json.Unmarshal([]byte(argStr), out); err == nil {
			return true
		}
	}
	return false
}

func classifyToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "classify_exam_meta",
		Desc: "识别题目学科与题型，返回中文标签。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"subject": {
				Type:     schema.String,
				Desc:     "学科，必须为中文标签",
				Enum:     []string{"数学", "物理", "化学", "生物", "未知"},
				Required: true,
			},
			"question_type": {
				Type:     schema.String,
				Desc:     "题型，必须为中文标签",
				Enum:     []string{"选择", "填空", "解答", "未知"},
				Required: true,
			},
			"title": {
				Type: schema.String,
				Desc: "short title",
			},
		}),
	}
}

func latexToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "submit_exam_latex",
		Desc: "返回结构化题目JSON片段：题干、选项、小问（latex片段），不要返回完整题目latex。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"question_type": {
				Type:     schema.String,
				Desc:     "题型，必须为中文标签",
				Enum:     []string{"选择", "填空", "解答", "未知"},
				Required: true,
			},
			"stem": {
				Type:     schema.String,
				Desc:     "题干latex片段，不含question环境",
				Required: true,
			},
			"options": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
				Desc: "选择题选项latex片段数组，例如 [\"A. ...\", \"B. ...\"]",
			},
			"sub_questions": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
				Desc: "解答题小问latex片段数组",
			},
			"latex_answer": {
				Type: schema.String,
				Desc: "optional short answer",
			},
		}),
	}
}

func solveToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "submit_latex_solution",
		Desc: "Return solved steps and final answer in latex format.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"latex_answer": {
				Type:     schema.String,
				Desc:     "final concise answer in latex",
				Required: true,
			},
			"latex_solution": {
				Type:     schema.String,
				Desc:     "latex solution with steps",
				Required: true,
			},
		}),
	}
}

func tagsToolInfo() *schema.ToolInfo {
	return &schema.ToolInfo{
		Name: "submit_tags",
		Desc: "Generate concise tags for exam question.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"tags": {
				Type: schema.Array,
				ElemInfo: &schema.ParameterInfo{
					Type: schema.String,
				},
				Desc:     "tags list",
				Required: true,
			},
		}),
	}
}

func unmarshalJSONFromText(content string, out any) error {
	if strings.TrimSpace(content) == "" {
		return errors.New("empty content")
	}

	jsonRe := regexp.MustCompile("(?s)```json\\s*(.*?)```")
	if m := jsonRe.FindStringSubmatch(content); len(m) >= 2 {
		return json.Unmarshal([]byte(strings.TrimSpace(m[1])), out)
	}

	objRe := regexp.MustCompile(`(?s)\{.*\}`)
	if m := objRe.FindString(content); strings.TrimSpace(m) != "" {
		return json.Unmarshal([]byte(strings.TrimSpace(m)), out)
	}

	return errors.New("no json in content")
}

func extractLatexBlock(content string) string {
	re := regexp.MustCompile("(?s)```latex\\s*(.*?)```")
	m := re.FindStringSubmatch(content)
	if len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}
	return strings.TrimSpace(content)
}

func inferAnswerFromSolution(content string) string {
	boxedRe := regexp.MustCompile(`\\boxed\{([^}]*)\}`)
	if m := boxedRe.FindStringSubmatch(content); len(m) >= 2 {
		return strings.TrimSpace(m[1])
	}

	lines := strings.Split(strings.TrimSpace(content), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" {
			return line
		}
	}

	return ""
}

func inferQuestionType(latex string) string {
	s := strings.ToLower(latex)
	switch {
	case strings.Contains(s, "\\begin{choices}"):
		return "选择"
	case strings.Contains(s, "\\fillin"):
		return "填空"
	case strings.Contains(s, "\\begin{enumerate}"):
		return "解答"
	default:
		return "未知"
	}
}

func resolveQuestionType(rawType string, latex string) string {
	if isMultiQuestionLatex(latex) {
		return "解答"
	}

	normalized := normalizeQuestionTypeLabel(rawType)
	if normalized != "未知" {
		return normalized
	}

	return inferQuestionType(latex)
}

// ============= DELETED ==============
// normalizeEssayLatexQuestion was deleted - now handled by JSON Schema output
// collectEssayItems was deleted - now handled by JSON Schema output
// cleanEssaySegment was deleted - now handled by JSON Schema output
// ====================================

func isMultiQuestionLatex(latex string) bool {
	if strings.TrimSpace(latex) == "" {
		return false
	}

	questionRe := regexp.MustCompile(`\\begin\{question\}(?:\[[^\]]*\])?`)
	if len(questionRe.FindAllString(latex, -1)) >= 2 {
		return true
	}

	numberedLineRe := regexp.MustCompile(`(?:^|\n)\s*(?:\d{1,2}[\.、．\)]|[（(]\d+[）)])\s*`)
	if len(numberedLineRe.FindAllString(latex, -1)) >= 2 {
		return true
	}

	return false
}

func inferTags(latex string) []string {
	typeTag := inferQuestionType(latex)
	return []string{"exam-zh", typeTag}
}

func normalizeSubjectLabel(input string) string {
	v := strings.TrimSpace(strings.ToLower(input))
	switch v {
	case "math", "数学":
		return "数学"
	case "physics", "物理":
		return "物理"
	case "chemistry", "化学":
		return "化学"
	case "biology", "生物":
		return "生物"
	case "", "unknown", "未知":
		return "未知"
	default:
		return strings.TrimSpace(input)
	}
}

func normalizeQuestionTypeLabel(input string) string {
	v := strings.TrimSpace(strings.ToLower(input))
	switch v {
	case "choice", "选择", "选择题", "single_choice", "multiple_choice":
		return "选择"
	case "fill_blank", "填空", "填空题":
		return "填空"
	case "essay", "解答", "解答题", "subjective", "大题":
		return "解答"
	case "", "unknown", "未知":
		return "未知"
	default:
		return strings.TrimSpace(input)
	}
}

func buildExamPrompt(questionType string, subject string) string {
	fence := "```"
	base :=
		"选择题（可能是单选，也可能是多选，不要默认单选）：\n\n" +
			fence + "latex\n" +
			"\\begin{question}[index=1]\n" +
			"    $\\frac{3}{4}$的相反数是\\pa\n" +
			"    \\begin{choices}[columns=4]\n" +
			"        \\item $-\\frac{3}{4}$\n" +
			"        \\item $\\frac{3}{4}$\n" +
			"        \\item $\\frac{4}{3}$\n" +
			"        \\item $-\\frac{4}{3}$\n" +
			"    \\end{choices}\n" +
			"\\end{question}\n" +
			fence + "\n\n" +
			"填空题：\n\n" +
			fence + "latex\n" +
			"\\begin{question}\n" +
			"    己知$\\triangle ABC$为锐角三角形，且$AB=5$，$AC=6$，$\\triangle ABC$的面积为$6\\sqrt{6}$，则$BC=$\\fillin[width = 4em][]。\n" +
			"\\end{question}\n" +
			fence + "\n\n" +
			"大题：\n\n" +
			fence + "latex\n" +
			"\\begin{question}[index=20]\n" +
			"    大题题干在这里。\n" +
			"    \\begin{enumerate}\n" +
			"        \\item 入射电子的德布罗意波长$\\lambda_e$；\n" +
			"        \\item 该靶原子K系特征X射线$K\\alpha$线的波长$\\lambda$；\n" +
			"        \\item 根据实验数据估算该靶原子M层的电离能$E_M$；\n" +
			"        \\item 有同学发现用带电粒子在电场中的运动也能完成对电子速度的测定，请设计实验方案，并指出需要测定的物理量和计算方法。\n" +
			"    \\end{enumerate}\n" +
			"\\end{question}\n" +
			fence + "\n\n" +
			"识别图中的题型，必须通过 toolcall 返回结构化JSON片段，不要输出完整question环境。要求：stem 为题干latex片段；选择题把每个选项放到 options 数组；解答题把每个小问放到 sub_questions 数组。特别注意：当存在 sub_questions 时，stem 只能保留总题干，不能包含任何(1)(2)等编号小问文本；fillin/公式等保留latex片段本身。"

	meta := fmt.Sprintf("\n\n已知分类结果：subject=%s, question_type=%s。请优先按该题型输出。", subject, questionType)
	return base + meta
}
