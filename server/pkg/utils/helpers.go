package utils

import "github.com/gin-gonic/gin"

func ResponseRenderer(message string, payload ...gin.H) gin.H {
	response := gin.H{
		"meta": gin.H{
			"message": message,
		},
	}
	if len(payload) > 0 {
		response["payload"] = payload[0]
	}
	return response
}
