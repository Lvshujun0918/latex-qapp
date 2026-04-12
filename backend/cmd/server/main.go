package main

import (
	"log"

	"latex-qapp/backend/internal/config"
	"latex-qapp/backend/internal/db"
	"latex-qapp/backend/internal/handler"
	"latex-qapp/backend/internal/router"
	"latex-qapp/backend/internal/service"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("db connect failed: %v", err)
	}

	authService := service.NewAuthService(database, cfg.JWTSecret)
	recordService := service.NewRecordService(database)
	aiService := service.NewAIService(cfg.QwenAPIKey, cfg.QwenBaseURL, cfg.QwenVisionModel, cfg.QwenTextModel)

	authHandler := handler.NewAuthHandler(authService)
	recordHandler := handler.NewRecordHandler(recordService)
	aiHandler := handler.NewAIHandler(aiService)
	statsHandler := handler.NewStatsHandler()
	pdfHandler := handler.NewPDFHandler(database, recordService)

	r := router.New(cfg, authHandler, recordHandler, aiHandler, statsHandler, pdfHandler)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server run failed: %v", err)
	}
}
