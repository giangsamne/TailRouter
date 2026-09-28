<div align="center">

# ⚡ TailRouter

**High-performance, zero-dependency bare-metal gateway & web dashboard for Tailscale Serve and Funnel.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Native%20Binaries)-success.svg)](https://github.com/giangsamne/TailRouter)
[![Memory Footprint](https://img.shields.io/badge/RAM-%3C%208MB%20RSS-blueviolet.svg)](https://github.com/giangsamne/TailRouter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ 1-Line Universal Auto-Installer (Recommended)

Run a single command in your terminal or PowerShell — it automatically detects your Operating System, hardware architecture (Apple Silicon / Intel / ARM64 / x86_64), and init system (`systemd`, `openrc`, `launchd`), then installs and starts TailRouter in one shot:

* **🐧 Linux & 🍏 macOS** (Terminal):
  ```bash
  curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/v1.0.1/install.sh | bash
  ```
* **🪟 Windows** (PowerShell):
  ```powershell
  irm https://raw.githubusercontent.com/giangsamne/TailRouter/v1.0.1/install.ps1 | iex
  ```

---

### 🌿 Dedicated OS Branches

TailRouter provides dedicated platform branches with optimized guides, service setups, and configurations:

| Branch | Platform | Features | Direct Link |
| :--- | :--- | :--- | :--- |
| **`main`** | 🌐 Multi-Platform (Hub) | Default hub: Unified guides, multi-OS installers, release artifacts | [**View `main` branch 👉**](https://github.com/giangsamne/TailRouter/tree/main) |
| **`linux`** | 🐧 Linux | Native CLI + Web Control (:65534), systemd & OpenRC (Alpine/Ubuntu/Arch) | [**View `linux` branch 👉**](https://github.com/giangsamne/TailRouter/tree/linux) |
| **`windows`** | 🪟 Windows | Native CLI (`.exe`) + Web Control (:65534), PowerShell installer, Startup autostart | [**View `windows` branch 👉**](https://github.com/giangsamne/TailRouter/tree/windows) |
| **`macos`** | 🍏 macOS | Native CLI + Web Control (:65534), launchd agent, Apple Silicon & Intel | [**View `macos` branch 👉**](https://github.com/giangsamne/TailRouter/tree/macos) |


---

## 💡 What is TailRouter?

**TailRouter** is an ultra-fast, zero-dependency bare-metal gateway running directly on port **65534** of your host machine. It eliminates the need for heavy reverse proxies (Nginx, Traefik, NPM) or virtual machines to publish local homelab services (3D printers, home automation, web apps, media servers, dev tools) to your **Tailnet** or the **public Internet**.

### 🌟 Key Highlights

- ⚡ **Zero Dependencies & Blazing Fast**:
  - Written in 100% pure static Go.
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
- 🔁 **Persistent Route Storage**:
  - Automatically saves routes to standard system locations (`~/.config/tailrouter/routes.json` or `%APPDATA%\TailRouter\routes.json`).
  - Survives updates, binary moves, and reboots seamlessly.
- 🔄 **Native Boot Autostart**:
  - Supports `systemd` (Ubuntu, Debian, Fedora), `openrc` (Alpine Linux), `launchd` (macOS), and Windows Registry autostart via `tailrouter service install`.

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

### 1. Launching TailRouter
Simply double-click `tailrouter` (or `tailrouter.exe` on Windows), or run it in your terminal:
```bash
tailrouter
```
* The gateway automatically starts on port `65534`.
* Your default web browser immediately opens the Web Dashboard at:
  👉 **`http://localhost:65534/router`**

### 2. Running as a Permanent Background Service
To make TailRouter start automatically every time your computer boots:
```bash
tailrouter service install
```
* **Linux**: Configured via `systemd` user service or `openrc`.
* **macOS**: Configured via `launchd` user agent.
* **Windows**: Configured via User Run registry key.

---

## ⌨️ CLI Command Reference

`tailrouter` provides an intuitive command-line interface:

```bash
# Run gateway (starts server and opens browser if no args provided)
tailrouter
tailrouter run [--port 65534] [--open]

# Inspect running gateway status and Tailscale connection
tailrouter status

# Scan active TCP ports on the host
tailrouter scan

# List all configured routes
tailrouter routes list

# Add a new route
tailrouter routes add <name> <path> <port> [serve|funnel]
# Example:
tailrouter routes add "Bambu 3D Printer" /bambu 8080 serve

# Delete a route by ID or path
tailrouter routes delete <id|path>
# Example:
tailrouter routes delete /bambu

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
├── cmd/
│   └── tailrouter/           # Single-binary engine: Daemon + Embedded Web UI + CLI (Go)
├── internal/
│   ├── gateway/              # Reverse proxy, REST API & embedded web server
│   ├── scanner/              # Host TCP port discovery engine
│   ├── config/               # JSON route persistence with standard system path resolution
│   ├── service/              # Systemd, OpenRC, Launchd, and Windows autostart
│   ├── tailscale/            # Tailscale Serve & Funnel CLI integration
│   └── browser/              # Cross-platform browser launcher
├── web/
│   ├── embed.go              # Go binary asset packaging via embed.FS
│   └── index.html            # Responsive SPA Web Dashboard with i18n
├── scripts/
│   ├── build_all.sh          # Multi-platform static compilation tool
│   └── snapshot_tag.sh       # Release snapshot packaging tool
├── install.sh                # 1-Line universal installer for Linux & macOS
├── install.ps1               # 1-Line installer for Windows (PowerShell)
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
