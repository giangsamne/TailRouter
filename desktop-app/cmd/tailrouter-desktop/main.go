package main

import (
	"os"
)

func main() {
	// Nếu người dùng truyền lệnh qua CLI (ví dụ: TailRouter.exe scan, status...)
	if len(os.Args) > 1 {
		// Gọi hàm CLI chính
		runCLI()
		return
	}

	// Chạy chế độ Desktop Normal (có UX/UI khay hệ thống, mở trình duyệt và chạy ngầm)
	runDesktop()
}
