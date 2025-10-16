package geoip

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/apimgr/echoip/src/iputil/geo"
	geoip2 "github.com/oschwald/geoip2-golang"
)

const (
	updateInterval = 7 * 24 * time.Hour // Weekly updates
)

// CDN URLs for databases (sapics/ip-location-db via jsdelivr)
const (
	cityIPv4URL  = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv4.mmdb"
	cityIPv6URL  = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-city-mmdb/geolite2-city-ipv6.mmdb"
	countryURL   = "https://cdn.jsdelivr.net/npm/@ip-location-db/geo-whois-asn-country-mmdb/geo-whois-asn-country.mmdb"
	asnURL       = "https://cdn.jsdelivr.net/npm/@ip-location-db/asn-mmdb/asn.mmdb"
)

// Manager handles GeoIP database management
type Manager struct {
	dataDir        string
	cityIPv4File   string
	cityIPv6File   string
	countryFile    string
	asnFile        string
	cityIPv4DB     *geoip2.Reader
	cityIPv6DB     *geoip2.Reader
	countryDB      *geoip2.Reader
	asnDB          *geoip2.Reader
	lastUpdate     time.Time
	updateInterval time.Duration
}

// NewManager creates a new GeoIP manager
func NewManager(dataDir string) *Manager {
	geoipDir := filepath.Join(dataDir, "geoip")
	return &Manager{
		dataDir:        geoipDir,
		cityIPv4File:   filepath.Join(geoipDir, "geolite2-city-ipv4.mmdb"),
		cityIPv6File:   filepath.Join(geoipDir, "geolite2-city-ipv6.mmdb"),
		countryFile:    filepath.Join(geoipDir, "geo-whois-asn-country.mmdb"),
		asnFile:        filepath.Join(geoipDir, "asn.mmdb"),
		updateInterval: updateInterval,
	}
}

