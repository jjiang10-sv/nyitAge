// package main

// import (
// 	"bytes"
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"fmt"
// 	"log"
// 	"math"
// 	"net"
// 	"os"
// 	"regexp"
// 	"sync"
// 	"time"
// )

// // ================== Core IPS Engine ==================
// type IntrusionPreventionSystem struct {
// 	config          *IPSConfig
// 	threatDetector  *ThreatDetector
// 	responseEngine  *ResponseEngine
// 	trafficAnalyzer *TrafficAnalyzer
// 	behaviorMonitor *BehaviorMonitor
// 	blockList       *BlockList
// 	rateLimit       *RateLimiter
// 	alertChannel    chan SecurityAlert
// 	running         bool
// 	mu              sync.RWMutex
// }

// type IPSConfig struct {
// 	ListenPort      int
// 	ProtectedHosts  []string
// 	BlockDuration   time.Duration
// 	MaxConnections  int
// 	RateLimitWindow time.Duration
// 	AlertThreshold  int
// 	AutoBlock       bool
// 	LogFile         string
// 	WhitelistedIPs  []string
// }

// // ================== Threat Detection Engine ==================
// type ThreatDetector struct {
// 	signatures      []ThreatSignature
// 	anomalyDetector *AnomalyDetector
// 	protocolFilters map[string]*ProtocolFilter
// 	patternMatcher  *PatternMatcher
// 	geoFilter       *GeoIPFilter
// }

// type ThreatSignature struct {
// 	ID          string
// 	Name        string
// 	Pattern     string
// 	Protocol    string
// 	Severity    int
// 	Action      string
// 	Description string
// 	Regex       *regexp.Regexp
// }

// type SecurityAlert struct {
// 	ID          string
// 	Timestamp   time.Time
// 	SourceIP    string
// 	DestIP      string
// 	Protocol    string
// 	ThreatType  string
// 	Severity    int
// 	Description string
// 	Evidence    string
// 	Action      string
// 	Blocked     bool
// }

// // ================== Response Engine ==================
// type ResponseEngine struct {
// 	actions    map[string]ResponseAction
// 	blockList  *BlockList
// 	notifier   *AlertNotifier
// 	quarantine *QuarantineManager
// 	honeypot   *HoneypotManager
// }

// type ResponseAction struct {
// 	Type     string
// 	Duration time.Duration
// 	Severity int
// 	Handler  func(alert SecurityAlert) error
// }

// // ================== Traffic Analysis ==================
// type TrafficAnalyzer struct {
// 	packetInspector  *PacketInspector
// 	flowAnalyzer     *FlowAnalyzer
// 	protocolAnalyzer *ProtocolAnalyzer
// 	dpiEngine        *DPIEngine
// }

// type PacketInspector struct {
// 	rules           []InspectionRule
// 	payloadAnalyzer *PayloadAnalyzer
// }

// type InspectionRule struct {
// 	Name     string
// 	Protocol string
// 	Pattern  []byte
// 	Offset   int
// 	Action   string
// 	Severity int
// }

// // ================== Behavior Monitoring ==================
// type BehaviorMonitor struct {
// 	hostProfiles map[string]*HostProfile
// 	anomalyRules []AnomalyRule
// 	baselineData *BaselineData
// 	learningMode bool
// 	mu           sync.RWMutex
// }

// type HostProfile struct {
// 	IP              string
// 	NormalBehavior  BehaviorPattern
// 	CurrentBehavior BehaviorPattern
// 	ThreatScore     float64
// 	LastUpdate      time.Time
// }

// type BehaviorPattern struct {
// 	ConnectionRate   float64
// 	PortScanScore    float64
// 	DataTransferRate float64
// 	ProtocolMix      map[string]float64
// 	TimePatterns     map[string]float64
// }

// // ================== Rate Limiting & Blocking ==================
// type RateLimiter struct {
// 	limits map[string]*RateLimit
// 	window time.Duration
// 	mu     sync.RWMutex
// }

// type RateLimit struct {
// 	Count    int
// 	Window   time.Time
// 	MaxCount int
// 	Blocked  bool
// }

// type BlockList struct {
// 	blockedIPs   map[string]*BlockEntry
// 	blockedPorts map[int]*BlockEntry
// 	whiteList    map[string]bool
// 	mu           sync.RWMutex
// }

// type BlockEntry struct {
// 	IP        string
// 	Reason    string
// 	Timestamp time.Time
// 	Duration  time.Duration
// 	Severity  int
// }

