// +build windows
package FKMessageBox

import (
	"syscall"
	"unsafe"
	"FKSystem"
	"fmt"
)

const(
	EnumOk = 1
	EnumCancel = 2
)

func messageBox(hwnd uintptr, caption, title string, flags uint) int {
	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		uintptr(flags))

	return int(ret)
}

func MessageBox_Notice(title, caption string) {
	OSType := FKSystem.GetSystemOSType()
	const (
		NULL  = 0
		MB_OK = 0
	)
	if OSType == FKSystem.EnumSystemWindows {
		messageBox(NULL, caption, title, MB_OK)
	} else {
		fmt.Println(caption)
	}
	return
}

func MessageBox_OkCancel(title, caption string) int {
	const (
		MB_OK  = 0
		MB_CANCEL = 1
	)
	return messageBox(0, caption, title, MB_OK | MB_CANCEL)
}