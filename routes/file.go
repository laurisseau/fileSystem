package routes

import (
	"fileSystem/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes initializes all API routes

func TestRoutes(r *gin.Engine) {
	routes := r.Group("/")
	{
		routes.GET("/", controllers.GetFile)
		routes.POST("/uploadblob", controllers.UploadHandler)
		routes.GET("/listblobs", controllers.ListBlobs)
		routes.DELETE("/deleteblob", controllers.DeleteBlob)
		routes.POST("/downloadBlob", controllers.DownloadFile)
	}
}


/*
func RegisterRoutes(r *gin.Engine) {
	userRoutes := r.Group("/users")
	{
		userRoutes.GET("/", controllers.GetUsers)
		userRoutes.POST("/", controllers.CreateUser)
	}
}
*/
