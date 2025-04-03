package controllers

import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
)

// handleError logs the error and sends an HTTP response if it's critical
func handleError(c *gin.Context, err error) {
	if err != nil {
		log.Println("Error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
