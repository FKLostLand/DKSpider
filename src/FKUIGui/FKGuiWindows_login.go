package FKUIGui

import (
	"FKConfig"
	"FKLog"
	"FKStatus"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func showLoginWindow() {
	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     globalInputor,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:   FKConfig.APP_FULL_NAME,
		MinSize: Size{250, 300},
		Layout:  VBox{ /*MarginsZero: true*/ },
		Children: []Widget{
			RadioButtonGroupBox{
				AssignTo: &mode,
				Title:    "*运行模式",
				Layout:   HBox{},
				MinSize:  Size{0, 60},

				DataMember: "Mode",
				Buttons: []RadioButton{
					{Text: GuiOpt.Mode[0].Key, Value: GuiOpt.Mode[0].Int},
					{Text: GuiOpt.Mode[1].Key, Value: GuiOpt.Mode[1].Int},
					{Text: GuiOpt.Mode[2].Key, Value: GuiOpt.Mode[2].Int},
				},
			},

			VSplitter{
				AssignTo: &host,
				MaxSize:  Size{0, 120},
				Children: []Widget{
					Label{
						Text: "分布式端口：（单机模式不填）",
					},
					NumberEdit{
						Value:    Bind("MasterPort"),
						Suffix:   "",
						Decimals: 0,
					},

					Label{
						Text: "主节点 URL：（客户端模式必填）",
					},
					LineEdit{
						Text: Bind("MasterIP"),
					},
				},
			},

			PushButton{
				Text:     "确认开始",
				MaxSize:  Size{80, 30},
				AssignTo: &runStopBtn,
				OnClicked: func() {
					if err := db.Submit(); err != nil {
						FKLog.G_Log.Error("%v", err)
						return
					}

					switch globalInputor.Mode {
					case FKStatus.OFFLINE:
						showOfflineWindow()

					case FKStatus.SERVER:
						showServerWindow()

					case FKStatus.CLIENT:
						showClientWindow()
					}

				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	if icon, err := walk.NewIconFromResourceId(3); err == nil {
		mw.SetIcon(icon)
	}
	// 运行窗体程序
	mw.Run()
}