// // ================== IPS Implementation ==================
// func NewIntrusionPreventionSystem(config *IPSConfig) *IntrusionPreventionSystem {
// 	ips := &IntrusionPreventionSystem{
// 		config:          config,
// 		threatDetector:  NewThreatDetector(),
// 		responseEngine:  NewResponseEngine(),
// 		trafficAnalyzer: NewTrafficAnalyzer(),
// 		behaviorMonitor: NewBehaviorMonitor(),
// 		blockList:       NewBlockList(config.WhitelistedIPs),
// 		rateLimit:       NewRateLimiter(config.RateLimitWindow),
// 		alertChannel:    make(chan SecurityAlert, 1000),
// 		running:         false,
// 	}

// 	ips.loadThreatSignatures()
// 	ips.setupResponseActions()
// 	return ips
// }

// func (ips *IntrusionPreventionSystem) Start() error {
// 	ips.mu.Lock()
// 	defer ips.mu.Unlock()

// 	if ips.running {
// 		return fmt.Errorf("IPS is already running")
// 	}

// 	log.Println("🛡️  Starting Intrusion Prevention System")

// 	// Start traffic monitoring
// 	go ips.startTrafficMonitoring()

// 	// Start behavior monitoring
// 	go ips.startBehaviorMonitoring()

// 	// Start alert processing
// 	go ips.processAlerts()

// 	// Start cleanup routine
// 	go ips.startCleanupRoutine()

// 	ips.running = true
// 	log.Printf("✅ IPS started on port %d protecting hosts: %v",
// 		ips.config.ListenPort, ips.config.ProtectedHosts)

// 	return nil
// }

// func (ips *IntrusionPreventionSystem) startTrafficMonitoring() {
// 	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ips.config.ListenPort))
// 	if err != nil {
// 		log.Fatalf("Failed to start traffic monitor: %v", err)
// 	}
// 	defer listener.Close()

// 	for ips.running {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			continue
// 		}
// 		go ips.handleConnection(conn)
// 	}
// }

// func (ips *IntrusionPreventionSystem) handleConnection(conn net.Conn) {
// 	defer conn.Close()

// 	clientIP := getClientIP(conn)

// 	// Check if IP is blocked
// 	if ips.blockList.IsBlocked(clientIP) {
// 		log.Printf("🚫 Blocked connection from %s", clientIP)
// 		ips.generateAlert("BLOCKED_CONNECTION", clientIP, "", "Connection blocked", 5)
// 		return
// 	}

// 	// Rate limiting check
// 	if ips.rateLimit.IsRateLimited(clientIP) {
// 		log.Printf("⚠️  Rate limited connection from %s", clientIP)
// 		ips.generateAlert("RATE_LIMITED", clientIP, "", "Rate limit exceeded", 4)
// 		ips.blockList.AddIP(clientIP, "Rate limit exceeded", time.Minute*5, 4)
// 		return
// 	}

// 	// Deep packet inspection
// 	buffer := make([]byte, 4096)
// 	n, err := conn.Read(buffer)
// 	if err != nil {
// 		return
// 	}

// 	// Analyze traffic
// 	threats := ips.analyzeTraffic(buffer[:n], clientIP, getServerIP(conn))

// 	// Process detected threats
// 	for _, threat := range threats {
// 		ips.alertChannel <- threat

// 		if threat.Severity >= 7 && ips.config.AutoBlock {
// 			duration := ips.calculateBlockDuration(threat.Severity)
// 			ips.blockList.AddIP(clientIP, threat.Description, duration, threat.Severity)
// 			log.Printf("🔒 Auto-blocked %s for %v due to: %s",
// 				clientIP, duration, threat.Description)
// 			return
// 		}
// 	}

// 	// Update behavior profile
// 	ips.behaviorMonitor.UpdateHostProfile(clientIP, buffer[:n])

// 	// Forward to protected host if clean
// 	if len(threats) == 0 {
// 		ips.forwardToProtectedHost(conn, buffer[:n])
// 	}
// }

// func (ips *IntrusionPreventionSystem) analyzeTraffic(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	var alerts []SecurityAlert

// 	// Signature-based detection
// 	sigAlerts := ips.threatDetector.DetectSignatures(data, srcIP, dstIP)
// 	alerts = append(alerts, sigAlerts...)

// 	// Protocol analysis
// 	protoAlerts := ips.trafficAnalyzer.AnalyzeProtocols(data, srcIP, dstIP)
// 	alerts = append(alerts, protoAlerts...)

