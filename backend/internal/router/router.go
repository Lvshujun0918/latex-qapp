package router

import (
	"latex-qapp/backend/internal/config"
	"latex-qapp/backend/internal/handler"
	"latex-qapp/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, authHandler *handler.AuthHandler, recordHandler *handler.RecordHandler, aiHandler *handler.AIHandler) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.AllowOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

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

		ai := api.Group("/ai")
		{
			// Keep vision endpoint public for dev/mobile capture flow where login may be bypassed.
			ai.POST("/vision/latex", aiHandler.VisionLatex)
		}

		records := api.Group("/records", middleware.JWTAuth(cfg.JWTSecret))
		{
			records.GET("", recordHandler.List)
			records.POST("", recordHandler.Create)
			records.GET(":id", recordHandler.Get)
			records.PUT(":id", recordHandler.Update)
			records.DELETE(":id", recordHandler.Delete)
		}
	}

	return r
}
