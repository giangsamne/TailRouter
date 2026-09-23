# Hướng Dẫn Chạy TailRouter Bằng Docker & Docker Compose

Nhánh `docker` cung cấp cấu hình tối ưu để triển khai **TailRouter** trên môi trường Docker, Unraid, TrueNAS SCALE, CasaOS, và Portainer.

---

## 🚀 1. Khởi Chạy Nhanh Bằng Docker Compose

```bash
# 1. Clone nhánh docker
git clone -b docker https://github.com/giangsamne/TailRouter.git tailrouter-docker
cd tailrouter-docker

# 2. Khởi chạy
docker compose up -d
```

Bảng quản trị sẽ sẵn sàng tại:
- **http://localhost:65534/router**
- **http://<TAILSCALE_IP>:65534/router**

---

## ⚙️ 2. Các Volume Quan Trọng
- `./config:/app/config`: Lưu danh sách các tuyến đường (routes) cố định.
- `/var/run/docker.sock:/var/run/docker.sock:ro`: Cho phép TailRouter tự động quét và nhận diện tên các container Docker đang chạy.
- `/var/run/tailscale/tailscaled.sock:/var/run/tailscale/tailscaled.sock:ro`: Cho phép TailRouter bên trong container giao tiếp với tiến trình Tailscale của máy chủ Host.
