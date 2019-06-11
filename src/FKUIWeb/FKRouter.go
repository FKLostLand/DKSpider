package FKUIWeb

import (
	"FKApp"
	"FKBase"
	"FKConfig"
	"FKLog"
	"FKStatus"
	ws "github.com/golang/net/websocket"
	"mime"
	"net/http"
	"text/template"
)

// 路由
func Router() {
	mime.AddExtensionType(".css", "text/css")
	// 设置websocket请求路由
	http.Handle("/ws", ws.Handler(wsHandle))
	// 设置websocket报告打印专用路由
	http.Handle("/ws/log", ws.Handler(wsLogHandle))
	// 设置http访问的路由
	http.HandleFunc("/", webHandle)
	//static file server
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(assetFS())))
}

// 处理web页面请求
func webHandle(rw http.ResponseWriter, req *http.Request) {
	sess, _ := G_SessionMgr.SessionStart(rw, req)
	defer sess.SessionRelease(rw)
	index, _ := viewsIndexHtmlBytes()
	t, err := template.New("index").Parse(string(index)) //解析模板文件
	// t, err := template.ParseFiles("web/views/index.html") //解析模板文件
	if err != nil {
		FKLog.G_Log.Error("%v", err)
	}
	//获取信息
	data := map[string]interface{}{
		"title":   FKConfig.APP_NAME,
		"logo":    FKConfig.APP_ICON_PNG,
		"version": FKConfig.APP_VERSION,
		"author":  FKConfig.APP_AUTHOR,
		"mode": map[string]int{
			"offline": FKStatus.OFFLINE,
			"server":  FKStatus.SERVER,
			"client":  FKStatus.CLIENT,
			"unset":   FKStatus.UNSET,
			"curr":    FKApp.G_App.GetAppConfig("Mode").(int),
		},
		"status": map[string]int{
			"stopped": FKStatus.UNINIT,
			"stop":    FKStatus.STOP,
			"run":     FKStatus.RUN,
			"pause":   FKStatus.PAUSE,
		},
		"port": FKApp.G_App.GetAppConfig("MasterPort").(int),
		"ip":   FKApp.G_App.GetAppConfig("MasterIP").(string),
	}
	t.Execute(rw, data) //执行模板的merger操作
}

func wsLogHandle(conn *ws.Conn) {
	defer func() {
		if p := recover(); p != nil {
			FKLog.G_Log.Error("%v", p)
		}
	}()
	// var err error
	sess, _ := G_SessionMgr.SessionStart(nil, conn.Request())
	sessID := sess.SessionID()
	connPool := G_LSC.connPool.Load().(map[string]*ws.Conn)
	if connPool[sessID] == nil {
		G_LSC.Add(sessID, conn)
	}
	defer func() {
		G_LSC.Remove(sessID)
	}()
	for {
		if err := ws.JSON.Receive(conn, nil); err != nil {
			return
		}
	}
}

func wsHandle(conn *ws.Conn) {
	defer func() {
		if p := recover(); p != nil {
			FKLog.G_Log.Error("%v", p)
		}
	}()
	sess, _ := G_SessionMgr.SessionStart(nil, conn.Request())
	sessID := sess.SessionID()
	if Sc.GetConn(sessID) == nil {
		Sc.Add(sessID, conn)
	}

	defer Sc.Remove(sessID, conn)

	go func() {
		var err error
		for info := range Sc.GetWchan(sessID).wchan {
			if _, err = ws.JSON.Send(conn, info); err != nil {
				return
			}
		}
	}()

	for {
		var req map[string]interface{}

		if err := ws.JSON.Receive(conn, &req); err != nil {
			// logs.Log.Debug("websocket接收出错断开 (%v) !", err)
			return
		}

		// log.Log.Debug("Received from web: %v", req)
		wsApi[FKBase.Atoa(req["operate"])](sessID, req)
	}
}
