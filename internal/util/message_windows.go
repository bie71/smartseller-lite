//go:build windows

package util

import (
	"syscall"
	"unsafe"
)

const (
	mbOK        = 0x00000000
	mbIconError = 0x00000010
)

func showErrorDialog(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")

	textPtr, _ := syscall.UTF16PtrFromString(message)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	proc.Call(0,
		uintptr(unsafe.Pointer(textPtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(mbOK|mbIconError),
	)
}
