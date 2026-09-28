<div align="center">

# 🍏 TailRouter for macOS

**High-performance Native CLI Engine & Web Control (:65534) for Apple Silicon & Intel Macs.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![macOS](https://img.shields.io/badge/macOS-Apple%20Silicon%20%7C%20Intel-000000.svg?logo=apple&logoColor=white)](https://github.com/giangsamne/TailRouter/tree/macos)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ 1-Line Universal Install for macOS (Terminal)

Run this single command in Terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/macos/install.sh | bash
```

* **Zero Dependencies**: Pure static Go binary (`tailrouter`), universal for Apple Silicon (M1/M2/M3/M4) and Intel.
* **No Unsigned App Bundles**: Avoids Gatekeeper warnings, quarantine attributes, or Cocoa wrapper complications.
* **Launchd Service**: Automatically configures clean, native macOS launchd agent autostart.
* **Direct Web Control**: Immediately ready at: **`http://localhost:65534/router`**.

---

### 📦 Manual Download (Standalone Zip)

Download the pre-compiled binary package:
* [**`TailRouter-macOS.zip` (GitHub Release)**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.0/TailRouter-macOS.zip)

To install or run manually:
1. Extract `TailRouter-macOS.zip`.
2. Double-click `install.command` (or run `./tailrouter-arm64` / `./tailrouter-amd64`).

---

### 🖥️ Dual Control Architecture: CLI + Web Control (:65534)

TailRouter on macOS runs as a single lightweight background process providing both CLI and Web interfaces:

#### 1. 🌐 Web Control Dashboard
* Access locally at: **`http://localhost:65534/router`**
* Or over your Tailnet: **`https://<your-tailscale-node>.ts.net/router`**
* Visually manage route proxies, toggle Serve/Funnel, scan local ports with a dark-mode web dashboard.

#### 2. ⌨️ macOS CLI Commands (Terminal)
```bash
# Start Gateway in foreground and open browser dashboard
tailrouter

# Inspect Gateway health and Tailscale connection
tailrouter status

# Scan open TCP ports across macOS
tailrouter scan

# List all configured routes
tailrouter routes list

# Add a new reverse-proxy route
tailrouter routes add "Bambu Printer" /bambu 8080 serve

# Delete a route
tailrouter routes delete /bambu

# Setup / Remove launchd background agent
tailrouter service install    # Configures ~/Library/LaunchAgents/com.tailrouter.gateway.plist
tailrouter service status     # Checks launchd agent status
tailrouter service uninstall  # Removes launchd agent
```

---

### 📂 Persistent Configuration

Route definitions are reliably saved to:
* `$HOME/Library/Application Support/TailRouter/routes.json`
* Survives updates, binary replacements, and macOS reboots.

---

### 🌿 Related Branches
* [**`main`**](https://github.com/giangsamne/TailRouter/tree/main): Multi-platform hub & documentation.
* [**`linux`**](https://github.com/giangsamne/TailRouter/tree/linux): Linux native CLI & Web Control.
* [**`windows`**](https://github.com/giangsamne/TailRouter/tree/windows): Windows native CLI & Web Control.
