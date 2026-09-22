"""
service_helper.py - Quản lý tự khởi động cùng hệ thống đa nền tảng
Hỗ trợ:
- macOS: launchd (LaunchAgent ~/Library/LaunchAgents/com.tailrouter.gateway.plist)
- Linux: systemd user service (~/.config/systemd/user/tailscale-port-router.service + linger)
- Windows: Startup Folder Shortcut (%APPDATA%\\Microsoft\\Windows\\Start Menu\\Programs\\Startup)
"""

import os
import platform
import shutil
import subprocess
import sys
from typing import Any, Dict, Tuple

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
SERVER_PY = os.path.join(BASE_DIR, "server.py")
SERVER_LOG = os.path.join(BASE_DIR, "server.log")


def get_platform_name() -> str:
    sys_plat = platform.system().lower()
    if "darwin" in sys_plat:
        return "macos"
    elif "windows" in sys_plat:
        return "windows"
    return "linux"


def get_macos_plist_path() -> str:
    return os.path.expanduser("~/Library/LaunchAgents/com.tailrouter.gateway.plist")


def get_linux_service_path() -> str:
    return os.path.expanduser("~/.config/systemd/user/tailscale-port-router.service")


def get_autostart_status() -> Dict[str, Any]:
    """Kiểm tra trạng thái tự khởi động trên hệ điều hành hiện tại."""
    plat = get_platform_name()

    if plat == "macos":
        plist = get_macos_plist_path()
        file_exists = os.path.exists(plist)
        is_loaded = False
        pid = None

        if file_exists:
            try:
                res = subprocess.run(["launchctl", "list"], capture_output=True, text=True, timeout=3)
                for line in res.stdout.splitlines():
                    if "com.tailrouter.gateway" in line:
                        is_loaded = True
                        parts = line.split()
                        if len(parts) >= 3 and parts[0] != "-":
                            try:
                                pid = int(parts[0])
                            except ValueError:
                                pass
                        break
            except Exception:
                pass

        return {
            "supported": True,
            "platform": "macos",
            "service_type": "launchd (LaunchAgent)",
            "enabled": file_exists and is_loaded,
            "file_exists": file_exists,
            "is_running": pid is not None,
            "pid": pid,
            "service_file": plist,
            "description": "Tự khởi động cùng macOS qua Apple launchd",
        }

    elif plat == "linux":
        service_file = get_linux_service_path()
        file_exists = os.path.exists(service_file)
        enabled = False
        running = False

        try:
            res_en = subprocess.run(
                ["systemctl", "--user", "is-enabled", "tailscale-port-router.service"],
                capture_output=True,
                text=True,
                timeout=3,
            )
            enabled = res_en.stdout.strip() == "enabled"

            res_st = subprocess.run(
                ["systemctl", "--user", "is-active", "tailscale-port-router.service"],
                capture_output=True,
                text=True,
                timeout=3,
            )
            running = res_st.stdout.strip() == "active"
        except Exception:
            pass

        return {
            "supported": True,
            "platform": "linux",
            "service_type": "systemd user service",
            "enabled": enabled,
            "file_exists": file_exists,
            "is_running": running,
            "service_file": service_file,
            "description": "Tự khởi động cùng Linux qua systemd (kèm Linger bật máy)",
        }

    elif plat == "windows":
        appdata = os.environ.get("APPDATA", "")
        startup_dir = os.path.join(appdata, r"Microsoft\Windows\Start Menu\Programs\Startup")
        vbs_file = os.path.join(startup_dir, "TailRouter.vbs")
        exists = os.path.exists(vbs_file)
        return {
            "supported": True,
            "platform": "windows",
            "service_type": "windows startup",
            "enabled": exists,
            "file_exists": exists,
            "service_file": vbs_file,
            "description": "Tự khởi động cùng Windows Startup Folder",
        }

    return {
        "supported": False,
        "platform": plat,
        "enabled": False,
        "description": "Chưa hỗ trợ tự động trên hệ điều hành này",
    }


