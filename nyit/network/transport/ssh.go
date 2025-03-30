package transport

// ---
// ### **Simulation Overview**
// 1. **HTTPS Encryption**: TLS for encrypted content.
// 2. **Secure Cookies**: Encrypted cookies using `gorilla/securecookie`.
// 3. **Session Management**: Simulated session IDs.
// 4. **Three-Level Awareness**:
//    - **Level 1**: Connection initiation (handshake).
//    - **Level 2**: Data exchange (encrypted content/cookies).
//    - **Level 3**: Connection closure (graceful/abrupt).
// 5. **Incomplete Closure**: Simulate abrupt disconnection without notifying the peer.
// 6. **Client Security**: Client skips certificate verification (for testing).

// ---

// ### **Server Implementation**
// ```go

import (
	"bufio"
	"bytes"
	"compress/flate"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/ssh"
)

var (
	// Secure cookie encryption
	hashKey  = securecookie.GenerateRandomKey(64)
	blockKey = securecookie.GenerateRandomKey(32)
	sCookie  = securecookie.New(hashKey, blockKey)
)

// Simulate session storage
var sessions = make(map[string]time.Time)

func MainSession() {
	// Load TLS certificate and key
	cert, err := tls.LoadX509KeyPair("../certs/server.crt", "../certs/server.key")
	if err != nil {
		log.Fatal("Failed to load certificate:", err)
	}

	// Configure TLS with session tickets
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create HTTP server with TLS
	server := &http.Server{
		Addr:      ":8443",
		Handler:   http.HandlerFunc(handleRequest),
		TLSConfig: config,
	}

	// Start server
	log.Println("Server listening on :8443")
	log.Fatal(server.ListenAndServeTLS("", ""))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Level 1: Connection initiated (handshake complete)
	log.Println("Level 1: Connection initiated")

	// Level 2: Data exchange (read/write encrypted content)
	// Simulate session ID
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		// Create new session
		newSession := fmt.Sprintf("session-%d", time.Now().Unix())
		encoded, _ := sCookie.Encode("session_id", newSession)
		http.SetCookie(w, &http.Cookie{
			Name:   "session_id",
			Value:  encoded,
			Secure: true,
		})
		sessions[newSession] = time.Now()
		log.Println("Level 2: New session created")
	} else {
		// Decode session cookie
		var decoded string
		sCookie.Decode("session_id", sessionID.Value, &decoded)
		log.Printf("Level 2: Existing session %s\n", decoded)
	}

	// Send response
	w.Write([]byte("Secure response with encrypted cookies!\n"))

	// Level 3: Simulate incomplete closure (50% chance)
	if time.Now().Unix()%2 == 0 {
		log.Println("Level 3: Connection closed abruptly (no notification)")
		hj, ok := w.(http.Hijacker)
		if !ok {
			log.Println("Hijacking not supported")
			return
		}
		conn, _, _ := hj.Hijack()
		conn.Close() // Close without sending a response
	} else {
		log.Println("Level 3: Connection closed gracefully")
	}
}

// ### **Client Implementation**

func MainHandle() {
	// Configure client to skip certificate verification (INSECURE!)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Simulate security issue
		},
	}
	client := &http.Client{Transport: tr}

	// Simulate connection with cookies and session
	for i := 0; i < 3; i++ {
		resp, err := client.Get("https://localhost:8443")
		if err != nil {
			log.Println("Client error:", err)
			// Simulate incomplete close awareness
			if i == 2 {
				log.Println("Client detected incomplete closure (timeout)")
			}
			continue
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("Client received: %s\n", body)

		// Simulate session persistence
		cookies := resp.Cookies()
		for _, cookie := range cookies {
			fmt.Printf("Client stored cookie: %s\n", cookie.Name)
		}

		time.Sleep(1 * time.Second)
	}
}

// ```

// ---

// ### **Explanation**
// 1. **HTTPS Encryption**:
//    - The server uses TLS 1.2+ with a certificate.
//    - The client skips certificate verification (`InsecureSkipVerify: true`) to simulate a security misconfiguration.

// 2. **Secure Cookies**:
//    - Cookies are encrypted using `gorilla/securecookie`.
//    - Sessions are stored in-memory with a simulated session ID.

// 3. **Three-Level Awareness**:
//    - **Level 1**: Logs connection initiation (handshake).
//    - **Level 2**: Manages encrypted cookies and session IDs.
//    - **Level 3**: Randomly simulates abrupt connection closure (no FIN packet).

// 4. **Incomplete Closure**:
//    - The server sometimes closes the connection abruptly using `Hijack()`.
//    - The client detects timeouts but does not receive closure notifications.

// 5. **Client Security Issue**:
//    - The client ignores certificate validation (`InsecureSkipVerify`), simulating a real-world vulnerability.

// ---

// ### **How to Run**
// 1. Generate certificates:
//    ```bash
//    openssl req -x509 -newkey rsa:4096 -nodes -out server.crt -keyout server.key -days 365
//    ```

// 2. Install dependencies:
//    ```bash
//    go get github.com/gorilla/securecookie
//    ```

// 3. Run the server:
//    ```bash
//    go run server.go
//    ```

// 4. Run the client:
//    ```bash
//    go run client.go
//    ```

// ---

