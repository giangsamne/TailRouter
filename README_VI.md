<div align="center">

# ⚡ TailRouter (Cổng 65534)

**Hệ thống Gateway điều phối Reverse-Proxy & Web Dashboard siêu nhẹ, zero-dependency, chạy trực tiếp trên máy vật lý kết hợp Tailscale Serve & Funnel.**

[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Stdlib)-success.svg)](https://docs.python.org/3/library/asyncio.html)
[![i18n](https://img.shields.io/badge/i18n-VI%20%7C%20EN%20%7C%20ZH%20%7C%20JA-orange.svg)](#-đa-ngôn-ngữ-i18n)

[**English Documentation**](README.md) | [**Tài Liệu Tiếng Việt**](README_VI.md)

</div>

---

### 📦 Tải Nhanh Bản Phát Hành v2.0 (Không Cần Cài Python)

Tải file chạy trực tiếp, nhấp đúp là dùng ngay:

| Hệ điều hành | 💻 Bản Normal (Desktop UX/UI + CLI) | ⚡ Bản Server (CLI Headless) | Tải Về Trực Tiếp |
| :--- | :--- | :--- | :--- |
| 🍏 **macOS** | **Menu Bar App** (`TailRouter.app`) | `tailrouter-server` | [**Tải `TailRouter-macOS.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-macOS.zip) |
| 🪟 **Windows** | **Khay Hệ Thống** (`TailRouter.exe`) | `tailrouter-server.exe` | [**Tải `TailRouter-Windows.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Windows.zip) |
| 🐧 **Linux** | **Desktop Launcher** (`tailrouter-desktop`) | `tailrouter-server` (x86_64 & ARM64) | [**Tải `TailRouter-Linux.tar.gz`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Linux.tar.gz) |

> 💡 **Cơ chế All-in-One Gateway**: Cả 2 bản đều nhúng sẵn toàn bộ Web Dashboard tại cổng **65534** (`/router`), hỗ trợ đầy đủ lệnh CLI, và khi đóng cửa sổ giao diện **ứng dụng vẫn chạy ngầm** ở khay hệ thống để duy trì định tuyến.

---

### 🌿 Các Nhánh Chuyên Biệt Theo Từng Hệ Điều Hành

| Nhánh (Branch) | Hệ điều hành | Mô tả | Đường dẫn xem code |
| :--- | :--- | :--- | :--- |
| **`macos`** | 🍏 macOS | Chỉ chứa mã nguồn Menu Bar app (Swift), native engine và hướng dẫn macOS. | [**Xem nhánh `macos` 👉**](https://github.com/giangsamne/TailRouter/tree/macos) |
| **`windows`** | 🪟 Windows | Chỉ chứa mã nguồn System Tray app, native engine và hướng dẫn Windows. | [**Xem nhánh `windows` 👉**](https://github.com/giangsamne/TailRouter/tree/windows) |
| **`linux`** | 🐧 Linux | Chỉ chứa mã nguồn Desktop app, systemd service, CLI và Raspberry Pi. | [**Xem nhánh `linux` 👉**](https://github.com/giangsamne/TailRouter/tree/linux) |
| **`v0`** | 🐍 Bản Thử Nghiệm Python | Bản prototype đầu tiên viết bằng Python 3 stdlib (lưu trữ tham khảo). | [**Xem nhánh `v0` 👉**](https://github.com/giangsamne/TailRouter/tree/v0) |
| **`desktop-app`** | 🌐 Tổng hợp (Hub) | Nhánh mặc định tổng hợp mã nguồn đa nền tảng. | [**Nhánh `desktop-app`**](https://github.com/giangsamne/TailRouter/tree/desktop-app) |

---

## 💡 Giới Thiệu Dự Án

**TailRouter** là giải pháp gateway chạy trực tiếp trên máy vật lý (bare-metal) tại cổng **65534**, giúp bạn gom và mở tất cả các cổng, ứng dụng nội bộ (máy in 3D Bambu Timelapse, NoMachine, Home Assistant, Docker container, Next.js, camera, v.v.) ra ngoài mạng **Tailscale** hoặc **Internet công cộng** một cách dễ dàng và an toàn.

Không cần cài máy ảo, không cần cấu hình phức tạp như Nginx/Traefik/NPM!

### 🌟 Tính Năng Nổi Bật
- 🪶 **Zero Dependencies (0 thư viện ngoài)**: Chạy hoàn toàn bằng thư viện tiêu chuẩn của Python 3 (`asyncio`, `urllib`, `json`, `subprocess`). Không cần cài `pip`, không cần `npm`. Tiêu thụ chỉ ~15MB RAM và khởi động trong 50ms.
- ⚡ **Chạy trực tiếp máy thật (Bare-metal)**: Đạt hiệu năng tối đa cho tác vụ nặng (như stream video timelapse máy in 3D, truyền tệp dung lượng lớn).
- 🎛️ **Tích hợp sâu Tailscale Serve & Funnel**:
  - **Serve**: Truy cập mã hóa nội bộ trong mạng Tailnet của bạn (`https://<node-name>.ts.net/<path>`).
  - **Funnel**: Mở công khai ra toàn thế giới qua HTTPS mà không cần mở cổng modem (NAT/Port Forwarding) hay cần IP tĩnh!
  - Chuyển đổi qua lại giữa Serve và Funnel chỉ với **1 cú click chuột** trên giao diện Web.
- 🖥️ **Web Dashboard hiện đại tại `/router`**:
  - Truy cập tại `http://localhost:65534/router` hoặc `https://<node-name>.ts.net/router`.
  - Giao diện Dark-Mode đẹp mắt, phản hồi tức thì.
- 🔍 **Tự động quét cổng (Auto Port Scanner)**:
  - Tự động phát hiện các cổng TCP đang mở trên máy (Linux `ss`, macOS `lsof`, Windows `netstat`) và các container Docker.
  - Tự động kiểm tra HTTP title và chữ ký dịch vụ.
  - Nút bấm 1-click **"+ Thêm vào Router"** tự điền thông tin để tạo route trong nháy mắt.
- 🌐 **Đa ngôn ngữ (i18n)**: Hỗ trợ sẵn 4 ngôn ngữ: 🇻🇳 Tiếng Việt, 🇺🇸 Tiếng Anh, 🇨🇳 Tiếng Trung, và 🇯🇵 Tiếng Nhật.
- 💻 **Chạy mượt trên mọi hệ điều hành**: Kèm sẵn script chạy cho **Linux**, **macOS**, **Windows**, và **Docker**.
- 🔄 **Tự khởi động cùng hệ thống**: Hỗ trợ Systemd user service (Linux), tự chạy lại các route đã lưu khi khởi động lại máy.

---

## 🚀 Hướng Dẫn Sử Dụng Nhanh

### Yêu cầu tiên quyết
- Máy tính đã cài đặt **Python 3.8+** trở lên.
- Đã cài và đăng nhập **[Tailscale](https://tailscale.com)**.

---

### 🐧 Dành cho Linux & 🍏 macOS

1. **Tải mã nguồn về máy**:
   ```bash
   git clone https://github.com/giangsamne/TailRouter.git
   cd TailRouter
   ```

2. **Khởi động**:
   ```bash
   # Chạy nền (daemon):
   ./start.sh

   # Hoặc chạy trực tiếp xem log trực quan:
   ./start.sh -f
   ```

3. **Dừng hoạt động**:
   ```bash
   ./stop.sh
   ```

4. **Cài đặt tự khởi động khi bật máy (Systemd)**:
   ```bash
   ./install-service.sh
   ```
   *Script tự động kích hoạt `loginctl enable-linger` giúp dịch vụ khởi động ngay khi máy bật mà không cần đăng nhập màn hình.*

---

### 🪟 Dành cho Windows

1. **Tải mã nguồn**:
   ```cmd
   git clone https://github.com/giangsamne/TailRouter.git
   cd TailRouter
   ```

2. **Khởi động**:
   Nhấp đúp chuột vào file `start.bat` (hoặc chạy qua CMD / PowerShell).

3. **Dừng hoạt động**:
   Nhấp đúp chuột vào file `stop.bat`.

---

### 🐳 Dành cho Docker / Docker Compose

Nếu bạn thích chạy qua Docker:
```bash
docker compose up -d
```
*(Cấu hình sử dụng `network_mode: host` để có thể nhận diện trực tiếp các cổng trên máy thật và daemon Tailscale).*

---

## 🧭 Hướng Dẫn Sử Dụng Web Dashboard (`/router`)

Sau khi khởi động, bạn mở trình duyệt và truy cập:
- **Từ máy thật**: [http://localhost:65534/router](http://localhost:65534/router)
- **Từ thiết bị khác trong Tailnet**: `http://<tailscale-ip>:65534/router`
- **Qua HTTPS MagicDNS**: `https://<ten-may>.ts.net/router`

### Các tính năng trên giao diện:
1. **Thông tin máy chủ**: Hiển thị Hostname, IP Tailscale, Tên miền MagicDNS FQDN, và thời gian hoạt động.
2. **Bảng Danh Sách Router**:
   - Đường dẫn (ví dụ: `/bambu`, `/nomachine`, `/api`).
   - Cổng đích (`127.0.0.1:8080`).
   - Chế độ: **🔒 Serve (Chỉ mạng nội bộ)** hoặc **🌐 Funnel (Công khai Internet)**.
   - Nút chuyển đổi nhanh Serve $\leftrightarrow$ Funnel tức thì.
   - Nút kiểm tra Ping dịch vụ đo độ trễ mili-giây.
   - Nút mở link MagicDNS và nút copy link vào clipboard.
3. **Tab Quét Cổng Tự Động**:
   - Tự dò các tiến trình đang mở cổng TCP và container Docker.
   - Bấm **"+ Thêm vào Router"** để đưa cổng vào danh sách phân phối.
4. **Nút đổi ngôn ngữ**:
   - Chọn Tiếng Việt, English, 中文, hoặc 日本語 ở góc trên màn hình.

---

## 📡 Tài Liệu REST API

| Giao thức | Đường dẫn | Chức năng |
|---|---|---|
| `GET` | `/api/status` | Xem thông tin hệ thống, trạng thái Tailscale |
| `GET` | `/api/routes` | Lấy danh sách tất cả các route |
| `POST` | `/api/routes` | Tạo một route mới |
| `PUT` | `/api/routes/<id>` | Cập nhật cấu hình route |
| `POST` | `/api/routes/<id>/toggle-mode` | Đổi nhanh giữa chế độ `serve` và `funnel` |
| `POST` | `/api/routes/<id>/ping` | Kiểm tra kết nối đến cổng dịch vụ đích |
| `DELETE` | `/api/routes/<id>` | Xóa route và hủy gán Tailscale |
| `GET` | `/api/scan` | Quét các cổng TCP và Docker đang mở |

---

## 📁 Cấu Trúc Thư Mục

```
TailRouter/
├── server.py              # Server chính xử lý Reverse-Proxy & REST API cổng 65534
├── port_scanner.py        # Module quét cổng TCP đa nền tảng & Docker
├── routes_manager.py      # Module quản lý danh sách route & lưu file JSON
├── tailscale_helper.py    # Wrapper gọi lệnh Tailscale CLI (Serve, Funnel, MagicDNS)
├── web/
│   └── index.html         # Giao diện Web SPA đa ngôn ngữ tại /router
├── config/
│   ├── routes.json        # File lưu danh sách route hiện tại
│   └── routes.example.json# File cấu hình mẫu
├── start.sh               # Script khởi động cho Linux & macOS
├── stop.sh                # Script dừng cho Linux & macOS
├── install-service.sh     # Script cài đặt dịch vụ Systemd tự chạy khi bật máy
├── start.bat              # Script khởi động cho Windows
├── stop.bat               # Script dừng cho Windows
├── Dockerfile             # File build Docker image
├── docker-compose.yml     # File khởi chạy Docker Compose
├── LICENSE                # Giấy phép MIT
├── README.md              # Hướng dẫn tiếng Anh
└── README_VI.md           # Hướng dẫn tiếng Việt
```

---

## 📜 Giấy Phép (License)

Dự án được phát hành theo giấy phép **MIT License**. Xem chi tiết tại [LICENSE](LICENSE).
