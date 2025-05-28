package webrtc_gateway

// WebRTCConfig holds the configuration for the WebRTC Gateway.
type WebRTCConfig struct {
	ListenAddress string
	SignalPort    int
	StunServers   []string
}

// LoadConfig returns a default WebRTCConfig instance.
// In a real application, this would load from a file or environment variables.
func LoadConfig() WebRTCConfig {
	return WebRTCConfig{
		ListenAddress: "0.0.0.0",
		SignalPort:    8080,
		StunServers: []string{
			"stun:stun.l.google.com:19302",
		},
	}
}
