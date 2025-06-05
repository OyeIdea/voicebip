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
	// import "github.com/pion/opus" // Conceptual import for Opus decoding
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

					// Conceptual Opus Decoder Initialization
					// Assumes the track is Opus, 48kHz, stereo (though we'll process as mono)
					// In a real scenario, sample rate and channels might be derived from SDP or track settings.
					// const opusSampleRate = 48000
					// const opusChannels = 2 // Pion's Opus library typically decodes stereo Opus to stereo PCM.
										 // We would then need to select one channel or mix down.
										 // For STT, mono is usually required.
					// decoder, err := opus.NewDecoder(opusSampleRate, opusChannels)
					// if err != nil {
					// 	log.Printf("%s[ERROR][OnTrack][Audio] Failed to create Opus decoder for SessionID %s: %v", WebRTCGatewayLogPrefix, sessionID, err)
					// 	return
					// }
					// log.Printf("%s[INFO][OnTrack][Audio] Conceptual Opus decoder initialized for SessionID %s", WebRTCGatewayLogPrefix, sessionID)

					// For STT, we usually need mono Linear16 PCM, often at 16kHz.
					// The pion/opus decoder outputs PCM matching Opus's sample rate (e.g., 48kHz) and channels.
					// Additional steps for resampling (e.g., 48kHz to 16kHz) and channel selection/mixing (stereo to mono)
					// would be needed after decoding. These are complex and omitted here.
					// We will placeholder by directly using rtpBuf and marking as LINEAR16,
					// assuming 16kHz mono for the STT service based on previous decisions.

				rtpBuf := make([]byte, 1500) // Standard MTU size for RTP
				var sequenceNumber uint32 = 0
				for {
					n, _, readErr := track.Read(rtpBuf)
					if readErr != nil {
						log.Printf("%s[ERROR][OnTrack][Audio] Error reading from track: SessionID=%s, TrackID=%s, Error=%v", WebRTCGatewayLogPrefix, sessionID, track.ID(), readErr)
						return // Stop processing this track
					}

					// --- OPUS DECODING & PCM PREPARATION (Conceptual) ---
					// 1. Decode Opus:
					//    The rtpBuf[:n] contains the raw Opus packet.
					//    pcmBuf := make([]int16, frameSize*opusChannels) // frameSize depends on Opus packet (e.g., 960 for 20ms at 48kHz)
					//    samplesDecoded, err := decoder.Decode(rtpBuf[:n], pcmBuf)
					//    if err != nil {
					//        log.Printf("%s[ERROR][OnTrack][Audio] Opus decode error for SessionID %s: %v", WebRTCGatewayLogPrefix, sessionID, err)
					//        continue
					//    }
					//
					// 2. Channel Selection/Mixing (if stereo Opus):
					//    monoPcmBuf := make([]int16, samplesDecoded) // Assuming we take the first channel or mix down
					//    for i := 0; i < samplesDecoded; i++ {
					//        monoPcmBuf[i] = pcmBuf[i*opusChannels] // Example: take first channel
					//    }
					//
					// 3. Resampling (e.g., from 48kHz Opus output to 16kHz for STT):
					//    resampledPcmBuf := resample(monoPcmBuf, opusSampleRate, 16000) // Conceptual resample function
					//
					// 4. Convert int16 PCM to byte slice (Linear16 is typically 16-bit signed little-endian):
					//    outputBytes := make([]byte, len(resampledPcmBuf)*2)
					//    for i, s := range resampledPcmBuf {
					//        binary.LittleEndian.PutUint16(outputBytes[i*2:], uint16(s))
					//    }
					//
					// For this placeholder stage, we are NOT performing actual decoding.
					// We continue to use the raw rtpBuf data but mark it as LINEAR16
					// for pipeline flow testing, assuming 16kHz mono based on downstream STT needs.
					// The actual data in rtpBuf[:n] is still Opus.
					audioDataForSegment := rtpBuf[:n]
					// --- END CONCEPTUAL DECODING ---

					segment := &real_time_processing.AudioSegment{
						SessionId:      sessionID,
						Timestamp:      time.Now().UnixMilli(),
						// AudioFormat is now LINEAR16, assuming the conceptual decoding above
						// would output 16kHz, 16-bit mono PCM.
						AudioFormat:    real_time_processing.AudioFormat_LINEAR16,
						SequenceNumber: sequenceNumber,
						Data:           audioDataForSegment, // This is still raw Opus data in this placeholder
						IsFinal:        false,
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
