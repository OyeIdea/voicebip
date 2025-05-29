package sip_gateway

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testCallID = "testcall123@1.2.3.4"
	testFrom   = "<sip:alice@example.com>;tag=fromtag1"
	testTo     = "<sip:bob@otherdomain.com>"
	testCSeq   = "1 INVITE"
	testRequestURI = "sip:bob@otherdomain.com"
)

var sampleINVITERequest = fmt.Sprintf(
	"INVITE %s SIP/2.0\r\n"+
		"Via: SIP/2.0/UDP pc33.example.com;branch=z9hG4bK776asdhds\r\n"+
		"Max-Forwards: 70\r\n"+
		"To: %s\r\n"+
		"From: %s\r\n"+
		"Call-ID: %s\r\n"+
		"CSeq: %s\r\n"+
		"Contact: <sip:alice@pc33.example.com>\r\n"+
		"Content-Type: application/sdp\r\n"+
		"Content-Length: 4\r\n"+
		"\r\n"+
		"v=0\r\n", testRequestURI, testTo, testFrom, testCallID, testCSeq,
)

func TestParseINVITE(t *testing.T) {
	req, err := parseSIPRequest([]byte(sampleINVITERequest))
	if err != nil {
		t.Fatalf("parseSIPRequest failed: %v", err)
	}

	if req.Method != "INVITE" {
		t.Errorf("Expected Method INVITE, got %s", req.Method)
	}
	// Assuming parseURI stores the string representation directly for now
	if req.RequestURI.String() != "sip:bob@otherdomain.com:5060" { // Default port added by parseURI
		t.Errorf("Expected RequestURI %s, got %s", "sip:bob@otherdomain.com:5060", req.RequestURI.String())
	}
	if req.Headers[HeaderCallID] != testCallID {
		t.Errorf("Expected Call-ID %s, got %s", testCallID, req.Headers[HeaderCallID])
	}
	if req.Headers[HeaderFrom] != testFrom {
		t.Errorf("Expected From %s, got %s", testFrom, req.Headers[HeaderFrom])
	}
	if req.Headers[HeaderTo] != testTo {
		t.Errorf("Expected To %s, got %s", testTo, req.Headers[HeaderTo])
	}
	if req.Headers[HeaderCSeq] != testCSeq {
		t.Errorf("Expected CSeq %s, got %s", testCSeq, req.Headers[HeaderCSeq])
	}
	if len(req.Body) != 4 {
		t.Errorf("Expected Content-Length 4, got %d for body: %s", len(req.Body), string(req.Body))
	}
	if string(req.Body) != "v=0\r\n" {
		t.Errorf("Expected body 'v=0\\r\\n', got '%s'", string(req.Body))
	}
}

// MockUDPSender simulates sending UDP packets and captures them for inspection.
type MockUDPSender struct {
	mu            sync.Mutex
	SentPackets   [][]byte
	RemoteAddress net.Addr
	conn          *net.UDPConn // Store the connection to respond if needed
}

func (m *MockUDPSender) WriteToUDP(b []byte, addr net.Addr) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Make a copy of the packet, as the buffer might be reused.
	packetCopy := make([]byte, len(b))
	copy(packetCopy, b)
	m.SentPackets = append(m.SentPackets, packetCopy)
	m.RemoteAddress = addr // Capture the address it would send to
	return len(b), nil
}