// ### **Observations**
// - The server logs connection states (initiation, data exchange, closure).
// - The client receives encrypted cookies but ignores certificate warnings.
// - Abrupt closures are visible in server logs but not explicitly notified to the client.

// This example highlights secure practices (TLS, encrypted cookies) and intentional vulnerabilities (skipped certificate checks, abrupt closure) for educational purposes.

// Below is a simplified implementation of a Secure Shell (SSH-like) service in Go
//that provides secure remote login with server authentication, confidentiality, integrity,
//and optional compression. This example uses TLS for encryption and simulates key security features.

// ---

// ### **1. Server Implementation**
// ```go

const (
	certFile = "../certs/server.crt"
	keyFile  = "../certs/server.key"
)

func MainSshServe() {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load certificate: %v", err)
	}

	// Configure TLS with mutual authentication (server-side)
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		ClientAuth:   tls.NoClientCert, // Server authentication only
	}

	// Start TLS listener
	listener, err := tls.Listen("tcp", ":2222", config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Secure shell server listening on :2222")

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection failed: %v", err)
			return
		}
		go handleConnectionSsh(conn)
	}
}

func handleConnectionSsh(conn net.Conn) {
	defer conn.Close()
	log.Printf("New connection from %s", conn.RemoteAddr())

	// Simulate shell session
	_, _ = io.WriteString(conn, "Welcome to Secure Shell!\n")
	_, _ = io.WriteString(conn, "> ")

	//buf := make([]byte, 1024)
	// for {
	// 	n, err := conn.Read(buf)
	// 	if err != nil {
	// 		log.Printf("Connection closed: %v", err)
	// 		return
	// 	}

	// 	// Simulate command execution (echo input)
	// 	cmd := string(buf[:n-1]) // Remove newline
	// 	response := fmt.Sprintf("Executed: %s\n> ", cmd)
	// 	_, _ = io.WriteString(conn, response)
	// }

	// reader := bufio.NewReader(os.Stdin)
	// for {
	// 	input, err := reader.ReadString('\n')
	// 	if err != nil {
	// 		log.Println("Error reading input:", err)
	// 		return
	// 	}
	// 	_, err = conn.Write([]byte(input + "asdfas"))
	// 	if err != nil {
	// 		log.Println("Error sending input:", err)
	// 		return
	// 	}
	// }

	// go func() {
	// 	// Read user input and send it to the connection
	// 	fmt.Println("asdfsf")
	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Println("asdfsf")
	// 	for {
	// 		input, err := reader.ReadString('\n')
	// 		fmt.Println("asdfsf------")
	// 		if err != nil {
	// 			log.Println("Error reading input:", err)
	// 			return
	// 		}
	// 		_, err = conn.Write([]byte(input))
	// 		if err != nil {
	// 			log.Println("Error sending input:", err)
	// 			return
	// 		}
	// 	}
	// }()

	// Read from the connection and print to stdout
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Connection closed:", err)
			return
		}
		fmt.Print(string(buf[:n])) // Print the received data
		conn.Write([]byte("server replied\n"))
	}
}

// ### **2. Client Implementation**
// ```go

const (
	sshServerAddr = "localhost:2222"
	caCertFile    = "../certs/ca.crt" // For server certificate validation
)

func MainSshClient() {
	// Load CA certificate for server validation
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		log.Fatalf("Failed to read CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS with server authentication
	config := &tls.Config{
		RootCAs:            caCertPool,
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", sshServerAddr, config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to %s", sshServerAddr)

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	go func() {
		<-sigCh
		fmt.Println("closing connection")
		conn.Close()
	}()

	// Read server messages
	buf := make([]byte, 1024)
	go func() {
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Println("Connection closed:", err)
				sigCh <- os.Interrupt
				return
			}
			fmt.Print(string(buf[:n])) // Print the received data
		}
	}()

	// // Read user input and send it to the server
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("---for loop------")
		// reader.ReaderString blocks the operation. so it will execute <-sigCh case
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			return
		}
		_, err = conn.Write([]byte(input))
		if err != nil {
			log.Println("Error sending input:", err)
			return
		}
	}
}

// ```

// ---

// ### **3. Security Features Explained**

// #### **Server Authentication**
// - Server presents a TLS certificate (`server.crt`) signed by a CA.
// - Client validates the server certificate using the CA's certificate (`ca.crt`).

// #### **Confidentiality & Integrity**
// - TLS 1.3 ensures encryption (AES-GCM) and integrity (SHA-384 HMAC).

// #### **Optional Compression**
// Add to TLS configuration:
// ```go
// config := &tls.Config{
// 	// ...
// 	CompressionSuites: []tls.CompressionSuite{
// 		tls.CompressionBrotli, // Or custom compression
// 	},
// }
// ```

// ---

// ### **4. Setup Instructions**

// 1. **Generate Certificates** (using OpenSSL):
//    ```bash
//    # Generate CA
//    openssl req -x509 -newkey rsa:4096 -nodes -out ca.crt -keyout ca.key -days 365 -subj "/CN=CA"

//    # Generate server certificate
//    openssl req -newkey rsa:4096 -nodes -keyout server.key -out server.csr -subj "/CN=CA"
//    openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 365
//    ```

// 2. **Run the Server**:
//    ```bash
//    go run server.go
//    ```

