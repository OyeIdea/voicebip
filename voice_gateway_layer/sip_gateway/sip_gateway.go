package sip_gateway

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // For grpc.WithTransportCredentials
	"revovoiceai/voice_gateway_layer/internal/protos/real_time_processing"
)

// DummySessionManagerClient is a placeholder implementation for SessionManagerClient.
// This can be removed if no longer used by tests directly, or kept for basic tests.
type DummySessionManagerClient struct{}

func (d *DummySessionManagerClient) RegisterSession(callID string, sessionType string, details map[string]string) error {
	log.Printf("%s[INFO][DummySMClient][RegisterSession] CallID: %s, Type: %s, Details: %v", SipGatewayLogPrefix, callID, sessionType, details)
	return nil
}

func (d *DummySessionManagerClient) UpdateSessionState(callID string, state string) error {
	log.Printf("%s[INFO][DummySMClient][UpdateSessionState] CallID: %s, State: %s", SipGatewayLogPrefix, callID, state)
	return nil
}

func (d *DummySessionManagerClient) DeregisterSession(callID string) error {
	log.Printf("%s[INFO][DummySMClient][DeregisterSession] CallID: %s", SipGatewayLogPrefix, callID)
	return nil
}

// ErrSIPParseError represents an error during SIP message parsing.
var ErrSIPParseError = errors.New("SIP message parse error")

// parseSIPRequest parses a raw SIP message.
func parseSIPRequest(data []byte) (*SIPRequest, error) {
	// log.Printf("%s[DEBUG][parseSIPRequest] Attempting to parse SIP request data: %s", SipGatewayLogPrefix, string(data))
	req := &SIPRequest{
		Headers: make(map[string]string),
	}

	// Find the end of the request line
	firstLineEnd := bytes.Index(data, []byte("\r\n"))
	if firstLineEnd == -1 {
		errMsg := "no CRLF found for request line"
		log.Printf("%s[ERROR][parseSIPRequest] Invalid SIP request: %s", SipGatewayLogPrefix, errMsg)
		return nil, fmt.Errorf("%w: %s", ErrSIPParseError, errMsg)
	}
	requestLine := string(data[:firstLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		errMsg := fmt.Sprintf("invalid SIP request line format: '%s'", requestLine)
		log.Printf("%s[ERROR][parseSIPRequest] %s", SipGatewayLogPrefix, errMsg)
		return nil, fmt.Errorf("%w: %s", ErrSIPParseError, errMsg)
	}
	req.Method = parts[0]

	parsedURI, err := parseURI(parts[1])
	if err != nil {
		log.Printf("%s[ERROR][parseSIPRequest] Failed to parse URI '%s': %v", SipGatewayLogPrefix, parts[1], err)
		return nil, fmt.Errorf("%w: failed to parse URI: %v", ErrSIPParseError, err)
	}
	req.RequestURI = parsedURI
	req.SIPVersion = parts[2]

	// Parse headers
	headerData := data[firstLineEnd+2:]
	currentPos := 0
	for currentPos < len(headerData) {
		lineEnd := bytes.Index(headerData[currentPos:], []byte("\r\n"))
		if lineEnd == -1 {
			if len(bytes.TrimSpace(headerData[currentPos:])) == 0 {
				break
			}
			errMsg := "invalid SIP headers: missing CRLF after a header line"
			log.Printf("%s[ERROR][parseSIPRequest] %s. Data: %s", SipGatewayLogPrefix, errMsg, string(headerData[currentPos:]))
			return nil, fmt.Errorf("%w: %s", ErrSIPParseError, errMsg)
		}
		if lineEnd == 0 {
			bodyStart := currentPos + 2
			if bodyStart < len(headerData) {
				req.Body = headerData[bodyStart:]
			}
			break
		}

		line := string(headerData[currentPos : currentPos+lineEnd])
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headerName := strings.TrimSpace(headerParts[0])
			headerValue := strings.TrimSpace(headerParts[1])
			req.Headers[headerName] = headerValue
		} else {
			log.Printf("%s[WARN][parseSIPRequest] Malformed header line (no colon): '%s'", SipGatewayLogPrefix, line)
		}
		currentPos += lineEnd + 2
	}
	// log.Printf("%s[DEBUG][parseSIPRequest] Successfully parsed SIP request: Method=%s, URI=%s, HeadersCount=%d", SipGatewayLogPrefix, req.Method, req.RequestURI, len(req.Headers))
	return req, nil
}