// 	// Payload inspection
// 	payloadAlerts := ips.trafficAnalyzer.InspectPayload(data, srcIP, dstIP)
// 	alerts = append(alerts, payloadAlerts...)

// 	// Anomaly detection
// 	anomalyAlerts := ips.detectAnomalies(data, srcIP, dstIP)
// 	alerts = append(alerts, anomalyAlerts...)

// 	return alerts
// }

// // ================== Threat Detection Implementation ==================
// func NewThreatDetector() *ThreatDetector {
// 	return &ThreatDetector{
// 		signatures:      loadThreatSignatures(),
// 		anomalyDetector: NewAnomalyDetector(),
// 		protocolFilters: initializeProtocolFilters(),
// 		patternMatcher:  NewPatternMatcher(),
// 		geoFilter:       NewGeoIPFilter(),
// 	}
// }

// func (td *ThreatDetector) DetectSignatures(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	var alerts []SecurityAlert
// 	content := string(data)

// 	for _, sig := range td.signatures {
// 		if sig.Regex.MatchString(content) {
// 			alert := SecurityAlert{
// 				ID:          generateAlertID(),
// 				Timestamp:   time.Now(),
// 				SourceIP:    srcIP,
// 				DestIP:      dstIP,
// 				Protocol:    sig.Protocol,
// 				ThreatType:  sig.Name,
// 				Severity:    sig.Severity,
// 				Description: sig.Description,
// 				Evidence:    sig.Pattern,
// 				Action:      sig.Action,
// 				Blocked:     false,
// 			}
// 			alerts = append(alerts, alert)
// 		}
// 	}

// 	return alerts
// }

// func loadThreatSignatures() []ThreatSignature {
// 	signatures := []ThreatSignature{
// 		// SQL Injection patterns
// 		{
// 			ID:          "SQL_001",
// 			Name:        "SQL_Injection",
// 			Pattern:     `(?i)(union\s+select|or\s+1\s*=\s*1|drop\s+table|exec\s*\(|script\s*>)`,
// 			Protocol:    "HTTP",
// 			Severity:    9,
// 			Action:      "BLOCK",
// 			Description: "SQL injection attack detected",
// 		},
// 		// XSS patterns
// 		{
// 			ID:          "XSS_001",
// 			Name:        "Cross_Site_Scripting",
// 			Pattern:     `(?i)(<script[^>]*>|javascript:|on\w+\s*=|eval\s*\(|alert\s*\()`,
// 			Protocol:    "HTTP",
// 			Severity:    8,
// 			Action:      "BLOCK",
// 			Description: "Cross-site scripting attack detected",
// 		},
// 		// Directory traversal
// 		{
// 			ID:          "DT_001",
// 			Name:        "Directory_Traversal",
// 			Pattern:     `(\.\./){2,}|\.\.\\|%2e%2e%2f|%252e%252e%252f`,
// 			Protocol:    "HTTP",
// 			Severity:    7,
// 			Action:      "BLOCK",
// 			Description: "Directory traversal attack detected",
// 		},
// 		// Command injection
// 		{
// 			ID:          "CI_001",
// 			Name:        "Command_Injection",
// 			Pattern:     `(?i)(;|\|)\s*(cat|ls|dir|type|echo|ping|wget|curl|nc|netcat)\s`,
// 			Protocol:    "HTTP",
// 			Severity:    9,
// 			Action:      "BLOCK",
// 			Description: "Command injection attack detected",
// 		},
// 		// Shellcode patterns
// 		{
// 			ID:          "SC_001",
// 			Name:        "Shellcode",
// 			Pattern:     `(\x90{4,}|\x31\xc0|\xeb\xfe|AAAA)`,
// 			Protocol:    "TCP",
// 			Severity:    10,
// 			Action:      "BLOCK",
// 			Description: "Shellcode detected in traffic",
// 		},
// 		// Port scan detection
// 		{
// 			ID:          "PS_001",
// 			Name:        "Port_Scan",
// 			Pattern:     `^.{0,10}$`, // Short packets often used in scanning
// 			Protocol:    "TCP",
// 			Severity:    5,
// 			Action:      "MONITOR",
// 			Description: "Potential port scan detected",
// 		},
// 	}

// 	// Compile regex patterns
// 	for i := range signatures {
// 		signatures[i].Regex = regexp.MustCompile(signatures[i].Pattern)
// 	}

// 	return signatures
// }

// // ================== Response Engine Implementation ==================
// func NewResponseEngine() *ResponseEngine {
// 	return &ResponseEngine{
// 		actions:    make(map[string]ResponseAction),
// 		blockList:  NewBlockList([]string{}),
// 		notifier:   NewAlertNotifier(),
// 		quarantine: NewQuarantineManager(),
// 		honeypot:   NewHoneypotManager(),
// 	}
// }

