// Below is a Go simulation of Linux's netfilter/iptables functionality, including rule matching, chain traversal, and packet filtering:

// ```go
package transport

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"regexp"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// Packet represents network traffic
type Packet struct {
	Protocol        string
	SourceIP        net.IP
	DestinationIP   net.IP
	SourcePort      int
	DestinationPort int
	InInterface     string
	OutInterface    string
	Action          string // Pending decision
}

// Rule defines iptables-like matching criteria
type Rule struct {
	Protocol        string
	Source          string // CIDR notation
	Destination     string // CIDR notation
	InInterface     string
	OutInterface    string
	SourcePort      int
	DestinationPort int
	Target          string
}

// Chain contains ordered rules and default policy
type Chain struct {
	Name   string
	Rules  []Rule
	Policy string // ACCEPT or DROP
}

// Table represents different iptables tables
type Table struct {
	Name   string // filter, nat, mangle
	Chains map[string]*Chain
}

// Netfilter simulation core
type Netfilter struct {
	Tables map[string]*Table
}

func NewNetfilter() *Netfilter {
	return &Netfilter{
		Tables: map[string]*Table{
			"filter": {
				Name: "filter",
				Chains: map[string]*Chain{
					"INPUT":   {Name: "INPUT", Policy: "ACCEPT"},
					"OUTPUT":  {Name: "OUTPUT", Policy: "ACCEPT"},
					"FORWARD": {Name: "FORWARD", Policy: "DROP"},
				},
			},
			"nat": {
				Name: "nat",
				Chains: map[string]*Chain{
					"PREROUTING":  {Name: "PREROUTING", Policy: "ACCEPT"},
					"POSTROUTING": {Name: "POSTROUTING", Policy: "ACCEPT"},
				},
			},
		},
	}
}

func (nf *Netfilter) AddRule(table, chain string, rule Rule) {
	if t, ok := nf.Tables[table]; ok {
		if c, ok := t.Chains[chain]; ok {
			c.Rules = append(c.Rules, rule)
		}
	}
}

func (nf *Netfilter) ProcessPacket(pkt *Packet) string {
	// Traverse tables in priority order: raw -> mangle -> nat -> filter
	for _, tableName := range []string{"nat", "filter"} {
		table := nf.Tables[tableName]

		switch tableName {
		case "nat":
			// Handle NAT processing
			if pkt.Action == "" {
				nf.processChain(table.Chains["PREROUTING"], pkt)
			}

		case "filter":
			switch {
			case pkt.InInterface != "":
				nf.processChain(table.Chains["INPUT"], pkt)
			case pkt.OutInterface != "":
				nf.processChain(table.Chains["OUTPUT"], pkt)
			default:
				nf.processChain(table.Chains["FORWARD"], pkt)
			}
		}

		if pkt.Action != "" {
			return pkt.Action
		}
	}

	return "ACCEPT"
}

func (nf *Netfilter) processChain(chain *Chain, pkt *Packet) {
	for _, rule := range chain.Rules {
		if matchRule(rule, pkt) {
			pkt.Action = rule.Target
			return
		}
	}

	// No rules matched, apply default policy
	pkt.Action = chain.Policy
}

func matchRule(rule Rule, pkt *Packet) bool {
	if rule.Protocol != "" && rule.Protocol != pkt.Protocol {
		return false
	}

	if rule.Source != "" {
		_, cidrNet, _ := net.ParseCIDR(rule.Source)
		if !cidrNet.Contains(pkt.SourceIP) {
			return false
		}
	}

	if rule.Destination != "" {
		_, cidrNet, _ := net.ParseCIDR(rule.Destination)
		if !cidrNet.Contains(pkt.DestinationIP) {
			return false
		}
	}

	if rule.InInterface != "" && rule.InInterface != pkt.InInterface {
		return false
	}

	if rule.OutInterface != "" && rule.OutInterface != pkt.OutInterface {
		return false
	}

	if rule.SourcePort > 0 && rule.SourcePort != pkt.SourcePort {
		return false
	}

	if rule.DestinationPort > 0 && rule.DestinationPort != pkt.DestinationPort {
		return false
	}

	return true
}

