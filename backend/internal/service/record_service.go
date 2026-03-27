package service

import (
	"encoding/json"
	"time"

	"latex-qapp/backend/internal/model"

	"gorm.io/gorm"
)

type RecordService struct {
	db *gorm.DB
}

type CreateRecordInput struct {
	Subject       string   `json:"subject"`
	QuestionType  string   `json:"question_type"`
	Difficulty    int      `json:"difficulty"`
	Title         string   `json:"title"`
	LatexSource   string   `json:"latex_source"`
	LatexAnswer   string   `json:"latex_answer"`
	QuestionTags  []string `json:"question_tags"`
	MistakeReason string   `json:"mistake_reason"`
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

func (s *RecordService) Create(userID uint, input CreateRecordInput) (*model.ErrorRecord, error) {
	tagsBytes, _ := json.Marshal(input.QuestionTags)
	now := time.Now().UnixMilli()

	if input.Subject == "" {
		input.Subject = "math"
	}
	if input.Difficulty <= 0 {
		input.Difficulty = 3
	}

	record := model.ErrorRecord{
		UserID:            userID,
		Subject:           input.Subject,
		QuestionType:      input.QuestionType,
		Difficulty:        input.Difficulty,
		Title:             input.Title,
		LatexSource:       input.LatexSource,
		LatexAnswer:       input.LatexAnswer,
		QuestionTagsJSON:  string(tagsBytes),
		LatexVersion:      1,
		LatexRenderStatus: "pending",
		MistakeReason:     input.MistakeReason,
		MasteryLevel:      0,
		ReviewCount:       0,
		SyncStatus:        "pending",
		LocalVersion:      now,
		ServerVersion:     now,
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
	now := time.Now().UnixMilli()

	record.Subject = input.Subject
	record.QuestionType = input.QuestionType
	record.Difficulty = input.Difficulty
	record.Title = input.Title
	record.LatexSource = input.LatexSource
	record.LatexAnswer = input.LatexAnswer
	record.QuestionTagsJSON = string(tagsBytes)
	record.MistakeReason = input.MistakeReason
	record.LatexVersion = record.LatexVersion + 1
	record.LocalVersion = now
	record.ServerVersion = now
	record.SyncStatus = "pending"

	if err := s.db.Save(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (s *RecordService) Delete(userID uint, id uint) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.ErrorRecord{}).Error
}