// func (ips *IntrusionPreventionSystem) setupResponseActions() {
// 	actions := map[string]ResponseAction{
// 		"BLOCK": {
// 			Type:     "BLOCK",
// 			Duration: time.Hour,
// 			Severity: 7,
// 			Handler:  ips.handleBlockAction,
// 		},
// 		"MONITOR": {
// 			Type:     "MONITOR",
// 			Duration: time.Minute * 30,
// 			Severity: 5,
// 			Handler:  ips.handleMonitorAction,
// 		},
// 		"QUARANTINE": {
// 			Type:     "QUARANTINE",
// 			Duration: time.Hour * 24,
// 			Severity: 9,
// 			Handler:  ips.handleQuarantineAction,
// 		},
// 		"REDIRECT": {
// 			Type:     "REDIRECT",
// 			Duration: time.Hour,
// 			Severity: 6,
// 			Handler:  ips.handleRedirectAction,
// 		},
// 	}

// 	ips.responseEngine.actions = actions
// }

// func (ips *IntrusionPreventionSystem) handleBlockAction(alert SecurityAlert) error {
// 	duration := ips.calculateBlockDuration(alert.Severity)
// 	ips.blockList.AddIP(alert.SourceIP, alert.Description, duration, alert.Severity)

// 	log.Printf("🔒 BLOCKED: %s for %v - %s",
// 		alert.SourceIP, duration, alert.Description)

// 	return nil
// }

// func (ips *IntrusionPreventionSystem) handleMonitorAction(alert SecurityAlert) error {
// 	log.Printf("👁️  MONITORING: %s - %s", alert.SourceIP, alert.Description)
// 	ips.behaviorMonitor.AddSuspiciousActivity(alert.SourceIP, alert)
// 	return nil
// }

// func (ips *IntrusionPreventionSystem) handleQuarantineAction(alert SecurityAlert) error {
// 	log.Printf("🏥 QUARANTINED: %s - %s", alert.SourceIP, alert.Description)
// 	return ips.responseEngine.quarantine.QuarantineHost(alert.SourceIP, alert.Description)
// }

// func (ips *IntrusionPreventionSystem) handleRedirectAction(alert SecurityAlert) error {
// 	log.Printf("↪️  REDIRECTED: %s to honeypot - %s", alert.SourceIP, alert.Description)
// 	return ips.responseEngine.honeypot.RedirectToHoneypot(alert.SourceIP)
// }

// // ================== Traffic Analysis Implementation ==================
// func NewTrafficAnalyzer() *TrafficAnalyzer {
// 	return &TrafficAnalyzer{
// 		packetInspector:  NewPacketInspector(),
// 		flowAnalyzer:     NewFlowAnalyzer(),
// 		protocolAnalyzer: NewProtocolAnalyzer(),
// 		dpiEngine:        NewDPIEngine(),
// 	}
// }

// func (ta *TrafficAnalyzer) AnalyzeProtocols(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	var alerts []SecurityAlert

// 	// HTTP analysis
// 	if isHTTP(data) {
// 		httpAlerts := ta.analyzeHTTP(data, srcIP, dstIP)
// 		alerts = append(alerts, httpAlerts...)
// 	}

// 	// TCP analysis
// 	tcpAlerts := ta.analyzeTCP(data, srcIP, dstIP)
// 	alerts = append(alerts, tcpAlerts...)

// 	// DNS analysis
// 	if isDNS(data) {
// 		dnsAlerts := ta.analyzeDNS(data, srcIP, dstIP)
// 		alerts = append(alerts, dnsAlerts...)
// 	}

// 	return alerts
// }

// func (ta *TrafficAnalyzer) InspectPayload(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	var alerts []SecurityAlert

// 	// Check for malicious patterns
// 	if isSuspiciousPayload(data) {
// 		alert := SecurityAlert{
// 			ID:          generateAlertID(),
// 			Timestamp:   time.Now(),
// 			SourceIP:    srcIP,
// 			DestIP:      dstIP,
// 			Protocol:    "TCP",
// 			ThreatType:  "MALICIOUS_PAYLOAD",
// 			Severity:    8,
// 			Description: "Suspicious payload detected",
// 			Evidence:    hex.EncodeToString(data[:min(32, len(data))]),
// 			Action:      "BLOCK",
// 			Blocked:     false,
// 		}
// 		alerts = append(alerts, alert)
// 	}