// parseURI is a simplified URI parser for SIP.
func parseURI(uriStr string) (*net.UDPAddr, error) {
	// log.Printf("%s[DEBUG][parseURI] Attempting to parse URI: %s", SipGatewayLogPrefix, uriStr)
	trimmedURI := strings.TrimPrefix(uriStr, "<")
	trimmedURI = strings.TrimSuffix(trimmedURI, ">")
	if strings.HasPrefix(trimmedURI, "sip:") {
		trimmedURI = trimmedURI[4:]
	}

	atParts := strings.Split(trimmedURI, "@")
	hostPortPart := atParts[len(atParts)-1]

	host, portStr, err := net.SplitHostPort(hostPortPart)
	if err != nil {
		// log.Printf("%s[DEBUG][parseURI] SplitHostPort failed for '%s' (err: %v), trying with default port 5060", SipGatewayLogPrefix, hostPortPart, err)
		addr, resolveErr := net.ResolveUDPAddr("udp", net.JoinHostPort(hostPortPart, "5060"))
		if resolveErr != nil {
			log.Printf("%s[ERROR][parseURI] Failed to resolve address '%s' with default port: %v (original SplitHostPort err: %v)", SipGatewayLogPrefix, hostPortPart, resolveErr, err)
			return nil, fmt.Errorf("failed to parse host/port or resolve address '%s': %w (original: %v)", hostPortPart, resolveErr, err)
		}
		// log.Printf("%s[DEBUG][parseURI] Successfully parsed URI '%s' with default port: %s", SipGatewayLogPrefix, uriStr, addr.String())
		return addr, nil
	}

	port, convErr := strconv.Atoi(portStr)
	if convErr != nil {
		log.Printf("%s[ERROR][parseURI] Invalid port in URI '%s': %v", SipGatewayLogPrefix, uriStr, convErr)
		return nil, fmt.Errorf("invalid port in URI '%s': %w", uriStr, convErr)
	}

	ipAddr, resolveErr := net.ResolveIPAddr("ip", host)
	if resolveErr != nil {
		ip := net.ParseIP(host)
		if ip == nil {
			log.Printf("%s[ERROR][parseURI] Could not resolve host '%s': %v", SipGatewayLogPrefix, host, resolveErr)
			return nil, fmt.Errorf("could not resolve host '%s': %w", host, resolveErr)
		}
		// log.Printf("%s[DEBUG][parseURI] Host '%s' is an IP, using it directly.", SipGatewayLogPrefix, host)
		return &net.UDPAddr{IP: ip, Port: port}, nil
	}
	// log.Printf("%s[DEBUG][parseURI] Successfully parsed URI '%s': %s:%d", SipGatewayLogPrefix, uriStr, ipAddr.IP.String(), port)
	return &net.UDPAddr{IP: ipAddr.IP, Port: port}, nil
}

