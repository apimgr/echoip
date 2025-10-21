package database

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Setting represents a configuration setting
type Setting struct {
	Key            string
	Value          string
	Type           string
	Category       string
	Description    string
	RequiresReload bool
}

// InitializeDefaultSettings creates default settings
func (db *DB) InitializeDefaultSettings() error {
	settings := []Setting{
		// CORS (default: allow all)
		{"server.cors_enabled", "true", "boolean", "server", "Enable CORS", false},
		{"server.cors_origins", `["*"]`, "json", "server", "Allowed origins", false},
		{"server.cors_methods", `["GET","POST","PUT","DELETE","OPTIONS"]`, "json", "server", "Allowed methods", false},
		{"server.cors_headers", `["Content-Type","Authorization"]`, "json", "server", "Allowed headers", false},
		{"server.cors_credentials", "false", "boolean", "server", "Allow credentials", false},

		// Rate Limiting
		{"rate.enabled", "true", "boolean", "rate", "Enable rate limiting", false},
		{"rate.global_rps", "100", "number", "rate", "Global requests per second", false},
		{"rate.global_burst", "200", "number", "rate", "Global burst allowance", false},
		{"rate.api_rps", "50", "number", "rate", "API requests per second", false},
		{"rate.api_burst", "100", "number", "rate", "API burst allowance", false},
		{"rate.admin_rps", "10", "number", "rate", "Admin requests per second", false},
		{"rate.admin_burst", "20", "number", "rate", "Admin burst allowance", false},

		// Request Limits
		{"request.timeout", "60", "number", "request", "Request timeout (seconds)", false},
		{"request.max_size", "10485760", "number", "request", "Max body size (bytes)", false},
		{"request.max_header_size", "1048576", "number", "request", "Max header size (bytes)", false},

		// Connection Limits
		{"connection.max_concurrent", "1000", "number", "connection", "Max concurrent requests", false},
		{"connection.idle_timeout", "120", "number", "connection", "Idle timeout (seconds)", false},
		{"connection.read_timeout", "10", "number", "connection", "Read timeout (seconds)", false},
		{"connection.write_timeout", "10", "number", "connection", "Write timeout (seconds)", false},

		// Security Headers
		{"security.frame_options", "DENY", "string", "security", "X-Frame-Options", false},
		{"security.content_type_options", "nosniff", "string", "security", "X-Content-Type-Options", false},
		{"security.xss_protection", "1; mode=block", "string", "security", "X-XSS-Protection", false},
		{"security.csp", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'", "string", "security", "Content-Security-Policy", false},
		{"security.hsts_enabled", "true", "boolean", "security", "Enable HSTS", false},
		{"security.hsts_max_age", "31536000", "number", "security", "HSTS max age (seconds)", false},

		// Logging
		{"log.level", "info", "string", "log", "Log level (debug, info, warn, error)", false},
		{"log.access_log", "true", "boolean", "log", "Enable access logging", false},
		{"log.security_log", "true", "boolean", "log", "Enable security event logging", false},
	}

	for _, s := range settings {
		_, err := db.Exec(`
			INSERT OR IGNORE INTO settings (key, value, type, category, description, requires_reload)
			VALUES (?, ?, ?, ?, ?, ?)
		`, s.Key, s.Value, s.Type, s.Category, s.Description, s.RequiresReload)

		if err != nil {
			return fmt.Errorf("failed to insert setting %s: %w", s.Key, err)
		}
	}

	return nil
}

// GetSetting retrieves a setting value
func (db *DB) GetSetting(key string) (interface{}, error) {
	var value, typ string
	err := db.QueryRow(`
		SELECT value, type FROM settings WHERE key = ?
	`, key).Scan(&value, &typ)

	if err != nil {
		return nil, err
	}

	return parseSetting(value, typ)
}

// SetSetting updates a setting value
func (db *DB) SetSetting(key string, value interface{}) error {
	valueStr := fmt.Sprintf("%v", value)

	_, err := db.Exec(`
		UPDATE settings SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?
	`, valueStr, key)

	return err
}

// GetAllSettings retrieves all settings
func (db *DB) GetAllSettings() (map[string]interface{}, error) {
	rows, err := db.Query(`SELECT key, value, type FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	settings := make(map[string]interface{})

	for rows.Next() {
		var key, value, typ string
		if err := rows.Scan(&key, &value, &typ); err != nil {
			continue
		}

		parsed, err := parseSetting(value, typ)
		if err != nil {
			continue
		}

		settings[key] = parsed
	}

	return settings, nil
}

// parseSetting parses a setting value based on its type
func parseSetting(value, typ string) (interface{}, error) {
	switch typ {
	case "boolean":
		return strconv.ParseBool(value)
	case "number":
		return strconv.Atoi(value)
	case "json":
		var result interface{}
		err := json.Unmarshal([]byte(value), &result)
		return result, err
	default:
		return value, nil
	}
}
