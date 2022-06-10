package LogToLogServer

import "github.com/gin-gonic/gin"

func InitLogToLogServer(R *gin.Engine) {
	//关闭gin色彩
	gin.DisableConsoleColor()
	//注册logserver
	R.Use(LogToLogServer())
	//注册console 输出
	R.Use(ConsoleLog())
	//异常保护
	R.Use(gin.Recovery())
	//运行日志服务器
	Run()
}
