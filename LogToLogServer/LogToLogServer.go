package LogToLogServer

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/gin-gonic/gin"
	"path"
	"runtime"
	"strconv"
	"time"
)

//日志服务器配置
var LogServer servercfg

//日志缓存管道
var logchan chan LogType

//服务器请求地址
var serverPostUrl string

var log *logs.BeeLogger
var logToLogServerErr *logs.BeeLogger

//日志发送到日志服务器
func LogToLogServer() gin.HandlerFunc {
	return logController
}

//控制台打印log
func ConsoleLog() gin.HandlerFunc {
	return consoleLogPrint
}

func logController(c *gin.Context) {
	//中间件起始时间
	t := time.Now()
	lt := new(LogType)
	//获取返回的内容
	blw := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
	c.Writer = blw
	c.Next()
	//获取返回的内容
	if LogServer.ResplogBTN {
		lt.RequstInfo.ReturnInfo = blw.body.String()
	}
	//for s, strings := range c.Request.Header {
	//	fmt.Println(s, strings)
	//}
	//项目名称
	lt.ProjectName = LogServer.ProjectName
	//应用名称
	lt.App = LogServer.App
	//获取cookie
	lt.RequstInfo.UserCookie, _ = c.Cookie("Cookie")
	lt.RequstInfo.UserSession, _ = c.Cookie("Session")
	//获取用户浏览器类型
	lt.RequstInfo.UserOS = c.GetHeader("User-Agent")
	//获取用户操作系统类型
	lt.RequstInfo.UserBrowserType = c.GetHeader("Sec-CH-UA-Platform")
	//获取用户ip
	lt.RequstInfo.UserIP = c.ClientIP()
	//获取请求的路由
	lt.RequstInfo.ResponseRouter = c.Request.RequestURI
	//获取请求的方法
	lt.RequstInfo.ResponseType = c.Request.Method
	lt.RequstInfo.State = strconv.Itoa(c.Writer.Status())
	lt.Type = "Debug"
	//错误处理
	if len(c.Errors) != 0 {
		lt.Type = "Err"
		for i, e := range c.Errors {
			if lt.Info == "" {
				//错误数量加错误信息    如果错误大于1  则其他错误 放到其他信息中
				lt.Info = strconv.Itoa(len(c.Errors)) + e.Error()
			} else {
				eunm := "Err" + strconv.Itoa(i)
				lt.OtherInfo = append(lt.OtherInfo, otherInfo{Key: eunm, Value: e.Error()})
			}
		}
	}

	lt.RequstInfo.Host = c.Request.Host

	//请求的头部信息放到LogType中
	for s, strings := range c.Request.Header {
		var str string
		for _, s2 := range strings {
			str = str + "," + s2
		}

		lt.RequstInfo.HeadInfo = append(lt.RequstInfo.HeadInfo, headInfo{Key: s, Value: str})
	}

	//中间件运行完毕后时间
	t2 := time.Since(t)
	lt.RequstInfo.ResponseTime = t2.String()
	logchan <- *lt
	//fmt.Println(lt.ReturnInfo, lt.RequstInfo.State, lt.RequstInfo.ResponseRouter)
}

//控制台log打印
func consoleLogPrint(c *gin.Context) {
	t := time.Now()
	c.Next()
	t2 := time.Since(t)

	switch {
	case c.Writer.Status() >= 400:
		logs.Error("	|	", c.Request.Method, "|	", c.ClientIP(), "|	", t2, "	|	", c.Writer.Status(), "|	", c.Request.RequestURI)
	default:
		logs.Info("	|	", c.Request.Method, "|	", c.ClientIP(), "|	", t2, "	|	", c.Writer.Status(), "|	", c.Request.RequestURI)
	}

}