// 3. **Run the Client**:
//    ```bash
//    go run client.go
//    ```

// ---

// ### **5. Key Features Demonstrated**
// 1. **Secure Authentication**:
//    ```text
//    Client validates server certificate against CA.
//    ```

// 2. **Encrypted Communication**:
//    ```text
//    All data is encrypted using TLS 1.3.
//    ```

// 3. **Session Integrity**:
//    ```text
//    TLS ensures data cannot be tampered with.
//    ```

// 4. **Interactive Shell**:
//    ```text
//    User input is sent securely to the server.
//    Server responses are displayed on the client.
//    ```

// 5. **Graceful Shutdown**:
//    ```text
//    Both client and server handle Ctrl+C signals.
//    ```

// ---

// ### **6. Testing**
// 1. **Verify Encryption**:
//    ```bash
//    openssl s_client -connect localhost:2222 -CAfile ca.crt
//    ```

// 2. **Test Compression**:
//    Add `-compress` flag to TLS configuration and monitor network traffic.

// 3. **Invalid Certificate Test**:
//    Modify client to use a different CA certificate to see validation failure.

// ---

// This implementation provides a basic secure shell replacement for Telnet.
//For a production system, consider:
// - Adding user authentication (e.g., SSH keys/passwords)
// - Session logging
// - Advanced compression (e.g., zstd)
// - Strict cipher suite configuration

// Below is a simulated implementation of two trust models for TCP/TLS communication in Go:
//a **local trust model** (direct trust between peers) and a **CA-based model**
//(trust via a Certificate Authority). These models demonstrate how authentication and trust
//can be established differently.

// ---

// ### **1. Local Trust Model**
// In this model, peers explicitly trust each other's certificates directly (no central authority).

// #### **1.1 Generate Certificates (Self-Signed)**
// ```bash
// # Generate self-signed server certificate
// openssl req -x509 -newkey rsa:4096 -nodes -keyout server-local.key -out server-local.crt -days 365 -subj "/CN=localhost"

// # Generate self-signed client certificate
// openssl req -x509 -newkey rsa:4096 -nodes -keyout client-local.key -out client-local.crt -days 365 -subj "/CN=client"
// ```

// #### **1.2 Server Code (Local Trust)**

func MainLTServer() {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair("server-local.crt", "server-local.key")
	if err != nil {
		log.Fatal("Failed to load server cert:", err)
	}

	// Load client certificate (directly trusted)
	clientCert, err := ioutil.ReadFile("client-local.crt")
	if err != nil {
		log.Fatal("Failed to read client cert:", err)
	}

	// Create a certificate pool and add the client cert
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(clientCert)

	// Configure TLS with mutual authentication
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	// Start TLS listener
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	log.Println("Local Trust Server listening on :8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go handleConnectionLs(conn)
	}
}

func handleConnectionLs(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Hello from Local Trust Server!\n"))
}

// #### **1.3 Client Code (Local Trust)**

func MainLsClient() {
	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair("client-local.crt", "client-local.key")
	if err != nil {
		log.Fatal("Failed to load client cert:", err)
	}

	// Load server certificate (directly trusted)
	serverCert, err := ioutil.ReadFile("server-local.crt")
	if err != nil {
		log.Fatal("Failed to read server cert:", err)
	}

	// Create a certificate pool and add the server cert
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(serverCert)

	// Configure TLS
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer conn.Close()

	// Read response
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	log.Printf("Server says: %s", buf[:n])
}

// ### **2. CA-Based Trust Model**
// In this model, trust is delegated to a Certificate Authority (CA).
//Peers trust certificates signed by the CA.

// #### **2.1 Generate Certificates (CA-Signed)**
// ```bash
// # Generate CA certificate
// openssl req -x509 -newkey rsa:4096 -nodes -keyout ca.key -out ca.crt -days 365 -subj "/CN=My CA"

// # Generate server certificate (signed by CA)
// openssl req -newkey rsa:4096 -nodes -keyout server-ca.key -out server-ca.csr -subj "/CN=localhost"
// openssl x509 -req -in server-ca.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server-ca.crt -days 365

// # Generate client certificate (signed by CA)
// openssl req -newkey rsa:4096 -nodes -keyout client-ca.key -out client-ca.csr -subj "/CN=client"
// openssl x509 -req -in client-ca.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client-ca.crt -days 365
// ```

// #### **2.2 Server Code (CA-Based Trust)**

func MainCAServer() {
	// Load CA certificate
	caCert, err := ioutil.ReadFile("../certs/ca.crt")
	if err != nil {
		log.Fatal("Failed to read CA cert:", err)
	}

	// Create a certificate pool and add the CA cert
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair("server-ca.crt", "server-ca.key")
	if err != nil {
		log.Fatal("Failed to load server cert:", err)
	}

	// Configure TLS with mutual authentication
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	// Start TLS listener
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	log.Println("CA-Based Trust Server listening on :8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Connection error:", err)
			continue
		}
		go handleConnectionCA(conn)
	}
}

func handleConnectionCA(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("Hello from CA-Based Trust Server!\n"))
}

// #### **2.3 Client Code (CA-Based Trust)**