// generateResponse creates a string representation of a SIP response.
func generateResponse(statusCode int, statusText string, request *SIPRequest, otherHeaders map[string]string) string {
	var response strings.Builder
	response.WriteString(fmt.Sprintf("%s %d %s\r\n", request.SIPVersion, statusCode, statusText))

	// Essential headers from request
	if via, ok := request.Headers[HeaderVia]; ok {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", HeaderVia, via))
	}
	if to, ok := request.Headers[HeaderTo]; ok {
		response.WriteString(fmt.Sprintf("%s: %s", HeaderTo, to))
		if statusCode != 100 && statusCode != 180 && statusCode != 183 { // Add tag for non-provisional responses to To header
			if !strings.Contains(to, ";tag=") { // Avoid adding if tag already present (e.g. from request)
				 response.WriteString(fmt.Sprintf(";tag=%d", time.Now().UnixNano()%100000))
			}
		}
		response.WriteString("\r\n")
	}
	if from, ok := request.Headers[HeaderFrom]; ok {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", HeaderFrom, from))
	}
	if callID, ok := request.Headers[HeaderCallID]; ok {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", HeaderCallID, callID))
	}
	if cseq, ok := request.Headers[HeaderCSeq]; ok {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", HeaderCSeq, cseq)) // CSeq is copied from request
	}
	
	// Add other headers provided
	for key, value := range otherHeaders {
		response.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	
	var body string
	if statusCode == 200 && request.Method == "INVITE" {
		// Basic SDP placeholder
		body = "v=0\r\n" +
			"o=- 0 0 IN IP4 127.0.0.1\r\n" + // Replace with actual IP later
			"s=-\r\n" +
			"c=IN IP4 127.0.0.1\r\n" + // Replace with actual IP later
			"t=0 0\r\n" +
			"m=audio 9 RTP/AVP 0\r\n" + // Port 9 is a common placeholder, replace later
			"a=rtpmap:0 PCMU/8000\r\n"
		response.WriteString(fmt.Sprintf("%s: application/sdp\r\n", HeaderContentType))
	}

	response.WriteString(fmt.Sprintf("%s: %d\r\n", HeaderContentLength, len(body)))
	response.WriteString("\r\n") // End of headers
	if len(body) > 0 {
		response.WriteString(body)
	}
	return response.String()
}

// StartSIPGateway starts the SIP gateway and listens for incoming messages.
func StartSIPGateway(cfg SIPConfig, smClient SessionManagerClient) {
	listenAddr := fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ListenPort)
	addr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP port: %v", err)
	}
	defer conn.Close()
	log.Printf("SIP Gateway listening on %s\n", listenAddr)

	buffer := make([]byte, 2048) // Buffer for incoming data

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v\n", err)
			continue
		}

		log.Printf("Received %d bytes from %s:\n%s\n", n, remoteAddr.String(), string(buffer[:n]))

		request, err := parseSIPRequest(buffer[:n])
		if err != nil {
			log.Printf("Error parsing SIP request: %v\n", err)
			// Consider sending a 400 Bad Request if parsing fails significantly
			continue
		}

		// Simple state machine based on method
		switch request.Method {
		case "INVITE":
			log.Printf("Received INVITE for Call-ID: %s\n", request.Headers[HeaderCallID])
			
			// Send 100 Trying
			tryingResponse := generateResponse(100, "Trying", request, nil)
			_, err = conn.WriteToUDP([]byte(tryingResponse), remoteAddr)
			if err != nil {
				log.Printf("Error sending 100 Trying: %v\n", err)
				continue
			}
			log.Printf("Sent 100 Trying to %s\n", remoteAddr.String())

			// Register session with the session manager
			sessionDetails := map[string]string{
				"from":           request.Headers[HeaderFrom],
				"to":             request.Headers[HeaderTo],
				"remote_address": remoteAddr.String(),
			}
			err = smClient.RegisterSession(request.Headers[HeaderCallID], "SIP", sessionDetails)
			if err != nil {
				log.Printf("Error registering session with Session Manager for Call-ID %s: %v", request.Headers[HeaderCallID], err)
				// Send 500 Server Internal Error if registration fails
				errorResponse := generateResponse(500, "Server Internal Error - Session Registration Failed", request, nil)
				_, writeErr := conn.WriteToUDP([]byte(errorResponse), remoteAddr)
				if writeErr != nil {
					log.Printf("Error sending 500 Server Internal Error: %v", writeErr)
				}
				continue // Skip further processing for this INVITE
			}
			log.Printf("Successfully registered session for Call-ID: %s with Session Manager", request.Headers[HeaderCallID])

			// Send 180 Ringing (after a slight delay to simulate processing)
			// In a real scenario, this would happen upon some external trigger or internal logic.
			time.Sleep(100 * time.Millisecond) 
			ringingResponse := generateResponse(180, "Ringing", request, nil)
			_, err = conn.WriteToUDP([]byte(ringingResponse), remoteAddr)
			if err != nil {
				log.Printf("Error sending 180 Ringing: %v\n", err)
				continue
			}
			log.Printf("Sent 180 Ringing to %s\n", remoteAddr.String())

			// Simulate call being answered after a delay
			// In a real scenario, this would be triggered by the callee picking up.
			time.Sleep(500 * time.Millisecond)

			// Simulate receiving RTP packets
			// In a real implementation, this would be a separate goroutine or callback handling actual RTP data.
			// For now, we'll just simulate a few packets after the call is established.
			go func(callID string, remoteAddr *net.UDPAddr) {
				// This is a placeholder for actual RTP packet handling.
				// In a real system, you'd receive RTP packets from remoteAddr (or a negotiated media port).
				// For now, let's simulate a few packets.
				for i := 0; i < 5; i++ { // Simulate 5 RTP packets
					time.Sleep(20 * time.Millisecond) // Simulate RTP packet interval

					// Create a dummy RTP payload
					dummyRTPPayload := []byte(fmt.Sprintf("RTP Packet %d for %s", i+1, callID))

					// Populate AudioSegment
					segment := &real_time_processing.AudioSegment{
						SessionId:      callID,
						Timestamp:      time.Now().UnixMilli(),
						AudioFormat:    real_time_processing.AudioFormat_PCMU, // Assuming PCMU for this example
						SequenceNumber: uint32(i + 1),
						Data:           dummyRTPPayload,
						IsFinal:        false,
					}
					sendAudioSegmentToSdm(segment)
				}
			}(request.Headers[HeaderCallID], remoteAddr)

			// Update session state to "active" in session manager
			err = smClient.UpdateSessionState(request.Headers[HeaderCallID], "active")
			if err != nil {
				log.Printf("Error updating session state to active for Call-ID %s: %v", request.Headers[HeaderCallID], err)
				// Non-fatal for the call flow, but log it. The call can proceed.
			} else {
				log.Printf("Successfully updated session state to active for Call-ID: %s", request.Headers[HeaderCallID])
			}

			okHeaders := map[string]string{
				HeaderContact: fmt.Sprintf("<sip:%s:%d>", cfg.ListenAddress, cfg.ListenPort), // Our contact
			}
			okResponse := generateResponse(200, "OK", request, okHeaders)
			_, err = conn.WriteToUDP([]byte(okResponse), remoteAddr)
			if err != nil {
				log.Printf("Error sending 200 OK for INVITE: %v\n", err)
				continue
			}
			log.Printf("Sent 200 OK for INVITE to %s\n", remoteAddr.String())
			log.Printf("RTP stream should be established for Call-ID: %s (placeholder)\n", request.Headers[HeaderCallID])

		case "ACK":
			log.Printf("Received ACK for Call-ID: %s\n", request.Headers[HeaderCallID])
			// The ACK confirms the 200 OK for INVITE.
			// Session state is already 'active'. No further state update needed here from ACK itself.
			// RTP session would be active.

		case "BYE":
			log.Printf("Received BYE for Call-ID: %s\n", request.Headers[HeaderCallID])

			// Deregister session from the session manager
			err = smClient.DeregisterSession(request.Headers[HeaderCallID])
			if err != nil {
				log.Printf("Error deregistering session with Session Manager for Call-ID %s: %v", request.Headers[HeaderCallID], err)
				// Non-fatal for BYE processing, proceed to send 200 OK for BYE.
			} else {
				log.Printf("Successfully deregistered session for Call-ID: %s with Session Manager", request.Headers[HeaderCallID])
			}

			// Send 200 OK for BYE
			byeOkResponse := generateResponse(200, "OK", request, nil)
			_, err = conn.WriteToUDP([]byte(byeOkResponse), remoteAddr)
			if err != nil {
				log.Printf("Error sending 200 OK for BYE: %v\n", err)
				continue
			}
			log.Printf("Sent 200 OK for BYE to %s\n", remoteAddr.String())
			log.Printf("RTP stream should be terminated for Call-ID: %s (placeholder)\n", request.Headers[HeaderCallID])

		default:
			log.Printf("Received unhandled method: %s\n", request.Method)
			// Optionally send a 405 Method Not Allowed or 501 Not Implemented
			// For now, just log.
		}
	}
}