func Run() {
	//这个Log 是客户端的日志 客户日志开启了文件名
	log = logs.NewLogger()
	log.SetLogger(logs.AdapterConsole)
	log.SetLogger(logs.AdapterFile, `{"filename":"LogClient.log","level":7,"maxlines":0,"maxsize":0,"daily":false,"maxdays":10,"color":true}`)
	log.SetLogFuncCallDepth(3)
	log.Async(1e3)
	log.EnableFuncCallDepth(true)

	//这个日志是系统的日志
	logs.SetLogger(logs.AdapterConsole)
	logs.SetLogger(logs.AdapterFile, `{"filename":"LogSystem.log","level":7,"maxlines":0,"maxsize":0,"daily":false,"maxdays":10,"color":true}`)
	//logs.SetLogger(logs.AdapterFile, `{"filename":"test.log","daily":"false"}`)
	//logs.SetLogFuncCallDepth(3)
	logs.Async(1e3)

	logToLogServerErr = logs.NewLogger()
	logToLogServerErr.SetLogger(logs.AdapterConsole)
	logToLogServerErr.SetLogger(logs.AdapterFile, `{"filename":"LogToLogServerErr.log","level":7,"maxlines":0,"maxsize":0,"daily":false,"maxdays":10,"color":true}`)
	//logs.SetLogger(logs.AdapterFile, `{"filename":"test.log","daily":"false"}`)
	//logs.SetLogFuncCallDepth(3)
	logToLogServerErr.Async(1e3)

	serverPostUrl = "http://" + LogServer.IpPost + "/putlog"
	logchan = make(chan LogType, 30000)
	//连接日志服务器
	go gotoserver()
	//logs.EnableFuncCallDepth(true)

	//
	//f := &logs.PatternLogFormatter{
	//	Pattern:    "%F:%n|%w%t>> %F",
	//	WhenFormat: "01-02-2016",
	//}
	//logs.RegisterFormatter("pattern", f)
	//_ = logs.SetGlobalFormatter("pattern")
	//
	//logs.Info("beelog", logs.LogMsg{})
	//logs.Info("this %s cat is %v years old", "yellow", 3)

}

func gotoserver() {
	if LogServer.IpPost == "" {
		logs.Error("日志服务器连接地址为空")
		return
	}

	url := "http://" + LogServer.IpPost + "/test"
	req := httplib.Get(url)
	r, err := req.Response()
	if err != nil {
		logs.Error("日志服务器连接错误", err)
		return
	}
	if r.StatusCode == 200 {
		logs.Debug("日志服务器连接成功")
	} else {
		logs.Error("日志服务器连接错误,服务器没有返回正确的状态码")
		return
	}

	for {
		select {
		case l := <-logchan:
			go postLogToServer(l)
		}
	}

}

func postLogToServer(l LogType) {
	req := httplib.Post(serverPostUrl).SetTimeout(60*time.Second, 60*time.Second)
	req.JSONBody(l)
	//req.Response()
	resp, err := req.Response()

	if err != nil {
		logs.GetLogger("请求日志服务器连接错误").Println(err)
		logToLogServerErr.Info("LogToLogServerErr", l)
		return
	}
	if resp.StatusCode != 200 {
		logs.GetLogger("日志服务器返回信息错误").Println(l)
		logToLogServerErr.Info("LogToLogServerErr", l)
		return
	}
}

//对结构体进行预处理
func typeInfo(v ...interface{}) *LogType {
	lt := new(LogType)
	//项目名称
	lt.ProjectName = LogServer.ProjectName
	//应用名称
	lt.App = LogServer.App
	//获取运行文件名称
	var calldepth = 2
	_, f, n, ok := runtime.Caller(calldepth)
	if ok {
		_, lt.FuncName = path.Split(f)
		lt.FuncName = lt.FuncName + ":" + strconv.Itoa(n)
	} else {
		lt.FuncName = "无法获取文件名称"
	}
	if len(v) <= 1 {
		for _, i2 := range v {
			lt.Info = fmt.Sprintf("%s", i2)
		}
	} else {
		for _, i2 := range v {
			msg := fmt.Sprintf("%s", i2)
			lt.Info = strconv.Itoa(len(v)) + "条信息，详见otherInfo表"
			lt.OtherInfo = append(lt.OtherInfo, otherInfo{Key: lt.FuncName, Value: msg})
		}
	}
	return lt
}

//紧急情况
func Emergency(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Emergency"
	logchan <- *lt
	log.Emergency("%s", v)
}

//级别警报
func Alert(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Alert"
	logchan <- *lt
	log.Alert("%s", v)
}

//级别关键
func Critical(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Critical"
	logchan <- *lt
	log.Critical("%s", v)
}

//错误信息
func Error(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Error"
	logchan <- *lt
	log.Error("%s", v)
}

//水平误差
func Warning(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Warning"
	logchan <- *lt
	log.Warning("%s", v)
}

//关键通知
func Notice(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Notice"
	logchan <- *lt
	log.Notice("%s", v)
}

//一般信息
func Info(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Info"
	logchan <- *lt
	log.Info("%s", v)
}

//调试信息
func Debug(v ...interface{}) {
	lt := typeInfo(v)
	lt.Type = "Debug"
	logchan <- *lt
	log.Debug("%s", v)
}

//获取响应体内容
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w CustomResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
