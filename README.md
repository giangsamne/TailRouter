<div align="center">

# ⚡ TailRouter

**High-performance, zero-dependency bare-metal gateway & desktop manager for Tailscale Serve and Funnel.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![Swift](https://img.shields.io/badge/Swift-5.9+-FA7343.svg?logo=swift&logoColor=white)](https://developer.apple.com/swift/)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Native%20Binaries)-success.svg)](https://github.com/giangsamne/TailRouter)
[![Memory Footprint](https://img.shields.io/badge/RAM-%3C%208MB%20RSS-blueviolet.svg)](https://github.com/giangsamne/TailRouter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

### ⚡ 1-Line Universal Auto-Installer (Recommended)

Run a single command in your terminal or PowerShell — it automatically detects your Operating System, hardware architecture (Apple Silicon / Intel / ARM64 / x86_64), and init system (`systemd`, `openrc`, `launchd`), then downloads, extracts, and configures TailRouter in one shot:

* **🐧 Linux & 🍏 macOS** (Terminal):
  ```bash
  curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/desktop-app/install.sh | bash
  ```
* **🪟 Windows** (PowerShell):
  ```powershell
  irm https://raw.githubusercontent.com/giangsamne/TailRouter/desktop-app/install.ps1 | iex
  ```
* **🌐 Universal Polyglot Script** (Single file for all 3 OS):
  ```bash
  # Windows: Double-click install.cmd (or run in Command Prompt)
  # Linux & macOS:
  ./install.cmd
  ```

---

### 📦 Quick Downloads (Zero Dependencies / Self-Contained)

Download standalone native executables — no runtime or package installation required:

| Operating System | 💻 Normal Edition (Desktop UX/UI + CLI) | ⚡ Server Edition (Headless CLI & Daemon) | Direct Download Link |
| :--- | :--- | :--- | :--- |
| 🍏 **macOS** | **Menu Bar App** (`TailRouter.app`) | `tailrouter-server` (Apple Silicon & Intel) | [**Download `TailRouter-macOS.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-macOS.zip) |
| 🪟 **Windows** | **System Tray App** (`TailRouter.exe`) | `tailrouter-server.exe` (x86_64) | [**Download `TailRouter-Windows.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Windows.zip) |
| 🐧 **Linux** | **Desktop Launcher** (`tailrouter-desktop`) | `tailrouter-server` (x86_64 & ARM64) | [**Download `TailRouter-Linux.tar.gz`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Linux.tar.gz) |
| 💾 **USB Flash Drive** | **Universal Offline Installer** (All 3 OS Bundled) | Self-contained, zero-internet offline package (~15MB) | [**Download `TailRouter-USB-Offline-Installer.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-USB-Offline-Installer.zip) |

> 💡 **All-in-One Architecture**: Both editions embed the full Web Dashboard at port **65534** (`/router`), support all CLI management commands, and when closing the window/UI, **the app continues running in the background** to maintain uninterrupted reverse-proxy routing.

---

### 💾 Offline USB Installation (Share & Install Anywhere)

Carry TailRouter on a USB drive and install offline on any computer:
1. Download [**`TailRouter-USB-Offline-Installer.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-USB-Offline-Installer.zip) and extract it to your USB drive root.
2. Plug the USB into any machine:
   - **🪟 Windows**: Double-click **`Setup.cmd`** (extracts and starts `TailRouter.exe` with desktop shortcut).
   - **🍏 macOS**: Double-click **`Setup.command`** in Finder (installs `TailRouter.app` into `/Applications` and launches).
   - **🐧 Linux**: Run **`./Setup.sh`** in terminal (installs binary and enables systemd/openrc service).

---

### 🌿 Dedicated OS Branches

For developers and users who only need code for their specific operating system:

| Branch | Target Platform | Description | Quick Link |
| :--- | :--- | :--- | :--- |
| **`desktop-app`** | 🌐 Multi-Platform (Hub) | **Default**: Unified hub containing all platform sources, Go engine, and build systems. | [**View `desktop-app` branch 👉**](https://github.com/giangsamne/TailRouter/tree/desktop-app) |
| **`macos`** | 🍏 macOS | Dedicated macOS Menu Bar app (Swift), native Go engine, and macOS instructions. | [**View `macos` branch 👉**](https://github.com/giangsamne/TailRouter/tree/macos) |
| **`windows`** | 🪟 Windows | Dedicated Windows System Tray app (`.exe`), Win32 notify icon, and Windows guide. | [**View `windows` branch 👉**](https://github.com/giangsamne/TailRouter/tree/windows) |
| **`linux`** | 🐧 Linux | Dedicated Linux Desktop launcher, systemd/openrc services, and Raspberry Pi / ARM64 engine. | [**View `linux` branch 👉**](https://github.com/giangsamne/TailRouter/tree/linux) |

---

## 💡 What is TailRouter?

**TailRouter** is an ultra-fast, zero-dependency bare-metal gateway running directly on port **65534** of your host machine. It eliminates the need for heavy reverse proxies (Nginx, Traefik, NPM) or virtual machines to publish local homelab services (3D printers, home automation, web apps, media servers, dev tools) to your **Tailnet** or the **public Internet**.

### 🌟 Key Highlights

- ⚡ **Zero Dependencies & Blazing Fast**:
  - Written in 100% static Go and Swift.
  - **No Python, no Node.js, and no external runtimes required**.
  - Starts in **< 10ms**, consumes **< 8 MB RAM**, and uses **0.0% idle CPU**.
- 🎛️ **Seamless Tailscale Serve & Funnel Integration**:
  - **Serve**: Encrypted, authenticated access strictly within your private Tailnet (`https://<node-name>.ts.net/<path>`).
  - **Funnel**: Fully public HTTPS route accessible across the Internet without port forwarding or public IPv4.
  - Switch between Serve and Funnel in 1 click or via CLI!
- 🖥️ **Embedded Web Dashboard (`/router`)**:
  - Responsive dark-mode SPA embedded inside the binary via `embed.FS`.
  - Accessible locally at `http://localhost:65534/router` and over Tailnet at `https://<node-name>.ts.net/router`.
- 🔍 **Host Port Discovery Scanner**:
  - Scans active TCP listening ports across the host system.
  - Probes HTTP titles and signatures automatically to pre-fill route names and paths.
- 🔁 **CLI & Daemon Live Synchronization**:
  - CLI commands (`routes add`, `routes delete`, `status`) dynamically communicate with the live Gateway daemon via local REST API.
  - Configuration changes automatically hot-reload from disk without process restarts.
- 🛡️ **Built-in Loop & Out-of-Memory Protection**:
  - Prevents recursive routing back to port 65534 (`508 Loop Detected`).
  - Auto-rotating log capped at 5MB.
- 🔄 **Native Boot Autostart**:
  - Supports `systemd` (Ubuntu, Debian, Fedora), `openrc` (Alpine Linux), `launchd` (macOS), and Windows Registry autostart.
- 🌐 **Multi-Language Support (i18n)**:
  - English, Vietnamese, Chinese, and Japanese included out-of-the-box.

---

## 🖥️ System Architecture

```
Internet / Tailnet
         │
         ▼
┌──────────────────────────────────────────────────────────────┐
│                  Tailscale WireGuard Mesh                     │
│         (MagicDNS: https://<your-node>.ts.net)               │
└──────────────────────────────┬───────────────────────────────┘
                               │
                ┌──────────────┴───────────────┐
                ▼ (Serve / Funnel)              ▼ (/router Web Dashboard)
┌──────────────────────────────────────────────────────────────┐
│ TailRouter Native Engine (Port 65534)                        │
│  - Dual-stack IPv4/IPv6 reverse proxy                         │
│  - REST API & Embedded Web Dashboard                         │
│  - Host TCP port scanner & route hot-reload                  │
└──────────────────────────────┬───────────────────────────────┘
                               │
    ┌──────────────────────────┼───────────────────────────────┐
    ▼                          ▼                               ▼
Port 8080                  Port 3000                       Port 22
(Bambu 3D / OctoPrint)     (Web App / Next.js)             (SSH / Remote)
```

---

## 🚀 Quick Start Guide

### Option A: Desktop Apps (Normal Edition)

* **🍏 macOS**:
  1. Download and unzip [**`TailRouter-macOS.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-macOS.zip).
  2. Drag `TailRouter.app` into `/Applications`.
  3. Launch `TailRouter` — an icon will appear in your top Menu Bar. Click to access the dashboard, scan ports, or configure autostart.
* **🪟 Windows**:
  1. Download and unzip [**`TailRouter-Windows.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Windows.zip).
  2. Run `TailRouter.exe`.
  3. A tray icon appears in the Windows System Tray. Right-click to open the dashboard or control routes.
* **🐧 Linux Desktop**:
  1. Download and extract [**`TailRouter-Linux.tar.gz`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Linux.tar.gz).
  2. Run `./tailrouter-desktop` to launch the background service and open the browser.

---

### Option B: Server CLI & Headless Daemon

For servers, Raspberry Pi, homelab nodes, or cloud VMs:

```bash
# 1. Download the static binary for your architecture (amd64 or arm64)
curl -sL https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Linux.tar.gz | tar -xz

# 2. Start the gateway daemon
./tailrouter-server run

# 3. Check gateway status
./tailrouter-server status

# 4. Scan open host ports
./tailrouter-server scan

# 5. Enable auto-start on system boot
./tailrouter-server service install
```

---

## ⌨️ CLI Command Reference

Both `tailrouter-server` and `tailrouter-desktop` include a full CLI:

```bash
# Run the gateway server (default port 65534)
tailrouter run [--port 65534]

# Scan active TCP ports on the host
tailrouter scan

# Inspect running gateway status and Tailscale connection
tailrouter status

# List all configured routes
tailrouter routes list

# Add a new route
tailrouter routes add <name> <path> <port> [serve|funnel]
# Example:
tailrouter routes add "Grafana Dashboard" /grafana 3000 serve

# Delete a route by ID or path
tailrouter routes delete <id|path>
# Example:
tailrouter routes delete /grafana

# Automatic Tailscale Serve configuration
tailrouter serve router     # Expose /router web interface over Tailscale Serve
tailrouter serve gateway    # Forward entire host gateway via Tailscale Serve
tailrouter serve reset      # Reset Tailscale Serve configuration

# System service management
tailrouter service install    # Enable automatic start on boot (systemd/openrc/launchd)
tailrouter service status     # Check background service health
tailrouter service uninstall  # Remove system service
```

---

## 📡 REST API Reference

TailRouter exposes a high-performance REST API on port `65534`:

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/status` | Node status, Tailscale state, gateway uptime, autostart info. |
| `GET` | `/api/routes` | List all configured routes with latency and hit counts. |
| `POST` | `/api/routes` | Register a new reverse-proxy route. |
| `GET` | `/api/routes/{id}` | Inspect details of a specific route. |
| `PUT` | `/api/routes/{id}` | Update route target, name, or settings. |
| `DELETE` | `/api/routes/{id}` | Delete a route and unmap it from Tailscale. |
| `POST` | `/api/routes/{id}/ping` | Health-check backend target and return latency in ms. |
| `POST` | `/api/routes/{id}/toggle` | Enable or disable a route without deleting it. |
| `POST` | `/api/routes/{id}/toggle-mode` | Switch route between `serve` (private) and `funnel` (public). |
| `GET` | `/api/scan` | Trigger immediate host TCP port discovery scan. |
| `GET` | `/api/autostart` | Inspect operating system autostart configuration. |
| `POST` | `/api/autostart/toggle` | Enable or disable autostart on boot. |

---

## 📁 Repository Structure

```text
TailRouter/
├── app/
│   ├── macos/                # macOS Swift Menu Bar app (TailRouterMenuApp.swift)
│   ├── windows/              # Windows System Tray launcher scripts
│   └── linux/                # Linux desktop entry & installer
├── cmd/
│   ├── tailrouter/           # Core Headless Server CLI engine (Go)
│   └── tailrouter-desktop/   # Desktop GUI launcher & CLI wrapper (Go)
├── internal/
│   ├── gateway/              # Reverse proxy, REST API & embedded web server
│   ├── scanner/              # Host TCP port discovery engine
│   ├── config/               # JSON route persistence with disk hot-reload
│   ├── service/              # Systemd, OpenRC, Launchd, and Windows autostart
│   ├── tailscale/            # Tailscale Serve & Funnel CLI integration
│   └── tray/                 # Win32 notify icon & system tray support
├── web/
│   ├── embed.go              # Go binary asset packaging via embed.FS
│   └── index.html            # Responsive SPA Web Dashboard with i18n
├── scripts/
│   ├── build_all.sh          # Multi-platform static compilation tool
│   └── snapshot_tag.sh       # All-in-One release snapshot packaging tool
├── LICENSE                   # MIT License
└── README.md                 # English documentation
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Feel free to open an issue or submit a pull request on the [GitHub Issues page](https://github.com/giangsamne/TailRouter/issues).

---

## 📜 License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for more details.
