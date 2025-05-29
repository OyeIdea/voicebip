package webrtc_gateway

import (
	"encoding/json"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Session Manager HTTP Server (similar to the one in sip_gateway_test)
func mockWebRTCSessionManagerServer(t *testing.T, expectedType string, failRegister bool, failUpdateState bool, failDeregister bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock SM for WebRTC GW received request: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodPost:
			if strings.HasSuffix(r.URL.Path, "/sessions") {
				if failRegister {
					http.Error(w, "Simulated SM registration error", http.StatusInternalServerError)
					return
				}
				var req CreateSessionRequest // Assumes this struct is accessible or mirrored
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err, "Mock SM failed to decode register request")
				require.NotEmpty(t, req.ID, "Mock SM: RegisterSession ID is empty")
				assert.Equal(t, expectedType, req.Type, "Mock SM: Unexpected session type for RegisterSession")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(map[string]interface{}{"id": req.ID, "type": req.Type, "state": "pending", "details": req.Details})
				return
			}
		case http.MethodPut:
			if strings.HasSuffix(r.URL.Path, "/state") {
				if failUpdateState {
					http.Error(w, "Simulated SM update state error", http.StatusInternalServerError)
					return
				}
				var req UpdateStateRequest // Assumes this struct is accessible or mirrored
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err, "Mock SM failed to decode update state request")
				assert.Equal(t, "active", req.State, "Mock SM: Expected state 'active' for update")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"id": "someID", "state": req.State}) // ID doesn't matter much for response here
				return
			}
		case http.MethodDelete:
			if strings.Contains(r.URL.Path, "/sessions/") {
				if failDeregister {
					http.Error(w, "Simulated SM deregistration error", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.NotFound(w, r)
	}))
}

func TestSignalMessageJSON(t *testing.T) {
	offerPayload := `{"sdp": "v=0...", "type": "offer"}`
	msg := SignalMessage{
		Type:    "offer",
		Payload: offerPayload,
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal SignalMessage: %v", err)
	}

	var unmarshalledMsg SignalMessage
	err = json.Unmarshal(jsonData, &unmarshalledMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal SignalMessage: %v", err)
	}

	if unmarshalledMsg.Type != msg.Type {
		t.Errorf("Expected Type %s, got %s", msg.Type, unmarshalledMsg.Type)
	}
	if unmarshalledMsg.Payload != msg.Payload {
		t.Errorf("Expected Payload %s, got %s", msg.Payload, unmarshalledMsg.Payload)
	}
}

