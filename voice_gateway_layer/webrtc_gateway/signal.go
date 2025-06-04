package webrtc_gateway

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid" // For generating unique session IDs
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"revovoiceai/voice_gateway_layer/internal/protos/real_time_processing"
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

		// Handle audio tracks
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			go func() {
				// Buffer for incoming RTP packets
				rtpBuf := make([]byte, 1500) // Standard MTU size for RTP
				var sequenceNumber uint32 = 0
				for {
					n, _, readErr := track.Read(rtpBuf)
					if readErr != nil {
						log.Printf("%s[ERROR][OnTrack][Audio] Error reading from track: SessionID=%s, TrackID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, track.ID(), readErr)
						return // Stop processing this track
					}

					// Create and populate AudioSegment
					// Note: For WebRTC, RTP packet data is directly the audio data for codecs like Opus.
					// No separate RTP header stripping is typically needed here as Pion handles it.

					// AudioFormat Handling Notes for OPUS:
					// WebRTC commonly uses the Opus codec for audio, which is a compressed format.
					// The AudioSegment below correctly marks the audio_format as OPUS.
					//
					// IMPORTANT CONSIDERATION FOR DOWNSTREAM STT:
					// Many Speech-to-Text (STT) services, especially those designed for live
					// streaming and high accuracy (like Deepgram when not in a simulated mode),
					// expect uncompressed audio data, typically Linear PCM (e.g., L16).
					//
					// Therefore, for a production pipeline involving such an STT service,
					// the Opus audio received here would need to be DECODED to PCM
					// before being sent to that STT engine. This decoding could potentially
					// happen:
					//    1. Within this WebRTC Gateway before creating the AudioSegment.
					//    2. In an intermediate audio processing microservice.
					//    3. Potentially in the StreamingDataManager if it's designed for such tasks.
					//
					// For current testing purposes, especially if the downstream STT is
					// simulated or if it's an STT service that explicitly supports direct
					// Opus ingestion (less common for live streaming APIs), sending the
					// OPUS format directly is acceptable for verifying pipeline flow.
					segment := &real_time_processing.AudioSegment{
						SessionId:      sessionID,
						Timestamp:      time.Now().UnixMilli(),
						AudioFormat:    real_time_processing.AudioFormat_OPUS, // Assuming Opus for WebRTC
						SequenceNumber: sequenceNumber,
						Data:           rtpBuf[:n], // Use the actual read bytes
						IsFinal:        false,      // VAD will determine this later
					}
					sendAudioSegmentToSdmWebRTC(segment)
					sequenceNumber++
				}
			}()
		}
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

// sendAudioSegmentToSdmWebRTC logs the audio segment for WebRTC.
// In a real system, this would send the segment to the Stream Dialog Manager (SDM) via gRPC.
func sendAudioSegmentToSdmWebRTC(segment *real_time_processing.AudioSegment) {
	sdmAddress := "localhost:50051"

	conn, err := grpc.Dial(sdmAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("%s[ERROR][sendAudioSegmentToSdmWebRTC] Failed to connect to StreamingDataManager at %s: %v", WebRTCGatewayLogPrefix, sdmAddress, err)
		return
	}
	defer conn.Close()

	client := real_time_processing.NewStreamIngestClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// log.Printf("%s[INFO][sendAudioSegmentToSdmWebRTC] Sending segment: SID=%s, Seq=%d", WebRTCGatewayLogPrefix, segment.SessionId, segment.SequenceNumber)
	response, err := client.IngestAudioSegment(ctx, segment)
	if err != nil {
		log.Printf("%s[ERROR][sendAudioSegmentToSdmWebRTC] Error calling IngestAudioSegment: %v", WebRTCGatewayLogPrefix, err)
		return
	}
	log.Printf("%s[INFO][sendAudioSegmentToSdmWebRTC] Received response from SDM: SID=%s, Seq=%d, Status='%s'", WebRTCGatewayLogPrefix, response.GetSessionId(), response.GetSequenceNumber(), response.GetStatusMessage())
}
