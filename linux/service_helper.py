"""
service_helper.py - Quản lý tự khởi động cùng hệ thống đa nền tảng
Hỗ trợ:
- macOS: launchd (LaunchAgent ~/Library/LaunchAgents/com.tailrouter.gateway.plist)
- Linux: systemd user service (~/.config/systemd/user/tailscale-port-router.service + linger)
         kèm fallback crontab (@reboot) cho các bản phân phối không dùng systemd (như Alpine Linux / OpenRC)
- Windows: Startup Folder Shortcut (%APPDATA%\\Microsoft\\Windows\\Start Menu\\Programs\\Startup)
"""

import glob
import os
import platform
import shutil
import subprocess
import sys
from typing import Any, Dict, List, Optional, Tuple

BASE_DIR = os.path.dirname(os.path.abspath(__file__))
SERVER_PY = os.path.join(BASE_DIR, "server.py")
SERVER_LOG = os.path.join(BASE_DIR, "server.log")
START_SH = os.path.join(BASE_DIR, "start.sh")


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


def clean_file_lines(file_path: str, keywords: List[str]) -> bool:
    """Xóa các dòng chứa từ khóa khỏi file văn bản."""
    if not os.path.exists(file_path):
        return False
    try:
        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
            lines = f.readlines()
        new_lines = [l for l in lines if not any(kw in l for kw in keywords)]
        if len(new_lines) != len(lines):
            with open(file_path, "w", encoding="utf-8") as f:
                f.writelines(new_lines)
            return True
    except Exception:
        pass
    return False


