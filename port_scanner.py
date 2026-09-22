"""
port_scanner.py - Quét tự động các cổng TCP và dịch vụ đang lắng nghe trên máy chủ
Hỗ trợ nhận diện Docker containers, tên tiến trình, và gửi HTTP probe để lấy tiêu đề trang.
"""

import asyncio
import re
import socket
import subprocess
from typing import Any, Dict, List, Optional


def get_docker_services() -> Dict[int, Dict[str, str]]:
    """Lấy danh sách các cổng được map bởi Docker containers."""
    docker_map = {}
    try:
        res = subprocess.run(
            ["docker", "ps", "--format", "{{.Names}}\t{{.Ports}}\t{{.Image}}"],
            capture_output=True,
            text=True,
            timeout=3,
        )
        if res.returncode == 0:
            for line in res.stdout.strip().split("\n"):
                if not line.strip():
                    continue
                parts = line.split("\t")
                if len(parts) >= 2:
                    name = parts[0]
                    ports_str = parts[1]
                    image = parts[2] if len(parts) > 2 else ""
                    matches = re.findall(r"(?:0\.0\.0\.0|:::|127\.0\.0\.1):(\d+)->", ports_str)
                    for p in matches:
                        port_num = int(p)
                        docker_map[port_num] = {
                            "container_name": name,
                            "image": image,
                            "type": "docker",
                        }
    except Exception:
        pass
    return docker_map


def parse_ss_listeners() -> List[Dict[str, Any]]:
    """Phân tích lệnh ss -tulpn để tìm các cổng TCP LISTEN trên Linux."""
    listeners = []
    try:
        res = subprocess.run(
            ["ss", "-tulpn"],
            capture_output=True,
            text=True,
            timeout=3,
        )
        if res.returncode == 0:
            for line in res.stdout.split("\n"):
                if "LISTEN" not in line or "tcp" not in line.lower():
                    continue
                parts = line.split()
                if len(parts) < 5:
                    continue
                local_addr = parts[4]
                last_colon = local_addr.rfind(":")
                if last_colon == -1:
                    continue
                ip_part = local_addr[:last_colon]
                port_part = local_addr[last_colon + 1 :]
                try:
                    port = int(port_part)
                except ValueError:
                    continue

                process = "N/A"
                if len(parts) >= 7:
                    user_info = " ".join(parts[6:])
                    proc_match = re.search(r'users:\(\("([^"]+)"', user_info)
                    if proc_match:
                        process = proc_match.group(1)

                listeners.append({"ip": ip_part, "port": port, "process": process})
    except Exception:
        pass
    return listeners


def parse_macos_listeners() -> List[Dict[str, Any]]:
    """Phân tích tiến trình và cổng lắng nghe trên macOS qua lsof kết hợp netstat."""
    listeners = []
    seen_ports = set()

    # 1. Thử lsof để lấy thông tin chi tiết tên tiến trình
    try:
        res = subprocess.run(
            ["lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n"],
            capture_output=True,
            text=True,
            timeout=3,
        )
        if res.returncode == 0:
            for line in res.stdout.split("\n")[1:]:
                parts = line.split()
                if len(parts) >= 9:
                    proc_name = parts[0]
                    addr_part = parts[8]  # *:8080 or 127.0.0.1:3000
                    if ":" in addr_part:
                        ip_p, port_p = addr_part.rsplit(":", 1)
                        try:
                            port = int(port_p)
                            listeners.append({"ip": ip_p, "port": port, "process": proc_name})
                            seen_ports.add(port)
                        except ValueError:
                            pass
    except Exception:
        pass

    # 2. Bổ sung netstat -an -p tcp để phát hiện mọi cổng LISTEN hệ thống kể cả khi không chạy dưới sudo
    try:
        res = subprocess.run(
            ["netstat", "-an", "-p", "tcp"],
            capture_output=True,
            text=True,
            timeout=3,
        )
        if res.returncode == 0:
            for line in res.stdout.splitlines():
                if "LISTEN" in line:
                    parts = line.split()
                    if len(parts) >= 4:
                        local_addr = parts[3]
                        last_dot = local_addr.rfind(".")
                        if last_dot != -1:
                            port_str = local_addr[last_dot + 1 :]
                            try:
                                port = int(port_str)
                                if port not in seen_ports and port != 65534:
                                    listeners.append({"ip": "0.0.0.0", "port": port, "process": "Service"})
                                    seen_ports.add(port)
                            except ValueError:
                                pass
    except Exception:
        pass

    return listeners


def parse_windows_listeners() -> List[Dict[str, Any]]:
    """Phân tích cổng lắng nghe trên Windows qua netstat."""
    listeners = []
    try:
        res = subprocess.run(
            ["netstat", "-ano", "-p", "tcp"],
            capture_output=True,
            text=True,
            timeout=3,
        )
        if res.returncode == 0:
            for line in res.stdout.split("\n"):
                if "LISTENING" in line:
                    parts = line.split()
                    if len(parts) >= 5:
                        addr = parts[1]  # 0.0.0.0:8080
                        pid = parts[4]
                        if ":" in addr:
                            ip_p, port_p = addr.rsplit(":", 1)
                            try:
                                port = int(port_p)
                                listeners.append({"ip": ip_p, "port": port, "process": f"PID:{pid}"})
                            except ValueError:
                                pass
    except Exception:
        pass
    return listeners


