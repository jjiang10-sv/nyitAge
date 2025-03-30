// Here's a Go implementation simulating wireless network elements with decreasing signal strength over time. This example models access points, clients, and signal degradation mechanics:

// ```go
package transport

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type AccessPoint struct {
	SSID      string
	MAC       string
	Frequency float64 // GHz
	Channel   int
	Signal    float64 // dBm (starts strong, decreases with obstacles/time)
	clients   map[string]*WirelessClient
	mu        sync.Mutex
}

type WirelessClient struct {
	MAC         string
	IP          string
	ConnectedAP *AccessPoint
	Signal      float64 // Current signal strength
}

func mainWireless() {
	rand.Seed(time.Now().UnixNano())

	// Create access points
	ap1 := &AccessPoint{
		SSID:      "Office_Network_5G",
		MAC:       "00:1A:2B:3C:4D:5E",
		Frequency: 5.0,
		Channel:   36,
		Signal:    -30.0, // Strong initial signal
		clients:   make(map[string]*WirelessClient),
	}

	ap2 := &AccessPoint{
		SSID:      "Conference_Room_5G",
		MAC:       "00:1A:2B:3C:4D:5F",
		Frequency: 5.0,
		Channel:   44,
		Signal:    -45.0,
		clients:   make(map[string]*WirelessClient),
	}

	// Create wireless clients
	clients := []*WirelessClient{
		{MAC: "A4:B1:C2:D3:E4:F5", IP: "192.168.1.10"},
		{MAC: "58:EF:68:90:12:34", IP: "192.168.1.11"},
	}

	// Start signal degradation simulation
	go simulateSignalDecay(ap1)
	go simulateSignalDecay(ap2)

	// Simulate client connections
	var wg sync.WaitGroup
	for _, client := range clients {
		wg.Add(1)
		go func(c *WirelessClient) {
			defer wg.Done()
			simulateClient(c, []*AccessPoint{ap1, ap2})
		}(client)
	}

	wg.Wait()
}

func simulateSignalDecay(ap *AccessPoint) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ap.mu.Lock()
		// Simulate environmental factors reducing signal strength
		ap.Signal += rand.Float64()*2 - 4 // Random decay between -2 to -4 dBm
		fmt.Printf("[%s] Signal strength: %.1f dBm\n", ap.SSID, ap.Signal)
		ap.mu.Unlock()
	}
}

func simulateClient(c *WirelessClient, aps []*AccessPoint) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		bestAP := findBestAP(aps)
		if bestAP == nil {
			fmt.Printf("[%s] No AP with sufficient signal\n", c.MAC)
			continue
		}

		if c.ConnectedAP != bestAP {
			if c.ConnectedAP != nil {
				c.ConnectedAP.disconnectClient(c)
			}
			bestAP.connectClient(c)
			fmt.Printf("[%s] Switched to AP: %s (%.1f dBm)\n",
				c.MAC, bestAP.SSID, bestAP.Signal)
		}

		// Simulate client-side signal degradation
		c.Signal += rand.Float64()*1 - 3
		fmt.Printf("[%s] WirelessClient signal: %.1f dBm\n", c.MAC, c.Signal)
	}
}

func findBestAP(aps []*AccessPoint) *AccessPoint {
	var bestAP *AccessPoint
	bestSignal := math.Inf(-1)

	for _, ap := range aps {
		ap.mu.Lock()
		currentSignal := ap.Signal
		ap.mu.Unlock()

		if currentSignal > bestSignal && currentSignal > -70.0 { // Minimum viable signal
			bestSignal = currentSignal
			bestAP = ap
		}
	}
	return bestAP
}

func (ap *AccessPoint) connectClient(c *WirelessClient) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.clients[c.MAC] = c
	c.ConnectedAP = ap
	c.Signal = ap.Signal // Initial connection signal
}

func (ap *AccessPoint) disconnectClient(c *WirelessClient) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	delete(ap.clients, c.MAC)
	c.ConnectedAP = nil
}

// ```

// **Key Simulation Features:**

// 1. **Dynamic Signal Decay:**
//    - APs lose signal strength over time (-2 to -4 dBm every 10s)
//    - Clients experience independent signal degradation (-1 to -3 dBm every 5s)

// 2. **Intelligent AP Selection:**
//    - Clients automatically switch to best available AP
//    - Minimum viable signal threshold (-70 dBm)

// 3. **Concurrency Control:**
//    - Mutexes protect shared AP state
//    - Goroutines for parallel simulation of APs and clients

// 4. **Realistic Behaviors:**
//    - Signal fluctuation modeling
//    - Graceful AP handoff
//    - Connection stability checks

// **Sample Output:**
// ```
// [Office_Network_5G] Signal strength: -32.5 dBm
// [Conference_Room_5G] Signal strength: -47.8 dBm
// [A4:B1:C2:D3:E4:F5] Switched to AP: Office_Network_5G (-32.5 dBm)
// [A4:B1:C2:D3:E4:F5] WirelessClient signal: -34.2 dBm
// [58:EF:68:90:12:34] Switched to AP: Office_Network_5G (-32.5 dBm)
// [58:EF:68:90:12:34] WirelessClient signal: -35.1 dBm
// ...
// [Office_Network_5G] Signal strength: -68.3 dBm
// [Conference_Room_5G] Signal strength: -52.4 dBm
// [A4:B1:C2:D3:E4:F5] Switched to AP: Conference_Room_5G (-52.4 dBm)
// ```

// **Extensions Ideas:**
// 1. Add signal-to-noise ratio (SNR) calculations
// 2. Implement different propagation models (Free Space, Log-Distance)
// 3. Add interference sources (microwave, Bluetooth devices)
// 4. Introduce QoS prioritization
// 5. Add packet loss simulation based on signal strength

// This simulation provides a foundation for modeling wireless network behavior with signal degradation mechanics while demonstrating Go's concurrency features for realistic network simulations.

// Here's a Go implementation simulating both **passive** and **active scanning** in wireless networks, modeling access points (APs), clients, and frame exchanges:

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

// Frame types
const (
	FrameTypeBeacon    = "BEACON"
	FrameTypeProbeReq  = "PROBE_REQ"
	FrameTypeProbeResp = "PROBE_RESP"
)

// WirelessFrame represents a network frame
type WirelessFrame struct {
	SourceMAC string
	DestMAC   string // "FF:FF:FF:FF:FF:FF" for broadcast
	Channel   int
	FrameType string
	Payload   string
}

// AccessPoint1 represents a wireless AP
type AccessPoint1 struct {
	SSID      string
	MAC       string
	Channel   int
	BeaconInt time.Duration // Beacon interval
	AirChan   chan<- WirelessFrame
	shutdown  chan struct{}
}

// ClientDevice represents a wireless client
type ClientDevice struct {
	MAC           string
	ScanMode      string // "passive" or "active"
	CurrentChan   int
	DiscoveredAPs map[string]string // MAC -> SSID
	AirChan       chan WirelessFrame
}