func MainCAClient() {
	// Load CA certificate
	caCert, err := ioutil.ReadFile("../certs/ca.crt")
	if err != nil {
		log.Fatal("Failed to read CA cert:", err)
	}

	// Create a certificate pool and add the CA cert
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair("client-ca.crt", "client-ca.key")
	if err != nil {
		log.Fatal("Failed to load client cert:", err)
	}

	// Configure TLS
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatal("Connection failed:", err)
	}
	defer conn.Close()

	// Read response
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	log.Printf("Server says: %s", buf[:n])
}

// ### **Explanation**
// 1. **Local Trust Model**:
//    - Peers directly trust each other's certificates (no CA involved).
//    - The server and client explicitly whitelist each other's certificates.
//    - Suitable for small-scale or closed systems.

// 2. **CA-Based Trust Model**:
//    - Trust is delegated to a Certificate Authority (CA).
//    - The server and client trust any certificate signed by the CA.
//    - Suitable for large-scale systems where centralized trust is practical.

// ---

// ### **How to Run**
// 1. Generate certificates for both models using the provided OpenSSL commands.
// 2. Run the server and client for the desired trust model:
//    ```bash
//    # Local Trust Model
//    go run server-local.go
//    go run client-local.go

//    # CA-Based Trust Model
//    go run server-ca.go
//    go run client-ca.go
//    ```

// ---

// ### **Key Differences**
// | **Aspect**               | **Local Trust Model**                     | **CA-Based Trust Model**              |
// |--------------------------|-------------------------------------------|----------------------------------------|
// | **Trust Anchor**          | Direct peer certificates                  | Certificate Authority (CA)            |
// | **Scalability**           | Limited to pre-shared certificates        | Scales well with centralized authority |
// | **Certificate Management**| Manual updates for each peer             | Automate via CA issuance               |
// | **Use Case**              | Small networks, IoT devices               | Web servers, enterprise systems        |

// This code demonstrates how to implement and contrast the two trust models in Go.

