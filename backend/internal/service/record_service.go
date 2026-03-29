package service

import (
	"encoding/json"
	"fmt"

	"latex-qapp/backend/internal/model"

	"gorm.io/gorm"
)

type RecordService struct {
	db *gorm.DB
}

type CreateRecordInput struct {
	Subject       string   `json:"subject"`
	QuestionType  string   `json:"question_type"`
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
		UserID:            userID,
		Subject:           input.Subject,
		QuestionType:      input.QuestionType,
		Title:             input.Title,
		LatexSource:       input.LatexSource,
		LatexAnswer:       input.LatexAnswer,
		QuestionTagsJSON:  string(tagsBytes),
		LatexVersion:      1,
		LatexRenderStatus: "pending",
		MistakeReason:     input.MistakeReason,
		MasteryLevel:      0,
		ReviewCount:       0,
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
	record.LatexSource = input.LatexSource
	record.LatexAnswer = input.LatexAnswer
	record.QuestionTagsJSON = string(tagsBytes)
	record.MistakeReason = input.MistakeReason
	record.LatexVersion = record.LatexVersion + 1

	if err := s.db.Save(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (s *RecordService) Delete(userID uint, id uint) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.ErrorRecord{}).Error
}