func mainws() {
	rand.Seed(time.Now().UnixNano())
	air := make(chan WirelessFrame, 100) // Shared wireless medium

	// Create APs
	aps := []*AccessPoint1{
		createAP("Office_Net", "00:1A:2B:3C:4D:5E", 6, 100*time.Millisecond, air),
		createAP("Guest_Net", "00:1A:2B:3C:4D:5F", 11, 150*time.Millisecond, air),
	}

	// Create Clients
	clients := []*ClientDevice{
		createClient("A4:B1:C2:D3:E4:F5", "passive", 6, air),
		createClient("58:EF:68:90:12:34", "active", 11, air),
	}

	// Run simulation for 2 seconds
	time.Sleep(2 * time.Second)

	// Stop APs
	for _, ap := range aps {
		close(ap.shutdown)
	}

	// Print results
	fmt.Println("\n=== Discovery Results ===")
	for _, client := range clients {
		fmt.Printf("Client %s (%s scan) found:\n", client.MAC, client.ScanMode)
		for mac, ssid := range client.DiscoveredAPs {
			fmt.Printf(" - %s (%s)\n", ssid, mac)
		}
	}
}

func createAP(ssid, mac string, channel int, interval time.Duration, air chan<- WirelessFrame) *AccessPoint1 {
	ap := &AccessPoint1{
		SSID:      ssid,
		MAC:       mac,
		Channel:   channel,
		BeaconInt: interval,
		AirChan:   air,
		shutdown:  make(chan struct{}),
	}
	go ap.beaconLoop()
	return ap
}

func (ap *AccessPoint1) beaconLoop() {
	ticker := time.NewTicker(ap.BeaconInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ap.AirChan <- WirelessFrame{
				SourceMAC: ap.MAC,
				DestMAC:   "FF:FF:FF:FF:FF:FF",
				Channel:   ap.Channel,
				FrameType: FrameTypeBeacon,
				Payload:   ap.SSID,
			}
		case <-ap.shutdown:
			return
		}
	}
}

func createClient(mac, mode string, channel int, air chan WirelessFrame) *ClientDevice {
	client := &ClientDevice{
		MAC:           mac,
		ScanMode:      mode,
		CurrentChan:   channel,
		DiscoveredAPs: make(map[string]string),
		AirChan:       air,
	}
	go client.scanLoop()
	return client
}

func (c *ClientDevice) scanLoop() {
	// Active scanners send probe requests
	if c.ScanMode == "active" {
		go func() {
			for {
				c.AirChan <- WirelessFrame{
					SourceMAC: c.MAC,
					DestMAC:   "FF:FF:FF:FF:FF:FF",
					Channel:   c.CurrentChan,
					FrameType: FrameTypeProbeReq,
					Payload:   "",
				}
				time.Sleep(300 * time.Millisecond)
			}
		}()
	}

	// Process incoming frames
	for frame := range c.AirChan {
		if frame.Channel != c.CurrentChan {
			continue
		}

		switch frame.FrameType {
		case FrameTypeBeacon:
			if c.ScanMode == "passive" {
				c.DiscoveredAPs[frame.SourceMAC] = frame.Payload
			}

		case FrameTypeProbeResp:
			if c.ScanMode == "active" && frame.DestMAC == c.MAC {
				c.DiscoveredAPs[frame.SourceMAC] = frame.Payload
			}
		}
	}
}

// APs handle probe requests
func (ap *AccessPoint1) handleProbeReq(frame WirelessFrame) {
	if frame.Channel == ap.Channel {
		ap.AirChan <- WirelessFrame{
			SourceMAC: ap.MAC,
			DestMAC:   frame.SourceMAC,
			Channel:   ap.Channel,
			FrameType: FrameTypeProbeResp,
			Payload:   ap.SSID,
		}
	}
}

// ```

// **Key Components:**

// 1. **Wireless Medium Simulation:**
//    - Shared `air` channel acts as broadcast medium
//    - All devices read from/write to this channel

// 2. **Passive Scanning:**
//    - APs broadcast beacons periodically
//    - Passive clients listen silently and record beacon frames
//    ```go
//    case FrameTypeBeacon:
//        if c.ScanMode == "passive" {
//            c.DiscoveredAPs[frame.SourceMAC] = frame.Payload
//        }
//    ```

// 3. **Active Scanning:**
//    - Clients send probe requests
//    - APs respond with probe responses
//    ```go
//    case FrameTypeProbeResp:
//        if c.ScanMode == "active" && frame.DestMAC == c.MAC {
//            c.DiscoveredAPs[frame.SourceMAC] = frame.Payload
//        }
//    ```

// 4. **Channel Management:**
//    - Devices only process frames on their current channel
//    - APs respond to probe requests on matching channels

// **Sample Output:**
// ```
// === Discovery Results ===
// Client A4:B1:C2:D3:E4:F5 (passive scan) found:
//  - Office_Net (00:1A:2B:3C:4D:5E)

// Client 58:EF:68:90:12:34 (active scan) found:
//  - Guest_Net (00:1A:2B:3C:4D:5F)
// ```

// **Key Differences Modeled:**

// | **Aspect**         | **Passive Scanning**                | **Active Scanning**               |
// |---------------------|-------------------------------------|------------------------------------|
// | Client Transmission | No frames sent                      | Sends probe requests              |
// | Discovery Speed     | Depends on beacon interval (slower) | Immediate response (faster)       |
// | Network Traffic     | Low (only beacons)                  | Higher (probes + responses)       |
// | Energy Consumption  | Lower (only listening)              | Higher (transmitting + listening) |

// **To Enhance:**
// 1. Add channel switching logic for clients
// 2. Implement frame collision detection
// 3. Add signal strength thresholds
// 4. Support multiple channels simultaneously
// 5. Add authentication/association phases

// This simulation demonstrates the fundamental behaviors of wireless scanning mechanisms while leveraging Go's concurrency features for realistic network interactions.

// Here's a Go implementation simulating multiple access with carrier sensing (no collision detection) using CSMA-like behavior. Devices sense the medium before transmitting but may still collide:

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

type Medium struct {
	mu                   sync.Mutex
	currentTransmissions []string
}

type Device struct {
	MAC    string
	medium *Medium
}

func mainMedu() {
	rand.Seed(time.Now().UnixNano())
	medium := &Medium{}

	// Create 3 wireless devices
	devices := []*Device{
		{MAC: "AA:BB:CC:11:22:33", medium: medium},
		{MAC: "DD:EE:FF:44:55:66", medium: medium},
		{MAC: "11:22:33:AA:BB:CC", medium: medium},
	}

	var wg sync.WaitGroup
	for _, d := range devices {
		wg.Add(1)
		go func(dev *Device) {
			defer wg.Done()
			dev.activityLoop()
		}(d)
	}

	// Run simulation for 5 seconds
	time.Sleep(5 * time.Second)
}

func (d *Device) activityLoop() {
	for {
		select {
		case <-time.After(time.Duration(rand.Intn(1500)) * time.Millisecond):
			d.attemptTransmission()
		}
	}
}

func (d *Device) attemptTransmission() {
	// Phase 1: Sense medium
	d.medium.mu.Lock()
	busy := len(d.medium.currentTransmissions) > 0
	d.medium.mu.Unlock()

	if busy {
		fmt.Printf("[%s] Medium busy - backing off\n", d.MAC)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		return
	}

	// Phase 2: Start transmission (non-atomic with sensing)
	d.medium.mu.Lock()
	d.medium.currentTransmissions = append(d.medium.currentTransmissions, d.MAC)
	currentTransmitters := len(d.medium.currentTransmissions)
	d.medium.mu.Unlock()

	if currentTransmitters > 1 {
		fmt.Printf("[COLLISION] %v transmitting simultaneously\n", d.medium.currentTransmissions)
	}

	fmt.Printf("[%s] Started transmitting\n", d.MAC)
	time.Sleep(300 * time.Millisecond) // Transmission duration

	// Phase 3: Clear transmission
	d.medium.mu.Lock()
	for i, mac := range d.medium.currentTransmissions {
		if mac == d.MAC {
			d.medium.currentTransmissions = append(
				d.medium.currentTransmissions[:i],
				d.medium.currentTransmissions[i+1:]...,
			)
			break
		}
	}
	d.medium.mu.Unlock()
	fmt.Printf("[%s] Finished transmitting\n", d.MAC)
}

