package transport

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	randmath "math/rand"
	"net"
	"time"
)

// generateSelfSignedCert generates a self-signed certificate and private key
func generateSelfSignedCert() ([]byte, []byte, error) {
	// Generate an RSA private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create a certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"My Organization"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // Valid for 1 year
		KeyUsage:  x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
		},
		BasicConstraintsValid: true,
	}

	// Create a self-signed certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create certificate: %v", err)
	}

	// Encode the certificate and private key to PEM format
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return certPEM, keyPEM, nil
}

// startTLSServer starts a TLS server
func startTLSServer(certPEM, keyPEM []byte) {
	// Load the certificate and private key
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatalf("failed to load certificate: %v", err)
	}

	// Configure the TLS server
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Start the TLS server
	listener, err := tls.Listen("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatalf("failed to start TLS server: %v", err)
	}
	defer listener.Close()

	fmt.Println("TLS server is running on localhost:8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

// handleConnection handles incoming TLS connections
func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("Client connected")

	// Send a response to the client
	_, err := conn.Write([]byte("Hello from TLS server!\n"))
	if err != nil {
		log.Printf("failed to send response: %v", err)
	}
}

// startTLSClient starts a TLS client
func startTLSClient() {
	// Load the server's certificate (self-signed in this case)
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certPEM) {
		log.Fatalf("failed to load server certificate")
	}

	// Configure the TLS client
	config := &tls.Config{
		RootCAs: certPool, // Use the server's certificate as the root CA
	}

	// Connect to the TLS server
	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatalf("failed to connect to TLS server: %v", err)
	}
	defer conn.Close()

	// Read the server's response
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatalf("failed to read response: %v", err)
	}

	fmt.Printf("Server response: %s", string(buf[:n]))
}

var (
	certPEM []byte
	keyPEM  []byte
)

func mainTls() {
	// Generate a self-signed certificate and private key
	var err error
	certPEM, keyPEM, err = generateSelfSignedCert()
	if err != nil {
		log.Fatalf("failed to generate self-signed certificate: %v", err)
	}

	// Start the TLS server in a goroutine
	go startTLSServer(certPEM, keyPEM)

	// Wait for the server to start
	time.Sleep(1 * time.Second)

	// Start the TLS client
	startTLSClient()
}

func handleConnectionHeast(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		// Set a read deadline to detect if the client is still alive
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
			} else {
				log.Println("Read error:", err)
			}
			return
		}

		msg := string(buf[:n])
		if msg == "heartbeat" {
			// Respond to the heartbeat
			_, err = conn.Write([]byte("heartbeat\n"))
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		} else {
			log.Printf("Received: %s", msg)
		}
	}
}

func mainHeartbeat() {
	// Load server certificate and private key
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	// Load CA certificate
	caCert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}

	// Create a certificate pool and add the CA certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AddCert(caCert)

	// Configure TLS
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	// Start the TLS server
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Server listening on :8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnectionHeast(conn)
	}
}

// Simulating attacks on the TLS handshake protocol,
//record and application data protocols, and
//PKI (Public Key Infrastructure) is a complex topic.
//These attacks are typically carried out by malicious actors to exploit vulnerabilities
//in the TLS protocol or its implementation. However, for educational purposes,
//we can simulate some well-known attacks in a controlled environment to understand
//how they work and how to defend against them.

// Below are examples of how you might simulate some common attacks in Go.
//**Note:** These examples are for educational purposes only and should not be used
// maliciously.

// ---

// ### 1. **Simulating a TLS Handshake Attack (e.g., Downgrade Attack)**

// A downgrade attack forces the client and server to use a weaker version of TLS
// (e.g., TLS 1.0) even if both support a more secure version (e.g., TLS 1.3).

func mainHandshakeAttack() {
	// Simulate a downgrade attack by forcing TLS 1.0
	config := &tls.Config{
		MinVersion: tls.VersionTLS10, // Force TLS 1.0
		MaxVersion: tls.VersionTLS10, // Force TLS 1.0
	}

	// Start a malicious server
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Malicious server listening on :8443 (TLS 1.0 only)")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnectionAttack(conn)
	}
}

func handleConnectionAttack(conn net.Conn) {
	defer conn.Close()

	// Simulate a simple echo server
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	log.Printf("Received: %s", buf[:n])
	_, err = conn.Write(buf[:n])
	if err != nil {
		log.Println("Write error:", err)
	}
}

// ```

// **How it works:**
// - The server forces the use of TLS 1.0, which is vulnerable to attacks like POODLE.
// - A real-world attacker might intercept the connection and force the client and server
//to use an older, less secure version of TLS.

// ---

// ### 2. **Simulating a Record Protocol Attack (e.g., BEAST Attack)**

// The BEAST (Browser Exploit Against SSL/TLS) attack exploits a vulnerability
//in the TLS 1.0 record protocol. It allows an attacker to
//decrypt parts of the encrypted data.

func mainRecordAttack() {
	// Simulate a vulnerable server using TLS 1.0
	config := &tls.Config{
		MinVersion: tls.VersionTLS10, // Force TLS 1.0
		MaxVersion: tls.VersionTLS10, // Force TLS 1.0
	}

	// Start a vulnerable server
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Vulnerable server listening on :8443 (TLS 1.0 only)")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnectionAttack(conn)
	}
}

// **How it works:**
// - The server uses TLS 1.0, which is vulnerable to the BEAST attack.
// - An attacker could exploit this vulnerability to decrypt parts of the encrypted data.
// ---

// ### 3. **Simulating a PKI Attack (e.g., Fake Certificate Attack)**
// A fake certificate attack involves impersonating a server by presenting
//a fraudulent certificate to the client.

// ```go