def parse_listeners_cross_platform() -> List[Dict[str, Any]]:
    """Tự động phân phối phương thức quét cổng theo hệ điều hành hiện tại."""
    import platform
    sys_plat = platform.system().lower()

    if "darwin" in sys_plat:
        listeners = parse_macos_listeners()
    elif "windows" in sys_plat:
        listeners = parse_windows_listeners()
    else:
        listeners = parse_ss_listeners()

    # Nếu không tìm được kết quả, thử dự phòng các phương thức khác
    if not listeners and "linux" in sys_plat:
        try:
            res = subprocess.run(["netstat", "-tulpn"], capture_output=True, text=True, timeout=2)
            # fallback parse netstat if needed
        except Exception:
            pass

    return listeners


async def probe_http_service(port: int, host: str = "127.0.0.1", timeout: float = 0.5) -> Dict[str, Any]:
    """Kiểm tra nhanh xem cổng có phục vụ giao thức HTTP không và lấy thông tin cơ bản."""
    result = {"is_http": False, "title": None, "server": None, "latency_ms": None}
    loop = asyncio.get_event_loop()
    start_t = loop.time()

    try:
        reader, writer = await asyncio.wait_for(
            asyncio.open_connection(host, port),
            timeout=timeout,
        )
        req = (
            f"GET / HTTP/1.1\r\n"
            f"Host: {host}:{port}\r\n"
            f"User-Agent: TailscalePortRouter/1.0\r\n"
            f"Accept: */*\r\n"
            f"Connection: close\r\n\r\n"
        )
        writer.write(req.encode())
        await writer.drain()

        raw_response = await asyncio.wait_for(reader.read(4096), timeout=timeout)
        latency = round((loop.time() - start_t) * 1000, 1)
        result["latency_ms"] = latency

        try:
            writer.close()
            await writer.wait_closed()
        except Exception:
            pass

        if raw_response.startswith(b"HTTP/1.") or raw_response.startswith(b"HTTP/2."):
            result["is_http"] = True
            text = raw_response.decode("utf-8", errors="ignore")
            srv_match = re.search(r"server:\s*([^\r\n]+)", text, re.IGNORECASE)
            if srv_match:
                result["server"] = srv_match.group(1).strip()
            title_match = re.search(r"<title[^>]*>([^<]+)</title>", text, re.IGNORECASE)
            if title_match:
                result["title"] = title_match.group(1).strip()
    except Exception:
        pass

    return result


async def scan_active_ports(configured_ports: Optional[List[int]] = None) -> List[Dict[str, Any]]:
    """Quét toàn bộ cổng đang mở trên máy chủ và bổ sung thông tin chi tiết."""
    if configured_ports is None:
        configured_ports = []

    docker_services = get_docker_services()
    raw_listeners = parse_listeners_cross_platform()

    ports_seen = {}
    for item in raw_listeners:
        p = item["port"]
        if p == 65534:  # Không tự quét chính gateway
            continue
        if p not in ports_seen:
            ports_seen[p] = item
        else:
            if ports_seen[p]["process"] == "N/A" and item["process"] != "N/A":
                ports_seen[p]["process"] = item["process"]

    for dp in docker_services:
        if dp not in ports_seen and dp != 65534:
            ports_seen[dp] = {"ip": "0.0.0.0", "port": dp, "process": "docker"}

    tasks = []
    port_list = sorted(ports_seen.keys())
    for p in port_list:
        tasks.append(probe_http_service(p))

    probe_results = await asyncio.gather(*tasks, return_exceptions=True)

    results = []
    for p, probe in zip(port_list, probe_results):
        info = ports_seen[p]
        docker_info = docker_services.get(p)
        
        probe_data = probe if isinstance(probe, dict) else {"is_http": False, "title": None, "server": None, "latency_ms": None}

        service_name = info["process"]
        if docker_info:
            service_name = f"Docker: {docker_info['container_name']}"
        elif p == 80:
            service_name = "HTTP Web"
        elif p == 443:
            service_name = "HTTPS Web"
        elif p == 22:
            service_name = "SSH Service"
        elif p == 4000:
            service_name = "NoMachine Remote Desktop"
        elif p == 3389:
            service_name = "Windows/Gnome Remote Desktop"
        elif p == 8080:
            service_name = "HTTP Service (8080)"
        elif p == 3000:
            service_name = "Node/React Web App"
        elif p == 5000:
            service_name = "Flask/Python App"

        suggested_path = f"/port-{p}"
        if docker_info:
            clean_name = re.sub(r"[^a-zA-Z0-9_-]", "", docker_info["container_name"]).lower()
            if clean_name:
                suggested_path = f"/{clean_name}"
        elif probe_data.get("title"):
            title_slug = re.sub(r"[^a-zA-Z0-9_-]", "-", probe_data["title"].strip()).lower()[:15].strip("-")
            if title_slug:
                suggested_path = f"/{title_slug}"

        # Tìm các route Tailscale đang mở trên cổng này
        ts_routes = []
        try:
            from tailscale_helper import get_tailscale_serve_routes
            all_ts_routes = get_tailscale_serve_routes()
            for tr in all_ts_routes:
                if tr.get("target_port") == p:
                    ts_routes.append(tr.get("path"))
        except Exception:
            pass

        is_conf = (p in configured_ports) or bool(ts_routes)

        results.append({
            "port": p,
            "ip": info["ip"],
            "process": info["process"],
            "service_name": service_name,
            "is_docker": bool(docker_info),
            "docker_container": docker_info["container_name"] if docker_info else None,
            "is_http": probe_data.get("is_http", False),
            "title": probe_data.get("title"),
            "server": probe_data.get("server"),
            "latency_ms": probe_data.get("latency_ms"),
            "suggested_path": ts_routes[0] if ts_routes else suggested_path,
            "tailscale_routes": ts_routes,
            "is_configured": is_conf,
        })

    return results


if __name__ == "__main__":
    import json
    data = asyncio.run(scan_active_ports())
    print(json.dumps(data, indent=2, ensure_ascii=False))
