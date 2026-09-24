//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/giangsamne/TailRouter/internal/gateway"
	"github.com/giangsamne/TailRouter/internal/tray"
)

func runDesktop() {
	// 1. Khởi động Gateway ngầm
	srv := gateway.NewServer(gateway.DefaultPort, "")
	go func() {
		if err := srv.Start(); err != nil {
			fmt.Printf("[TailRouter] Lỗi khởi động: %v\n", err)
			os.Exit(1)
		}
	}()

	// 2. Chờ 500ms để Gateway sẵn sàng và mở Web Dashboard trên trình duyệt
	go func() {
		time.Sleep(500 * time.Millisecond)
		dashboardURL := fmt.Sprintf("http://localhost:%d/router", gateway.DefaultPort)
		_ = tray.OpenBrowser(dashboardURL)
	}()

	// 3. Giữ tiến trình chạy ngầm liên tục
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	srv.Close()
	os.Exit(0)
}
