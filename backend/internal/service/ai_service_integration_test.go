package service

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 运行方式（PowerShell）：
// $env:QWEN_API_KEY="你的key"
// $env:AI_TEST_IMAGE_PATH="C:\\path\\to\\question.jpg"
// go test ./internal/service -run TestAIService_GenerateLatexDraft_FromLocalImage_Integration -v
func TestAIService_GenerateLatexDraft_FromLocalImage_Integration(t *testing.T) {
	apiKey := strings.TrimSpace(os.Getenv("QWEN_API_KEY"))
	if apiKey == "" {
		t.Skip("skip integration test: QWEN_API_KEY is empty")
	}
	t.Log("api key loaded, length=", len(apiKey))

	imagePath := strings.TrimSpace(os.Getenv("AI_TEST_IMAGE_PATH"))
	if imagePath == "" {
		t.Skip("skip integration test: AI_TEST_IMAGE_PATH is empty")
	}

	data, err := os.ReadFile(filepath.Clean(imagePath))
	if err != nil {
		t.Fatalf("read image failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("image file is empty")
	}
	t.Log("image file ready, path=", imagePath)

	baseURL := getEnvOrDefault("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1")
	visionModel := getEnvOrDefault("QWEN_VISION_MODEL", "qwen-vl-max-latest")
	textModel := getEnvOrDefault("QWEN_TEXT_MODEL", "qwen-plus")

	svc := NewAIService(apiKey, baseURL, visionModel, textModel)
	imageBase64 := base64.StdEncoding.EncodeToString(data)

	result, err := svc.GenerateLatexDraft(imageBase64)
	if err != nil {
		t.Fatalf("GenerateLatexDraft failed: %v", err)
	}
	if result == nil {
		t.Fatal("GenerateLatexDraft returned nil result")
	}
	if strings.TrimSpace(result.LatexQuestion) == "" {
		t.Fatalf("expected non-empty latex_question, got empty; trace=%v", result.AgentTrace)
	}
	if len(result.Tags) == 0 {
		t.Fatalf("expected non-empty tags, got empty; trace=%v", result.AgentTrace)
	}

	t.Logf("subject=%s questionType=%s title=%s", result.Subject, result.QuestionType, result.Title)
	t.Logf("tags=%v", result.Tags)
	t.Logf("question=%v", result.LatexQuestion)
	t.Logf("answer=%v", result.LatexAnswer)
	t.Logf("sulution=%v", result.LatexSolution)
	t.Logf("agentTrace=%v", result.AgentTrace)
}

func getEnvOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
