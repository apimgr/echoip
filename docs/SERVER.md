# Server Administration Guide

Complete guide for installing, configuring, and administering **echoip**.

---

## Installation

### Binary Installation (Linux)

#### Download

```bash
# AMD64
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-linux-amd64

# ARM64
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-linux-arm64

# Make executable
chmod +x echoip
sudo mv echoip /usr/local/bin/
```

#### Verify Installation

```bash
echoip -version
# Output: 0.0.1

echoip -status
# Exit code: 0
```

---

### Systemd Service

#### Create User

```bash
sudo useradd -r -s /bin/false echoip
```

#### Create Directories

```bash
sudo mkdir -p /var/lib/echoip /var/log/echoip /etc/echoip
sudo chown echoip:echoip /var/lib/echoip /var/log/echoip /etc/echoip
```

#### Service File

Create `/etc/systemd/system/echoip.service`:

```ini
[Unit]
Description=echoip - IP address lookup service
After=network.target
Documentation=https://github.com/apimgr/echoip

[Service]
Type=simple
User=echoip
Group=echoip
ExecStart=/usr/local/bin/echoip -l :8080 -d /var/lib/echoip -r -s
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/echoip /var/log/echoip

# Environment
Environment="CONFIG_DIR=/etc/echoip"
Environment="DATA_DIR=/var/lib/echoip"
Environment="LOGS_DIR=/var/log/echoip"

[Install]
WantedBy=multi-user.target
```

#### Enable and Start

```bash
sudo systemctl daemon-reload
sudo systemctl enable echoip
sudo systemctl start echoip
sudo systemctl status echoip
```

---

### Docker Installation

#### Docker Compose (Production)

Download `docker-compose.yml`:

```bash
curl -L -o docker-compose.yml https://raw.githubusercontent.com/apimgr/echoip/master/docker-compose.yml
```

Start service:

```bash
docker-compose up -d
```

Access:

```bash
curl http://172.17.0.1:64180/
```

View logs:

```bash
docker-compose logs -f
```

Stop:

```bash
docker-compose down
```

#### Docker Run

```bash
docker run -d \
  --name echoip \
  -p 172.17.0.1:64180:80 \
  -v ./data:/data \
  -e DATA_DIR=/data \
  --restart unless-stopped \
  ghcr.io/apimgr/echoip:latest
```

---

## Configuration

### Command-Line Flags

```
-d string
    Data directory for GeoIP databases (default "data")

-l string
    Listening address - supports IPv4, IPv6, or dual-stack (default ":8080")
    Examples: ":8080", "0.0.0.0:8080", "[::]:8080", "192.168.1.100:8080"

-t string
    Path to template directory (default "src/server/templates")

-r
    Perform reverse hostname lookups

-p
    Enable port connectivity testing

-s
    Show sponsor logo in web interface

-C int
    Response cache size in entries (0 to disable caching)

-H value
    Trust remote IP from header (can be specified multiple times)
    Examples: -H X-Real-IP -H X-Forwarded-For

-P
    Enable profiling handlers at /debug/pprof

-version
    Show version information and exit

-status
    Health check - exits with code 0 (for Docker healthchecks)
```

### Environment Variables

Docker and systemd deployments support environment variables:

```bash
PORT=80                    # Server port
ADDRESS=0.0.0.0            # Listen address (:: for dual-stack IPv6)
CONFIG_DIR=/config         # Configuration directory
DATA_DIR=/data             # Data directory (GeoIP databases)
LOGS_DIR=/logs             # Logs directory
```

---

## GeoIP Databases

### Automatic Download

