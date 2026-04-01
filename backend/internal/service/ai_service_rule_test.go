package service

import (
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

// TestNormalizeEssayLatexQuestion_AlwaysSingleQuestionWithEnumerate was removed
// This functionality is now handled by JSON Schema output from the Graph pipeline.
// The Graph nodes ensure that multi-question inputs are properly formatted as structured JSON.
