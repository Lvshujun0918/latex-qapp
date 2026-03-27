package handler

import (
	"strconv"

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
	httputil.OK(c, items)
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
	httputil.OK(c, item)
}

func (h *RecordHandler) Get(c *gin.Context) {
	userID, _ := c.Get("userID")
	id, _ := strconv.Atoi(c.Param("id"))
	item, err := h.recordService.GetByID(userID.(uint), uint(id))
	if err != nil {
		httputil.BadRequest(c, "record not found")
		return
	}
	httputil.OK(c, item)
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
	httputil.OK(c, item)
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
