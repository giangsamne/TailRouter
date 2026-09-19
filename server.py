"""
server.py - TailRouter & Gateway (Port 65534)
Hệ thống Gateway điều phối Reverse-Proxy và Dashboard quản lý trên cổng 65534
"""

import asyncio
import json
import logging
import os
import re
import socket
import sys
import time
import urllib.parse
from typing import Any, Dict, List, Optional, Tuple

from port_scanner import scan_active_ports
from routes_manager import RoutesManager
from tailscale_helper import (
    apply_route_tailscale,
    get_tailscale_info,
    remove_route_tailscale,
    run_tailscale_cmd,
    sync_all_routes_tailscale,
)

# Cấu hình logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%H:%M:%S",
)
logger = logging.getLogger("tailscale_router")

PORT = 65534
HOST = "0.0.0.0"
BASE_DIR = os.path.dirname(os.path.abspath(__file__))
WEB_DIR = os.path.join(BASE_DIR, "web")
INDEX_HTML = os.path.join(WEB_DIR, "index.html")

START_TIME = time.time()
routes_mgr = RoutesManager()
recent_logs: List[Dict[str, Any]] = []
MAX_LOGS = 100


def add_log(method: str, path: str, status_code: int, target: str, latency_ms: float = 0.0) -> None:
    """Ghi nhận nhật ký truy cập ngắn gọn."""
    entry = {
        "time": time.strftime("%H:%M:%S"),
        "method": method,
        "path": path,
        "status": status_code,
        "target": target,
        "latency_ms": round(latency_ms, 1),
    }
    recent_logs.append(entry)
    if len(recent_logs) > MAX_LOGS:
        recent_logs.pop(0)


