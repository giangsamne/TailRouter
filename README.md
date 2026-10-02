<div align="center">

# ⚡ TailRouter (Python Edition)

**Lightweight, pure-standard-library bare-metal gateway & web dashboard for Tailscale Serve and Funnel.**

[![Release](https://img.shields.io/badge/Release-v1.1.1-blue.svg?logo=github)](https://github.com/giangsamne/TailRouter/releases)
[![Python](https://img.shields.io/badge/Python-3.8+-3776AB.svg?logo=python&logoColor=white)](https://python.org)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-brightgreen.svg)](https://github.com/giangsamne/TailRouter)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20(Pure%20Standard%20Library)-success.svg)](https://github.com/giangsamne/TailRouter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ 1-Line Universal Auto-Installer (v1.1.1)

Run a single command in your terminal or PowerShell — it installs and starts TailRouter v1.1.1 in one shot:

* **🐧 Linux & 🍏 macOS** (Terminal):
  ```bash
  curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/v1.1.1/install.sh | bash
  ```

* **🪟 Windows** (PowerShell):
  ```powershell
  irm https://raw.githubusercontent.com/giangsamne/TailRouter/v1.1.1/install.ps1 | iex
  ```

---

## 🆕 What's New in v1.1.1

Version **v1.1.1 (Python Edition)** brings essential scan synchronization improvements, full application lifecycle controls, and standardized system settings:

### 1. 🔍 Two-Way Auto-Scan Port Synchronization & Assigned Badge (v1.1.1)
- **Resolved Unassigned Status Bug**: Fixed an issue where ports previously added to the Router continued to display `Unassigned` (`Chưa gán`) with the `+ Add to Router` (`+ Thêm vào Router`) button after page reload.
- **Real-Time Bidirectional Cross-Referencing**: The Auto-Scan engine actively cross-references detected host listening ports with the configured routing table (`routes.json`).
- **Visual Status Badges**: Configured ports immediately display a prominent green **`✓ Assigned` (`✓ Đã gán route`)** status badge alongside active Tailscale route path tags (**`⚡ Route: /ssh`**, **`⚡ Route: /port-3389`**).
- **Safe Action Control**: The action button automatically locks into a disabled **`Added` (`Đã thêm`)** state to prevent duplicate route mappings.
- **Dynamic Reactive Refresh**: The scan table re-renders in real-time when routes are added, deleted, or whenever status polls without requiring a manual rescan.

### 2. ⚙️ 5th Tab: System Settings & Control (Settings Tab)
- **Clutter-Free Main Dashboard**: Relocated all persistent banners (`Tailscale Serve`, `Auto-Start on Boot`, `Operator Warning`) inside the Settings tab.
- **🌐 Remote Access via Tailscale Serve**: Enable/disable `/router` mapped to your Tailscale MagicDNS domain over standard HTTPS (port 443).
- **🚀 Auto-Start on Boot Management**: Configure background daemon autostart on macOS (`launchd`) and Linux (`systemd`, `openrc`, `crontab`).
- **⚠️ Danger Zone & System Reset**:
  - 🗑️ **Reset Router Configuration**: Clear all registered routes back to 0 and reset Tailscale Serve.
  - 🧹 **Clear Proxy Logs**: Purge in-memory live traffic inspection logs.
  - 💥 **Clean App Removal**: Remove configuration files and release port 65534.
- **ℹ️ Project & Repository Info**: Displays engine version, author **Giang3DLab / Giang Sam (MIT License)**, and official GitHub link.

### 3. 🔍 Ascending Numerical Port Order (Auto-Scan)
- Active host TCP listening ports and Docker containers are **automatically sorted in ascending numeric order** (`22 -> 80 -> 443 -> 3000 -> 8080...`).

### 4. 📡 New Management REST APIs
- `POST /api/routes/reset`: Reset all routes to 0.
- `POST /api/logs/clear`: Purge proxy traffic logs.

---

### 🌿 Dedicated OS Branches

TailRouter provides dedicated platform branches containing optimized configurations and raw scripts for each operating system:

| Branch | Platform | Features | Direct Link |
| :--- | :--- | :--- | :--- |
| **`main`** | 🌐 Multi-Platform (Hub) | Go Native Bare-Metal v2.1.1 (< 8MB RAM, 0 dependencies) | [**View `main` branch 👉**](https://github.com/giangsamne/TailRouter/tree/main) |
| **`linux`** | 🐧 Linux | Raw data & Linux scripts, systemd & OpenRC (Alpine/Ubuntu/Arch) | [**View `linux` branch 👉**](https://github.com/giangsamne/TailRouter/tree/linux) |
| **`windows`** | 🪟 Windows | Raw data & Windows scripts, PowerShell installer, Startup autostart | [**View `windows` branch 👉**](https://github.com/giangsamne/TailRouter/tree/windows) |
| **`macos`** | 🍏 macOS | Raw data & macOS scripts, launchd agent, Apple Silicon & Intel | [**View `macos` branch 👉**](https://github.com/giangsamne/TailRouter/tree/macos) |

---

## 💡 What is TailRouter?

**TailRouter** is an ultra-lightweight, zero-external-dependency bare-metal gateway running directly on port **65534** of your host machine. Built purely on the **Python Standard Library** (`asyncio`, `socket`, `urllib`), it eliminates the need for heavy reverse proxies (Nginx, Traefik, NPM) or virtual machines to publish local homelab services to your **Tailnet** or the **public Internet**.

### 🌟 Key Highlights

- ⚡ **Pure Standard Library**:
  - 100% Python standard library. Zero `pip install`, zero external modules.
- 🎛️ **Seamless Tailscale Serve & Funnel Integration**:
  - **Serve**: Encrypted, authenticated access strictly within your private Tailnet (`https://<node-name>.ts.net/<path>`).
  - **Funnel**: Fully public HTTPS route accessible across the Internet without port forwarding or public IPv4.
- 🖥️ **Embedded 5-Tab Web Dashboard**:
  - Responsive dark-mode SPA accessible at `http://localhost:65534/router`.
  - Multi-language support: Tiếng Việt, English, 中文, 日本語.
- 🔍 **Host Port Discovery Scanner**:
  - Scans active TCP listening ports and sorts them in ascending numerical order.

---

## 🚀 Quick Run

```bash
git clone -b v1.1.0 https://github.com/giangsamne/TailRouter.git
cd TailRouter
./start.sh
```

---

## 📜 License

Distributed under the **MIT License**. Author: **Giang3DLab / Giang Sam**. See [LICENSE](LICENSE) for more details.