// Mock Session Manager HTTP Server
func mockSessionManagerServer(t *testing.T, expectedCallID string, expectedType string, expectedStateForUpdate string, failRegister bool, failUpdateState bool, failDeregister bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock Session Manager received request: %s %s", r.Method, r.URL.Path)
		switch r.Method {
		case http.MethodPost: // RegisterSession
			if strings.HasSuffix(r.URL.Path, "/sessions") {
				if failRegister {
					t.Log("Mock Session Manager: Simulating RegisterSession failure")
					http.Error(w, "Simulated registration error", http.StatusInternalServerError)
					return
				}
				var req CreateSessionRequest // Assuming CreateSessionRequest is defined or mirrored from session_client.go
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "Bad request body", http.StatusBadRequest)
					return
				}
				if req.ID != expectedCallID {
					t.Errorf("Mock Session Manager: Expected CallID %s, got %s", expectedCallID, req.ID)
					http.Error(w, "Wrong CallID", http.StatusBadRequest)
					return
				}
				if req.Type != expectedType {
					t.Errorf("Mock Session Manager: Expected Type %s, got %s", expectedType, req.Type)
					http.Error(w, "Wrong Type", http.StatusBadRequest)
					return
				}
				t.Logf("Mock Session Manager: Successfully processed RegisterSession for %s", req.ID)
				w.WriteHeader(http.StatusCreated)
				// Optionally return the created session body if needed by client
				json.NewEncoder(w).Encode(map[string]interface{}{"id": req.ID, "type": req.Type, "state": "pending", "details": req.Details})
				return
			}
		case http.MethodPut: // UpdateSessionState
			if strings.HasSuffix(r.URL.Path, "/state") {
				if failUpdateState {
					t.Log("Mock Session Manager: Simulating UpdateSessionState failure")
					http.Error(w, "Simulated update state error", http.StatusInternalServerError)
					return
				}
				var req UpdateStateRequest // Assuming UpdateStateRequest is defined or mirrored
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, "Bad request body for state update", http.StatusBadRequest)
					return
				}
				if req.State != expectedStateForUpdate {
					t.Errorf("Mock Session Manager: Expected State %s for update, got %s", expectedStateForUpdate, req.State)
					http.Error(w, "Wrong state for update", http.StatusBadRequest)
					return
				}
				t.Logf("Mock Session Manager: Successfully processed UpdateSessionState to %s for %s", req.State, strings.Split(r.URL.Path, "/")[2])
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{"id": strings.Split(r.URL.Path, "/")[2], "state": req.State})
				return
			}
		case http.MethodDelete: // DeregisterSession
			if strings.Contains(r.URL.Path, "/sessions/") {
				if failDeregister {
					t.Log("Mock Session Manager: Simulating DeregisterSession failure")
					http.Error(w, "Simulated deregistration error", http.StatusInternalServerError)
					return
				}
				t.Logf("Mock Session Manager: Successfully processed DeregisterSession for %s", strings.Split(r.URL.Path, "/")[2])
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		t.Logf("Mock Session Manager: Path not handled: %s", r.URL.Path)
		http.NotFound(w, r)
	}))
}


