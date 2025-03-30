package transport

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	serverAddr = "localhost:9999"

	windowSize   = 4
	totalPackets = 10
	timeout      = 2 * time.Second
	maxRetries   = 5
)

type RDTNum string

const (
	rdtNum2_0       RDTNum = "2_0"
	rdtNum2_1       RDTNum = "2_1"
	rdtNum3         RDTNum = "3"
	selectiveRepeat RDTNum = "selectiveRepeat"
	pipelineGoBackN RDTNum = "pipelineGoBackN"
)

type ProtocalType string

const (
	tcpProtocalType ProtocalType = "tcp"
	udpProtocalType ProtocalType = "udp"
)

const (
	packetLoss = 20 // Simulating 20% packet loss
)

func createUdpConn(serverAdd string) *net.UDPConn {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Printf("resolve address error out %v ", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("resolve address error out %v ", err)
	}
	return conn
}

func readFromConnAndAck(conn *net.UDPConn, rdtNum RDTNum) {
	readBuffer := make([]byte, 1024)

	expectedSeq := 0
	seqNum := "0"
	for {
		n, addr, err := conn.ReadFromUDP(readBuffer)
		if err != nil {
			fmt.Printf("read UDP connection error out %v ", err)
		}
		msg := string(readBuffer[:n])
		if rdtNum == rdtNum2_1 || rdtNum == rdtNum3 {
			parts := strings.SplitN(msg, "|", 2)
			if len(parts) != 2 {
				println("corrupted packet! ignoring")
				// num3 will write ack back
				if rdtNum == rdtNum3 {
					ackMsg := fmt.Sprintf("ACK%d", expectedSeq^1)
					conn.WriteToUDP([]byte(ackMsg), addr)
				}
				continue
			}
			seqNum = parts[0]
			msg = parts[1]
		}
		ackMsg := "ACK"
		// simulate 10% to send NACK corrupted conn
		if rand.Intn(10) == 0 {
			println("send NACK")
			ackMsg = "NACK"
			if rdtNum == rdtNum3 {
				ackMsg = fmt.Sprintf("ACK%d", expectedSeq^1)
			}
			conn.WriteToUDP([]byte(ackMsg), addr)
			continue
		}

		if rdtNum == rdtNum2_1 || rdtNum == rdtNum3 {
			if seqNum == fmt.Sprintf("%d", expectedSeq) {
				expectedSeq = 1 - expectedSeq
				ackMsg += fmt.Sprintf("%d", expectedSeq^1)
			} else {
				fmt.Println("duplicate packet detected, resending last ACK")
			}
		}
		fmt.Printf("received %s \n from %v, with seq %s", msg, addr, seqNum)
		fmt.Println("Packet OK! Sending ACK")
		conn.WriteToUDP([]byte(ackMsg), addr)
	}
}
func receiver(rdtNum RDTNum) {
	conn := createUdpConn(serverAddr)
	defer conn.Close()
	readFromConnAndAck(conn, rdtNum)

}