// 	// Entropy analysis
// 	entropy := calculateEntropy(data)
// 	if entropy > 7.5 {
// 		alert := SecurityAlert{
// 			ID:          generateAlertID(),
// 			Timestamp:   time.Now(),
// 			SourceIP:    srcIP,
// 			DestIP:      dstIP,
// 			Protocol:    "TCP",
// 			ThreatType:  "HIGH_ENTROPY",
// 			Severity:    6,
// 			Description: fmt.Sprintf("High entropy payload detected (%.2f)", entropy),
// 			Evidence:    fmt.Sprintf("Entropy: %.2f", entropy),
// 			Action:      "MONITOR",
// 			Blocked:     false,
// 		}
// 		alerts = append(alerts, alert)
// 	}

// 	return alerts
// }

// // ================== Behavior Monitoring Implementation ==================
// func NewBehaviorMonitor() *BehaviorMonitor {
// 	return &BehaviorMonitor{
// 		hostProfiles: make(map[string]*HostProfile),
// 		anomalyRules: loadAnomalyRules(),
// 		baselineData: NewBaselineData(),
// 		learningMode: true,
// 		mu:           sync.RWMutex{},
// 	}
// }

// func (bm *BehaviorMonitor) UpdateHostProfile(ip string, data []byte) {
// 	bm.mu.Lock()
// 	defer bm.mu.Unlock()

// 	profile, exists := bm.hostProfiles[ip]
// 	if !exists {
// 		profile = &HostProfile{
// 			IP:              ip,
// 			NormalBehavior:  BehaviorPattern{ProtocolMix: make(map[string]float64)},
// 			CurrentBehavior: BehaviorPattern{ProtocolMix: make(map[string]float64)},
// 			ThreatScore:     0.0,
// 			LastUpdate:      time.Now(),
// 		}
// 		bm.hostProfiles[ip] = profile
// 	}

// 	// Update current behavior
// 	bm.updateBehaviorPattern(&profile.CurrentBehavior, data)
// 	profile.LastUpdate = time.Now()

// 	// Calculate threat score
// 	profile.ThreatScore = bm.calculateThreatScore(profile)
// }

// func (bm *BehaviorMonitor) AddSuspiciousActivity(ip string, alert SecurityAlert) {
// 	bm.mu.Lock()
// 	defer bm.mu.Unlock()

// 	profile, exists := bm.hostProfiles[ip]
// 	if !exists {
// 		return
// 	}

// 	// Increase threat score based on alert severity
// 	profile.ThreatScore += float64(alert.Severity) * 0.1
// }

// // ================== Rate Limiting Implementation ==================
// func NewRateLimiter(window time.Duration) *RateLimiter {
// 	return &RateLimiter{
// 		limits: make(map[string]*RateLimit),
// 		window: window,
// 		mu:     sync.RWMutex{},
// 	}
// }

// func (rl *RateLimiter) IsRateLimited(ip string) bool {
// 	rl.mu.Lock()
// 	defer rl.mu.Unlock()

// 	limit, exists := rl.limits[ip]
// 	if !exists {
// 		rl.limits[ip] = &RateLimit{
// 			Count:    1,
// 			Window:   time.Now(),
// 			MaxCount: 100, // Default limit
// 			Blocked:  false,
// 		}
// 		return false
// 	}

// 	// Reset window if expired
// 	if time.Since(limit.Window) > rl.window {
// 		limit.Count = 1
// 		limit.Window = time.Now()
// 		limit.Blocked = false
// 		return false
// 	}

// 	limit.Count++
// 	if limit.Count > limit.MaxCount {
// 		limit.Blocked = true
// 		return true
// 	}

// 	return false
// }

// // ================== Block List Implementation ==================
// func NewBlockList(whitelist []string) *BlockList {
// 	whiteMap := make(map[string]bool)
// 	for _, ip := range whitelist {
// 		whiteMap[ip] = true
// 	}

// 	return &BlockList{
// 		blockedIPs:   make(map[string]*BlockEntry),
// 		blockedPorts: make(map[int]*BlockEntry),
// 		whiteList:    whiteMap,
// 		mu:           sync.RWMutex{},
// 	}
// }

// func (bl *BlockList) IsBlocked(ip string) bool {
// 	bl.mu.RLock()
// 	defer bl.mu.RUnlock()

// 	// Check whitelist first
// 	if bl.whiteList[ip] {
// 		return false
// 	}

// 	entry, exists := bl.blockedIPs[ip]
// 	if !exists {
// 		return false
// 	}

