package response

import "github.com/gin-gonic/gin"

func Success(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{
		"success": true,
		"data":    data,
	})
}

func Error(c *gin.Context, code, message string) {
	c.JSON(400, gin.H{
		"success": false,
		"code":    code,
		"message": message,
	})
}
