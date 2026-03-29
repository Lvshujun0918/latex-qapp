package model

import "time"

type ErrorRecord struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	UserID            uint      `gorm:"index;not null" json:"user_id"`
	Subject           string    `gorm:"size:32;not null" json:"subject"`
	QuestionType      string    `gorm:"size:32" json:"question_type"`
	Title             string    `gorm:"size:255" json:"title"`
	LatexSource       string    `gorm:"type:text;not null" json:"latex_source"`
	LatexAnswer       string    `gorm:"type:text" json:"latex_answer"`
	QuestionTagsJSON  string    `gorm:"type:text" json:"question_tags_json"`
	LatexVersion      int       `gorm:"default:1;not null" json:"latex_version"`
	LatexRenderStatus string    `gorm:"size:16;default:pending;not null" json:"latex_render_status"`
	MistakeReason     string    `gorm:"type:text" json:"mistake_reason"`
	MasteryLevel      int       `gorm:"default:0;not null" json:"mastery_level"`
	ReviewCount       int       `gorm:"default:0;not null" json:"review_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
