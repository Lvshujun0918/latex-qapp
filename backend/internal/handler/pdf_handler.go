package handler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"latex-qapp/backend/internal/model"
	"latex-qapp/backend/internal/service"
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type PDFHandler struct {
	recordService *service.RecordService
	templatePath  string
	outputDir     string
	jobs          map[string]pdfJobDetail
	mu            sync.RWMutex
}

type pdfExportRequest struct {
	RecordIDs []uint `json:"record_ids"`
}

type pdfJobDetail struct {
	JobID         string           `json:"jobId"`
	Status        string           `json:"status"`
	Progress      int              `json:"progress"`
	SelectedCount int              `json:"selected_count"`
	PDFFileURL    string           `json:"pdf_file_url"`
	Message       string           `json:"message"`
	Questions     []pdfJobQuestion `json:"questions"`
}

type pdfJobQuestion struct {
	ID           uint   `json:"id"`
	Index        int    `json:"index"`
	Title        string `json:"title"`
	Subject      string `json:"subject"`
	QuestionType string `json:"question_type"`
}

func NewPDFHandler(recordService *service.RecordService) *PDFHandler {
	templatePath := strings.TrimSpace(os.Getenv("TEMPLATE_TEX_PATH"))
	if templatePath == "" {
		templatePath = filepath.Clean(filepath.Join(".", "template.tex"))
	}

	outputDir := strings.TrimSpace(os.Getenv("PDF_OUTPUT_DIR"))
	if outputDir == "" {
		outputDir = filepath.Clean(filepath.Join(".", "public", "pdfs"))
	}

	return &PDFHandler{
		recordService: recordService,
		templatePath:  templatePath,
		outputDir:     outputDir,
		jobs:          map[string]pdfJobDetail{},
	}
}

func (h *PDFHandler) Export(c *gin.Context) {
	var req pdfExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid export payload")
		return
	}

	if len(req.RecordIDs) == 0 {
		httputil.BadRequest(c, "record_ids is required")
		return
	}
	userIDAny, ok := c.Get("userID")
	if !ok {
		httputil.Unauthorized(c, "missing user identity")
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		httputil.Unauthorized(c, "invalid user identity")
		return
	}

	records, err := h.recordService.GetByIDs(userID, req.RecordIDs)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if len(records) == 0 {
		httputil.BadRequest(c, "no records found")
		return
	}

	jobID := fmt.Sprintf("job-%d", time.Now().UnixMilli())
	jobDir := filepath.Join(h.outputDir, jobID)
	if err := os.MkdirAll(jobDir, 0o755); err != nil {
		httputil.InternalError(c, fmt.Sprintf("create output dir failed: %v", err))
		return
	}

	templateBytes, err := os.ReadFile(h.templatePath)
	if err != nil {
		httputil.InternalError(c, fmt.Sprintf("read template failed: %v", err))
		return
	}

	subject := normalizeSubjectForPDF(records[0].Subject)
	title := fmt.Sprintf("错题导出（%s）", time.Now().Format("2006-01-02"))
	content := buildTemplateContent(records)

	texText := string(templateBytes)
	texText = strings.ReplaceAll(texText, "{{title}}", escapeLatexText(title))
	texText = strings.ReplaceAll(texText, "{{subject}}", escapeLatexText(subject))
	texText = strings.ReplaceAll(texText, "{{content}}", content)

	texPath := filepath.Join(jobDir, "paper.tex")
	if err := os.WriteFile(texPath, []byte(texText), 0o644); err != nil {
		httputil.InternalError(c, fmt.Sprintf("write tex failed: %v", err))
		return
	}
	buildDir := filepath.Join(jobDir, "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		httputil.InternalError(c, fmt.Sprintf("create build dir failed: %v", err))
		return
	}

	cmd := exec.Command("latexmk", "-synctex=1", "-interaction=nonstopmode", "-file-line-error", "-halt-on-error", "-outdir=build", "-xelatex", "paper.tex")
	cmd.Dir = jobDir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		httputil.InternalError(c, fmt.Sprintf("xelatex compile failed: %v\n%s", err, tailString(out.String(), 1200)))
		return
	}

	pdfPath := filepath.Join(buildDir, "paper.pdf")
	if _, err := os.Stat(pdfPath); err != nil {
		httputil.InternalError(c, "pdf not generated")
		return
	}

	publicPDFPath := filepath.Join(jobDir, "paper.pdf")
	if err := copyFile(pdfPath, publicPDFPath); err != nil {
		httputil.InternalError(c, fmt.Sprintf("copy pdf failed: %v", err))
		return
	}

	job := pdfJobDetail{
		JobID:         jobID,
		Status:        "done",
		Progress:      100,
		SelectedCount: len(records),
		PDFFileURL:    fmt.Sprintf("/public/pdfs/%s/paper.pdf", jobID),
		Message:       "pdf generated successfully",
		Questions:     buildJobQuestions(records),
	}

	h.mu.Lock()
	h.jobs[jobID] = job
	h.mu.Unlock()

	httputil.OK(c, job)
}

