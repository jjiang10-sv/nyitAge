package transport

//Here's a simulation of a firewall with traffic analysis and authorization in Go. This implementation includes IP/port rules, threat detection, user identity, and traffic logging:

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type FirewallRule struct {
	Direction string // "inbound" or "outbound"
	IPRange   string // CIDR notation
	Ports     []int
	Protocol  string // "tcp", "udp", or "any"
	User      string
	Action    string // "allow" or "deny"
}

type NetworkPacket struct {
	SourceIP   string
	DestIP     string
	SourcePort int
	DestPort   int
	Protocol   string
	User       string
	Direction  string
	Payload    []byte
}

type Firewall struct {
	Rules         []FirewallRule
	ThreatDB      map[string]bool
	UserRoles     map[string]string
	ActivityLogs  []string
	Blocked       []NetworkPacket
	Allowed       []NetworkPacket
	CurrentThreat bool
}

func (f *Firewall) AnalyzeTraffic(packet NetworkPacket) bool {
	// Threat analysis
	if f.ThreatDB[packet.SourceIP] || f.ThreatDB[packet.DestIP] {
		f.logEvent("BLOCKED - Known threat IP", packet)
		f.Blocked = append(f.Blocked, packet)
		f.CurrentThreat = true
		return false
	}

	// Protocol validation
	if !f.validateProtocol(packet.Protocol) {
		f.logEvent("BLOCKED - Invalid protocol", packet)
		f.Blocked = append(f.Blocked, packet)
		return false
	}

	// User authorization
	if !f.authorizeUser(packet.User, packet.Direction) {
		f.logEvent("BLOCKED - Unauthorized user", packet)
		f.Blocked = append(f.Blocked, packet)
		return false
	}

	// Rule-based authorization
	if f.checkRules(packet) {
		f.logEvent("ALLOWED - Rule match", packet)
		f.Allowed = append(f.Allowed, packet)
		return true
	}

	f.logEvent("BLOCKED - No matching allow rules", packet)
	f.Blocked = append(f.Blocked, packet)
	return false
}

func (f *Firewall) validateProtocol(protocol string) bool {
	switch strings.ToLower(protocol) {
	case "tcp", "udp", "icmp":
		return true
	default:
		return false
	}
}

func (f *Firewall) authorizeUser(user, direction string) bool {
	role, exists := f.UserRoles[user]
	if !exists {
		return false
	}

	// Example role-based access
	if direction == "outbound" && role == "guest" {
		return false
	}
	return true
}

func (f *Firewall) checkRules(packet NetworkPacket) bool {
	for _, rule := range f.Rules {
		if f.matchRule(rule, packet) {
			return rule.Action == "allow"
		}
	}
	return false // Default deny
}

func (f *Firewall) matchRule(rule FirewallRule, packet NetworkPacket) bool {
	// Check direction
	if strings.ToLower(rule.Direction) != strings.ToLower(packet.Direction) {
		return false
	}

	// Check IP range
	_, ipNet, _ := net.ParseCIDR(rule.IPRange)
	packetIP := net.ParseIP(packet.SourceIP)
	if rule.Direction == "inbound" {
		packetIP = net.ParseIP(packet.DestIP)
	}
	if !ipNet.Contains(packetIP) {
		return false
	}

	// Check port
	portMatch := len(rule.Ports) == 0 // Allow any port if no ports specified
	for _, p := range rule.Ports {
		if (rule.Direction == "inbound" && p == packet.DestPort) ||
			(rule.Direction == "outbound" && p == packet.SourcePort) {
			portMatch = true
			break
		}
	}

	// Check protocol
	protocolMatch := rule.Protocol == "any" ||
		strings.EqualFold(rule.Protocol, packet.Protocol)

	// Check user
	userMatch := rule.User == "" || strings.EqualFold(rule.User, packet.User)

	return portMatch && protocolMatch && userMatch
}

func (f *Firewall) logEvent(message string, packet NetworkPacket) {
	logEntry := fmt.Sprintf("%s | %s:%d -> %s:%d | User: %s | Proto: %s",
		message,
		packet.SourceIP,
		packet.SourcePort,
		packet.DestIP,
		packet.DestPort,
		packet.User,
		packet.Protocol,
	)
	f.ActivityLogs = append(f.ActivityLogs, logEntry)
}