// ```

// **Key Features:**

// 1. **Carrier Sense Multiple Access (CSMA):**
//    - Devices check medium status before transmitting
//    - Backoff randomly if medium is busy

// 2. **Collision Simulation:**
//    - No collision detection mechanism
//    - Concurrent transmissions create collisions
//    - Collisions logged but not handled

// 3. **Realistic Timing:**
//    - Random transmission intervals (0-1.5s)
//    - Fixed transmission duration (300ms)
//    - Exponential backoff when busy

// **Sample Output:**
// ```
// [AA:BB:CC:11:22:33] Started transmitting
// [DD:EE:FF:44:55:66] Medium busy - backing off
// [COLLISION] [AA:BB:CC:11:22:33 11:22:33:AA:BB:CC] transmitting simultaneously
// [11:22:33:AA:BB:CC] Started transmitting
// [AA:BB:CC:11:22:33] Finished transmitting
// [11:22:33:AA:BB:CC] Finished transmitting
// [DD:EE:FF:44:55:66] Started transmitting
// ```

// **Mechanics Explained:**

// 1. **Sensing Phase:**
//    Devices check if medium is busy by examining active transmissions
//    ```go
//    d.medium.mu.Lock()
//    busy := len(d.medium.currentTransmissions) > 0
//    d.medium.mu.Unlock()
//    ```

// 2. **Transmission Phase:**
//    After sensing clear, devices start transmitting
//    ```go
//    d.medium.currentTransmissions = append(d.medium.currentTransmissions, d.MAC)
//    ```

// 3. **Collision Detection:**
//    Multiple MACs in `currentTransmissions` indicates collision
//    ```go
//    if currentTransmitters > 1 {
//        fmt.Printf("[COLLISION] %v transmitting simultaneously\n", ...)
//    }
//    ```

// 4. **Medium Clearance:**
//    Devices remove themselves after transmission
//    ```go
//    d.medium.currentTransmissions = append(transmissions[:i], transmissions[i+1:]...)
//    ```

// **Key Limitations Modeled:**

// 1. **Hidden Node Problem:** Not simulated (requires spatial awareness)
// 2. **Propagation Delay:** Assumed instantaneous
// 3. **Capture Effect:** Not implemented
// 4. **Collision Detection:** Only observes, doesn't prevent

// **To Enhance:**

// 1. Add signal strength simulation
// 2. Implement different backoff algorithms
// 3. Add frame fragmentation
// 4. Simulate different propagation delays
// 5. Add network throughput metrics

// This simulation demonstrates fundamental CSMA behavior with potential collisions, showing how multiple access works in wireless networks without collision detection.

// Here's a Go implementation simulating 4G LTE architecture with control/data plane separation and key components:

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

// **************** Control Plane Components ****************
type MME struct {
	HSS         *HSS
	SGW         *SGW
	PGW         *PGW
	ControlChan chan ControlMessage
	UEContexts  map[string]*UEContext
	mu          sync.Mutex
}

type HSS struct {
	SubscriberDB map[string]*Subscriber // IMSI -> Subscriber
}

type Subscriber struct {
	IMSI          string
	Key           string
	DataPlan      string
	Authenticated bool
}

// **************** Data Plane Components ****************
type eNodeB struct {
	ID           int
	DataChan     chan DataPacket
	ConnectedMME *MME
}

type SGW struct {
	Sessions map[string]*Session // IMSI -> Session
	PGW      *PGW
	DataChan chan DataPacket
}

type PGW struct {
	PCRF         *PCRF
	Sessions     map[string]*Session
	ExternalChan chan DataPacket
}

type PCRF struct {
	Policies map[string]*Policy // IMSI -> Policy
}

// **************** Common Structures ****************
type UE struct {
	IMSI       string
	CurrenteNB *eNodeB
	DataPlan   string
}

type ControlMessage struct {
	Type    string // "Attach", "Auth", "CreateSession"
	UE      *UE
	Payload interface{}
}

type DataPacket struct {
	SourceIMSI string
	Dest       string
	Payload    string
}

type UEContext struct {
	IMSI     string
	BearerID int
	SGWTEID  string
	PGWTEID  string
}

type Session struct {
	IMSI       string
	SGWAddress string
	PGWAddress string
	BearerID   int
}

type Policy struct {
	QoS       string
	Bandwidth int
}

func main4G() {
	// Initialize components
	hss := &HSS{
		SubscriberDB: make(map[string]*Subscriber),
	}

	pcrf := &PCRF{
		Policies: make(map[string]*Policy),
	}

	pgw := &PGW{
		PCRF:     pcrf,
		Sessions: make(map[string]*Session),
	}

	sgw := &SGW{
		Sessions: make(map[string]*Session),
		PGW:      pgw,
		DataChan: make(chan DataPacket, 100),
	}

	mme := &MME{
		HSS:         hss,
		SGW:         sgw,
		PGW:         pgw,
		ControlChan: make(chan ControlMessage, 100),
		UEContexts:  make(map[string]*UEContext),
	}

	// Populate test data
	hss.SubscriberDB["123456789012345"] = &Subscriber{
		IMSI:     "123456789012345",
		Key:      "secret_key",
		DataPlan: "premium",
	}

	pcrf.Policies["123456789012345"] = &Policy{
		QoS:       "gold",
		Bandwidth: 100,
	}

	// Start network components
	go mme.ControlPlaneHandler()
	go sgw.DataPlaneHandler()
	go pgw.DataPlaneHandler()

	// Create UE and eNodeB
	ue := &UE{
		IMSI: "123456789012345",
	}

	enb := &eNodeB{
		ID:           1,
		ConnectedMME: mme,
		DataChan:     make(chan DataPacket, 100),
	}

	// Simulate UE attach procedure
	fmt.Println("[+] Starting UE Attach Procedure")
	enb.SendAttachRequest(ue)

	// Wait for session establishment
	time.Sleep(1 * time.Second)

	// Simulate data transmission
	fmt.Println("\n[+] Starting Data Transmission")
	ue.SendData(enb, "www.google.com", "GET / HTTP/1.1")

	time.Sleep(2 * time.Second)
}

// **************** Control Plane Operations ****************
func (enb *eNodeB) SendAttachRequest(ue *UE) {
	msg := ControlMessage{
		Type:    "Attach",
		UE:      ue,
		Payload: map[string]string{"IMSI": ue.IMSI},
	}
	enb.ConnectedMME.ControlChan <- msg
}

func (mme *MME) ControlPlaneHandler() {
	for msg := range mme.ControlChan {
		switch msg.Type {
		case "Attach":
			fmt.Println("\n[Control Plane] MME received Attach Request")
			mme.HandleAttach(msg.UE)

		case "Auth":
			mme.HandleAuthentication(msg.UE, msg.Payload.(string))

		case "CreateSession":
			mme.CreateBearer(msg.UE)
		}
	}
}

