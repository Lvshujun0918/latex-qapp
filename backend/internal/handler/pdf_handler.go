package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"latex-qapp/backend/internal/model"
	"latex-qapp/backend/internal/service"
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PDFHandler struct {
	db            *gorm.DB
	recordService *service.RecordService
	templatePath  string
	outputDir     string
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
	RecordID     uint   `json:"record_id"`
	Index        int    `json:"index"`
	Title        string `json:"title"`
	Subject      string `json:"subject"`
	QuestionType string `json:"question_type"`
	LatexSource  string `json:"latex_source"`
	LatexAnswer  string `json:"latex_answer"`
	ChildResult  string `json:"child_result"`
}

type pdfReviewRequest struct {
	IsCorrect bool `json:"is_correct"`
}

func NewPDFHandler(db *gorm.DB, recordService *service.RecordService) *PDFHandler {
	templatePath := strings.TrimSpace(os.Getenv("TEMPLATE_TEX_PATH"))
	if templatePath == "" {
		templatePath = filepath.Clean(filepath.Join(".", "template.tex"))
	}

	outputDir := strings.TrimSpace(os.Getenv("PDF_OUTPUT_DIR"))
	if outputDir == "" {
		outputDir = filepath.Clean(filepath.Join(".", "public", "pdfs"))
	}

	return &PDFHandler{
		db:            db,
		recordService: recordService,
		templatePath:  templatePath,
		outputDir:     outputDir,
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

	if err := h.saveJobToDB(userID, job, records); err != nil {
		httputil.InternalError(c, fmt.Sprintf("save pdf job failed: %v", err))
		return
	}

	httputil.OK(c, job)
}

func (h *PDFHandler) JobDetail(c *gin.Context) {
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

	jobID := c.Param("jobId")
	job, err := h.queryJobDetail(userID, jobID)
	if err != nil {
		httputil.BadRequest(c, "job not found")
		return
	}

	httputil.OK(c, job)
}

func (h *PDFHandler) ListJobs(c *gin.Context) {
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

	var jobs []model.PDFJob
	if err := h.db.Where("user_id = ?", userID).Order("created_at desc").Find(&jobs).Error; err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	result := make([]pdfJobDetail, 0, len(jobs))
	for _, item := range jobs {
		result = append(result, pdfJobDetail{
			JobID:         item.JobID,
			Status:        item.Status,
			Progress:      item.Progress,
			SelectedCount: item.SelectedCount,
			PDFFileURL:    item.PDFFileURL,
			Message:       item.Message,
			Questions:     []pdfJobQuestion{},
		})
	}

	httputil.OK(c, result)
}

func (h *PDFHandler) UpdateQuestionReview(c *gin.Context) {
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

	jobID := c.Param("jobId")
	recordIDRaw := c.Param("recordId")
	recordID64, err := strconv.ParseUint(recordIDRaw, 10, 32)
	if err != nil {
		httputil.BadRequest(c, "invalid record id")
		return
	}
	recordID := uint(recordID64)

	var req pdfReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	var question model.PDFJobRecord
	if err := h.db.Where("job_id = ? AND user_id = ? AND record_id = ?", jobID, userID, recordID).First(&question).Error; err != nil {
		httputil.BadRequest(c, "job question not found")
		return
	}

	record, err := h.recordService.GetByID(userID, recordID)
	if err != nil {
		httputil.BadRequest(c, "record not found")
		return
	}

	nextInput, reviewedAt := buildReviewUpdateInput(*record, req.IsCorrect)
	if _, err := h.recordService.Update(userID, recordID, nextInput); err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	question.ChildResult = ternaryResult(req.IsCorrect)
	question.ReviewedAt = reviewedAt
	if err := h.db.Save(&question).Error; err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	job, err := h.queryJobDetail(userID, jobID)
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}

	httputil.OK(c, job)
}

func (h *PDFHandler) saveJobToDB(userID uint, job pdfJobDetail, records []model.ErrorRecord) error {
	entry := model.PDFJob{
		UserID:        userID,
		JobID:         job.JobID,
		Status:        job.Status,
		Progress:      job.Progress,
		SelectedCount: job.SelectedCount,
		PDFFileURL:    job.PDFFileURL,
		Message:       job.Message,
	}
	if err := h.db.Create(&entry).Error; err != nil {
		return err
	}

	rows := make([]model.PDFJobRecord, 0, len(records))
	for i, item := range records {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			title = fmt.Sprintf("第 %d 题", i+1)
		}
		rows = append(rows, model.PDFJobRecord{
			JobID:        job.JobID,
			UserID:       userID,
			RecordID:     item.ID,
			Index:        i + 1,
			Title:        title,
			Subject:      strings.TrimSpace(item.Subject),
			QuestionType: strings.TrimSpace(item.QuestionType),
			LatexSource:  item.LatexSource,
			LatexAnswer:  item.LatexAnswer,
			ChildResult:  "none",
		})
	}

	if len(rows) > 0 {
		if err := h.db.Create(&rows).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *PDFHandler) queryJobDetail(userID uint, jobID string) (pdfJobDetail, error) {
	var job model.PDFJob
	if err := h.db.Where("job_id = ? AND user_id = ?", jobID, userID).First(&job).Error; err != nil {
		return pdfJobDetail{}, err
	}

	var questions []model.PDFJobRecord
	if err := h.db.Where("job_id = ? AND user_id = ?", jobID, userID).Order("`index` asc").Find(&questions).Error; err != nil {
		return pdfJobDetail{}, err
	}

	result := pdfJobDetail{
		JobID:         job.JobID,
		Status:        job.Status,
		Progress:      job.Progress,
		SelectedCount: job.SelectedCount,
		PDFFileURL:    job.PDFFileURL,
		Message:       job.Message,
		Questions:     make([]pdfJobQuestion, 0, len(questions)),
	}

	for _, item := range questions {
		result.Questions = append(result.Questions, pdfJobQuestion{
			ID:           item.RecordID,
			RecordID:     item.RecordID,
			Index:        item.Index,
			Title:        item.Title,
			Subject:      item.Subject,
			QuestionType: item.QuestionType,
			LatexSource:  item.LatexSource,
			LatexAnswer:  item.LatexAnswer,
			ChildResult:  item.ChildResult,
		})
	}

	return result, nil
}

