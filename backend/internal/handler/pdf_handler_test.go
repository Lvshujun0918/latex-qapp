package handler

import (
	"strings"
	"testing"

	"latex-qapp/backend/internal/model"
)

func TestWrapQuestionWithIndex_ReplacesExistingIndex(t *testing.T) {
	input := "\\begin{question}[index=99]\n题干内容\n\\end{question}"
	got := wrapQuestionWithIndex(input, 1)

	if !strings.Contains(got, "\\begin{question}[index=1]") {
		t.Fatalf("expected index replaced to 1, got: %s", got)
	}
	if strings.Contains(got, "index=99") {
		t.Fatalf("expected old index removed, got: %s", got)
	}
}

func TestWrapQuestionWithIndex_WrapsRawContent(t *testing.T) {
	input := "这是一个未包裹题目"
	got := wrapQuestionWithIndex(input, 3)

	expectedPrefix := "\\begin{question}[index=3]"
	expectedSuffix := "\\end{question}"

	if !strings.HasPrefix(got, expectedPrefix) {
		t.Fatalf("expected prefix %q, got: %s", expectedPrefix, got)
	}
	if !strings.Contains(got, input) {
		t.Fatalf("expected original content preserved, got: %s", got)
	}
	if !strings.HasSuffix(got, expectedSuffix) {
		t.Fatalf("expected suffix %q, got: %s", expectedSuffix, got)
	}
}

func TestBuildTemplateContent_IncrementsQuestionIndex(t *testing.T) {
	records := []model.ErrorRecord{
		{LatexSource: "\\begin{question}[index=9]A\\end{question}"},
		{LatexSource: "B"},
	}

	got := buildTemplateContent(records)

	if !strings.Contains(got, "\\begin{question}[index=1]") {
		t.Fatalf("expected first question index=1, got: %s", got)
	}
	if !strings.Contains(got, "\\begin{question}[index=2]") {
		t.Fatalf("expected second question index=2, got: %s", got)
	}
}