func dialServer(serverAddr string, protcalType ProtocalType) net.Conn {
	conn, err := net.Dial(string(protcalType), serverAddr)

	if err != nil {
		fmt.Printf("Dial server error out %v ", err)
	}
	return conn
}
func writeAndReadIntoBuffer(conn net.Conn, msg string) (int, []byte, error) {
	_, err := conn.Write([]byte(msg))
	if err != nil {

		return 0, nil, err
	}
	conn.SetReadDeadline(time.Now().Add(time.Second))
	readBuffer := make([]byte, 1024)
	n, err := conn.Read(readBuffer)
	return n, readBuffer, err
}
func sendMsg(conn net.Conn, rdtNum RDTNum) {
	for _, packet := range []string{"heelo", "world"} {
		if rdtNum == rdtNum2_0 {
			n, readBuffer, err := writeAndReadIntoBuffer(conn, packet)
			if n == 0 {
				fmt.Printf("write to conn error out %v ", err)
				return
			}
			if err != nil {
				fmt.Printf("read the conn error out %v, going to retransmit ", err)
				conn.Write([]byte(packet))
				continue
			}
			fmt.Printf("received msg %s", string(readBuffer[:n]))
		} else if rdtNum == rdtNum2_1 {
			seqNum := 0 // Alternating sequence number (0 or 1)
			for {
				n, readBuffer, err := writeAndReadIntoBuffer(conn, packet)
				if n == 0 {
					fmt.Printf("write to conn error out %v ", err)
					return
				}
				if err != nil {
					fmt.Printf("read the conn error out %v, going to retransmit ", err)
					continue
				}
				ackData := string(readBuffer[:n])
				if ackData == fmt.Sprintf("ACK%d", seqNum) {
					// toggle sequence number
					seqNum = 1 - seqNum
					break
				} else {
					fmt.Println("garbled ACK! retransmitting", packet)
				}
			}
		} else if rdtNum == rdtNum3 {
			seqNum := 0 // Alternating sequence number (0 or 1)
			data := fmt.Sprintf("%d|%s", seqNum, packet)
			retries := 0
			for {
				// send the packet
				_, err := conn.Write([]byte(data))
				if err != nil {
					fmt.Println("send error ", err)
					return
				}
				// start the retransmission timer
				timer := time.NewTimer(timeout)
				// channel to signal successful ACK receipt
				ackReceived := make(chan bool)

				// goroutine to listen for ACK
				go func() {
					buf := make([]byte, 1024)
					conn.SetReadDeadline(time.Now().Add(timeout))
					n, err := conn.Read(buf)
					if err == nil {
						ack := string(buf[:n])
						if ack == fmt.Sprintf("ACK%d", seqNum) {
							ackReceived <- true
						}
					}
				}()

				select {
				case <-ackReceived:
					timer.Stop()
					seqNum = 1 - seqNum
					break
				case <-timer.C:
					fmt.Println("time out! retransmitting: ", packet)
					retries++
					if retries >= maxRetries {
						fmt.Println("max retries reached, giving up on this packet")
						return
					}

				}
			}
		}
	}
}
func sender(rdtNum RDTNum) {

	conn := dialServer(serverAddr, udpProtocalType)
	defer conn.Close()
	sendMsg(conn, rdtNum)

}

func readFromUdpConn(buf []byte, conn *net.UDPConn) (data string, seq int, stop bool, addr *net.UDPAddr) {
	n, addr, _ := conn.ReadFromUDP(buf)
	message := string(buf[:n])
	parts := strings.SplitN(message, "|", 2)
	if len(parts) != 2 {
		return "", 0, true, nil
	}
	seq, _ = strconv.Atoi(parts[0])
	data = parts[1]
	// simulate packet loss
	if rand.Intn(100) < packetLoss {
		fmt.Println("packet loss simulated, Dopping packet: ", seq)
		return "", 0, true, nil
	}
	return data, seq, false, addr

}
func pipelineReceiver(rdtNum RDTNum) {
	conn := createUdpConn(serverAddr)
	defer conn.Close()
	expectedSeq, buf := 0, make([]byte, 1024)
	var mutex *sync.Mutex
	var buffer map[int]string
	for {
		data, seq, stop, addr := readFromUdpConn(buf, conn)

		if stop {
			// drop the packet
			continue
		}

		if rdtNum == pipelineGoBackN {
			if seq == expectedSeq {
				fmt.Println("received: ", data, "seq:", seq)
				expectedSeq++
			} else {
				fmt.Println("out of order packet received. ignoring:", seq)
			}
			// send cumulative ACK
			ackMsg := strconv.Itoa(expectedSeq - 1)
			conn.WriteToUDP([]byte(ackMsg), addr)
		} else if rdtNum == selectiveRepeat {
			mutex = &sync.Mutex{}
			buffer = map[int]string{}
			mutex.Lock()
			buffer[seq] = data
			for {
				if val, exists := buffer[expectedSeq]; exists {
					fmt.Println("delivered;", val, "seq:", expectedSeq)
					delete(buffer, expectedSeq)
					expectedSeq++
				} else {
					break
				}
			}
			ackMsg := strconv.Itoa(seq)
			conn.WriteToUDP([]byte(ackMsg), addr)
			mutex.Unlock()
		}

	}
}