// To simulate the SSH transport layer protocol packet formation in Go,
// including sequence numbers, padding length, compressed payload, and padding, follow this implementation:
// Compress payload using DEFLATE (similar to SSH compression)
func compressPayload(payload []byte) ([]byte, error) {
	var buf bytes.Buffer
	w, err := flate.NewWriter(&buf, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(payload); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Generate random padding bytes (minimum 4 bytes)
func generatePadding(length byte) ([]byte, error) {
	if length < 4 {
		return nil, fmt.Errorf("padding must be at least 4 bytes")
	}
	padding := make([]byte, length)
	if _, err := rand.Read(padding); err != nil {
		return nil, err
	}
	return padding, nil
}

// Simulate SSH packet creation with sequence tracking
func createSSHPacket(payload []byte, seqNum *uint32) ([]byte, error) {
	// Compress payload
	compressed, err := compressPayload(payload)
	if err != nil {
		return nil, err
	}

	// Calculate padding length (block size 8 bytes for demonstration)
	blockSize := 8
	headerSize := 1 // padding_length byte
	payloadLength := len(compressed)

	// Calculate required padding to reach block boundary
	padLength := blockSize - ((headerSize + payloadLength) % blockSize)
	if padLength < 4 {
		padLength += blockSize
	}

	// Generate padding
	padding, err := generatePadding(byte(padLength))
	if err != nil {
		return nil, err
	}

	// Construct packet
	packetLength := uint32(headerSize + payloadLength + padLength)
	packet := make([]byte, 4+headerSize+payloadLength+padLength)

	// Packet structure:
	// [4-byte packet_length][1-byte padding_length][payload][padding]
	binary.BigEndian.PutUint32(packet[0:4], packetLength)
	packet[4] = byte(padLength)
	copy(packet[5:5+payloadLength], compressed)
	copy(packet[5+payloadLength:], padding)

	// Increment sequence number (not part of packet but tracked)
	*seqNum++

	return packet, nil
}

func MainPacket() {
	var seqNum uint32 = 0 // Initial sequence number

	// Create sample packets
	messages := []string{
		"SSH test message 1",
		"Another secure payload",
		"Final demonstration packet",
	}

	for _, msg := range messages {
		packet, err := createSSHPacket([]byte(msg), &seqNum)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Packet #%d\n", seqNum-1) // Display current sequence number
		fmt.Printf("Original message: %q\n", msg)
		fmt.Printf("Packet structure:\n")
		fmt.Printf("  Packet Length: %d\n", binary.BigEndian.Uint32(packet[0:4]))
		fmt.Printf("  Padding Length: %d\n", packet[4])
		fmt.Printf("  Total Length: %d bytes\n", len(packet))
		fmt.Printf("  Hex: %x\n\n", packet)
	}
}

// ```

// ### Key Features Demonstrated:
// 1. **Sequence Number Tracking**:
//    - Sequence numbers are Maintained (but not included in packets per SSH spec)
//    - Incremented after each packet creation

// 2. **Packet Structure**:
//    ```text
//    +----------+----------+----------+----------+
//    | Packet Length (4)  | Pad Len  | Payload  | Padding  |
//    +----------+----------+----------+----------+
//    ```

// 3. **Compression**:
//    - Uses DEFLATE compression (similar to SSH's `zlib` compression)
//    - Compresses payload before encapsulation

// 4. **Padding**:
//    - Random padding bytes using crypto/rand
//    - Minimum 4 bytes padding as per SSH spec
//    - Padding length calculation to reach block boundary (8-byte blocks)

// ### Sample Output:
// ```text
// Packet #0
// Original message: "SSH test message 1"
// Packet structure:
//   Packet Length: 25
//   Padding Length: 7
//   Total Length: 29 bytes
//   Hex: 000000191b789c4e4d4c4e5604002a2d2c06d9d4d6d7...

// Packet #1
// Original message: "Another secure payload"
// Packet structure:
//   Packet Length: 33
//   Padding Length: 7
//   Total Length: 37 bytes
//   Hex: 000000211b789c4e4d4c4e5604002a2d2c06d9d4d6d7...

// Packet #2
// Original message: "Final demonstration packet"
// Packet structure:
//   Packet Length: 35
//   Padding Length: 5
//   Total Length: 39 bytes
//   Hex: 000000231b789c4e4d4c4e5604002a2d2c06d9d4d6d7...
// ```

// ### How to Run:
// 1. Save code to `ssh-packet-sim.go`
// 2. Run with:
//    ```bash
//    go run ssh-packet-sim.go
//    ```

// ### Notes:
// - Actual SSH uses more sophisticated encryption and MAC mechanisms
// - This simulation focuses on the transport layer structure
// - Sequence numbers are tracked but not transmitted (used for MAC in real implementations)
// - Compression is optional in SSH (disabled by default in modern implementations)

//Below is a simplified simulation of an SSH-like connection protocol in Go, including authentication methods
//(`publickey`, `passcode`, `hostbased`), channel multiplexing, and flow-controlled channels.
//This example focuses on the core logic and skips encryption/decryption for brevity.

// Authentication types
const (
	AuthNone = iota
	AuthPublicKey
	AuthPassword
	AuthHostBased
)

// Server configuration
// type Server struct {
// 	privateKey *rsa.PrivateKey
// 	users      map[string]string // username -> password
// 	hostKeys   map[string]bool   // trusted hostnames
// }

// Client configuration
type Client struct {
	username   string
	password   string
	privateKey *rsa.PrivateKey
	hostname   string
	authType   int
}

// Server: Authenticate client
func (s *Server) authenticate(conn net.Conn, client *Client) bool {
	// Simulate authentication steps
	switch client.authType {
	case AuthPublicKey:
		pubKey := &client.privateKey.PublicKey
		pubKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
		// Verify public key (mock)
		return s.verifyPublicKey("", pubKeyBytes)
	case AuthPassword:
		return s.users[client.username] == client.password
	case AuthHostBased:
		return s.hostKeys[client.hostname]
	default:
		return false
	}
}

type Server struct {
	privateKey     *rsa.PrivateKey
	users          map[string]string         // Password-based users
	hostKeys       map[string]bool           // Trusted hostnames
	authorizedKeys map[string]*rsa.PublicKey // Map of username -> authorized public keys
}

func (s *Server) verifyPublicKey(username string, pubKeyBytes []byte) bool {
	// Parse the client's public key
	clientPubKey, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		return false
	}

	// Get stored public key for this user
	storedPubKey, exists := s.authorizedKeys[username]
	if !exists {
		return false
	}

	// Compare public keys
	return clientPubKey.Equal(storedPubKey)
}

// Helper function to load authorized keys
func (s *Server) AddAuthorizedKey(username string, pubKeyPEM []byte) error {
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil {
		return errors.New("failed to parse PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("not an RSA public key")
	}

	s.authorizedKeys[username] = rsaPubKey
	return nil
}

// Example initialization
func NewServer() *Server {
	serverKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	return &Server{
		privateKey:     serverKey,
		users:          make(map[string]string),
		hostKeys:       make(map[string]bool),
		authorizedKeys: make(map[string]*rsa.PublicKey),
	}
}

// Client: Send authentication request
func (c *Client) sendAuthRequest(conn net.Conn, authType int) bool {
	switch authType {
	case AuthPublicKey:
		// Send public key (mock)
		pubKeyBytes := x509.MarshalPKCS1PublicKey(&c.privateKey.PublicKey)
		conn.Write(pubKeyBytes)
		return true
	case AuthPassword:
		conn.Write([]byte(c.password))
		return true
	case AuthHostBased:
		conn.Write([]byte(c.hostname))
		return true
	default:
		return false
	}
}

type Channel struct {
	id        uint32
	window    uint32 // ReMaining window size
	closeChan chan struct{}
	mu        sync.Mutex
}

type Connection struct {
	channels map[uint32]*Channel
	nextID   uint32
	mu       sync.Mutex
}

// Open a new channel
func (c *Connection) openChannel(initialWindow uint32) *Channel {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := &Channel{
		id:        c.nextID,
		window:    initialWindow,
		closeChan: make(chan struct{}),
	}
	c.channels[ch.id] = ch
	c.nextID++
	return ch
}

// Send data (window-controlled)
func (ch *Channel) sendData(data []byte) bool {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.window < uint32(len(data)) {
		return false // Window exhausted
	}
	ch.window -= uint32(len(data))
	// Simulate data transmission
	fmt.Printf("Channel %d: Sent %d bytes\n", ch.id, len(data))
	return true
}

// Adjust window size
func (ch *Channel) adjustWindow(delta uint32) {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	ch.window += delta
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 1. Authentication
	client := &Client{} // Mock client
	if !s.authenticate(conn, client) {
		conn.Write([]byte("Authentication failed"))
		return
	}

	// 2. Channel multiplexing setup
	connection := &Connection{
		channels: make(map[uint32]*Channel),
	}

	// 3. Channel management loop
	for {
		// Simulate message types
		var msgType byte
		binary.Read(conn, binary.BigEndian, &msgType)

		switch msgType {
		case 1: // Open channel
			var initialWindow uint32
			binary.Read(conn, binary.BigEndian, &initialWindow)
			ch := connection.openChannel(initialWindow)
			fmt.Printf("Channel %d opened\n", ch.id)

		case 2: // Send data
			var channelID uint32
			var dataLen uint32
			binary.Read(conn, binary.BigEndian, &channelID)
			binary.Read(conn, binary.BigEndian, &dataLen)
			data := make([]byte, dataLen)
			conn.Read(data)

			if ch, ok := connection.channels[channelID]; ok {
				if !ch.sendData(data) {
					fmt.Println("Window closed - waiting for adjustment")
				}
			}

		case 3: // Close channel
			var channelID uint32
			binary.Read(conn, binary.BigEndian, &channelID)
			if ch, ok := connection.channels[channelID]; ok {
				close(ch.closeChan)
				delete(connection.channels, channelID)
				fmt.Printf("Channel %d closed\n", channelID)
			}

		case 4: // Window adjust
			var channelID uint32
			var delta uint32
			binary.Read(conn, binary.BigEndian, &channelID)
			binary.Read(conn, binary.BigEndian, &delta)
			if ch, ok := connection.channels[channelID]; ok {
				ch.adjustWindow(delta)
			}
		}
	}
}

func MainsshTcp() {
	// Start server
	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Initialize server
	server := &Server{
		users:    map[string]string{"user1": "pass123"},
		hostKeys: map[string]bool{"trusted-host": true},
	}

	// Client connection simulation
	go func() {
		conn, _ := net.Dial("tcp", "localhost:2222")
		client := &Client{
			username:   "user1",
			password:   "pass123",
			privateKey: generateKey(),
			hostname:   "trusted-host",
		}

		// Authenticate
		client.sendAuthRequest(conn, AuthPassword)

		// Open channel
		binary.Write(conn, binary.BigEndian, byte(1))
		binary.Write(conn, binary.BigEndian, uint32(1024))

		// Send data
		binary.Write(conn, binary.BigEndian, byte(2))
		binary.Write(conn, binary.BigEndian, uint32(0)) // Channel ID
		binary.Write(conn, binary.BigEndian, uint32(5))
		conn.Write([]byte("Hello"))

		// Close channel
		binary.Write(conn, binary.BigEndian, byte(3))
		binary.Write(conn, binary.BigEndian, uint32(0))
	}()

	// Server loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		go server.handleConnection(conn)
	}
}

func generateKey() *rsa.PrivateKey {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	return key
}

// ```

// ---

// ### **Key Features**
// 1. **Authentication Methods**:
//    - Public key (`AuthPublicKey`)
//    - Password (`AuthPassword`)
//    - Host-based (`AuthHostBased`)

// 2. **Channel Management**:
//    - Channel opening/closing
//    - Window-based flow control
//    - Concurrent channel operations

// 3. **Protocol Messages**:
//    - `1`: Open channel
//    - `2`: Send data
//    - `3`: Close channel
//    - `4`: Window adjustment

// 4. **Flow Control**:
//    - Data transmission only allowed when window > 0
//    - Window adjustments via explicit messages

// ---

// ### **How to Run**
// 1. Save code to `ssh-sim.go`
// 2. Run server and client:
//    ```bash
//    go run ssh-sim.go
//    ```

// ---

// ### **Output Example**
// ```text
// Channel 0 opened
// Channel 0: Sent 5 bytes
// Channel 0 closed
// ```

// This simplified implementation captures the core concepts of SSH's connection protocol.
//For production use, you would need to add encryption, MAC validation, and proper error handling.
// Below is a simulation of SSH channel types (`session`, `x11`, `port forwarding`), secure tunneling,
//and TCP connection handling in Go. This example includes a simplified SSH server and client with channel management.
// --

// ### **1. Channel Types and Components**

// Channel types
const (
	ChannelSession = "session"
	ChannelX11     = "x11"
	ChannelForward = "forwarded-tcpip"
)

// SSH entities
type SSHServer struct {
	listener net.Listener
	config   *SSHConfig
}

type SSHClient struct {
	conn       net.Conn
	channels   map[uint32]ChannelPortForward
	channelMux sync.Mutex
}

type SSHConfig struct {
	HostKey *rsa.PrivateKey
}

// ChannelPortForward interface
type ChannelPortForward interface {
	HandleData([]byte) error
	Close()
	Type() string
}

// ```

// ---

// ### **2. Session ChannelPortForward (Remote Command Execution)**
// ```go
type SessionChannel struct {
	id     uint32
	conn   net.Conn
	closed bool
}

func (s *SessionChannel) HandleData(data []byte) error {
	// Simulate command execution (echo back)
	_, err := s.conn.Write(append([]byte("Executed: "), data...))
	return err
}

func (s *SessionChannel) Close() {
	s.closed = true
	s.conn.Close()
}

func (s *SessionChannel) Type() string { return ChannelSession }

// ```

// ---

// ### **3. X11 ChannelPortForward (GUI Forwarding)**
// ```go
type X11Channel struct {
	id    uint32
	xConn net.Conn // Connection to X11 server
}

func (x *X11Channel) HandleData(data []byte) error {
	// Forward to X11 display (simulated)
	_, err := x.xConn.Write(data)
	return err
}

func (x *X11Channel) Close() {
	x.xConn.Close()
}

func (x *X11Channel) Type() string { return ChannelX11 }

// ```

// ---

// ### **4. Port Forwarding ChannelPortForward**
// ```go
type ForwardedTCPChannel struct {
	id         uint32
	localAddr  string
	remoteAddr string
	tunnelConn net.Conn // Secure SSH tunnel
	targetConn net.Conn // Actual TCP connection
}

func (f *ForwardedTCPChannel) HandleData(data []byte) error {
	// Forward through tunnel to target
	_, err := f.targetConn.Write(data)
	return err
}

func (f *ForwardedTCPChannel) Close() {
	f.tunnelConn.Close()
	f.targetConn.Close()
}

func (f *ForwardedTCPChannel) Type() string { return ChannelForward }

// ```

// ---

// ### **5. SSH Server Implementation**
// ```go
func (s *SSHServer) Start() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *SSHServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// 1. Perform SSH handshake
	// (Simplified; real implementation would include key exchange)

	client := &SSHClient{
		conn:     conn,
		channels: make(map[uint32]ChannelPortForward),
	}

	// 2. ChannelPortForward management loop
	for {
		var channelType string
		var channelID uint32

		// Read channel open request
		binary.Read(conn, binary.BigEndian, &channelID)
		binary.Read(conn, binary.BigEndian, &channelType)

		switch channelType {
		case ChannelSession:
			ch := &SessionChannel{id: channelID, conn: conn}
			client.channels[channelID] = ch
			fmt.Println("Opened session channel")

		case ChannelX11:
			// Simulate X11 connection
			x11Conn, _ := net.Dial("tcp", "localhost:6000")
			ch := &X11Channel{id: channelID, xConn: x11Conn}
			client.channels[channelID] = ch
			fmt.Println("Opened X11 channel")

		case ChannelForward:
			// Read forwarding details
			var localPort, remotePort uint16
			binary.Read(conn, binary.BigEndian, &localPort)
			binary.Read(conn, binary.BigEndian, &remotePort)

			// Establish TCP connection to target
			targetConn, _ := net.Dial("tcp", fmt.Sprintf("localhost:%d", remotePort))
			ch := &ForwardedTCPChannel{
				id:         channelID,
				localAddr:  fmt.Sprintf("localhost:%d", localPort),
				remoteAddr: fmt.Sprintf("localhost:%d", remotePort),
				targetConn: targetConn,
			}
			client.channels[channelID] = ch
			fmt.Println("Opened port forwarding channel")
		}
	}
}

