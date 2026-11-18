package main

import (
	"bytes"
	"embed"
	"log"
	"net/http"
	"os"
)

//go:embed index.html demo.html
var content embed.FS

const demoMockScript = `<script>
(function() {
  const startTs = Date.now();
  let mockServeConfigured = true;
  let mockAutostart = true;
  let mockRoutes = [
    { id: "route_bambu", name: "Bambu 3D Printer Cam", path: "/bambu", target_host: "127.0.0.1", target_port: 8080, strip_prefix: true, enabled: true, mode: "serve", hits: 142, last_status: "online", last_latency_ms: 0.42, notes: "Bambu Lab A1 Timelapse & Proxy" },
    { id: "route_git", name: "GitAutoHub Backup", path: "/git", target_host: "127.0.0.1", target_port: 18090, strip_prefix: false, enabled: true, mode: "serve", hits: 89, last_status: "online", last_latency_ms: 0.31, notes: "Local Git WebUI" },
    { id: "route_ha", name: "Home Assistant", path: "/ha", target_host: "127.0.0.1", target_port: 8123, strip_prefix: false, enabled: true, mode: "funnel", hits: 315, last_status: "online", last_latency_ms: 0.55, notes: "Smart Home Control (Public Funnel)" },
    { id: "system-tailrouter-gateway", name: "TailRouter Gateway", path: "/router", target_host: "127.0.0.1", target_port: 65534, strip_prefix: false, enabled: true, mode: "serve", hits: 24, last_status: "online", last_latency_ms: 0.12, notes: "Trang điều khiển TailRouter qua Tailscale Serve (/router)" }
  ];
  let mockLogs = [
    { time: "13:32:10", method: "GET", path: "/bambu/api/status", target: "127.0.0.1:8080", status: 200, latency_ms: 0.41, client_ip: "100.100.1.8" },
    { time: "13:32:14", method: "GET", path: "/git/", target: "127.0.0.1:18090", status: 200, latency_ms: 0.31, client_ip: "100.100.1.12" },
    { time: "13:32:19", method: "POST", path: "/ha/api/webhook", target: "127.0.0.1:8123", status: 200, latency_ms: 0.58, client_ip: "Funnel" },
    { time: "13:32:25", method: "GET", path: "/router", target: "127.0.0.1:65534", status: 200, latency_ms: 0.15, client_ip: "100.100.1.3" }
  ];
  const mockPorts = [
    { port: 22, process: "sshd", service: "SSH", description: "OpenSSH Remote Access", protocol: "tcp", latency_ms: 0.2, suggested_path: "/ssh" },
    { port: 3000, process: "docker:grafana", service: "HTTP Web", description: "Grafana Monitoring", title: "Grafana", protocol: "http", latency_ms: 0.4, suggested_path: "/grafana" },
    { port: 8080, process: "bambu-proxy", service: "HTTP Web", description: "Bambu Timelapse Web Proxy", title: "Bambu Cam", protocol: "http", latency_ms: 0.4, suggested_path: "/bambu" },
    { port: 8123, process: "homeassistant", service: "HTTP Web", description: "Home Assistant Smart Home", title: "Home Assistant", protocol: "http", latency_ms: 0.5, suggested_path: "/ha" },
    { port: 18090, process: "git-webui", service: "HTTP Web", description: "GitAutoHub - Homelab Git Auto Backup", title: "GitAutoHub", protocol: "http", latency_ms: 0.3, suggested_path: "/git" }
  ];

  function jsonRes(obj, status = 200) {
    return Promise.resolve(new Response(JSON.stringify(obj), {
      status,
      headers: { "Content-Type": "application/json; charset=utf-8" }
    }));
  }

  const origFetch = window.fetch.bind(window);
  window.fetch = function(input, init = {}) {
    const url = typeof input === "string" ? input : (input && input.url) || "";
    if (!url.startsWith("/api/")) return origFetch(input, init);
    const method = (init.method || "GET").toUpperCase();
    const path = url.split("?")[0];

    if (path === "/api/status") {
      return jsonRes({
        status: "online",
        version: "2.2.0",
        pid: 1838166,
        uptime_seconds: Math.floor((Date.now() - startTs) / 1000) + 14280,
        routes_count: mockRoutes.length,
        routes: mockRoutes,
        platform: "linux",
        arch: "amd64",
        tailscale: {
          running: true,
          node_name: "Giang3DLab-HomeLab",
          ipv4: "100.100.1.3",
          ips: ["100.100.1.3"],
          fqdn: "homelab.tailaf2792.ts.net",
          serve_configured: mockServeConfigured,
          needs_operator: false
        },
        autostart: {
          supported: true,
          platform: "linux",
          enabled: mockAutostart,
          active: mockAutostart,
          service_type: "systemd",
          service_file: "~/.config/systemd/user/tailrouter.service"
        },
        logs: mockLogs
      });
    }

    if (path === "/api/scan") {
      return jsonRes({ success: true, ports: mockPorts });
    }

    if (path === "/api/routes" && method === "POST") {
      const body = JSON.parse(init.body || "{}");
      const nr = {
        id: "route_" + Date.now(),
        name: body.name || "New Service",
        path: body.path || "/app",
        target_host: body.target_host || "127.0.0.1",
        target_port: Number(body.target_port) || 8080,
        strip_prefix: !!body.strip_prefix,
        enabled: body.enabled !== false,
        mode: body.mode || "serve",
        hits: 0,
        last_status: "online",
        last_latency_ms: 0.35,
        notes: body.notes || ""
      };
      mockRoutes.unshift(nr);
      return jsonRes({ success: true, route: nr, message: "Đã thêm tuyến đường thành công (Demo)" });
    }

    if (path === "/api/routes/reset" && method === "POST") {
      mockRoutes = [];
      return jsonRes({ success: true, message: "Đã đặt lại 0 routes (Demo)" });
    }

    if (path.startsWith("/api/routes/")) {
      const rest = path.slice("/api/routes/".length).split("/");
      const id = decodeURIComponent(rest[0]);
      const action = rest[1] || "";
      const idx = mockRoutes.findIndex(r => r.id === id);
      if (idx === -1) return jsonRes({ error: "Not found" }, 404);
      const r = mockRoutes[idx];

      if (method === "DELETE") {
        mockRoutes.splice(idx, 1);
        return jsonRes({ success: true, message: "Đã xóa route (Demo)" });
      }
      if (method === "PUT") {
        const body = JSON.parse(init.body || "{}");
        Object.assign(r, body, { target_port: Number(body.target_port) || r.target_port });
        return jsonRes({ success: true, route: r, message: "Đã cập nhật route (Demo)" });
      }
      if (action === "toggle") {
        r.enabled = !r.enabled;
        return jsonRes({ success: true, route: r });
      }
      if (action === "toggle-mode") {
        r.mode = r.mode === "serve" ? "funnel" : "serve";
        return jsonRes({ success: true, route: r, message: "Đã chuyển sang " + r.mode.toUpperCase() });
      }
      if (action === "ping") {
        r.last_latency_ms = Number((Math.random() * 0.45 + 0.2).toFixed(3));
        r.last_status = "online";
        return jsonRes({ success: true, latency_ms: r.last_latency_ms, status: "online" });
      }
    }

    if (path === "/api/autostart" && method === "POST") {
      mockAutostart = !mockAutostart;
      return jsonRes({ success: true, enabled: mockAutostart, message: mockAutostart ? "Đã bật tự khởi động (Demo)" : "Đã tắt tự khởi động (Demo)" });
    }

    if (path === "/api/logs/clear" && method === "POST") {
      mockLogs = [];
      return jsonRes({ success: true, message: "Đã xóa nhật ký (Demo)" });
    }

    if (path.startsWith("/api/tailscale/")) {
      if (path.endsWith("/reset")) mockServeConfigured = false;
      else mockServeConfigured = true;
      return jsonRes({ success: true, message: "Đã cập nhật Tailscale Serve (Demo)" });
    }

    if (path === "/api/system/uninstall") {
      return jsonRes({ success: true, message: "Demo mode: Uninstall simulated safely." });
    }

    return jsonRes({ success: true });
  };
})();
</script>`

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "49005"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/demo-app", func(w http.ResponseWriter, r *http.Request) {
		data, err := content.ReadFile("demo.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		injected := bytes.Replace(data, []byte("<head>"), []byte("<head>\n"+demoMockScript), 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		_, _ = w.Write(injected)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/index.html" {
			http.NotFound(w, r)
			return
		}
		data, err := content.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		_, _ = w.Write(data)
	})

	addr := "127.0.0.1:" + port
	log.Printf("TailRouter Landing Page listening on http://%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
