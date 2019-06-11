package exec

import (
	"FKConfig"
	"FKUICmd"
	"FKUIGui"
	"FKUIWeb"
	"os"
	"os/exec"
	"os/signal"
)

func runPlatform(UIType string) {
	exec.Command("cmd.exe", "/c", "title", FKConfig.APP_FULL_NAME).Start()

	switch UIType {
	case "gui":
		FKUIGui.Main()
	case "cmd":
		FKUICmd.Main()
	case "web":
		fallthrough
	default:
		ctrl := make(chan os.Signal, 1)
		signal.Notify(ctrl, os.Interrupt, os.Kill)
		go FKUIWeb.Main()
		<-ctrl
	}
}

func parseDifferentUIFlag(){
	FKUIWeb.ParseFlag()
	FKUIGui.ParseFlag()
	FKUICmd.ParseFlag()
}