// TestWebSocketSignaling_OfferAnswer_WithSessionManager tests the signaling flow
// including interactions with a mocked Session Manager.
func TestWebSocketSignaling_OfferAnswer_WithSessionManager(t *testing.T) {
	tests := []struct {
		name                    string
		failSMRegister          bool
		failSMUpdateState       bool // Note: UpdateState failure currently only logs, doesn't stop flow
		expectWebSocketErrorMsg string // Expected error message type from WebSocket if SM registration fails
	}{
		{"Success", false, false, ""},
		{"SMRegisterFails", true, false, "error"}, // Expect an error signal if SM registration fails
		// {"SMUpdateStateFails", false, true, ""}, // Update state failure is logged, does not send WS error
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoadConfig() // Use default config which now includes SM endpoint
			mockSM := mockWebRTCSessionManagerServer(t, "WebRTC", tt.failSMRegister, tt.failSMUpdateState, false)
			defer mockSM.Close()
			cfg.SessionManagerAPIEndpoint = mockSM.URL // Override with mock server URL

			// smClient will be initialized inside StartWebRTCGateway, which then passes it to HandleWebSocketConnections
			// For this test, we need to simulate that. We'll create the client and pass it.
			smHttpClient := NewHTTPMeetingSessionManagerClient(cfg.SessionManagerAPIEndpoint)

			peerConnectionConfig := webrtc.Configuration{
				ICEServers: []webrtc.ICEServer{{URLs: cfg.StunServers}},
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				HandleWebSocketConnections(w, r, peerConnectionConfig, smHttpClient)
			}))
			defer server.Close()

			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
			ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err, "Failed to connect to WebSocket server")
			defer ws.Close()

			log.Println("Test client connected to WebSocket server for test:", tt.name)

			// --- Simulate Client Sending an Offer ---
			mockOfferSDP := webrtc.SessionDescription{
				Type: webrtc.SDPTypeOffer,
				SDP:  "v=0\r\no=- 12345 67890 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\nm=audio 9 RTP/AVP 0\r\na=rtpmap:0 PCMU/8000\r\n",
			}
			offerPayloadBytes, _ := json.Marshal(mockOfferSDP)
			offerMsg := SignalMessage{Type: "offer", Payload: string(offerPayloadBytes)}
			err = ws.WriteJSON(offerMsg)
			require.NoError(t, err, "Failed to send offer message")
			log.Println("Test client sent offer for test:", tt.name)

			// --- Client Receives Messages ---
			var answerReceived bool
			var errorReceived bool
			var receivedErrorPayload string

			// Read messages with a timeout
			for i := 0; i < 5; i++ { // Try to read a few messages (answer, candidates, or error)
				var receivedMsg SignalMessage
				ws.SetReadDeadline(time.Now().Add(2 * time.Second))
				err := ws.ReadJSON(&receivedMsg)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						log.Println("Test client ReadJSON timeout for test:", tt.name)
						break // Timeout is expected if no more messages
					}
					log.Printf("Test client error reading message for test %s: %v", tt.name, err)
					break
				}

				log.Printf("Test client received message: Type=%s for test: %s", receivedMsg.Type, tt.name)

				switch receivedMsg.Type {
				case "answer":
					answerReceived = true
					// Basic validation of answer
					var answerSDP webrtc.SessionDescription
					err := json.Unmarshal([]byte(receivedMsg.Payload), &answerSDP)
					require.NoError(t, err, "Failed to unmarshal answer SDP for test: %s", tt.name)
					assert.Equal(t, webrtc.SDPTypeAnswer, answerSDP.Type, "Expected SDPTypeAnswer for test: %s", tt.name)
				case "candidate":
					log.Println("Test client received ICE candidate for test:", tt.name)
				case "error":
					errorReceived = true
					receivedErrorPayload = receivedMsg.Payload
					log.Printf("Test client received error from server: %s for test: %s", receivedErrorPayload, tt.name)
					// If we expect an error, we can break here. Otherwise, it's a test failure.
					if tt.expectWebSocketErrorMsg != "" {
						goto EndReadLoop // Using goto to break outer loop for readability in this case
					}
				}
			}
			EndReadLoop:

			if tt.expectWebSocketErrorMsg != "" {
				assert.True(t, errorReceived, "Expected a WebSocket error message but didn't get one for test: %s", tt.name)
				assert.Contains(t, receivedErrorPayload, "Session registration failed", "Expected error payload to indicate session registration failure for test: %s", tt.name)
				assert.False(t, answerReceived, "Did not expect an answer message when SM registration fails for test: %s", tt.name)
			} else {
				assert.False(t, errorReceived, "Received unexpected WebSocket error: %s for test: %s", receivedErrorPayload, tt.name)
				assert.True(t, answerReceived, "Did not receive an answer message for test: %s", tt.name)
			}
			log.Println("TestWebSocketSignaling_OfferAnswer_WithSessionManager completed test:", tt.name)
		})
	}
}

// TestStartWebRTCGateway_Conceptual is a placeholder for testing server startup.
// Actual testing of a running HTTP server is more of an integration test.
func TestStartWebRTCGateway_Conceptual(t *testing.T) {
	// This test is mostly conceptual.
	// To truly test StartWebRTCGateway, you'd run it in a goroutine,
	// then make HTTP/WebSocket requests to its port, and then shut it down.
	// This is complex for a unit test.

	// We can verify that LoadConfig works as expected.
	cfg := LoadConfig()
	if cfg.SignalPort == 0 {
		t.Error("LoadConfig returned a zero SignalPort, which is unlikely to be correct.")
	}
	if cfg.ListenAddress == "" {
		t.Error("LoadConfig returned an empty ListenAddress.")
	}
	if len(cfg.StunServers) == 0 {
		t.Log("Warning: LoadConfig returned no STUN servers. This might be intended for some environments but often required.")
	}

	t.Log("TestStartWebRTCGateway_Conceptual: Verified LoadConfig. Full server start test is more of an integration test.")
}
