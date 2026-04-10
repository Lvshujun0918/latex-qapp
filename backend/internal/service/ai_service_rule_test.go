package service

import (
	"strings"
	"testing"
)

func TestResolveQuestionType_MultiQuestionDefaultsEssay(t *testing.T) {
	latex := "\\begin{question}[index=1]\nA\n\\end{question}\n\\begin{question}[index=2]\nB\n\\end{question}"
	got := resolveQuestionType("填空", latex)
	if got != "解答" {
		t.Fatalf("expected 解答 for multi-question latex, got %s", got)
	}
}

func TestResolveQuestionType_KeepKnownSingleQuestionType(t *testing.T) {
	latex := "\\begin{question}[index=1]\n题干\\fillin[width = 4em][]。\n\\end{question}"
	got := resolveQuestionType("填空", latex)
	if got != "填空" {
		t.Fatalf("expected 填空 for single-question typed latex, got %s", got)
	}
}

func TestResolveQuestionType_UnknownFallsBackInference(t *testing.T) {
	latex := "\\begin{question}[index=1]\n题干\\begin{enumerate}\\item 小问\\end{enumerate}\n\\end{question}"
	got := resolveQuestionType("未知", latex)
	if got != "解答" {
		t.Fatalf("expected 解答 when unknown type and latex contains enumerate, got %s", got)
	}
}

func TestIsMultiQuestionLatex_NumberedLines(t *testing.T) {
	latex := "1. 第一题\n2. 第二题"
	if !isMultiQuestionLatex(latex) {
		t.Fatal("expected numbered multi-line latex to be multi-question")
	}
}

func TestNormalizeLatexOutput_EssayStemWithoutSubQuestionDuplication(t *testing.T) {
	out := &LatexOutput{
		QuestionType: "解答",
		Stem:         "一质点的运动函数为$\\vec{r} = 2t\\vec{i} + 3t^3\\vec{j}$（SI），求：（1）$t=1\\mathrm{s}$时刻的速度；（2）$1\\mathrm{s}\\sim3\\mathrm{s}$时间内的平均速度$\\bar{v}$和平均加速度$\\bar{a}$。",
		SubQuestions: []string{
			"（1）$t=1\\mathrm{s}$时刻的速度",
			"（2）$1\\mathrm{s}\\sim3\\mathrm{s}$时间内的平均速度$\\bar{v}$和平均加速度$\\bar{a}$",
		},
	}

	normalizeLatexOutput(out)

	if out.Stem == "" {
		t.Fatal("expected non-empty stem")
	}
	if strings.Contains(out.Stem, "（1）") || strings.Contains(out.Stem, "（2）") {
		t.Fatalf("stem still contains numbered sub questions: %s", out.Stem)
	}
	if !strings.Contains(out.Stem, "一质点的运动函数") {
		t.Fatalf("stem lost main context: %s", out.Stem)
	}
	if len(out.SubQuestions) != 2 {
		t.Fatalf("expected 2 sub questions, got %d", len(out.SubQuestions))
	}
	if out.SubQuestions[0] != "$t=1\\mathrm{s}$时刻的速度" {
		t.Fatalf("unexpected first sub question: %s", out.SubQuestions[0])
	}
}

func TestNormalizeLatexOutput_EssayFallbackStemWhenOnlyItems(t *testing.T) {
	out := &LatexOutput{
		QuestionType: "解答",
		Stem:         "（1）证明A；（2）证明B",
		SubQuestions: []string{"（1）证明A", "（2）证明B"},
	}

	normalizeLatexOutput(out)

	if out.Stem != "请解答下列问题" {
		t.Fatalf("expected fallback stem, got %s", out.Stem)
	}
}

// TestNormalizeEssayLatexQuestion_AlwaysSingleQuestionWithEnumerate was removed
// This functionality is now handled by JSON Schema output from the Graph pipeline.
// The Graph nodes ensure that multi-question inputs are properly formatted as structured JSON.
