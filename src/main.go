package main

import (
	"FKExec"
	_ "FKScript"
)

/*
	入口函数
*/
func main() {
	exec.Init(true)
	exec.Run()
}

