package handler

import (
	"latex-qapp/backend/internal/service"
	"latex-qapp/backend/pkg/httputil"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in service.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	user, err := h.authService.Register(in)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	httputil.OK(c, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var in service.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		httputil.BadRequest(c, "invalid payload")
		return
	}

	token, user, err := h.authService.Login(in)
	if err != nil {
		httputil.Unauthorized(c, err.Error())
		return
	}

	httputil.OK(c, gin.H{"accessToken": token, "user": user})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := c.Get("userID")
	user, err := h.authService.Me(userID.(uint))
	if err != nil {
		httputil.Unauthorized(c, "user not found")
		return
	}
	httputil.OK(c, user)
}
