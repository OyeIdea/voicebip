package session_manager

// SessionManagerConfig holds the configuration for the Session Manager service.
type SessionManagerConfig struct {
	ListenAddress string
	ApiPort       int
}

// LoadConfig returns a default SessionManagerConfig instance.
// In a real application, this would load from a file or environment variables.
func LoadConfig() SessionManagerConfig {
	return SessionManagerConfig{
		ListenAddress: "0.0.0.0",
		ApiPort:       8000,
	}
}