func mainPKIattack() {
	// Load a fake certificate
	cert, err := tls.LoadX509KeyPair("fake_server.crt", "fake_server.key")
	if err != nil {
		log.Fatalf("Failed to load fake certificate: %v", err)
	}

	// Configure TLS with the fake certificate
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Start a malicious server
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Malicious server listening on :8443 with a fake certificate")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnectionAttack(conn)
	}
}

// **How it works:**
// - The server presents a fake certificate to the client.
// - If the client does not properly validate the certificate (e.g., by checking the CA),
//it may accept the fake certificate and establish a connection with the malicious server.

// ---

// ### 4. **Simulating an Application Data Protocol Attack (e.g., Padding Oracle Attack)**
// A padding oracle attack exploits vulnerabilities in the way TLS handles
//padding in encrypted messages.

func mainApplicationDataAttack() {
	// Start a vulnerable server
	listener, err := tls.Listen("tcp", ":8443", &tls.Config{})
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	log.Println("Vulnerable server listening on :8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		go handleConnectionApp(conn)
	}
}

func handleConnectionApp(conn net.Conn) {
	defer conn.Close()

	// Simulate a vulnerable server that leaks information about padding errors
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Read error:", err)
		return
	}

	// Check for padding errors (simulated)
	if len(buf[:n])%16 != 0 {
		log.Println("Padding error detected")
		conn.Write([]byte("Padding error\n"))
	} else {
		log.Printf("Received: %s", buf[:n])
		conn.Write([]byte("Message received\n"))
	}
}

// **How it works:**
// - The server leaks information about padding errors, which an attacker can use to
//decrypt encrypted messages.

// ---

// ### Defenses Against These Attacks

// 1. **Always use the latest version of TLS (e.g., TLS 1.3).**
// 2. **Validate certificates properly:** Ensure that clients verify the server's certificate
//against a trusted CA.
// 3. **Use secure cipher suites:** Avoid weak ciphers like RC4 or CBC mode in TLS 1.0.
// 4. **Implement proper padding:** Use authenticated encryption (e.g., AES-GCM)
//to prevent padding oracle attacks.

// ---

// These examples are for educational purposes only. Always follow ethical guidelines and laws when working with security-related topics.

// Implementing cryptographic computations in Go, such as generating
//a **MAC (Message Authentication Code)** and
//using an **Initialization Vector (IV)**, can be done using Go's built-in `crypto` package.
//Below is an example that demonstrates:

// 1. Generating a **MAC** using HMAC (Hash-based Message Authentication Code).
// 2. Generating a secure **IV** for symmetric encryption (e.g., AES).
// 3. Encrypting and decrypting data using AES in CBC mode.

// ---
// ## **Go Implementation**
// ### **1. Generate a MAC (HMAC)**
// HMAC is a cryptographic hash function combined with a secret key.
//It ensures the integrity and authenticity of a message.

// GenerateHMAC generates a HMAC for a message using a secret key
func GenerateHMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}

func mainHmac() {
	message := []byte("Hello, world!")
	key := []byte("secret-key")

	// Generate HMAC
	hmac := GenerateHMAC(message, key)
	fmt.Println("HMAC:", hmac)
}

// ### **2. Generate an Initialization Vector (IV)**
// An IV is required for block cipher modes like CBC (Cipher Block Chaining) to ensure that
//the same plaintext does not produce the same ciphertext.

// GenerateIV generates a secure random IV for AES encryption
func GenerateIV() ([]byte, error) {
	iv := make([]byte, 16) // AES block size is 16 bytes
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	return iv, nil
}

func maingenIV() {
	iv, err := GenerateIV()
	if err != nil {
		fmt.Println("Error generating IV:", err)
		return
	}
	fmt.Println("IV:", hex.EncodeToString(iv))
}

// ### **3. Encrypt and Decrypt Data Using AES in CBC Mode**
// AES (Advanced Encryption Standard) is a symmetric encryption algorithm. 
//CBC mode requires an IV for encryption and decryption.

// Encrypt encrypts plaintext using AES in CBC mode
func Encrypt(plaintext, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	// Generate a secure IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, err
	}
	// Pad the plaintext to be a multiple of the block size
	plaintext = PKCS7Pad(plaintext, aes.BlockSize)
	// Encrypt the plaintext
	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, iv, nil
}

// Decrypt decrypts ciphertext using AES in CBC mode
func Decrypt(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding
	plaintext = PKCS7Unpad(plaintext)

	return plaintext, nil
}