GeoIP databases auto-download from [sapics/ip-location-db](https://github.com/sapics/ip-location-db) on first run.

**Databases**:
- Country: ~9.4 MB (IPv4 + IPv6)
- City: ~26 MB (IPv4)
- ASN: ~11 MB (IPv4 + IPv6)

**First Run**:
```bash
./echoip -d data -l :8080

# Output:
# echoip 0.0.1 (commit: abc123, built: 2025-10-16)
# IPv6 support enabled
# GeoIP databases not found, downloading...
# Downloading GeoLite2-country from sapics/ip-location-db...
# Downloading GeoLite2-city from sapics/ip-location-db...
# Downloading GeoLite2-asn from sapics/ip-location-db...
# GeoIP databases loaded successfully
```

### Manual Download

```bash
make geoip-download
```

### Weekly Updates

The server automatically checks for database updates:
- **Check interval**: Every 24 hours
- **Update threshold**: 7 days
- **Process**: Downloads latest → Reloads databases → No restart needed

---

## IPv6 Configuration

### Dual-Stack (Recommended)

Listen on both IPv4 and IPv6:

```bash
echoip -l :8080
# Automatically binds to both 0.0.0.0:8080 and [::]:8080
```

### IPv6 Only

```bash
echoip -l [::]:8080
```

### IPv4 Only

```bash
echoip -l 0.0.0.0:8080
```

### Docker IPv6

Enable IPv6 in `/etc/docker/daemon.json`:

```json
{
  "ipv6": true,
  "fixed-cidr-v6": "2001:db8:1::/64"
}
```

Restart Docker:

```bash
sudo systemctl restart docker
```

Update `docker-compose.yml`:

```yaml
environment:
  - ADDRESS=::    # Dual-stack

networks:
  echoip:
    enable_ipv6: true
    ipam:
      config:
        - subnet: 172.18.0.0/16
        - subnet: 2001:db8:1::/64
```

---

## Reverse Proxy Configuration

### nginx

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name ifconfig.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Run echoip with trusted header:

```bash
echoip -l :8080 -H X-Real-IP
```

### Caddy

```
ifconfig.example.com {
    reverse_proxy localhost:8080
}
```

### Traefik

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.echoip.rule=Host(`ifconfig.example.com`)"
  - "traefik.http.services.echoip.loadbalancer.server.port=8080"
```

---

## Performance Tuning

### Response Caching

Enable caching for better performance:

```bash
echoip -l :8080 -C 10000
# Caches up to 10,000 responses
```

**Cache Strategy**:
- Keys: IP address + endpoint
- TTL: Until cache full (LRU eviction)
- Recommended size: 5000-20000 entries

### Disable Features

For minimal resource usage:

```bash
echoip -l :8080
# No reverse lookup (-r)
# No port testing (-p)
# No caching (default)
```

---

## Monitoring

### Health Checks

```bash
# HTTP health check
curl http://localhost:8080/health

# Binary health check (Docker)
echoip -status
# Exit code: 0 (healthy)
```

### Logs

**Systemd**:
```bash
sudo journalctl -u echoip -f
```

**Docker**:
```bash
docker-compose logs -f
```

### Metrics

Enable profiling:

```bash
echoip -l :8080 -P
```

Access pprof:

```bash
# CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 > cpu.prof

# Heap profile
curl http://localhost:8080/debug/pprof/heap > heap.prof

# View in browser
go tool pprof -http=:8081 cpu.prof
```

---

## Troubleshooting

### GeoIP Download Fails

**Symptom**: "Failed to download GeoIP databases"

**Solutions**:
1. Check internet connectivity
2. Manual download: `make geoip-download`
3. Use pre-downloaded databases: Place in `data/` directory

**Fallback**: Server continues without GeoIP (returns IP only)

### Port Already in Use

**Symptom**: "address already in use"

**Solution**:
```bash
# Find process using port
sudo lsof -i :8080

# Use different port
echoip -l :8081
```

### Permission Denied

**Symptom**: Cannot create data directory

**Solution**:
```bash
# Specify writable directory
echoip -d /tmp/echoip-data -l :8080
```

### IPv6 Not Working

**Symptom**: IPv6 requests fail

**Solutions**:
1. Check IPv6 connectivity: `ping6 google.com`
2. Use dual-stack: `echoip -l :8080` (not `-l 0.0.0.0:8080`)
3. Docker: Enable IPv6 in daemon.json

---

## Backup & Restore

### Backup

```bash
# Backup GeoIP databases
tar -czf echoip-backup.tar.gz data/

# Backup with configuration
tar -czf echoip-full-backup.tar.gz \
  /etc/echoip \
  /var/lib/echoip \
  /var/log/echoip
```

### Restore

```bash
# Restore GeoIP databases
tar -xzf echoip-backup.tar.gz

# Restart service
sudo systemctl restart echoip
```

---

## Upgrading

### Binary Upgrade

```bash
# Stop service
sudo systemctl stop echoip

# Download new version
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-linux-amd64
chmod +x echoip
sudo mv echoip /usr/local/bin/

# Start service
sudo systemctl start echoip
```

### Docker Upgrade

```bash
# Pull latest image
docker-compose pull

# Recreate container
docker-compose up -d

# Or with downtime
docker-compose down
docker-compose up -d
```

---

## Security Considerations

1. **Run as non-root** - systemd service uses dedicated user
2. **Docker non-root** - Container runs as UID 65534
3. **Rate limiting** - Implement at reverse proxy level
4. **HTTPS** - Use reverse proxy (nginx, Caddy, Traefik)
5. **Firewall** - Restrict access to management ports
6. **Updates** - Keep binary and Docker images updated

---

## Support

- **Documentation**: [https://echoip.readthedocs.io](https://echoip.readthedocs.io)
- **Issues**: [GitHub Issues](https://github.com/apimgr/echoip/issues)
- **Source**: [GitHub Repository](https://github.com/apimgr/echoip)
