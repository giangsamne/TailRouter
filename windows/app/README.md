# TailRouter Desktop Apps (macOS, Windows, Linux)

Thư mục này chứa toàn bộ các công cụ đóng gói và giao diện khay hệ thống (Menu Bar / System Tray) của **TailRouter** trên nhánh `desktop-app`.

---

## 🍏 1. macOS (Menu Bar Native App)
Ứng dụng **`TailRouter.app`** được viết bằng **Swift Native**, chạy hoàn toàn ngầm không chiếm chỗ thanh Dock (`LSUIElement = true`), hiển thị icon trên thanh Menu Bar góc trên bên phải.

### Tính năng:
- Tự động phát hiện và khởi động `server.py` ngầm khi mở app.
- Icon Menu Bar trực quan (🟢 Đang chạy, 🔴 Đã dừng).
- Bấm vào icon để mở nhanh Bảng Quản Trị `/router`, xem IP Tailscale, quét cổng.
- Bật/tắt tự khởi động cùng máy Mac (Login Items).

### Cách đóng gói:
```bash
./app/macos/build_app.sh
```
Sau khi đóng gói, file **`TailRouter.app`** sẽ xuất hiện tại thư mục `dist/`. Kéo vào `/Applications` để sử dụng.

---

## 🪟 2. Windows (System Tray Agent)
Ứng dụng khay hệ thống Windows nằm ở góc dưới bên phải (Notification Area cạnh đồng hồ), chạy hoàn toàn ngầm không mở cửa sổ Command Prompt đen.

### Cách sử dụng:
1. Nhấp đúp vào **`app/windows/TailRouter-Tray.vbs`** để khởi chạy ngầm.
2. Icon TailRouter sẽ xuất hiện dưới góc phải màn hình.
3. Nhấp đúp hoặc chuột phải vào icon để mở Bảng Quản Trị `/router`, khởi động lại hoặc thoát.

---

## 🐧 3. Linux (Desktop Application & Launcher)
Tích hợp trực tiếp vào GNOME, KDE, XFCE Application Launcher với đầy đủ icon vector.

### Cách cài đặt:
```bash
./app/linux/install-desktop.sh
```
Sau khi cài đặt, bạn có thể tìm thấy biểu tượng **TailRouter** trong danh sách ứng dụng của hệ thống để mở nhanh 1-click.