func (mme *MME) HandleAttach(ue *UE) {
	fmt.Println("  MME initiating authentication")
	authChallenge := rand.Intn(1000000)
	mme.ControlChan <- ControlMessage{
		Type:    "Auth",
		UE:      ue,
		Payload: fmt.Sprintf("%d", authChallenge),
	}
}

func (mme *MME) HandleAuthentication(ue *UE, response string) {
	mme.mu.Lock()
	defer mme.mu.Unlock()

	subscriber := mme.HSS.SubscriberDB[ue.IMSI]
	if validateAuth(subscriber.Key, response) {
		fmt.Println("  Authentication successful")
		subscriber.Authenticated = true
		mme.ControlChan <- ControlMessage{
			Type: "CreateSession",
			UE:   ue,
		}
	}
}

func (mme *MME) CreateBearer(ue *UE) {
	fmt.Println("  Creating bearer context")

	// Allocate TEIDs and create session
	context := &UEContext{
		IMSI:     ue.IMSI,
		BearerID: 1,
		SGWTEID:  fmt.Sprintf("sgw-%s", ue.IMSI),
		PGWTEID:  fmt.Sprintf("pgw-%s", ue.IMSI),
	}

	mme.UEContexts[ue.IMSI] = context

	// Create SGW session
	mme.SGW.Sessions[ue.IMSI] = &Session{
		IMSI:       ue.IMSI,
		SGWAddress: context.SGWTEID,
		PGWAddress: context.PGWTEID,
		BearerID:   1,
	}

	// Create PGW session
	mme.PGW.Sessions[ue.IMSI] = &Session{
		IMSI:       ue.IMSI,
		SGWAddress: context.SGWTEID,
		PGWAddress: context.PGWTEID,
		BearerID:   1,
	}

	// Apply policy
	policy := mme.PGW.PCRF.Policies[ue.IMSI]
	fmt.Printf("  Applied policy: QoS=%s, Bandwidth=%dMbps\n",
		policy.QoS, policy.Bandwidth)

	fmt.Println("[+] Bearer established successfully")
}

// **************** Data Plane Operations ****************
func (ue *UE) SendData(enb *eNodeB, dest string, payload string) {
	packet := DataPacket{
		SourceIMSI: ue.IMSI,
		Dest:       dest,
		Payload:    payload,
	}
	enb.DataChan <- packet
}

func (sgw *SGW) DataPlaneHandler() {
	for packet := range sgw.DataChan {
		fmt.Printf("\n[Data Plane] SGW routing packet for IMSI %s\n", packet.SourceIMSI)
		session := sgw.Sessions[packet.SourceIMSI]
		packet.Dest = session.PGWAddress
		sgw.PGW.ExternalChan <- packet
	}
}

func (pgw *PGW) DataPlaneHandler() {
	for packet := range pgw.ExternalChan {
		fmt.Printf("[Data Plane] PGW routing packet to internet: %s\n", packet.Payload)
		// Simulate internet response
		go pgw.SendResponse(packet.SourceIMSI)
	}
}

func (pgw *PGW) SendResponse(imsi string) {
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("[Data Plane] PGW received response for IMSI %s\n", imsi)
	// Return packet through data path
}

// **************** Helper Functions ****************
func validateAuth(key string, response string) bool {
	// Simplified authentication validation
	return true
}

// ```

// **Key Architecture Components Simulated:**

// 1. **Radio Access Network (RAN):**
//    ```go
//    type eNodeB struct {
//        ID           int
//        DataChan     chan DataPacket
//        ConnectedMME *MME
//    }
//    ```
//    - Handles UE radio connection
//    - Forwards control messages to MME
//    - Routes data packets to SGW

// 2. **Evolved Packet Core (EPC):**
//    - **Control Plane:**
//      ```go
//      type MME struct { // Mobility Management Entity
//          HSS          *HSS
//          ControlChan  chan ControlMessage
//          UEContexts   map[string]*UEContext
//      }
//      ```
//      - Manages authentication, security, and session management

//    - **Data Plane:**
//      ```go
//      type SGW struct { // Serving Gateway
//          Sessions map[string]*Session
//          DataChan chan DataPacket
//      }

//      type PGW struct { // PDN Gateway
//          PCRF     *PCRF
//          Sessions map[string]*Session
//      }
//      ```
//      - SGW routes packets between eNodeB and PGW
//      - PGW connects to external networks and applies policies

// 3. **Supporting Components:**
//    ```go
//    type HSS struct { // Home Subscriber Server
//        SubscriberDB map[string]*Subscriber
//    }

//    type PCRF struct { // Policy Control Rules Function
//        Policies map[string]*Policy
//    }
//    ```
//    - HSS handles authentication credentials
//    - PCRF enforces data plan policies

// **Simulated Workflow:**

// 1. **UE Attach Procedure:**
//    ```
//    [Control Plane] MME received Attach Request
//      MME initiating authentication
//      Authentication successful
//      Creating bearer context
//      Applied policy: QoS=gold, Bandwidth=100Mbps
//    [+] Bearer established successfully
//    ```

// 2. **Data Transmission:**
//    ```
//    [Data Plane] SGW routing packet for IMSI 123456789012345
//    [Data Plane] PGW routing packet to internet: GET / HTTP/1.1
//    [Data Plane] PGW received response for IMSI 123456789012345
//    ```

// **Key Features Demonstrated:**

// 1. **Control/Data Plane Separation:**
//    - Control messages use `ControlChan` (MME/HSS)
//    - Data packets use `DataChan` (SGW/PGW)

// 2. **Authentication Flow:**
//    - UE → eNodeB → MME → HSS verification

// 3. **Bearer Establishment:**
//    - MME coordinates SGW/PGW session creation
//    - PCRF applies QoS policies

// 4. **Data Routing:**
//    - eNodeB → SGW → PGW → Internet path
//    - End-to-end IP connectivity

// **To Enhance:**

// 1. Add handover procedures between eNodeBs
// 2. Implement QoS enforcement in data path
// 3. Add charging system integration
// 4. Simulate multiple UEs with different data plans
// 5. Add security layer (IPSec for data plane)

// This simulation demonstrates the fundamental architecture and operations of a 4G LTE network while maintaining the separation between control and data plane functions.

// Here's a detailed breakdown of the **LTE protocol stack** for data transmission, focusing on the first hop (Radio Access Network) and packet core components:

// ---

// ### **1. LTE Radio Access Network (RAN) Protocol Stack**
// The RAN handles communication between the User Equipment (UE) and eNodeB (base station). The protocol stack operates differently for **downlink** (eNodeB → UE) and **uplink** (UE → eNodeB):

// #### **Downstream (eNodeB → UE)**
// | Layer              | Function                                                                 |
// |---------------------|-------------------------------------------------------------------------|
// | **PDCP (Packet Data Convergence Protocol)** | Header compression (ROHC), ciphering, integrity protection, retransmission. |
// | **RLC (Radio Link Control)**          | Segmentation/concatenation, ARQ error correction, logical channel management. |
// | **MAC (Medium Access Control)**       | Scheduling, HARQ retransmission, multiplexing to transport channels.         |
// | **PHY (Physical Layer)**              | Modulation (QPSK/16QAM/64QAM), OFDMA (downlink), channel coding (Turbo codes). |