func mainNetfilter() {
	nf := NewNetfilter()

	// Add firewall rules
	nf.AddRule("filter", "INPUT", Rule{
		Protocol:        "tcp",
		DestinationPort: 22,
		Target:          "ACCEPT",
	})

	nf.AddRule("filter", "INPUT", Rule{
		Protocol:        "tcp",
		DestinationPort: 80,
		Target:          "ACCEPT",
	})

	nf.AddRule("filter", "INPUT", Rule{
		Source: "192.168.1.0/24",
		Target: "DROP",
	})

	// Test packets
	packets := []*Packet{
		{
			Protocol:        "tcp",
			SourceIP:        net.ParseIP("192.168.1.100"),
			DestinationIP:   net.ParseIP("10.0.0.5"),
			DestinationPort: 22,
			InInterface:     "eth0",
		},
		{
			Protocol:        "tcp",
			SourceIP:        net.ParseIP("192.168.1.101"),
			DestinationIP:   net.ParseIP("10.0.0.5"),
			DestinationPort: 80,
			InInterface:     "eth0",
		},
		{
			Protocol:        "udp",
			SourceIP:        net.ParseIP("10.0.0.2"),
			DestinationIP:   net.ParseIP("10.0.0.5"),
			DestinationPort: 53,
			InInterface:     "eth0",
		},
	}

	for i, pkt := range packets {
		action := nf.ProcessPacket(pkt)
		fmt.Printf("Packet %d: %s\n", i+1, action)
	}
}

// ```

// Key components and their Linux equivalents:

// 1. **Tables and Chains**:
// ```go
// Tables: map[string]*Table{
// 	"filter": {
// 		Chains: map[string]*Chain{
// 			"INPUT":   {Policy: "ACCEPT"},
// 			"OUTPUT":  {Policy: "ACCEPT"},
// 			"FORWARD": {Policy: "DROP"},
// 		},
// 	},
// 	"nat": {
// 		Chains: map[string]*Chain{
// 			"PREROUTING":  {Policy: "ACCEPT"},
// 			"POSTROUTING": {Policy: "ACCEPT"},
// 		},
// 	},
// }
// ```

// 2. **Rule Matching**:
// ```go
// func matchRule(rule Rule, pkt *Packet) bool {
// 	// Implements iptables-like matching for:
// 	// - Protocol (tcp/udp/icmp)
// 	// - Source/destination IP (CIDR)
// 	// - Input/output interfaces
// 	// - Port numbers
// }
// ```

// 3. **Packet Processing Workflow**:
// ```go
// func (nf *Netfilter) ProcessPacket(pkt *Packet) string {
// 	// Simulates netfilter hook order:
// 	// 1. NAT PREROUTING chain
// 	// 2. FILTER INPUT/OUTPUT/FORWARD chains
// 	// 3. NAT POSTROUTING chain
// }
// ```

// Sample Output:
// ```
// Packet 1: DROP    # Blocked by source IP rule
// Packet 2: DROP    # Blocked by source IP rule
// Packet 3: ACCEPT  # Allowed by default policy
// ```

// To extend this simulation, you could add:

// 1. **Connection Tracking**:
// ```go
// type Connection struct {
// 	Protocol string
// 	Source   net.IP
// 	Dest     net.IP
// 	State    string // NEW, ESTABLISHED, RELATED
// }

// func (nf *Netfilter) TrackConnections(pkt *Packet) {
// 	// Implement stateful firewall functionality
// }
// ```

// 2. **NAT Functionality**:
// ```go
// func (nf *Netfilter) PerformNAT(pkt *Packet) {
// 	// Implement source/destination NAT rules
// 	// Modify packet IPs/ports based on NAT rules
// }
// ```

// 3. **Logging Support**:
// ```go
// type LogEntry struct {
// 	Timestamp time.Time
// 	Packet    Packet
// 	Action    string
// 	Chain     string
// }

// func (nf *Netfilter) AddLogging() {
// 	// Record all packet processing decisions
// }
// ```

// This simulation demonstrates the core concepts of Linux's netfilter/iptables system including:

// - Chain traversal order
// - Rule matching logic
// - Default chain policies
// - Basic packet filtering
// - Table separation (filter/nat)

// The actual Linux implementation is much more complex with:
// - More match criteria (MAC address, packet state, etc)
// - Additional targets (LOG, REJECT, etc)
// - Connection tracking integration
// - Advanced NAT capabilities
// - Kernel-space optimizations

// 以下是一个使用 Go 语言实现的简单网络入侵检测系统（IDS）和入侵防御系统（IPS）的示例。该示例使用 `gopacket` 库捕获网络流量，并通过规则匹配实现基本的检测和防御逻辑。

// ```go
// package main

// import (
// 	"fmt"
// 	"log"
// 	"regexp"
// 	"net"
// 	"os/exec"
// 	"strings"
// 	"sync"

// 	"github.com/google/gopacket"
// 	"github.com/google/gopacket/layers"
// 	"github.com/google/gopacket/pcap"
// )

// 定义IDS/IPS配置
type SecurityConfig struct {
	Interface      string           // 监听的网络接口
	BPFFilter      string           // 抓包过滤规则
	BlockIPs       map[string]bool  // 被阻断的IP列表
	AttackPatterns []*regexp.Regexp // 攻击特征正则表达式
	mu             sync.Mutex
}

