package handler

import (
	"fmt"
	"time"

	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type PDFHandler struct{}

func NewPDFHandler() *PDFHandler {
	return &PDFHandler{}
}

func (h *PDFHandler) Export(c *gin.Context) {
	jobID := fmt.Sprintf("job-%d", time.Now().UnixMilli())
	httputil.OK(c, gin.H{
		"jobId":   jobID,
		"status":  "queued",
		"message": "pdf export task created",
	})
}

func (h *PDFHandler) JobDetail(c *gin.Context) {
	jobID := c.Param("jobId")
	httputil.OK(c, gin.H{
		"jobId":        jobID,
		"status":       "queued",
		"progress":     0,
		"pdf_file_url": "",
	})
}
