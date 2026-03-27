package httputil

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) {
	c.JSON(200, gin.H{"ok": true, "data": data})
}

func BadRequest(c *gin.Context, msg string) {
	c.JSON(400, gin.H{"ok": false, "error": msg})
}

func Unauthorized(c *gin.Context, msg string) {
	c.JSON(401, gin.H{"ok": false, "error": msg})
}

func InternalError(c *gin.Context, msg string) {
	c.JSON(500, gin.H{"ok": false, "error": msg})
}
