package FKUIGui

import (
	"FKApp"
	"FKConfig"
	. "github.com/lxn/walk/declarative"
)

func showClientWindow() {
	mw.Close()
	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     globalInputor,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:    FKConfig.APP_FULL_NAME + " 【 运行模式 -> 客户端 】",
		MinSize:  Size{1100, 600},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			// Composite{
			// 	Layout:  HBox{},
			// 	MaxSize: Size{1100, 150},
			// 	Children: []Widget{
			// 		PushButton{
			// 			MaxSize:  Size{1000, 150},
			// 			Text:     "断开服务器连接",
			// 			AssignTo: &runStopBtn,
			// 		},
			// 	},
			// },
		},
	}.Create()); err != nil {
		panic(err)
	}

	CreateLogWindow()

	// 初始化应用
	Init()

	// 执行任务
	go FKApp.G_App.Run()

	// 运行窗体程序
	mw.Run()
}