def enable_autostart() -> Tuple[bool, str]:
    """Kích hoạt tự khởi động khi máy tính bật nguồn."""
    plat = get_platform_name()
    python_bin = sys.executable or shutil.which("python3") or "/usr/bin/python3"

    if plat == "macos":
        plist_path = get_macos_plist_path()
        os.makedirs(os.path.dirname(plist_path), exist_ok=True)

        plist_content = f"""<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.tailrouter.gateway</string>
    <key>ProgramArguments</key>
    <array>
        <string>{python_bin}</string>
        <string>{SERVER_PY}</string>
    </array>
    <key>WorkingDirectory</key>
    <string>{BASE_DIR}</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{SERVER_LOG}</string>
    <key>StandardErrorPath</key>
    <string>{SERVER_LOG}</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
        <key>SHLVL</key>
        <string>1</string>
    </dict>
</dict>
</plist>
"""
        try:
            with open(plist_path, "w", encoding="utf-8") as f:
                f.write(plist_content.strip() + "\n")

            # Nếu chưa có job com.tailrouter.gateway trong launchctl, thử nạp để launchd quản lý
            try:
                loaded_check = subprocess.run(["launchctl", "list"], capture_output=True, text=True, timeout=3)
                if "com.tailrouter.gateway" not in loaded_check.stdout:
                    subprocess.run(["launchctl", "load", "-w", plist_path], capture_output=True, timeout=5)
            except Exception:
                pass

            return True, "Đã kích hoạt tự khởi động thành công trên macOS (launchd LaunchAgent)!"
        except Exception as e:
            return False, f"Lỗi tạo cấu hình macOS: {e}"

    elif plat == "linux":
        service_file = get_linux_service_path()
        os.makedirs(os.path.dirname(service_file), exist_ok=True)

        service_content = f"""[Unit]
Description=TailRouter & Gateway (Port 65534)
After=network.target tailscaled.service

[Service]
Type=simple
WorkingDirectory={BASE_DIR}
ExecStart={python_bin} {SERVER_PY}
Restart=always
RestartSec=5
StandardOutput=append:{SERVER_LOG}
StandardError=append:{SERVER_LOG}

[Install]
WantedBy=default.target
"""
        try:
            with open(service_file, "w", encoding="utf-8") as f:
                f.write(service_content)

            subprocess.run(["systemctl", "--user", "daemon-reload"], capture_output=True, timeout=5)
            subprocess.run(["systemctl", "--user", "enable", "tailscale-port-router.service"], capture_output=True, timeout=5)
            
            # Enable linger để chạy khi máy bật dù chưa login
            user_name = os.environ.get("USER", "")
            if user_name:
                subprocess.run(["loginctl", "enable-linger", user_name], capture_output=True, timeout=5)

            return True, "Đã kích hoạt tự khởi động thành công trên Linux (systemd + linger)!"
        except Exception as e:
            return False, f"Lỗi cài đặt systemd: {e}"

    elif plat == "windows":
        appdata = os.environ.get("APPDATA", "")
        if not appdata:
            return False, "Không tìm thấy thư mục APPDATA của Windows."
        startup_dir = os.path.join(appdata, r"Microsoft\Windows\Start Menu\Programs\Startup")
        os.makedirs(startup_dir, exist_ok=True)
        vbs_path = os.path.join(startup_dir, "TailRouter.vbs")
        bat_path = os.path.join(BASE_DIR, "start.bat")

        vbs_content = f'''Set WshShell = CreateObject("WScript.Shell")\nWshShell.Run "{bat_path}", 0, False\n'''
        try:
            with open(vbs_path, "w", encoding="utf-8") as f:
                f.write(vbs_content)
            return True, "Đã kích hoạt tự khởi động thành công trên Windows Startup!"
        except Exception as e:
            return False, f"Lỗi tạo file Windows Startup: {e}"

    return False, f"Nền tảng {plat} chưa được hỗ trợ tự động."


def disable_autostart() -> Tuple[bool, str]:
    """Tắt tính năng tự khởi động cùng hệ thống."""
    plat = get_platform_name()

    if plat == "macos":
        plist_path = get_macos_plist_path()
        if os.path.exists(plist_path):
            try:
                os.remove(plist_path)
            except Exception as e:
                return False, f"Lỗi gỡ bỏ file LaunchAgent: {e}"
            return True, "Đã tắt tự khởi động trên macOS (đã gỡ LaunchAgent khỏi thư mục khởi động)."
        return True, "Tự khởi động vốn chưa được bật trên macOS."

    elif plat == "linux":
        service_file = get_linux_service_path()
        subprocess.run(["systemctl", "--user", "disable", "tailscale-port-router.service"], capture_output=True, timeout=5)
        if os.path.exists(service_file):
            try:
                os.remove(service_file)
                subprocess.run(["systemctl", "--user", "daemon-reload"], capture_output=True, timeout=5)
            except Exception:
                pass
        return True, "Đã tắt tự khởi động trên Linux (đã gỡ systemd service)."

    elif plat == "windows":
        appdata = os.environ.get("APPDATA", "")
        if appdata:
            vbs_path = os.path.join(appdata, r"Microsoft\Windows\Start Menu\Programs\Startup", "TailRouter.vbs")
            if os.path.exists(vbs_path):
                try:
                    os.remove(vbs_path)
                except Exception:
                    pass
        return True, "Đã tắt tự khởi động trên Windows."

    return False, f"Nền tảng {plat} chưa được hỗ trợ."