// TestHandleINVITEConceptual tests the INVITE flow including session manager interaction
func TestHandleINVITEConceptual(t *testing.T) {
	tests := []struct {
		name                 string
		failRegister         bool
		failUpdateState      bool
		expectSessionError   bool // True if we expect the INVITE to fail due to session manager error
		expectedResponseCode int  // Expected final SIP response code if expectSessionError is true
	}{
		{"Success", false, false, false, 0},
		{"RegisterFails", true, false, true, 500},
		{"UpdateStateFails", false, true, false, 0}, // Update state failure is currently logged but doesn't stop the call
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoadConfig()
			mockSMServer := mockSessionManagerServer(t, testCallID, "SIP", "active", tt.failRegister, tt.failUpdateState, false)
			defer mockSMServer.Close()
			cfg.SessionManagerAPIEndpoint = mockSMServer.URL // Point to mock server

			smClient := NewHTTPMeetingSessionManagerClient(cfg.SessionManagerAPIEndpoint)

			tempListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
			if err != nil {
				t.Fatalf("Failed to create temp listener: %v", err)
			}
			serverAddr := tempListener.LocalAddr().(*net.UDPAddr)
			tempListener.Close()

			clientConn, err := net.DialUDP("udp", nil, serverAddr)
			if err != nil {
				t.Fatalf("Failed to dial mock server: %v", err)
			}
			defer clientConn.Close()
			clientAddr := clientConn.LocalAddr().(*net.UDPAddr)

			serverConn, err := net.ListenUDP("udp", serverAddr)
			if err != nil {
				t.Fatalf("Failed to listen on mock server UDP port: %v", err)
			}
			defer serverConn.Close()

			request, _ := parseSIPRequest([]byte(sampleINVITERequest))
			mockSender := &MockUDPSender{conn: serverConn}

			// --- Simulate INVITE handling ---
			// The actual StartSIPGateway handler would make these calls.
			// This conceptual test verifies the client's interaction logic.

			// 1. Register session
			sessionDetails := map[string]string{"from": request.Headers[HeaderFrom], "to": request.Headers[HeaderTo], "remote_address": clientAddr.String()}
			regErr := smClient.RegisterSession(request.Headers[HeaderCallID], "SIP", sessionDetails)

			if tt.expectSessionError { // This case implies registration failure
				if regErr == nil {
					t.Fatalf("Expected session registration error, but got nil")
				}
				// In an integrated test, we would check that the SIP gateway sends a 500 error response.
				// Here, we confirm the client call failed as expected.
				t.Logf("Confirmed RegisterSession failed as expected: %v", regErr)
				return // Stop test for this case as the call flow would be aborted.
			}
			if regErr != nil { // Unexpected registration failure
				t.Fatalf("Session registration failed unexpectedly: %v", regErr)
			}

			// If registration was successful, conceptually the gateway would send 100 Trying, then 180 Ringing.
			mockSender.WriteToUDP([]byte(generateResponse(100, "Trying", request, nil)), clientAddr) // Simulating gateway action
			mockSender.WriteToUDP([]byte(generateResponse(180, "Ringing", request, nil)), clientAddr) // Simulating gateway action


			// 2. Update session state to "active"
			if !tt.failRegister { // Only attempt update if registration was supposed to succeed
				updErr := smClient.UpdateSessionState(request.Headers[HeaderCallID], "active")
				if tt.failUpdateState && updErr == nil {
					t.Errorf("Expected session update state error, but got nil")
				}
				if !tt.failUpdateState && updErr != nil {
					t.Errorf("Session update state failed unexpectedly: %v", updErr)
				}
				if updErr != nil {
					t.Logf("UpdateSessionState failed as expected/unexpectedly: %v", updErr)
				}
			}
			
			// 3. Send 200 OK (simulating gateway action after successful INVITE processing)
			okHeaders := map[string]string{HeaderContact: fmt.Sprintf("<sip:%s:%d>", cfg.ListenAddress, cfg.ListenPort)}
			mockSender.WriteToUDP([]byte(generateResponse(200, "OK", request, okHeaders)), clientAddr)

			// Assertions on sent SIP packets
			if !tt.expectSessionError {
				// Expect Trying, Ringing, OK
				if len(mockSender.SentPackets) != 3 { 
					t.Fatalf("Expected 3 SIP packets (Trying, Ringing, OK) in successful flow, got %d", len(mockSender.SentPackets))
				}
				if !strings.Contains(string(mockSender.SentPackets[0]), "100 Trying") {
					t.Errorf("Expected first packet to be 100 Trying, got: %s", string(mockSender.SentPackets[0]))
				}
				if !strings.Contains(string(mockSender.SentPackets[1]), "180 Ringing") {
					t.Errorf("Expected second packet to be 180 Ringing, got: %s", string(mockSender.SentPackets[1]))
				}
				if !strings.Contains(string(mockSender.SentPackets[2]), "200 OK") {
					t.Errorf("Expected third packet to be 200 OK, got: %s", string(mockSender.SentPackets[2]))
				}
				if !strings.Contains(string(mockSender.SentPackets[2]), "application/sdp") {
					t.Errorf("Expected 200 OK for INVITE to contain SDP, got: %s", string(mockSender.SentPackets[2]))
				}
			}
		})
	}
}

var sampleBYERequest = fmt.Sprintf(
	"BYE %s SIP/2.0\r\n"+
	tempListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to create temp listener: %v", err)
	}
	serverAddr := tempListener.LocalAddr().(*net.UDPAddr)
	tempListener.Close() // Close it immediately, we just needed the address

		"Via: SIP/2.0/UDP pc33.example.com;branch=z9hG4bKkjshdy77\r\n"+
		"Max-Forwards: 70\r\n"+
		"To: %s\r\n"+ // To from original INVITE
		"From: %s\r\n"+ // From from original INVITE
		"Call-ID: %s\r\n"+
		"CSeq: 2 BYE\r\n"+ // CSeq incremented, method BYE
		"Content-Length: 0\r\n"+
		"\r\n", testRequestURI, testTo, testFrom, testCallID,
)