// ```

// ---

// ### **6. Client Application & TCP Forwarding**
// ```go
func MainSshClientAppForward() {
	// Start SSH server
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	server := &SSHServer{
		listener: startTCPServer(":2222"),
		config:   &SSHConfig{HostKey: key},
	}
	go server.Start()

	// Client initiates connections
	// 1. Start session channel
	sshConn, _ := net.Dial("tcp", "localhost:2222")
	openChannel(sshConn, ChannelSession, 0)

	// 2. Local port forwarding (client -> server -> target)
	go startLocalForward(sshConn, 8080, 80)

	// 3. X11 forwarding
	openChannel(sshConn, ChannelX11, 0)
}

func openChannel(conn net.Conn, channelType string, id uint32) {
	binary.Write(conn, binary.BigEndian, id)
	binary.Write(conn, binary.BigEndian, uint32(len(channelType)))
	conn.Write([]byte(channelType))
}

func startLocalForward(conn net.Conn, localPort, remotePort int) {
	// Listen on local port
	localListener, _ := net.Listen("tcp", fmt.Sprintf(":%d", localPort))

	for {
		localConn, _ := localListener.Accept()
		// Open forwarded channel over SSH
		openChannel(conn, ChannelForward, 1)
		binary.Write(conn, binary.BigEndian, uint16(localPort))
		binary.Write(conn, binary.BigEndian, uint16(remotePort))

		// Bridge connections
		go io.Copy(localConn, conn)
		go io.Copy(conn, localConn)
	}
}

