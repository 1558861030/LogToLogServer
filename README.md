# LogToLogServer


go get 	github.com/1558861030/LogToLogServer

与LogServer 项目 配合使用



func main() {

	//配置日志服务器ip
	LogToLogServer.LogServer.IpPost = "127.0.0.1:9100"
	//是否记录每个用户返回的信息
	LogToLogServer.LogServer.ResplogBTN = true
	//项目名称
	LogToLogServer.LogServer.ProjectName = "mydwl"
	//应用名称
	LogToLogServer.LogServer.App = "logtest"

	//新建gin路由
	Gin := gin.New()
	//注册日志服务器
	LogToLogServer.InitLogToLogServer(Gin)
	//注册路由
	router.InitRouter(Gin)
	Gin.Run(":80")

}
