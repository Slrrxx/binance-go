package binance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func executeDownloadedUpdate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%w: stat downloaded file: %v", ErrUpdateUnavailable, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%w: downloaded path is a directory", ErrUpdateUnavailable)
	}
	cmd := exec.Command(path)
	cmd.Dir = filepath.Dir(path)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	detachProcess(cmd)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("%w: start downloaded file: %v", ErrUpdateUnavailable, err)
	}
	_ = cmd.Process.Release()
	return nil
}
