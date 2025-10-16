# echoip

![Build Status](https://github.com/apimgr/echoip/workflows/Docker%20Build%20and%20Push/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

A simple, fast IP address lookup service with full IPv4/IPv6 support and GeoIP location data.

## About

**echoip** is a lightweight IP address lookup service that provides:

- **Instant IP Detection** - Returns your public IP address
- **GeoIP Location Data** - Country, city, coordinates, timezone, ASN
- **Full IPv6 Support** - Dual-stack IPv4 and IPv6
- **RESTful API** - Versioned API at `/api/v1`
- **Multiple Output Formats** - JSON, plain text, or custom
- **Auto-Updating GeoIP** - Weekly automatic database updates
- **Mobile-Friendly** - Responsive web interface
- **Fast & Lightweight** - Single static binary <15MB
- **No API Keys Required** - GeoIP databases auto-download

## Production Installation

### Binary Installation

#### Linux (systemd)

```bash
# Download binary (AMD64)
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-linux-amd64
chmod +x echoip
sudo mv echoip /usr/local/bin/

# Or use installation script
curl -fsSL https://raw.githubusercontent.com/apimgr/echoip/master/scripts/install.sh | sudo bash
```

The installation script will:
- Download the appropriate binary for your system
- Create systemd service
- Create dedicated user and directories
- Start the service automatically

**Service Management**:
```bash
sudo systemctl status echoip
sudo systemctl start echoip
sudo systemctl stop echoip
sudo systemctl restart echoip
sudo journalctl -u echoip -f
```

**Configuration**: Edit `/etc/systemd/system/echoip.service`

#### macOS

```bash
# Download binary (ARM64 for Apple Silicon, AMD64 for Intel)
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-darwin-arm64
chmod +x echoip
sudo mv echoip /usr/local/bin/

# Run
echoip -l :8080
```

#### Windows

```powershell
# Download binary
Invoke-WebRequest -Uri "https://github.com/apimgr/echoip/releases/latest/download/echoip-windows-amd64.exe" -OutFile "echoip.exe"

# Run
.\echoip.exe -l :8080
```

### Environment Variables

```bash
# Linux/macOS
export DATA_DIR=/var/lib/echoip
export CONFIG_DIR=/etc/echoip
export LOGS_DIR=/var/log/echoip

# Windows
set DATA_DIR=C:\ProgramData\echoip
set CONFIG_DIR=C:\ProgramData\echoip\config
```

---

## Docker Deployment

### Docker Compose (Production)

**Download**:
```bash
curl -L -o docker-compose.yml https://raw.githubusercontent.com/apimgr/echoip/master/docker-compose.yml
```

**Start**:
```bash
docker-compose up -d
```

**Access**:
```bash
curl http://172.17.0.1:64180/
```

**Configuration**: `docker-compose.yml`
```yaml
services:
  echoip:
    image: ghcr.io/apimgr/echoip:latest
    ports:
      - "172.17.0.1:64180:80"
    volumes:
      - ./rootfs/data/echoip:/data
    restart: unless-stopped
```

### Docker Compose (Development)

```bash
# Build dev image
make docker-dev

# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Access
curl http://localhost:64181/

# Cleanup
docker-compose -f docker-compose.test.yml down
rm -rf /tmp/echoip/rootfs
```

### Docker Run

```bash
docker run -d \
  --name echoip \
  -p 8080:80 \
  -v ./data:/data \
  -e DATA_DIR=/data \
  --restart unless-stopped \
  ghcr.io/apimgr/echoip:latest
```

---

## API Usage

### Quick Examples

#### Get Your IP

```bash
# Plain text
curl https://your-server.com/
# Output: 203.0.113.42

# JSON
curl https://your-server.com/json
```

```json
{
  "ip": "203.0.113.42",
  "ip_decimal": 3405803306,
  "country": "United States",
  "country_iso": "US",
  "city": "Mountain View",
  "latitude": 37.386,
  "longitude": -122.0838,
  "timezone": "America/Los_Angeles",
  "asn": "AS15169",
  "asn_org": "Google LLC"
}
```

#### Lookup Specific IP

```bash
# IPv4
curl https://your-server.com/8.8.8.8

# IPv6
curl https://your-server.com/2001:4860:4860::8888
```

#### Get Specific Fields

```bash
curl https://your-server.com/ip          # Your IP address
curl https://your-server.com/country     # Your country
curl https://your-server.com/city        # Your city
curl https://your-server.com/asn         # Your ASN
```

### API v1 Endpoints

```bash
# Your IP info (JSON)
curl https://your-server.com/api/v1

# Lookup specific IP
curl https://your-server.com/api/v1/ip/8.8.8.8

# Get specific fields
curl https://your-server.com/api/v1/country
curl https://your-server.com/api/v1/city
curl https://your-server.com/api/v1/asn
```

### IPv6 Support

```bash
# Force IPv4
curl -4 https://your-server.com/

# Force IPv6
curl -6 https://your-server.com/

# Lookup IPv6 address
curl https://your-server.com/2001:4860:4860::8888
```

### Rate Limiting

For automated use, please limit requests to **1 request per minute**.

---

## Development

### Requirements

- **Go**: 1.23 or higher
- **Make**: GNU Make
- **Docker**: For containerized testing (recommended)
- **Git**: For version control

### Build System & Testing

#### Makefile Targets

```bash
# Build for all platforms (8 targets)
make build
# Outputs to: binaries/ and releases/

# Run tests
make test

# Code quality
make lint

# Build Docker images
make docker         # Multi-arch production images
make docker-dev     # Local development image
make docker-test    # Test with docker-compose

# Download GeoIP databases
make geoip-download

# Create GitHub release
make release        # Auto-increments version

# Development
make run            # Run locally
make run-full       # Run with all features

# Cleanup
make clean          # Clean all artifacts
make clean-binaries # Clean binaries only
make clean-docker   # Clean Docker artifacts

# Help
make help           # Show all targets
```

#### Platform Support

Builds for 8 platforms:
- Linux: amd64, arm64
- macOS: amd64, arm64
- Windows: amd64, arm64
- FreeBSD: amd64, arm64

#### Versioning

Version managed in `release.txt`:
```bash
cat release.txt
# Output: 0.0.1

# Version embedded in binary
./echoip -version
# Output: 0.0.1
```

### Development Mode

#### Local Development

```bash
# Clone repository
git clone https://github.com/apimgr/echoip.git
cd echoip

# Download GeoIP databases (optional)
make geoip-download

# Run in development mode
make run

# Run with all features
make run-full
```

#### Development Flags

```bash
# Minimal (no GeoIP)
./echoip -l :8080

# With GeoIP
./echoip -l :8080 -d data

# With all features
./echoip -l :8080 -d data -r -p -s

# With custom directories
./echoip \
  -l :8080 \
  -d /tmp/echoip/data \
  -t src/server/templates
```

#### Debug Features

```bash
# Enable profiling
./echoip -l :8080 -P

# Access profiler
curl http://localhost:8080/debug/pprof/
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/heap
```

### CI/CD

#### GitHub Actions

- **Triggers**: Push to main/master, monthly schedule, manual
- **Workflows**:
  - `release.yml` - Binary builds for 8 platforms
  - `docker.yml` - Multi-arch Docker images (amd64, arm64)
- **Schedule**: 1st of month, 3:00 AM UTC

#### Jenkins Pipeline

- **Server**: jenkins.casjay.cc
- **Agents**: amd64, arm64
- **Stages**: Test → Build → Docker → Push → Release
- **Artifacts**: Binaries, Docker images, GitHub releases

### Testing

```bash
# Run Go tests
make test

# Test Docker image
make docker-test

# Run test script
./test/test-docker.sh
```

---

## License & Credits

### License

MIT License - See [LICENSE.md](LICENSE.md)

### Credits

- **Original**: [mpolden/echoip](https://github.com/mpolden/echoip)
- **GeoIP Data**: [sapics/ip-location-db](https://github.com/sapics/ip-location-db)
  - GeoLite2 databases from MaxMind
  - Public domain country data
- **Maintained by**: [apimgr](https://github.com/apimgr)

### Attribution

GeoIP data from:
- MaxMind GeoLite2 (https://www.maxmind.com/) - CC BY-SA 4.0
- Aggregated via sapics/ip-location-db
- Country data: Public domain (WHOIS aggregation)