def update_crontab(lines: List[str]) -> bool:
    """Cập nhật crontab người dùng hiện tại."""
    cleaned = "\n".join(lines).strip()
    if not cleaned:
        res = subprocess.run(["crontab", "-r"], capture_output=True)
        return res.returncode == 0
    res = subprocess.run(["crontab", "-"], input=cleaned + "\n", text=True, capture_output=True)
    return res.returncode == 0


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
            "service_type": "launchd",
            "method": "launchd",
            "enabled": file_exists and is_loaded,
            "file_exists": file_exists,
            "is_running": pid is not None,
            "pid": pid,
            "service_file": plist,
            "description": "Tự khởi động cùng macOS qua Apple launchd",
        }

    elif plat == "linux":
        home = os.path.expanduser("~")
        service_file = get_linux_service_path()
        has_systemd = shutil.which("systemctl") is not None
        enabled = False
        running = False
        service_type = "systemd user service" if has_systemd else "crontab"
        method = "systemd" if has_systemd else "crontab"
        desc = "Tự khởi động cùng Linux qua systemd (kèm Linger bật máy)" if has_systemd else "Tự khởi động cùng Linux qua crontab (@reboot)"

        # 1. Kiểm tra systemd nếu có
        if has_systemd and os.path.exists(service_file):
            try:
                res_en = subprocess.run(
                    ["systemctl", "--user", "is-enabled", "tailscale-port-router.service"],
                    capture_output=True,
                    text=True,
                    timeout=3,
                )
                if res_en.stdout.strip() == "enabled":
                    enabled = True
                    method = "systemd"
                    service_type = "systemd user service"
                    desc = "Tự khởi động cùng Linux qua systemd (kèm Linger bật máy)"

                res_st = subprocess.run(
                    ["systemctl", "--user", "is-active", "tailscale-port-router.service"],
                    capture_output=True,
                    text=True,
                    timeout=3,
                )
                running = res_st.stdout.strip() == "active"
            except Exception:
                pass

        # 2. Kiểm tra OpenRC nếu có file /etc/init.d/tailrouter
        if not enabled and os.path.exists("/etc/init.d/tailrouter"):
            enabled = True
            running = True
            method = "openrc"
            service_type = "openrc"
            service_file = "/etc/init.d/tailrouter"
            desc = "Tự khởi động cùng Linux qua OpenRC service"

        # 3. Kiểm tra Crontab (@reboot)
        if not enabled:
            try:
                c_res = subprocess.run(["crontab", "-l"], capture_output=True, text=True, timeout=3)
                if c_res.returncode == 0:
                    c_text = c_res.stdout
                    if "server.py" in c_text or "tailrouter" in c_text or "start.sh" in c_text:
                        enabled = True
                        running = True
                        method = "crontab"
                        service_type = "crontab"
                        service_file = "crontab (@reboot)"
                        desc = "Tự khởi động cùng Linux qua crontab (@reboot)"
            except Exception:
                pass

        # Nếu là root, kiểm tra thêm trong /var/spool/cron/crontabs/
        if not enabled and os.geteuid() == 0:
            for ct in glob.glob("/var/spool/cron/crontabs/*"):
                try:
                    with open(ct, "r", encoding="utf-8", errors="ignore") as f:
                        ct_c = f.read()
                    if "server.py" in ct_c or "tailrouter" in ct_c or "start.sh" in ct_c:
                        enabled = True
                        running = True
                        method = "crontab"
                        service_type = "crontab"
                        service_file = f"{ct} (@reboot)"
                        desc = "Tự khởi động cùng Linux qua crontab (@reboot)"
                        break
                except Exception:
                    pass

        # 4. Kiểm tra shell profile
        if not enabled:
            check_profiles = []
            for pf in [".profile", ".ashrc", ".bashrc", ".bash_profile"]:
                check_profiles.append(os.path.join(home, pf))
            if os.geteuid() == 0:
                for uh in glob.glob("/home/*"):
                    for pf in [".profile", ".ashrc", ".bashrc", ".bash_profile"]:
                        check_profiles.append(os.path.join(uh, pf))
            for p in check_profiles:
                if os.path.exists(p):
                    try:
                        with open(p, "r", encoding="utf-8", errors="ignore") as f:
                            p_txt = f.read()
                        if "server.py" in p_txt or "tailrouter" in p_txt:
                            enabled = True
                            running = True
                            method = "profile"
                            service_type = "profile"
                            service_file = p
                            desc = f"Tự khởi động qua profile ({os.path.basename(p)})"
                            break
                    except Exception:
                        pass

        return {
            "supported": True,
            "platform": "linux",
            "service_type": service_type,
            "method": method,
            "enabled": enabled,
            "file_exists": os.path.exists(service_file) if isinstance(service_file, str) and "/" in service_file else enabled,
            "is_running": running or enabled,
            "service_file": service_file,
            "description": desc,
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
            "method": "startup_folder",
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
        has_systemd = shutil.which("systemctl") is not None
        home = os.path.expanduser("~")

        # Trường hợp 1: Linux có systemd
        if has_systemd:
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
                subprocess.run(["systemctl", "--user", "start", "tailscale-port-router.service"], capture_output=True, timeout=5)
                
                # Enable linger để chạy khi máy bật dù chưa login
                user_name = os.environ.get("USER", "")
                if user_name:
                    subprocess.run(["loginctl", "enable-linger", user_name], capture_output=True, timeout=5)

                return True, "Đã kích hoạt tự khởi động thành công trên Linux qua systemd (+ linger)!"
            except Exception as e:
                return False, f"Lỗi cài đặt systemd: {e}"

        # Trường hợp 2: Fallback crontab (@reboot) cho Alpine Linux OpenRC / BusyBox
        has_crontab = shutil.which("crontab") is not None
        if has_crontab:
            cron_cmd = f"@reboot sleep 5 && /bin/sh {START_SH} > {SERVER_LOG} 2>&1" if os.path.exists(START_SH) else f"@reboot sleep 5 && {python_bin} {SERVER_PY} > {SERVER_LOG} 2>&1"
            
            c_res = subprocess.run(["crontab", "-l"], capture_output=True, text=True, timeout=3)
            current_cron = c_res.stdout if c_res.returncode == 0 else ""
            lines = [l for l in current_cron.splitlines() if l.strip()]

            # Kiểm tra xem đã có entry chưa
            already_exists = any("server.py" in l or "tailrouter" in l or "start.sh" in l for l in lines)
            if not already_exists:
                lines.append(cron_cmd)
                if update_crontab(lines):
                    return True, "Đã kích hoạt tự khởi động thành công trên Linux qua crontab (@reboot)!"
                return False, "Không thể cập nhật cấu hình crontab."
            return True, "Tự khởi động qua crontab (@reboot) đã được cấu hình từ trước."

        return False, "Không tìm thấy systemctl hoặc crontab trên hệ thống để cài đặt tự khởi động."

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
    home = os.path.expanduser("~")

    if plat == "macos":
        plist_path = get_macos_plist_path()
        if os.path.exists(plist_path):
            try:
                subprocess.run(["launchctl", "unload", "-w", plist_path], capture_output=True, timeout=5)
                os.remove(plist_path)
            except Exception as e:
                return False, f"Lỗi gỡ bỏ file LaunchAgent: {e}"
            return True, "Đã tắt tự khởi động trên macOS."
        return True, "Tự khởi động vốn chưa được bật trên macOS."

    elif plat == "linux":
        # 1. Gỡ systemd nếu có
        if shutil.which("systemctl"):
            subprocess.run(["systemctl", "--user", "stop", "tailscale-port-router.service"], capture_output=True, timeout=5)
            subprocess.run(["systemctl", "--user", "disable", "tailscale-port-router.service"], capture_output=True, timeout=5)
            service_file = get_linux_service_path()
            if os.path.exists(service_file):
                try:
                    os.remove(service_file)
                    subprocess.run(["systemctl", "--user", "daemon-reload"], capture_output=True, timeout=5)
                except Exception:
                    pass
            if os.geteuid() == 0:
                subprocess.run(["systemctl", "stop", "tailscale-port-router.service"], capture_output=True, timeout=5)
                subprocess.run(["systemctl", "disable", "tailscale-port-router.service"], capture_output=True, timeout=5)

        # 2. Gỡ OpenRC service nếu có
        if os.path.exists("/etc/init.d/tailrouter"):
            try:
                subprocess.run(["rc-service", "tailrouter", "stop"], capture_output=True, timeout=5)
                subprocess.run(["rc-update", "del", "tailrouter", "default"], capture_output=True, timeout=5)
                os.remove("/etc/init.d/tailrouter")
            except Exception:
                pass

        # 3. Gỡ khỏi crontab
        if shutil.which("crontab"):
            try:
                c_res = subprocess.run(["crontab", "-l"], capture_output=True, text=True, timeout=3)
                if c_res.returncode == 0:
                    remaining = []
                    for l in c_res.stdout.splitlines():
                        if not any(kw in l for kw in ["server.py", "tailrouter", "start.sh", "tailscale-port-router"]):
                            remaining.append(l)
                    update_crontab(remaining)
            except Exception:
                pass

        # Nếu là root, dọn dẹp crontabs của mọi user
        if os.geteuid() == 0:
            for ct in glob.glob("/var/spool/cron/crontabs/*") + glob.glob("/var/spool/cron/*"):
                if not ct.endswith("crontabs"):
                    clean_file_lines(ct, ["server.py", "tailrouter", "start.sh", "tailscale-port-router"])

        # 4. Dọn dẹp shell profiles để tránh tự khởi động lại khi login
        clean_profiles = []
        for pf in [".profile", ".ashrc", ".bashrc", ".bash_profile", ".zshrc"]:
            clean_profiles.append(os.path.join(home, pf))
        if os.geteuid() == 0:
            for uh in glob.glob("/home/*"):
                for pf in [".profile", ".ashrc", ".bashrc", ".bash_profile", ".zshrc"]:
                    clean_profiles.append(os.path.join(uh, pf))
            for pf in [".profile", ".ashrc", ".bashrc", ".bash_profile", ".zshrc"]:
                clean_profiles.append(os.path.join("/root", pf))

        for p in clean_profiles:
            clean_file_lines(p, ["server.py", "tailrouter", "tailscale-port-router"])

        return True, "Đã tắt tự khởi động thành công trên Linux."

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


get_service_status = get_autostart_status


if __name__ == "__main__":
    action = sys.argv[1].lower() if len(sys.argv) > 1 else "status"
    if action in ["enable", "on", "install"]:
        ok, msg = enable_autostart()
        print(f"[{'OK' if ok else 'FAIL'}] {msg}")
    elif action in ["disable", "off", "uninstall"]:
        ok, msg = disable_autostart()
        print(f"[{'OK' if ok else 'FAIL'}] {msg}")
    else:
        import json
        st = get_autostart_status()
        print(json.dumps(st, indent=2, ensure_ascii=False))
