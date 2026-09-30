#!/usr/bin/env python3
import asyncio
import random
import sys

# Pool of realistic service names for mock services
SERVICE_TEMPLATES = [
    "Web Dashboard", "Grafana Monitor", "Prometheus Metrics", "Redis UI",
    "Home Assistant", "Jellyfin Media", "Nextcloud Storage", "Nginx Proxy",
    "Pi-hole Admin", "Portainer CE", "Uptime Kuma", "Vaultwarden",
    "Transmission Torrent", "qBittorrent UI", "AdGuard DNS", "WireGuard UI",
    "Syncthing Sync", "Duplicati Backup", "Mealie Recipes", "Paperless Docs",
    "PhotoPrism Gallery", "Immich Photos", "Sonarr Series", "Radarr Cinema",
    "Lidarr Tracks", "Readarr Books", "Bazarr Subtitles", "Prowlarr Indexer",
    "Overseerr Media", "Calibre Books", "Navidrome Streamer", "Audiobookshelf",
    "Homebridge IoT", "Zigbee2MQTT", "Mosquitto Broker", "Node-RED Flow",
    "TasmoAdmin IoT", "ESPHome Panel", "Scrutiny SMART", "Netdata Metrics",
    "Glances System", "Cockpit Console", "Webmin Control", "phpMyAdmin DB",
    "Adminer DB", "PostgreSQL Admin", "Gitea Repo", "Forgejo Git",
    "Drone CI", "Woodpecker CI", "MinIO S3 Storage", "Ollama LLM Engine",
    "LocalAI Server", "Stable Diffusion UI", "ComfyUI Pipeline", "ChromaDB Vector",
    "Qdrant Search", "Meilisearch Engine", "Typesense Node", "Elasticsearch Core",
    "Kibana Viz", "RabbitMQ Broker", "Kafka Streams", "EMQX Broker",
    "InfluxDB TimeSeries", "ClickHouse Analytics", "MongoDB Express", "Cassandra Cluster",
    "PocketBase App", "Supabase Gateway", "Appwrite Cloud", "Directus Headless",
    "Strapi CMS", "Ghost Publishing", "WordPress Blog", "Drupal Hub",
    "Matrix Synapse", "Element Chat", "Vault Secrets", "Traefik Router",
    "Caddy Web Server", "Apache Traffic", "Haproxy Balancer", "Varnish Cache",
    "Memcached Memory", "Squid Proxy", "Shadowsocks Server", "OpenVPN Gateway",
    "Tailscale Exit Node", "ZeroTier Member", "K3s Cluster API", "MicroK8s Edge",
    "Rancher Manager", "Portainer Agent", "Cadvisor Containers", "Fluentd Log Pipe",
    "Loki Log Store", "Jaeger Tracing", "Tempo Telemetry", "OpenTelemetry Collector"
]

RESERVED_PORTS = {22, 53, 80, 443, 3389, 41641, 44698, 65534}

def generate_services(count=100, seed=42):
    random.seed(seed)
    available_ports = [p for p in range(1025, 65530) if p not in RESERVED_PORTS]
    selected_ports = random.sample(available_ports, count)
    
    services = []
    for i, port in enumerate(selected_ports):
        base_name = SERVICE_TEMPLATES[i % len(SERVICE_TEMPLATES)]
        tag = f"#{random.randint(10, 99)}" if i >= len(SERVICE_TEMPLATES) else ""
        full_name = f"{base_name} {tag}".strip()
        services.append((port, full_name))
    return services

def make_handler(port, name):
    async def handle_client(reader, writer):
        try:
            line = await asyncio.wait_for(reader.readline(), timeout=2.0)
            body = f"""<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>{name}</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0b0f19; color: #f8fafc; padding: 2.5rem;">
  <div style="max-width: 600px; margin: 0 auto; background: #182238; border: 1px solid #2a3b5c; border-radius: 12px; padding: 2rem;">
    <h1 style="color: #38bdf8; margin-top: 0;">⚡ {name}</h1>
    <p style="font-size: 1.1rem;">Cổng dịch vụ ảo thử nghiệm: <strong style="color: #4ade80;">{port}</strong></p>
    <p>Trạng thái: <span style="background: #064e3b; color: #6ee7b7; padding: 0.2rem 0.6rem; border-radius: 9999px; font-weight: 600;">🟢 Online (Mock Service)</span></p>
    <p style="color: #94a3b8; font-size: 0.9rem; margin-top: 1.5rem;">Dịch vụ được tạo tự động để thử nghiệm tính năng quét cổng (Auto-Scan) và định tuyến (Routing) của TailRouter.</p>
  </div>
</body>
</html>""".encode("utf-8")
            resp = (
                f"HTTP/1.1 200 OK\r\n"
                f"Content-Type: text/html; charset=utf-8\r\n"
                f"Content-Length: {len(body)}\r\n"
                f"Connection: close\r\n\r\n"
            ).encode("utf-8") + body
            writer.write(resp)
            await writer.drain()
        except Exception:
            pass
        finally:
            try:
                writer.close()
                await writer.wait_closed()
            except Exception:
                pass
    return handle_client

async def main():
    count = 100
    if len(sys.argv) > 1 and sys.argv[1].isdigit():
        count = int(sys.argv[1])

    services = generate_services(count)
    servers = []
    
    print(f"🚀 Đang khởi tạo {len(services)} cổng dịch vụ ảo ngẫu nhiên...", flush=True)
    
    bound_count = 0
    for port, name in services:
        try:
            srv = await asyncio.start_server(make_handler(port, name), '0.0.0.0', port)
            servers.append(srv)
            bound_count += 1
        except Exception as e:
            print(f"⚠️ Không thể bind cổng {port}: {e}", flush=True)

    print(f"✅ Đã kích hoạt thành công {bound_count}/{len(services)} cổng ảo!", flush=True)
    print("📋 Danh sách các cổng mở:", flush=True)
    ports_list = [str(p) for p, _ in services]
    for i in range(0, len(ports_list), 10):
        print("   " + ", ".join(ports_list[i:i+10]), flush=True)

    print("\n💡 Bấm Ctrl+C để dừng hoặc chạy 'pkill -f mock_100_ports.py'", flush=True)

    await asyncio.gather(*(s.serve_forever() for s in servers))

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except (KeyboardInterrupt, SystemExit):
        print("\n🛑 Đã dừng toàn bộ cổng ảo.")