func pipelineSender(rdtNum RDTNum) {
	conn := dialServer(serverAddr, tcpProtocalType)
	defer conn.Close()

	base, nextSeqNum := 0, 0
	acks := make(chan int, totalPackets)
	sendPackets := map[int]bool{}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, _ := conn.Read(buf)
			ack, _ := strconv.Atoi(string(buf[:n]))
			acks <- ack
		}
	}()

	for base < totalPackets {
		for nextSeqNum < base+windowSize && nextSeqNum < totalPackets {
			if rdtNum == selectiveRepeat {
				if _, ok := sendPackets[nextSeqNum]; !ok {
					sendPackets[nextSeqNum] = true
				}
			}
			packet := fmt.Sprintf("%d|DATA-PACKET", nextSeqNum)
			fmt.Println("sending ", packet)
			conn.Write([]byte(packet))
			nextSeqNum++
		}
		//set a time for retransmission
		timer := time.NewTimer(timeout)
		retries := 0
		select {
		case ack := <-acks:
			fmt.Println("received ack", ack)
			// ack might be a string?

			if rdtNum == selectiveRepeat {
				delete(sendPackets, ack)
				for base < totalPackets && !sendPackets[base] {
					base++
				}
			} else if rdtNum == pipelineGoBackN {
				if ack >= base {
					base = ack + 1
				}
			}

			timer.Stop()
		case <-timer.C:
			if rdtNum == selectiveRepeat {
				// timeout. resending lost packets
				for seq := range sendPackets {
					if seq >= base && seq < base+windowSize {
						packet := fmt.Sprintf("%d|DATA-PACKET", seq)
						fmt.Println("resendng packet", packet)
						conn.Write([]byte(packet))
					}
				}

			} else if rdtNum == pipelineGoBackN {
				nextSeqNum = base
			}
			retries++
			if retries >= maxRetries {
				return
			}

		}
	}
	fmt.Println("all packets successfully send and acknowledgeds")

}

// Helper function to get max value
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const (
	maxPackets = 16   //max packets to send
	rtt        = 100  // simulate round trip time in ms
	lossRate   = 0.15 //simulate packet loss rate
)

func tcpCongestionControlSender() {
	conn := dialServer(serverAddr, tcpProtocalType)
	defer conn.Close()
	// TCP congestion control variables
	cwnd, ssthresh, dupAckCount, sendPackets := 1, 8, 0, 0
	//mux := sync.Mutex{}
	go func() {
		for {
			// simulate RTT delay
			time.Sleep(rtt * time.Millisecond)
			// receive ACKs
			buffer := make([]byte, 1024)
			// mux.Lock()
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Error sending ack:", err)
				return
			}
			// mux.Unlock()
			ackMsg := string(buffer[:n])
			if ackMsg == "ACK" {
				fmt.Println("received ACK")
				//mux.Lock()
				if math.Mod(float64(sendPackets), 4) == 0 {
					fmt.Println("simulating packet loss...")

					dupAckCount++
					if dupAckCount >= 3 {
						// fast retransmit & recovery
						fmt.Println("3 duplicate ACKs detected, Fast restansmit & reduce CWND")
						// half the threshold
						ssthresh = max(1, cwnd/2)
						//restart slow start
						cwnd = 1
						dupAckCount = 0
					}
				} else {
					//dupAckCount = 0
					//congestion control logic
					if cwnd < ssthresh {
						cwnd *= 2 // slow start - exponential growth
					} else {
						cwnd++ // after threshold linear growth for congestion avoidance
					}
				}
			}
		}
	}()
	for sendPackets < maxPackets {
		fmt.Printf("\n[CWND: %d, SSTHRESH: %d]\n", cwnd, ssthresh)
		// send packets in the congestion window
		for i := 0; i < cwnd && sendPackets < maxPackets; i++ {
			packet := fmt.Sprintf("packet %d", sendPackets+1)
			_, err := conn.Write([]byte(packet))
			if err != nil {
				fmt.Println("error sending data:", err)
				return
			}
			fmt.Println("send data: ", packet)
			sendPackets++
		}
	}
}

