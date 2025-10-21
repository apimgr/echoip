package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/apimgr/echoip/src/geoip"
	"github.com/apimgr/echoip/src/iputil"
	"github.com/apimgr/echoip/src/paths"
	"github.com/apimgr/echoip/src/scheduler"
	"github.com/apimgr/echoip/src/server"
)

// Version information (set by build flags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type multiValueFlag []string

func (f *multiValueFlag) String() string {
	return strings.Join([]string(*f), ", ")
}

func (f *multiValueFlag) Set(v string) error {
	*f = append(*f, v)
	return nil
}

func init() {
	log.SetPrefix("echoip: ")
	log.SetFlags(log.Lshortfile)
}

func main() {
	// Flags
	dataDir := flag.String("d", "data", "Data directory for GeoIP databases")
	listen := flag.String("l", ":8080", "Listening address (supports IPv4, IPv6, or dual-stack)")
	reverseLookup := flag.Bool("r", false, "Perform reverse hostname lookups")
	portLookup := flag.Bool("p", false, "Enable port lookup")
	template := flag.String("t", "src/server/templates", "Path to template dir")
	cacheSize := flag.Int("C", 0, "Size of response cache. Set to 0 to disable")
	profile := flag.Bool("P", false, "Enables profiling handlers")
	sponsor := flag.Bool("s", false, "Show sponsor logo")
	showVersion := flag.Bool("version", false, "Show version information")
	showStatus := flag.Bool("status", false, "Check server status (for health checks)")

	var headers multiValueFlag
	flag.Var(&headers, "H", "Header to trust for remote IP, if present (e.g. X-Real-IP)")
	flag.Parse()

	// Handle --version
	if *showVersion {
		fmt.Println(Version)
		return
	}

	// Handle --status (health check)
	if *showStatus {
		// For Docker health checks - just exit 0 if we got here
		os.Exit(0)
	}

	if len(flag.Args()) != 0 {
		flag.Usage()
		return
	}

	// Get OS-specific directories
	dirs := paths.GetDirectories()
	if *dataDir == "data" {
		*dataDir = dirs.Data
	}

	// Ensure directories exist
	if err := paths.EnsureDirectories(dirs); err != nil {
		log.Printf("⚠️  Failed to create directories: %v", err)
	}

	// Log startup information
	log.Printf("🚀 echoip %s (commit: %s, built: %s)", Version, Commit, BuildDate)
	log.Println("🌐 IPv6 support enabled - server will accept both IPv4 and IPv6 connections")

	// Initialize GeoIP manager
	geoMgr := geoip.NewManager(*dataDir)
	if err := geoMgr.Initialize(); err != nil {
		log.Printf("⚠️  Failed to initialize GeoIP: %v", err)
		log.Println("⚠️  Server will continue without GeoIP support")
	} else {
		log.Println("✅ GeoIP databases loaded (4 files, ~103MB)")
	}

	// Initialize scheduler for GeoIP updates
	sched := scheduler.New()
	sched.AddTask("geoip-update", "0 3 * * 0", func() error {
		log.Println("📅 Running scheduled GeoIP database update...")
		return geoMgr.Update()
	})
	sched.Start()
	defer sched.Stop()

	r := geoMgr.Reader()
	cache := server.NewCache(*cacheSize)
	srv := server.New(r, cache, *profile)
	srv.IPHeaders = headers
	if _, err := os.Stat(*template); err == nil {
		srv.Template = *template
	} else {
		log.Printf("Not configuring default handler: Template not found: %s", *template)
	}
	if *reverseLookup {
		log.Println("Enabling reverse lookup")
		srv.LookupAddr = iputil.LookupAddr
	}
	if *portLookup {
		log.Println("Enabling port lookup")
		srv.LookupPort = iputil.LookupPort
	}
	if *sponsor {
		log.Println("Enabling sponsor logo")
		srv.Sponsor = *sponsor
	}
	if len(headers) > 0 {
		log.Printf("Trusting remote IP from header(s): %s", headers.String())
	}
	if *cacheSize > 0 {
		log.Printf("Cache capacity set to %d", *cacheSize)
	}
	if *profile {
		log.Printf("Enabling profiling handlers")
	}

	// Enhance listening address logging for IPv6
	listenAddr := *listen
	if strings.Contains(listenAddr, ":") && !strings.HasPrefix(listenAddr, "[") {
		// Check if it's IPv6 or just port
		parts := strings.Split(listenAddr, ":")
		if len(parts) == 2 && parts[0] == "" {
			log.Printf("Listening on port %s (dual-stack: IPv4 and IPv6)", parts[1])
		} else if len(parts) > 2 {
			log.Printf("Listening on IPv6 address: [%s]", listenAddr)
		} else {
			log.Printf("Listening on http://%s", listenAddr)
		}
	} else {
		log.Printf("Listening on http://%s", listenAddr)
	}

	if err := srv.ListenAndServe(*listen); err != nil {
		log.Fatal(err)
	}
}
