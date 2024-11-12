package main

import (
	AuthController "example/go-backend/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

func StartServer() {
	router := gin.Default()
	router.Use(CORSMiddleware())
	router.POST("/", AuthController.Signup)
	router.POST("/signin", AuthController.SignIn)
	router.Run("localhost:8080")
}

func main() {

	StartServer()
}
