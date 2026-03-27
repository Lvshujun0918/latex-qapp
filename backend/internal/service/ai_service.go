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
	"github.com/cloudwego/eino/compose"
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
	pipeline   compose.Runnable[*agentPipelineState, *agentPipelineState]

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
	if err := s.ensureEinoPipeline(); err != nil {
		return nil, err
	}

	if s.apiKey == "" {
		return nil, errors.New("QWEN_API_KEY is empty")
	}

	state := &agentPipelineState{
		ImageBase64: imageBase64,
		Trace:       []string{},
	}

	outState, err := s.pipeline.Invoke(context.Background(), state)
	if err != nil {
		return nil, err
	}

	meta := outState.Meta
	latexOut := outState.LatexOut
	solveOut := outState.SolveOut
	tagOut := outState.TagOut
	rawContent := outState.RawContent

	if meta == nil || latexOut == nil || solveOut == nil || tagOut == nil {
		return nil, errors.New("agent pipeline returned incomplete result")
	}

	return &VisionResult{
		LatexQuestion: latexOut.LatexQuestion,
		LatexAnswer:   latexOut.LatexAnswer,
		LatexSolution: solveOut.LatexSolution,
		Tags:          tagOut.Tags,
		Subject:       meta.Subject,
		Title:         meta.Title,
		QuestionType:  meta.QuestionType,
		RawContent:    rawContent,
		AgentTrace:    outState.Trace,
	}, nil
}

func (s *AIService) classifyImageMeta(ctx context.Context, imageBase64 string) (*classifyResult, error) {
	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64)
	messages := []*schema.Message{
		schema.SystemMessage("You are an exam classification agent."),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{Type: schema.ChatMessagePartTypeText, Text: "先判断图片的学科和题目类型。question_type 仅可为 choice/fill_blank/essay/unknown。"},
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

	prompt := fmt.Sprintf("请对以下%s题进行解答，输出可直接展示的latex解答步骤，不要解释框架外内容。\\n\\n题目latex:\\n%s", meta.QuestionType, latexOut.LatexQuestion)
	messages := []*schema.Message{
		schema.SystemMessage("You are a math exam solving agent."),
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
	if strings.TrimSpace(out.LatexSolution) == "" {
		out.LatexSolution = extractLatexBlock(rawContent)
	}

	return out, nil
}

func (s *AIService) generateTags(ctx context.Context, meta *classifyResult, latexOut *latexResult, solveOut *solveResult) (*tagResult, error) {

	prompt := fmt.Sprintf(
		"根据题目与解答生成标签，该题目是%s的%s，并补充2-4个知识点标签。返回 toolcall。\\n\\n题目latex:\\n%s\\n\\n解答latex:\\n%s",
		meta.Subject,
		meta.QuestionType,
		latexOut.LatexQuestion,
		solveOut.LatexSolution,
	)

	messages := []*schema.Message{
		schema.SystemMessage("You are a tagging agent for exam questions."),
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

func (s *AIService) ensureEinoPipeline() error {
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

		chain := compose.NewChain[*agentPipelineState, *agentPipelineState]()

		chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, st *agentPipelineState) (*agentPipelineState, error) {
			st.Trace = append(st.Trace, "step1: classify subject and question type")
			meta, err := s.classifyImageMeta(ctx, st.ImageBase64)
			if err != nil {
				return nil, err
			}
			if meta.Subject == "" {
				meta.Subject = "math"
			}
			if meta.QuestionType == "" {
				meta.QuestionType = "unknown"
			}
			st.Meta = meta
			return st, nil
		}))

		chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, st *agentPipelineState) (*agentPipelineState, error) {
			st.Trace = append(st.Trace, "step2: render exam-zh latex by question type")
			latexOut, rawContent, err := s.generateLatexByType(ctx, st.ImageBase64, st.Meta)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(latexOut.LatexQuestion) == "" {
				latexOut.LatexQuestion = extractLatexBlock(rawContent)
			}
			if strings.TrimSpace(latexOut.LatexQuestion) == "" {
				return nil, errors.New("empty latex_question from model")
			}
			st.LatexOut = latexOut
			st.RawContent = rawContent
			return st, nil
		}))

		chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, st *agentPipelineState) (*agentPipelineState, error) {
			st.Trace = append(st.Trace, "step3: solve the latex question")
			solveOut, err := s.solveByLatex(ctx, st.Meta, st.LatexOut)
			if err != nil {
				return nil, err
			}
			st.SolveOut = solveOut
			return st, nil
		}))

		chain.AppendLambda(compose.InvokableLambda(func(ctx context.Context, st *agentPipelineState) (*agentPipelineState, error) {
			st.Trace = append(st.Trace, "step4: generate tags from latex and solution")
			tagOut, err := s.generateTags(ctx, st.Meta, st.LatexOut, st.SolveOut)
			if err != nil {
				return nil, err
			}
			if len(tagOut.Tags) == 0 {
				tagOut.Tags = inferTags(st.LatexOut.LatexQuestion)
			}
			st.TagOut = tagOut
			return st, nil
		}))

		r, err := chain.Compile(ctx)
		if err != nil {
			s.initErr = fmt.Errorf("compile eino pipeline failed: %w", err)
			return
		}

		s.pipeline = r
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
	base := "识别试卷题目并排版为`exam-zh`可用的latex格式的完美可用的提示词如下：\n\n" +
		"选择题：\n\n" +
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
		"识别图中的题型，按照上面的格式排版图片中的内容，使用latex代码块返回，括号用\\pa来表示，不需要其他内容，不必解题。"

	meta := fmt.Sprintf("\n\n已知分类结果：subject=%s, question_type=%s。请优先按该题型输出。", subject, questionType)
	return base + meta
}
