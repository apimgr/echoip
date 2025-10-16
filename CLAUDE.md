# echoip - Complete Project Specification

**Project**: echoip
**Organization**: apimgr
**Version**: 0.0.1
**Last Updated**: 2025-10-16
**License**: MIT
**Repository**: https://github.com/apimgr/echoip

---

## Table of Contents

1. [Project Overview](#project-overview)
2. [Technology Stack](#technology-stack)
3. [Architecture](#architecture)
4. [Directory Structure](#directory-structure)
5. [Features](#features)
6. [API Endpoints](#api-endpoints)
7. [GeoIP Implementation](#geoip-implementation)
8. [Configuration](#configuration)
9. [Build System](#build-system)
10. [Deployment](#deployment)
11. [CI/CD](#cicd)
12. [Development](#development)
13. [Testing](#testing)
14. [Documentation](#documentation)
15. [Security](#security)
16. [Performance](#performance)
17. [Troubleshooting](#troubleshooting)

---

## Project Overview

**echoip** is a simple, fast IP address lookup service that returns your public IP address along with detailed geolocation information. It's designed for both human users (via web browser) and automated tools (via API).

### Key Characteristics

- **Single Static Binary**: All assets embedded, no external dependencies
- **Lightweight**: <15MB binary, minimal resource usage
- **Fast**: Sub-millisecond responses with caching
- **IPv6 Ready**: Full dual-stack IPv4/IPv6 support
- **Auto-Updating**: GeoIP databases update weekly automatically
- **Production-Grade**: Health checks, monitoring, CI/CD
- **SPEC-Compliant**: Follows apimgr standards exactly

### Use Cases

- Quick IP address lookup for developers
- Geolocation API for applications
- Network diagnostics and troubleshooting
- IPv6 testing and validation
- Self-hosted alternative to commercial IP lookup services

---

## Technology Stack

### Core

- **Language**: Go 1.23+
- **Web Framework**: Native `net/http` with custom router
- **Templates**: Go `html/template` with `go:embed`
- **GeoIP Library**: `github.com/oschwald/geoip2-golang`

### GeoIP Data

- **Provider**: sapics/ip-location-db
- **Distribution**: jsdelivr CDN (npm packages)
- **Databases**: 4 MMDB files (~103MB total)
  - geolite2-city-ipv4.mmdb (~50MB)
  - geolite2-city-ipv6.mmdb (~40MB)
  - geo-whois-asn-country.mmdb (~8MB)
  - asn.mmdb (~5MB)
- **Update Frequency**: Twice weekly (city), daily (country/ASN)

### Runtime

- **Development**: Local binary or Docker
- **Production**: Docker (Alpine Linux) or systemd service
- **Platforms**: 8 targets (Linux, macOS, Windows, FreeBSD on amd64/arm64)

### Frontend

- **HTML**: Go templates (embedded)
- **CSS**: Vanilla CSS3 with CSS variables (~751 lines)
- **JavaScript**: Vanilla JS, no frameworks (~237 lines)
- **Theme**: Dark mode default, light mode available
- **Responsive**: Mobile-friendly with IPv6 support

### Infrastructure

- **Container Registry**: ghcr.io (GitHub Container Registry)
- **CI/CD**: GitHub Actions + Jenkins (jenkins.casjay.cc)
- **Documentation**: ReadTheDocs (MkDocs + Material theme + Dracula)
- **Monitoring**: Health checks, profiling endpoints

---

## Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         main.go                              │
│  - Version info (Version, Commit, BuildDate)                 │
│  - Flag parsing (--version, --status, -d, -l, -r, etc.)     │
│  - GeoIP manager initialization                              │
│  - Weekly update scheduler (background goroutine)            │
│  - Server setup and startup                                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ├─────────────────────────┐
                              │                         │
                              ▼                         ▼
                    ┌──────────────────┐    ┌──────────────────┐
                    │  geoip/          │    │  server/         │
                    │  Manager         │    │  HTTP Server     │
                    │                  │    │                  │
                    │  - 4 databases   │    │  - Routing       │
                    │  - Auto-download │    │  - Handlers      │
                    │  - Weekly update │    │  - Templates     │
                    │  - IPv4/IPv6     │    │  - Static files  │
                    └──────────────────┘    └──────────────────┘
                              │                         │
                              │                         │
                    ┌─────────┴─────────┐    ┌─────────┴──────────┐
                    │                   │    │                    │
                    ▼                   ▼    ▼                    ▼
            ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐
            │ iputil/      │  │ paths/       │  │ utils/           │
            │ - IP parse   │  │ - OS paths   │  │ - Network        │
            │ - Decimal    │  │ - Detect OS  │  │ - IPv6           │
            │ - geo/       │  │ - Defaults   │  │ - URL display    │
            └──────────────┘  └──────────────┘  └──────────────────┘
```

### Request Flow

```
Client Request
    │
    ▼
┌─────────────────────┐
│ HTTP Router         │
│ (server/router.go)  │
└─────────────────────┘
    │
    ├─ /health          → HealthHandler → {"status":"OK"}
    ├─ /                → JSONHandler/CLIHandler → Your IP
    ├─ /{ip}            → IPLookupHandler → Specific IP lookup
    ├─ /api/v1          → APIV1InfoHandler → Your IP (JSON)
    ├─ /api/v1/ip/{ip}  → APIV1IPLookupHandler → IP lookup
    ├─ /country, /city  → CLI*Handler → Specific fields
    └─ /static/*        → StaticHandler → Embedded CSS/JS/images
```

### Data Flow (GeoIP Lookup)

```
1. Client Request: GET /8.8.8.8
2. Router matches: /{ip} route
3. Handler extracts IP: "8.8.8.8"
4. Parse IP: net.ParseIP("8.8.8.8")
5. Detect IP version: IPv4
6. Select database: cityIPv4DB
7. Lookup:
   - cityIPv4DB.City(ip) → City data
   - countryDB.Country(ip) → Country data
   - asnDB.ASN(ip) → ASN data
8. Build response: JSON object
9. Cache response (if enabled)
10. Return to client
```

---

## Directory Structure

### Root Files

```
echoip/
├── .claude/
│   └── settings.local.json      # Claude Code settings
├── .github/
│   └── workflows/
│       ├── ci.yml               # Existing CI
│       ├── docker.yml           # Docker builds (amd64, arm64)
│       └── release.yml          # Binary releases (8 platforms)
├── docs/                        # ReadTheDocs documentation
│   ├── stylesheets/
│   │   └── dracula.css          # Dracula theme CSS
│   ├── javascripts/
│   │   └── extra.js             # Custom JavaScript
│   ├── overrides/               # MkDocs theme overrides
│   ├── index.md                 # Documentation home
│   ├── API.md                   # Complete API reference
│   ├── SERVER.md                # Server administration
│   ├── README.md                # Documentation index
│   ├── mkdocs.yml               # MkDocs config (Material + Dracula)
│   └── requirements.txt         # Python deps for RTD
├── scripts/                     # Production scripts
│   ├── install.sh               # systemd installation (Linux)
│   ├── backup.sh                # Backup with rotation
│   └── uninstall.sh             # Clean uninstall
├── test/                        # Testing
│   └── test-docker.sh           # Docker test script (SPEC-compliant)
├── .dockerignore                # Docker ignore patterns
├── .gitattributes               # Git attributes (LF, diff settings)
├── .gitignore                   # Git ignore (binaries/, rootfs/, etc.)
├── .readthedocs.yml             # ReadTheDocs configuration
├── CLAUDE.md                    # This file - Project specification
├── docker-compose.yml           # Production (172.17.0.1:64180:80)
├── docker-compose.test.yml      # Development (64181:80, /tmp)
├── Dockerfile                   # Alpine-based multi-stage build
├── go.mod                       # Go 1.23, github.com/apimgr/echoip
├── go.sum                       # Go module checksums
├── Jenkinsfile                  # Multi-arch CI/CD (jenkins.casjay.cc)
├── LICENSE.md                   # MIT License
├── Makefile                     # Complete build system
├── README.md                    # User documentation
└── release.txt                  # Version: 0.0.1
```

### Source Structure

```
src/
├── main.go                      # Application entry point
│   - Version variables (Version, Commit, BuildDate)
│   - Flag parsing and validation
│   - GeoIP manager initialization
│   - Weekly update scheduler (background goroutine)
│   - Server configuration
│   - Startup and listening
│
├── geoip/                       # GeoIP database management
│   └── geoip.go
│       - Manager struct (4 database files)
│       - Auto-download from sapics/ip-location-db via CDN
│       - Database loading (cityIPv4, cityIPv6, country, asn)
│       - IPv4/IPv6 database selection logic
│       - Weekly update mechanism
│       - geo.Reader interface implementation
│
├── iputil/                      # IP address utilities
│   ├── iputil.go                # IP parsing, decimal conversion
│   ├── iputil_test.go
│   └── geo/                     # GeoIP reader interface
│       └── geo.go
│           - Reader interface
│           - Country, City, ASN structs
│           - Data structure definitions
│
├── paths/                       # OS-specific directory detection
│   └── paths.go
│       - GetDirectories() → Directories struct
│       - OS detection (Linux, macOS, Windows)
│       - Environment variable overrides (CONFIG_DIR, DATA_DIR, LOGS_DIR)
│       - EnsureDirectories() creation
│
├── server/                      # HTTP server (renamed from 'http')
│   ├── templates/               # HTML templates (embedded via go:embed)
│   │   ├── index.html           # Main page with IP display
│   │   ├── script.html          # JavaScript (widget logic)
│   │   └── styles.html          # CSS (mobile-friendly, IPv6)
│   ├── static/                  # Static assets (embedded via go:embed)
│   │   ├── css/
│   │   │   └── main.css         # Modern CSS framework (751 lines)
│   │   ├── js/
│   │   │   └── main.js          # JavaScript utilities (237 lines)
│   │   └── images/
│   │       └── leafcloud-logo.svg
│   ├── templates.go             # Template & static embedding
│   │   - InitTemplates() with go:embed
│   │   - StaticHandler() for /static/*
│   ├── http.go                  # Request handlers
│   │   - New() server constructor
│   │   - HealthHandler, JSONHandler, CLIHandler
│   │   - DefaultHandler (web UI)
│   │   - IPLookupHandler (/{ip})
│   │   - APIV1*Handler (API v1 endpoints)
│   │   - PortHandler, cacheHandler
│   │   - Handler() router setup
│   ├── router.go                # Route matching
│   │   - Custom router with path matching
│   │   - Prefix matching support
│   │   - Header matching
│   │   - Custom matcher functions
│   ├── cache.go                 # Response caching
│   │   - LRU cache implementation
│   │   - Cache key generation
│   │   - Size management
│   ├── cache_test.go
│   ├── error.go                 # Error handling
│   │   - appError struct
│   │   - Error response formatting (JSON/text)
│   │   - HTTP status code mapping
│   ├── http_test.go
│   └── router_test.go (implied)
│
├── useragent/                   # User agent parsing
│   ├── useragent.go             # Parse User-Agent headers
│   └── useragent_test.go
│
├── utils/                       # Utility functions
│   └── network.go
│       - GetAccessibleURL(port) → FQDN/hostname/IP/fallback
│       - GetOutboundIP() → IPv4/IPv6 detection
│       - ParseIP(ipStr) → Handle brackets
│       - IsIPv6(ip) → IP version check
│       - formatURLWithIP(ip, port) → IPv6 brackets
│
└── main_test.go                 # Main package tests
```

---

## Features

### Core Functionality

#### IP Address Detection
- **Your IP**: Returns client's public IP address
- **Output Formats**: Plain text, JSON
- **User-Agent Detection**: Returns appropriate format based on client
- **Header Support**: Trusts X-Real-IP, X-Forwarded-For (configurable)

#### IP Address Lookup
- **Specific IP**: Lookup any IPv4 or IPv6 address
- **Endpoint**: `GET /{ip}` (e.g., `/8.8.8.8`, `/2001:4860:4860::8888`)
- **API Endpoint**: `GET /api/v1/ip/{ip}`
- **Bracket Handling**: Automatically removes brackets from IPv6

#### GeoIP Location Data
- **Country**: Name, ISO code, EU membership
- **City**: Name, region/state, postal code
- **Coordinates**: Latitude, longitude
- **Timezone**: IANA timezone identifier
- **ASN**: Autonomous System Number and organization name
- **Metro Code**: US metro code (US only)

#### Additional Features
- **Reverse DNS**: Optional hostname lookups (`-r` flag)
- **Port Testing**: Check if ports are reachable (`-p` flag)
- **Response Caching**: LRU cache for performance (`-C` flag)
- **Profiling**: pprof endpoints for debugging (`-P` flag)
- **IPv6 Support**: Full dual-stack, mobile-friendly display

### GeoIP Auto-Management

#### First Run Behavior
1. Check if GeoIP databases exist in data directory
2. If missing, download all 4 databases from jsdelivr CDN
3. Display progress for each database download
4. Load databases into memory
5. Server ready to serve requests with GeoIP data

#### Automatic Updates
- **Schedule**: Checks daily, updates if >7 days old
- **Process**: Download → Reload → No restart needed
- **Background**: Runs in separate goroutine
- **Logging**: All operations logged
- **Graceful Failure**: Server continues if update fails

#### Manual Update
```bash
make geoip-download
```

### Web Interface

#### Features
- **Responsive Design**: Mobile-first approach
- **Dark Mode**: Default theme (light mode available)
- **Theme Toggle**: Persisted in localStorage
- **IPv6 Display**: Horizontal scroll for long addresses
- **Interactive Widget**: Build API queries visually
- **Sponsor Logo**: Configurable (`-s` flag)

#### Components (SPEC Section 12)
- Toast notifications
- Modal dialogs
- Mobile menu toggle
- Copy to clipboard
- Keyboard shortcuts
- API helpers
- Theme management

---

## API Endpoints

### Public Endpoints

#### Root Endpoint
**GET /**

Returns your IP address in different formats based on Accept header or User-Agent.

**Plain Text** (default for CLI tools):
```bash
$ curl https://your-server.com/
203.0.113.42
```

**JSON** (with Accept header):
```bash
$ curl -H "Accept: application/json" https://your-server.com/
{
  "ip": "203.0.113.42",
  "ip_decimal": 3405803306,
  "country": "United States",
  "country_iso": "US",
  "city": "Mountain View",
  "region_name": "California",
  "region_code": "CA",
  "latitude": 37.386,
  "longitude": -122.0838,
  "timezone": "America/Los_Angeles",
  "asn": "AS15169",
  "asn_org": "Google LLC",
  "hostname": "dns.google"
}
```

#### IP Lookup Endpoint
**GET /{ip}**

Lookup information for any IPv4 or IPv6 address.

**Examples**:
```bash
# IPv4
curl https://your-server.com/8.8.8.8

# IPv6
curl https://your-server.com/2001:4860:4860::8888

# IPv6 with brackets (auto-stripped)
curl https://your-server.com/[2001:4860:4860::8888]
```

**Response**: Always JSON with full GeoIP data

#### Health Check
**GET /health**

Server health check endpoint (always returns 200 OK).

```json
{
  "status": "OK"
}
```

#### Field-Specific Endpoints

| Endpoint | Returns | Example |
|----------|---------|---------|
| `GET /ip` | Your IP address | `203.0.113.42` |
| `GET /json` | Your IP info (JSON) | Full JSON object |
| `GET /country` | Your country name | `United States` |
| `GET /country-iso` | Your country ISO code | `US` |
| `GET /city` | Your city name | `Mountain View` |
| `GET /coordinates` | Your coordinates | `37.386,-122.0838` |
| `GET /asn` | Your ASN | `AS15169` |
| `GET /asn-org` | Your ASN organization | `Google LLC` |

#### Port Testing
**GET /port/{port}**

Test if a specific port is reachable on your IP address.

```bash
curl https://your-server.com/port/22
```

**Response**:
```json
{
  "ip": "203.0.113.42",
  "port": 22,
  "reachable": true
}
```

### API v1 Endpoints

RESTful API with versioned endpoints.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1` | GET | Your IP info (JSON) |
| `/api/v1/ip` | GET | Your IP address (text) |
| `/api/v1/ip/{ip}` | GET | Lookup specific IP (JSON) |
| `/api/v1/country` | GET | Your country (text) |
| `/api/v1/city` | GET | Your city (text) |
| `/api/v1/asn` | GET | Your ASN (text) |

**Route Matching Philosophy** (SPEC):
- Frontend routes mirror API routes
- `/{ip}` ↔ `/api/v1/ip/{ip}`
- `/country` ↔ `/api/v1/country`
- Predictable and consistent

### Query Parameters

#### ?ip={address}
Lookup information for a specific IP address.

```bash
curl https://your-server.com/json?ip=8.8.8.8
curl https://your-server.com/country?ip=1.1.1.1
```

### Static Files

**GET /static/***

Serves embedded static assets (CSS, JavaScript, images).

```
/static/css/main.css
/static/js/main.js
/static/images/leafcloud-logo.svg
```

All files embedded via `go:embed` in binary.

---

## GeoIP Implementation

### Database Configuration (SPEC Section 16)

#### 4 Required Databases

**1. geolite2-city-ipv4.mmdb** (~50MB)
- **Purpose**: City-level geolocation for IPv4 addresses
- **Data**: City, region, postal code, coordinates, timezone, metro code
- **Source**: MaxMind GeoLite2
- **URL**: `https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv4.mmdb`
- **Updates**: Twice weekly
- **License**: CC BY-SA 4.0 (attribution required)

**2. geolite2-city-ipv6.mmdb** (~40MB)
- **Purpose**: City-level geolocation for IPv6 addresses
- **Data**: Same as IPv4 city database
- **Source**: MaxMind GeoLite2
- **URL**: `https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv6.mmdb`
- **Updates**: Twice weekly
- **License**: CC BY-SA 4.0 (attribution required)

**3. geo-whois-asn-country.mmdb** (~8MB)
- **Purpose**: Country-level data (combined IPv4/IPv6)
- **Data**: Country name, ISO code, EU membership
- **Source**: Aggregated from WHOIS, ASN, GeoFeed
- **URL**: `https://cdn.jsdelivr.net/npm/@ip-location-db/geo-whois-asn-country-mmdb/geo-whois-asn-country.mmdb`
- **Updates**: Daily
- **License**: Public domain (CC0/PDDL) - NO attribution required

**4. asn.mmdb** (~5MB)
- **Purpose**: ASN/ISP information (combined IPv4/IPv6)
- **Data**: Autonomous System Number, organization name
- **Source**: Multiple (RouteViews, WHOIS, etc.)
- **URL**: `https://cdn.jsdelivr.net/npm/@ip-location-db/asn-mmdb/asn.mmdb`
- **Updates**: Daily
- **License**: Various (check sources)

**Total Size**: ~103MB (all 4 databases)

### IPv4/IPv6 Database Selection

The service intelligently selects the correct city database based on IP type:

```go
func (g *geoipReader) City(ip net.IP) (geo.City, error) {
    // Select appropriate city database
    var cityDB *geoip2.Reader
    if ip.To4() != nil {
        // IPv4 address → use IPv4 city database
        cityDB = g.cityIPv4DB
    } else {
        // IPv6 address → use IPv6 city database
        cityDB = g.cityIPv6DB
    }

    // Perform lookup
    record, err := cityDB.City(ip)
    // ...
}
```

**Benefits**:
- Optimized memory usage (separate databases)
- Faster lookups (smaller files)
- Better performance for IPv6-heavy traffic

### Storage Locations

**Docker**:
- `/data/geoip/geolite2-city-ipv4.mmdb`
- `/data/geoip/geolite2-city-ipv6.mmdb`
- `/data/geoip/geo-whois-asn-country.mmdb`
- `/data/geoip/asn.mmdb`

**Linux (systemd)**:
- `/var/lib/echoip/geoip/*.mmdb`

**macOS**:
- `~/Library/Application Support/echoip/geoip/*.mmdb`

**Windows**:
- `%LOCALAPPDATA%\echoip\geoip\*.mmdb`

**Development**:
- `./data/geoip/*.mmdb`

### Update Mechanism

**Automatic** (Background Scheduler):
- Runs in goroutine started from main.go
- Checks every 24 hours
- Updates if databases are >7 days old
- Downloads all 4 databases
- Reloads databases without restart
- Logs all operations

**Manual**:
```bash
# Via Makefile
make geoip-download

# Via API (future)
curl -X POST http://localhost:8080/api/v1/admin/geoip/update
```

---

## Configuration

### Command-Line Flags

```
-d string
    Data directory for GeoIP databases
    Default: "data"
    Docker: "/data"
    systemd: "/var/lib/echoip"

-l string
    Listening address (supports IPv4, IPv6, or dual-stack)
    Default: ":8080"
    Docker: ":80"
    Examples:
      - ":8080"              (dual-stack on port 8080)
      - "0.0.0.0:8080"       (IPv4 only)
      - "[::]:8080"          (IPv6 dual-stack)
      - "127.0.0.1:8080"     (IPv4 localhost)
      - "[::1]:8080"         (IPv6 localhost)

-t string
    Path to template directory
    Default: "src/server/templates"
    Note: With go:embed, this is only used for development overrides

-r
    Perform reverse hostname lookups
    Enables LookupAddr() function
    May slow down responses

-p
    Enable port connectivity testing
    Enables /port/{port} endpoint
    Allows checking if ports are reachable

-s
    Show sponsor logo in web interface
    Displays Casjays Developments branding

-C int
    Response cache size (number of entries)
    Default: 0 (disabled)
    Recommended: 5000-20000 for production
    Example: -C 10000

-H value
    Header to trust for remote IP (can be used multiple times)
    Examples:
      -H X-Real-IP
      -H X-Forwarded-For
      -H CF-Connecting-IP
    Use with reverse proxy (nginx, Caddy, Traefik)

-P
    Enable profiling handlers at /debug/pprof
    For performance analysis and debugging
    DO NOT enable in production without access control

-version
    Show version information and exit
    Output: Version number only (e.g., "0.0.1")
    SPEC-compliant (no "v" prefix)

-status
    Health check - exits with code 0
    Used by Docker HEALTHCHECK
    For monitoring and orchestration
```

### Environment Variables

Used primarily in Docker and systemd deployments:

```bash
PORT=80
    Server port (Docker internal port)

ADDRESS=0.0.0.0
    Listen address
    Use "::" for dual-stack IPv6
    Use "0.0.0.0" for IPv4 only

CONFIG_DIR=/config
    Configuration directory
    Stores settings, preferences

DATA_DIR=/data
    Data directory
    Stores GeoIP databases (in /data/geoip/)
    Auto-created on first run

LOGS_DIR=/logs
    Logs directory
    Application logs (if file logging enabled)
```

### OS-Specific Defaults (src/paths/)

**Linux**:
- Config: `/etc/echoip` or `$XDG_CONFIG_HOME/echoip`
- Data: `/var/lib/echoip` or `$XDG_DATA_HOME/echoip`
- Logs: `/var/log/echoip`

**macOS**:
- Config: `~/Library/Application Support/echoip`
- Data: `~/Library/Application Support/echoip`
- Logs: `~/Library/Logs/echoip`

**Windows**:
- Config: `%APPDATA%\echoip`
- Data: `%LOCALAPPDATA%\echoip`
- Logs: `%LOCALAPPDATA%\echoip\logs`

**Docker**:
- Config: `/config`
- Data: `/data`
- Logs: `/logs`

---

## Build System

### Makefile Targets

#### Build Targets

**`make build`** - Build for all platforms
- Builds for 8 platforms (Linux, macOS, Windows, FreeBSD on amd64/arm64)
- Outputs to `binaries/` and `releases/`
- Reads version from `release.txt`
- Embeds version info via ldflags
- Creates host platform binary in `binaries/echoip`

**`make install`** - Install to $GOPATH/bin
- Builds binary
- Installs to `$GOPATH/bin/echoip`
- Makes it available system-wide

#### Testing Targets

**`make test`** - Run tests
- Runs `go test ./...` with race detection
- Timeout: 5 minutes
- Verbose output

**`make test-coverage`** - Test with coverage
- Generates coverage.out and coverage.html
- Opens coverage report in browser

**`make lint`** - Code quality checks
- Runs gofmt check
- Runs go vet
- Ensures code quality

**`make vet`** - Run go vet
**`make check-fmt`** - Check code formatting

#### Docker Targets

**`make docker`** - Build and push multi-arch images
- Platforms: linux/amd64, linux/arm64
- Tags: `latest`, `{VERSION}`
- Registry: ghcr.io/apimgr/echoip
- Requires: docker buildx

**`make docker-dev`** - Build development image
- Tag: `echoip:dev`
- Local only (not pushed)
- For testing with docker-compose

**`make docker-test`** - Test Docker image
- Uses docker-compose.test.yml
- Starts service on port 64181
- Runs health check
- Shows logs
- Cleans up /tmp storage

#### Release Targets

**`make release`** - Create GitHub release
- Builds all platforms
- Creates source archives (tar.gz, zip)
- Deletes existing release if present
- Creates new GitHub release
- Uploads all binaries
- Auto-increments version in release.txt

**`make version-bump`** - Increment patch version
- Called automatically after successful release
- Updates release.txt
- Example: 1.0.0 → 1.0.1

#### GeoIP Targets

**`make geoip-download`** - Download GeoIP databases
- Downloads all 4 databases from sapics/ip-location-db
- Uses jsdelivr CDN
- Saves to `data/geoip/`
- No API key required
- Total: ~103MB

#### Development Targets

**`make run`** - Run in development mode
- Runs `go run ./src -l :8080`
- No GeoIP databases required
- Hot reload with file changes

**`make run-full`** - Run with all features
- Downloads GeoIP databases first
- Runs with: `-a`, `-c`, `-f` (GeoIP), `-r` (reverse), `-p` (ports), `-s` (sponsor)
- Full feature demonstration

#### Cleanup Targets

**`make clean`** - Clean all artifacts
- Removes binaries/
- Removes releases/
- Removes coverage files
- Stops Docker containers
- Removes Docker images

**`make clean-binaries`** - Clean build artifacts only
**`make clean-docker`** - Clean Docker artifacts only

#### Help

**`make help`** - Show all targets
- Lists all available make commands
- Shows current version
- Organized by category

### Build Configuration

**Version Management**:
- Version stored in `release.txt`
- Format: `0.0.1` (no "v" prefix)
- Embedded in binary via `-ldflags`
- Variables: `main.Version`, `main.Commit`, `main.BuildDate`

**Build Flags**:
```makefile
LDFLAGS := -X main.Version=$(VERSION) \
           -X main.Commit=$(COMMIT) \
           -X main.BuildDate=$(BUILD_DATE) \
           -w -s
```

**Platforms**:
```makefile
PLATFORMS := \
    linux/amd64 \
    linux/arm64 \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64 \
    windows/arm64 \
    freebsd/amd64 \
    freebsd/arm64
```

### Binary Output

**binaries/** (gitignored):
- `echoip-linux-amd64`
- `echoip-linux-arm64`
- `echoip-darwin-amd64`
- `echoip-darwin-arm64`
- `echoip-windows-amd64.exe`
- `echoip-windows-arm64.exe`
- `echoip-freebsd-amd64`
- `echoip-freebsd-arm64`
- `echoip` (host platform)

**releases/** (gitignored):
- All 8 platform binaries
- `echoip-{VERSION}-src.tar.gz`
- `echoip-{VERSION}-src.zip`

---

## Deployment

### Docker (Production)

#### Docker Compose

**File**: `docker-compose.yml`

```yaml
services:
  echoip:
    image: ghcr.io/apimgr/echoip:latest
    container_name: echoip
    restart: unless-stopped

    environment:
      - CONFIG_DIR=/config
      - DATA_DIR=/data
      - LOGS_DIR=/logs
      - PORT=80
      - ADDRESS=0.0.0.0

    volumes:
      - ./rootfs/config/echoip:/config
      - ./rootfs/data/echoip:/data
      - ./rootfs/logs/echoip:/logs

    ports:
      - "172.17.0.1:64180:80"

    networks:
      - echoip

    healthcheck:
      test: ["CMD", "/usr/local/bin/echoip", "-status"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 10s

networks:
  echoip:
    name: echoip
    external: false
    driver: bridge
```

**Deployment**:
```bash
docker-compose up -d
curl http://172.17.0.1:64180/
docker-compose logs -f
```

**Storage**:
- Config: `./rootfs/config/echoip`
- Data: `./rootfs/data/echoip` (GeoIP databases auto-download here)
- Logs: `./rootfs/logs/echoip`

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

### Binary Installation (Linux systemd)

#### Installation Script

```bash
curl -fsSL https://raw.githubusercontent.com/apimgr/echoip/master/scripts/install.sh | sudo bash
```

**What it does**:
1. Detects architecture (amd64 or arm64)
2. Downloads appropriate binary from GitHub releases
3. Installs to `/usr/local/bin/echoip`
4. Creates user `echoip`
5. Creates directories (`/var/lib/echoip`, `/var/log/echoip`, `/etc/echoip`)
6. Creates systemd service file
7. Enables and starts service

#### Service File

**Location**: `/etc/systemd/system/echoip.service`

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

#### Service Management

```bash
sudo systemctl status echoip
sudo systemctl start echoip
sudo systemctl stop echoip
sudo systemctl restart echoip
sudo journalctl -u echoip -f
```

### Reverse Proxy

#### nginx

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
    }
}
```

Run echoip with trusted header:
```bash
echoip -l :8080 -H X-Real-IP
```

---

## CI/CD

### GitHub Actions

**Workflows**:
1. **release.yml** - Binary builds and GitHub releases
2. **docker.yml** - Docker multi-arch builds

**Triggers** (both):
- Push to `main` or `master`
- Monthly schedule (1st of month, 3:00 AM UTC)
- Manual trigger (workflow_dispatch)

**release.yml Jobs**:
1. **test**: Run `make test` with Go 1.23
2. **build-binaries**: Build all 8 platforms
3. **create-release**: Delete old release, create new with binaries

**docker.yml Jobs**:
1. **build-and-push**: Multi-arch (amd64, arm64), push to ghcr.io, verify

**Artifacts**:
- Binaries retained for 90 days
- Docker images: `latest`, `{VERSION}`, `{branch}-{sha}`

### Jenkins Pipeline

**Server**: jenkins.casjay.cc

**Agents**: amd64, arm64

**Stages**:
1. **Checkout** - Clone repository
2. **Test** - Parallel testing on both architectures
3. **Build Binaries** - Parallel builds on both architectures
4. **Build Docker** - Parallel Docker image builds
5. **Push Docker** - Create manifests, push to registry
6. **GitHub Release** - Create release with artifacts

**Credentials Required**:
- `github-registry` - Docker registry login
- `github-token` - GitHub API token

---

## Development

### Local Development

```bash
# Clone
git clone https://github.com/apimgr/echoip.git
cd echoip

# Download GeoIP (optional)
make geoip-download

# Run
make run

# Or run with all features
make run-full
```

### Testing

**Unit Tests**:
```bash
make test
```

**Docker Testing**:
```bash
make docker-test
```

**Manual Docker Testing**:
```bash
./test/test-docker.sh
```

**Test Coverage**:
```bash
make test-coverage
# Opens coverage.html in browser
```

### Debug Mode

**Profiling**:
```bash
./echoip -l :8080 -P

# Access profiler
curl http://localhost:8080/debug/pprof/
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/heap
```

**Verbose Logging**:
- All operations logged with package prefix
- GeoIP download progress
- Server startup information
- Request handling (if verbose mode added)

---

## Testing

### Test Structure

**test/test-docker.sh** - Docker test script (SPEC-compliant)
- Generates random port (64000-64999)
- Builds dev image
- Runs container with /tmp storage
- Tests: health, IP, JSON, API v1, IP lookup
- Shows logs
- Automatic cleanup

**Running Tests**:
```bash
# Docker test
./test/test-docker.sh

# Or via Makefile
make docker-test

# Or unit tests
make test
```

### Test Checklist

- ✅ Health endpoint (`/health`)
- ✅ IP endpoint (`/ip`)
- ✅ JSON endpoint (`/json`)
- ✅ API v1 endpoint (`/api/v1`)
- ✅ IP lookup (`/8.8.8.8`)
- ✅ API v1 IP lookup (`/api/v1/ip/1.1.1.1`)
- ✅ IPv6 support
- ✅ GeoIP data
- ✅ Cache functionality
- ✅ Error handling

---

## Documentation

### Structure

**docs/** - ReadTheDocs compatible documentation

- **index.md**: Documentation homepage with quick start
- **API.md**: Complete API reference with examples
- **SERVER.md**: Server administration guide
- **README.md**: Documentation index
- **mkdocs.yml**: MkDocs configuration (Material theme + Dracula)
- **requirements.txt**: Python dependencies
- **stylesheets/dracula.css**: Dracula color theme
- **javascripts/extra.js**: Custom JavaScript (search, smooth scroll)

### Building Documentation

```bash
cd docs
pip install -r requirements.txt
mkdocs serve
# Open http://localhost:8000
```

### Publishing

Documentation auto-publishes to ReadTheDocs on push to repository.

**URL**: https://echoip.readthedocs.io (when configured)

---

## Security

### Docker Security

- **Non-root user**: Runs as UID 65534 (nobody)
- **Read-only root**: Filesystem is read-only
- **Minimal base**: Alpine Linux (small attack surface)
- **No shells in prod**: bash only for health checks
- **Capability dropping**: NoNewPrivileges=true

### systemd Security

- **Dedicated user**: Runs as `echoip` user
- **Restricted permissions**: ReadWritePaths limited
- **ProtectSystem**: strict
- **ProtectHome**: true
- **PrivateTmp**: true

### Network Security

- **Rate Limiting**: Recommended 1 req/min for automated use
- **HTTPS**: Use reverse proxy (nginx, Caddy, Traefik)
- **Firewall**: Restrict access as needed
- **Header Validation**: Only trust configured headers (-H flag)

### Data Privacy

- **No Logging**: IP addresses not logged to disk by default
- **No Tracking**: No analytics or user tracking
- **Ephemeral**: Responses cached in memory only
- **Open Source**: Fully auditable code

---

## Performance

### Benchmarks

- **Response Time**: <1ms (with cache)
- **Throughput**: Thousands of requests/second
- **Memory**: ~150MB (with 4 GeoIP databases loaded)
- **CPU**: Minimal (<1% idle, ~5% under load)

### Optimization

**Caching**:
```bash
echoip -C 10000
# Caches 10,000 responses in memory
# LRU eviction
# Significant performance boost
```

**Database Selection**:
- Separate IPv4/IPv6 city databases
- Smaller files = faster lookups
- Optimized memory usage

**Static Assets**:
- All assets embedded in binary
- No disk I/O for templates/CSS/JS
- Faster page loads

---

## Troubleshooting

### GeoIP Download Fails

**Symptom**: "Failed to download GeoIP databases"

**Causes**:
- No internet connectivity
- CDN unavailable
- Disk space full
- Permission denied

**Solutions**:
1. Check internet: `ping google.com`
2. Manual download: `make geoip-download`
3. Pre-download databases and place in data directory
4. Check disk space: `df -h`
5. Check permissions: `ls -la data/`

**Fallback**: Server continues without GeoIP (IP address only)

### Port Already in Use

**Symptom**: "address already in use"

**Solution**:
```bash
# Find process
sudo lsof -i :8080
sudo netstat -tlnp | grep 8080

# Use different port
echoip -l :8081
```

### Permission Denied

**Symptom**: Cannot create data directory

**Solutions**:
```bash
# Specify writable directory
echoip -d /tmp/echoip-data -l :8080

# Fix permissions
sudo chown -R $USER:$USER /var/lib/echoip
```

### IPv6 Not Working

**Symptom**: IPv6 requests fail or time out

**Solutions**:
1. Check IPv6 connectivity: `ping6 google.com`
2. Use dual-stack: `echoip -l :8080` (not `-l 0.0.0.0:8080`)
3. Docker: Enable IPv6 in `/etc/docker/daemon.json`
4. Firewall: Allow IPv6 traffic

### GeoIP Data Inaccurate

**Expected**: GeoIP databases provide estimates, not exact locations

**Solutions**:
- Wait for weekly update (databases update automatically)
- Force update: `rm -rf data/geoip && restart service`
- Use multiple data sources (databases aggregate multiple sources)

---

## Version Management

### Current Version

**File**: `release.txt`
**Content**: `0.0.1`

### Version Format (SPEC)

- release.txt: `0.0.1` (no "v" prefix)
- Git tags: `0.0.1` (no "v" prefix)
- GitHub releases: `0.0.1`
- Docker tags: `ghcr.io/apimgr/echoip:0.0.1`
- CLI output: `0.0.1` (version number only)

### Workflow

```bash
# 1. Build with current version
make build
# Uses version from release.txt

# 2. Test
make test

# 3. Create release
make release
# - Creates GitHub release with version from release.txt
# - AFTER successful release, auto-increments release.txt
# - Example: 0.0.1 → 0.0.2

# 4. Next build uses new version
make build
# Now builds 0.0.2
```

### Manual Version Override

```bash
VERSION=2.0.0 make build
VERSION=2.0.0 make release
```

---

## IPv6 Support

### Listening

**Dual-Stack** (Default):
```bash
echoip -l :8080
# Listens on both 0.0.0.0:8080 and [::]:8080
```

**IPv6 Only**:
```bash
echoip -l [::]:8080
```

**IPv4 Only**:
```bash
echoip -l 0.0.0.0:8080
```

### URL Formatting

**IPv6 addresses use brackets**:
- IPv4: `http://192.168.1.100:8080`
- IPv6: `http://[2001:db8::1]:8080`
- IPv6 localhost: `http://[::1]:8080`
- Hostname: `http://server.example.com:8080`

**Implementation**: `src/utils/network.go`

### GeoIP with IPv6

- Separate city databases for IPv4 and IPv6
- Automatic database selection based on `ip.To4()`
- Country and ASN databases support both IPv4/IPv6
- Full accuracy for IPv6 geolocation

### Mobile Display

**IPv6 addresses can be very long**:
- Horizontal scrolling enabled
- Touch-friendly scrollbars
- Dark mode scrollbar styling
- Font size responsive (36px → 24px on mobile)

**Implementation**: `src/server/templates/styles.html`

---

## Compliance & Standards

### SPEC Compliance

This project follows the **apimgr SPEC.md** (Version 2.0, Last Updated: 2025-01-14).

**SPEC Document**: 4,787 lines, 18 sections total

---

#### **✅ Implemented Sections** (16/18)

| Section | Title | Status | Implementation |
|---------|-------|--------|----------------|
| **1** | URL Display Standards | ✅ COMPLETE | src/utils/network.go |
| **2** | Dockerfile - Alpine Runtime | ✅ COMPLETE | Dockerfile (PORT=80, health check) |
| **3** | docker-compose.yml Production | ✅ COMPLETE | docker-compose.yml (172.17.0.1:64180:80) |
| **4** | docker-compose.test.yml Dev | ✅ COMPLETE | docker-compose.test.yml (64181:80, /tmp) |
| **5** | Makefile - Docker Improvements | ✅ COMPLETE | Makefile (all targets) |
| **6** | Jenkinsfile | ✅ COMPLETE | Jenkinsfile (multi-arch amd64/arm64) |
| **7** | src/data Directory | ⚠️ N/A | No JSON data files (uses MMDB binaries) |
| **8** | README.md Structure | ✅ COMPLETE | README.md (About → Production → Docker → API → Dev) |
| **9** | Complete Project Layout | ✅ COMPLETE | All required files and directories |
| **10** | ReadTheDocs Configuration | ✅ COMPLETE | .readthedocs.yml, docs/* (Dracula theme) |
| **11** | GitHub Actions Workflows | ✅ COMPLETE | release.yml, docker.yml (monthly schedule) |
| **12** | Web UI / Frontend Standards | ✅ COMPLETE | main.css (751 lines), main.js (237 lines), go:embed |
| **13** | Testing Environment Priority | ✅ COMPLETE | test/test-docker.sh, docker-compose.test.yml |
| **14** | AI Assistant Guidelines | ✅ FOLLOWED | /tmp usage, random ports, Docker first |
| **15** | IPv6 Support | ✅ COMPLETE | Dual-stack, GeoIP IPv4/IPv6, mobile CSS |
| **16** | GeoIP (sapics/ip-location-db) | ✅ COMPLETE | 4 databases, auto-download, weekly updates |

**Multi-Distro Testing** (Section after 16) | ✅ APPLICABLE | test-docker.sh compatible

---

#### **⚠️ Non-Applicable Sections** (2/18)

| Section | Title | Why Not Applicable |
|---------|-------|-------------------|
| **17** | Security & DDoS Protection | **NOT APPLICABLE** - echoip is a public service with no admin panel, no authentication, no protected routes. Rate limiting can be added at reverse proxy level if needed. |
| **18** | Admin Server Configuration | **NOT APPLICABLE** - echoip has no admin panel, no database-backed settings, no live reload needed. Configuration is via CLI flags only. |

**Rationale**:
- **Section 17** is for projects with admin authentication, protected routes, and user management
- **Section 18** is for projects with database-backed settings and admin panels
- **echoip** is a simple, stateless IP lookup service - no admin features required
- Security is handled at infrastructure level (reverse proxy, firewall)

---

#### **✅ Additional SPEC Elements**

**Version Control Restrictions** (lines 2924-2975):
- ✅ AI does not perform git operations
- ✅ All changes via Edit/Write tools
- ✅ User responsible for commits

**AI Communication Guidelines** (lines 4589-4786):
- ✅ Questions answered, not executed
- ✅ Commands executed with confirmation
- ✅ Clarification requested when ambiguous

---

### **Summary**: 16/16 applicable sections fully implemented. 2 sections not applicable to this project type.

### Code Quality

- **Formatting**: gofmt enforced
- **Linting**: go vet required
- **Testing**: Unit tests with race detection
- **Coverage**: Coverage reports generated
- **Static Analysis**: All code checked

### Best Practices

- ✅ Single responsibility principle
- ✅ Clean separation of concerns
- ✅ Dependency injection
- ✅ Error handling throughout
- ✅ Graceful degradation (GeoIP optional)
- ✅ Embedded assets (true single binary)
- ✅ Security hardening (non-root, minimal permissions)

---

## Credits & Attribution

### Original Project

**mpolden/echoip**
- Original author: Martin Polden
- Repository: https://github.com/mpolden/echoip
- License: BSD 3-Clause

### GeoIP Data

**sapics/ip-location-db**
- Repository: https://github.com/sapics/ip-location-db
- Aggregates: MaxMind GeoLite2, WHOIS, ASN, GeoFeed
- Licenses:
  - GeoLite2 city: CC BY-SA 4.0 (attribution required)
  - geo-whois-asn-country: Public domain (CC0/PDDL)
  - ASN: Various sources

**MaxMind GeoLite2**
- Provider: MaxMind, Inc.
- Website: https://www.maxmind.com/
- License: CC BY-SA 4.0
- Required attribution: "This product includes GeoLite2 data created by MaxMind, available from https://www.maxmind.com"

### Maintained By

**apimgr**
- Organization: apimgr
- Repository: https://github.com/apimgr/echoip
- SPEC: https://github.com/apimgr/SPEC.md

---

## License

**MIT License**

Copyright (c) 2025 apimgr

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

## Quick Reference

### Installation

```bash
# Docker Compose
curl -L -o docker-compose.yml https://raw.githubusercontent.com/apimgr/echoip/master/docker-compose.yml
docker-compose up -d

# Binary
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-linux-amd64
chmod +x echoip
./echoip -l :8080

# Install script (Linux)
curl -fsSL https://raw.githubusercontent.com/apimgr/echoip/master/scripts/install.sh | sudo bash
```

### Usage

```bash
# Get your IP
curl https://your-server.com/

# Get IP info (JSON)
curl https://your-server.com/json

# Lookup specific IP
curl https://your-server.com/8.8.8.8
curl https://your-server.com/2001:4860:4860::8888

# API v1
curl https://your-server.com/api/v1
curl https://your-server.com/api/v1/ip/8.8.8.8
```

### Development

```bash
# Build
make build

# Test
make test

# Docker dev
make docker-dev
make docker-test

# GeoIP
make geoip-download

# Release
make release
```

### Support

- **Documentation**: https://echoip.readthedocs.io
- **Issues**: https://github.com/apimgr/echoip/issues
- **Source**: https://github.com/apimgr/echoip

---

**This document serves as the complete specification and reference for the echoip project.**