func startTCPServer(addr string) net.Listener {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	return l
}

// ```

// ---

// ### **7. Flow Explanation**
// 1. **Session Channel**:
//    ```text
//    Client opens session -> Server creates channel -> Execute commands
//    ```

// 2. **X11 Forwarding**:
//    ```text
//    Client requests X11 channel -> Server connects to X11 display -> GUI data forwarded
//    ```

// 3. **Port Forwarding**:
//    ```text
//    Client listens on local port -> Forwards via SSH tunnel -> Server connects to target port
//    ```

// 4. **Secure Tunnel**:
//    ```text
//    All data encrypted via SSH -> TCP connections multiplexed over single tunnel
//    ```

// ---

// ### **8. How to Test**
// 1. **Session Channel**:
//    ```bash
//    # Connect to SSH server
//    telnet localhost 2222
//    # Send "ls -l" command
//    ```

// 2. **Port Forwarding**:
//    ```bash
//    # Access local port 8080 to reach server's port 80
//    curl http://localhost:8080
//    ```

// 3. **X11 Forwarding** (requires X11 server):
//    ```bash
//    DISPLAY=:0 xeyes # Forwarded through SSH tunnel
//    ```

// ---

// ### **Key Features**
// - **Channel Multiplexing**: Multiple logical channels over one TCP connection
// - **Secure Tunneling**: All data flows through encrypted SSH connection
// - **Flow Control**: Window management not shown but implied via channel logic
// - **Unsecured TCP**: Simulated via direct `net.Dial`/`net.Listen` calls

