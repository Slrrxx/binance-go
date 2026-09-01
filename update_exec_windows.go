//go:build windows

package binance

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}

func launchUpdate(path string) *exec.Cmd {
	cmd := exec.Command("cmd.exe", "/C", "start", "", path)
	detachProcess(cmd)
	return cmd
}
