package exec

import (
	"FKConfig"
	"FKUICmd"
	"FKUIWeb"
	"os"
	"os/exec"
	"os/signal"
)

func runPlatform(UIType string) {
	exec.Command("/bin/sh", "-c", "title", FKConfig.APP_FULL_NAME).Start()

	switch UIType {
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
	FKUICmd.ParseFlag()
}