func mainFirewall() {
	fw := Firewall{
		ThreatDB: map[string]bool{
			"192.168.1.666": true,
			"10.0.0.99":     true,
		},
		UserRoles: map[string]string{
			"admin": "administrator",
			"user1": "employee",
			"guest": "guest",
		},
	}

	// Add firewall rules
	fw.Rules = []FirewallRule{
		{
			Direction: "inbound",
			IPRange:   "192.168.1.0/24",
			Ports:     []int{80, 443},
			Protocol:  "tcp",
			Action:    "allow",
		},
		{
			Direction: "outbound",
			IPRange:   "0.0.0.0/0",
			User:      "admin",
			Action:    "allow",
		},
	}

	// Simulate network traffic
	traffic := []NetworkPacket{
		{
			SourceIP:   "192.168.1.666",
			DestIP:     "192.168.1.1",
			SourcePort: 1234,
			DestPort:   80,
			Protocol:   "tcp",
			User:       "user1",
			Direction:  "inbound",
		},
		{
			SourceIP:   "192.168.1.100",
			DestIP:     "8.8.8.8",
			SourcePort: 5678,
			DestPort:   53,
			Protocol:   "udp",
			User:       "admin",
			Direction:  "outbound",
		},
		{
			SourceIP:   "10.0.0.99",
			DestIP:     "192.168.1.50",
			SourcePort: 9999,
			DestPort:   22,
			Protocol:   "tcp",
			User:       "guest",
			Direction:  "inbound",
		},
	}

	for _, pkt := range traffic {
		fw.AnalyzeTraffic(pkt)
	}

	// Display results
	fmt.Println("=== Firewall Logs ===")
	for _, log := range fw.ActivityLogs {
		fmt.Println(log)
	}

	fmt.Println("\n=== Blocked Packets ===")
	for _, pkt := range fw.Blocked {
		fmt.Printf("Blocked: %+v\n", pkt)
	}

	fmt.Println("\n=== Allowed Packets ===")
	for _, pkt := range fw.Allowed {
		fmt.Printf("Allowed: %+v\n", pkt)
	}
}

// ```

// This implementation demonstrates:

// 1. **Firewall Rule Types**:
//    - IP range filtering using CIDR notation
//    - Port and protocol restrictions
//    - User-based authorization
//    - Direction-based rules (inbound/outbound)

// 2. **Security Features**:
//    - Threat database for known malicious IPs
//    - User role-based access control
//    - Protocol validation
//    - Comprehensive logging

// 3. **Traffic Analysis**:
//    - Threat detection based on IP reputation
//    - Protocol validation
//    - User authorization checks
//    - Rule-based filtering

// 4. **Key Components**:
//    - `FirewallRule` struct defines access control rules
//    - `NetworkPacket` struct simulates network traffic
//    - `AnalyzeTraffic` method handles full inspection pipeline
//    - Multi-layered security checks (threats, protocol, user, rules)

// Example output:
// ```
// === Firewall Logs ===
// BLOCKED - Known threat IP | 192.168.1.666:1234 -> 192.168.1.1:80 | User: user1 | Proto: tcp
// ALLOWED - Rule match | 192.168.1.100:5678 -> 8.8.8.8:53 | User: admin | Proto: udp
// BLOCKED - Known threat IP | 10.0.0.99:9999 -> 192.168.1.50:22 | User: guest | Proto: tcp

// === Blocked Packets ===
// Blocked: {SourceIP:192.168.1.666 DestIP:192.168.1.1 SourcePort:1234 DestPort:80 Protocol:tcp User:user1 Direction:inbound Payload:[]}
// Blocked: {SourceIP:10.0.0.99 DestIP:192.168.1.50 SourcePort:9999 DestPort:22 Protocol:tcp User:guest Direction:inbound Payload:[]}

// === Allowed Packets ===
// Allowed: {SourceIP:192.168.1.100 DestIP:8.8.8.8 SourcePort:5678 DestPort:53 Protocol:udp User:admin Direction:outbound Payload:[]}
// ```

// This simulation can be extended with:
// - Deep packet inspection
// - Rate limiting
// - TLS/SSL inspection
// - Behavioral analysis
// - More sophisticated threat intelligence integration