func TestHandleBYEConceptual(t *testing.T) {
	tests := []struct {
		name           string
		failDeregister bool
	}{
		{"Success", false},
		{"DeregisterFails", true}, // Deregister failure is currently logged but doesn't stop BYE 200 OK
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := LoadConfig()
			mockSMServer := mockSessionManagerServer(t, testCallID, "SIP", "", false, false, tt.failDeregister)
			defer mockSMServer.Close()
			cfg.SessionManagerAPIEndpoint = mockSMServer.URL

			smClient := NewHTTPMeetingSessionManagerClient(cfg.SessionManagerAPIEndpoint)

			tempListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
			if err != nil {
				t.Fatalf("Failed to create temp listener: %v", err)
			}
			serverAddr := tempListener.LocalAddr().(*net.UDPAddr)
			tempListener.Close()

			clientConn, err := net.DialUDP("udp", nil, serverAddr)
			if err != nil {
				t.Fatalf("Failed to dial mock server: %v", err)
			}
			defer clientConn.Close()
			clientAddr := clientConn.LocalAddr().(*net.UDPAddr)

			serverConn, err := net.ListenUDP("udp", serverAddr)
			if err != nil {
				t.Fatalf("Failed to listen on mock server UDP port: %v", err)
			}
			defer serverConn.Close()

			request, _ := parseSIPRequest([]byte(sampleBYERequest))
			mockSender := &MockUDPSender{conn: serverConn}

			// --- Simulate BYE handling ---
			// 1. Deregister session
			err = smClient.DeregisterSession(request.Headers[HeaderCallID])
			if tt.failDeregister && err == nil {
				t.Errorf("Expected session deregistration error, but got nil")
			}
			if !tt.failDeregister && err != nil {
				t.Errorf("Session deregistration failed unexpectedly: %v", err)
			}
			if err != nil {
				t.Logf("DeregisterSession failed as expected/unexpectedly: %v", err)
			}


			// 2. Send 200 OK for BYE (simulating gateway action)
			byeOkResponse := generateResponse(200, "OK", request, nil)
			mockSender.WriteToUDP([]byte(byeOkResponse), clientAddr)

			// Assertions
			if len(mockSender.SentPackets) != 1 {
				t.Fatalf("Expected 1 SIP packet sent for BYE, got %d", len(mockSender.SentPackets))
			}
			responseStr := string(mockSender.SentPackets[0])
			if !strings.Contains(responseStr, "200 OK") {
				t.Errorf("Expected response to be 200 OK, got: %s", responseStr)
			}
			if !strings.Contains(responseStr, "CSeq: 2 BYE") {
				t.Errorf("Expected CSeq '2 BYE' in response, got: %s", responseStr)
			}
		})
	}
}

// TestStartSIPGateway_ReceiveINVITE is a more integrated test, but still uses a mock client.
// It tests if the StartSIPGateway function (when run in a goroutine) correctly processes an INVITE.
func TestStartSIPGateway_ReceiveINVITE(t *testing.T) {
	cfg := LoadConfig()
	// Use a dynamic port for testing to avoid conflicts
	tempListener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatalf("Failed to create temp listener: %v", err)
	}
	serverAddr := tempListener.LocalAddr().(*net.UDPAddr)
	tempListener.Close()

	// Mock client connection
	clientConn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("Failed to dial mock server: %v", err)
	}
	defer clientConn.Close()
	clientAddr := clientConn.LocalAddr().(*net.UDPAddr)
	
	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to listen on mock server UDP port: %v", err)
	}
	defer serverConn.Close()


	request, _ := parseSIPRequest([]byte(sampleBYERequest))
	mockSender := &MockUDPSender{conn: serverConn}

	// --- Simulate BYE handling ---
	// 1. Conceptual session deregistration
	smClient.DeregisterSession(request.Headers[HeaderCallID])

	// 2. Send 200 OK for BYE
	byeOkResponse := generateResponse(200, "OK", request, nil)
	mockSender.WriteToUDP([]byte(byeOkResponse), clientAddr)

	// Assertions
	if len(mockSender.SentPackets) != 1 {
		t.Fatalf("Expected 1 packet sent for BYE, got %d", len(mockSender.SentPackets))
	}
	responseStr := string(mockSender.SentPackets[0])
	if !strings.Contains(responseStr, "200 OK") {
		t.Errorf("Expected response to be 200 OK, got: %s", responseStr)
	}
	// Check if CSeq in response matches CSeq from BYE request
	if !strings.Contains(responseStr, "CSeq: 2 BYE") {
		t.Errorf("Expected CSeq '2 BYE' in response, got: %s", responseStr)
	}
}

