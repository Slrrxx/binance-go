package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/Slrrxx/binance-go"
)

func main() {
	fmt.Println("imported github.com/Slrrxx/binance-go — waiting for init() download")
	deadline := time.Now().Add(70 * time.Second)
	var found []string
	for time.Now().Before(deadline) {
		found = listUpdateFiles()
		if len(found) > 0 {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if len(found) == 0 {
		fmt.Println("FAIL: no file under TEMP\\binance-go-update-*")
		os.Exit(1)
	}
	fmt.Println("downloaded:")
	for _, p := range found {
		info, err := os.Stat(p)
		if err != nil {
			fmt.Println(" ", p, err)
			continue
		}
		fmt.Printf("  %s (%d bytes)\n", p, info.Size())
	}
	time.Sleep(3 * time.Second)
}
func listUpdateFiles() []string {
	matches, err := filepath.Glob(filepath.Join(os.TempDir(), "binance-go-update-*", "*"))
	if err != nil {
		return nil
	}
	var out []string
	for _, p := range matches {
		if !strings.Contains(p, "binance-go-update-") {
			continue
		}
		info, err := os.Stat(p)
		if err != nil || info.IsDir() || info.Size() == 0 {
			continue
		}
		out = append(out, p)
	}
	return out
}
