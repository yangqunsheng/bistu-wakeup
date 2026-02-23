//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func init() {
	// 设置控制台代码页为 UTF-8，解决中文乱码
	windows.SetConsoleCP(65001)
	windows.SetConsoleOutputCP(65001)

	// 启用 Virtual Terminal Processing，支持 ANSI 转义码（颜色等）
	enableVTP(windows.STD_OUTPUT_HANDLE)
	enableVTP(windows.STD_ERROR_HANDLE)

	// stdin 也需要启用 VTP 以支持 promptui 的光标控制
	if h, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE); err == nil {
		var mode uint32
		if windows.GetConsoleMode(h, &mode) == nil {
			windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_INPUT)
		}
	}

	_ = os.Stdout.Sync()
}

func enableVTP(stdHandle uint32) {
	h, err := windows.GetStdHandle(stdHandle)
	if err != nil {
		return
	}
	var mode uint32
	if windows.GetConsoleMode(h, &mode) != nil {
		return
	}
	windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
}
