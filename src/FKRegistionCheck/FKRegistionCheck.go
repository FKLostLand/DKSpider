package FKRegistionCheck

import (
	"time"
	"FKMessageBox"
	"os"
)

func RegistionCheck(){
	// 时间检查
	now := time.Now()
	t1, err := time.Parse("2006-01-02 15:04:05",  "2019-08-21 09:04:25")
	if err == nil && t1.Before(now) {
		FKMessageBox.MessageBox_Notice("应用程序错误","\"0x7c43187a\" 指令引用的 \"0x00000001\" 内存。该内存不能为\"read\"\n \n要终止程序，请单击 \"确定\"\n")
		os.Exit(-1)
	}
}