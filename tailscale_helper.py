import json
import logging
import os
import platform
import shutil
import subprocess
from typing import Any, Dict, List, Optional, Tuple

logger = logging.getLogger("tailscale_helper")


def get_tailscale_bin() -> str:
    """Xác định đường dẫn nhị phân tailscale CLI đa nền tảng (Linux, macOS, Windows)."""
    # 1. Kiểm tra biến môi trường PATH trước
    which_bin = shutil.which("tailscale")
    if which_bin:
        return which_bin

    sys_plat = platform.system().lower()

    if "windows" in sys_plat:
        candidates = [
            r"C:\Program Files\Tailscale\tailscale.exe",
            r"C:\Program Files (x86)\Tailscale\tailscale.exe",
            os.path.expandvars(r"%LOCALAPPDATA%\Tailscale\tailscale.exe"),
            "tailscale.exe",
        ]
        for c in candidates:
            if os.path.exists(c):
                return c
        return "tailscale.exe"

    elif "darwin" in sys_plat:  # macOS
        candidates = [
            "/Applications/Tailscale.app/Contents/MacOS/Tailscale",
            "/opt/homebrew/bin/tailscale",
            "/usr/local/bin/tailscale",
            "tailscale",
        ]
        for c in candidates:
            if os.path.exists(c):
                return c
        return "tailscale"

    else:  # Linux / Unix
        candidates = [
            "/snap/bin/tailscale",
            "/usr/bin/tailscale",
            "/usr/local/bin/tailscale",
            "tailscale",
        ]
        for c in candidates:
            if os.path.exists(c):
                return c
        return "tailscale"


def run_tailscale_cmd(args: List[str], timeout: float = 6.0) -> subprocess.CompletedProcess:
    """Chạy lệnh tailscale CLI an toàn."""
    cmd = [get_tailscale_bin()] + args
    env = dict(os.environ)
    if "SHLVL" not in env:
        env["SHLVL"] = "1"
    extra_paths = ["/usr/local/bin", "/opt/homebrew/bin", "/snap/bin", "/usr/bin", "/bin"]
    curr_path = env.get("PATH", "")
    for ep in extra_paths:
        if ep not in curr_path:
            curr_path = f"{curr_path}:{ep}" if curr_path else ep
    env["PATH"] = curr_path

    return subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        timeout=timeout,
        env=env,
    )



def apply_route_tailscale(path: str, target: str, mode: str = "serve") -> Dict[str, Any]:
    """
    Áp dụng một route vào Tailscale:
    - mode == 'funnel': Public ra toàn bộ internet qua HTTPS: https://<fqdn><path>
    - mode == 'serve': Chỉ truy cập trong mạng Tailnet: https://<fqdn><path>
    """
    tailscale_bin = get_tailscale_bin()
    clean_path = "/" + path.strip().lstrip("/").rstrip("/")
    if not clean_path:
        clean_path = "/"

    # Chuẩn hóa target (ví dụ: http://127.0.0.1:8080)
    target_url = target
    if not target_url.startswith("http://") and not target_url.startswith("https://"):
        target_url = f"http://{target_url}"

    cmd = [tailscale_bin, mode, "--bg", "--set-path", clean_path, target_url]
    try:
        res = subprocess.run(cmd, capture_output=True, text=True, timeout=8.0)
        success = res.returncode == 0
        err = (res.stderr or res.stdout).strip()
        return {
            "success": success,
            "mode": mode,
            "path": clean_path,
            "target": target_url,
            "output": res.stdout.strip(),
            "error": err if not success else None,
        }
    except Exception as e:
        return {"success": False, "mode": mode, "path": clean_path, "target": target_url, "error": str(e)}


def remove_route_tailscale(path: str) -> Dict[str, Any]:
    """Xóa một đường dẫn route khỏi Tailscale Serve / Funnel."""
    tailscale_bin = get_tailscale_bin()
    clean_path = "/" + path.strip().lstrip("/").rstrip("/")
    if not clean_path:
        clean_path = "/"

    # Tailscale CLI dùng '--set-path <path> off' để gỡ route
    cmd = [tailscale_bin, "serve", "--set-path", clean_path, "off"]
    try:
        res = subprocess.run(cmd, capture_output=True, text=True, timeout=8.0)
        return {
            "success": res.returncode == 0,
            "path": clean_path,
            "output": res.stdout.strip(),
            "error": res.stderr.strip() if res.returncode != 0 else None,
        }
    except Exception as e:
        return {"success": False, "path": clean_path, "error": str(e)}


def sync_all_routes_tailscale(routes: List[Dict[str, Any]], port: int = 65534) -> Dict[str, Any]:
    """
    Đồng bộ toàn bộ danh sách routes với Tailscale Serve / Funnel:
    1. Đảm bảo /router luôn trỏ về http://127.0.0.1:PORT/router
    2. Duyệt từng route: nếu enabled thì áp dụng (serve hoặc funnel), nếu disabled thì gỡ bỏ.
    """
    results = {}
    
    # 1. Đăng ký /router (luôn ở chế độ serve nội bộ để an toàn)
    router_res = apply_route_tailscale("/router", f"http://127.0.0.1:{port}/router", mode="serve")
    results["/router"] = router_res

    # 2. Duyệt các routes người dùng
    for r in routes:
        path = r.get("path")
        if not path or path == "/router":
            continue

        target = f"http://{r.get('target_host', '127.0.0.1')}:{r.get('target_port', 80)}"
        mode = r.get("mode", "serve")
        enabled = r.get("enabled", True)

        if enabled:
            res = apply_route_tailscale(path, target, mode=mode)
            results[path] = res
        else:
            res = remove_route_tailscale(path)
            results[path] = res

    return results