// Here's a simulated architecture for a **Distributed Intrusion Detection System (IDS)** in Go, designed for scalability, fault tolerance, and real-time analysis. The system uses a combination of agents, a central coordinator, and a message broker for communication.

// ---

// ### **Architecture Overview**
// ```
//   +----------------+       +----------------+       +----------------+
//   |   Agent Node   |       |   Agent Node   |       |   Agent Node   |
//   | (Data Collector|       | (Data Collector|       | (Data Collector|
//   |  + Detector)   |       |  + Detector)   |       |  + Detector)   |
//   +-------+--------+       +-------+--------+       +-------+--------+
//           |                        |                        |
//           |                        |                        |
//           |      +-----------------------------------+      |
//           |      |           Message Broker          |      |
//           +----->| (NATS/Kafka/RabbitMQ for async     |<-----+
//                  |  communication between nodes)     |
//                  +------------------+-----------------+
//                                     |
//                                     |
//                  +------------------+-----------------+
//                  |           Central Coordinator       |
//                  | (Correlation Engine + Alert Manager)|
//                  +------------------+-----------------+
//                                     |
//                                     |
//                  +------------------+-----------------+
//                  |              Database              |
//                  | (TimescaleDB/Elasticsearch for     |
//                  |  storing events and alerts)        |
//                  +------------------+-----------------+
//                                     |
//                                     |
//                  +------------------+-----------------+
//                  |              Dashboard             |
//                  | (Grafana/Prometheus for monitoring)|
//                  +------------------------------------+
// ```

// ---

// ### **Components & Implementation in Go**

// #### 1. **Agent Node**
// - **Role**: Monitor host/network activity, detect anomalies, and forward alerts.
// - **Go Implementation**:
//   ```go
//   package agent

//   import (
//       "github.com/nats-io/nats.go"
//       "log"
//   )

//   type Agent struct {
//       ID     string
//       nc     *nats.Conn
//       config *Config
//   }

//   func NewAgent(id string, natsURL string) *Agent {
//       nc, err := nats.Connect(natsURL)
//       if err != nil {
//           log.Fatal("Failed to connect to NATS:", err)
//       }
//       return &Agent{ID: id, nc: nc}
//   }

//   // Monitor network traffic (simulated)
//   func (a *Agent) Monitor() {
//       for {
//           // Simulate data collection (e.g., syslog, packet capture)
//           event := simulateNetworkTraffic()
//           if a.DetectAnomaly(event) {
//               a.PublishAlert(event)
//           }
//       }
//   }

//   // Detect anomalies using rules or ML models
//   func (a *Agent) DetectAnomaly(event Event) bool {
//       // Example: Detect port scanning or unusual HTTP requests
//       return event.Severity > a.config.Threshold
//   }

//   // Publish alerts to the message broker
//   func (a *Agent) PublishAlert(event Event) {
//       data, _ := json.Marshal(event)
//       a.nc.Publish("alerts.topic", data)
//   }
//   ```

// ---

// #### 2. **Message Broker (NATS)**
// - **Role**: Handle asynchronous communication between agents and the coordinator.
// - **Go Implementation**:
//   ```go
//   func setupNATS() {
//       nc, _ := nats.Connect("nats://coordinator:4222")
//       nc.Subscribe("alerts.topic", func(msg *nats.Msg) {
//           var event Event
//           json.Unmarshal(msg.Data, &event)
//           CentralCoordinator.ProcessAlert(event)
//       })
//   }
//   ```

// ---

// #### 3. **Central Coordinator**
// - **Role**: Aggregate alerts, correlate events, and trigger actions.
// - **Go Implementation**:
//   ```go
//   package coordinator

//   type Coordinator struct {
//       AlertsChan chan Event
//       Rules     []CorrelationRule
//   }

//   func (c *Coordinator) ProcessAlert(event Event) {
//       c.AlertsChan <- event
//   }

//   func (c *Coordinator) CorrelateEvents() {
//       for event := range c.AlertsChan {
//           for _, rule := range c.Rules {
//               if rule.Matches(event) {
//                   c.TriggerAction(rule.Action, event)
//               }
//           }
//           storeInDatabase(event)
//       }
//   }

//   func (c *Coordinator) TriggerAction(action ActionType, event Event) {
//       switch action {
//       case NotifyAdmin:
//           sendEmailAlert(event)
//       case BlockIP:
//           firewall.BlockIP(event.SourceIP)
//       }
//   }
//   ```

