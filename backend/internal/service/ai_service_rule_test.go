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

func TestNormalizeEssayLatexQuestion_AlwaysSingleQuestionWithEnumerate(t *testing.T) {
	latex := "\\begin{question}[index=1]\\n第一问\\n\\end{question}\\n\\begin{question}[index=2]\\n第二问\\n\\end{question}"
	got := normalizeEssayLatexQuestion(latex)

	if strings.Count(got, "\\begin{question}") != 1 {
		t.Fatalf("expected exactly one question block, got: %s", got)
	}
	if !strings.Contains(got, "\\begin{question}[index=20]") {
		t.Fatalf("expected normalized index=20 format, got: %s", got)
	}
	if !strings.Contains(got, "\\begin{enumerate}") || !strings.Contains(got, "\\end{enumerate}") {
		t.Fatalf("expected enumerate wrapper, got: %s", got)
	}
	if strings.Count(got, "\\item ") < 2 {
		t.Fatalf("expected at least 2 list items from multi-question input, got: %s", got)
	}
}