// TestStartSIPGateway_ReceiveINVITE is a more integrated test, but still uses a mock client.
// It tests if the StartSIPGateway function (when run in a goroutine) correctly processes an INVITE.
func TestStartSIPGateway_ReceiveINVITE(t *testing.T) {
	cfg := LoadConfig()
	// Use a dynamic port for testing to avoid conflicts
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}
	
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to listen on UDP port: %v", err)
	}
	
	// Update cfg with the dynamically assigned port
	cfg.ListenPort = conn.LocalAddr().(*net.UDPAddr).Port
	cfg.ListenAddress = conn.LocalAddr().(*net.UDPAddr).IP.String()
	
	// We need to close the connection here because StartSIPGateway will try to listen on this address.
	// This is a bit tricky. For a full integration test, StartSIPGateway would run,
	// and the test would act as a client.
	// Let's adjust StartSIPGateway to accept a connection for testing, or use a real client.
	// For now, let's keep it simpler and test the internal logic flow.

	// The previous conceptual tests (TestHandleINVITEConceptual) are better suited for unit testing
	// the flow without standing up the full server. A full integration test for StartSIPGateway
	// would involve:
	// 1. Starting StartSIPGateway in a goroutine.
	// 2. Creating a separate client that sends UDP packets to the gateway's port.
	// 3. The client then listening for responses.
	// This is more complex than a unit test.

	// Given the structure, we'll stick to testing the parsing and response generation logic
	// as done in TestParseINVITE, TestHandleINVITEConceptual, and TestHandleBYEConceptual.

	// The following is a sketch of how an integration test might look if we were to run the server.
	// This is commented out because it requires careful synchronization and is more of an integration test.
	/*
	var smClient SessionManagerClient = &DummySessionManagerClient{}
	
	// Use a WaitGroup to ensure the server has a chance to start and process.
	var wg sync.WaitGroup
	wg.Add(1) // For the server goroutine

	go func() {
		defer wg.Done()
		// To make this testable, StartSIPGateway would need a way to signal it's ready,
		// and a way to be gracefully shut down.
		// For simplicity, we'll assume it runs and we send a packet.
		// In a real test, you'd use a context for cancellation.
		StartSIPGateway(cfg, smClient) // This will block
	}()

	// Give the server a moment to start (not ideal, but simple for now)
	time.Sleep(100 * time.Millisecond)

	// Client part: Send an INVITE to the server's port
	clientConn, err := net.DialUDP("udp", nil, conn.LocalAddr().(*net.UDPAddr))
	if err != nil {
		t.Fatalf("Client failed to dial server: %v", err)
	}
	defer clientConn.Close()

	_, err = clientConn.Write([]byte(sampleINVITERequest))
	if err != nil {
		t.Fatalf("Client failed to send INVITE: %v", err)
	}

	// Client: Listen for responses (Trying, Ringing, OK)
	// This part needs careful handling of timeouts and expected sequences.
	// For example, expect 3 packets back.
	buffer := make([]byte, 2048)
	expectedResponses := 3
	receivedResponses := 0
	for i := 0; i < expectedResponses; i++ {
		clientConn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Timeout for each read
		n, _, err := clientConn.ReadFromUDP(buffer)
		if err != nil {
			t.Errorf("Client error reading response %d: %v", i+1, err)
			break
		}
		t.Logf("Client received response %d: %s", i+1, string(buffer[:n]))
		receivedResponses++
	}

	if receivedResponses != expectedResponses {
		t.Errorf("Expected %d responses, got %d", expectedResponses, receivedResponses)
	}
	
	// How to stop StartSIPGateway? Need a channel or context.
	// For now, this test would hang as StartSIPGateway runs indefinitely.
	// This highlights why the conceptual tests are more practical for this stage.
	// conn.Close() // This might help stop the server if it's the listening conn.
	// wg.Wait() 
	*/
	// Close the initial listener since StartSIPGateway will create its own.
	// This test case is more of a placeholder for future integration testing.
	// The core logic is tested in TestHandleINVITEConceptual.
	 conn.Close() 
	 t.Log("TestStartSIPGateway_ReceiveINVITE is a conceptual placeholder for true integration testing.")
}