func (h *PDFHandler) JobDetail(c *gin.Context) {
	jobID := c.Param("jobId")

	h.mu.RLock()
	job, ok := h.jobs[jobID]
	h.mu.RUnlock()
	if !ok {
		httputil.BadRequest(c, "job not found")
		return
	}

	httputil.OK(c, job)
}

func buildTemplateContent(records []model.ErrorRecord) string {
	blocks := make([]string, 0, len(records))
	for i, item := range records {
		blocks = append(blocks, wrapQuestionWithIndex(item.LatexSource, i+1))
	}
	return strings.Join(blocks, "\n\n")
}

func buildJobQuestions(records []model.ErrorRecord) []pdfJobQuestion {
	questions := make([]pdfJobQuestion, 0, len(records))
	for i, item := range records {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("第 %d 题", i+1)
		}

		questions = append(questions, pdfJobQuestion{
			ID:           item.ID,
			Index:        i + 1,
			Title:        title,
			Subject:      strings.TrimSpace(item.Subject),
			QuestionType: strings.TrimSpace(item.QuestionType),
		})
	}

	return questions
}

func wrapQuestionWithIndex(raw string, index int) string {
	content := raw
	if content == "" {
		return fmt.Sprintf("\\begin{question}[index=%d]\n（空题目）\n\\end{question}", index)
	}

	beginRe := regexp.MustCompile(`\\begin{question}(?:\[[^\]]*\])?`)
	if beginRe.MatchString(content) {
		return beginRe.ReplaceAllString(content, fmt.Sprintf(`\begin{question}[index=%d]`, index))
	}

	return fmt.Sprintf("\\begin{question}[index=%d]\n%s\n\\end{question}", index, content)
}

func normalizeSubjectForPDF(subject string) string {
	v := strings.TrimSpace(strings.ToLower(subject))
	switch v {
	case "math", "数学":
		return "数学"
	case "physics", "物理":
		return "物理"
	case "chemistry", "化学":
		return "化学"
	case "biology", "生物":
		return "生物"
	default:
		if strings.TrimSpace(subject) == "" {
			return "综合"
		}
		return strings.TrimSpace(subject)
	}
}

func escapeLatexText(input string) string {
	replacer := strings.NewReplacer(
		`\\`, `\\textbackslash{}`,
		`{`, `\\{`,
		`}`, `\\}`,
		`#`, `\\#`,
		`$`, `\\$`,
		`%`, `\\%`,
		`&`, `\\&`,
		`_`, `\\_`,
	)
	return replacer.Replace(strings.TrimSpace(input))
}

func tailString(input string, max int) string {
	if len(input) <= max {
		return input
	}
	return input[len(input)-max:]
}

func copyFile(srcPath, dstPath string) error {
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := out.ReadFrom(in); err != nil {
		return err
	}

	return out.Sync()
}