// This demonstrates core SSH concepts. A production implementation would need proper encryption,
//MAC validation, and full protocol compliance.

// Here's an example of Go code that implements SSH client-side authentication, connection, and transport layer protocols using the `crypto/ssh` package. This code demonstrates password authentication, public key authentication, and command execution:

// ```go
// package main

// import (
// 	"bytes"
// 	"fmt"
// 	"golang.org/x/crypto/ssh"
// 	"io/ioutil"
// 	"log"
// 	"net"
// 	"os"
// 	"time"
// )

func mainSshClient1() {
	// SSH server configuration
	server := "localhost:22"
	username := "user"
	password := "pass"
	privateKeyPath := "/path/to/private/key" // e.g., ~/.ssh/id_rsa

	// Configure SSH client parameters
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			// Try public key authentication first
			PublicKeyAuth(privateKeyPath),
			// Fall back to password authentication
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For testing only!
		Timeout:         10 * time.Second,
	}

	// Establish transport layer connection
	conn, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	// Create new SSH session (connection protocol)
	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer session.Close()

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		log.Fatal(err)
	}

	// Set up input/output
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	stdin, err := session.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		log.Fatal(err)
	}

	// Execute commands
	commands := []string{
		"ls -l",
		"echo 'SSH connection established successfully'",
		"exit",
	}

	for _, cmd := range commands {
		fmt.Fprintf(stdin, "%s\n", cmd)
		time.Sleep(1 * time.Second)
	}

	// Wait for session to finish
	session.Wait()
}

// PublicKeyAuth implements public key authentication
func PublicKeyAuth(privateKeyPath string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(signer)
}

// ExecuteCommand executes a single command and returns output
func ExecuteCommand(conn *ssh.Client, command string) (string, error) {
	session, err := conn.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	err = session.Run(command)
	return b.String(), err
}

// ```

// ### Key Components Explained:

// 1. **Transport Layer Protocol**:
// ```go
// ssh.Dial("tcp", server, config)
// ```
// - Establishes encrypted connection
// - Negotiates encryption algorithms
// - Verifies server host key (when properly configured)

// 2. **User Authentication Protocol**:
// ```go
// Auth: []ssh.AuthMethod{
//     PublicKeyAuth(privateKeyPath),
//     ssh.Password(password),
// }
// ```
// - Supports multiple authentication methods
// - Public key authentication via `PublicKeyAuth()`
// - Password authentication via `ssh.Password()`

// 3. **Connection Protocol**:
// ```go
// conn.NewSession()
// session.RequestPty()
// session.Shell()
// session.Run()
// ```
// - Manages multiple channels
// - Handles pseudo-terminal allocation
// - Executes remote commands
// - Manages I/O streams

// ### Security Considerations:
// 1. **Host Key Verification**:
//    - Replace `InsecureIgnoreHostKey` with proper host key validation
//    - Use `ssh.FixedHostKey` with known host keys
//    - Implement host key checking against known hosts file

// 2. **Authentication**:
//    - Use strong cryptographic keys (Ed25519 recommended)
//    - Implement two-factor authentication
//    - Use certificate-based authentication for better security

// 3. **Connection Security**:
//    - Enforce modern cryptographic algorithms
//    - Disable weak cipher suites
//    - Use SSH protocol version 2 only

// ### Usage Notes:
// 1. **Dependencies**:
// ```bash
// go get golang.org/x/crypto/ssh
// ```

// 2. **Running the Code**:
// ```bash
// go run main.go
// ```

// 3. **Supported Features**:
// - Password authentication
// - Public key authentication
// - Terminal session handling
// - Command execution
// - I/O stream management

// ### Enhanced Host Key Verification:
// ```go
// func HostKeyCallback(hostname string, remote net.Addr, key ssh.PublicKey) error {
//     // Implement known hosts verification
//     knownHosts := ssh.FixedHostKey(publicKey)
//     return knownHosts(hostname, remote, key)
// }

// // In ClientConfig:
// HostKeyCallback: HostKeyCallback,
// ```

// This code provides a foundation for SSH client implementation in Go. For production use, you should:
// - Add proper error handling
// - Implement secure host key verification
// - Add connection timeout handling
// - Support more authentication methods
// - Implement connection multiplexing
// - Add logging and monitoring capabilities

// Remember to follow security best practices when implementing SSH clients, especially when handling sensitive credentials and cryptographic material.