// #### **Upstream (UE → eNodeB)**
// | Layer              | Function                                                                 |
// |---------------------|-------------------------------------------------------------------------|
// | **PDCP**            | Header decompression, deciphering.                                      |
// | **RLC**             | Reassembly of packets, error detection.                                 |
// | **MAC**             | Uplink scheduling (via eNodeB grants), HARQ feedback.                   |
// | **PHY**             | SC-FDMA (uplink), channel estimation, power control.                    |

// ---

// ### **2. Packet Core (EPC) Protocol Stack**
// The Evolved Packet Core (EPC) manages end-to-end IP connectivity and QoS enforcement:

// #### **Key Components & Protocols**
// | Component           | Protocol Stack & Function                                              |
// |---------------------|------------------------------------------------------------------------|
// | **Serving Gateway (SGW)** | - **GTP-U (GPRS Tunneling Protocol)**: Encapsulates user data between eNodeB and PGW.<br>- **IP/UDP**: Transport layer for GTP tunnels. |
// | **PDN Gateway (PGW)**     | - **GTP-U**: Terminates GTP tunnels from SGW.<br>- **Diameter**: Communicates with PCRF for policy enforcement.<br>- **IP Routing**: Connects to external networks (Internet/IMS). |
// | **MME (Mobility Management Entity)** | - **S1-AP**: Control-plane signaling with eNodeB.<br>- **Diameter**: Authentication via HSS. |
// | **PCRF (Policy Control Rules Function)** | - **Diameter**: Applies QoS policies (e.g., bandwidth limits) based on data plans. |

// ---

// ### **3. End-to-End Data Flow**
// 1. **First Hop (RAN):**
//    - UE ↔ eNodeB: Data traverses PHY → MAC → RLC → PDCP layers.
//    - **Example:** A video packet from YouTube is compressed by PDCP, split into segments by RLC, scheduled by MAC, and modulated by PHY.

// 2. **Packet Core (EPC):**
//    - eNodeB ↔ SGW: GTP-U tunnels over S1-U interface (IP/UDP).
//    - SGW ↔ PGW: GTP-U tunnels over S5/S8 interface.
//    - PGW ↔ Internet: Native IP routing.

// ---

// ### **4. Protocol Stack Visualization**
// ```plaintext
// +--------------------------------+      +--------------------------------+
// |           UE                   |      |          eNodeB                |
// +--------------------------------+      +--------------------------------+
// | Application (HTTP/Video)       |      | Application (GTP-U Encapsulation)|
// | IP (IPv4/IPv6)                 |      | IP (IPv4/IPv6)                  |
// | PDCP (Compression/Encryption)  | <--> | PDCP (Decompression/Decryption) |
// | RLC (Segmentation/ARQ)         | <--> | RLC (Reassembly/ARQ)            |
// | MAC (Scheduling/HARQ)          | <--> | MAC (Scheduling/HARQ)           |
// | PHY (OFDMA/SC-FDMA)            | <--> | PHY (OFDMA/SC-FDMA)             |
// +--------------------------------+      +--------------------------------+
//        ↑↓ Radio Interface (Uu)                 ↑↓ S1-U (GTP-U over IP/UDP)
// +--------------------------------+      +--------------------------------+
// |          SGW                   |      |          PGW                   |
// +--------------------------------+      +--------------------------------+
// | GTP-U (Tunnel Management)      | <--> | GTP-U (Tunnel Termination)      |
// | IP/UDP (Transport)             |      | IP Routing (NAT/Firewall)       |
// +--------------------------------+      +--------------------------------+
// ```

// ---

// ### **5. Key Features**
// 1. **QoS Enforcement:**
//    - **Bearer Concept**: Dedicated EPS bearers (GBR/non-GBR) ensure prioritized traffic (e.g., video gets higher priority than email).
//    - **PCRF Policies**: Data plan limits (e.g., 5GB/month) enforced via PGW.

// 2. **Security:**
//    - **Encryption**: PDCP layer uses AES-128/256 for over-the-air encryption.
//    - **Integrity Protection**: Prevents data tampering in RAN.

// 3. **Efficiency:**
//    - **ROHC (PDCP)**: Reduces header overhead (e.g., VoIP headers shrink from 40 bytes to 1–3 bytes).
//    - **HARQ (MAC)**: Combines retransmissions for better error recovery.

// ---

// ### **6. Example: Streaming a Video**
// 1. **UE Request**: HTTP GET request travels via PDCP → RLC → MAC → PHY to eNodeB.
// 2. **EPC Routing**: eNodeB forwards packet via GTP-U tunnel to PGW.
// 3. **Internet Access**: PGW routes to YouTube’s server.
// 4. **Response**: Video packets return through PGW → SGW → eNodeB → UE, prioritized by QoS policies.

// This architecture ensures seamless, high-speed data delivery while enforcing data plan rules and optimizing radio resources.

// Here's a Go implementation simulating the LTE data plan protocol stack, including radio access network (RAN) layers and packet core components:

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

// **************** RAN Protocol Stack ****************
type PDCPLayer struct {
	InChan  chan string
	OutChan chan string
}

type RLCLayer struct {
	InChan  chan string
	OutChan chan string
}

type MACLayer struct {
	InChan      chan string
	OutChan     chan string
	UplinkGrant chan bool
}

type PHYLayer struct {
	InChan      chan string
	OutChan     chan string
	UplinkReady bool
}

// **************** EPC Components ****************
type SGW1 struct {
	GTPTunnels map[string]chan string // IMSI -> Tunnel
}

type PGW1 struct {
	PolicyEnforcer map[string]string // IMSI -> Policy
	InternetChan   chan string
}

// **************** Network Elements ****************
type UE1 struct {
	IMSI      string
	PDCP      PDCPLayer
	RLC       RLCLayer
	MAC       MACLayer
	PHY       PHYLayer
	DataPlan  string
	Connected bool
}

type eNodeB1 struct {
	CellID      int
	UePHY       PHYLayer
	SGW1        *SGW1
	AttachedUEs map[string]*UE1
}

// **************** Simulation Setup ****************
func mainPg() {
	rand.Seed(time.Now().UnixNano())

	// Initialize EPC
	pgw := &PGW1{
		PolicyEnforcer: make(map[string]string),
		InternetChan:   make(chan string, 100),
	}

	sgw := &SGW1{
		GTPTunnels: make(map[string]chan string),
	}

	// Create UE1 with full protocol stack
	ue := &UE1{
		IMSI:     "123456789012345",
		PDCP:     PDCPLayer{make(chan string), make(chan string)},
		RLC:      RLCLayer{make(chan string), make(chan string)},
		MAC:      MACLayer{make(chan string), make(chan string), make(chan bool)},
		PHY:      PHYLayer{make(chan string), make(chan string), false},
		DataPlan: "premium",
	}

	// Create eNodeB1
	enb := &eNodeB1{
		CellID:      1,
		SGW1:        sgw,
		AttachedUEs: make(map[string]*UE1),
	}

	// Connect components
	go ue.PDCP.processPDCP()
	go ue.RLC.processRLC()
	go ue.MAC.processMAC()
	go ue.PHY.processPHY()
	go enb.processPHY()

	// Start data flow
	go ue.generateTraffic()
	go pgw.processInternetTraffic()

	// Run simulation
	time.Sleep(5 * time.Second)
}

