package webrtc_gateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid" // For generating unique session IDs
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
		return true
	},
}

// HandleWebSocketConnections is the HTTP handler for WebSocket signaling.
func HandleWebSocketConnections(w http.ResponseWriter, r *http.Request, peerConnectionConfig webrtc.Configuration, smClient SessionManagerClient) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("%s[ERROR][HandleWebSocketConnections] Failed to upgrade connection to WebSocket: %v. RemoteAddr: %s", WebRTCGatewayLogPrefix, err, r.RemoteAddr)
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	log.Printf("%s[INFO][HandleWebSocketConnections] WebSocket connection established. SessionID: %s, RemoteAddr: %s", WebRTCGatewayLogPrefix, sessionID, conn.RemoteAddr().String())

	api := webrtc.NewAPI(webrtc.WithMediaEngine(&webrtc.MediaEngine{}))
	peerConnection, err := api.NewPeerConnection(peerConnectionConfig)
	if err != nil {
		log.Printf("%s[ERROR][HandleWebSocketConnections] Failed to create PeerConnection: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
		// Attempt to send error over WebSocket, but connection might be unstable
		wsErr := conn.WriteJSON(SignalMessage{Type: "error", Payload: fmt.Sprintf("Failed to create PeerConnection: %v", err)})
		if wsErr != nil {
			log.Printf("%s[ERROR][HandleWebSocketConnections] Failed to send PeerConnection creation error over WebSocket: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, wsErr)
		}
		return
	}
	defer func() {
		if err := peerConnection.Close(); err != nil {
			log.Printf("%s[ERROR][HandleWebSocketConnections] Error closing PeerConnection: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
		}
		log.Printf("%s[INFO][HandleWebSocketConnections] PeerConnection closed. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
	}()

	initialDetails := map[string]string{
		"remote_address": conn.RemoteAddr().String(),
		"user_agent":     r.UserAgent(),
	}
	err = smClient.RegisterSession(sessionID, "WebRTC", initialDetails)
	if err != nil {
		log.Printf("%s[ERROR][HandleWebSocketConnections] Error registering session with Session Manager: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
		wsErr := conn.WriteJSON(SignalMessage{Type: "error", Payload: fmt.Sprintf("Session registration failed: %v", err)})
		if wsErr != nil {
			log.Printf("%s[ERROR][HandleWebSocketConnections] Failed to send session registration error over WebSocket: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, wsErr)
		}
		return
	}
	log.Printf("%s[INFO][HandleWebSocketConnections] Session registered successfully with Session Manager. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
	var sessionActive bool

	defer func() {
		log.Printf("%s[INFO][HandleWebSocketConnections] Deregistering session with Session Manager. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
		if err := smClient.DeregisterSession(sessionID); err != nil {
			log.Printf("%s[ERROR][HandleWebSocketConnections] Error deregistering session: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
		} else {
			log.Printf("%s[INFO][HandleWebSocketConnections] Session deregistered successfully. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
		}
	}()

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			log.Printf("%s[DEBUG][OnICECandidate] Nil candidate received, usually means ICE gathering is complete. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
			return
		}
		candidateJSON, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Printf("%s[ERROR][OnICECandidate] Error marshalling ICE candidate: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
			return
		}
		log.Printf("%s[INFO][OnICECandidate] Sending ICE candidate: SessionID=%s, Candidate=%s", WebRTCGatewayLogPrefix, sessionID, string(candidateJSON))
		if err := conn.WriteJSON(SignalMessage{Type: "candidate", Payload: string(candidateJSON)}); err != nil {
			log.Printf("%s[ERROR][OnICECandidate] Error sending ICE candidate over WebSocket: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
		}
	})

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("%s[INFO][OnICEConnectionStateChange] ICE Connection State changed: SessionID=%s, State=%s", WebRTCGatewayLogPrefix, sessionID, state.String())
		switch state {
		case webrtc.ICEConnectionStateConnected:
			log.Printf("%s[INFO][OnICEConnectionStateChange] ICE Connected. Session is active. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
			if !sessionActive {
				err := smClient.UpdateSessionState(sessionID, "active")
				if err != nil {
					log.Printf("%s[ERROR][OnICEConnectionStateChange] Error updating session state to active: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				} else {
					log.Printf("%s[INFO][OnICEConnectionStateChange] Session state updated to active. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
					sessionActive = true
				}
			}
		case webrtc.ICEConnectionStateFailed, webrtc.ICEConnectionStateDisconnected, webrtc.ICEConnectionStateClosed:
			log.Printf("%s[INFO][OnICEConnectionStateChange] ICE Disconnected/Failed/Closed. SessionID: %s, State: %s. Closing WebSocket.", WebRTCGatewayLogPrefix, sessionID, state.String())
			conn.Close() // Triggers deferred DeregisterSession
		}
	})

	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("%s[INFO][OnTrack] Received remote track: SessionID=%s, Kind=%s, Codec=%s, SSRC=%d, ID=%s, StreamID=%s",
			WebRTCGatewayLogPrefix, sessionID, track.Kind(), track.Codec().MimeType, track.SSRC(), track.ID(), track.StreamID())
	})

	for {
		var msg SignalMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				log.Printf("%s[INFO][HandleWebSocketConnections] WebSocket connection closed gracefully or as expected: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
			} else {
				log.Printf("%s[WARN][HandleWebSocketConnections] Error reading JSON from WebSocket: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
			}
			break 
		}

		log.Printf("%s[INFO][HandleWebSocketConnections] Received WebSocket message: SessionID=%s, Type=%s", WebRTCGatewayLogPrefix, sessionID, msg.Type)

		switch msg.Type {
		case "offer":
			offer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(msg.Payload), &offer); err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Offer] Error unmarshalling offer SDP: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Invalid offer SDP"})
				continue
			}

			log.Printf("%s[INFO][HandleWebSocketConnections][Offer] Received SDP Offer. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
			if err := peerConnection.SetRemoteDescription(offer); err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Offer] Error setting remote description: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: fmt.Sprintf("Failed to set remote description: %v", err)})
				continue
			}

			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Offer] Error creating answer SDP: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: fmt.Sprintf("Failed to create answer: %v", err)})
				continue
			}
			
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
			if err := peerConnection.SetLocalDescription(answer); err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Offer] Error setting local description: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: fmt.Sprintf("Failed to set local description: %v", err)})
				continue
			}
			
			<-gatherComplete 
			
			log.Printf("%s[INFO][HandleWebSocketConnections][Offer] Created SDP Answer, ICE gathering complete. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
			answerPayload, err := json.Marshal(peerConnection.LocalDescription())
			if err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Offer] Error marshalling answer SDP: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Failed to marshal answer"})
				continue
			}

			log.Printf("%s[INFO][HandleWebSocketConnections][Offer] Sending SDP Answer. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
			if err := conn.WriteJSON(SignalMessage{Type: "answer", Payload: string(answerPayload)}); err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Offer] Error sending answer SDP over WebSocket: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
			}

		case "candidate":
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(msg.Payload), &candidate); err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Candidate] Error unmarshalling ICE candidate: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: "Invalid ICE candidate"})
				continue
			}

			log.Printf("%s[INFO][HandleWebSocketConnections][Candidate] Received ICE Candidate: SessionID=%s, Candidate=%s", WebRTCGatewayLogPrefix, sessionID, candidate.Candidate)
			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Printf("%s[ERROR][HandleWebSocketConnections][Candidate] Error adding ICE candidate: SessionID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, err)
				conn.WriteJSON(SignalMessage{Type: "error", Payload: fmt.Sprintf("Failed to add ICE candidate: %v", err)})
				continue
			}
			log.Printf("%s[INFO][HandleWebSocketConnections][Candidate] Added ICE Candidate. SessionID: %s", WebRTCGatewayLogPrefix, sessionID)
		
		default:
			log.Printf("%s[WARN][HandleWebSocketConnections] Received unknown message type: SessionID=%s, Type=%s", WebRTCGatewayLogPrefix, sessionID, msg.Type)
			conn.WriteJSON(SignalMessage{Type: "error", Payload: "Unknown message type"})
		}
	}
}
