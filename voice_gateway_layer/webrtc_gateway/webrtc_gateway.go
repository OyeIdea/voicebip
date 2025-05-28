package webrtc_gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pion/webrtc/v3"
)

// StartWebRTCGateway initializes and starts the WebRTC gateway.
func StartWebRTCGateway(cfg WebRTCConfig, smClient SessionManagerClient) {
	// Prepare the WebRTC PeerConnection configuration
	// For this basic implementation, we'll use a public STUN server.
	// In a production environment, you'd likely have your own TURN servers.
	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: cfg.StunServers,
			},
		},
	}

	// Set up the HTTP handler for WebSocket signaling
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		HandleWebSocketConnections(w, r, peerConnectionConfig, smClient)
	})

	// Start the HTTP server for signaling
	listenAddr := fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.SignalPort)
	log.Printf("WebRTC Gateway signaling server starting on %s", listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Failed to start WebRTC signaling server: %v", err)
	}
}

// Example main function (can be removed or kept for standalone testing)
// func main() {
// 	cfg := LoadConfig()
// 	var smClient SessionManagerClient = &DummyWebRTCSessionManagerClient{} // Use dummy client
// 	StartWebRTCGateway(cfg, smClient)
// }
