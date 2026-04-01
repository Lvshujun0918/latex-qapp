package service

// ============= Graph Input/Output Types =============

// ClassifyOutput: 输出格式（Tool Call + JSON Schema）
type ClassifyOutput struct {
	Subject      string `json:"subject"`       // 数学/物理/化学/生物
	QuestionType string `json:"question_type"` // 选择/填空/解答
	Title        string `json:"title"`         // 题干简述
}

// LatexOutput: LaTeX 渲染结果
type LatexOutput struct {
	QuestionType string   `json:"question_type"`
	Stem         string   `json:"stem"`
	Options      []string `json:"options,omitempty"`
	SubQuestions []string `json:"sub_questions,omitempty"`
	LatexAnswer  string   `json:"latex_answer,omitempty"`
}

// TagsOutput: 标签提取结果
type TagsOutput struct {
	Tags []string `json:"tags"`
}

// SolveOutput: 解答结果
type SolveOutput struct {
	LatexAnswer   string `json:"latex_answer"`
	LatexSolution string `json:"latex_solution"`
}

// ============= Graph State =============

// PipelineState: Graph 流程中的共享状态
type PipelineState struct {
	// Input
	ImageBase64 string

	// Intermediate outputs
	ClassifyOut *ClassifyOutput
	LatexOut    *LatexOutput
	TagsOut     *TagsOutput
	SolveOut    *SolveOutput

	// For streaming & debugging
	RawContent string
	Trace      []string
}

// ============= Final Output (compatible with existing code) =============

// VisionResult remains the same - final output
// type VisionResult struct {
//     LatexQuestion string   `json:"latex_question"`
//     LatexAnswer   string   `json:"latex_answer"`
//     LatexSolution string   `json:"latex_solution"`
//     Tags          []string `json:"tags"`
//     Subject       string   `json:"subject"`
//     Title         string   `json:"title"`
//     QuestionType  string   `json:"question_type"`
//     RawContent    string   `json:"raw_content"`
//     AgentTrace    []string `json:"agent_trace"`
// }
