//go:build windows

package binance

import (
	"os/exec"
	"syscall"
)

const (
	windowsCreateNewProcessGroup = 0x00000200
	windowsDetachedProcess       = 0x00000008
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CreationFlags: windowsDetachedProcess | windowsCreateNewProcessGroup,
	}
}
