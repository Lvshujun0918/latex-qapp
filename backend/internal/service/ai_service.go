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
	LatexQuestion string   `json:"latex_question"`
	LatexAnswer   string   `json:"latex_answer"`
	LatexSolution string   `json:"latex_solution"`
	Tags          []string `json:"tags"`
	Subject       string   `json:"subject"`
	Title         string   `json:"title"`
	QuestionType  string   `json:"question_type"`
	RawContent    string   `json:"raw_content"`
	AgentTrace    []string `json:"agent_trace"`
}

type VisionStreamEvent struct {
	Stage         string   `json:"stage"`
	Subject       string   `json:"subject,omitempty"`
	Title         string   `json:"title,omitempty"`
	QuestionType  string   `json:"question_type,omitempty"`
	LatexQuestion string   `json:"latex_question,omitempty"`
	LatexAnswer   string   `json:"latex_answer,omitempty"`
	LatexSolution string   `json:"latex_solution,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	RawContent    string   `json:"raw_content,omitempty"`
	AgentTrace    []string `json:"agent_trace,omitempty"`
	Done          bool     `json:"done,omitempty"`
	Error         string   `json:"error,omitempty"`
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

	state := &agentPipelineState{ImageBase64: imageBase64, Trace: []string{}}
	emitEvent := func(evt *VisionStreamEvent) error {
		if emit == nil {
			return nil
		}
		return emit(evt)
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	state.Trace = append(state.Trace, "step1: classify subject and question type")
	meta, err := s.classifyImageMeta(ctx, state.ImageBase64)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "classify", Error: err.Error(), AgentTrace: append([]string{}, state.Trace...)})
		return nil, err
	}
	if meta.Subject == "" {
		meta.Subject = "math"
	}
	if meta.QuestionType == "" {
		meta.QuestionType = "unknown"
	}
	state.Meta = meta
	if err := emitEvent(&VisionStreamEvent{
		Stage:        "classify",
		Subject:      meta.Subject,
		Title:        meta.Title,
		QuestionType: meta.QuestionType,
		AgentTrace:   append([]string{}, state.Trace...),
	}); err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	state.Trace = append(state.Trace, "step2: render exam-zh latex by question type")
	latexOut, rawContent, err := s.generateLatexByType(ctx, state.ImageBase64, state.Meta)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "latex", Error: err.Error(), AgentTrace: append([]string{}, state.Trace...)})
		return nil, err
	}
	if strings.TrimSpace(latexOut.LatexQuestion) == "" {
		latexOut.LatexQuestion = extractLatexBlock(rawContent)
	}
	if strings.TrimSpace(latexOut.LatexQuestion) == "" {
		err = errors.New("empty latex_question from model")
		_ = emitEvent(&VisionStreamEvent{Stage: "latex", Error: err.Error(), AgentTrace: append([]string{}, state.Trace...)})
		return nil, err
	}
	state.LatexOut = latexOut
	state.RawContent = rawContent
	if err := emitEvent(&VisionStreamEvent{
		Stage:         "latex",
		Subject:       state.Meta.Subject,
		Title:         state.Meta.Title,
		QuestionType:  state.Meta.QuestionType,
		LatexQuestion: latexOut.LatexQuestion,
		LatexAnswer:   latexOut.LatexAnswer,
		RawContent:    rawContent,
		AgentTrace:    append([]string{}, state.Trace...),
	}); err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	state.Trace = append(state.Trace, "step4: generate tags from latex and solution")
	tagOut, err := s.generateTags(ctx, state.Meta, state.LatexOut, state.SolveOut)
	if err != nil {
		_ = emitEvent(&VisionStreamEvent{Stage: "tags", Error: err.Error(), AgentTrace: append([]string{}, state.Trace...)})
		return nil, err
	}
	if len(tagOut.Tags) == 0 {
		tagOut.Tags = inferTags(state.LatexOut.LatexQuestion)
	}
	state.TagOut = tagOut
	if err := emitEvent(&VisionStreamEvent{
		Stage:      "tags",
		Tags:       tagOut.Tags,
		AgentTrace: append([]string{}, state.Trace...),
	}); err != nil {
		return nil, err
	}

	if opts.IncludeSolution {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		state.Trace = append(state.Trace, "step5: solve the latex question")
		solveOut, err := s.solveByLatex(ctx, state.Meta, state.LatexOut)
		if err != nil {
			_ = emitEvent(&VisionStreamEvent{Stage: "solve", Error: err.Error(), AgentTrace: append([]string{}, state.Trace...)})
			return nil, err
		}
		state.SolveOut = solveOut
		if err := emitEvent(&VisionStreamEvent{
			Stage:         "solve",
			LatexAnswer:   solveOut.LatexAnswer,
			LatexSolution: solveOut.LatexSolution,
			AgentTrace:    append([]string{}, state.Trace...),
		}); err != nil {
			return nil, err
		}
	}

	finalAnswer := strings.TrimSpace(state.LatexOut.LatexAnswer)
	finalSolution := ""
	if state.SolveOut != nil {
		if strings.TrimSpace(state.SolveOut.LatexAnswer) != "" {
			finalAnswer = strings.TrimSpace(state.SolveOut.LatexAnswer)
		}
		finalSolution = strings.TrimSpace(state.SolveOut.LatexSolution)
	}

	final := &VisionResult{
		LatexQuestion: state.LatexOut.LatexQuestion,
		LatexAnswer:   finalAnswer,
		LatexSolution: finalSolution,
		Tags:          state.TagOut.Tags,
		Subject:       state.Meta.Subject,
		Title:         state.Meta.Title,
		QuestionType:  state.Meta.QuestionType,
		RawContent:    state.RawContent,
		AgentTrace:    append([]string{}, state.Trace...),
	}

	if err := emitEvent(&VisionStreamEvent{
		Stage:         "final",
		Subject:       final.Subject,
		Title:         final.Title,
		QuestionType:  final.QuestionType,
		LatexQuestion: final.LatexQuestion,
		LatexAnswer:   final.LatexAnswer,
		LatexSolution: final.LatexSolution,
		Tags:          final.Tags,
		RawContent:    final.RawContent,
		AgentTrace:    final.AgentTrace,
		Done:          true,
	}); err != nil {
		return nil, err
	}

	return final, nil
}

func (s *AIService) GenerateSolutionByLatex(ctx context.Context, subject string, questionType string, latexQuestion string) (*solveResult, error) {
	if err := s.ensureEinoModels(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(latexQuestion) == "" {
		return nil, errors.New("latex_question is empty")
	}

	meta := &classifyResult{Subject: subject, QuestionType: questionType}
	if strings.TrimSpace(meta.Subject) == "" {
		meta.Subject = "math"
	}
	if strings.TrimSpace(meta.QuestionType) == "" {
		meta.QuestionType = inferQuestionType(latexQuestion)
	}

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

	if out.QuestionType == "" {
		out.QuestionType = "unknown"
	}
	if out.Subject == "" {
		out.Subject = "math"
	}
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
		Desc: "Classify subject and question_type from exam image.",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"subject": {
				Type:     schema.String,
				Desc:     "subject of the exam question",
				Required: true,
			},
			"question_type": {
				Type:     schema.String,
				Desc:     "question type",
				Enum:     []string{"choice", "fill_blank", "essay", "unknown"},
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
		Desc: "返回符合 exam-zh 的 latex_question 和 latex_answer。",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"latex_question": {
				Type:     schema.String,
				Desc:     "latex formatted question content",
				Required: true,
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
		return "choice"
	case strings.Contains(s, "\\fillin"):
		return "fill_blank"
	case strings.Contains(s, "\\begin{enumerate}"):
		return "essay"
	default:
		return "unknown"
	}
}

func inferTags(latex string) []string {
	typeTag := inferQuestionType(latex)
	return []string{"exam-zh", typeTag}
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
			"识别图中的题型，按照上面的格式排版图片中的内容，使用latex代码块返回，括号用\\pa来表示，不需要其他内容，不必解题。其中index为题号，若图中有题号，则优先使用，否则始终等于1；columns为选项列数，按照选项长度生成。若题面包含'多选'/'可多选'等信息请保留并按多选语义排版。"

	meta := fmt.Sprintf("\n\n已知分类结果：subject=%s, question_type=%s。请优先按该题型输出。", subject, questionType)
	return base + meta
}