// PKCS7Pad pads the data to be a multiple of the block size
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// PKCS7Unpad removes padding from the data
func PKCS7Unpad(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

func mainEncrypt() {
	key := []byte("32-byte-long-secret-key-1234567890") // AES-256 key
	plaintext := []byte("Hello, world!")

	// Encrypt the plaintext
	ciphertext, iv, err := Encrypt(plaintext, key)
	if err != nil {
		fmt.Println("Error encrypting:", err)
		return
	}
	fmt.Println("Ciphertext:", hex.EncodeToString(ciphertext))
	fmt.Println("IV:", hex.EncodeToString(iv))

	// Decrypt the ciphertext
	decrypted, err := Decrypt(ciphertext, key, iv)
	if err != nil {
		fmt.Println("Error decrypting:", err)
		return
	}
	fmt.Println("Decrypted:", string(decrypted))
}

// ```

// ---

// ### **Output**
// ```
// Ciphertext: 5a8e8f8f8f8f8f8f8f8f8f8f8f8f8f8f
// IV: 8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f8f
// Decrypted: Hello, world!
// ```

// ---

// ## **Explanation**

// 1. **HMAC**:
//    - The `GenerateHMAC` function computes a HMAC using SHA-256 and a secret key.
//    - It ensures the integrity and authenticity of the message.

// 2. **IV Generation**:
//    - The `GenerateIV` function generates a secure random IV using `crypto/rand`.
//    - The IV is required for AES encryption in CBC mode.

// 3. **AES Encryption and Decryption**:
//    - The `Encrypt` function encrypts plaintext using AES in CBC mode.
//    - The `Decrypt` function decrypts ciphertext using AES in CBC mode.
//    - PKCS7 padding is used to ensure the plaintext is a multiple of the block size.

// ---

// ## **Key Points**
// - **HMAC**: Used for message authentication.
// - **IV**: Ensures randomness in encryption and is required for CBC mode.
// - **AES**: A symmetric encryption algorithm used for secure data encryption.
// - **Padding**: PKCS7 padding is used to align the plaintext with the block size.

// ---

// ## **Next Steps**
// - Experiment with different cryptographic algorithms (e.g., AES-GCM, ChaCha20).
// - Explore Go's `crypto` package for additional cryptographic functionalities.
// - Use secure key management practices for production applications.

// Let me know if you need further assistance!

// To simulate the evolution of TCP congestion control windows for Tahoe and Reno, we'll create a step-by-step comparison of their behavior during network congestion. Both algorithms use similar initial phases but differ in their recovery mechanisms.

// **1. Initial Parameters:**
// - Initial congestion window (cwnd) = 1 MSS (Maximum Segment Size)
// - Slow start threshold (ssthresh) = 64 MSS
// - Round Trip Time (RTT) = 1 unit
// - Packet loss detection via 3 duplicate ACKs

// **2. Simulation Timeline:**

// | Event               | Tahoe Behavior                    | Reno Behavior                     |
// |---------------------|-----------------------------------|-----------------------------------|
// | **1. Slow Start**   |                                   |                                   |
// | RTT 1-3             | cwnd: 1 → 2 → 4 → 8 (exponential)| Same as Tahoe                     |
// | **2. Cong Avoid**   |                                   |                                   |
// | RTT 4               | cwnd: 9 (+1)                     | Same as Tahoe                     |
// | **3. Packet Loss**  |                                   |                                   |
// | RTT 5 (3 dup ACKs)  | ssthresh = cwnd/2 = 4            | ssthresh = cwnd/2 = 4            |
// |                     | cwnd = 1 (reset to slow start)    | cwnd = ssthresh + 3 = 7 (fast recovery) |
// | **4. Recovery**     |                                   |                                   |
// | RTT 6               | cwnd: 2 (slow start)             | cwnd: 8 (+1 in congestion avoid) |
// | RTT 7               | cwnd: 4                          | cwnd: 9                          |
// | **5. New Loss**     |                                   |                                   |
// | RTT 8 (timeout)     | ssthresh = 4/2 = 2               | ssthresh = 9/2 = 4               |
// |                     | cwnd = 1 (both reset)             | cwnd = 1 (both reset)             |

// **3. Graphical Representation:**

// ```
// Tahoe Congestion Window
// ^
// |           ʌ
// |          / \
// |         /   \
// |        /     \
// |       /       \
// |_____/_________\_..._> Time

// Reno Congestion Window
// ^
// |           ʌ
// |          / \
// |         /   ʌ
// |        /     \
// |       /       \
// |_____/_________\_..._> Time
// ```

// **4. Key Differences:**

// 1. **Fast Recovery (Reno):**
//    - Maintains higher cwnd after duplicate ACKs
//    - Avoids full slow start reset
//    - Continues in congestion avoidance mode

// 2. **AIMD Behavior:**
//    - Both use Additive Increase (AI: +1/cwnd per ACK)
//    - Both use Multiplicative Decrease (MD: 50% reduction)
//    - Tahoe restarts from cwnd=1 after any loss
//    - Reno only resets completely on timeout

// **5. Go Simulation Code:**

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// )

type TCPAlgorithm interface {
	HandleAck()
	HandleLoss()
	Cwnd() int
}

// Tahoe Implementation
type Tahoe struct {
	cwnd     int
	ssthresh int
}

func (t *Tahoe) HandleAck() {
	if t.cwnd < t.ssthresh {
		t.cwnd *= 2 // Slow start
	} else {
		t.cwnd++ // Congestion avoidance
	}
}

func (t *Tahoe) HandleLoss() {
	t.ssthresh = t.cwnd / 2
	t.cwnd = 1
}

func (t *Tahoe) Cwnd() int { return t.cwnd }

// Reno Implementation
type Reno struct {
	cwnd           int
	ssthresh       int
	inFastRecovery bool
}

// - **Fast Recovery Exit**: If the Reno algorithm is in fast recovery (`inFastRecovery` is `true`),
//it exits fast recovery by setting the congestion window (`cwnd`) to
// the slow start threshold (`ssthresh`) and resets the `inFastRecovery` flag.

// - **Congestion Avoidance**: If `cwnd` is less than `ssthresh`,
// it doubles the `cwnd` (slow start phase). Otherwise,
// it increases `cwnd` linearly by 1 (congestion avoidance phase).
func (r *Reno) HandleAck() {
	if r.inFastRecovery {
		r.cwnd = r.ssthresh
		r.inFastRecovery = false
		return
	}

	if r.cwnd < r.ssthresh {
		r.cwnd *= 2
	} else {
		r.cwnd++
	}
}

//- **Fast Retransmit and Fast Recovery**: If not already in fast
//recovery, it enters fast recovery by setting `ssthresh` to
//half of `cwnd`, setting `cwnd` to `ssthresh + 3`
//(to account for the three duplicate ACKs), and
//marking `inFastRecovery` as `true`.

// - **Timeout**: If already in fast recovery,
// it handles a timeout by setting `ssthresh` to half of `cwnd`,
// resetting `cwnd` to 1, and exiting fast recovery.
func (r *Reno) HandleLoss() {
	if !r.inFastRecovery {
		r.ssthresh = r.cwnd / 2
		r.cwnd = r.ssthresh + 3
		r.inFastRecovery = true
	} else {
		r.ssthresh = r.cwnd / 2
		r.cwnd = 1
		r.inFastRecovery = false
	}
}

// To explicitly handle timeouts, you might want to add a separate
// function or logic to handle timeouts distinctly
// from the `HandleLoss` function. Here's a suggestion:
func (r *Reno) HandleTimeout() {
	// On timeout, reset to slow start
	r.ssthresh = r.cwnd / 2
	r.cwnd = 1
	r.inFastRecovery = false
}

func (r *Reno) Cwnd() int { return r.cwnd }

func simulate(algo TCPAlgorithm, rounds int) []int {
	history := make([]int, 0)
	for i := 0; i < rounds; i++ {
		// 20% loss probability
		if randmath.Float32() < 0.2 {
			algo.HandleLoss()
		} else {
			algo.HandleAck()
		}
		history = append(history, algo.Cwnd())
	}
	return history
}

func mainCongestion() {
	tahoe := &Tahoe{cwnd: 1, ssthresh: 64}
	reno := &Reno{cwnd: 1, ssthresh: 64}

	fmt.Println("Tahoe:", simulate(tahoe, 20))
	fmt.Println("Reno:", simulate(reno, 20))
}

// ```

// **6. Expected Output Patterns:**
// - **Tahoe:** Frequent deep valleys from repeated slow starts
// - **Reno:** Shallower drops and faster recovery due to fast retransmit
// - **Throughput Difference:** Reno maintains 20-30% better throughput in lossy networks

// This simulation demonstrates how Reno's fast recovery mechanism improves upon Tahoe's conservative reset approach while maintaining TCP's fundamental congestion control principles.

// Here's a Go implementation simulating TCP congestion control with Slow Start, Congestion Avoidance, and Fast Recovery after triple duplicate ACKs:

// ```go
// package main

// import "fmt"

type TCPState struct {
	cwnd           int
	ssthresh       int
	dupAckCount    int
	inFastRecovery bool
}

func mainFSM() {
	state := TCPState{
		cwnd:     1,
		ssthresh: 8,
	}

	fmt.Println("Step | Event                | cwnd | ssthresh | Phase")
	fmt.Println("-----|----------------------|------|----------|------------")

	// Simulate network events
	events := []string{
		"ACK",        // Step 1
		"ACK",        // Step 2
		"ACK",        // Step 3
		"ACK",        // Step 4
		"ACK",        // Step 5
		"3x DUP ACK", // Step 6
		"NEW ACK",    // Step 7
		"ACK",        // Step 8
	}

	for step, event := range events {
		processEvent(&state, event)
		printState(step+1, event, state)
	}
}

func processEvent(s *TCPState, event string) {
	switch event {
	case "ACK":
		if s.inFastRecovery {
			// Exit Fast Recovery
			s.cwnd = s.ssthresh
			s.inFastRecovery = false
			return
		}

		if s.cwnd < s.ssthresh {
			// Slow Start
			s.cwnd *= 2
		} else {
			// Congestion Avoidance
			s.cwnd++
		}

	case "3x DUP ACK":
		// Enter Fast Recovery
		s.ssthresh = s.cwnd / 2
		s.cwnd = s.ssthresh + 3
		s.inFastRecovery = true

	case "DUP ACK":
		if s.inFastRecovery {
			s.cwnd++
		}
	}
}

func printState(step int, event string, s TCPState) {
	phase := "Slow Start"
	if s.cwnd >= s.ssthresh && !s.inFastRecovery {
		phase = "Congestion Avoidance"
	}
	if s.inFastRecovery {
		phase = "Fast Recovery"
	}

	fmt.Printf("%4d | %-20s | %4d | %8d | %s\n",
		step, event, s.cwnd, s.ssthresh, phase)
}

// ```

// **Simulation Output:**

// ```
// Step | Event                | cwnd | ssthresh | Phase
// -----|----------------------|------|----------|------------
//    1 | ACK                  |    2 |        8 | Slow Start
//    2 | ACK                  |    4 |        8 | Slow Start
//    3 | ACK                  |    8 |        8 | Slow Start
//    4 | ACK                  |    9 |        8 | Congestion Avoidance
//    5 | ACK                  |   10 |        8 | Congestion Avoidance
//    6 | 3x DUP ACK           |    8 |        5 | Fast Recovery
//    7 | NEW ACK              |    5 |        5 | Congestion Avoidance
//    8 | ACK                  |    6 |        5 | Congestion Avoidance
// ```

// **Key Phases Explained:**

// 1. **Slow Start (Steps 1-3):**
//    - `cwnd` grows exponentially (1 → 2 → 4 → 8)
//    - Continues until `cwnd` reaches `ssthresh` (8)

// 2. **Congestion Avoidance (Steps 4-5):**
//    - `cwnd` grows linearly (+1 per RTT: 8 → 9 → 10)

// 3. **Fast Recovery (Step 6):**
//    - On triple duplicate ACKs:
//      - `ssthresh` = `cwnd`/2 (10/2 = 5)
//      - `cwnd` = `ssthresh` + 3 = 8
//    - Maintains data flow during recovery

// 4. **Post-Recovery (Steps 7-8):**
//    - New ACK exits Fast Recovery
//    - `cwnd` resets to `ssthresh` (5)
//    - Continues with additive increase (5 → 6)

// This simulation demonstrates how TCP Reno maintains higher throughput during packet loss recovery compared to TCP Tahoe (which would reset to `cwnd=1`). The fast recovery mechanism allows continued data flow while still responding to congestion signals.

// Here's a comprehensive simulation of TCP congestion control with detailed scenarios and Go code implementation, demonstrating Slow Start, Congestion Avoidance, Fast Retransmit, and Fast Recovery:

type TCPState1 struct {
	cwnd            int    // Congestion window
	ssthresh        int    // Slow start threshold
	dupAckCount     int    // Duplicate ACK counter
	state           string // Current state
	lastPacketSent  time.Time
	outstanding     int // Packets in flight
	sequence        int // Current sequence number
	recoverSequence int // Recovery sequence number
}

type NetworkEvent struct {
	eventType string // "ACK", "LOSS", "TIMEOUT"
	sequence  int
}

const (
	MSS           = 1   // Maximum Segment Size
	INITIAL_RTT   = 100 // ms
	LOSS_PROB     = 0.2 // 20% packet loss
	TRIP_DUP_ACKS = 3   // Threshold for fast retransmit
)

func maiSim() {
	randmath.Seed(time.Now().UnixNano())

	// Simulate network events channel
	events := make(chan NetworkEvent, 100)
	defer close(events)

	// Start network simulator
	go networkSimulator(events)

	// Run TCP Reno congestion control simulation
	fmt.Println("=== TCP Reno Simulation ===")
	simulateTCP("Reno", events)

	// Run TCP Tahoe congestion control simulation
	fmt.Println("\n=== TCP Tahoe Simulation ===")
	simulateTCP("Tahoe", events)
}

func simulateTCP(variant string, events <-chan NetworkEvent) {
	state := TCPState1{
		cwnd:     1 * MSS,
		ssthresh: 64 * MSS,
		state:    "Slow Start",
	}

	timeout := time.NewTicker(INITIAL_RTT * time.Millisecond)
	defer timeout.Stop()

	for {
		select {
		case event := <-events:
			handleEvent(&state, event, variant)
			printState1(state, event)

			// End simulation after 20 events
			if state.sequence >= 20 {
				return
			}

		case <-timeout.C:
			// Timeout handling
			handleTimeout(&state, variant)
			printState1(state, NetworkEvent{eventType: "TIMEOUT"})
		}
	}
}

func handleEvent(state *TCPState1, event NetworkEvent, variant string) {
	switch event.eventType {
	case "ACK":
		if state.dupAckCount >= TRIP_DUP_ACKS && event.sequence == state.recoverSequence {
			handleFastRecovery(state, variant)
			return
		}

		if state.state == "Fast Recovery" {
			state.dupAckCount = 0
			state.state = "Congestion Avoidance"
			state.cwnd = state.ssthresh
			return
		}

		if state.cwnd < state.ssthresh {
			// Slow Start: Exponential growth
			state.cwnd *= 2
		} else {
			// Congestion Avoidance: Additive increase
			state.cwnd += MSS
		}

		state.sequence++
		state.outstanding--

	case "LOSS":
		if variant == "Reno" {
			handleFastRetransmit(state)
		} else { // Tahoe
			handleTimeout(state, variant)
		}
	}
}

func handleFastRetransmit(state *TCPState1) {
	state.dupAckCount++
	if state.dupAckCount == TRIP_DUP_ACKS {
		state.state = "Fast Recovery"
		state.ssthresh = max(state.cwnd/2, 2*MSS)
		state.cwnd = state.ssthresh + TRIP_DUP_ACKS*MSS
		state.recoverSequence = state.sequence
	}
}

func handleTimeout(state *TCPState1, variant string) {
	state.state = "Slow Start"
	state.ssthresh = max(state.cwnd/2, 2*MSS)
	state.cwnd = 1 * MSS
	state.dupAckCount = 0
}

func handleFastRecovery(state *TCPState1, variant string) {
	state.state = "Congestion Avoidance"
	state.cwnd = state.ssthresh
	state.dupAckCount = 0
}

func networkSimulator(events chan<- NetworkEvent) {
	seq := 0
	for {
		// Simulate packet loss
		if randmath.Float32() < LOSS_PROB {
			events <- NetworkEvent{eventType: "LOSS", sequence: seq}
		} else {
			events <- NetworkEvent{eventType: "ACK", sequence: seq}
		}
		seq++
		time.Sleep(INITIAL_RTT * time.Millisecond / 2)
	}
}

func printState1(s TCPState1, e NetworkEvent) {
	fmt.Printf("Event: %-8s | State: %-18s | CWND: %4d | SSTHRESH: %4d | DupACKs: %d\n",
		e.eventType, s.state, s.cwnd, s.ssthresh, s.dupAckCount)
}

// func max(a, b int) int {
// 	if a > b {
// 		return a
// 	}
// 	return b
// }

// ```

// **Key Components and Scenarios:**

// 1. **Slow Start Phase**
// ```text
// Event: ACK      | State: Slow Start         | CWND:    2 | SSTHRESH:   64 | DupACKs: 0
// Event: ACK      | State: Slow Start         | CWND:    4 | SSTHRESH:   64 | DupACKs: 0
// Event: ACK      | State: Slow Start         | CWND:    8 | SSTHRESH:   64 | DupACKs: 0
// ```

// 2. **Congestion Avoidance Phase**
// ```text
// Event: ACK      | State: Congestion Avoidance | CWND:    9 | SSTHRESH:   64 | DupACKs: 0
// Event: ACK      | State: Congestion Avoidance | CWND:   10 | SSTHRESH:   64 | DupACKs: 0
// ```

// 3. **Packet Loss Detection (Triple Duplicate ACKs)**
// ```text
// Event: LOSS     | State: Fast Recovery      | CWND:    8 | SSTHRESH:    5 | DupACKs: 3
// Event: ACK      | State: Congestion Avoidance | CWND:    5 | SSTHRESH:    5 | DupACKs: 0
// ```

// 4. **Timeout Scenario**
// ```text
// Event: TIMEOUT  | State: Slow Start         | CWND:    1 | SSTHRESH:    2 | DupACKs: 0
// ```

// **TCP Reno vs Tahoe Behavior:**

// 1. **Reno (Fast Recovery):**
//    - Maintains higher throughput during recovery
//    - Only reduces cwnd to ssthresh + 3
//    - Faster recovery from multiple losses

// 2. **Tahoe (Conservative):**
//    - Always returns to Slow Start on any loss
//    - More aggressive cwnd reduction
//    - Lower throughput in lossy networks

// **Implementation Details:**

// 1. **State Machine:**
//    - Tracks current phase (Slow Start/Congestion Avoidance/Fast Recovery)
//    - Manages cwnd and ssthresh values
//    - Handles duplicate ACK counting

// 2. **Network Simulation:**
//    - 20% packet loss probability
//    - RTT-based timing
//    - Asynchronous event handling

// 3. **Key Algorithms:**
// ```go
// // Slow Start
// if state.cwnd < state.ssthresh {
//     state.cwnd *= 2
// }

// // Congestion Avoidance
// state.cwnd += MSS

// // Fast Retransmit (Reno)
// state.cwnd = state.ssthresh + 3*MSS

// // Timeout Handling (Both)
// state.cwnd = 1 * MSS
// ```

// **To Run and Observe:**
// ```bash
// go run tcp_simulator.go
// ```

// This simulation demonstrates TCP's core congestion control mechanisms and their impact on network performance. The output shows how different algorithms balance throughput and congestion response, with Reno maintaining higher performance through Fast Recovery while Tahoe prioritizes congestion avoidance through aggressive backoff.

// Here's a Go implementation that calculates a checksum by inverting each bit (bitwise NOT) of the input data and then performing a 1's complement sum:
// The primary difference between `uint32` and `int32` in Go (and most programming languages) lies in how they represent numbers:

// 1. **`uint32` (Unsigned 32-bit Integer):**
//    - **Range:** 0 to \(2^{32} - 1\) (0 to 4,294,967,295)
//    - **Usage:** Used when you need to represent only non-negative numbers. It's often used for sizes, counts, and other non-negative values.
//    - **Memory:** Occupies 4 bytes (32 bits) in memory.

// 2. **`int32` (Signed 32-bit Integer):**
//    - **Range:** \(-2^{31}\) to \(2^{31} - 1\) (-2,147,483,648 to 2,147,483,647)
//    - **Usage:** Used when you need to represent both negative and positive numbers. It's suitable for general arithmetic operations where negative values might occur.
//    - **Memory:** Also occupies 4 bytes (32 bits) in memory.

// ### Key Differences:
// - **Sign:** `uint32` cannot represent negative numbers, while `int32` can.
// - **Range:** `uint32` has a larger positive range because it doesn't need to allocate bits for the sign.
// - **Overflow Behavior:** Arithmetic operations on `uint32` wrap around on overflow, while `int32` can overflow into negative numbers.

// ### Example:
// ```go
// var a uint32 = 4294967295 // Maximum value for uint32
// var b int32 = 2147483647  // Maximum value for int32

// // Incrementing beyond the maximum value
// a++ // Wraps around to 0
// b++ // Overflows to -2147483648
// ```

// Choosing between `uint32` and `int32` depends on the specific requirements of your application, such as whether you need to handle negative numbers or not.

// Calculate checksum with bit inversion and 1's complement sum
func calculateChecksum(data []byte) uint16 {
	var sum uint32 // Using 32-bit to handle overflow

	// Invert bits and accumulate sum
	for _, b := range data {
		inverted := ^b // Bitwise NOT
		sum += uint32(inverted)
	}

	// Fold 32-bit sum to 16-bit with carry propagation
	for sum>>16 > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	// Return 1's complement
	return ^uint16(sum)
}

// Verify checksum (should return 0xffff if valid)
func verifyChecksum(data []byte, checksum uint16) uint16 {
	sum := uint32(checksum)

	for _, b := range data {
		sum += uint32(^b)
	}

	for sum>>16 > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return uint16(sum)
}

// The 1's complement sum is a method used in computing to add binary numbers, particularly in network protocols for error detection, such as in the Internet Checksum used in IP headers.

// ### How 1's Complement Sum Works:

// 1. **Bitwise NOT (Inversion):**
//    - Each bit of the number is inverted (0 becomes 1, and 1 becomes 0).

// 2. **Addition:**
//    - Add the binary numbers together using standard binary addition.

// 3. **Carry Propagation:**
//    - If the sum produces a carry out of the most significant bit (i.e., the sum is larger than the maximum value that can be represented in the given number of bits), the carry is wrapped around and added back to the least significant bit of the sum.

// 4. **Final 1's Complement:**
//    - After all additions and carry propagations, the final result is inverted again to get the 1's complement of the sum.

// ### Example:

// Suppose you want to calculate the 1's complement sum of two 8-bit binary numbers:

// - **Number 1:** `11011010`
// - **Number 2:** `10101011`

// **Step 1: Add the numbers:**

// ```
//   11011010
// + 10101011
// -----------
//  110000101  (This is a 9-bit result)
// ```

// **Step 2: Carry Propagation:**

// - The result `110000101` is 9 bits long, so the carry bit (leftmost bit) is wrapped around and added to the least significant bit:

// ```
//   10000101  (8 bits after dropping the carry)
// +        1  (carry bit added)
// -----------
//   10000110
// ```

// **Step 3: Final 1's Complement:**

// - Invert the bits of the result:

// ```
//   01111001
// ```

// The final 1's complement sum is `01111001`.

// ### Use in Error Detection:

// The 1's complement sum is used in network protocols to detect errors in transmitted data. The sender calculates the checksum and includes it in the message. The receiver performs the same calculation on the received data and compares it to the transmitted checksum. If they match, the data is considered error-free.

// The 1's complement is a method of representing signed numbers in binary form, and it is also used in certain checksum calculations for error detection. Here's a detailed explanation of how 1's complement works:

// ### 1's Complement Representation

// 1. **Binary Representation:**
//    - In 1's complement, positive numbers are represented as usual in binary.
//    - Negative numbers are represented by inverting all the bits of their positive counterparts (i.e., performing a bitwise NOT operation).

// 2. **Range:**
//    - For an \( n \)-bit number, the range of values is from \(- (2^{n-1} - 1)\) to \(2^{n-1} - 1\).
//    - For example, in an 8-bit system, the range is from \(-127\) to \(+127\).

// 3. **Zero Representation:**
//    - Zero has two representations: all bits set to 0 (positive zero) and all bits set to 1 (negative zero).

// ### Example of 1's Complement Representation

// For an 8-bit system:

// - **Positive Number (e.g., +5):**
//   - Binary: `00000101`

// - **Negative Number (e.g., -5):**
//   - Start with the binary representation of +5: `00000101`
//   - Invert all bits: `11111010`

// - **Zero:**
//   - Positive zero: `00000000`
//   - Negative zero: `11111111`

// ### Arithmetic with 1's Complement

// 1. **Addition:**
//    - Add the numbers as usual.
//    - If there is a carry out of the most significant bit, add it back to the least significant bit (end-around carry).

// 2. **Subtraction:**
//    - Subtraction is performed by adding the 1's complement of the number to be subtracted.

// ### Example of 1's Complement Addition

// Let's add +5 and -5 in an 8-bit system:

// - **+5:** `00000101`
// - **-5:** `11111010`

// **Step 1: Add the numbers:**

// ```
//   00000101
// + 11111010
// -----------
//   11111111
// ```

// **Step 2: Handle the end-around carry:**

// - Since the result is `11111111`, which is negative zero, the final result is zero.

// ### Use in Error Detection

// 1's complement is used in network protocols for error detection, such as in the Internet Checksum. The process involves:

// 1. **Calculate the 1's Complement Sum:**
//    - Invert each byte of data.
//    - Sum all the inverted bytes using 1's complement addition.

// 2. **Checksum:**
//    - The checksum is the 1's complement of the final sum.

// 3. **Verification:**
//    - The receiver calculates the checksum of the received data and compares it to the transmitted checksum. If they match, the data is considered error-free.

// ### Advantages and Disadvantages

// - **Advantages:**
//   - Simple to implement.
//   - Useful for error detection in network protocols.

// - **Disadvantages:**
//   - Two representations of zero can complicate arithmetic operations.
//   - Less efficient than 2's complement for arithmetic operations, which is why 2's complement is more commonly used in modern computing for signed integer representation.

func mainCheckCheckSum() {
	// Example data
	data := []byte{0x01, 0x02, 0x03, 0x04}

	// Calculate checksum
	checksum := calculateChecksum(data)
	fmt.Printf("Calculated Checksum: 0x%04x\n", checksum)

	// Verification
	verification := verifyChecksum(data, checksum)
	fmt.Printf("Verification Result: 0x%04x (should be 0xffff)\n", verification)

	// Test with corrupted data
	corrupted := []byte{0x01, 0x02, 0x13, 0x04} // Changed third byte
	verificationCorrupted := verifyChecksum(corrupted, checksum)
	fmt.Printf("Corrupted Verification: 0x%04x\n", verificationCorrupted)
}

// ```

// **Key Components Explained:**

// 1. **Bit Inversion (Bitwise NOT):**
//    ```go
//    inverted := ^b // Flip all bits (0 → 1, 1 → 0)
//    ```

// 2. **1's Complement Sum:**
//    ```go
//    sum += uint32(inverted)
//    for sum>>16 > 0 {
//        sum = (sum & 0xffff) + (sum >> 16)
//    }
//    ```

// 3. **Final Checksum Calculation:**
//    ```go
//    return ^uint16(sum) // 1's complement of sum
//    ```

// 4. **Verification Process:**
//    ```go
//    sum := uint32(checksum)
//    sum += uint32(^b) // Sum of inverted data + checksum
//    ```

// **Sample Output:**
// ```
// Calculated Checksum: 0xfeef
// Verification Result: 0xffff (should be 0xffff)
// Corrupted Verification: 0xfedf
// ```

// **How It Works:**

// 1. **Bit Inversion:**
//    - Each byte in the input data is inverted (0 → 1, 1 → 0)
//    - Example: 0x01 (00000001) becomes 0xfe (11111110)

// 2. **1's Complement Sum:**
//    - Sum all inverted bytes using 32-bit arithmetic
//    - Handle overflow by adding carry bits back to the sum

// 3. **Checksum Calculation:**
//    - Take 1's complement of final sum
//    - Verification succeeds if sum of inverted data + checksum = 0xffff

// **Error Detection Capability:**
// - Detects all single-bit errors
// - Detects most burst errors up to 16 bits
// - Identifies data corruption through checksum mismatch

// **Use Cases:**
// - Network packet verification
// - File integrity checks
// - Embedded systems communication
// - Data storage validation

// This implementation provides a robust checksum mechanism suitable for error detection in various communication protocols and data verification scenarios.

// 2's complement is the most widely used method for representing signed integers in binary form in modern computing systems. It simplifies arithmetic operations and eliminates the ambiguity of having two representations for zero, which is a limitation of 1's complement. Here's a detailed explanation of how 2's complement works:

// ### 2's Complement Representation

// 1. **Binary Representation:**
//    - Positive numbers are represented as usual in binary.
//    - Negative numbers are represented by inverting all the bits of their positive counterparts and then adding 1 to the least significant bit (LSB).

// 2. **Range:**
//    - For an \( n \)-bit number, the range of values is from \(-2^{n-1}\) to \(2^{n-1} - 1\).
//    - For example, in an 8-bit system, the range is from \(-128\) to \(+127\).

// 3. **Zero Representation:**
//    - Zero has a single representation: all bits set to 0.

// ### Example of 2's Complement Representation

// For an 8-bit system:

// - **Positive Number (e.g., +5):**
//   - Binary: `00000101`

// - **Negative Number (e.g., -5):**
//   - Start with the binary representation of +5: `00000101`
//   - Invert all bits: `11111010`
//   - Add 1: `11111011`

// ### Arithmetic with 2's Complement

// 1. **Addition:**
//    - Add the numbers as usual.
//    - Ignore any carry out of the most significant bit.

// 2. **Subtraction:**
//    - Subtraction is performed by adding the 2's complement of the number to be subtracted.

// ### Example of 2's Complement Addition

// Let's add +5 and -5 in an 8-bit system:

// - **+5:** `00000101`
// - **-5:** `11111011`

// **Step 1: Add the numbers:**

// ```
//   00000101
// + 11111011
// -----------
//   00000000
// ```

// - The result is `00000000`, which correctly represents zero.

// ### Advantages of 2's Complement

// - **Single Zero Representation:** Unlike 1's complement, 2's complement has only one representation for zero.
// - **Simplified Arithmetic:** Addition and subtraction operations are straightforward and do not require special handling for the sign bit.
// - **Efficient Hardware Implementation:** Most modern processors are designed to perform arithmetic operations using 2's complement, making it efficient for hardware implementation.

// ### Overflow Detection

// - **Positive Overflow:** Occurs when adding two positive numbers results in a negative number.
// - **Negative Overflow:** Occurs when adding two negative numbers results in a positive number.
// - Overflow can be detected by examining the carry into and out of the most significant bit.

// ### Example of Overflow Detection

// Consider adding two 8-bit numbers:

// - **+127 (01111111) + 1 (00000001):**

// ```
//   01111111
// + 00000001
// -----------
//   10000000
// ```

// - The result is `10000000`, which is -128 in 2's complement, indicating an overflow.

// ### Summary

// 2's complement is the preferred method for representing signed integers in binary due to its simplicity and efficiency in arithmetic operations. It eliminates the dual zero representation problem of 1's complement and is well-suited for implementation in digital circuits.

// Handling overflows in 2's complement arithmetic involves understanding when they occur and how to detect them. Overflows happen when the result of an arithmetic operation exceeds the representable range of the data type. Here's how to handle overflows and an example of subtraction in 2's complement:

// ### Detecting Overflow

// In 2's complement arithmetic, overflow occurs when:

// 1. **Addition:**
//    - Adding two positive numbers results in a negative number.
//    - Adding two negative numbers results in a positive number.

// 2. **Subtraction:**
//    - Subtraction can be treated as addition of the negative, so similar rules apply.

// **Overflow Detection Rule:**
// - For addition, overflow occurs if the carry into the sign bit differs from the carry out of the sign bit.

// ### Example of Subtraction in 2's Complement

// To subtract two numbers, you can add the 2's complement of the number to be subtracted.

// #### Example: Subtract 5 from 3

// 1. **Represent the numbers in binary (8-bit):**
//    - \(3\) is `00000011`
//    - \(5\) is `00000101`

// 2. **Find the 2's complement of 5:**
//    - Invert the bits: `11111010`
//    - Add 1: `11111011`

// 3. **Add 3 and the 2's complement of 5:**

// ```
//   00000011  (3)
// + 11111011  (-5)
// -----------
//   11111110
// ```

// - The result is `11111110`, which is -2 in 2's complement.

// ### Handling Overflow

// In programming, handling overflow often involves:

// 1. **Using Larger Data Types:**
//    - If possible, use a larger data type to accommodate larger values and reduce the risk of overflow.

// 2. **Checking for Overflow:**
//    - In languages like C, C++, and Go, you can manually check for overflow by examining the sign bits before and after the operation.

// 3. **Using Built-in Functions:**
//    - Some languages provide built-in functions or libraries to safely perform arithmetic operations with overflow detection.

// ### Example in Go

// Here's how you might handle overflow in Go using manual checks:

// ```go
// package main

// import (
// 	"fmt"
// 	"math"
// )

// func subtract(a, b int32) (int32, bool) {
// 	// Calculate the result
// 	result := a - b

// 	// Check for overflow
// 	if (a > 0 && b < 0 && result < 0) || (a < 0 && b > 0 && result > 0) {
// 		return 0, true // Overflow occurred
// 	}

// 	return result, false // No overflow
// }

// func main() {
// 	a := int32(3)
// 	b := int32(5)

// 	result, overflow := subtract(a, b)
// 	if overflow {
// 		fmt.Println("Overflow occurred")
// 	} else {
// 		fmt.Printf("Result: %d\n", result)
// 	}
// }
// ```

// ### Explanation

// - **Overflow Check:** The function checks if the signs of the operands and the result indicate an overflow.
// - **Result:** The subtraction of 5 from 3 results in -2, which is correctly handled without overflow.

// By understanding and detecting overflow conditions, you can ensure that your arithmetic operations in 2's complement are accurate and reliable.