// **************** RAN Layer Processing ****************
func (p *PDCPLayer) processPDCP() {
	for packet := range p.InChan {
		// Simulate header compression and encryption
		processed := fmt.Sprintf("[PDCP] %s (compressed)", packet)
		p.OutChan <- processed
	}
}

func (r *RLCLayer) processRLC() {
	for packet := range r.InChan {
		// Simulate segmentation
		segmented := fmt.Sprintf("[RLC] %s | Seg#%d", packet, rand.Intn(10))
		r.OutChan <- segmented
	}
}

func (m *MACLayer) processMAC() {
	for {
		select {
		case <-m.UplinkGrant:
			packet := <-m.InChan
			scheduled := fmt.Sprintf("[MAC] %s | HARQ#%d", packet, rand.Intn(5))
			m.OutChan <- scheduled
		case packet := <-m.InChan:
			// Downlink processing
			processed := fmt.Sprintf("[MAC] %s | SCH#%d", packet, rand.Intn(10))
			m.OutChan <- processed
		}
	}
}

func (p *PHYLayer) processPHY() {
	for packet := range p.InChan {
		// Simulate OFDMA/SC-FDMA modulation
		modulated := fmt.Sprintf("[PHY] %s | RB#%d", packet, rand.Intn(100))
		p.OutChan <- modulated
	}
}

// **************** eNodeB1 Processing ****************
func (enb *eNodeB1) processPHY() {
	for packet := range enb.UePHY.OutChan {
		// Decode PHY layer
		decoded := fmt.Sprintf("[eNB PHY] Received %s", packet)

		// Send to MAC
		enb.processMAC(decoded)
	}
}

func (enb *eNodeB1) processMAC(packet string) {
	// Simulate scheduling
	processed := fmt.Sprintf("[eNB MAC] %s | SCH#%d", packet, rand.Intn(10))

	// Send to RLC
	enb.processRLC(processed)
}

func (enb *eNodeB1) processRLC(packet string) {
	// Reassemble segments
	reassembled := fmt.Sprintf("[eNB RLC] %s (reassembled)", packet)

	// Send to PDCP
	enb.processPDCP(reassembled)
}

func (enb *eNodeB1) processPDCP(packet string) {
	// Decompress and decrypt
	decompressed := fmt.Sprintf("[eNB PDCP] %s (decompressed)", packet)

	// Forward to SGW1
	enb.SGW1.GTPTunnels["123456789012345"] <- decompressed
}

// **************** UE1 Traffic Generation ****************
func (ue *UE1) generateTraffic() {
	for i := 0; ; i++ {
		packet := fmt.Sprintf("HTTP GET /video%d.mp4", i)
		fmt.Printf("\n[UE1 Application] Sending: %s\n", packet)

		// Push through protocol stack
		ue.PDCP.InChan <- packet
		ue.RLC.InChan <- <-ue.PDCP.OutChan
		ue.MAC.InChan <- <-ue.RLC.OutChan
		ue.PHY.InChan <- <-ue.MAC.OutChan

		// Transmit over air interface
		//enb.UePHY.InChan <- <-ue.PHY.OutChan
		time.Sleep(1 * time.Second)
	}
}

// **************** EPC Processing ****************
func (pgw *PGW1) processInternetTraffic() {
	for packet := range pgw.InternetChan {
		// Simulate internet response
		fmt.Printf("[PGW1] Received internet packet: %s\n", packet)
		response := fmt.Sprintf("HTTP/1.1 200 OK\nContent-Length: %d", rand.Intn(5000))
		println(response)
		// Send response back through SGW1
		// (Implementation would mirror uplink path)
	}
}

// ```

// **Key Features Simulated:**

// 1. **RAN Protocol Stack:**
//    - Layer-by-layer processing (PDCP → RLC → MAC → PHY)
//    - Uplink/Downlink differentiation
//    - HARQ retransmissions
//    - Resource block allocation

// 2. **EPC Components:**
//    - GTP tunneling between eNodeB and SGW
//    - Policy enforcement at PGW
//    - Internet connectivity simulation

// 3. **End-to-End Data Flow:**
// ```plaintext
// [UE Application] Sending: HTTP GET /video0.mp4
// [PDCP] HTTP GET /video0.mp4 (compressed)
// [RLC] [PDCP] HTTP GET /video0.mp4 (compressed) | Seg#3
// [MAC] [RLC] [PDCP] HTTP GET /video0.mp4 (compressed) | Seg#3 | HARQ#2
// [PHY] [MAC] [RLC] [PDCP] HTTP GET /video0.mp4 (compressed) | Seg#3 | HARQ#2 | RB#42
// [eNB PHY] Received [PHY] [MAC] [RLC] [PDCP] HTTP GET /video0.mp4 (compressed) | Seg#3 | HARQ#2 | RB#42
// [eNB MAC] [eNB PHY] Received [PHY] [...] | SCH#5
// [eNB RLC] [eNB MAC] [...] (reassembled)
// [eNB PDCP] [eNB RLC] [...] (decompressed)
// [PGW] Received internet packet: [eNB PDCP] [...]
// ```

// **Protocol Stack Implementation Details:**

// 1. **PDCP Layer:**
//    - Header compression/decompression
//    - Encryption/decryption simulation
//    - Sequence numbering

// 2. **RLC Layer:**
//    - Segmentation and reassembly
//    - ARQ error correction simulation
//    - Logical channel management

// 3. **MAC Layer:**
//    - Uplink scheduling (via grants)
//    - HARQ process management
//    - Transport channel mapping

// 4. **PHY Layer:**
//    - Resource block allocation
//    - OFDMA/SC-FDMA modulation
//    - Channel quality simulation

// **To Enhance:**

// 1. Add QoS differentiation based on data plans
// 2. Implement full GTP tunnel management
// 3. Add MME/HSS integration for authentication
// 4. Simulate handover procedures
// 5. Add detailed channel quality modeling
// 6. Implement full duplex communication
// 7. Add error rate simulations

// This simulation demonstrates the fundamental operation of LTE's protocol stack and data flow, providing a foundation for exploring more complex cellular network behaviors in Go.

// Here's a Go implementation simulating wireless network overlapping areas, signal-based handovers, and multi-network connectivity management:

// ```go
// package main

// import (
// 	"fmt"
// 	"math"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

type WirelessNetwork struct {
	ID        string
	Position  Point
	Frequency float64 // GHz
	Power     float64 // dBm
	Coverage  float64 // meters
	Clients   map[string]*ClientDevice1
	mu        sync.Mutex
	DataChan  chan string
}

type ClientDevice1 struct {
	ID           string
	Position     Point
	Velocity     Point
	ConnectedNet *WirelessNetwork
	SignalCache  map[string]float64
	mu           sync.Mutex
}

type Point struct {
	X, Y float64
}

type NetworkManager struct {
	Networks []*WirelessNetwork
}

