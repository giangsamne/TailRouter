//go:build windows

package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/giangsamne/TailRouter/internal/gateway"
	"github.com/giangsamne/TailRouter/internal/tray"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW    = user32.NewProc("RegisterClassExW")
	procCreateWindowExW     = user32.NewProc("CreateWindowExW")
	procDefWindowProcW      = user32.NewProc("DefWindowProcW")
	procDestroyWindow       = user32.NewProc("DestroyWindow")
	procPostQuitMessage     = user32.NewProc("PostQuitMessage")
	procGetMessageW         = user32.NewProc("GetMessageW")
	procTranslateMessage    = user32.NewProc("TranslateMessage")
	procDispatchMessageW    = user32.NewProc("DispatchMessageW")
	procLoadIconW           = user32.NewProc("LoadIconW")
	procCreatePopupMenu     = user32.NewProc("CreatePopupMenu")
	procAppendMenuW         = user32.NewProc("AppendMenuW")
	procTrackPopupMenu      = user32.NewProc("TrackPopupMenu")
	procGetCursorPos        = user32.NewProc("GetCursorPos")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
	procShell_NotifyIconW   = shell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandleW    = kernel32.NewProc("GetModuleHandleW")
)

const (
	WM_USER          = 0x0400
	WM_TRAYICON      = WM_USER + 100
	WM_COMMAND       = 0x0111
	WM_LBUTTONUP     = 0x0202
	WM_RBUTTONUP     = 0x0205
	WM_LBUTTONDBLCLK = 0x0203

	NIM_ADD    = 0
	NIM_MODIFY = 1
	NIM_DELETE = 2

	NIF_MESSAGE = 1
	NIF_ICON    = 2
	NIF_TIP     = 4

	MF_STRING    = 0x0000
	MF_SEPARATOR = 0x0800
	TPM_BOTTOMALIGN = 0x0020
	TPM_RIGHTALIGN  = 0x0008

	IDI_APPLICATION = 32512

	CMD_OPEN_DASHBOARD = 1001
	CMD_SCAN_PORTS     = 1002
	CMD_QUIT           = 1003
)

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
}

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

var (
	globalHWnd    uintptr
	globalNID     NOTIFYICONDATAW
	globalServer  *gateway.Server
)

func runDesktop() {
	// 1. Khởi chạy Gateway Server ngầm
	globalServer = gateway.NewServer(gateway.DefaultPort, "")
	go func() {
		_ = globalServer.Start()
	}()

	// Mở trình duyệt sau 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = tray.OpenBrowser(fmt.Sprintf("http://localhost:%d/router", gateway.DefaultPort))
	}()

	// 2. Tạo cửa sổ ẩn và biểu tượng khay hệ thống (System Tray)
	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className, _ := syscall.UTF16PtrFromString("TailRouterTrayWindowClass")

	wndClass := WNDCLASSEXW{
		CbSize:      uint32(unsafe.Sizeof(WNDCLASSEXW{})),
		LpfnWndProc: syscall.NewCallback(wndProc),
		HInstance:   hInstance,
		LpszClassName: className,
	}

	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wndClass)))

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("TailRouter Tray"))),
		0,
		0, 0, 0, 0,
		0, 0, hInstance, 0,
	)
	globalHWnd = hwnd

	// Load Icon
	hIcon, _, _ := procLoadIconW.Call(0, uintptr(IDI_APPLICATION))

	// Cấu hình NotifyIcon
	globalNID = NOTIFYICONDATAW{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATAW{})),
		HWnd:             hwnd,
		UID:              1,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: WM_TRAYICON,
		HIcon:            hIcon,
	}
	tip, _ := syscall.UTF16FromString("TailRouter Gateway (Cổng 65534)")
	copy(globalNID.SzTip[:], tip)

	procShell_NotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&globalNID)))

	// Message Loop
	var msg MSG
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}

	// Dọn dẹp
	procShell_NotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&globalNID)))
	if globalServer != nil {
		globalServer.Close()
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAYICON:
		switch lParam {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
			_ = tray.OpenBrowser(fmt.Sprintf("http://localhost:%d/router", gateway.DefaultPort))
		case WM_RBUTTONUP:
			showContextMenu(hwnd)
		}
		return 0

	case WM_COMMAND:
		switch wParam {
		case CMD_OPEN_DASHBOARD:
			_ = tray.OpenBrowser(fmt.Sprintf("http://localhost:%d/router", gateway.DefaultPort))
		case CMD_SCAN_PORTS:
			_ = tray.OpenBrowser(fmt.Sprintf("http://localhost:%d/router#scan", gateway.DefaultPort))
		case CMD_QUIT:
			procDestroyWindow.Call(hwnd)
			procPostQuitMessage.Call(0)
		}
		return 0

	case 0x0002: // WM_DESTROY
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func showContextMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()

	titleStr, _ := syscall.UTF16PtrFromString("⚡ TailRouter v2.0 (Active)")
	procAppendMenuW.Call(hMenu, MF_STRING, 0, uintptr(unsafe.Pointer(titleStr)))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)

	dashStr, _ := syscall.UTF16PtrFromString("🌐 Mở Bảng Quản Trị (/router)")
	procAppendMenuW.Call(hMenu, MF_STRING, uintptr(CMD_OPEN_DASHBOARD), uintptr(unsafe.Pointer(dashStr)))

	scanStr, _ := syscall.UTF16PtrFromString("🔍 Quét Cổng Đang Mở")
	procAppendMenuW.Call(hMenu, MF_STRING, uintptr(CMD_SCAN_PORTS), uintptr(unsafe.Pointer(scanStr)))

	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)

	quitStr, _ := syscall.UTF16PtrFromString("❌ Thoát TailRouter")
	procAppendMenuW.Call(hMenu, MF_STRING, uintptr(CMD_QUIT), uintptr(unsafe.Pointer(quitStr)))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	procSetForegroundWindow.Call(hwnd)
	procTrackPopupMenu.Call(hMenu, TPM_BOTTOMALIGN|TPM_RIGHTALIGN, uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
}
