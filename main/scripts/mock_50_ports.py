#!/usr/bin/env python3
import asyncio
import sys

SERVICES = [
    (8001, "Web Dashboard Alpha"),
    (8002, "Grafana Metrics Monitor"),
    (8003, "Prometheus Server"),
    (8004, "Redis Commander UI"),
    (8005, "Home Assistant Hub"),
    (8006, "Jellyfin Media Server"),
    (8007, "Nextcloud File Cloud"),
    (8008, "Nginx Proxy Manager"),
    (8009, "Pi-hole Network Admin"),
    (8010, "Portainer Docker CE"),
    (8011, "Uptime Kuma Health"),
    (8012, "Vaultwarden Passwords"),
    (8013, "Transmission Torrent"),
    (8014, "qBittorrent Web UI"),
    (8015, "AdGuard Home DNS"),
    (8016, "WireGuard Easy VPN"),
    (8017, "Syncthing File Sync"),
    (8018, "Duplicati Backup"),
    (8019, "Mealie Recipe Manager"),
    (8020, "Paperless-ngx Archive"),
    (8021, "Photoprism Photos"),
    (8022, "Immich Photo Engine"),
    (8023, "Sonarr TV Shows"),
    (8024, "Radarr Movie Library"),
    (8025, "Lidarr Music Library"),
    (8026, "Readarr E-Book Library"),
    (8027, "Bazarr Subtitle Manager"),
    (8028, "Prowlarr Indexer"),
    (8029, "Overseerr Requests"),
    (8030, "Calibre-web Library"),
    (8031, "Navidrome Music Streamer"),
    (8032, "Audiobookshelf Player"),
    (8033, "Homebridge Apple Home"),
    (8034, "Zigbee2MQTT Gateway"),
    (8035, "Mosquitto MQTT Broker"),
    (8036, "Node-RED Automation"),
    (8037, "TasmoAdmin IoT"),
    (8038, "ESPHome Dashboard"),
    (8039, "Scrutiny SMART Monitor"),
    (8040, "Netdata Performance"),
    (8041, "Glances System Monitor"),
    (8042, "Cockpit Server Admin"),
    (8043, "Webmin Unix Control"),
    (8044, "phpMyAdmin Database"),
    (8045, "Adminer Lightweight DB"),
    (8046, "PostgreSQL pgAdmin"),
    (8047, "Gitea Git Server"),
    (8048, "Forgejo Software Forge"),
    (8049, "Drone CI Pipeline"),
    (8050, "Woodpecker CI Engine"),
]

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
<body style="font-family: sans-serif; background: #0f172a; color: #f8fafc; padding: 2rem;">
  <h1>⚡ {name}</h1>
  <p>Cổng dịch vụ ảo thử nghiệm: <strong>{port}</strong></p>
  <p>Trạng thái: <span style="color: #4ade80;">🟢 Online (Mock Service)</span></p>
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
    servers = []
    print(f"🚀 Đang khởi tạo {len(SERVICES)} cổng dịch vụ ảo (8001 - 8050)...", flush=True)
    for port, name in SERVICES:
        try:
            srv = await asyncio.start_server(make_handler(port, name), '0.0.0.0', port)
            servers.append(srv)
        except Exception as e:
            print(f"⚠️ Không thể bind cổng {port}: {e}", flush=True)
    
    print(f"✅ Đã kích hoạt thành công {len(servers)}/{len(SERVICES)} cổng ảo!", flush=True)
    print("💡 Bấm Ctrl+C để dừng hoặc chạy 'pkill -f mock_50_ports.py'", flush=True)

    await asyncio.gather(*(s.serve_forever() for s in servers))

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except (KeyboardInterrupt, SystemExit):
        print("🛑 Đã dừng toàn bộ cổng ảo.")