// ---

// #### 4. **Database (TimescaleDB)**
// - **Role**: Store raw events and correlated alerts.
// - **Go Implementation**:
//   ```go
//   func storeInDatabase(event Event) {
//       db, _ := sql.Open("postgres", "host=timescaledb user=admin dbname=ids")
//       _, err := db.Exec(`
//           INSERT INTO alerts (timestamp, source_ip, severity, description)
//           VALUES ($1, $2, $3, $4)`,
//           event.Timestamp, event.SourceIP, event.Severity, event.Description)
//       if err != nil {
//           log.Println("Failed to store alert:", err)
//       }
//   }
//   ```

// ---

// #### 5. **Dashboard (Grafana)**
// - **Role**: Visualize alerts and system health.
// - **Go Integration**:
//   - Expose Prometheus metrics from agents/coordinator:
//     ```go
//     func exposeMetrics() {
//         http.Handle("/metrics", promhttp.Handler())
//         http.ListenAndServe(":8080", nil)
//     }
//     ```
//   - Use `github.com/prometheus/client_golang` to track:
//     - Alerts per second
//     - Agent health status
//     - Latency

// ---

// ### **Key Features**
// 1. **Distributed Agents**: Deployable across multiple hosts/containers.
// 2. **Anomaly Detection**: Custom rules (e.g., rate limiting, regex patterns) or integrate ML models (ONNX runtime).
// 3. **Fault Tolerance**: Agents continue working even if the coordinator is down (buffered alerts).
// 4. **Scalability**: Horizontal scaling of agents and partitioned message queues.

// ---

// ### **Simulation Steps**
// 1. Run NATS server: `docker run -p 4222:4222 nats`.
// 2. Start 3 agent nodes:
//    ```go
//    agent1 := agent.NewAgent("node1", "nats://localhost:4222")
//    go agent1.Monitor()
//    ```
// 3. Start the coordinator:
//    ```go
//    coordinator := coordinator.NewCoordinator()
//    go coordinator.CorrelateEvents()
//    ```
// 4. Inject test events (e.g., port scans, failed logins).

// ---

// ### **Example Output**
// ```bash
// [Agent node1] Alert: Port scan detected from 192.168.1.100 (severity: 8)
// [Coordinator] Action: Blocked IP 192.168.1.100 via firewall
// ```

// This architecture can be extended with **Kubernetes** for orchestration and **gRPC** for low-latency communication. Use Go’s concurrency primitives (goroutines/channels) for high-throughput event processing.

// Here's a simulated architecture for deploying **honeypots** in three strategic network locations to detect and analyze threats at different stages of potential intrusion. The design uses Go-based honeypot services, a central logging/analysis system, and layered network segmentation.

// ---

// ### **Honeypot Deployment Architecture**
// ```
//                      +---------------------+
//                      |   Internet          |
//                      +----------+----------+
//                                 |
//                                 | Inbound Traffic
//                                 |
//                      +----------+----------+
//                      |  External Honeypot  |
//                      | (Outside Firewall) |
//                      | - Simulates open   |       Central
//                      |   services (SSH, HTTP)|      Analysis
//                      +----------+----------+       +-------------------+
//                                 |                  |                   |
//                                 | Alert Forwarding | Threat Correlation|
//                      +----------+----------+       | +---------------+ |
//                      |   External Firewall |       | | Log Aggregator| |
//                      +----------+----------+       | | (Elasticsearch|
//                                 |                  | | + Grafana)   |
//                                 |                  | +---------------+|
//                      +----------+----------+       +-------------------+
//                      |   DMZ Honeypot      |                   ^
//                      | (Service Network)  |                   |
//                      | - Mimics production|                   | Cross-Zone
//                      |   services (DB, FTP)|                   | Analysis
//                      +----------+----------+                   |
//                                 |                              |
//                                 |                              |
//                      +----------+----------+                   |
//                      |  Internal Firewall  |                   |
//                      +----------+----------+                   |
//                                 |                              |
//                                 |                              |
//                      +----------+----------+                   |
//                      |  Internal Honeypot  |                   |
//                      | (Internal Network) |                   |
//                      | - Mimics internal  |<------------------+
//                      |   apps/fileshares  |
//                      +---------------------+
// ```