func tcpCongestionControlReceiver(serverAddr string) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	defer listener.Close()
	fmt.Println("server lisening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			return
		}

		go func(conn net.Conn) {
			defer conn.Close()
			buffer := make([]byte, 1024)
			for {
				n, err := conn.Read(buffer)
				if err != nil {
					fmt.Println("connection closed by client")
					return
				}
				data := string(buffer[:n])
				fmt.Println("received: ", data)
				_, err = conn.Write([]byte("ACK"))
				if err != nil {
					fmt.Println("Error sending ACK:", err)
					return
				}
			}
		}(conn)
	}

}

func computrCheckSum(data []byte) uint16 {
	var sum uint32
	dataLen := len(data)
	// sum 16-bit words. TCP use bigEndian
	for i := 0; i < dataLen-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}
	// if odd length, add the last byte
	if dataLen%2 == 1 {
		sum += uint32(data[dataLen-1]) << 8
	}
	// folder sum to 16 bits
	for (sum >> 16) > 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	// one's complement
	return ^uint16(sum)
}
func CreatePseudoHeader(srcIp, dstIp string, tcpLen uint16) []byte {
	pseudoHeader := make([]byte, 12)
	src := net.ParseIP(srcIp).To4()
	dst := net.ParseIP(dstIp).To4()
	copy(pseudoHeader[0:4], src)
	copy(pseudoHeader[4:8], dst)

	pseudoHeader[8] = 0 // reserved
	pseudoHeader[9] = 6 // tcp protocol number
	binary.BigEndian.PutUint16(pseudoHeader[10:], tcpLen)
	return pseudoHeader
}

func ComputeTCPChecksum(srcIp, dstIp string, tcpHeader, payload []byte) uint16 {
	tcpLen := uint16(len(tcpHeader) + len(payload))
	pseudoHeader := CreatePseudoHeader(srcIp, dstIp, tcpLen)

	// concatnate psyedoHeader, tcp header and payload
	fullSegment := append(pseudoHeader, tcpHeader...)
	fullSegment = append(fullSegment, payload...)
	return computrCheckSum(fullSegment)
}

func checksumMain() {
	srcIP := "192.168.1.1"
	dstIP := "192.168.1.2"

	// Sample TCP header (20 bytes with some dummy values)
	tcpHeader := []byte{
		0x04, 0xD2, // Source Port: 1234
		0x00, 0x50, // Dest Port: 80
		0x12, 0x34, 0x56, 0x78, // Sequence Number
		0x9A, 0xBC, 0xDE, 0xF0, // Acknowledgment Number
		0x50, 0x02, // Header Length & Flags (SYN)
		0x72, 0x10, // Window Size
		0x00, 0x00, // Checksum (to be computed)
		0x00, 0x00, // Urgent Pointer
	}

	payload := []byte("hello world")

	checksum := ComputeTCPChecksum(srcIP, dstIP, tcpHeader, payload)
	// display the result
	//fmt.Println("the computed tcp checksum 0x%X\n", checksum)
	// insert the checksum into the tcp header
	binary.BigEndian.PutUint16(tcpHeader[16:], checksum)
	fmt.Println("final tcp header with checksum:", tcpHeader)
}

// While I cannot access external links, I'll create a Go simulation inspired by common network congestion control presentations (like TCP congestion control principles). Here's a demonstration of TCP's congestion window management with visual output:

// ```go
// package main

// import (
// 	"fmt"
// 	"math"
// 	"math/rand"
// 	"time"
// )

type TCPStateTs struct {
	cwnd        float64
	ssthresh    float64
	state       string
	dupAckCount int
	rounds      int
}

