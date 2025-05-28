package webrtc_gateway

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

// SignalMessage defines the structure for messages exchanged over WebSocket.
type SignalMessage struct {
	Type    string `json:"type"` // "offer", "answer", "candidate"
	Payload string `json:"payload"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for simplicity in this example.
		// In a production environment, you should implement proper origin checking.
		return true
	},
}

// SessionManagerClient defines the interface for interacting with a session manager.
// This is a conceptual interface, copied from sip_gateway for consistency.
type SessionManagerClient interface {
	RegisterSession(sessionID string, sessionDetails map[string]string) error
	DeregisterSession(sessionID string) error
}

// DummyWebRTCSessionManagerClient is a placeholder implementation for SessionManagerClient.
type DummyWebRTCSessionManagerClient struct{}

func (d *DummyWebRTCSessionManagerClient) RegisterSession(sessionID string, sessionDetails map[string]string) error {
	log.Printf("WEBRTC_SESSION_MANAGER_CLIENT: Registering session for SessionID: %s, Details: %v\n", sessionID, sessionDetails)
	return nil
}

func (d *DummyWebRTCSessionManagerClient) DeregisterSession(sessionID string) error {
	log.Printf("WEBRTC_SESSION_MANAGER_CLIENT: Deregistering session for SessionID: %s\n", sessionID)
	return nil
}

// HandleWebSocketConnections is the HTTP handler for WebSocket signaling.
func HandleWebSocketConnections(w http.ResponseWriter, r *http.Request, peerConnectionConfig webrtc.Configuration, smClient SessionManagerClient) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket connection established")
	sessionID := conn.RemoteAddr().String() // Use remote address as a simple session ID

	// Create a new PeerConnection
	// Note: In a more complex app, the API (media engine, settings) would be created once.
	// For simplicity, creating API per connection here, but this is not optimal.
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&webrtc.MediaEngine{})) // Basic media engine
	peerConnection, err := api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		log.Printf("Failed to create PeerConnection: %v", err)
		conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to create PeerConnection"})
		return
	}
	defer peerConnection.Close()

	// Use a WaitGroup to wait for ICE gathering to complete before sending the answer
	// This is a common pattern.
	var iceGatheringComplete sync.WaitGroup
	iceGatheringComplete.Add(1) // For the OnICEGatheringDone event

	// (Conceptual) Register session
	smClient.RegisterSession(sessionID, map[string]string{"type": "webrtc", "remoteAddr": sessionID})

	// Set up OnICECandidate handler
	// This sends candidates back to the client
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		candidateJSON, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Printf("Error marshalling ICE candidate: %v", err)
			return
		}
		if err := conn.WriteJSON(SignalMessage{Type: "candidate", Payload: string(candidateJSON)}); err != nil {
			log.Printf("Error sending ICE candidate over WebSocket: %v", err)
		} else {
			log.Printf("Sent ICE candidate: %s", string(candidateJSON))
		}
	})
	
	peerConnection.OnICEGatheringDone(func() {
        iceGatheringComplete.Done()
		log.Println("ICE Gathering Complete.")
    })


	// Set up OnTrack handler
	// This is where you would handle incoming media streams
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Received remote track: Kind %s, Codec %s, SSRC %d, ID %s, StreamID %s\n",
			track.Kind(), track.Codec().MimeType, track.SSRC(), track.ID(), track.StreamID())
		// In a real application, you would read from the track here.
		// For example, if it's an audio track, you might forward it to another service
		// or process it. For now, we just log its properties.
		// go func() {
		// 	buffer := make([]byte, 1500)
		// 	for {
		// 		_, _, readErr := track.Read(buffer)
		// 		if readErr != nil {
		// 			log.Printf("Error reading from track %s: %v", track.ID(), readErr)
		// 			return
		// 		}
		// 		// Do something with the buffer
		// 	}
		// }()
	})

	// Handle incoming messages from the WebSocket client
	for {
		var msg SignalMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Printf("WebSocket connection closed: %v", err)
			}
			// (Conceptual) Deregister session
			smClient.DeregisterSession(sessionID)
			break
		}

		log.Printf("Received WebSocket message: Type=%s", msg.Type)

		switch msg.Type {
		case "offer":
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(msg.Payload), &offer); err != nil {
				log.Printf("Error unmarshalling offer SDP: %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Invalid offer SDP"})
				continue
			}

			log.Println("Received SDP Offer")
			if err := peerConnection.SetRemoteDescription(offer); err != nil {
				log.Printf("Error setting remote description (offer): %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to set remote description"})
				continue
			}

			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				log.Printf("Error creating answer SDP: %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to create answer"})
				continue
			}

			// Sets the LocalDescription, and starts listining for UDP packets
			// Block until ICE Gathering is complete, disabling trickle ICE
			// This is not strictly required in all cases.
			// For simplicity, we wait here. In a production system, you might handle candidates as they arrive (trickle ICE).
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)


			if err := peerConnection.SetLocalDescription(answer); err != nil {
				log.Printf("Error setting local description (answer): %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to set local description"})
				continue
			}
			
			// Wait for ICE gathering to complete
			<-gatherComplete // This replaces the iceGatheringComplete.Wait() for this specific flow.
			// iceGatheringComplete.Wait() // Wait for all candidates to be gathered (if not using trickle ICE for answer)
			
			log.Println("Created SDP Answer, ICE gathering complete.")
			answerPayload, err := json.Marshal(peerConnection.LocalDescription())
			if err != nil {
				log.Printf("Error marshalling answer SDP: %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to marshal answer"})
				continue
			}

			if err := conn.WriteJSON(SignalMessage{Type: "answer", Payload: string(answerPayload)}); err != nil {
				log.Printf("Error sending answer SDP over WebSocket: %v", err)
			} else {
				log.Println("Sent SDP Answer")
			}

		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(msg.Payload), &candidate); err != nil {
				log.Printf("Error unmarshalling ICE candidate: %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Invalid ICE candidate"})
				continue
			}

			log.Printf("Received ICE Candidate: %v", candidate.Candidate)
			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Printf("Error adding ICE candidate: %v", err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to add ICE candidate"})
				continue
			}
			log.Println("Added ICE Candidate")
		
		default:
			log.Printf("Received unknown message type: %s", msg.Type)
			conn.WriteJSON(SignalMessage{Type: "error", Payload: "Unknown message type"})
		}
	}
}
