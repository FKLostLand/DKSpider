package FKUIGui

import (
	"FKApp"
	"FKConfig"
	"FKLog"
	"FKStatus"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func showOfflineWindow() {
	mw.Close()

	if err := (MainWindow{
		AssignTo: &mw,
		DataBinder: DataBinder{
			AssignTo:       &db,
			DataSource:     globalInputor,
			ErrorPresenter: ErrorPresenterRef{&ep},
		},
		Title:   FKConfig.APP_FULL_NAME + " 【 运行模式 ->  单机 】",
		MinSize: Size{1100, 700},
		Layout:  VBox{MarginsZero: true},
		Children: []Widget{

			Composite{
				AssignTo: &setting,
				Layout:   Grid{Columns: 2},
				Children: []Widget{
					// 任务列表
					TableView{
						ColumnSpan:            1,
						MinSize:               Size{850, 450},
						AlternatingRowBGColor: walk.RGB(255, 255, 224),
						CheckBoxes:            true,
						ColumnsOrderable:      true,
						Columns: []TableViewColumn{
							{Title: "#", Width: 50},
							{Title: "任务", Width: 150 /*, Format: "%.2f", Alignment: AlignFar*/},
							{Title: "描述", Width: 630},
						},
						Model: spiderMenu,
					},

					VSplitter{
						ColumnSpan: 1,
						MinSize:    Size{250, 450},
						Children: []Widget{

							VSplitter{
								Children: []Widget{
									Label{
										Text: "自定义配置（用“<>”包围，支持多关键字）",
									},
									LineEdit{
										Text: Bind("Keywords"),
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*采集上限（默认限制URL数）：",
									},
									NumberEdit{
										Value:    Bind("RequestLimit"),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*并发协程：（1~99999）",
									},
									NumberEdit{
										Value:    Bind("MaxThreadNum", Range{1, 99999}),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*分批输出大小：（1~5,000,000 条数据）",
									},
									NumberEdit{
										Value:    Bind("DockerCap", Range{1, 5000000}),
										Suffix:   "",
										Decimals: 0,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*暂停时长参考:",
									},
									ComboBox{
										Value:         Bind("MedianPauseTime", SelRequired{}),
										DisplayMember: "Key",
										BindingMember: "Int64",
										Model:         GuiOpt.Pausetime,
									},
								},
							},

							VSplitter{
								Children: []Widget{
									Label{
										Text: "*代理IP更换频率:",
									},
									ComboBox{
										Value:         Bind("UpdateProxyIntervale", SelRequired{}),
										DisplayMember: "Key",
										BindingMember: "Int64",
										Model:         GuiOpt.ProxyMinute,
									},
								},
							},

							RadioButtonGroupBox{
								ColumnSpan: 1,
								Title:      "*输出方式",
								Layout:     HBox{},
								DataMember: "OutputType",
								Buttons:    radioBtnList,
							},
						},
					},
				},
			},

			Composite{
				Layout: HBox{},
				Children: []Widget{
					VSplitter{
						Children: []Widget{
							// 必填项错误检查
							LineErrorPresenter{
								AssignTo: &ep,
							},
						},
					},

					HSplitter{
						MaxSize: Size{220, 50},
						Children: []Widget{
							Label{
								Text: "继承并保存成功记录",
							},
							CheckBox{
								Checked: Bind("IsInheritSuccess"),
							},
						},
					},
					HSplitter{
						MaxSize: Size{220, 50},
						Children: []Widget{
							Label{
								Text: "继承并保存失败记录",
							},
							CheckBox{
								Checked: Bind("IsInheritFailure"),
							},
						},
					},

					HSplitter{
						MaxSize: Size{90, 50},
						Children: []Widget{
							PushButton{
								Text:      "暂停/恢复",
								AssignTo:  &pauseRecoverBtn,
								OnClicked: offlinePauseRecover,
							},
						},
					},
					HSplitter{
						MaxSize: Size{90, 50},
						Children: []Widget{
							PushButton{
								Text:      "开始运行",
								AssignTo:  &runStopBtn,
								OnClicked: offlineRunStop,
							},
						},
					},
				},
			},
		},
	}.Create()); err != nil {
		panic(err)
	}

	CreateLogWindow()

	pauseRecoverBtn.SetVisible(false)

	// 初始化应用
	Init()

	// 运行窗体程序
	mw.Run()
}

// 暂停\恢复
func offlinePauseRecover() {
	switch FKApp.G_App.Status() {
	case FKStatus.RUN:
		pauseRecoverBtn.SetText("恢复运行")
	case FKStatus.PAUSE:
		pauseRecoverBtn.SetText("暂停")
	}
	FKApp.G_App.PauseRecover()
}

// 开始\停止控制
func offlineRunStop() {
	if !FKApp.G_App.IsStopped() {
		go func() {
			runStopBtn.SetEnabled(false)
			runStopBtn.SetText("停止中…")
			pauseRecoverBtn.SetVisible(false)
			pauseRecoverBtn.SetText("暂停")
			FKApp.G_App.Stop()
			offlineResetBtn()
		}()
		return
	}

	if err := db.Submit(); err != nil {
		FKLog.G_Log.Error("%v", err)
		return
	}

	// 读取任务
	globalInputor.Spiders = spiderMenu.GetChecked()

	runStopBtn.SetText("停止")

	// 记录配置信息
	SetTaskConf()

	// 更新蜘蛛队列
	SpiderPrepare()

	go func() {
		pauseRecoverBtn.SetText("暂停")
		pauseRecoverBtn.SetVisible(true)
		FKApp.G_App.Run()
		offlineResetBtn()
		pauseRecoverBtn.SetVisible(false)
		pauseRecoverBtn.SetText("暂停")
	}()
}

// Offline 模式下按钮状态控制
func offlineResetBtn() {
	runStopBtn.SetEnabled(true)
	runStopBtn.SetText("开始运行")
}
