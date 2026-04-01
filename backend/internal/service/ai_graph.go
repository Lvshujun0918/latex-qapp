package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ============= Graph Node Functions =============

// nodeClassifyImage: Step 1 - Classify question type and subject from image
func (s *AIService) nodeClassifyImage(ctx context.Context, state *PipelineState) (*PipelineState, error) {
	if s.visionChat == nil {
		return nil, errors.New("vision model not initialized")
	}

	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", state.ImageBase64)
	messages := []*schema.Message{
		schema.SystemMessage("你是一位试题分类专家。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{Type: schema.ChatMessagePartTypeText, Text: "先判断图片的学科和题目类型。学科类型为数学/物理/化学/生物，题目类型为选择/填空/解答。只需要根据图片内容进行判断，不要根据其他信息猜测。"},
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

	out := &ClassifyOutput{}
	if !unmarshalToolCallArguments(msg, "classify_exam_meta", out) {
		_ = unmarshalJSONFromText(msg.Content, out)
	}

	out.QuestionType = normalizeQuestionTypeLabel(out.QuestionType)
	out.Subject = normalizeSubjectLabel(out.Subject)

	state.ClassifyOut = out
	state.Trace = append(state.Trace, "step1: classify subject and question type - OK")
	return state, nil
}

// nodeGenerateLatex: Step 2 - Generate LaTeX from image
func (s *AIService) nodeGenerateLatex(ctx context.Context, state *PipelineState) (*PipelineState, error) {
	if s.visionChat == nil {
		return nil, errors.New("vision model not initialized")
	}
	if state.ClassifyOut == nil {
		return nil, errors.New("classify output not found")
	}

	imageURL := fmt.Sprintf("data:image/jpeg;base64,%s", state.ImageBase64)
	messages := []*schema.Message{
		schema.SystemMessage("你是一位图片解析转Latex高手，请按照提示词和用户提供的图片工作。"),
		{
			Role: schema.User,
			UserInputMultiContent: []schema.MessageInputPart{
				{Type: schema.ChatMessagePartTypeText, Text: buildExamPrompt(state.ClassifyOut.QuestionType, state.ClassifyOut.Subject)},
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
		return nil, err
	}

	out := &LatexOutput{}
	if !unmarshalToolCallArguments(msg, "submit_exam_latex", out) {
		_ = unmarshalJSONFromText(msg.Content, out)
	}
	if strings.TrimSpace(out.Stem) == "" {
		return nil, errors.New("empty stem from submit_exam_latex")
	}
	out.QuestionType = normalizeQuestionTypeLabel(out.QuestionType)
	if out.QuestionType == "未知" {
		out.QuestionType = state.ClassifyOut.QuestionType
	}

	state.LatexOut = out
	state.RawContent = msg.Content

	assembled := assembleQuestionLatex(out)
	if isMultiQuestionLatex(assembled) {
		state.ClassifyOut.QuestionType = "解答"
	}

	state.Trace = append(state.Trace, "step2: generate latex - OK")
	return state, nil
}

// nodeGenerateTags: Step 3 - Extract tags
func (s *AIService) nodeGenerateTags(ctx context.Context, state *PipelineState) (*PipelineState, error) {
	if s.textChat == nil {
		return nil, errors.New("text model not initialized")
	}
	if state.ClassifyOut == nil || state.LatexOut == nil {
		return nil, errors.New("required outputs not found")
	}

	solutionText := "(未生成解答)"
	if state.SolveOut != nil && strings.TrimSpace(state.SolveOut.LatexSolution) != "" {
		solutionText = state.SolveOut.LatexSolution
	}

	prompt := fmt.Sprintf(
		"根据题目与（可选）解答生成标签，该题目是%s的%s，并补充2-4个知识点标签。返回 toolcall。\n\n题目latex:\n%s\n\n解答latex:\n%s",
		state.ClassifyOut.Subject,
		state.ClassifyOut.QuestionType,
		assembleQuestionLatex(state.LatexOut),
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

	out := &TagsOutput{}
	if !unmarshalToolCallArguments(msg, "submit_tags", out) {
		_ = unmarshalJSONFromText(msg.Content, out)
	}
	if len(out.Tags) == 0 {
		out.Tags = inferTags(assembleQuestionLatex(state.LatexOut))
	}

	state.TagsOut = out
	state.Trace = append(state.Trace, "step3: generate tags - OK")
	return state, nil
}

// nodeSolveQuestion: Step 4 - Solve the question (conditional)
func (s *AIService) nodeSolveQuestion(ctx context.Context, state *PipelineState) (*PipelineState, error) {
	if s.textChat == nil {
		return nil, errors.New("text model not initialized")
	}
	if state.ClassifyOut == nil || state.LatexOut == nil {
		return nil, errors.New("required outputs not found")
	}

	prompt := fmt.Sprintf(
		"请对以下%s题进行解答，输出可直接展示的latex结果。必须通过 toolcall 返回两个字段：latex_answer(最终答案) 与 latex_solution(分步解答)。不要输出框架外内容。\n\n题目latex:\n%s",
		state.ClassifyOut.QuestionType,
		assembleQuestionLatex(state.LatexOut),
	)

	messages := []*schema.Message{
		schema.SystemMessage("你是一位试题解答专家。"),
		schema.UserMessage(prompt),
	}

	msg, err := s.generateWithForcedTool(ctx, s.textChat, messages, solveToolInfo(), 0.2)
	if err != nil {
		return nil, err
	}

	out := &SolveOutput{}
	if !unmarshalToolCallArguments(msg, "submit_latex_solution", out) {
		_ = unmarshalJSONFromText(msg.Content, out)
	}
	if strings.TrimSpace(out.LatexAnswer) == "" {
		out.LatexAnswer = inferAnswerFromSolution(msg.Content)
	}
	if strings.TrimSpace(out.LatexSolution) == "" {
		out.LatexSolution = extractLatexBlock(msg.Content)
	}

	state.SolveOut = out
	state.Trace = append(state.Trace, "step4: solve question - OK")
	return state, nil
}

// nodeMergeFinalResult: Step 5 - Merge results into final output
func nodeMergeFinalResult(ctx context.Context, state *PipelineState) (*VisionResult, error) {
	if state.LatexOut == nil || state.ClassifyOut == nil {
		return nil, errors.New("required state not found")
	}

	finalAnswer := strings.TrimSpace(state.LatexOut.LatexAnswer)
	finalSolution := ""
	if state.SolveOut != nil {
		if strings.TrimSpace(state.SolveOut.LatexAnswer) != "" {
			finalAnswer = strings.TrimSpace(state.SolveOut.LatexAnswer)
		}
		finalSolution = strings.TrimSpace(state.SolveOut.LatexSolution)
	}

	tags := []string{}
	if state.TagsOut != nil {
		tags = state.TagsOut.Tags
	}
	sourceBytes, _ := json.Marshal(state.LatexOut)

	return &VisionResult{
		QuestionJSON:  state.LatexOut,
		LatexSource:   string(sourceBytes),
		LatexAnswer:   finalAnswer,
		LatexSolution: finalSolution,
		Tags:          tags,
		Subject:       state.ClassifyOut.Subject,
		Title:         state.ClassifyOut.Title,
		QuestionType:  state.ClassifyOut.QuestionType,
		RawContent:    state.RawContent,
		AgentTrace:    state.Trace,
	}, nil
}

func assembleQuestionLatex(out *LatexOutput) string {
	if out == nil {
		return ""
	}
	qt := normalizeQuestionTypeLabel(out.QuestionType)
	stem := strings.TrimSpace(out.Stem)
	if stem == "" {
		stem = "（空题目）"
	}

	switch qt {
	case "选择":
		parts := make([]string, 0, len(out.Options))
		for _, opt := range out.Options {
			clean := strings.TrimSpace(opt)
			if clean != "" {
				parts = append(parts, "\\item "+clean)
			}
		}
		if len(parts) == 0 {
			return stem
		}
		return stem + "\n\\begin{choices}\n" + strings.Join(parts, "\n") + "\n\\end{choices}"
	case "解答":
		parts := make([]string, 0, len(out.SubQuestions))
		for _, sq := range out.SubQuestions {
			clean := strings.TrimSpace(sq)
			if clean != "" {
				parts = append(parts, "\\item "+clean)
			}
		}
		if len(parts) == 0 {
			return stem
		}
		return stem + "\n\\begin{enumerate}\n" + strings.Join(parts, "\n") + "\n\\end{enumerate}"
	default:
		return stem
	}
}

// ============= Graph Building =============

// BuildVisionGraph: Build the complete vision processing graph
func (s *AIService) BuildVisionGraph(ctx context.Context, includeSolution bool) (*compose.Graph[*PipelineState, *VisionResult], error) {
	if err := s.ensureEinoModels(); err != nil {
		return nil, err
	}

	// Create graph with graph-level state
	graph := compose.NewGraph[*PipelineState, *VisionResult]()

	const (
		nodeClassify = "classify"
		nodeLatex    = "latex"
		nodeTags     = "tags"
		nodeSolve    = "solve"
		nodeMerge    = "merge"
	)

	// Step 1: Classify image
	classifyLambda := compose.InvokableLambda(
		func(ctx context.Context, state *PipelineState) (*PipelineState, error) {
			return s.nodeClassifyImage(ctx, state)
		},
	)
	_ = graph.AddLambdaNode(nodeClassify, classifyLambda)

	// Step 2: Generate LaTeX
	latexLambda := compose.InvokableLambda(
		func(ctx context.Context, state *PipelineState) (*PipelineState, error) {
			return s.nodeGenerateLatex(ctx, state)
		},
	)
	_ = graph.AddLambdaNode(nodeLatex, latexLambda)

	// Step 3: Generate Tags
	tagsLambda := compose.InvokableLambda(
		func(ctx context.Context, state *PipelineState) (*PipelineState, error) {
			return s.nodeGenerateTags(ctx, state)
		},
	)
	_ = graph.AddLambdaNode(nodeTags, tagsLambda)

	// Step 4: Solve (conditional based on includeSolution)
	if includeSolution {
		solveLambda := compose.InvokableLambda(
			func(ctx context.Context, state *PipelineState) (*PipelineState, error) {
				return s.nodeSolveQuestion(ctx, state)
			},
		)
		_ = graph.AddLambdaNode(nodeSolve, solveLambda)
	}

	// Step 5: Merge final result
	mergeLambda := compose.InvokableLambda(nodeMergeFinalResult)
	_ = graph.AddLambdaNode(nodeMerge, mergeLambda)

	// Build edges
	_ = graph.AddEdge(compose.START, nodeClassify)
	_ = graph.AddEdge(nodeClassify, nodeLatex)
	_ = graph.AddEdge(nodeLatex, nodeTags)

	if includeSolution {
		_ = graph.AddEdge(nodeTags, nodeSolve)
		_ = graph.AddEdge(nodeSolve, nodeMerge)
	} else {
		_ = graph.AddEdge(nodeTags, nodeMerge)
	}

	_ = graph.AddEdge(nodeMerge, compose.END)

	return graph, nil
}