// 	// Check if block has expired
// 	if time.Since(entry.Timestamp) > entry.Duration {
// 		delete(bl.blockedIPs, ip)
// 		return false
// 	}

// 	return true
// }

// func (bl *BlockList) AddIP(ip, reason string, duration time.Duration, severity int) {
// 	bl.mu.Lock()
// 	defer bl.mu.Unlock()

// 	// Don't block whitelisted IPs
// 	if bl.whiteList[ip] {
// 		return
// 	}

// 	bl.blockedIPs[ip] = &BlockEntry{
// 		IP:        ip,
// 		Reason:    reason,
// 		Timestamp: time.Now(),
// 		Duration:  duration,
// 		Severity:  severity,
// 	}
// }

// // ================== Alert Processing ==================
// func (ips *IntrusionPreventionSystem) processAlerts() {
// 	for alert := range ips.alertChannel {
// 		// Log alert
// 		ips.logAlert(alert)

// 		// Execute response action
// 		if action, exists := ips.responseEngine.actions[alert.Action]; exists {
// 			if err := action.Handler(alert); err != nil {
// 				log.Printf("Error executing action %s: %v", alert.Action, err)
// 			}
// 		}

// 		// Send notifications for high severity alerts
// 		if alert.Severity >= 8 {
// 			ips.responseEngine.notifier.SendAlert(alert)
// 		}
// 	}
// }

// // ================== Helper Functions ==================
// func (ips *IntrusionPreventionSystem) generateAlert(threatType, srcIP, dstIP, description string, severity int) {
// 	alert := SecurityAlert{
// 		ID:          generateAlertID(),
// 		Timestamp:   time.Now(),
// 		SourceIP:    srcIP,
// 		DestIP:      dstIP,
// 		ThreatType:  threatType,
// 		Severity:    severity,
// 		Description: description,
// 		Action:      ips.determineAction(severity),
// 		Blocked:     false,
// 	}

// 	select {
// 	case ips.alertChannel <- alert:
// 	default:
// 		log.Println("Alert channel full, dropping alert")
// 	}
// }

// func (ips *IntrusionPreventionSystem) determineAction(severity int) string {
// 	if severity >= 9 {
// 		return "QUARANTINE"
// 	} else if severity >= 7 {
// 		return "BLOCK"
// 	} else if severity >= 5 {
// 		return "MONITOR"
// 	}
// 	return "LOG"
// }

// func (ips *IntrusionPreventionSystem) calculateBlockDuration(severity int) time.Duration {
// 	switch {
// 	case severity >= 9:
// 		return time.Hour * 24
// 	case severity >= 7:
// 		return time.Hour * 2
// 	case severity >= 5:
// 		return time.Minute * 30
// 	default:
// 		return time.Minute * 5
// 	}
// }

// func (ips *IntrusionPreventionSystem) detectAnomalies(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	var alerts []SecurityAlert

// 	// Check for port scan patterns
// 	if len(data) < 10 && len(data) > 0 {
// 		alert := SecurityAlert{
// 			ID:          generateAlertID(),
// 			Timestamp:   time.Now(),
// 			SourceIP:    srcIP,
// 			DestIP:      dstIP,
// 			ThreatType:  "PORT_SCAN",
// 			Severity:    5,
// 			Description: "Potential port scan detected (short packet)",
// 			Action:      "MONITOR",
// 			Blocked:     false,
// 		}
// 		alerts = append(alerts, alert)
// 	}

// 	// Check for unusual packet sizes
// 	if len(data) > 8192 {
// 		alert := SecurityAlert{
// 			ID:          generateAlertID(),
// 			Timestamp:   time.Now(),
// 			SourceIP:    srcIP,
// 			DestIP:      dstIP,
// 			ThreatType:  "LARGE_PACKET",
// 			Severity:    4,
// 			Description: fmt.Sprintf("Unusually large packet (%d bytes)", len(data)),
// 			Action:      "MONITOR",
// 			Blocked:     false,
// 		}
// 		alerts = append(alerts, alert)
// 	}

// 	return alerts
// }

// func (ips *IntrusionPreventionSystem) forwardToProtectedHost(conn net.Conn, initialData []byte) {
// 	// Simple forwarding to first protected host for demo
// 	if len(ips.config.ProtectedHosts) == 0 {
// 		return
// 	}

// 	targetHost := ips.config.ProtectedHosts[0]
// 	target, err := net.Dial("tcp", targetHost)
// 	if err != nil {
// 		log.Printf("Failed to connect to protected host %s: %v", targetHost, err)
// 		return
// 	}
// 	defer target.Close()