// Initialize downloads databases if needed and loads them
func (m *Manager) Initialize() error {
	// Create data directory
	if err := os.MkdirAll(m.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Check if databases exist
	needsDownload := !m.databasesExist()

	if needsDownload {
		log.Println("GeoIP databases not found, downloading...")
		if err := m.DownloadDatabases(); err != nil {
			return fmt.Errorf("failed to download GeoIP databases: %w", err)
		}
	}

	// Load databases
	if err := m.loadDatabases(); err != nil {
		return fmt.Errorf("failed to load GeoIP databases: %w", err)
	}

	m.lastUpdate = time.Now()
	log.Println("GeoIP databases loaded successfully")
	return nil
}

// databasesExist checks if all required databases exist
func (m *Manager) databasesExist() bool {
	files := []string{m.cityIPv4File, m.cityIPv6File, m.countryFile, m.asnFile}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// loadDatabases loads all GeoIP databases into memory
func (m *Manager) loadDatabases() error {
	var err error

	// Load City IPv4
	m.cityIPv4DB, err = geoip2.Open(m.cityIPv4File)
	if err != nil {
		return fmt.Errorf("failed to load city IPv4 database: %w", err)
	}

	// Load City IPv6
	m.cityIPv6DB, err = geoip2.Open(m.cityIPv6File)
	if err != nil {
		return fmt.Errorf("failed to load city IPv6 database: %w", err)
	}

	// Load Country
	m.countryDB, err = geoip2.Open(m.countryFile)
	if err != nil {
		return fmt.Errorf("failed to load country database: %w", err)
	}

	// Load ASN
	m.asnDB, err = geoip2.Open(m.asnFile)
	if err != nil {
		return fmt.Errorf("failed to load ASN database: %w", err)
	}

	return nil
}

// DownloadDatabases downloads all 4 GeoIP databases from sapics/ip-location-db via CDN
func (m *Manager) DownloadDatabases() error {
	databases := map[string]struct {
		url      string
		filepath string
	}{
		"geolite2-city-ipv4":      {cityIPv4URL, m.cityIPv4File},
		"geolite2-city-ipv6":      {cityIPv6URL, m.cityIPv6File},
		"geo-whois-asn-country":   {countryURL, m.countryFile},
		"asn":                     {asnURL, m.asnFile},
	}

	for name, db := range databases {
		log.Printf("  Downloading %s.mmdb...", name)

		if err := m.downloadFile(db.url, db.filepath); err != nil {
			return fmt.Errorf("failed to download %s: %w", name, err)
		}
	}

	return nil
}

// downloadFile downloads a file from URL to local path
func (m *Manager) downloadFile(url, localPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// ShouldUpdate checks if databases need updating
func (m *Manager) ShouldUpdate() bool {
	return time.Since(m.lastUpdate) > m.updateInterval
}

// Update updates the databases if needed
func (m *Manager) Update() error {
	if !m.ShouldUpdate() {
		return nil
	}

	log.Println("Updating GeoIP databases...")
	if err := m.DownloadDatabases(); err != nil {
		return err
	}

	// Reload databases
	if err := m.loadDatabases(); err != nil {
		return err
	}

	m.lastUpdate = time.Now()
	log.Println("GeoIP databases updated successfully")
	return nil
}

// Reader returns a geo.Reader interface wrapping the databases
func (m *Manager) Reader() geo.Reader {
	return &geoipReader{
		cityIPv4DB: m.cityIPv4DB,
		cityIPv6DB: m.cityIPv6DB,
		countryDB:  m.countryDB,
		asnDB:      m.asnDB,
	}
}

// geoipReader implements geo.Reader interface with IPv4/IPv6 database selection
type geoipReader struct {
	cityIPv4DB *geoip2.Reader
	cityIPv6DB *geoip2.Reader
	countryDB  *geoip2.Reader
	asnDB      *geoip2.Reader
}

func (g *geoipReader) Country(ip net.IP) (geo.Country, error) {
	country := geo.Country{}
	if g.countryDB == nil {
		return country, nil
	}

	record, err := g.countryDB.Country(ip)
	if err != nil {
		return country, err
	}

	if c, exists := record.Country.Names["en"]; exists {
		country.Name = c
	}
	if c, exists := record.RegisteredCountry.Names["en"]; exists && country.Name == "" {
		country.Name = c
	}
	if record.Country.IsoCode != "" {
		country.ISO = record.Country.IsoCode
	}
	if record.RegisteredCountry.IsoCode != "" && country.ISO == "" {
		country.ISO = record.RegisteredCountry.IsoCode
	}
	isEU := record.Country.IsInEuropeanUnion || record.RegisteredCountry.IsInEuropeanUnion
	country.IsEU = &isEU

	return country, nil
}

func (g *geoipReader) City(ip net.IP) (geo.City, error) {
	city := geo.City{}

	// Select appropriate city database based on IP version
	var cityDB *geoip2.Reader
	if ip.To4() != nil {
		// IPv4
		cityDB = g.cityIPv4DB
	} else {
		// IPv6
		cityDB = g.cityIPv6DB
	}

	if cityDB == nil {
		return city, nil
	}

	record, err := cityDB.City(ip)
	if err != nil {
		return city, err
	}

	if c, exists := record.City.Names["en"]; exists {
		city.Name = c
	}
	if len(record.Subdivisions) > 0 {
		if c, exists := record.Subdivisions[0].Names["en"]; exists {
			city.RegionName = c
		}
		if record.Subdivisions[0].IsoCode != "" {
			city.RegionCode = record.Subdivisions[0].IsoCode
		}
	}
	city.Latitude = record.Location.Latitude
	city.Longitude = record.Location.Longitude

	// Metro code is US Only
	if record.Location.MetroCode > 0 && record.Country.IsoCode == "US" {
		city.MetroCode = record.Location.MetroCode
	}
	if record.Postal.Code != "" {
		city.PostalCode = record.Postal.Code
	}
	if record.Location.TimeZone != "" {
		city.Timezone = record.Location.TimeZone
	}

	return city, nil
}

func (g *geoipReader) ASN(ip net.IP) (geo.ASN, error) {
	asn := geo.ASN{}
	if g.asnDB == nil {
		return asn, nil
	}

	record, err := g.asnDB.ASN(ip)
	if err != nil {
		return asn, err
	}

	if record.AutonomousSystemNumber > 0 {
		asn.AutonomousSystemNumber = record.AutonomousSystemNumber
	}
	if record.AutonomousSystemOrganization != "" {
		asn.AutonomousSystemOrganization = record.AutonomousSystemOrganization
	}

	return asn, nil
}

func (g *geoipReader) IsEmpty() bool {
	return g.cityIPv4DB == nil && g.cityIPv6DB == nil && g.countryDB == nil
}
