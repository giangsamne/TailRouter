FROM python:3.12-slim

WORKDIR /app

# Install curl, iproute2 for network inspection
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl iproute2 procps && \
    rm -rf /var/lib/apt/lists/*

COPY . /app

EXPOSE 65534

CMD ["python3", "server.py"]
