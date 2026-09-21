"""
routes_manager.py - Quản lý cấu hình định tuyến (Routes) và lưu trữ JSON
"""

import asyncio
import json
import os
import re
import socket
import time
from typing import Any, Dict, List, Optional, Tuple

CONFIG_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), "config")
CONFIG_FILE = os.path.join(CONFIG_DIR, "routes.json")

RESERVED_PATHS = ["/router", "/api", "/favicon.ico"]


class RoutesManager:
    def __init__(self, config_path: str = CONFIG_FILE):
        self.config_path = config_path
        os.makedirs(os.path.dirname(self.config_path), exist_ok=True)
        self._routes: Dict[str, Dict[str, Any]] = {}
        self.load()

    def load(self) -> None:
        """Tải danh sách route từ file JSON."""
        if os.path.exists(self.config_path):
            try:
                with open(self.config_path, "r", encoding="utf-8") as f:
                    data = json.load(f)
                    if isinstance(data, list):
                        self._routes = {r["id"]: r for r in data if "id" in r}
                    elif isinstance(data, dict):
                        self._routes = data
                    return
            except Exception as e:
                print(f"[RoutesManager] Lỗi đọc config: {e}")
        
        # Mặc định khởi tạo cấu hình mẫu nếu có container Bambu 8080
        self._routes = {}
        self._init_defaults()
        self.save()

    def _init_defaults(self) -> None:
        """Thêm các route gợi ý mặc định nếu có dịch vụ sẵn."""
        # Kiểm tra xem cổng 8080 có mở không để add mẫu
        route_id = "route_bambu_8080"
        self._routes[route_id] = {
            "id": route_id,
            "name": "Bambu 3D Timelapse",
            "path": "/bambu",
            "target_host": "127.0.0.1",
            "target_port": 8080,
            "strip_prefix": True,
            "mode": "serve",
            "enabled": True,
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            "hits": 0,
            "last_status": "unknown",
            "last_latency_ms": None,
            "notes": "Tự động phát hiện container Bambu Lab A1",
        }

    def save(self) -> bool:
        """Lưu danh sách route vào file JSON."""
        try:
            temp_file = f"{self.config_path}.tmp"
            with open(temp_file, "w", encoding="utf-8") as f:
                json.dump(list(self._routes.values()), f, indent=2, ensure_ascii=False)
            os.replace(temp_file, self.config_path)
            return True
        except Exception as e:
            print(f"[RoutesManager] Lỗi lưu config: {e}")
            return False

    def list_routes(self) -> List[Dict[str, Any]]:
        """Trả về toàn bộ route đã sắp xếp theo ngày tạo hoặc path."""
        routes = list(self._routes.values())
        routes.sort(key=lambda x: x.get("path", ""))
        return routes

    def get_configured_ports(self) -> List[int]:
        """Lấy danh sách các port đã được cấu hình trong router."""
        return [r["target_port"] for r in self._routes.values() if "target_port" in r]

    def add_route(self, data: Dict[str, Any]) -> Tuple[bool, str, Optional[Dict[str, Any]]]:
        """Thêm route mới với kiểm tra ràng buộc đường dẫn và cổng."""
        raw_path = data.get("path", "").strip()
        if not raw_path:
            return False, "Đường dẫn (path) không được để trống.", None

        # Chuẩn hóa path: bắt đầu bằng / và bỏ trailing slash
        path = "/" + raw_path.lstrip("/").rstrip("/")
        if not path:
            path = "/"

        # Kiểm tra đường dẫn cấm
        for reserved in RESERVED_PATHS:
            if path == reserved or path.startswith(f"{reserved}/"):
                return False, f"Đường dẫn '{path}' trùng với hệ thống quản trị ({reserved}). Vui lòng chọn path khác.", None

        # Kiểm tra trùng lặp
        for r in self._routes.values():
            if r["path"].lower() == path.lower():
                return False, f"Đường dẫn '{path}' đã tồn tại cho route '{r.get('name', r['id'])}'.", None

        # Kiểm tra target port
        try:
            port = int(data.get("target_port", 0))
            if port < 1 or port > 65535:
                return False, "Cổng đích (target port) phải trong khoảng 1 - 65535.", None
        except (ValueError, TypeError):
            return False, "Cổng đích không hợp lệ.", None

        route_id = f"route_{int(time.time())}_{port}"
        name = data.get("name", "").strip() or f"Dịch vụ cổng {port}"
        target_host = data.get("target_host", "").strip() or "127.0.0.1"
        strip_prefix = bool(data.get("strip_prefix", False))
        mode = "funnel" if str(data.get("mode", "serve")).strip().lower() == "funnel" else "serve"
        enabled = bool(data.get("enabled", True))
        notes = data.get("notes", "").strip()

        new_route = {
            "id": route_id,
            "name": name,
            "path": path,
            "target_host": target_host,
            "target_port": port,
            "strip_prefix": strip_prefix,
            "mode": mode,
            "enabled": enabled,
            "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
            "hits": 0,
            "last_status": "unknown",
            "last_latency_ms": None,
            "notes": notes,
        }

        self._routes[route_id] = new_route
        self.save()
        return True, "Thêm route thành công.", new_route

    def update_route(self, route_id: str, data: Dict[str, Any]) -> Tuple[bool, str]:
        """Cập nhật route đã tồn tại."""
        if route_id not in self._routes:
            return False, "Route không tồn tại."

        route = self._routes[route_id]

        if "path" in data:
            new_path = "/" + data["path"].strip().lstrip("/").rstrip("/")
            if not new_path:
                new_path = "/"
            for reserved in RESERVED_PATHS:
                if new_path == reserved or new_path.startswith(f"{reserved}/"):
                    return False, f"Đường dẫn '{new_path}' bị xung đột với hệ thống quản trị."
            for rid, r in self._routes.items():
                if rid != route_id and r["path"].lower() == new_path.lower():
                    return False, f"Đường dẫn '{new_path}' đã được dùng bởi route khác."
            route["path"] = new_path

        if "name" in data and data["name"].strip():
            route["name"] = data["name"].strip()

        if "target_host" in data and data["target_host"].strip():
            route["target_host"] = data["target_host"].strip()

        if "target_port" in data:
            try:
                p = int(data["target_port"])
                if 1 <= p <= 65535:
                    route["target_port"] = p
            except ValueError:
                pass

        if "strip_prefix" in data:
            route["strip_prefix"] = bool(data["strip_prefix"])

        if "mode" in data:
            m = str(data["mode"]).strip().lower()
            route["mode"] = "funnel" if m == "funnel" else "serve"

        if "enabled" in data:
            route["enabled"] = bool(data["enabled"])

        if "notes" in data:
            route["notes"] = data["notes"].strip()

        self.save()
        return True, "Cập nhật route thành công."

    def toggle_mode(self, route_id: str) -> Tuple[bool, str, str]:
        """Chuyển đổi giữa chế độ SERVE (nội bộ) và FUNNEL (công khai)."""
        if route_id in self._routes:
            curr = self._routes[route_id].get("mode", "serve")
            new_mode = "funnel" if curr == "serve" else "serve"
            self._routes[route_id]["mode"] = new_mode
            self.save()
            return True, f"Đã chuyển chế độ sang {new_mode.upper()}.", new_mode
        return False, "Route không tồn tại.", "serve"

    def delete_route(self, route_id: str) -> Tuple[bool, str]:
        """Xóa route khỏi cấu hình."""
        if route_id in self._routes:
            del self._routes[route_id]
            self.save()
            return True, "Đã xóa route."
        return False, "Route không tồn tại."

    def toggle_route(self, route_id: str) -> Tuple[bool, str, bool]:
        """Bật/tắt trạng thái route."""
        if route_id in self._routes:
            curr = self._routes[route_id].get("enabled", True)
            self._routes[route_id]["enabled"] = not curr
            self.save()
            return True, f"Đã {'bật' if not curr else 'tắt'} route.", not curr
        return False, "Route không tồn tại.", False

    def increment_hits(self, route_id: str) -> None:
        """Tăng bộ đếm lượt truy cập."""
        if route_id in self._routes:
            self._routes[route_id]["hits"] = self._routes[route_id].get("hits", 0) + 1

    def match_route(self, request_path: str) -> Optional[Tuple[Dict[str, Any], str]]:
        """
        Tìm route khớp với đường dẫn (Longest Prefix Match).
        Trả về (route_object, rewrite_path).
        """
        best_match = None
        best_len = -1

        for route in self._routes.values():
            if not route.get("enabled", True):
                continue
            rpath = route["path"]

            # Exact match hoặc prefix match (/bambu hoặc /bambu/...)
            if request_path == rpath or request_path.startswith(f"{rpath}/") or (rpath == "/" and request_path.startswith("/")):
                if len(rpath) > best_len:
                    best_len = len(rpath)
                    best_match = route

        if not best_match:
            return None

        # Tính toán target path nếu strip_prefix bật
        target_path = request_path
        if best_match.get("strip_prefix", False) and best_match["path"] != "/":
            prefix = best_match["path"]
            if target_path.startswith(prefix):
                target_path = target_path[len(prefix):]
                if not target_path.startswith("/"):
                    target_path = "/" + target_path

        return best_match, target_path

    async def ping_target(self, target_host: str, target_port: int, timeout: float = 1.0) -> Tuple[bool, Optional[float]]:
        """Kiểm tra kết nối TCP tới đích và tính latency."""
        t0 = time.time()
        try:
            _, writer = await asyncio.wait_for(
                asyncio.open_connection(target_host, target_port),
                timeout=timeout,
            )
            latency = round((time.time() - t0) * 1000, 1)
            writer.close()
            try:
                await writer.wait_closed()
            except Exception:
                pass
            return True, latency
        except Exception:
            return False, None

    async def check_all_health(self) -> None:
        """Kiểm tra sức khỏe đồng thời cho tất cả các route."""
        tasks = []
        routes_list = list(self._routes.values())
        for r in routes_list:
            tasks.append(self.ping_target(r["target_host"], r["target_port"]))

        results = await asyncio.gather(*tasks, return_exceptions=True)
        for r, res in zip(routes_list, results):
            if isinstance(res, tuple) and res[0]:
                r["last_status"] = "online"
                r["last_latency_ms"] = res[1]
            else:
                r["last_status"] = "offline"
                r["last_latency_ms"] = None

    def sync_from_tailscale(self, discovered_routes: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """
        Đồng bộ các routes đang mở trên Tailscale Serve / Funnel vào routes_manager:
        - Nếu route chưa có trong cấu hình: tự động thêm mới vào _routes.
        - Nếu route đã có: cập nhật mode ('serve' / 'funnel') nếu khác.
        Trả về danh sách các route mới được thêm.
        """
        new_added = []
        modified = False

        path_map = {r["path"].lower(): r for r in self._routes.values()}

        for disc in discovered_routes:
            d_path = disc.get("path", "").strip()
            if not d_path or d_path == "/router":
                continue

            clean_path = "/" + d_path.lstrip("/").rstrip("/")
            target_host = disc.get("target_host", "127.0.0.1")
            target_port = disc.get("target_port", 80)
            mode = disc.get("mode", "serve")

            if clean_path.lower() in path_map:
                existing = path_map[clean_path.lower()]
                if existing.get("mode") != mode:
                    existing["mode"] = mode
                    modified = True
                if not existing.get("enabled", True):
                    existing["enabled"] = True
                    modified = True
            else:
                route_id = f"route_ts_{int(time.time())}_{target_port}_{abs(hash(clean_path)) % 10000}"
                clean_name = clean_path.lstrip("/").replace("-", " ").title()
                suggested_name = f"⚡ {clean_name}"
                new_route = {
                    "id": route_id,
                    "name": suggested_name,
                    "path": clean_path,
                    "target_host": target_host,
                    "target_port": target_port,
                    "strip_prefix": False,
                    "mode": mode,
                    "enabled": True,
                    "created_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
                    "hits": 0,
                    "last_status": "online",
                    "last_latency_ms": None,
                    "notes": "Tự động phát hiện từ lệnh Tailscale CLI",
                }
                self._routes[route_id] = new_route
                path_map[clean_path.lower()] = new_route
                new_added.append(new_route)
                modified = True

        if modified:
            self.save()

        return new_added

