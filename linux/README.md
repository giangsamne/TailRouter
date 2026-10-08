<div align="center">

# 🐧 TailRouter for Linux

**High-performance Native CLI Engine & Web Control (:65534) for Ubuntu, Debian, Fedora, Arch & Alpine Linux.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg?logo=go&logoColor=white)](https://golang.org)
[![Linux](https://img.shields.io/badge/Linux-Ubuntu%20%7C%20Debian%20%7C%20Alpine%20%7C%20Arch%20%7C%20Fedora-FCC624.svg?logo=linux&logoColor=black)](https://github.com/giangsamne/TailRouter/tree/linux)
[![Arch](https://img.shields.io/badge/arch-x86__64%20%7C%20ARM64-lightgrey.svg)](https://github.com/giangsamne/TailRouter/tree/linux)
[![Init](https://img.shields.io/badge/Init-systemd%20%7C%20OpenRC-blue.svg)](https://github.com/giangsamne/TailRouter/tree/linux)
[![Tailscale](https://img.shields.io/badge/Tailscale-Serve%20%26%20Funnel-5056EC.svg?logo=tailscale&logoColor=white)](https://tailscale.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

---

### ⚡ 1-Line Universal Install for Linux

Run this single command in your Linux terminal (Ubuntu, Debian, Fedora, Arch, Alpine, Raspberry Pi):

```bash
curl -fsSL https://raw.githubusercontent.com/giangsamne/TailRouter/v2.0.1/install.sh | bash
```

* **Zero Dependencies**: Pure static Go binary. Compatible with both `glibc` and `musl` (Alpine Linux).
* **Auto Architecture Detection**: Automatically chooses `x86_64` (AMD64) or `aarch64` (ARM64).
* **System Service**: Automatically sets up background autostart via `systemd` or `openrc`.
* **Direct Web Control**: Immediately ready at: **`http://localhost:65534/router`**.

---

### 📦 Manual Download (Standalone Tarball)

Download the pre-compiled binary package:
* [**`TailRouter-Linux.tar.gz` (GitHub Release)**](https://github.com/giangsamne/TailRouter/releases/download/v2.0.1/TailRouter-Linux.tar.gz)

To extract and run manually:
```bash
tar -xzf TailRouter-Linux.tar.gz
# Run CLI & Web Control
./tailrouter
```

---

### 🖥️ Dual Control Architecture: CLI + Web Control (:65534)

TailRouter on Linux runs as a single lightweight background binary providing both CLI and Web interfaces:

#### 1. 🌐 Web Control Dashboard
* Access locally at: **`http://localhost:65534/router`**
* Or over your Tailnet: **`https://<your-tailscale-node>.ts.net/router`**
* Visually add routes, toggle Serve/Funnel modes, scan ports, and monitor backend latency.

#### 2. ⌨️ Linux CLI Commands
```bash
# Start Gateway in foreground (opens browser by default)
tailrouter

# Inspect Gateway health and Tailscale connection
tailrouter status

# Scan open TCP ports across the Linux host
tailrouter scan

# List all configured routes
tailrouter routes list

# Add a new reverse-proxy route
tailrouter routes add <name> <path> <port> [serve|funnel]
# Example:
tailrouter routes add "Bambu Printer" /bambu 8080 serve

# Delete a route
tailrouter routes delete <id|path>

# Setup / Remove background boot service
tailrouter service install    # Enables systemd / openrc service
tailrouter service status     # Checks service status
tailrouter service uninstall  # Removes system service
```

---

### 📂 Persistent Configuration

Route definitions are reliably saved to:
* `$HOME/.config/tailrouter/routes.json` (or `/etc/tailrouter/routes.json` if running as root)
* Survives updates, binary replacements, and system reboots.

---

### 🌿 Related Branches
* [**`main`**](https://github.com/giangsamne/TailRouter/tree/main): Multi-platform hub & documentation.
* [**`windows`**](https://github.com/giangsamne/TailRouter/tree/windows): Windows native CLI & Web Control.
* [**`macos`**](https://github.com/giangsamne/TailRouter/tree/macos): macOS native CLI & Web Control.
