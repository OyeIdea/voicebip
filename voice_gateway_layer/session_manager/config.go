package session_manager

import (
	"log"
	"os"
	"strconv"
)

// SessionManagerConfig holds the configuration for the Session Manager service.
type SessionManagerConfig struct {
	ListenAddress string
	ApiPort       int
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// LoadConfig returns a SessionManagerConfig instance, loading values from
// environment variables with specified fallbacks.
func LoadConfig() SessionManagerConfig {
	listenAddress := getEnv("SESSION_MANAGER_LISTEN_ADDRESS", "0.0.0.0")

	apiPortStr := getEnv("SESSION_MANAGER_API_PORT", "8000")
	apiPort, err := strconv.Atoi(apiPortStr)
	if err != nil {
		// Use the LogPrefix defined in session_manager.go or api.go if available,
		// otherwise, a generic prefix.
		// For now, assuming LogPrefix is not directly accessible here or to avoid import cycle,
		// using a local prefix.
		log.Printf("[SESSION_MANAGER_CONFIG][WARN] Invalid SESSION_MANAGER_API_PORT value '%s', using default 8000. Error: %v", apiPortStr, err)
		apiPort = 8000
	}

	return SessionManagerConfig{
		ListenAddress: listenAddress,
		ApiPort:       apiPort,
	}
}
