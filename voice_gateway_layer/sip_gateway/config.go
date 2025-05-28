package sip_gateway

// SIPConfig holds the configuration for the SIP Gateway.
type SIPConfig struct {
	ListenAddress string
	ListenPort    int
}

// LoadConfig returns a default SIPConfig instance.
// In a real application, this would load from a file or environment variables.
func LoadConfig() SIPConfig {
	return SIPConfig{
		ListenAddress: "0.0.0.0",
		ListenPort:    5060, // Standard SIP port
	}
}
