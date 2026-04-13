package service

import (
	"encoding/json"
	"fmt"
	"time"

	"latex-qapp/backend/internal/model"

	"gorm.io/gorm"
)

type RecordService struct {
	db *gorm.DB
}

type CreateRecordInput struct {
	Subject              string     `json:"subject"`
	QuestionType         string     `json:"question_type"`
	Title                string     `json:"title"`
	LatexSource          JSONText   `json:"latex_source"`
	AnswerMode           string     `json:"answer_mode"`
	AnswerText           string     `json:"answer_text"`
	AnswerImageDataURL   string     `json:"answer_image_data_url"`
	AnalysisMode         string     `json:"analysis_mode"`
	AnalysisText         string     `json:"analysis_text"`
	AnalysisImageDataURL string     `json:"analysis_image_data_url"`
	LatexAnswer          string     `json:"latex_answer"`
	QuestionTags         []string   `json:"question_tags"`
	MistakeReason        string     `json:"mistake_reason"`
	MasteryLevel         *int       `json:"mastery_level"`
	ReviewCount          *int       `json:"review_count"`
	ReviewEaseFactor     *float64   `json:"review_ease_factor"`
	LastReviewResult     *string    `json:"last_review_result"`
	LastReviewedAt       *time.Time `json:"last_reviewed_at"`
	NextReviewAt         *time.Time `json:"next_review_at"`
}

// JSONText accepts either a JSON string literal or a JSON object/array,
// and persists the raw JSON text for later PDF assembly.
type JSONText string

func (t *JSONText) UnmarshalJSON(data []byte) error {
	raw := string(data)
	if raw == "null" {
		*t = ""
		return nil
	}

	var plain string
	if err := json.Unmarshal(data, &plain); err == nil {
		*t = JSONText(plain)
		return nil
	}

	*t = JSONText(raw)
	return nil
}

func NewRecordService(db *gorm.DB) *RecordService {
	return &RecordService{db: db}
}

func (s *RecordService) ListByUser(userID uint) ([]model.ErrorRecord, error) {
	var items []model.ErrorRecord
	err := s.db.Where("user_id = ?", userID).Order("id desc").Find(&items).Error
	return items, err
}

func (s *RecordService) GetByID(userID uint, id uint) (*model.ErrorRecord, error) {
	var item model.ErrorRecord
	err := s.db.Where("id = ? AND user_id = ?", id, userID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *RecordService) GetByIDs(userID uint, ids []uint) ([]model.ErrorRecord, error) {
	if len(ids) == 0 {
		return []model.ErrorRecord{}, nil
	}

	var items []model.ErrorRecord
	if err := s.db.Where("user_id = ? AND id IN ?", userID, ids).Find(&items).Error; err != nil {
		return nil, err
	}

	byID := make(map[uint]model.ErrorRecord, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}

	ordered := make([]model.ErrorRecord, 0, len(ids))
	for _, id := range ids {
		item, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("record %d not found or not owned by user", id)
		}
		ordered = append(ordered, item)
	}

	return ordered, nil
}

func (s *RecordService) Create(userID uint, input CreateRecordInput) (*model.ErrorRecord, error) {
	tagsBytes, _ := json.Marshal(input.QuestionTags)

	if input.Subject == "" {
		input.Subject = "math"
	}

	record := model.ErrorRecord{
		UserID:               userID,
		Subject:              input.Subject,
		QuestionType:         input.QuestionType,
		Title:                input.Title,
		LatexSource:          string(input.LatexSource),
		AnswerMode:           normalizeSourceMode(input.AnswerMode),
		AnswerText:           input.AnswerText,
		AnswerImageDataURL:   input.AnswerImageDataURL,
		AnalysisMode:         normalizeSourceMode(input.AnalysisMode),
		AnalysisText:         input.AnalysisText,
		AnalysisImageDataURL: input.AnalysisImageDataURL,
		LatexAnswer:          input.LatexAnswer,
		QuestionTagsJSON:     string(tagsBytes),
		LatexVersion:         1,
		LatexRenderStatus:    "pending",
		MistakeReason:        input.MistakeReason,
		MasteryLevel:         0,
		ReviewCount:          0,
		ReviewEaseFactor:     2.5,
		LastReviewResult:     "none",
	}

	if input.MasteryLevel != nil {
		record.MasteryLevel = clampNonNegative(*input.MasteryLevel)
	}
	if input.ReviewCount != nil {
		record.ReviewCount = clampNonNegative(*input.ReviewCount)
	}
	if input.ReviewEaseFactor != nil {
		record.ReviewEaseFactor = clampEase(*input.ReviewEaseFactor)
	}
	if input.LastReviewResult != nil {
		record.LastReviewResult = normalizeReviewResult(*input.LastReviewResult)
	}
	if input.LastReviewedAt != nil {
		record.LastReviewedAt = input.LastReviewedAt
	}
	if input.NextReviewAt != nil {
		record.NextReviewAt = input.NextReviewAt
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *RecordService) Update(userID uint, id uint, input CreateRecordInput) (*model.ErrorRecord, error) {
	record, err := s.GetByID(userID, id)
	if err != nil {
		return nil, err
	}

	tagsBytes, _ := json.Marshal(input.QuestionTags)

	record.Subject = input.Subject
	record.QuestionType = input.QuestionType
	record.Title = input.Title
	record.LatexSource = string(input.LatexSource)
	record.AnswerMode = normalizeSourceMode(input.AnswerMode)
	record.AnswerText = input.AnswerText
	record.AnswerImageDataURL = input.AnswerImageDataURL
	record.AnalysisMode = normalizeSourceMode(input.AnalysisMode)
	record.AnalysisText = input.AnalysisText
	record.AnalysisImageDataURL = input.AnalysisImageDataURL
	record.LatexAnswer = input.LatexAnswer
	record.QuestionTagsJSON = string(tagsBytes)
	record.MistakeReason = input.MistakeReason
	record.LatexVersion = record.LatexVersion + 1
	if input.MasteryLevel != nil {
		record.MasteryLevel = clampNonNegative(*input.MasteryLevel)
	}
	if input.ReviewCount != nil {
		record.ReviewCount = clampNonNegative(*input.ReviewCount)
	}
	if input.ReviewEaseFactor != nil {
		record.ReviewEaseFactor = clampEase(*input.ReviewEaseFactor)
	}
	if input.LastReviewResult != nil {
		record.LastReviewResult = normalizeReviewResult(*input.LastReviewResult)
	}
	if input.LastReviewedAt != nil {
		record.LastReviewedAt = input.LastReviewedAt
	}
	if input.NextReviewAt != nil {
		record.NextReviewAt = input.NextReviewAt
	}

	if err := s.db.Save(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func clampEase(value float64) float64 {
	if value < 1.3 {
		return 1.3
	}
	if value > 3.0 {
		return 3.0
	}
	return value
}

func normalizeReviewResult(value string) string {
	switch value {
	case "correct", "wrong", "none":
		return value
	default:
		return "none"
	}
}

func normalizeSourceMode(value string) string {
	switch value {
	case "image":
		return "image"
	default:
		return "ai"
	}
}

func (s *RecordService) Delete(userID uint, id uint) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.ErrorRecord{}).Error
}