// 	// Send initial data
// 	target.Write(initialData)

// 	// Relay traffic bidirectionally
// 	go func() {
// 		buf := make([]byte, 4096)
// 		for {
// 			n, err := conn.Read(buf)
// 			if err != nil {
// 				break
// 			}
// 			target.Write(buf[:n])
// 		}
// 	}()

// 	buf := make([]byte, 4096)
// 	for {
// 		n, err := target.Read(buf)
// 		if err != nil {
// 			break
// 		}
// 		conn.Write(buf[:n])
// 	}
// }

// func (ips *IntrusionPreventionSystem) startBehaviorMonitoring() {
// 	ticker := time.NewTicker(time.Minute * 5)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			ips.analyzeBehaviorProfiles()
// 		}
// 	}
// }

// func (ips *IntrusionPreventionSystem) analyzeBehaviorProfiles() {
// 	ips.behaviorMonitor.mu.RLock()
// 	defer ips.behaviorMonitor.mu.RUnlock()

// 	for ip, profile := range ips.behaviorMonitor.hostProfiles {
// 		if profile.ThreatScore > 0.7 {
// 			ips.generateAlert("HIGH_THREAT_SCORE", ip, "",
// 				fmt.Sprintf("Host threat score: %.2f", profile.ThreatScore), 7)
// 		}
// 	}
// }

// func (ips *IntrusionPreventionSystem) startCleanupRoutine() {
// 	ticker := time.NewTicker(time.Hour)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-ticker.C:
// 			ips.cleanupExpiredBlocks()
// 		}
// 	}
// }

// func (ips *IntrusionPreventionSystem) cleanupExpiredBlocks() {
// 	ips.blockList.mu.Lock()
// 	defer ips.blockList.mu.Unlock()

// 	for ip, entry := range ips.blockList.blockedIPs {
// 		if time.Since(entry.Timestamp) > entry.Duration {
// 			delete(ips.blockList.blockedIPs, ip)
// 			log.Printf("🔓 Unblocked %s (block expired)", ip)
// 		}
// 	}
// }

// func (ips *IntrusionPreventionSystem) logAlert(alert SecurityAlert) {
// 	logEntry := fmt.Sprintf("[%s] %s -> %s: %s (Severity: %d, Action: %s)\n",
// 		alert.Timestamp.Format(time.RFC3339),
// 		alert.SourceIP, alert.DestIP,
// 		alert.Description, alert.Severity, alert.Action)

// 	// Log to file if configured
// 	if ips.config.LogFile != "" {
// 		file, err := os.OpenFile(ips.config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 		if err == nil {
// 			file.WriteString(logEntry)
// 			file.Close()
// 		}
// 	}

// 	// Log to console
// 	fmt.Print(logEntry)
// }

// // ================== Utility Functions ==================
// func getClientIP(conn net.Conn) string {
// 	if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
// 		return addr.IP.String()
// 	}
// 	return "unknown"
// }

// func getServerIP(conn net.Conn) string {
// 	if addr, ok := conn.LocalAddr().(*net.TCPAddr); ok {
// 		return addr.IP.String()
// 	}
// 	return "unknown"
// }

// func generateAlertID() string {
// 	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
// 	return hex.EncodeToString(hash[:4])
// }

// func isHTTP(data []byte) bool {
// 	return bytes.HasPrefix(data, []byte("GET")) ||
// 		bytes.HasPrefix(data, []byte("POST")) ||
// 		bytes.HasPrefix(data, []byte("PUT")) ||
// 		bytes.HasPrefix(data, []byte("DELETE"))
// }

// func isDNS(data []byte) bool {
// 	return len(data) > 12 && data[2]&0x80 == 0 // DNS query
// }

// func isSuspiciousPayload(data []byte) bool {
// 	// Check for common shellcode patterns
// 	patterns := [][]byte{
// 		{0x90, 0x90, 0x90, 0x90}, // NOP sled
// 		{0x31, 0xc0},             // xor eax, eax
// 		{0xeb, 0xfe},             // jmp short $
// 	}

// 	for _, pattern := range patterns {
// 		if bytes.Contains(data, pattern) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func calculateEntropy(data []byte) float64 {
// 	if len(data) == 0 {
// 		return 0
// 	}

// 	freq := make(map[byte]int)
// 	for _, b := range data {
// 		freq[b]++
// 	}

// 	entropy := 0.0
// 	length := float64(len(data))