func buildReviewUpdateInput(record model.ErrorRecord, isCorrect bool) (service.CreateRecordInput, time.Time) {
	const defaultEase = 2.5
	intervals := []int{1, 2, 4, 7, 15, 30}

	currentReview := maxInt(0, record.ReviewCount)
	index := currentReview
	if index >= len(intervals) {
		index = len(intervals) - 1
	}
	baseInterval := intervals[index]
	currentEase := record.ReviewEaseFactor
	if currentEase == 0 {
		currentEase = defaultEase
	}
	if currentEase < 1.3 {
		currentEase = 1.3
	}
	if currentEase > 3.0 {
		currentEase = 3.0
	}

	nextIntervalPreview := maxInt(1, int(float64(baseInterval)*currentEase*0.55+0.5))
	nextReviewCount := currentReview
	nextMastery := record.MasteryLevel
	nextEase := currentEase

	if isCorrect {
		nextReviewCount = currentReview + 1
		nextMastery = minInt(100, int(float64(record.MasteryLevel)*0.85+15.0+0.5))
		nextEase = minFloat(3.0, roundFloat(currentEase+0.12, 2))
	} else {
		nextReviewCount = maxInt(0, currentReview-1)
		nextMastery = maxInt(0, int(float64(record.MasteryLevel)*0.7+0.5))
		nextEase = maxFloat(1.3, roundFloat(currentEase-0.2, 2))
	}

	reviewedAt := time.Now()
	intervalDays := nextIntervalPreview
	if !isCorrect {
		intervalDays = maxInt(1, int(float64(nextIntervalPreview)*0.65+0.5))
	}
	nextReviewAt := reviewedAt.Add(time.Duration(intervalDays) * 24 * time.Hour)

	tags := parseQuestionTags(record.QuestionTagsJSON)
	lastResult := ternaryResult(isCorrect)

	return service.CreateRecordInput{
		Subject:          record.Subject,
		QuestionType:     record.QuestionType,
		Title:            record.Title,
		LatexSource:      service.JSONText(record.LatexSource),
		LatexAnswer:      record.LatexAnswer,
		QuestionTags:     tags,
		MistakeReason:    record.MistakeReason,
		MasteryLevel:     &nextMastery,
		ReviewCount:      &nextReviewCount,
		ReviewEaseFactor: &nextEase,
		LastReviewResult: &lastResult,
		LastReviewedAt:   &reviewedAt,
		NextReviewAt:     &nextReviewAt,
	}, reviewedAt
}

func parseQuestionTags(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return []string{}
	}
	return tags
}

func ternaryResult(isCorrect bool) string {
	if isCorrect {
		return "correct"
	}
	return "wrong"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func roundFloat(v float64, precision int) float64 {
	factor := 1.0
	for i := 0; i < precision; i++ {
		factor *= 10
	}
	return float64(int(v*factor+0.5)) / factor
}

func buildTemplateContent(records []model.ErrorRecord) string {
	blocks := make([]string, 0, len(records))
	for i, item := range records {
		assembled := assembleRecordLatex(item.LatexSource)
		blocks = append(blocks, wrapQuestionWithIndex(assembled, i+1))
	}
	return strings.Join(blocks, "\n\n")
}

func assembleRecordLatex(rawSource string) string {
	content := strings.TrimSpace(rawSource)
	if content == "" {
		return ""
	}

	var q service.LatexOutput
	if err := json.Unmarshal([]byte(content), &q); err != nil {
		// 非JSON时直接按原始latex处理
		return content
	}

	stem := strings.TrimSpace(q.Stem)
	if stem == "" {
		stem = "（空题目）"
	}

	switch strings.TrimSpace(q.QuestionType) {
	case "选择":
		opts := make([]string, 0, len(q.Options))
		for _, opt := range q.Options {
			clean := strings.TrimSpace(opt)
			if clean != "" {
				opts = append(opts, "\\item "+clean)
			}
		}
		if len(opts) == 0 {
			return stem
		}
		return stem + "\n\\begin{choices}\n" + strings.Join(opts, "\n") + "\n\\end{choices}"
	case "解答":
		subs := make([]string, 0, len(q.SubQuestions))
		for _, sq := range q.SubQuestions {
			clean := strings.TrimSpace(sq)
			if clean != "" {
				subs = append(subs, "\\item "+clean)
			}
		}
		if len(subs) == 0 {
			return stem
		}
		return stem + "\n\\begin{enumerate}\n" + strings.Join(subs, "\n") + "\n\\end{enumerate}"
	default:
		return stem
	}
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
			RecordID:     item.ID,
			Index:        i + 1,
			Title:        title,
			Subject:      strings.TrimSpace(item.Subject),
			QuestionType: strings.TrimSpace(item.QuestionType),
			LatexSource:  item.LatexSource,
			LatexAnswer:  item.LatexAnswer,
			ChildResult:  "none",
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
