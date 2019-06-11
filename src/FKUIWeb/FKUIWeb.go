package FKUIWeb

import (
	"FKApp"
	"FKLog"
	"FKStatus"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

var (
	serverIP      *string
	serverPort    *int
	webServerAddr string
	spiderMenu    []map[string]string
)

// 解析参数
func ParseFlag() {
	serverIP = flag.String("WebServerIP", "0.0.0.0", "   <Web服务器IP>")
	serverPort = flag.Int("WebServerPort", 9090, "   <Web服务器端口>")
}

// 执行入口
func Main() {
	fmt.Println("正在运行网页模式...")

	FKApp.G_App.SetLog(G_LSC).SetAppConfig("Mode", FKStatus.GlobalRuntimeTaskConfig.Mode)

	spiderMenu = func() (spMenu []map[string]string) {
		// 获取蜘蛛家族
		for _, sp := range FKApp.G_App.GetSpiderTypeList() {
			spMenu = append(spMenu, map[string]string{"name": sp.GetName(),
				"description": sp.GetDescription()})
		}
		return spMenu
	}()

	// web服务器地址
	webServerAddr = *serverIP + ":" + strconv.Itoa(*serverPort)
	// 预绑定路由
	Router()
	FKLog.G_Log.App("Web server running on %v", webServerAddr)

	// 自动打开web浏览器
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "http://localhost:"+strconv.Itoa(*serverPort))
	case "darwin":
		cmd = exec.Command("open", "http://localhost:"+strconv.Itoa(*serverPort))
	}
	if cmd != nil {
		go func() {
			FKLog.G_Log.App("Open the default browser after two seconds...")
			time.Sleep(time.Second * 2)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}()
	}

	// 监听端口
	err := http.ListenAndServe(webServerAddr, nil)
	if err != nil {
		FKLog.G_Log.Emergency("ListenAndServe: %v", err)
	}
}