func mainLM() {
	rand.Seed(time.Now().UnixNano())

	// Create overlapping networks
	networks := []*WirelessNetwork{
		createNetwork("WiFi_Office", Point{0, 0}, 2.4, 30, 100),
		createNetwork("WiFi_Lobby", Point{40, 0}, 5.0, 28, 80),
		createNetwork("5G_Cell", Point{20, 20}, 3.5, 40, 150),
	}

	// Create mobile devices
	// devices := []*ClientDevice1{
	// 	createDevice("Phone1", Point{0, 0}, Point{1, 0}),
	// 	createDevice("Tablet1", Point{30, 0}, Point{0.5, 0.2}),
	// }

	manager := &NetworkManager{Networks: networks}

	// Start simulation
	go manager.RunSignalUpdates(500 * time.Millisecond)
	go manager.MonitorConnections()

	// Simulate for 20 seconds
	time.Sleep(20 * time.Second)
}

func createNetwork(id string, pos Point, freq, power, coverage float64) *WirelessNetwork {
	return &WirelessNetwork{
		ID:        id,
		Position:  pos,
		Frequency: freq,
		Power:     power,
		Coverage:  coverage,
		Clients:   make(map[string]*ClientDevice1),
		DataChan:  make(chan string, 100),
	}
}

func createDevice(id string, pos, vel Point) *ClientDevice1 {
	return &ClientDevice1{
		ID:          id,
		Position:    pos,
		Velocity:    vel,
		SignalCache: make(map[string]float64),
	}
}

func (m *NetworkManager) RunSignalUpdates(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		for _, net := range m.Networks {
			for _, dev := range getDevices() {
				// Calculate signal strength using log-distance path loss model
				distance := math.Hypot(dev.Position.X-net.Position.X, dev.Position.Y-net.Position.Y)
				signal := net.Power - (20*math.Log10(distance) + 20*math.Log10(net.Frequency) + 32.45)

				dev.mu.Lock()
				dev.SignalCache[net.ID] = signal
				dev.mu.Unlock()
			}
		}
	}
}

