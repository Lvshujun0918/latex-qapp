package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type AIService struct {
	apiKey      string
	baseURL     string
	visionModel string
	textModel   string
	httpClient  *http.Client
}

type VisionResult struct {
	LatexQuestion string   `json:"latex_question"`
	LatexAnswer   string   `json:"latex_answer"`
	Tags          []string `json:"tags"`
	Subject       string   `json:"subject"`
	Title         string   `json:"title"`
	QuestionType  string   `json:"question_type"`
	RawContent    string   `json:"raw_content"`
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
	if s.apiKey == "" {
		return nil, errors.New("QWEN_API_KEY is empty")
	}

	toolSchema := map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        "submit_exam_latex",
			"description": "返回与exam_zh标签相关的题目信息，latex_question字段必须包含题目的latex格式内容，其他字段从图片中识别或者根据latex_question推断得出",
			"parameters": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"latex_question": map[string]any{"type": "string"},
					"latex_answer":   map[string]any{"type": "string"},
					"tags":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					"subject":        map[string]any{"type": "string"},
					"title":          map[string]any{"type": "string"},
					"question_type":  map[string]any{"type": "string"},
				},
				"required": []string{"latex_question", "tags", "question_type"},
			},
		},
	}

	messages := []map[string]any{
		{
			"role":    "system",
			"content": "你是一位图片解析转Latex高手，请按照提示词和用户提供的图片工作",
		},
		{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": buildExamPrompt()},
				{"type": "image_url", "image_url": map[string]any{"url": fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64)}},
			},
		},
	}

	payload := map[string]any{
		"model":       s.visionModel,
		"messages":    messages,
		"tools":       []map[string]any{toolSchema},
		"tool_choice": map[string]any{"type": "function", "function": map[string]any{"name": "submit_exam_latex"}},
		"temperature": 0.1,
	}

	respBody, err := s.postChatCompletions(payload)
	if err != nil {
		return nil, err
	}

	result, err := parseVisionResult(respBody)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(result.LatexQuestion) == "" {
		result.LatexQuestion = extractLatexBlock(result.RawContent)
	}

	if len(result.Tags) == 0 {
		result.Tags = inferTags(result.LatexQuestion)
	}
	if result.QuestionType == "" {
		result.QuestionType = inferQuestionType(result.LatexQuestion)
	}
	if result.Subject == "" {
		result.Subject = "math"
	}

	return result, nil
}

func (s *AIService) postChatCompletions(payload map[string]any) ([]byte, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(http.MethodPost, s.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("qwen api error: %s", string(data))
	}
	return data, nil
}

func parseVisionResult(respBody []byte) (*VisionResult, error) {
	var raw map[string]any
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, err
	}

	res := &VisionResult{}
	choices, _ := raw["choices"].([]any)
	if len(choices) == 0 {
		return nil, errors.New("empty choices")
	}
	firstChoice, _ := choices[0].(map[string]any)
	msg, _ := firstChoice["message"].(map[string]any)
	content, _ := msg["content"].(string)
	res.RawContent = content

	toolCalls, _ := msg["tool_calls"].([]any)
	if len(toolCalls) > 0 {
		firstCall, _ := toolCalls[0].(map[string]any)
		fnObj, _ := firstCall["function"].(map[string]any)
		argStr, _ := fnObj["arguments"].(string)
		if argStr != "" {
			_ = json.Unmarshal([]byte(argStr), &res)
		}
	}

	if res.LatexQuestion == "" && content != "" {
		res.LatexQuestion = extractLatexBlock(content)
	}

	return res, nil
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

func buildExamPrompt() string {
	return `
	选择题：
	\begin{question}[index=1]
	\frac{3}{4}$的相反数是\\pa
	\begin{choices}[columns=4]
	\item $-\\frac{3}{4}$
	\item $\\frac{3}{4}$
	\item $\\frac{4}{3}$
	\item $-\\frac{4}{3}$
	\end{choices}
	\end{question}
	填空题：
	\begin{question}
	己知$\triangle ABC$为锐角三角形，且$AB=5$，$AC=6$，$\triangle ABC$的面积为$6\sqrt{6}$，则$BC=$\fillin[width = 4em][]。
	\end{question}
	大题：
	\begin{question}[index=20]
	大题题干在这里。
	\begin{enumerate}
	\item 入射电子的德布罗意波长$\lambda_e$；
	\item 该靶原子K系特征X射线$K\alpha$线的波长$\lambda$；
	\item 根据实验数据估算该靶原子M层的电离能$E_M$；
	\item 有同学发现用带电粒子在电场中的运动也能完成对电子速度的测定，请设计实验方案，并指出需要测定的物理量和计算方法。
	\end{enumerate}
	\end{question}
	识别图中的题型，按照上面的对应格式以latex排版图片中的内容，使用toolcall返回，括号用\pa来表示，不需要其他内容，不必解题。
	`
}