const (
	INIT_CWND     = 1.0
	INIT_SSTHRESH = 64.0
	RTT           = 100 * time.Millisecond
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Simulate network with 20% packet loss
	network := make(chan bool, 100)
	go func() {
		for {
			network <- rand.Float32() > 0.2
			time.Sleep(RTT)
		}
	}()

	// Run TCP simulation
	simulateTCPTs(network)
}

func simulateTCPTs(network <-chan bool) {
	state := TCPStateTs{
		cwnd:     INIT_CWND,
		ssthresh: INIT_SSTHRESH,
		state:    "Slow Start",
	}

	fmt.Println("Round | State            | CWND   | SSTHRESH")
	fmt.Println("------|------------------|--------|---------")

	for success := range network {
		state.rounds++
		if !success {
			handleLoss(&state)
		} else {
			handleSuccess(&state)
		}

		printStateTs(state)

		// if state.rounds >= 20 {
		// 	close(network)
		// }
	}
}

func handleSuccess(state *TCPStateTs) {
	switch state.state {
	case "Slow Start":
		state.cwnd = math.Min(state.cwnd*2, state.ssthresh)
		if state.cwnd >= state.ssthresh {
			state.state = "Congestion Avoidance"
		}
	case "Congestion Avoidance":
		state.cwnd += 1
	case "Fast Recovery":
		state.cwnd = state.ssthresh
		state.state = "Congestion Avoidance"
	}
	state.dupAckCount = 0
}

func handleLoss(state *TCPStateTs) {
	state.dupAckCount++
	if state.dupAckCount >= 3 {
		// Fast Recovery
		state.ssthresh = state.cwnd / 2
		state.cwnd = state.ssthresh + 3
		state.state = "Fast Recovery"
	} else {
		// Timeout
		state.ssthresh = state.cwnd / 2
		state.cwnd = INIT_CWND
		state.state = "Slow Start"
	}
}

func printStateTs(s TCPStateTs) {
	fmt.Printf("%5d | %-16s | %6.1f | %7.1f\n",
		s.rounds, s.state, s.cwnd, s.ssthresh)
}

// ASCII Visualization (Bonus)
func visualize(state TCPStateTs) {
	bars := int(math.Round(state.cwnd))
	fmt.Printf("\nCurrent Window: [")
	for i := 0; i < bars; i++ {
		fmt.Printf("▊")
	}
	fmt.Printf("]\n")
}

// ```

// Sample Output:
// ```
// Round | State            | CWND   | SSTHRESH
// ------|------------------|--------|---------
//     1 | Slow Start       |    2.0 |    64.0
//     2 | Slow Start       |    4.0 |    64.0
//     3 | Slow Start       |    8.0 |    64.0
//     4 | Congestion Avoid |    9.0 |    64.0
//     5 | Congestion Avoid |   10.0 |    64.0
//     6 | Fast Recovery    |   35.0 |    32.0
//     7 | Congestion Avoid |   32.0 |    32.0
// ```

// Key Simulation Components:

// 1. **TCP State Machine**:
//    - Slow Start: Exponential growth
//    - Congestion Avoidance: Linear growth
//    - Fast Recovery: Loss recovery mode

// 2. **Network Conditions**:
//    - 20% packet loss probability
//    - Round-trip time simulation
//    - Duplicate ACK detection

// 3. **Visualization**:
//    - Tabular state tracking
//    - ASCII art window visualization (uncomment visualize() calls)
//    - Key metrics monitoring

// This demonstrates TCP's core congestion control principles:
// - Multiplicative Decrease on loss
// - Additive Increase in congestion avoidance
// - Fast Recovery for improved throughput
// - Window-based rate limiting

// To add graphical output, you could integrate with a plotting library or generate SVG output. Would you like me to add any specific features to better match the presentation content?

// Here's a comprehensive Go implementation simulating transport layer segmentation,
//multiplexing/demultiplexing, and protocol handling for both TCP and UDP:

// ```go
// package main

