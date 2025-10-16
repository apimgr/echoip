package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

// Directories holds the application directories
type Directories struct {
	Config string
	Data   string
	Logs   string
}

// GetDirectories returns OS-specific directories
// Priority: Environment variables > OS defaults
func GetDirectories() Directories {
	// Check environment variables first
	configDir := os.Getenv("CONFIG_DIR")
	dataDir := os.Getenv("DATA_DIR")
	logsDir := os.Getenv("LOGS_DIR")

	// If not set, use OS-specific defaults
	if configDir == "" {
		configDir = getDefaultConfigDir()
	}
	if dataDir == "" {
		dataDir = getDefaultDataDir()
	}
	if logsDir == "" {
		logsDir = getDefaultLogsDir()
	}

	return Directories{
		Config: configDir,
		Data:   dataDir,
		Logs:   logsDir,
	}
}

// getDefaultConfigDir returns OS-specific config directory
func getDefaultConfigDir() string {
	switch runtime.GOOS {
	case "linux":
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			return filepath.Join(configHome, "echoip")
		}
		return "/etc/echoip"
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "echoip")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "echoip")
	default:
		return filepath.Join(".", "config")
	}
}

// getDefaultDataDir returns OS-specific data directory
func getDefaultDataDir() string {
	switch runtime.GOOS {
	case "linux":
		if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
			return filepath.Join(dataHome, "echoip")
		}
		return "/var/lib/echoip"
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "echoip")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "echoip")
	default:
		return filepath.Join(".", "data")
	}
}

// getDefaultLogsDir returns OS-specific logs directory
func getDefaultLogsDir() string {
	switch runtime.GOOS {
	case "linux":
		return "/var/log/echoip"
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", "echoip")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "echoip", "logs")
	default:
		return filepath.Join(".", "logs")
	}
}

// EnsureDirectories creates all required directories
func EnsureDirectories(dirs Directories) error {
	for _, dir := range []string{dirs.Config, dirs.Data, dirs.Logs} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
