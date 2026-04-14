package handler

import (
	"os"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

type SystemHandler struct {
	startedAt time.Time
	version   string
}

func NewSystemHandler(startedAt time.Time) *SystemHandler {
	return &SystemHandler{
		startedAt: startedAt,
		version:   resolveAppVersion(),
	}
}

func (h *SystemHandler) Runtime(c *gin.Context) {
	now := time.Now()
	c.JSON(200, gin.H{
		"status":         "ok",
		"version":        h.version,
		"go_version":     runtimeVersion(),
		"started_at":     h.startedAt.UTC().Format(time.RFC3339),
		"server_time":    now.UTC().Format(time.RFC3339),
		"uptime_seconds": int(now.Sub(h.startedAt).Seconds()),
	})
}

func resolveAppVersion() string {
	if fromEnv := os.Getenv("APP_VERSION"); fromEnv != "" {
		return fromEnv
	}

	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}

	return "dev"
}

func runtimeVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		return info.GoVersion
	}
	return "unknown"
}
