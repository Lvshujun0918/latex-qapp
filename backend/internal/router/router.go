package router

import (
	"latex-qapp/backend/internal/config"
	"latex-qapp/backend/internal/handler"
	"latex-qapp/backend/internal/middleware"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New(
	cfg config.Config,
	authHandler *handler.AuthHandler,
	recordHandler *handler.RecordHandler,
	aiHandler *handler.AIHandler,
	statsHandler *handler.StatsHandler,
	pdfHandler *handler.PDFHandler,
) *gin.Engine {
	r := gin.Default()

	allowedOrigins := buildAllowedOrigins(cfg.AllowOrigin)
	corsCfg := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
		ExposeHeaders: []string{
			"Content-Disposition",
			"Content-Type",
		},
		AllowCredentials: false,
	}

	if containsWildcardOrigin(allowedOrigins) {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = allowedOrigins
	}

	r.Use(cors.New(corsCfg))
	r.Static("/public", "./public")

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.JWTAuth(cfg.JWTSecret), authHandler.Me)
		}

		ai := api.Group("/ai", middleware.JWTAuth(cfg.JWTSecret))
		{
			ai.POST("/vision/latex", aiHandler.VisionLatex)
			ai.POST("/vision/latex/stream", aiHandler.VisionLatexStream)
			ai.POST("/solve", aiHandler.SolveLatex)
			ai.POST("/solve/stream", aiHandler.SolveLatexStream)
		}

		records := api.Group("/records", middleware.JWTAuth(cfg.JWTSecret))
		{
			records.GET("", recordHandler.List)
			records.POST("", recordHandler.Create)
			records.GET(":id", recordHandler.Get)
			records.PUT(":id", recordHandler.Update)
			records.DELETE(":id", recordHandler.Delete)
		}

		stats := api.Group("/stats", middleware.JWTAuth(cfg.JWTSecret))
		{
			stats.GET("/overview", statsHandler.Overview)
			stats.GET("/by-category", statsHandler.ByCategory)
			stats.GET("/trending", statsHandler.Trending)
		}

		pdf := api.Group("/pdf", middleware.JWTAuth(cfg.JWTSecret))
		{
			pdf.POST("/export", pdfHandler.Export)
			pdf.GET("/jobs", pdfHandler.ListJobs)
			pdf.GET("/jobs/:jobId", pdfHandler.JobDetail)
			pdf.POST("/jobs/:jobId/questions/:recordId/review", pdfHandler.UpdateQuestionReview)
		}
	}

	return r
}

func buildAllowedOrigins(raw string) []string {
	origins := make([]string, 0, 8)
	seen := map[string]struct{}{}

	push := func(origin string) {
		v := strings.TrimSpace(origin)
		if v == "" {
			return
		}
		if _, ok := seen[v]; ok {
			return
		}
		seen[v] = struct{}{}
		origins = append(origins, v)
	}

	for _, chunk := range strings.Split(raw, ",") {
		push(chunk)
	}

	if len(origins) == 0 {
		push("*")
	}

	// Keep local development and Capacitor webview origins working by default.
	push("http://localhost:5173")
	push("http://localhost:5174")
	push("http://127.0.0.1:5173")
	push("http://127.0.0.1:5174")
	push("capacitor://localhost")
	push("http://localhost")

	return origins
}

func containsWildcardOrigin(origins []string) bool {
	for _, origin := range origins {
		if strings.TrimSpace(origin) == "*" {
			return true
		}
	}
	return false
}
