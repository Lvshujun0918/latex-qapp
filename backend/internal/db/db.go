package db

import (
	"latex-qapp/backend/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.User{}, &model.ErrorRecord{}, &model.PDFJob{}, &model.PDFJobRecord{}); err != nil {
		return nil, err
	}

	return db, nil
}
