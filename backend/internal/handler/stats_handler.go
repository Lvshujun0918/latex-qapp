package handler

import (
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct{}

func NewStatsHandler() *StatsHandler {
	return &StatsHandler{}
}

func (h *StatsHandler) Overview(c *gin.Context) {
	httputil.OK(c, gin.H{
		"total_errors":    0,
		"subjects":        []any{},
		"pending_reviews": 0,
		"mastery_average": 0,
	})
}

func (h *StatsHandler) ByCategory(c *gin.Context) {
	httputil.OK(c, []any{})
}

func (h *StatsHandler) Trending(c *gin.Context) {
	httputil.OK(c, []any{})
}
