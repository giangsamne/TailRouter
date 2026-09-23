<div align="center">

# ⚡ TailRouter

**High-performance, zero-dependency bare-metal gateway & web manager for Tailscale Serve and Funnel.**

[![Python 3.8+](https://img.shields.io/badge/python-3.8+-blue.svg)](https://www.python.org/downloads/)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Stdlib)-success.svg)](https://docs.python.org/3/library/asyncio.html)
[![i18n](https://img.shields.io/badge/i18n-VI%20%7C%20EN%20%7C%20ZH%20%7C%20JA-orange.svg)](#-internationalization-i18n)

[**Tiếng Việt (README_VI.md)**](README_VI.md) | [**English Documentation**](README.md)

</div>

---

### 📦 Quick Downloads v2.0 (Zero Dependencies / No Python Needed)

Download standalone native executables — double-click to run:

| Operating System | 💻 Normal Edition (Desktop UX/UI + CLI) | ⚡ Server Edition (Headless CLI) | Direct Download Link |
| :--- | :--- | :--- | :--- |
| 🍏 **macOS** | **Menu Bar App** (`TailRouter.app`) | `tailrouter-server` | [**Download `TailRouter-macOS.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-macOS.zip) |
| 🪟 **Windows** | **System Tray** (`TailRouter.exe`) | `tailrouter-server.exe` | [**Download `TailRouter-Windows.zip`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Windows.zip) |
| 🐧 **Linux** | **Desktop Launcher** (`tailrouter-desktop`) | `tailrouter-server` (x86_64 & ARM64) | [**Download `TailRouter-Linux.tar.gz`**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-Linux.tar.gz) |

> 💡 **All-in-One Gateway Architecture**: Both editions embed the full Web Dashboard at port **65534** (`/router`), support all CLI commands, and when closing the window/UI, **the app continues running in the background** to maintain network routing.

---

### 🌿 Dedicated OS Branches

| Branch | Platform | Description | Link |
| :--- | :--- | :--- | :--- |
| **`macos`** | 🍏 macOS | Dedicated macOS Menu Bar app (Swift), native engine & guide. | [**View `macos` branch 👉**](https://github.com/giangsamne/TailRouter/tree/macos) |
| **`windows`** | 🪟 Windows | Dedicated Windows System Tray app, native engine & guide. | [**View `windows` branch 👉**](https://github.com/giangsamne/TailRouter/tree/windows) |
| **`linux`** | 🐧 Linux | Dedicated Linux Desktop launcher, systemd daemon & Raspberry Pi. | [**View `linux` branch 👉**](https://github.com/giangsamne/TailRouter/tree/linux) |
| **`desktop-app`** | 🌐 Main Hub | Default multi-platform hub branch. | [**Branch `desktop-app`**](https://github.com/giangsamne/TailRouter/tree/desktop-app) |

> 📜 *Looking for the historical experimental Python prototype? Check out [**Release v0.1.0**](https://github.com/giangsamne/TailRouter/releases/tag/v0.1.0).*

---

## 💡 What is TailRouter?

**TailRouter** is a lightweight, bare-metal gateway running directly on port **65534** on your host machine. It eliminates the need for heavy virtual machines or complex reverse-proxy setups (Nginx, Traefik, NPM) to expose local homelab services (3D printers, home automation, web apps, media servers, Docker containers) to your Tailnet or the public Internet.

### 🌟 Key Highlights
- 🪶 **Zero External Dependencies**: Built entirely with Python 3's standard library (`asyncio`, `urllib`, `json`, `subprocess`). No `pip install`, no virtualenv, no node_modules. Uses ~15MB RAM and starts in under 50ms.
- ⚡ **Direct Bare-Metal Performance**: Runs straight on host hardware for maximum I/O throughput and ultra-low latency.
- 🎛️ **Tailscale Serve & Funnel Integration**:
  - **Serve**: Encrypted, authenticated access strictly within your private Tailnet (`https://<node-name>.ts.net/<path>`).
  - **Funnel**: Fully public HTTPS route accessible to the entire world without requiring port forwarding or a public IPv4.
  - Switch between Serve and Funnel with a single click in the UI!
- 🖥️ **Web Dashboard (`/router`)**: Responsive Dark-mode SPA at `http://<ip>:65534/router` and `https://<node-name>.ts.net/router`.
- 🔍 **Auto Port Scanner**: Scans open TCP ports and Docker containers on the host, inspects HTTP response headers, and lets you add routes with a single click.
- 🌐 **Multi-Language (i18n)**: Out-of-the-box support for 🇻🇳 Tiếng Việt, 🇺🇸 English, 🇨🇳 中文, and 🇯🇵 日本語.
- 💻 **Cross-Platform**: Ready-to-use scripts for **Linux**, **macOS**, **Windows**, and **Docker**.
- 🔄 **Auto-Start on Boot**: Systemd user service support with auto-recovery and persistent state.

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
               ┌───────────────┴───────────────┐
               ▼ (Serve / Funnel)              ▼ (/router Web UI)
┌──────────────────────────────────────┐       │
│ TailRouter (Port 65534)   │◄──────┘
│  - Async reverse proxy               │
│  - REST API & Web Dashboard          │
│  - Auto Port & Container Scanner     │
└──────────────┬───────────────────────┘
               │
   ┌───────────┼────────────────────────┐
   ▼           ▼                        ▼
Port 8080   Port 3000                Port 4000
(Bambu)     (Web / Next.js)         (NoMachine / Remote)
```

---

## 🚀 Quick Start

### Prerequisites
- Python 3.8 or newer (standard installation).
- [Tailscale](https://tailscale.com) installed and logged in.

---

### 🐧 Linux & 🍏 macOS

1. **Clone the repository**:
   ```bash
   git clone https://github.com/giangsamne/TailRouter.git
   cd TailRouter
   ```

2. **Start the gateway**:
   ```bash
   # Start as a background daemon:
   ./start.sh

   # Or run in foreground with live console logs:
   ./start.sh -f
   ```

3. **Stop the gateway**:
   ```bash
   ./stop.sh
   ```

4. **Enable auto-start on boot (Systemd)**:
   ```bash
   ./install-service.sh
   ```
   *The script enables lingering so the service launches on system boot even before user login.*

---

### 🪟 Windows

1. **Clone or download the repo**:
   ```cmd
   git clone https://github.com/giangsamne/TailRouter.git
   cd TailRouter
   ```

2. **Start the gateway**:
   Double-click `start.bat` or run:
   ```cmd
   start.bat
   ```

3. **Stop the gateway**:
   Double-click `stop.bat` or run:
   ```cmd
   stop.bat
   ```

---

### 🐳 Docker & Docker Compose (Optional)

```bash
docker compose up -d
```
*(Uses `network_mode: host` to directly access host ports and Tailscale daemon).*

---

## 🧭 Web Management Interface (`/router`)

Once started, open your browser and navigate to:
- **Local Host**: [http://localhost:65534/router](http://localhost:65534/router)
- **Within Tailnet**: `http://<tailscale-ip>:65534/router`
- **Via Tailscale Serve**: `https://<your-node>.ts.net/router`

### Key Features of the Dashboard:
1. **Node Status Card**: Shows machine hostname, Tailscale IP, MagicDNS FQDN, and system uptime.
2. **Active Routes Table**:
   - Path (`/bambu`, `/grafana`, `/api`, etc.).
   - Target destination (`127.0.0.1:8080`).
   - Mode badge: **🔒 Serve (Tailnet Only)** or **🌐 Funnel (Public Internet)**.
   - Quick Mode Toggle button (switch between Serve and Funnel in 1 click).
   - Test Ping (measures backend latency in milliseconds).
   - Direct MagicDNS link button with 1-click clipboard copy.
3. **Auto Port Scanner Tab**:
   - Lists all listening TCP ports and running Docker containers.
   - Probes HTTP titles and signatures automatically.
   - Single-click **"+ Add to Router"** button to prefill route creation modal.
4. **Multi-Language Selector**:
   - Switch between Vietnamese, English, Chinese, and Japanese.
   - Saved automatically to browser `localStorage`.

---

## 📡 REST API Documentation

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/status` | Get node info, Tailscale status, and gateway health. |
| `GET` | `/api/routes` | List all configured routes. |
| `POST` | `/api/routes` | Create a new route. |
| `PUT` | `/api/routes/<id>` | Update an existing route. |
| `POST` | `/api/routes/<id>/toggle-mode` | Toggle route between `serve` and `funnel`. |
| `POST` | `/api/routes/<id>/ping` | Health-check backend target and return latency ms. |
| `DELETE` | `/api/routes/<id>` | Delete route and unmap from Tailscale. |
| `GET` | `/api/scan` | Trigger host TCP port and Docker container scan. |

---

## 📁 Repository Structure

```
TailRouter/
├── server.py              # Main async HTTP reverse-proxy & REST API server
├── port_scanner.py        # Cross-platform TCP port & Docker container scanner
├── routes_manager.py      # Route state persistence & health checker
├── tailscale_helper.py    # Tailscale CLI wrapper (Serve, Funnel, MagicDNS discovery)
├── web/
│   └── index.html         # Responsive SPA Web Dashboard with i18n
├── config/
│   ├── routes.json        # Active route configuration (auto-created)
│   └── routes.example.json# Example route configuration template
├── start.sh               # Startup script for Linux / macOS
├── stop.sh                # Stop script for Linux / macOS
├── install-service.sh     # Systemd user service installer
├── start.bat              # Startup batch script for Windows
├── stop.bat               # Stop batch script for Windows
├── Dockerfile             # Container definition
├── docker-compose.yml     # Docker compose recipe
├── LICENSE                # MIT License
├── README.md              # English documentation
└── README_VI.md           # Vietnamese documentation
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Feel free to check [issues page](https://github.com/giangsamne/TailRouter/issues).

---

## 📜 License

Distributed under the **MIT License**. See [LICENSE](LICENSE) for more details.
