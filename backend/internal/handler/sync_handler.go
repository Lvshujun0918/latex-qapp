package handler

import (
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type SyncHandler struct{}

func NewSyncHandler() *SyncHandler {
	return &SyncHandler{}
}

func (h *SyncHandler) Push(c *gin.Context) {
	httputil.OK(c, gin.H{"synced": true, "mode": "push", "message": "sync push accepted"})
}

func (h *SyncHandler) Pull(c *gin.Context) {
	httputil.OK(c, gin.H{"synced": true, "mode": "pull", "changes": []any{}})
}
