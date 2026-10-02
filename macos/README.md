<div align="center">

# 🐍 TailRouter (Python Edition v1.1.1)

**Lightweight, zero-external-dependency gateway & web dashboard for Tailscale Serve and Funnel.**

[![Python](https://img.shields.io/badge/Python-3.8+-3776AB.svg?logo=python&logoColor=white)](https://python.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter/tree/v1.1.1)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Stdlib)-success.svg)](https://github.com/giangsamne/TailRouter/tree/v1.1.1)
[![Memory Footprint](https://img.shields.io/badge/RAM-%3C%2025MB%20RSS-blueviolet.svg)](https://github.com/giangsamne/TailRouter/tree/v1.1.1)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ Quick Start (Recommended)

TailRouter Python Edition relies 100% on the Python Standard Library. **No `pip install`**, virtual environment, or external wheels are required:

* **🐧 Linux & 🍏 macOS** (Terminal):
  ```bash
  git clone -b v1.1.1 https://github.com/giangsamne/TailRouter.git
  cd TailRouter
  ./start.sh
  ```

* **🪟 Windows** (Command Prompt / PowerShell):
  ```cmd
  git clone -b v1.1.1 https://github.com/giangsamne/TailRouter.git
  cd TailRouter
  start.bat
  ```

* **Direct Foreground Execution** (All Operating Systems):
  ```bash
  python3 server.py
  ```

> 💡 **Ready in Seconds**: Once started, open **`http://localhost:65534/router`** to access the Web Control Dashboard.

---

### 🏷️ Releases & Engine Comparison

| Release | Technology / Engine | Characteristics | Direct Link |
| :--- | :--- | :--- | :--- |
| **`v2.1.0`** *(Latest)* | ⚡ Go Native Bare-Metal | Single pre-compiled binary (`< 8MB RAM`, `0.0% CPU`), no runtime required, multi-platform | [**View v2.1.0 Release 👉**](https://github.com/giangsamne/TailRouter/releases/tag/v2.1.0) |
| **`v1.1.1`** | 🐍 Pure Python 3 Engine | Zero external dependencies (Pure stdlib), ascending port scanner with process & PID resolution | [**View v1.1.1 Tag 👉**](https://github.com/giangsamne/TailRouter/tree/v1.1.1) |

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

**TailRouter** is a lightweight reverse-proxy gateway and dashboard running directly on port **65534** of your host machine. It makes publishing local homelab services (3D printers, home automation, web apps, media servers, dev tools) to your **Tailnet** or the **public Internet** effortless without third-party Docker proxies.

### 🌟 Key Highlights

- ⚡ **Zero External Dependencies**:
  - Written in 100% Python Standard Library (`asyncio`, `urllib`, `socket`, `subprocess`).
  - No `pip install` required. Works out of the box on Linux, macOS, and Windows.
- 🎛️ **Tailscale Serve & Funnel Integration**:
  - **Serve**: Encrypted, authenticated access strictly within your private Tailnet (`https://<node-name>.ts.net/<path>`).
  - **Funnel**: Fully public HTTPS route accessible across the Internet without port forwarding or public IPv4.
  - Switch between Serve and Funnel in 1 click!
- 🖥️ **Embedded Web Dashboard (`/router`)**:
  - Responsive dark-mode SPA served directly on port `65534`.
  - Accessible locally at `http://localhost:65534/router` and over Tailnet at `https://<node-name>.ts.net/router`.
- 🔍 **Ascending Port Discovery Scanner**:
  - Automatically discovers open TCP listening ports and sorts them strictly in ascending numerical order (`22`, `53`, `80`, `443`, `8080`...).
  - Displays occupying software name, process PID, and Docker container tags.
- 🔁 **Persistent Route Storage**:
  - Routes are saved in standard location (`config/routes.json`) and persist across reboots and service restarts.
- 🔄 **Boot Autostart**:
  - Run `./install-service.sh` to configure systemd (Linux) or launchd (macOS) autostart on boot.

---

## 🖥️ System Architecture

```text
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
│ TailRouter Python Gateway (Port 65534)                       │
│  - Asyncio reverse proxy & streaming engine                  │
│  - REST API & Web Dashboard                                  │
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

* **Linux / macOS**:
  ```bash
  # Background daemon
  ./start.sh

  # Or foreground with live logs:
  python3 server.py
  ```

* **Windows**:
  ```cmd
  # Background execution
  start.bat

  # Or foreground:
  python server.py
  ```

Open your web browser at: 👉 **`http://localhost:65534/router`**

### 2. Stopping TailRouter

* **Linux / macOS**:
  ```bash
  ./stop.sh
  ```

* **Windows**:
  ```cmd
  stop.bat
  ```

### 3. Autostart on Boot (Linux & macOS)

To install TailRouter as an automatic background service:
```bash
./install-service.sh
```

---

## ⌨️ Script & Command Reference

| Script | Platform | Description |
| :--- | :--- | :--- |
| `python3 server.py` | Cross-platform | Run Gateway in foreground with real-time log output |
| `./start.sh` | Linux / macOS | Launch Gateway as a background daemon process |
| `./stop.sh` | Linux / macOS | Gracefully terminate background Gateway process |
| `start.bat` | Windows | Launch Gateway in background window on Windows |
| `stop.bat` | Windows | Terminate running Gateway process on Windows |
| `./install-service.sh` | Linux / macOS | Install systemd or launchd user service for boot autostart |

---

## 📡 REST API Reference

TailRouter exposes a REST API on port `65534`:

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
├── server.py                 # Core Asyncio Gateway, Reverse Proxy & API Server
├── port_scanner.py           # Host TCP port discovery scanner with process & PID resolution
├── routes_manager.py         # JSON route manager & thread-safe persistence
├── tailscale_helper.py       # Tailscale Serve & Funnel CLI integration
├── service_helper.py         # Platform autostart manager (systemd, launchd, registry)
├── config/
│   └── routes.example.json   # Template route configuration
├── web/
│   └── index.html            # Responsive SPA Web Dashboard with i18n
├── start.sh                  # Background daemon launcher (Linux/macOS)
├── stop.sh                   # Process terminator (Linux/macOS)
├── start.bat                 # Background launcher (Windows)
├── stop.bat                  # Process terminator (Windows)
├── install-service.sh        # System service installer (Linux/macOS)
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
