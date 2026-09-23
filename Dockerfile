FROM python:3.12-slim

WORKDIR /app

# Cài đặt curl, iproute2, procps, ca-certificates và Tailscale CLI
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl iproute2 procps ca-certificates iptables && \
    curl -fsSL https://tailscale.com/install.sh | sh && \
    rm -rf /var/lib/apt/lists/*

COPY . /app

EXPOSE 65534

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD curl -f http://127.0.0.1:65534/api/status || exit 1

CMD ["python3", "server.py"]
