package utils

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// getAccessibleURL returns the most relevant URL for accessing the server
// Priority: FQDN > hostname > public IP > fallback
// NEVER shows localhost, 127.0.0.1, or 0.0.0.0
func GetAccessibleURL(port string) string {
	// Try to get hostname
	hostname, err := os.Hostname()
	if err == nil && hostname != "" && hostname != "localhost" {
		// Try to resolve hostname to see if it's a valid FQDN
		if addrs, err := net.LookupHost(hostname); err == nil && len(addrs) > 0 {
			return formatURLWithHost(hostname, port)
		}
	}

	// Try to get outbound IP (most likely accessible IP)
	if ip := GetOutboundIP(); ip != "" {
		return formatURLWithIP(ip, port)
	}

	// Fallback to hostname if we have one
	if hostname != "" && hostname != "localhost" {
		return formatURLWithHost(hostname, port)
	}

	// Last resort: use a generic message
	return fmt.Sprintf("http://<your-host>:%s", port)
}

// GetOutboundIP gets the preferred outbound IP of this machine
func GetOutboundIP() string {
	// Try IPv4 first
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	// Try IPv6
	conn, err = net.Dial("udp", "[2001:4860:4860::8888]:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}

	return ""
}

// formatURLWithIP formats a URL with IP address (handles IPv6 brackets)
func formatURLWithIP(ip, port string) string {
	// IPv6 addresses need brackets
	if strings.Contains(ip, ":") {
		return fmt.Sprintf("http://[%s]:%s", ip, port)
	}
	return fmt.Sprintf("http://%s:%s", ip, port)
}

// formatURLWithHost formats a URL with hostname
func formatURLWithHost(hostname, port string) string {
	return fmt.Sprintf("http://%s:%s", hostname, port)
}

// IsIPv6 checks if the given IP is IPv6
func IsIPv6(ip net.IP) bool {
	return ip.To4() == nil
}

// ParseIP parses an IP address string, removing brackets if present
func ParseIP(ipStr string) (net.IP, error) {
	// Remove brackets if present (from URL)
	ipStr = strings.Trim(ipStr, "[]")

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	return ip, nil
}
