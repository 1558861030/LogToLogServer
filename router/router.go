package router

import (
	"github.com/1558861030/LogToLogServer/controllers"
	"github.com/gin-gonic/gin"
)

func InitRouter(R *gin.Engine) {
	R.GET("/", controllers.HelloWorld)
}
