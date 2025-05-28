package sip_gateway

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

// DummySessionManagerClient is a placeholder implementation for SessionManagerClient.
type DummySessionManagerClient struct{}

func (d *DummySessionManagerClient) RegisterSession(callID string, sessionDetails map[string]string) error {
	log.Printf("SESSION_MANAGER_CLIENT: Registering session for Call-ID: %s, Details: %v\n", callID, sessionDetails)
	return nil
}

func (d *DummySessionManagerClient) DeregisterSession(callID string) error {
	log.Printf("SESSION_MANAGER_CLIENT: Deregistering session for Call-ID: %s\n", callID)
	return nil
}

// parseSIPRequest parses a raw SIP message.
func parseSIPRequest(data []byte) (*SIPRequest, error) {
	req := &SIPRequest{
		Headers: make(map[string]string),
	}
	
	// Find the end of the request line
	firstLineEnd := bytes.Index(data, []byte("\r\n"))
	if firstLineEnd == -1 {
		return nil, fmt.Errorf("invalid SIP request: no CRLF found for request line")
	}
	requestLine := string(data[:firstLineEnd])
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid SIP request line: %s", requestLine)
	}
	req.Method = parts[0]
	// For simplicity, we're not fully parsing the URI struct here.
	// req.RequestURI = &url.URL{Path: parts[1]} 
	req.RequestURI, _ = parseURI(parts[1]) // Store as string for now
	req.SIPVersion = parts[2]

	// Parse headers
	headerSection := data[firstLineEnd+2:]
	currentLineStart := 0
	for currentLineStart < len(headerSection) {
		lineEnd := bytes.Index(headerSection[currentLineStart:], []byte("\r\n"))
		if lineEnd == -1 { // Should end with an empty line
			if len(bytes.TrimSpace(headerSection[currentLineStart:])) == 0 {
				break // End of headers
			}
			return nil, fmt.Errorf("invalid SIP headers: missing CRLF")
		}
		if lineEnd == 0 { // Empty line indicates end of headers
			// Check for body
			if len(headerSection) > currentLineStart+2 {
				req.Body = headerSection[currentLineStart+2:]
			}
			break
		}

		line := string(headerSection[currentLineStart : currentLineStart+lineEnd])
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headerName := strings.TrimSpace(headerParts[0])
			headerValue := strings.TrimSpace(headerParts[1])
			req.Headers[headerName] = headerValue
		} else {
			// Could be a multi-line header, not handling for simplicity
		}
		currentLineStart += lineEnd + 2 // Move to the start of the next line
	}
	return req, nil
}

// parseURI is a simplified URI parser for SIP.
func parseURI(uriStr string) (*net.UDPAddr, error) {
    // This is a very basic parser, assuming sip:user@host:port or sip:host:port
    // A proper implementation would use a full URI parsing library.
    trimmed := strings.TrimPrefix(uriStr, "<")
    trimmed = strings.TrimSuffix(trimmed, ">")
    if strings.HasPrefix(trimmed, "sip:") {
        trimmed = trimmed[4:]
    }
    
    parts := strings.Split(trimmed, "@")
    hostPortStr := parts[len(parts)-1] // Take the last part, which should be host:port

    host, portStr, err := net.SplitHostPort(hostPortStr)
    if err != nil {
        // If SplitHostPort fails, maybe it's just a host without a port (default 5060)
        // Or it could be an IP address without a port.
        // For simplicity, let's try to resolve it as a host and assume default port
        // This is not robust for production.
        addr, resolveErr := net.ResolveUDPAddr("udp", net.JoinHostPort(hostPortStr, "5060"))
        if resolveErr != nil {
            return nil, fmt.Errorf("failed to parse host/port or resolve address %s: %v, original error: %v", hostPortStr, resolveErr, err)
        }
        return addr, nil
    }

    port, err := strconv.Atoi(portStr)
    if err != nil {
        return nil, fmt.Errorf("invalid port in URI %s: %v", uriStr, err)
    }

    // Resolve IP address for the host
    ipAddr, err := net.ResolveIPAddr("ip", host)
    if err != nil {
        // Try to see if host itself is an IP
        ip := net.ParseIP(host)
        if ip == nil {
             return nil, fmt.Errorf("could not resolve host %s: %v", host, err)
        }
         return &net.UDPAddr{IP: ip, Port: port}, nil
    }
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

			// Conceptual session registration
			smClient.RegisterSession(request.Headers[HeaderCallID], map[string]string{
				"from": request.Headers[HeaderFrom],
				"to":   request.Headers[HeaderTo],
			})

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
			// Usually, no response is sent for ACK in this context.
			// The ACK confirms the 200 OK for INVITE.
			// RTP session would be active.

		case "BYE":
			log.Printf("Received BYE for Call-ID: %s\n", request.Headers[HeaderCallID])
			
			// Conceptual session deregistration
			smClient.DeregisterSession(request.Headers[HeaderCallID])

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
// func main() {
// 	cfg := LoadConfig()
// 	var smClient SessionManagerClient = &DummySessionManagerClient{} // Use dummy client
// 	StartSIPGateway(cfg, smClient)
// }