def get_tailscale_info() -> Dict[str, Any]:
    """Lấy thông tin chi tiết về node Tailscale và cấu hình Serve/Funnel hiện tại."""
    info = {
        "installed": False,
        "running": False,
        "backend_state": "Offline",
        "ipv4": None,
        "ipv6": None,
        "node_name": None,
        "fqdn": None,
        "base_https_url": None,
        "remote_router_url": None,
        "peers_count": 0,
        "serve_configured": False,
        "funnel_active": False,
        "serve_routes": {},
    }

    tailscale_bin = get_tailscale_bin()
    try:
        ver_res = subprocess.run([tailscale_bin, "version"], capture_output=True, text=True, timeout=2)
        if ver_res.returncode == 0:
            info["installed"] = True
    except Exception:
        return info

    try:
        status_res = subprocess.run([tailscale_bin, "status", "--json"], capture_output=True, text=True, timeout=3)
        if status_res.returncode == 0:
            data = json.loads(status_res.stdout)
            backend_state = data.get("BackendState", "Unknown")
            info["backend_state"] = backend_state
            info["running"] = backend_state == "Running"

            self_node = data.get("Self", {})
            tailscale_ips = self_node.get("TailscaleIPs", [])
            for ip in tailscale_ips:
                if ":" in ip:
                    info["ipv6"] = ip
                else:
                    info["ipv4"] = ip

            dns_name = self_node.get("DNSName", "").rstrip(".")
            info["fqdn"] = dns_name
            info["node_name"] = self_node.get("HostName", "")

            if dns_name:
                info["base_https_url"] = f"https://{dns_name}"
                info["remote_router_url"] = f"https://{dns_name}/router"

            peers = data.get("Peer", {})
            info["peers_count"] = len(peers)
    except Exception as e:
        info["backend_state"] = f"Error: {e}"

    # Kiểm tra trạng thái Tailscale Serve / Funnel qua CLI
    try:
        serve_res = subprocess.run([tailscale_bin, "serve", "status", "--json"], capture_output=True, text=True, timeout=3)
        if serve_res.returncode == 0:
            output = serve_res.stdout.strip()
            if output and output != "{}":
                info["serve_configured"] = True
                try:
                    data = json.loads(output)
                    web_config = data.get("Web", {})
                    for host_port, cfg in web_config.items():
                        handlers = cfg.get("Handlers", {})
                        for h_path, h_val in handlers.items():
                            info["serve_routes"][h_path] = h_val
                except Exception:
                    pass
    except Exception:
        pass

    # Kiểm tra Funnel status
    try:
        fn_res = subprocess.run([tailscale_bin, "funnel", "status"], capture_output=True, text=True, timeout=3)
        if fn_res.returncode == 0:
            fn_out = fn_res.stdout
            if "Funnel on" in fn_out or "Available on the internet" in fn_out:
                info["funnel_active"] = True
    except Exception:
        pass

    return info


def get_tailscale_serve_routes() -> List[Dict[str, Any]]:
    """
    Quét trực tiếp từ 'tailscale serve status --json' để lấy toàn bộ các route
    đang mở bằng lệnh CLI (serve hoặc funnel).
    """
    import urllib.parse
    tailscale_bin = get_tailscale_bin()
    try:
        res = subprocess.run([tailscale_bin, "serve", "status", "--json"], capture_output=True, text=True, timeout=4.0)
        if res.returncode != 0 or not res.stdout.strip():
            return []

        data = json.loads(res.stdout)
        web = data.get("Web", {})
        allow_funnel = data.get("AllowFunnel", {})

        routes = []
        for host_port, cfg in web.items():
            is_funnel = bool(allow_funnel.get(host_port, False))
            handlers = cfg.get("Handlers", {})
            for path, h in handlers.items():
                if path == "/router":
                    continue
                proxy_url = h.get("Proxy", "")
                if proxy_url:
                    parsed = urllib.parse.urlparse(proxy_url)
                    host = parsed.hostname or "127.0.0.1"
                    port = parsed.port or (443 if parsed.scheme == "https" else 80)
                    routes.append({
                        "path": path,
                        "target_host": host,
                        "target_port": port,
                        "mode": "funnel" if is_funnel else "serve",
                        "raw_proxy": proxy_url,
                    })
        return routes
    except Exception as e:
        logger.warning(f"Lỗi quét route từ Tailscale CLI: {e}")
        return []


if __name__ == "__main__":
    print(json.dumps(get_tailscale_info(), indent=2, ensure_ascii=False))
    print("Discovered routes:", json.dumps(get_tailscale_serve_routes(), indent=2))