// 	for _, count := range freq {
// 		if count > 0 {
// 			p := float64(count) / length
// 			entropy -= p * math.Log2(p)
// 		}
// 	}

// 	return entropy
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// // ================== Stub implementations for demo ==================
// func NewAnomalyDetector() *AnomalyDetector     { return &AnomalyDetector{} }
// func NewPatternMatcher() *PatternMatcher       { return &PatternMatcher{} }
// func NewGeoIPFilter() *GeoIPFilter             { return &GeoIPFilter{} }
// func NewAlertNotifier() *AlertNotifier         { return &AlertNotifier{} }
// func NewQuarantineManager() *QuarantineManager { return &QuarantineManager{} }
// func NewHoneypotManager() *HoneypotManager     { return &HoneypotManager{} }
// func NewPacketInspector() *PacketInspector     { return &PacketInspector{} }
// func NewFlowAnalyzer() *FlowAnalyzer           { return &FlowAnalyzer{} }
// func NewProtocolAnalyzer() *ProtocolAnalyzer   { return &ProtocolAnalyzer{} }
// func NewDPIEngine() *DPIEngine                 { return &DPIEngine{} }
// func NewBaselineData() *BaselineData           { return &BaselineData{} }

// type AnomalyDetector struct{}
// type PatternMatcher struct{}
// type GeoIPFilter struct{}
// type AlertNotifier struct{}
// type QuarantineManager struct{}
// type HoneypotManager struct{}
// type FlowAnalyzer struct{}
// type ProtocolAnalyzer struct{}
// type DPIEngine struct{}
// type BaselineData struct{}
// type PayloadAnalyzer struct{}
// type AnomalyRule struct{}

// func initializeProtocolFilters() map[string]*ProtocolFilter { return make(map[string]*ProtocolFilter) }
// func loadAnomalyRules() []AnomalyRule                       { return []AnomalyRule{} }

// type ProtocolFilter struct{}

// func (an *AlertNotifier) SendAlert(alert SecurityAlert)                                 {}
// func (qm *QuarantineManager) QuarantineHost(ip, reason string) error                    { return nil }
// func (hm *HoneypotManager) RedirectToHoneypot(ip string) error                          { return nil }
// func (bm *BehaviorMonitor) updateBehaviorPattern(pattern *BehaviorPattern, data []byte) {}
// func (bm *BehaviorMonitor) calculateThreatScore(profile *HostProfile) float64           { return 0.0 }
// func (ta *TrafficAnalyzer) analyzeHTTP(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	return []SecurityAlert{}
// }
// func (ta *TrafficAnalyzer) analyzeTCP(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	return []SecurityAlert{}
// }
// func (ta *TrafficAnalyzer) analyzeDNS(data []byte, srcIP, dstIP string) []SecurityAlert {
// 	return []SecurityAlert{}
// }

// // ================== Main Demo Function ==================
// func main() {
// 	config := &IPSConfig{
// 		ListenPort:      8080,
// 		ProtectedHosts:  []string{"192.168.1.100:80", "192.168.1.101:80"},
// 		BlockDuration:   time.Hour,
// 		MaxConnections:  1000,
// 		RateLimitWindow: time.Minute,
// 		AlertThreshold:  5,
// 		AutoBlock:       true,
// 		LogFile:         "ips.log",
// 		WhitelistedIPs:  []string{"127.0.0.1", "192.168.1.1"},
// 	}

// 	ips := NewIntrusionPreventionSystem(config)

// 	fmt.Println("🛡️  Intrusion Prevention System - Protecting Victim Hosts")
// 	fmt.Println("=========================================================")

// 	if err := ips.Start(); err != nil {
// 		log.Fatalf("Failed to start IPS: %v", err)
// 	}

// 	// Demo threat simulation
// 	go simulateThreats()

// 	// Keep running
// 	select {}
// }

// func simulateThreats() {
// 	time.Sleep(5 * time.Second)

// 	threats := []string{
// 		"GET /admin' OR 1=1--",
// 		"<script>alert('xss')</script>",
// 		"../../../../etc/passwd",
// 		"; cat /etc/shadow",
// 		string([]byte{0x90, 0x90, 0x90, 0x90, 0x31, 0xc0}),
// 	}

// 	for i, threat := range threats {
// 		fmt.Printf("\n🧪 Simulating threat %d: %s\n", i+1, threat)

// 		conn, err := net.Dial("tcp", "localhost:8080")
// 		if err != nil {
// 			continue
// 		}

// 		conn.Write([]byte(threat))
// 		conn.Close()

// 		time.Sleep(2 * time.Second)
// 	}
// }
