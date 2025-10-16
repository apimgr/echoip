# echoip Documentation

Welcome to the **echoip** documentation! This service provides fast, accurate IP address lookups with GeoIP location data and full IPv4/IPv6 support.

## Quick Start

### Get Your IP Address

```bash
curl https://your-server.com/
# Output: 203.0.113.42
```

### Get IP Information (JSON)

```bash
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

### Lookup Specific IP

```bash
# IPv4
curl https://your-server.com/8.8.8.8

# IPv6
curl https://your-server.com/2001:4860:4860::8888
```

## Features

- **Fast & Lightweight**: Single static binary, <15MB
- **IPv6 Support**: Full dual-stack IPv4/IPv6
- **GeoIP Data**: Country, city, ASN, coordinates, timezone
- **Auto-Updates**: Weekly GeoIP database updates
- **RESTful API**: `/api/v1` endpoints
- **Multiple Formats**: JSON, plain text, or custom
- **Mobile-Friendly**: Responsive web interface
- **Port Testing**: Check port connectivity
- **Reverse DNS**: Optional hostname lookups
- **No API Key**: GeoIP databases auto-download

## Documentation Sections

- **[API Reference](API.md)** - Complete API documentation
- **[Server Guide](SERVER.md)** - Installation and administration
- **[GitHub Repository](https://github.com/apimgr/echoip)** - Source code

## Installation

### Docker (Recommended)

```bash
docker run -d -p 8080:80 ghcr.io/apimgr/echoip:latest
curl http://localhost:8080/
```

### Binary

```bash
curl -L -o echoip https://github.com/apimgr/echoip/releases/latest/download/echoip-linux-amd64
chmod +x echoip
./echoip -l :8080
```

### Build from Source

```bash
git clone https://github.com/apimgr/echoip.git
cd echoip
make build
./binaries/echoip -l :8080
```

## Support

- **Issues**: [GitHub Issues](https://github.com/apimgr/echoip/issues)
- **License**: MIT
- **Credits**: Based on [mpolden/echoip](https://github.com/mpolden/echoip)
