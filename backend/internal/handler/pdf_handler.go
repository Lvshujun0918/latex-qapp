package handler

import (
	"fmt"
	"time"

	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type PDFHandler struct{}

type pdfExportRequest struct {
	RecordIDs []uint `json:"record_ids"`
}

func NewPDFHandler() *PDFHandler {
	return &PDFHandler{}
}

func (h *PDFHandler) Export(c *gin.Context) {
	var req pdfExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "invalid export payload")
		return
	}

	if len(req.RecordIDs) == 0 {
		httputil.BadRequest(c, "record_ids is required")
		return
	}

	jobID := fmt.Sprintf("job-%d", time.Now().UnixMilli())
	httputil.OK(c, gin.H{
		"jobId":          jobID,
		"status":         "queued",
		"selected_count": len(req.RecordIDs),
		"message":        "pdf export task created",
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
