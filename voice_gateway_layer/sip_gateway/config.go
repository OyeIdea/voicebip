package sip_gateway

import (
	"log"
	"os"
	"strconv"
)

// SIPConfig holds the configuration for the SIP Gateway.
type SIPConfig struct {
	ListenAddress           string
	ListenPort              int
	SessionManagerAPIEndpoint string
}

// getEnv retrieves an environment variable or returns a fallback value.
func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// LoadConfig returns a SIPConfig instance, loading values from
// environment variables with specified fallbacks.
func LoadConfig() SIPConfig {
	listenAddress := getEnv("SIP_LISTEN_ADDRESS", "0.0.0.0")

	listenPortStr := getEnv("SIP_LISTEN_PORT", "5060")
	listenPort, err := strconv.Atoi(listenPortStr)
	if err != nil {
		// Assuming SipGatewayLogPrefix might not be directly accessible here or to avoid import cycle,
		// using a local prefix or the one from session_client.go if appropriate.
		// For now, using a generic log statement.
		log.Printf("[SIP_GATEWAY_CONFIG][WARN] Invalid SIP_LISTEN_PORT value '%s', using default 5060. Error: %v", listenPortStr, err)
		listenPort = 5060
	}

	sessionManagerAPIEndpoint := getEnv("SIP_SESSION_MANAGER_API_ENDPOINT", "http://localhost:8000")

	return SIPConfig{
		ListenAddress:           listenAddress,
		ListenPort:              listenPort,
		SessionManagerAPIEndpoint: sessionManagerAPIEndpoint,
	}
}