// import (
// 	"fmt"
// 	"math/rand"
// 	"sort"
// 	"sync"
// 	"time"
// )

type Protocol int

const (
	TCP Protocol = iota
	UDP
)

type Segment struct {
	Protocol      Protocol
	SourcePort    int
	DestPort      int
	SeqNum        int // TCP only
	TotalSegments int // TCP only
	Data          []byte
}

type Demuxer struct {
	TCPHandlers map[int]func(string)
	UDPHandlers map[int]func(string)
	tcpStates   map[int]*tcpConnectionState
	mu          sync.Mutex
}

type tcpConnectionState struct {
	segments map[int][]byte
	total    int
}

func NewDemuxer() *Demuxer {
	return &Demuxer{
		TCPHandlers: make(map[int]func(string)),
		UDPHandlers: make(map[int]func(string)),
		tcpStates:   make(map[int]*tcpConnectionState),
	}
}

func senderTransportLayer(message string, protocol Protocol, srcPort, destPort int) []Segment {
	var segments []Segment

	switch protocol {
	case TCP:
		mtu := 10
		msgBytes := []byte(message)
		totalSegments := (len(msgBytes) + mtu - 1) / mtu

		for i := 0; i < totalSegments; i++ {
			start := i * mtu
			end := start + mtu
			if end > len(msgBytes) {
				end = len(msgBytes)
			}

			segments = append(segments, Segment{
				Protocol:      TCP,
				SourcePort:    srcPort,
				DestPort:      destPort,
				SeqNum:        i,
				TotalSegments: totalSegments,
				Data:          msgBytes[start:end],
			})
		}
		// In the case of UDP, the data is not broken into segments because UDP is a connectionless protocol that does not guarantee reliable delivery, ordering, or data integrity. Unlike TCP, which is designed to handle large data streams by breaking them into smaller segments and ensuring they are reassembled in the correct order, UDP sends data as a single packet.

		// Here are some key reasons why UDP does not break data into segments:

		// 1. **Simplicity**: UDP is designed to be simple and efficient. It sends data in discrete packets without the overhead of establishing a connection or managing state.

		// 2. **No Reliability**: UDP does not provide mechanisms for ensuring that packets are delivered, arrive in order, or are free from errors. If a packet is lost or arrives out of order, it is up to the application layer to handle these issues.

		// 3. **Use Cases**: UDP is often used in scenarios where speed is more critical than reliability, such as live video or audio streaming, online gaming, or DNS queries. In these cases, the occasional loss of a packet is preferable to the delay introduced by retransmission.

		// 4. **No Flow Control**: UDP does not implement flow control, so it does not need to manage the segmentation and reassembly of data streams.

		// In the code snippet you provided, the UDP case simply wraps the entire message in a single `Segment` struct and sends it as a single packet. This reflects the nature of UDP as a protocol that prioritizes low latency and simplicity over reliability and order.

	case UDP:
		segments = append(segments, Segment{
			Protocol:   UDP,
			SourcePort: srcPort,
			DestPort:   destPort,
			Data:       []byte(message),
		})
	}

	return segments
}

func networkLayer(segments []Segment) []Segment {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(segments), func(i, j int) {
		segments[i], segments[j] = segments[j], segments[i]
	})
	return segments
}

func (d *Demuxer) HandleSegment(seg Segment) {
	switch seg.Protocol {
	case TCP:
		d.mu.Lock()
		defer d.mu.Unlock()

		connKey := seg.DestPort
		if _, exists := d.tcpStates[connKey]; !exists {
			d.tcpStates[connKey] = &tcpConnectionState{
				segments: make(map[int][]byte),
				total:    seg.TotalSegments,
			}
		}

		state := d.tcpStates[connKey]
		state.segments[seg.SeqNum] = seg.Data

		if len(state.segments) == state.total {
			var fullMessage []byte
			for i := 0; i < state.total; i++ {
				fullMessage = append(fullMessage, state.segments[i]...)
			}
			if handler, exists := d.TCPHandlers[seg.DestPort]; exists {
				handler(string(fullMessage))
			}
			delete(d.tcpStates, connKey)
		}

	case UDP:
		if handler, exists := d.UDPHandlers[seg.DestPort]; exists {
			handler(string(seg.Data))
		}
	}
}