async def pipe_stream(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    """Chuyển tiếp luồng nhị phân trực tiếp hai chiều (dành cho WebSocket & streaming)."""
    try:
        while True:
            chunk = await reader.read(65536)
            if not chunk:
                break
            writer.write(chunk)
            await writer.drain()
    except Exception:
        pass
    finally:
        try:
            writer.close()
        except Exception:
            pass


class HTTPRequest:
    def __init__(self, method: str, full_path: str, version: str, headers: Dict[str, str], body: bytes):
        self.method = method
        self.full_path = full_path
        self.version = version
        self.headers = headers
        self.body = body

        parsed = urllib.parse.urlparse(full_path)
        self.path = parsed.path
        self.query = parsed.query

    def get_header(self, name: str, default: str = "") -> str:
        for k, v in self.headers.items():
            if k.lower() == name.lower():
                return v
        return default

    def is_websocket(self) -> bool:
        upgrade = self.get_header("upgrade").lower()
        conn = self.get_header("connection").lower()
        return "websocket" in upgrade or "upgrade" in conn


async def parse_http_request(reader: asyncio.StreamReader) -> Optional[HTTPRequest]:
    """Đọc và phân tích dòng yêu cầu và headers của HTTP request."""
    try:
        line = await reader.readline()
        if not line:
            return None
        line_str = line.decode("utf-8", errors="ignore").strip()
        parts = line_str.split(" ")
        if len(parts) < 3:
            return None

        method, full_path, version = parts[0], parts[1], parts[2]
        headers = {}

        while True:
            header_line = await reader.readline()
            if not header_line or header_line == b"\r\n" or header_line == b"\n":
                break
            h_str = header_line.decode("utf-8", errors="ignore").strip()
            if ":" in h_str:
                k, v = h_str.split(":", 1)
                headers[k.strip()] = v.strip()

        # Đọc body nếu có Content-Length
        body = b""
        content_length_str = headers.get("Content-Length") or headers.get("content-length")
        if content_length_str:
            try:
                cl = int(content_length_str)
                if cl > 0 and cl <= 10 * 1024 * 1024:  # Đọc tối đa 10MB cho API requests
                    body = await reader.readexactly(cl)
            except Exception:
                pass

        return HTTPRequest(method, full_path, version, headers, body)
    except Exception as e:
        logger.error(f"Lỗi parse HTTP request: {e}")
        return None


async def send_json_response(writer: asyncio.StreamWriter, status: int, data: Any) -> None:
    """Gửi phản hồi JSON chuẩn."""
    payload = json.dumps(data, ensure_ascii=False).encode("utf-8")
    headers = [
        f"HTTP/1.1 {status} {'OK' if status == 200 else 'Error'}",
        "Content-Type: application/json; charset=utf-8",
        f"Content-Length: {len(payload)}",
        "Access-Control-Allow-Origin: *",
        "Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS",
        "Access-Control-Allow-Headers: Content-Type, Authorization",
        "Connection: close",
        "\r\n",
    ]
    writer.write("\r\n".join(headers).encode("utf-8") + payload)
    await writer.drain()


async def send_html_response(writer: asyncio.StreamWriter, html_content: str, status: int = 200) -> None:
    """Gửi phản hồi HTML."""
    payload = html_content.encode("utf-8")
    headers = [
        f"HTTP/1.1 {status} OK",
        "Content-Type: text/html; charset=utf-8",
        f"Content-Length: {len(payload)}",
        "Connection: close",
        "\r\n",
    ]
    writer.write("\r\n".join(headers).encode("utf-8") + payload)
    await writer.drain()


async def handle_api_request(req: HTTPRequest, writer: asyncio.StreamWriter, client_ip: str) -> bool:
    """Xử lý các endpoint REST API cho giao diện quản trị."""
    path = req.path
    method = req.method

    # Xử lý CORS Preflight
    if method == "OPTIONS":
        headers = [
            "HTTP/1.1 204 No Content",
            "Access-Control-Allow-Origin: *",
            "Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS",
            "Access-Control-Allow-Headers: Content-Type, Authorization",
            "Connection: close",
            "\r\n\r\n",
        ]
        writer.write("\r\n".join(headers).encode("utf-8"))
        await writer.drain()
        return True

    # 1. API Trạng thái tổng quan
    if path == "/api/status" and method == "GET":
        ts_info = get_tailscale_info()
        routes = routes_mgr.list_routes()
        uptime_sec = int(time.time() - START_TIME)

        total_hits = sum(r.get("hits", 0) for r in routes)
        active_cnt = sum(1 for r in routes if r.get("enabled", True))
        data = {
            "version": "1.0.0",
            "uptime_seconds": uptime_sec,
            "port": PORT,
            "tailscale": ts_info,
            "total_routes": len(routes),
            "active_routes": active_cnt,
            "stats": {
                "total_routes": len(routes),
                "active_routes": active_cnt,
                "total_hits": total_hits,
                "uptime_seconds": uptime_sec,
            },
            "routes": routes,
            "logs": recent_logs[-20:],
            "recent_logs": recent_logs[-20:],
        }
        await send_json_response(writer, 200, data)
        return True

    # 2. API Quét cổng tự động (Auto Scan)
    if path == "/api/scan" and method == "GET":
        configured_ports = routes_mgr.get_configured_ports()
        scan_results = await scan_active_ports(configured_ports)
        await send_json_response(writer, 200, {"ports": scan_results, "count": len(scan_results)})
        return True

    # 3. API Quản lý Routes (Lấy danh sách / Thêm mới)
    if path == "/api/routes":
        if method == "GET":
            await routes_mgr.check_all_health()
            await send_json_response(writer, 200, {"routes": routes_mgr.list_routes()})
            return True
        elif method == "POST":
            try:
                body_json = json.loads(req.body.decode("utf-8")) if req.body else {}
                success, msg, new_route = routes_mgr.add_route(body_json)
                if success:
                    # Đồng bộ ngay vào Tailscale Serve / Funnel nếu route được kích hoạt
                    if new_route.get("enabled", True):
                        target = f"http://{new_route['target_host']}:{new_route['target_port']}"
                        apply_route_tailscale(new_route["path"], target, mode=new_route.get("mode", "serve"))
                    await send_json_response(writer, 201, {"success": True, "message": msg, "route": new_route})
                else:
                    await send_json_response(writer, 400, {"success": False, "message": msg})
            except Exception as e:
                await send_json_response(writer, 400, {"success": False, "message": f"Dữ liệu không hợp lệ: {e}"})
            return True

    # 4. API Thao tác trên từng Route cụ thể (/api/routes/<id>...)
    route_match = re.match(r"^/api/routes/([^/]+)(?:/(toggle|toggle-mode|ping))?$", path)
    if route_match:
        route_id = route_match.group(1)
        sub_action = route_match.group(2)

        if sub_action == "toggle" and method == "POST":
            ok, msg, new_state = routes_mgr.toggle_route(route_id)
            if ok:
                for r in routes_mgr.list_routes():
                    if r["id"] == route_id:
                        target = f"http://{r['target_host']}:{r['target_port']}"
                        if new_state:
                            apply_route_tailscale(r["path"], target, mode=r.get("mode", "serve"))
                        else:
                            remove_route_tailscale(r["path"])
                        break
            await send_json_response(writer, 200 if ok else 400, {"success": ok, "message": msg, "enabled": new_state})
            return True

        if sub_action == "toggle-mode" and method == "POST":
            ok, msg, new_mode = routes_mgr.toggle_mode(route_id)
            if ok:
                for r in routes_mgr.list_routes():
                    if r["id"] == route_id:
                        target = f"http://{r['target_host']}:{r['target_port']}"
                        if r.get("enabled", True):
                            apply_route_tailscale(r["path"], target, mode=new_mode)
                        break
            await send_json_response(writer, 200 if ok else 400, {"success": ok, "message": msg, "mode": new_mode})
            return True

        if sub_action == "ping" and method == "POST":
            for r in routes_mgr.list_routes():
                if r["id"] == route_id:
                    online, latency = await routes_mgr.ping_target(r["target_host"], r["target_port"])
                    r["last_status"] = "online" if online else "offline"
                    r["last_latency_ms"] = latency
                    routes_mgr.save()
                    await send_json_response(writer, 200, {"online": online, "latency_ms": latency})
                    return True
            await send_json_response(writer, 404, {"error": "Route không tồn tại"})
            return True

        if method == "DELETE":
            for r in routes_mgr.list_routes():
                if r["id"] == route_id:
                    remove_route_tailscale(r["path"])
                    break
            ok, msg = routes_mgr.delete_route(route_id)
            await send_json_response(writer, 200 if ok else 400, {"success": ok, "message": msg})
            return True

        if method == "PUT":
            try:
                body_json = json.loads(req.body.decode("utf-8")) if req.body else {}
                ok, msg = routes_mgr.update_route(route_id, body_json)
                if ok:
                    for r in routes_mgr.list_routes():
                        if r["id"] == route_id:
                            target = f"http://{r['target_host']}:{r['target_port']}"
                            if r.get("enabled", True):
                                apply_route_tailscale(r["path"], target, mode=r.get("mode", "serve"))
                            else:
                                remove_route_tailscale(r["path"])
                            break
                await send_json_response(writer, 200 if ok else 400, {"success": ok, "message": msg})
            except Exception as e:
                await send_json_response(writer, 400, {"success": False, "message": str(e)})
            return True

    # 5. API Tailscale Serve Quick Setup & Sync
    if path == "/api/tailscale/serve" and method == "POST":
        try:
            body_json = json.loads(req.body.decode("utf-8")) if req.body else {}
            action = body_json.get("action", "serve_router")
            target_port = body_json.get("port", PORT)

            if action == "serve_router":
                res = configure_serve_router(target_port)
            elif action == "serve_all":
                res = configure_serve_gateway(target_port)
            elif action == "reset":
                res = reset_serve_config()
            else:
                res = configure_serve_router(target_port)

            status_code = 200 if res.get("success") else 400
            await send_json_response(writer, status_code, res)
        except Exception as e:
            await send_json_response(writer, 500, {"success": False, "error": str(e)})
        return True

    return False


async def handle_reverse_proxy(
    req: HTTPRequest,
    client_reader: asyncio.StreamReader,
    client_writer: asyncio.StreamWriter,
    client_ip: str,
    matched_route: Dict[str, Any],
    new_path: str,
) -> None:
    """Chuyển tiếp yêu cầu HTTP hoặc WebSocket tới dịch vụ đích."""
    target_host = matched_route["target_host"]
    target_port = matched_route["target_port"]
    start_t = time.time()

    # Bổ sung query string nếu có
    if req.query:
        upstream_path = f"{new_path}?{req.query}"
    else:
        upstream_path = new_path

    try:
        upstream_reader, upstream_writer = await asyncio.wait_for(
            asyncio.open_connection(target_host, target_port),
            timeout=3.0,
        )
    except Exception as e:
        latency = (time.time() - start_t) * 1000
        add_log(req.method, req.full_path, 502, f"{target_host}:{target_port}", latency)
        error_html = f"""<!DOCTYPE html>
<html lang="vi">
<head><meta charset="utf-8"><title>502 Bad Gateway - TailRouter</title>
<style>
body {{ font-family: system-ui, -apple-system, sans-serif; background: #0f172a; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }}
.card {{ background: #1e293b; padding: 2.5rem; border-radius: 1rem; max-width: 520px; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.5); border: 1px solid #334155; }}
h1 {{ color: #ef4444; font-size: 1.5rem; margin-top: 0; }}
p {{ line-height: 1.6; color: #94a3b8; }}
code {{ background: #0f172a; padding: 0.2rem 0.4rem; border-radius: 0.25rem; color: #38bdf8; }}
.btn {{ display: inline-block; margin-top: 1rem; padding: 0.6rem 1.2rem; background: #2563eb; color: white; text-decoration: none; border-radius: 0.5rem; font-weight: 500; }}
</style></head>
<body><div class="card">
<h1>⚠️ 502 Bad Gateway</h1>
<p>TailRouter không thể kết nối tới dịch vụ đích tại <code>{target_host}:{target_port}</code> cho route <code>{matched_route.get('path')}</code>.</p>
<p><strong>Nguyên nhân:</strong> Dịch vụ chưa khởi động hoặc cổng bị chặn (Lỗi: {e}).</p>
<a href="/router" class="btn">👉 Mở Bảng Quản Trị /router</a>
</div></body></html>"""
        await send_html_response(client_writer, error_html, 502)
        return

    # Chuẩn bị header cho upstream
    upstream_headers = []
    upstream_headers.append(f"{req.method} {upstream_path} {req.version}")

    is_ws = req.is_websocket()
    host_header_set = False

    for k, v in req.headers.items():
        k_lower = k.lower()
        if k_lower == "host":
            upstream_headers.append(f"Host: {target_host}:{target_port}")
            host_header_set = True
        else:
            upstream_headers.append(f"{k}: {v}")

    if not host_header_set:
        upstream_headers.append(f"Host: {target_host}:{target_port}")

    # Bổ sung Proxy Headers
    upstream_headers.append(f"X-Real-IP: {client_ip}")
    upstream_headers.append(f"X-Forwarded-For: {client_ip}")
    upstream_headers.append(f"X-Forwarded-Proto: http")
    upstream_headers.append(f"X-Forwarded-Prefix: {matched_route['path']}")

    header_payload = "\r\n".join(upstream_headers) + "\r\n\r\n"
    upstream_writer.write(header_payload.encode("utf-8"))

    # Gửi kèm body nếu có
    if req.body:
        upstream_writer.write(req.body)
    await upstream_writer.drain()

    routes_mgr.increment_hits(matched_route["id"])

    # Xử lý WebSocket
    if is_ws:
        logger.info(f"[WebSocket] Bắt đầu phiên pipe {req.path} -> {target_host}:{target_port}")
        add_log("WS", req.path, 101, f"{target_host}:{target_port}", 0)
        await asyncio.gather(
            pipe_stream(client_reader, upstream_writer),
            pipe_stream(upstream_reader, client_writer),
        )
        return

    # Xử lý phản hồi HTTP thông thường
    try:
        status_line = await upstream_reader.readline()
        if not status_line:
            client_writer.close()
            return
        
        client_writer.write(status_line)

        status_code = 200
        try:
            status_parts = status_line.decode("utf-8", errors="ignore").split(" ")
            if len(status_parts) >= 2:
                status_code = int(status_parts[1])
        except Exception:
            pass

        content_length = None
        is_chunked = False

        while True:
            resp_header_line = await upstream_reader.readline()
            if not resp_header_line or resp_header_line == b"\r\n" or resp_header_line == b"\n":
                client_writer.write(b"\r\n")
                break
            
            client_writer.write(resp_header_line)
            h_str = resp_header_line.decode("utf-8", errors="ignore").strip()
            if ":" in h_str:
                hk, hv = h_str.split(":", 1)
                hk_l = hk.strip().lower()
                hv_v = hv.strip().lower()
                if hk_l == "content-length":
                    try:
                        content_length = int(hv_v)
                    except ValueError:
                        pass
                elif hk_l == "transfer-encoding" and "chunked" in hv_v:
                    is_chunked = True

        await client_writer.drain()

        # Stream body từ upstream về client
        if content_length is not None:
            bytes_left = content_length
            while bytes_left > 0:
                chunk = await upstream_reader.read(min(bytes_left, 65536))
                if not chunk:
                    break
                client_writer.write(chunk)
                await client_writer.drain()
                bytes_left -= len(chunk)
        else:
            # Chunked hoặc stream video/dữ liệu cho tới khi EOF
            while True:
                chunk = await upstream_reader.read(65536)
                if not chunk:
                    break
                client_writer.write(chunk)
                await client_writer.drain()

        latency = (time.time() - start_t) * 1000
        add_log(req.method, req.full_path, status_code, f"{target_host}:{target_port}", latency)

    except Exception as e:
        logger.error(f"Lỗi truyền dữ liệu proxy: {e}")
    finally:
        try:
            upstream_writer.close()
            await upstream_writer.wait_closed()
        except Exception:
            pass


async def handle_client(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    """Xử lý từng kết nối TCP đến gateway 65534."""
    peername = writer.get_extra_info("peername")
    client_ip = peername[0] if peername else "127.0.0.1"

    req = await parse_http_request(reader)
    if not req:
        try:
            writer.close()
        except Exception:
            pass
        return

    path = req.path

    # Chuyển hướng trang chủ "/" về "/router" nếu chưa có route "/" được định nghĩa riêng
    if path == "/" or path == "":
        match = routes_mgr.match_route("/")
        if not match:
            # 302 Redirect về /router
            redir_resp = [
                "HTTP/1.1 302 Found",
                "Location: /router",
                "Content-Length: 0",
                "Connection: close",
                "\r\n\r\n",
            ]
            writer.write("\r\n".join(redir_resp).encode("utf-8"))
            await writer.drain()
            try:
                writer.close()
            except Exception:
                pass
            return

    # Phục vụ Giao diện Dashboard tại /router
    if path == "/router" or path == "/router/":
        if os.path.exists(INDEX_HTML):
            with open(INDEX_HTML, "r", encoding="utf-8") as f:
                html = f.read()
            await send_html_response(writer, html)
        else:
            await send_html_response(writer, "<h1>404 - Chưa tìm thấy file index.html giao diện /router</h1>", 404)
        try:
            writer.close()
        except Exception:
            pass
        return

    # Phục vụ Favicon
    if path == "/favicon.ico" or path == "/favicon.svg":
        headers = [
            "HTTP/1.1 200 OK",
            "Content-Type: image/svg+xml",
            "Connection: close",
            "\r\n\r\n",
        ]
        svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="#3b82f6"><path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/></svg>'
        writer.write("\r\n".join(headers).encode("utf-8") + svg.encode("utf-8"))
        await writer.drain()
        try:
            writer.close()
        except Exception:
            pass
        return

    # Xử lý các REST API
    if path.startswith("/api/"):
        handled = await handle_api_request(req, writer, client_ip)
        if handled:
            try:
                writer.close()
            except Exception:
                pass
            return

    # Xử lý Reverse Proxy cho các route đã định nghĩa
    match_result = routes_mgr.match_route(path)
    if match_result:
        matched_route, new_path = match_result
        await handle_reverse_proxy(req, reader, writer, client_ip, matched_route, new_path)
    else:
        # Route không tồn tại
        add_log(req.method, path, 404, "None", 0)
        not_found_html = f"""<!DOCTYPE html>
<html lang="vi">
<head><meta charset="utf-8"><title>404 Route Not Found</title>
<style>
body {{ font-family: system-ui, -apple-system, sans-serif; background: #0b0f19; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }}
.box {{ background: #161f30; padding: 2.5rem; border-radius: 1rem; max-width: 500px; border: 1px solid #1e293b; text-align: center; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5); }}
h1 {{ color: #38bdf8; margin: 0 0 1rem; font-size: 2rem; }}
p {{ color: #94a3b8; font-size: 1rem; line-height: 1.5; }}
code {{ color: #f43f5e; background: #0f172a; padding: 0.2rem 0.5rem; border-radius: 0.3rem; }}
.btn {{ display: inline-block; margin-top: 1.5rem; padding: 0.75rem 1.5rem; background: #3b82f6; color: white; text-decoration: none; border-radius: 0.5rem; font-weight: 600; }}
.btn:hover {{ background: #2563eb; }}
</style></head>
<body><div class="box">
<h1>404 - Chưa Cấu Hình Route</h1>
<p>Đường dẫn <code>{path}</code> chưa được cấu hình định tuyến trên cổng 65534 của máy chủ.</p>
<a href="/router" class="btn">⚙️ Mở Bảng Quản Trị /router để thêm</a>
</div></body></html>"""
        await send_html_response(writer, not_found_html, 404)

    try:
        writer.close()
        await writer.wait_closed()
    except Exception:
        pass


async def background_health_checker() -> None:
    """Nhiệm vụ kiểm tra định kỳ trạng thái các route."""
    while True:
        try:
            await asyncio.sleep(30)
            await routes_mgr.check_all_health()
        except asyncio.CancelledError:
            break
        except Exception:
            pass


async def main() -> None:
    # Ghi PID file
    pid_file = os.path.join(BASE_DIR, "server.pid")
    try:
        with open(pid_file, "w") as f:
            f.write(str(os.getpid()))
    except Exception:
        pass

    server = await asyncio.start_server(handle_client, HOST, PORT, reuse_address=True)
    addrs = ", ".join(str(sock.getsockname()) for sock in server.sockets)
    logger.info(f"============================================================")
    logger.info(f"🚀 TailRouter & Gateway đang chạy trên cổng {PORT}!")
    logger.info(f"👉 Giao diện quản trị máy chủ: http://localhost:{PORT}/router")
    
    ts_info = get_tailscale_info()
    fqdn = ts_info.get("fqdn")
    if ts_info.get("ipv4"):
        logger.info(f"🌐 Tailnet IP trực tiếp:     http://{ts_info['ipv4']}:{PORT}/router")
    if fqdn:
        logger.info(f"🔗 Tên miền MagicDNS FQDN:   https://{fqdn}/router")

    # TỰ ĐỘNG ĐỒNG BỘ TAILSCALE SERVE & FUNNEL CHO TẤT CẢ ROUTERS
    logger.info(f"⚙️  Đang tự động đồng bộ Tailscale Serve & Funnel cho các router...")
    sync_res = sync_all_routes_tailscale(routes_mgr.list_routes(), port=PORT)
    for p, r in sync_res.items():
        m_str = r.get("mode", "serve").upper()
        if r.get("success"):
            logger.info(f"  ✓ https://{fqdn}{p} [{m_str}] -> OK")
        else:
            logger.warning(f"  ✗ {p} [{m_str}] -> {r.get('error')}")

    logger.info(f"============================================================")

    # Chạy health check ngầm
    checker_task = asyncio.create_task(background_health_checker())

    async with server:
        try:
            await server.serve_forever()
        except (KeyboardInterrupt, asyncio.CancelledError):
            pass
        finally:
            checker_task.cancel()
            try:
                if os.path.exists(pid_file):
                    os.remove(pid_file)
            except Exception:
                pass


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n[TailRouter] Đã dừng máy chủ.")
