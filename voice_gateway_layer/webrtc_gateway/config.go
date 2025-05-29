package webrtc_gateway

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// WebRTCConfig holds the configuration for the WebRTC Gateway.
type WebRTCConfig struct {
	ListenAddress           string
	SignalPort              int
	StunServers             []string
	SessionManagerAPIEndpoint string
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// LoadConfig returns a WebRTCConfig instance, loading values from
// environment variables with specified fallbacks.
func LoadConfig() WebRTCConfig {
	listenAddress := getEnv("WEBRTC_LISTEN_ADDRESS", "0.0.0.0")

	signalPortStr := getEnv("WEBRTC_SIGNAL_PORT", "8080")
	signalPort, err := strconv.Atoi(signalPortStr)
	if err != nil {
		// Assuming WebRTCGatewayLogPrefix might not be directly accessible here, using a local prefix.
		log.Printf("[WEBRTC_GATEWAY_CONFIG][WARN] Invalid WEBRTC_SIGNAL_PORT value '%s', using default 8080. Error: %v", signalPortStr, err)
		signalPort = 8080
	}

	stunServersStr := getEnv("WEBRTC_STUN_SERVERS", "stun:stun.l.google.com:19302")
	var stunServersList []string
	if strings.TrimSpace(stunServersStr) != "" {
		stunServersList = strings.Split(stunServersStr, ",")
		for i, server := range stunServersList {
			stunServersList[i] = strings.TrimSpace(server)
		}
	} else {
		// Fallback to default if the env var is set but empty after trimming, or if Split results in empty.
		// This case is mostly covered by getEnv's fallback, but good to be robust.
		stunServersList = []string{"stun:stun.l.google.com:19302"}
	}


	sessionManagerAPIEndpoint := getEnv("WEBRTC_SESSION_MANAGER_API_ENDPOINT", "http://localhost:8000")

	return WebRTCConfig{
		ListenAddress:           listenAddress,
		SignalPort:              signalPort,
		StunServers:             stunServersList,
		SessionManagerAPIEndpoint: sessionManagerAPIEndpoint,
	}
}
