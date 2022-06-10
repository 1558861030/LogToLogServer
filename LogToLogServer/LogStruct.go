package LogToLogServer

type servercfg struct {
	//日志服务器的ip地址和端口
	IpPost string
	//项目名称  可能一个项目中包含多个应用
	ProjectName string
	//应用名称
	App string
	//响应内容是否记录   只建议在返回json，string 等内容使用    静态内容没必要
	ResplogBTN bool
}

type LogType struct {
	//项目名称
	ProjectName string `gorm:"index"`
	//应用名称
	App string `gorm:"index"`
	//函数名称
	FuncName string
	//类型
	Type string `gorm:"index"`
	//请求信息
	RequstInfo requst
	//信息
	Info string
	//其他信息
	OtherInfo []otherInfo `gorm:"foreignkey:LogId"`
}
type requst struct {
	//用户IP
	UserIP string
	//用户cookie
	UserCookie string
	//用户session
	UserSession string
	//客户端类型
	UserOS string
	//浏览器类型
	UserBrowserType string
	//请求响应时间
	ResponseTime string
	//请求主机名称
	Host string
	//请求类型
	ResponseType string
	//状态码
	State string
	//请求路由
	ResponseRouter string
	//返回信息
	ReturnInfo string
	//头部信息
	HeadInfo []headInfo
}
type otherInfo struct {
	Key   string
	Value string
}
type headInfo struct {
	Key   string
	Value string
}