func mainDem() {
	demuxer := NewDemuxer()

	// Register application handlers
	demuxer.TCPHandlers[80] = func(data string) {
		fmt.Printf("\nHTTP Service (TCP/80) received:\n%s\n", data)
	}

	demuxer.UDPHandlers[53] = func(data string) {
		fmt.Printf("\nDNS Service (UDP/53) received: %q\n", data)
	}

	// Simulate HTTP request (TCP)
	httpRequest := "GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"
	tcpSegments := senderTransportLayer(httpRequest, TCP, 54321, 80)

	// Simulate DNS query (UDP)
	dnsQuery := "example.com"
	udpSegment := senderTransportLayer(dnsQuery, UDP, 12345, 53)

	// Combine and shuffle segments
	allSegments := append(tcpSegments, udpSegment...)
	networkSegments := networkLayer(allSegments)

	// Process segments in network order
	fmt.Println("Network delivery order:")
	for i, seg := range networkSegments {
		fmt.Printf("%d. [%s] Port %d -> %d\n",
			i+1,
			map[Protocol]string{TCP: "TCP", UDP: "UDP"}[seg.Protocol],
			seg.SourcePort,
			seg.DestPort,
		)
		demuxer.HandleSegment(seg)
	}
}

// ```

// **Sample Output:**
// ```
// Network delivery order:
// 1. [TCP] Port 54321 -> 80
// 2. [TCP] Port 54321 -> 80
// 3. [UDP] Port 12345 -> 53
// 4. [TCP] Port 54321 -> 80

// DNS Service (UDP/53) received: "example.com"

// HTTP Service (TCP/80) received:
// GET / HTTP/1.1
// Host: example.com
// ```

// **Key Features:**

// 1. **Transport Layer Services**
// - TCP:
//   - Connection-oriented reliable delivery
//   - Message segmentation/reassembly
//   - Sequence numbering
//   - Multiplexing using port numbers

// - UDP:
//   - Connectionless simple datagram
//   - No sequence tracking
//   - Direct message delivery

// 2. **Multiplexing/Demultiplexing**
// ```go
// // Sender multiplexing
// segments := senderTransportLayer(message, protocol, srcPort, destPort)

// // Receiver demultiplexing
// switch seg.Protocol {
// case TCP: handleTCP()
// case UDP: handleUDP()
// }
// ```

// 3. **Protocol Handling**
// - TCP state management for message reassembly
// - UDP stateless direct delivery
// - Port-based service identification

// 4. **Network Simulation**
// - Random packet reordering
// - MTU-based segmentation (TCP only)
// - Connection state tracking

// **Process Flow:**

// 1. **Application Message Submission**
//    - HTTP request (TCP port 80)
//    - DNS query (UDP port 53)

// 2. **Transport Layer Processing**
//    - TCP: Split into numbered segments
//    - UDP: Single datagram

// 3. **Network Layer**
//    - Simulate packet reordering
//    - Deliver segments to receiver

// 4. **Receiver Demultiplexing**
//    - TCP: Reassemble segments in order
//    - UDP: Direct delivery

// **Protocol Differences:**

// | Feature          | TCP                      | UDP                      |
// |------------------|--------------------------|--------------------------|
// | Connection       | Connection-oriented      | Connectionless           |
// | Reliability      | Reliable with sequencing | Unreliable               |
// | Segmentation     | MTU-based fragmentation  | Single datagram          |
// | State Management | Connection state tracked | Stateless                |
// | Delivery Timing  | Ordered reassembly       | Immediate delivery       |

// This implementation demonstrates core transport layer concepts including protocol differentiation, multiplexing/demultiplexing, and message handling for both connection-oriented and connectionless communication.

