# API Reference

Complete API reference for **echoip** - IP address lookup service.

## Base URL

```
http://your-server.com
```

---

## Endpoints

### Get Your IP Address

#### `GET /`

Returns your IP address in different formats based on the `Accept` header or User-Agent.

**Request**:
```bash
curl https://your-server.com/
```

**Response** (plain text):
```
203.0.113.42
```

**Request** (JSON):
```bash
curl -H "Accept: application/json" https://your-server.com/
```

**Response** (JSON):
```json
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

---

### Lookup Specific IP

#### `GET /{ip}`

Lookup information for a specific IP address (IPv4 or IPv6).

**IPv4 Example**:
```bash
curl https://your-server.com/8.8.8.8
```

**IPv6 Example**:
```bash
curl https://your-server.com/2001:4860:4860::8888
```

**Response**:
```json
{
  "ip": "8.8.8.8",
  "ip_decimal": 134744072,
  "country": "United States",
  "country_iso": "US",
  "asn": "AS15169",
  "asn_org": "Google LLC"
}
```

---

### Health Check

#### `GET /health`

Server health check endpoint.

**Request**:
```bash
curl https://your-server.com/health
```

**Response**:
```json
{
  "status": "OK"
}
```

---

### Get Specific Fields

#### `GET /ip`
Returns just your IP address (text).

```bash
curl https://your-server.com/ip
# Output: 203.0.113.42
```

#### `GET /country`
Returns your country name (text).

```bash
curl https://your-server.com/country
# Output: United States
```

#### `GET /country-iso`
Returns your country ISO code (text).

```bash
curl https://your-server.com/country-iso
# Output: US
```

#### `GET /city`
Returns your city name (text).

```bash
curl https://your-server.com/city
# Output: Mountain View
```

#### `GET /coordinates`
Returns your coordinates (text).

```bash
curl https://your-server.com/coordinates
# Output: 37.386,-122.0838
```

#### `GET /asn`
Returns your Autonomous System Number (text).

```bash
curl https://your-server.com/asn
# Output: AS15169
```

#### `GET /asn-org`
Returns your ASN organization (text).

```bash
curl https://your-server.com/asn-org
# Output: Google LLC
```

---

## API v1 Endpoints

RESTful API with versioned endpoints.

### `GET /api/v1`

Get your IP information (JSON).

**Request**:
```bash
curl https://your-server.com/api/v1
```

**Response**:
```json
{
  "ip": "203.0.113.42",
  "ip_decimal": 3405803306,
  "country": "United States",
  "country_iso": "US",
  ...
}
```

### `GET /api/v1/ip`

Get your IP address (text).

**Request**:
```bash
curl https://your-server.com/api/v1/ip
```

**Response**:
```
203.0.113.42
```

### `GET /api/v1/ip/{ip}`

Lookup specific IP address (JSON).

**Request**:
```bash
curl https://your-server.com/api/v1/ip/8.8.8.8
```

**Response**:
```json
{
  "ip": "8.8.8.8",
  "ip_decimal": 134744072,
  "country": "United States",
  ...
}
```

### `GET /api/v1/country`

Get your country (text).

### `GET /api/v1/city`

Get your city (text).

### `GET /api/v1/asn`

Get your ASN (text).

---

## Query Parameters

### `?ip={address}`

Lookup information for a specific IP address.

**Example**:
```bash
curl https://your-server.com/json?ip=8.8.8.8
```

---

## Port Testing

### `GET /port/{port}`

Test if a specific port is reachable on your IP address.

**Request**:
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

---

## Response Formats

### JSON

Use `Accept: application/json` header or access `/json` endpoint.

```bash
curl -H "Accept: application/json" https://your-server.com/
```

### Plain Text

Default for CLI tools (curl, wget, etc.).

```bash
curl https://your-server.com/
```

---

## IPv6 Support

### Force IPv4 or IPv6

Use your client's flags:

```bash
# Force IPv4
curl -4 https://your-server.com/

# Force IPv6
curl -6 https://your-server.com/
```

### IPv6 Lookups

```bash
# Lookup IPv6 address
curl https://your-server.com/2001:4860:4860::8888
```

---

## Rate Limiting

**Automated Use**: Please limit requests to **1 request per minute**.

Requests exceeding this limit may be rate-limited (HTTP 429) or dropped.

---

## Error Responses

### 400 Bad Request

```json
{
  "status": 400,
  "error": "Invalid IP address: invalid"
}
```

### 404 Not Found

```json
{
  "status": 404,
  "error": "Not found"
}
```

### 500 Internal Server Error

```json
{
  "status": 500,
  "error": "Internal server error"
}
```

---

## Examples

### cURL

```bash
curl https://your-server.com/
curl https://your-server.com/json
curl https://your-server.com/8.8.8.8
curl https://your-server.com/api/v1/ip/8.8.8.8
```

### HTTPie

```bash
http https://your-server.com/
http https://your-server.com/json
```

### wget

```bash
wget -qO- https://your-server.com/
```

### PowerShell

```powershell
Invoke-RestMethod https://your-server.com/json
```

---

## GeoIP Data

GeoIP data is provided by **GeoLite2** databases from [sapics/ip-location-db](https://github.com/sapics/ip-location-db).

- **Country**: ISO code, name, EU membership
- **City**: Name, region, postal code
- **Coordinates**: Latitude, longitude
- **Timezone**: IANA timezone identifier
- **ASN**: Autonomous System Number and organization

**Update Frequency**: Twice weekly (automatic)

---

## Technical Details

- **Version**: Check with `echoip -version`
- **Health**: Check with `/health` endpoint
- **Performance**: Sub-millisecond response times (with cache)
- **Capacity**: Handles thousands of requests per second