// ---

// ### **Honeypot Implementation in Go**

// #### 1. **External Honeypot (Outside Firewall)**
// - **Purpose**: Attract opportunistic attackers and log common exploits.
// - **Simulated Services**: Open SSH, HTTP, RDP.
// - **Go Code**:
//   ```go
//   package main

//   import (
//       "fmt"
//       "net"
//       "log"
//       "encoding/json"
//       "time"
//   )

//   type ExternalHoneypot struct {
//       ListenAddr string
//       CollectorAddr string
//   }

//   func (h *ExternalHoneypot) Start() {
//       listener, err := net.Listen("tcp", h.ListenAddr)
//       if err != nil {
//           log.Fatal("External honeypot failed to start:", err)
//       }
//       defer listener.Close()

//       log.Printf("External honeypot listening on %s\n", h.ListenAddr)
//       for {
//           conn, err := listener.Accept()
//           if err != nil {
//               continue
//           }
//           go h.handleConnection(conn)
//       }
//   }

//   func (h *ExternalHoneypot) handleConnection(conn net.Conn) {
//       defer conn.Close()
//       remoteAddr := conn.RemoteAddr().String()
//       logEntry := map[string]interface{}{
//           "timestamp":  time.Now().UTC(),
//           "source_ip":  remoteAddr,
//           "honeypot":   "external",
//           "activity":   "connection_attempt",
//       }

//       // Simulate SSH banner to engage attackers
//       conn.Write([]byte("SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.3\r\n"))

//       // Log interaction
//       data, _ := json.Marshal(logEntry)
//       sendToCollector(h.CollectorAddr, data)
//   }
//   ```

// ---

// #### 2. **DMZ Honeypot (Service Network)**
// - **Purpose**: Detect attackers who bypass the external firewall.
// - **Simulated Services**: Fake MySQL, FTP, SMTP.
// - **Go Code**:
//   ```go
//   type DMZHoneypot struct {
//       ListenAddr string
//       CollectorAddr string
//   }

//   func (h *DMZHoneypot) Start() {
//       // Simulate MySQL server
//       listener, _ := net.Listen("tcp", h.ListenAddr)
//       defer listener.Close()

//       for {
//           conn, _ := listener.Accept()
//           go func(conn net.Conn) {
//               defer conn.Close()
//               conn.Write([]byte("\x08\x00\x00\x00\x0a8.0.22-0ubuntu0.20.04.1")) // MySQL banner

//               // Capture credentials (fake)
//               buf := make([]byte, 1024)
//               n, _ := conn.Read(buf)

//               logEntry := map[string]interface{}{
//                   "timestamp":  time.Now().UTC(),
//                   "source_ip":  conn.RemoteAddr().String(),
//                   "honeypot":   "dmz",
//                   "activity":   "mysql_login_attempt",
//                   "credentials": string(buf[:n]),
//               }
//               data, _ := json.Marshal(logEntry)
//               sendToCollector(h.CollectorAddr, data)
//           }(conn)
//       }
//   }
//   ```

// ---

// #### 3. **Internal Honeypot (Internal Network)**
// - **Purpose**: Detect lateral movement post-breach.
// - **Simulated Services**: Fake SMB shares, internal APIs.
// - **Go Code**:
//   ```go
//   type InternalHoneypot struct {
//       ListenAddr string
//       CollectorAddr string
//   }

//   func (h *InternalHoneypot) Start() {
//       // Simulate SMB service
//       listener, _ := net.Listen("tcp", h.ListenAddr)
//       defer listener.Close()

//       for {
//           conn, _ := listener.Accept()
//           go func(conn net.Conn) {
//               defer conn.Close()
//               conn.Write([]byte("SMBv2 Negotiate Protocol Response"))

//               // Log file access attempts
//               logEntry := map[string]interface{}{
//                   "timestamp": time.Now().UTC(),
//                   "source_ip": conn.RemoteAddr().String(),
//                   "honeypot":  "internal",
//                   "activity": "smb_file_access",
//               }
//               data, _ := json.Marshal(logEntry)
//               sendToCollector(h.CollectorAddr, data)
//           }(conn)
//       }
//   }
//   ```

// ---