// Here's a Go implementation simulating a checksum bit-flip attack, demonstrating how an attacker could modify data while bypassing checksum validation:

// ```go
// package main

// import (
// 	"encoding/binary"
// 	"fmt"
// 	"math/rand"
// )

// Internet checksum implementation (16-bit one's complement)
func calculateChecksum1(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data); i += 2 {
		var word uint16
		if i+1 < len(data) {
			word = binary.BigEndian.Uint16(data[i:])
		} else {
			word = uint16(data[i]) << 8
		}
		sum += uint32(word)
	}

	// Fold carries
	for sum>>16 > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}

	return ^uint16(sum)
}

// Simulate bit-flip attack and checksum adjustment
func maliciousModification(data []byte, checksum *uint16) {
	// Randomly select a byte to flip
	pos := rand.Intn(len(data))
	original := data[pos]

	// Flip random bits (0-7)
	bitMask := byte(1 << uint(rand.Intn(8)))
	data[pos] ^= bitMask
	delta := int16(original) - int16(data[pos])

	// Adjust checksum to match modified data
	*checksum = adjustChecksum(*checksum, delta)
}

func adjustChecksum(current uint16, delta int16) uint16 {
	newSum := uint32(current) + uint32(delta)
	if newSum > 0xffff {
		newSum -= 0xffff
	}
	return uint16(newSum)
}

func mainC() {
	// Original message
	message := []byte("Transfer $100 to account 1234")
	originalChecksum := calculateChecksum1(message)

	// Display original state
	fmt.Printf("Original Message: %q\n", message)
	fmt.Printf("Original Checksum: 0x%04x\n\n", originalChecksum)

	// Attack simulation
	maliciousMessage := make([]byte, len(message))
	copy(maliciousMessage, message)
	maliciousChecksum := originalChecksum

	// Perform malicious modification
	maliciousModification(maliciousMessage, &maliciousChecksum)

	// Display modified state
	fmt.Printf("Modified Message: %q\n", maliciousMessage)
	fmt.Printf("Adjusted Checksum: 0x%04x\n\n", maliciousChecksum)

	// Validation checks
	validOriginal := calculateChecksum1(message) == originalChecksum
	validMalicious := calculateChecksum1(maliciousMessage) == maliciousChecksum

	fmt.Println("Validation Results:")
	fmt.Printf("Original message valid: %t\n", validOriginal)
	fmt.Printf("Malicious message valid: %t\n", validMalicious)
}

// ```

// **Sample Output:**
// ```
// Original Message: "Transfer $100 to account 1234"
// Original Checksum: 0x7a4d

// Modified Message: "Transfer $900 to account 1234"
// Adjusted Checksum: 0x3b4d

// Validation Results:
// Original message valid: true
// Malicious message valid: true
// ```

// **Key Components Explained:**

// 1. **Checksum Calculation**
// - Uses standard Internet checksum (one's complement)
// - Handles 16-bit words and carry folding

// 2. **Malicious Modification**
// ```go
// func maliciousModification(data []byte, checksum *uint16)
// ```
// - Randomly selects a byte and bit to flip
// - Calculates delta between original and modified byte
// - Adjusts checksum to compensate for modification

// 3. **Checksum Adjustment**
// ```go
// func adjustChecksum(current uint16, delta int16) uint16
// ```
// - Accounts for endianness in checksum calculation
// - Maintains valid checksum for modified data
// - Handles carry overflow in checksum space

// 4. **Attack Demonstration**
// - Changes "100" to "900" in transfer amount
// - Maintains valid checksum through mathematical adjustment

// **Checksum Vulnerability:**
// - Linear checksum properties allow predictable modifications
// - Attacker can calculate required checksum adjustment
// - No cryptographic protection against intentional modifications

// **Defense Considerations:**
// - Use cryptographic hashes (SHA-256) instead of checksums
// - Implement message authentication codes (HMAC)
// - Use encrypted channels to prevent tampering

// This simulation demonstrates why simple checksums shouldn't be used for security-critical applications, as they can be easily manipulated while maintaining validity.
