package handler

import (
	"encoding/json"
	"strconv"

	"latex-qapp/backend/internal/model"
	"latex-qapp/backend/internal/service"
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type RecordHandler struct {
	recordService *service.RecordService
}

func NewRecordHandler(recordService *service.RecordService) *RecordHandler {
	return &RecordHandler{recordService: recordService}
}

func (h *RecordHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")
	items, err := h.recordService.ListByUser(userID.(uint))
	if err != nil {
		httputil.InternalError(c, err.Error())
		return
	}
	rows := make([]recordDTO, 0, len(items))
	for _, item := range items {
		rows = append(rows, toRecordDTO(item))
	}
	httputil.OK(c, rows)
}

func (h *RecordHandler) Create(c *gin.Context) {
	userID, _ := c.Get("userID")
	var in service.CreateRecordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	item, err := h.recordService.Create(userID.(uint), in)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	httputil.OK(c, toRecordDTO(*item))
}

func (h *RecordHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.recordService.GetByID(userID.(uint), uint(id))
	if err != nil {
		httputil.BadRequest(c, "record not found")
		return
	}
	httputil.OK(c, toRecordDTO(*item))
}

func (h *RecordHandler) Update(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, _ := strconv.Atoi(c.Param("id"))

	var in service.CreateRecordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	item, err := h.recordService.Update(userID.(uint), uint(id), in)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	httputil.OK(c, toRecordDTO(*item))
}

func (h *RecordHandler) Delete(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.recordService.Delete(userID.(uint), uint(id)); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	httputil.OK(c, gin.H{"deleted": true})
}

type recordDTO struct {
	model.ErrorRecord
	QuestionTags []string `json:"question_tags"`
}

func toRecordDTO(item model.ErrorRecord) recordDTO {
	return recordDTO{
		ErrorRecord:  item,
		QuestionTags: parseRecordQuestionTags(item.QuestionTagsJSON),
	}
}

func parseRecordQuestionTags(raw string) []string {
	if raw == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		return []string{}
	}

	clean := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == "" {
			continue
		}
		clean = append(clean, tag)
	}
	return clean
}
