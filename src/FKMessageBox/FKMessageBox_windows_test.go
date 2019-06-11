package FKMessageBox

import (
	"testing"
)

func TestMessageBox(t *testing.T) {
	MessageBox_Notice("测试", "这是提示信息，你必须接受")
	a := MessageBox_OkCancel("测试", "这是选择，你可以做出选择？")
	t.Log(a)
}

