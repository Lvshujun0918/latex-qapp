package model

import "time"

type ErrorRecord struct {
	ID                   uint       `gorm:"primaryKey" json:"id"`
	UserID               uint       `gorm:"index;not null" json:"user_id"`
	Subject              string     `gorm:"size:32;not null" json:"subject"`
	QuestionType         string     `gorm:"size:32" json:"question_type"`
	Title                string     `gorm:"size:255" json:"title"`
	LatexSource          string     `gorm:"type:text;not null" json:"latex_source"`
	SolutionMode         string     `gorm:"size:16;default:ai;not null" json:"solution_mode"`
	AnswerText           string     `gorm:"type:text" json:"answer_text"`
	AnalysisText         string     `gorm:"type:text" json:"analysis_text"`
	SolutionImageDataURL string     `gorm:"type:text" json:"solution_image_data_url"`
	QuestionTagsJSON     string     `gorm:"type:text" json:"question_tags_json"`
	LatexVersion         int        `gorm:"default:1;not null" json:"latex_version"`
	LatexRenderStatus    string     `gorm:"size:16;default:pending;not null" json:"latex_render_status"`
	LatexAnswer          string     `gorm:"type:text" json:"latex_answer"`
	MistakeReason        string     `gorm:"type:text" json:"mistake_reason"`
	MasteryLevel         int        `gorm:"default:0;not null" json:"mastery_level"`
	ReviewCount          int        `gorm:"default:0;not null" json:"review_count"`
	ReviewEaseFactor     float64    `gorm:"type:decimal(4,2);default:2.50;not null" json:"review_ease_factor"`
	LastReviewResult     string     `gorm:"size:16;default:none;not null" json:"last_review_result"`
	LastReviewedAt       *time.Time `json:"last_reviewed_at"`
	NextReviewAt         *time.Time `json:"next_review_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}