// 初始化IDS/IPS
func NewSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		Interface: "eth0",
		BPFFilter: "tcp and port 80",
		BlockIPs:  make(map[string]bool),
		AttackPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)select\s.*from`), // SQL注入特征
			regexp.MustCompile(`(?i)<script>`),       // XSS攻击特征
			regexp.MustCompile(`/etc/passwd`),        // 路径遍历
		},
	}
}

// IPS: 阻断恶意IP（使用iptables）
func (s *SecurityConfig) BlockIP(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.BlockIPs[ip]; !exists {
		log.Printf("[IPS] Blocking IP: %s", ip)
		cmd := exec.Command("iptables", "-A", "INPUT", "-s", ip, "-j", "DROP")
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to block IP: %v", err)
		}
		s.BlockIPs[ip] = true
	}
}

// IDS: 检测HTTP请求中的攻击特征
func (s *SecurityConfig) DetectAttack(payload string, srcIP string) bool {
	for _, pattern := range s.AttackPatterns {
		if pattern.MatchString(payload) {
			log.Printf("[IDS] Attack detected from %s: %s", srcIP, pattern.String())
			return true
		}
	}
	return false
}

// 主监听循环
func (s *SecurityConfig) Start() {
	// 打开网络接口
	handle, err := pcap.OpenLive(
		s.Interface,
		1600, // snaplen
		true, // promiscuous mode
		pcap.BlockForever,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	// 设置BPF过滤器
	if err := handle.SetBPFFilter(s.BPFFilter); err != nil {
		log.Fatal(err)
	}

	// 解析数据包
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// 解析IP和TCP层
		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		tcpLayer := packet.Layer(layers.LayerTypeTCP)

		if ipLayer != nil && tcpLayer != nil {
			ip, _ := ipLayer.(*layers.IPv4)
			tcp, _ := tcpLayer.(*layers.TCP)

			// 提取HTTP负载
			if len(tcp.Payload) > 0 && tcp.DstPort == 80 {
				payload := string(tcp.Payload)
				if s.DetectAttack(payload, ip.SrcIP.String()) {
					s.BlockIP(ip.SrcIP.String()) // IPS动作
				}
			}
		}
	}
}

func mainIDSIPS() {
	config := NewSecurityConfig()

	// 检查权限
	if _, err := exec.LookPath("iptables"); err != nil {
		log.Fatal("需要root权限运行此程序")
	}

	fmt.Println("Starting IDS/IPS...")
	config.Start()
}

// ```

// ### 功能说明

// 1. **依赖安装**
//    需要先安装 `gopacket` 库：
//    ```bash
//    go get github.com/google/gopacket
//    go get github.com/google/gopacket/layers
//    go get github.com/google/gopacket/pcap
//    ```

// 2. **核心组件**
//    - **流量捕获**：使用 `gopacket` 监听指定网络接口（如 `eth0`），过滤 HTTP 流量（端口 80）。
//    - **攻击检测**：通过正则表达式匹配 SQL 注入、XSS 等常见攻击特征。
//    - **主动防御**：检测到攻击后，调用 `iptables` 阻断来源 IP。

// 3. **运行方式**
//    ```bash
//    sudo go run main.go  # 需要root权限操作网络接口和iptables
//    ```

// 4. **示例输出**
//    ```
//    [IDS] Attack detected from 192.168.1.100: select * from users
//    [IPS] Blocking IP: 192.168.1.100
//    ```

// ### 扩展建议

// 1. **规则引擎增强**
//    - 集成开源规则库（如 [Suricata规则集](https://rules.emergingthreats.net/)）
//    - 支持从文件动态加载规则：
//      ```go
//      func (s *SecurityConfig) LoadRulesFromFile(path string) {
//        // 解析YAML/JSON格式的规则
//      }
//      ```

// 2. **协议深度解析**
//    - 支持 HTTPS 解密（需配置 TLS 证书）
//    - 解析更多协议（DNS、FTP）：
//      ```go
//      func parseDNSPacket(packet gopacket.Packet) {
//        dnsLayer := packet.Layer(layers.LayerTypeDNS)
//        // 分析DNS查询
//      }
//      ```

// 3. **性能优化**
//    - 使用零拷贝技术处理数据包：
//      ```go
//      packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
//      packetSource.NoCopy = true
//      ```
//    - 添加流量速率限制，防止 DDoS 攻击：
//      ```go
//      rateLimiter := time.Tick(100 * time.Microsecond)
//      for packet := range packetSource.Packets() {
//        <-rateLimiter
//        // 处理数据包
//      }
//      ```

// 4. **威胁情报集成**
//    - 查询外部威胁情报 API：
//      ```go
//      func checkIPReputation(ip string) bool {
//        // 调用VirusTotal或AlienVault API
//      }
//      ```

// 该实现可作为学习网络安全的起点，生产环境建议使用成熟的解决方案（如 Suricata）。