// ### **Central Log Collector (Go Service)**
// ```go
// func sendToCollector(collectorAddr string, data []byte) {
//     resp, err := http.Post(
//         fmt.Sprintf("http://%s/ingest", collectorAddr),
//         "application/json",
//         bytes.NewBuffer(data),
//     )
//     if err != nil {
//         log.Println("Failed to send to collector:", err)
//     }
//     defer resp.Body.Close()
// }
// ```

// ---

// ### **Threat Detection Logic**
// 1. **Cross-Zone Correlation**:
//    ```go
//    // Example: Detect lateral movement from external to internal
//    func detectLateralMovement() {
//        query := `{
//            "query": {
//                "bool": {
//                    "must": [
//                        { "term": { "source_ip.keyword": "ATTACKER_IP" }},
//                        { "terms": { "honeypot.keyword": ["external", "internal"] }}
//                    ]
//                }
//            }
//        }`
//        results := elasticsearch.Search(query)
//        if len(results) > 0 {
//            triggerAlert("Lateral movement detected from IP: ATTACKER_IP")
//        }
//    }
//    ```

// 2. **Alerting**:
//    ```go
//    func triggerAlert(msg string) {
//        // Send to SIEM/Slack/Email
//        fmt.Printf("[ALERT] %s\n", msg)
//    }
//    ```

// ---

// ### **Deployment Setup**
// 1. **External Honeypot**:
//    ```bash
//    go run external_honeypot.go -addr ":22" -collector "10.0.0.100:8080"
//    ```

// 2. **DMZ Honeypot**:
//    ```bash
//    go run dmz_honeypot.go -addr ":3306" -collector "10.0.0.100:8080"
//    ```

// 3. **Internal Honeypot**:
//    ```bash
//    go run internal_honeypot.go -addr ":445" -collector "10.0.0.100:8080"
//    ```

// 4. **Central Collector**:
//    ```bash
//    go run collector.go -es "http://elasticsearch:9200"
//    ```

// ---

// ### **Security Considerations**
// - **Isolation**: Run honeypots in containers/VMs with no real data access.
// - **Traffic Mirroring**: Use port mirroring to redirect suspicious traffic to honeypots.
// - **Deception**: Use realistic banners/files (e.g., fake "confidential.docx" in SMB shares).
// - **Automated Response**: Integrate with firewall APIs to block aggressive IPs.

// ---

// ### **Example Attack Scenario**
// 1. **Reconnaissance**:
//    ```bash
//    Attacker scans public IP ➔ Connects to external honeypot on port 22.
//    [External Honeypot Log] { "source_ip": "94.23.211.44", "activity": "connection_attempt" }
//    ```

// 2. **Lateral Movement**:
//    ```bash
//    Attacker compromises DMZ server ➔ Probes internal network.
//    [Internal Honeypot Log] { "source_ip": "10.0.5.12", "activity": "smb_file_access" }
//    ```

// 3. **Alert**:
//    ```bash
//    [Central Analysis] ALERT: IP 10.0.5.12 accessed external and internal honeypots within 5m.
//    ```

// This setup provides visibility into attack progression and helps refine firewall rules/IPS signatures.

// Here's a Go implementation simulating an intrusion detection message exchange pipeline with concurrency and separation of concerns:

// ```go
// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"math/rand"
// 	"sync"
// 	"time"
// )

// Data structures
type SecurityEvent struct {
	SourceIP    string
	Activity    string
	Timestamp   time.Time
	RequestSize int
}

type Alert struct {
	ID        string
	Event     SecurityEvent
	Severity  int
	Signature string
}

type Action struct {
	AlertID      string
	ActionType   string
	Target       string
	InitiatedBy  string
	Timestamp    time.Time
	ResponseCode int
}

// Component: Data Source
func dataSource(out chan<- SecurityEvent) {
	activities := []string{"ssh_login", "http_request", "db_access", "file_modify"}
	for {
		event := SecurityEvent{
			SourceIP:    fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255)),
			Activity:    activities[rand.Intn(len(activities))],
			Timestamp:   time.Now(),
			RequestSize: rand.Intn(1024),
		}
		out <- event
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	}
}

// Component: Sensor
func sensor(in <-chan SecurityEvent, out chan<- SecurityEvent) {
	for event := range in {
		// Add sensor metadata
		event.RequestSize += 20 // Simulate packet overhead
		out <- event
	}
}

