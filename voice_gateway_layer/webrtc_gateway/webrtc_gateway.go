package webrtc_gateway

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pion/webrtc/v3"
)

// StartWebRTCGateway initializes and starts the WebRTC gateway.
func StartWebRTCGateway(cfg WebRTCConfig) {
	log.Printf("%s[INFO][StartWebRTCGateway] Initializing WebRTC Gateway...", WebRTCGatewayLogPrefix)

	smHTTPClient := NewHTTPMeetingSessionManagerClient(cfg.SessionManagerAPIEndpoint)

	peerConnectionConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: cfg.StunServers,
			},
		},
	}
	log.Printf("%s[DEBUG][StartWebRTCGateway] PeerConnection configuration prepared with STUN servers: %v", WebRTCGatewayLogPrefix, cfg.StunServers)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s[INFO][StartWebRTCGateway] Incoming WebSocket request for /ws from %s", WebRTCGatewayLogPrefix, r.RemoteAddr)
		HandleWebSocketConnections(w, r, peerConnectionConfig, smHTTPClient)
	})

	listenAddr := fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.SignalPort)
	log.Printf("%s[INFO][StartWebRTCGateway] WebRTC Gateway signaling server starting on http://%s", WebRTCGatewayLogPrefix, listenAddr)

	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("%s[FATAL][StartWebRTCGateway] Failed to start WebRTC signaling server on %s: %v", WebRTCGatewayLogPrefix, listenAddr, err)
	}
}

// Example main function (can be removed or kept for standalone testing)
// func main() {
// 	log.Printf("%s[INFO][main] Starting WebRTC Gateway service...", WebRTCGatewayLogPrefix)
// 	cfg := LoadConfig()
// 	StartWebRTCGateway(cfg)
// }
