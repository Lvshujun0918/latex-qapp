package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	DatabaseDSN     string
	JWTSecret       string
	QwenAPIKey      string
	QwenBaseURL     string
	QwenVisionModel string
	QwenTextModel   string
	AllowOrigin     string
}

func Load() Config {
	loadDotEnv()

	return Config{
		Port:            getenv("PORT", "8080"),
		DatabaseDSN:     getenv("DATABASE_DSN", "file:app.db?cache=shared&mode=rwc"),
		JWTSecret:       getenv("JWT_SECRET", "change-me-in-production"),
		QwenAPIKey:      getenv("QWEN_API_KEY", ""),
		QwenBaseURL:     getenv("QWEN_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
		QwenVisionModel: getenv("QWEN_VISION_MODEL", "qwen-vl-max-latest"),
		QwenTextModel:   getenv("QWEN_TEXT_MODEL", "qwen-plus"),
		AllowOrigin:     getenv("ALLOW_ORIGIN", "*"),
	}
}

func loadDotEnv() {
	candidates := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
		filepath.Join("backend", ".env"),
	}

	for _, file := range candidates {
		if _, err := os.Stat(file); err != nil {
			continue
		}

		if err := godotenv.Overload(file); err == nil {
			return
		}
	}
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getenvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}
