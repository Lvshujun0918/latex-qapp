package model

import "time"

type PDFJob struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`
	JobID         string         `gorm:"size:64;uniqueIndex;not null" json:"job_id"`
	Status        string         `gorm:"size:16;not null;default:queued" json:"status"`
	Progress      int            `gorm:"not null;default:0" json:"progress"`
	SelectedCount int            `gorm:"not null;default:0" json:"selected_count"`
	PDFFileURL    string         `gorm:"size:512" json:"pdf_file_url"`
	Message       string         `gorm:"size:255" json:"message"`
	Questions     []PDFJobRecord `gorm:"foreignKey:JobID;references:JobID" json:"questions,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type PDFJobRecord struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	JobID        string    `gorm:"size:64;index;not null" json:"job_id"`
	UserID       uint      `gorm:"index;not null" json:"user_id"`
	RecordID     uint      `gorm:"index;not null" json:"record_id"`
	Index        int       `gorm:"not null" json:"index"`
	Title        string    `gorm:"size:255" json:"title"`
	Subject      string    `gorm:"size:32" json:"subject"`
	QuestionType string    `gorm:"size:32" json:"question_type"`
	LatexSource  string    `gorm:"type:text" json:"latex_source"`
	LatexAnswer  string    `gorm:"type:text" json:"latex_answer"`
	ChildResult  string    `gorm:"size:16;not null;default:none" json:"child_result"`
	ReviewedAt   time.Time `json:"reviewed_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