// main function to run the gateway (example usage)
// main function to run the gateway (example usage)
// func main() {
// 	cfg := LoadConfig()
// 	// Initialize the actual HTTP client for the session manager
// 	smClient := NewHTTPMeetingSessionManagerClient(cfg.SessionManagerAPIEndpoint)
// 	StartSIPGateway(cfg, smClient)
// }

// sendAudioSegmentToSdm logs the audio segment.
// In a real system, this would send the segment to the Stream Dialog Manager (SDM) via gRPC.
func sendAudioSegmentToSdm(segment *real_time_processing.AudioSegment) {
	// Target for StreamingDataManager gRPC service
	sdmAddress := "localhost:50051"

	// Set up a connection to the server.
	// Using grpc.WithTransportCredentials(insecure.NewCredentials()) instead of grpc.WithInsecure() for newer gRPC versions.
	conn, err := grpc.Dial(sdmAddress, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Printf("SIPGateway: Failed to connect to StreamingDataManager at %s: %v", sdmAddress, err)
		return
	}
	defer conn.Close()

	client := real_time_processing.NewStreamIngestClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // 5-second timeout for the call
	defer cancel()

	// log.Printf("SIPGateway: Sending AudioSegment to SDM: SID=%s, Seq=%d", segment.SessionId, segment.SequenceNumber)
	response, err := client.IngestAudioSegment(ctx, segment)
	if err != nil {
		log.Printf("SIPGateway: Error calling IngestAudioSegment on StreamingDataManager: %v", err)
		return
	}

	log.Printf("SIPGateway: Received response from StreamingDataManager: SID=%s, Seq=%d, Status='%s'", response.GetSessionId(), response.GetSequenceNumber(), response.GetStatusMessage())
}