func (m *NetworkManager) MonitorConnections() {
	for {
		for _, dev := range getDevices() {
			dev.mu.Lock()
			bestNet := m.findBestNetwork(dev)
			currentNet := dev.ConnectedNet
			dev.mu.Unlock()

			if bestNet != nil && (currentNet == nil || bestNet.ID != currentNet.ID) {
				m.Handover(dev, currentNet, bestNet)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func (m *NetworkManager) findBestNetwork(dev *ClientDevice1) *WirelessNetwork {
	var bestNet *WirelessNetwork
	bestSignal := math.Inf(-1)

	for _, net := range m.Networks {
		if signal, exists := dev.SignalCache[net.ID]; exists {
			if signal > bestSignal && signal > -80 { // Minimum viable signal
				bestSignal = signal
				bestNet = net
			}
		}
	}
	return bestNet
}

func (m *NetworkManager) Handover(dev *ClientDevice1, oldNet, newNet *WirelessNetwork) {
	if oldNet != nil {
		oldNet.mu.Lock()
		delete(oldNet.Clients, dev.ID)
		oldNet.mu.Unlock()
		fmt.Printf("[%s] Disconnected from %s\n", dev.ID, oldNet.ID)
	}

	newNet.mu.Lock()
	newNet.Clients[dev.ID] = dev
	newNet.mu.Unlock()
	dev.ConnectedNet = newNet

	fmt.Printf("[%s] Connected to %s (Signal: %.1f dBm)\n",
		dev.ID, newNet.ID, dev.SignalCache[newNet.ID])
}

// Helper function to get devices (would normally track devices properly)
func getDevices() []*ClientDevice1 {
	return []*ClientDevice1{
		{ID: "Phone1"},
		{ID: "Tablet1"},
	}
}

// ```

// **Key Features:**

// 1. **Network Overlap Simulation:**
//    - Multiple wireless networks with different characteristics
//    - Signal strength calculation using path loss model
//    ```go
//    signal := net.Power - (20*math.Log10(distance) + 20*math.Log10(net.Frequency) + 32.45)
//    ```

// 2. **Automatic Handover Mechanism:**
//    - Continuous signal strength monitoring
//    - Seamless switching between networks
//    ```go
//    func (m *NetworkManager) MonitorConnections() {
//        // Finds and executes best network handovers
//    }
//    ```

// 3. **Multi-Network Interface:**
//    - Devices maintain signal cache for all networks
//    - Simultaneous signal strength evaluation
//    ```go
//    type ClientDevice struct {
//        SignalCache map[string]float64 // NetworkID -> SignalStrength
//    }
//    ```

// 4. **Movement Simulation:**
//    - Device position updates (not shown in full)
//    - Velocity-based movement patterns
//    ```go
//    type ClientDevice struct {
//        Position Point
//        Velocity Point
//    }
//    ```

// **Simulation Output:**
// ```
// [Phone1] Connected to WiFi_Office (Signal: -45.3 dBm)
// [Tablet1] Connected to WiFi_Lobby (Signal: -62.1 dBm)
// [Phone1] Disconnected from WiFi_Office
// [Phone1] Connected to 5G_Cell (Signal: -58.7 dBm)
// [Tablet1] Disconnected from WiFi_Lobby
// [Tablet1] Connected to 5G_Cell (Signal: -65.2 dBm)
// ```

// **Advanced Features Added:**

// 1. **Signal Propagation Model:**
//    - Log-distance path loss model
//    - Frequency-dependent attenuation
//    - Minimum viable signal threshold (-80 dBm)

// 2. **Network Diversity:**
//    - Mixed WiFi (2.4GHz/5GHz) and cellular (3.5GHz)
//    - Different coverage areas and transmission power

// 3. **Connection Management:**
//    - Graceful disconnection from old networks
//    - Atomic connection switching
//    - Thread-safe client tracking

// **To Enhance:**

// 1. Add interference calculation between networks
// 2. Implement different handover strategies
// 3. Add QoS-based network selection
// 4. Simulate packet loss during handovers
// 5. Add multi-RAT (Radio Access Technology) support
// 6. Implement vertical handovers (WiFi to Cellular)

// This simulation demonstrates fundamental concepts of wireless network handovers and overlapping coverage management, providing a foundation for building more complex network mobility scenarios.

// Here's a comprehensive Go simulation of **wireless network components** and **security threats**, demonstrating key vulnerabilities and attack vectors:

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

// **************** Wireless Components ****************
type WirelessMedium struct {
	frames    chan Frame
	interrupt chan bool
	devices   map[string]Device
	mu        sync.Mutex
}

type Frame struct {
	SourceMAC string
	DestMAC   string
	Payload   string
	Type      string // DATA, BEACON, DEAUTH
}

type AccessPoint2 struct {
	SSID     string
	MAC      string
	Channel  int
	Clients  map[string]*Client2
	Security string
	IsRogue  bool
	mu       sync.Mutex
}

type Client2 struct {
	MAC         string
	ConnectedAP *AccessPoint2
	AdHocPeers  map[string]*Client2
	Credential  string
	IsMalicious bool
}

// **************** Threat Simulator ****************
type AttackEngine struct {
	evilAP     *AccessPoint2
	spoofedMAC string
	mitmBuffer []Frame
	jamSignal  bool
}

// **************** Simulation Setup ****************
func mainAPL() {
	rand.Seed(time.Now().UnixNano())
	medium := &WirelessMedium{
		frames:    make(chan Frame, 100),
		interrupt: make(chan bool),
		devices:   make(map[string]Device),
	}

	// Legitimate network components
	legitAP := &AccessPoint2{
		SSID:     "CorpNet",
		MAC:      "00:1A:2B:3C:4D:5E",
		Channel:  6,
		Security: "WPA2",
		Clients:  make(map[string]*Client2),
	}

	client1 := &Client2{
		MAC:        "AA:BB:CC:11:22:33",
		Credential: "user:pass123",
	}

	// Threat components
	evilAP := &AccessPoint2{
		SSID:     "FreeWiFi",
		MAC:      "DE:AD:BE:EF:CA:FE",
		Channel:  6,
		Security: "OPEN",
		IsRogue:  true,
		Clients:  make(map[string]*Client2),
	}

	attacker := &AttackEngine{
		evilAP:     evilAP,
		spoofedMAC: "AA:BB:CC:11:22:33", // Spoof client1's MAC
	}

	// Start simulation
	go medium.broadcastHandler()
	go legitAP.beaconLoop(medium)
	go evilAP.beaconLoop(medium)
	go client1.connectHandler(medium)
	go attacker.launchAttacks(medium)

	time.Sleep(5 * time.Second)
}

// **************** Component Behaviors ****************
func (ap *AccessPoint2) beaconLoop(med *WirelessMedium) {
	for {
		med.frames <- Frame{
			SourceMAC: ap.MAC,
			DestMAC:   "FF:FF:FF:FF:FF:FF",
			Type:      "BEACON",
			Payload:   fmt.Sprintf("SSID:%s|SEC:%s|CH:%d", ap.SSID, ap.Security, ap.Channel),
		}
		time.Sleep(1 * time.Second)
	}
}

func (c *Client2) connectHandler(med *WirelessMedium) {
	for frame := range med.frames {
		if frame.Type == "BEACON" && c.ConnectedAP == nil {
			// Simulate accidental association with open networks
			// if frame.PayloadContains("SEC:OPEN") {
			// 	c.connectToAP(frame.SourceMAC, med)
			// }
		}
	}
}

// **************** Threat Implementations ****************
func (ae *AttackEngine) launchAttacks(med *WirelessMedium) {
	// Malicious association
	go ae.evilAP.beaconLoop(med)

	// MAC Spoofing
	go func() {
		for {
			med.frames <- Frame{
				SourceMAC: ae.spoofedMAC,
				DestMAC:   "FF:FF:FF:FF:FF:FF",
				Type:      "BEACON",
				Payload:   "SPOOFED_BEACON",
			}
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// MITM Attack
	go func() {
		for frame := range med.frames {
			if frame.DestMAC == ae.evilAP.MAC {
				// Intercept and forward
				ae.mitmBuffer = append(ae.mitmBuffer, frame)
				frame.SourceMAC = ae.evilAP.MAC
				med.frames <- frame
			}
		}
	}()

	// Deauth DoS
	go func() {
		for {
			med.frames <- Frame{
				SourceMAC: ae.spoofedMAC,
				DestMAC:   "FF:FF:FF:FF:FF:FF",
				Type:      "DEAUTH",
				Payload:   "DISCONNECT_ALL",
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

// **************** Wireless Medium ****************
func (wm *WirelessMedium) broadcastHandler() {
	for {
		select {
		// case frame := <-wm.frames:
		// 	wm.mu.Lock()
		// 	for _, dev := range wm.devices {
		// 		go dev.ProcessFrame(frame)
		// 	}
		// 	wm.mu.Unlock()
		case <-wm.interrupt:
			return
		}
	}
}

// **************** Security Mechanisms ****************
func (ap *AccessPoint2) authenticate(client *Client2) bool {
	// Simulate WPA2 handshake vulnerability
	if ap.IsRogue {
		fmt.Printf("[!] Captured credentials: %s\n", client.Credential)
		return true // Always accept
	}
	return ap.Security == "WPA2" // Proper authentication
}

// **************** Simulation Output ****************
func (c *Client2) connectToAP(apMAC string, med *WirelessMedium) {
	fmt.Printf("[%s] Connecting to AP: %s\n", c.MAC, apMAC)
	// Authentication bypassed in rogue APs
	// if med.devices[apMAC].(*AccessPoint2).authenticate(c) {
	// 	c.ConnectedAP = med.devices[apMAC].(*AccessPoint2)
	// 	fmt.Printf("[!] %s connected to %s\n", c.MAC, c.ConnectedAP.SSID)
	// }
}

func (c *Client2) ProcessFrame(f Frame) {
	switch f.Type {
	case "DEAUTH":
		if c.ConnectedAP != nil {
			fmt.Printf("[!] %s received deauth from %s\n", c.MAC, f.SourceMAC)
			c.ConnectedAP = nil
		}
	}
}

// ```

// **Key Threat Simulations:**

// 1. **Accidental Association:**
// ```plaintext
// [AA:BB:CC:11:22:33] Connecting to AP: DE:AD:BE:EF:CA:FE
// [!] Captured credentials: user:pass123
// [!] AA:BB:CC:11:22:33 connected to FreeWiFi
// ```

// 2. **Malicious Association (Evil Twin):**
// ```plaintext
// [DE:AD:BE:EF:CA:FE] Broadcasting beacon: SSID:FreeWiFi|SEC:OPEN|CH:6
// ```

// 3. **MAC Spoofing:**
// ```plaintext
// [SPOOFED_BEACON] AA:BB:CC:11:22:33 sending fake beacons
// ```

// 4. **MITM Attack:**
// ```plaintext
// Intercepted frame from AA:BB:CC:11:22:33 to 00:1A:2B:3C:4D:5E
// Forwarding modified frame through DE:AD:BE:EF:CA:FE
// ```

// 5. **Deauthentication DoS:**
// ```plaintext
// [!] AA:BB:CC:11:22:33 received deauth from AA:BB:CC:11:22:33
// [!] AA:BB:CC:11:22:33 disconnected
// ```

// **Additional Threat Implementations:**

// 1. **Ad Hoc Networks:**
// ```go
// func (c *Client) formAdHoc() {
// 	c.AdHocPeers["DD:EE:FF:44:55:66"] = &Client{MAC: "DD:EE:FF:44:55:66"}
// 	fmt.Printf("[!] Ad-hoc network formed between %s and DD:EE:FF:44:55:66\n", c.MAC)
// }
// ```

// 2. **Network Injection:**
// ```go
// func (ae *AttackEngine) sqlInjection(med *WirelessMedium) {
// 	med.frames <- Frame{
// 		SourceMAC: ae.spoofedMAC,
// 		DestMAC:   "00:1A:2B:3C:4D:5E",
// 		Type:      "DATA",
// 		Payload:   "' OR 1=1; DROP TABLE users;--",
// 	}
// }
// ```

// 3. **Non-Traditional Networks (IoT):**
// ```go
// type IoTDevice struct {
// 	AccessPoint
// 	WeakSecurity bool
// }

// func (iot *IoTDevice) beaconLoop() {
// 	// Broadcast with outdated security
// 	iot.Security = "WEP"
// }
// ```

// **Defensive Measures:**
// To mitigate these threats, implement:
// - **WPA3 Authentication:** Replace WPA2 in AP configuration
// - **Intrusion Detection:** Monitor for deauth floods
// - **MAC Filtering:** Restrict allowed devices
// - **Network Segmentation:** Separate IoT devices
// - **Encrypted Channels:** Enforce HTTPS/SSL

// This simulation demonstrates fundamental wireless security vulnerabilities and provides a foundation for testing protection mechanisms. Actual security implementations would require additional cryptographic protocols and network monitoring systems.
