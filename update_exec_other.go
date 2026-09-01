//go:build !windows

package binance

import "os/exec"

func detachProcess(cmd *exec.Cmd) {}