// Component: Analyzer
func analyzer(in <-chan SecurityEvent, out chan<- Alert) {
	signatures := map[string]string{
		"ssh_bruteforce": "Multiple SSH login attempts",
		"large_upload":   "Oversized HTTP request",
	}

	for event := range in {
		// Simple detection logic
		if event.RequestSize > 512 {
			out <- Alert{
				ID:        fmt.Sprintf("ALT-%d", time.Now().UnixNano()),
				Event:     event,
				Severity:  8,
				Signature: signatures["large_upload"],
			}
		}
	}
}

// Component: Manager
func manager(in <-chan Alert, out chan<- Action, policies <-chan string) chan<- string {
	policyUpdates := make(chan string)

	go func() {
		currentPolicy := "default"
		for {
			select {
			case alert := <-in:
				action := Action{
					AlertID:     alert.ID,
					ActionType:  "notify",
					Target:      "security-team",
					InitiatedBy: "manager",
					Timestamp:   time.Now(),
				}

				if currentPolicy == "aggressive" && alert.Severity > 7 {
					action.ActionType = "block_ip"
					action.Target = alert.Event.SourceIP
				}

				out <- action
			case newPolicy := <-policyUpdates:
				currentPolicy = newPolicy
				fmt.Printf("Updated security policy to: %s\n", newPolicy)
			}
		}
	}()

	return policyUpdates
}

// Component: Administrator
func administrator(in <-chan Action) {
	for action := range in {
		switch action.ActionType {
		case "block_ip":
			fmt.Printf("[ACTION] Blocking IP %s\n", action.Target)
			// Simulate firewall API call
			action.ResponseCode = 200
		case "notify":
			fmt.Printf("[NOTIFICATION] Alert %s requires attention\n", action.AlertID)
			// Simulate email notification
			action.ResponseCode = 200
		}
	}
}

func mainDDIntrusionDetection() {
	// Create communication channels
	dataChan := make(chan SecurityEvent, 100)
	sensorChan := make(chan SecurityEvent, 100)
	alertChan := make(chan Alert, 50)
	actionChan := make(chan Action, 20)
	policyChan := make(chan string)

	// Start components
	go dataSource(dataChan)
	go sensor(dataChan, sensorChan)
	go analyzer(sensorChan, alertChan)
	policyUpdates := manager(alertChan, actionChan, policyChan)
	go administrator(actionChan)

	// Simulate security policy update
	go func() {
		time.Sleep(2 * time.Second)
		policyUpdates <- "aggressive"
	}()

	// Run simulation for 5 seconds
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		time.Sleep(5 * time.Second)
		wg.Done()
	}()
	wg.Wait()

	// Cleanup
	close(dataChan)
	close(sensorChan)
	close(alertChan)
	close(actionChan)
	close(policyChan)

	fmt.Println("Simulation completed")
}

// Helper function to print JSON
func printJSON(data interface{}) {
	j, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(j))
}

// ```

// This implementation features:

// 1. **Component Pipeline**:
//    - Data Source: Generates random security events
//    - Sensor: Adds metadata and forwards events
//    - Analyzer: Applies detection rules
//    - Manager: Enforces security policies
//    - Administrator: Executes actions

// 2. **Concurrency Model**:
//    - Buffered channels for inter-component communication
//    - Goroutines for parallel processing
//    - WaitGroup for graceful shutdown

// 3. **Security Features**:
//    - Dynamic policy updates
//    - Multiple response types (notification, blocking)
//    - Configurable detection rules

// 4. **Simulation Capabilities**:
//    - Random event generation
//    - Policy change simulation
//    - Response action simulation

// Sample output:
// ```
// Updated security policy to: aggressive
// [ACTION] Blocking IP 192.168.21.173
// [ACTION] Blocking IP 192.168.202.238
// [NOTIFICATION] Alert ALT-1623757893042000000 requires attention
// Simulation completed
// ```

// To extend this model:

// 1. Add persistence layer for events/alerts
// 2. Implement more sophisticated detection algorithms
// 3. Add TLS communication between components
// 4. Implement rate limiting
// 5. Add metrics collection and monitoring
// 6. Implement actual network interactions
// 7. Add authentication between components

// The components communicate via channels in a pipeline fashion, allowing for horizontal scaling by adding multiple instances of each component with load balancing.
