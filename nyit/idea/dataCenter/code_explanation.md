# BCube OpenFlow Configuration - Detailed Code Explanation

## Table of Contents
1. [Overview](#overview)
2. [BCube Topology Structure](#bcube-topology-structure)
3. [Host Routing Configuration](#host-routing-configuration)
4. [Linux Networking Commands](#linux-networking-commands)
5. [OpenFlow Rules](#openflow-rules)
6. [Complete Data Flow Example](#complete-data-flow-example)

---

## Overview

This code implements a **BCube(3,2) topology** with **Software-Defined Networking (SDN)** using OpenFlow. The key challenge is creating 5 specific bidirectional paths where hosts have **multiple network interfaces** and must use the correct interface for each path.

### The Problem
In BCube topology:
- Each host has **8 network interfaces** (eth0-eth7)
- Each interface connects to a **different switch** at different levels
- Only **eth0 has an IP address** (10.0.0.X) by default
- Other interfaces (eth1-eth7) have **no IP addresses**
- We need to force traffic through **specific interfaces** for each path

---

## BCube Topology Structure

### Topology Layout
```
BCube(3,2): k=3, n=2
- 8 mini-cubes (2^k)
- 2 hosts per cube (n)
- 4 switch levels (k+1): Level 0, 1, 2, 3
- Total: 16 hosts, 32 switches
```

### Host Connections Example
```
h00 has 8 interfaces:
├── eth0 → s00 (Level-0, has IP 10.0.0.1)
├── eth1 → s10 (Level-1, no IP)
├── eth2 → s20 (Level-2, no IP)
├── eth3 → s22 (Level-2, no IP)
├── eth4 → s30 (Level-3, no IP) ← Used for RED & GREEN paths
├── eth5 → s32 (Level-3, no IP)
├── eth6 → s34 (Level-3, no IP)
└── eth7 → s36 (Level-3, no IP)
```

### The 5 Required Paths
1. **RED**: h00 ↔ h40 via s30
2. **GREEN**: h00 ↔ h50 via s30 → h40 → s14 (h40 acts as relay)
3. **BLUE**: h20 ↔ h30 via s12
4. **PURPLE**: h60 ↔ h61 via s06
5. **BLACK**: h60 ↔ h70 via s16

---

## Host Routing Configuration

The [`configure_host_routing()`](assignment1.py:57) function solves the multi-interface problem using three Linux networking commands.

### Step 1: Find the Correct Interface

```python
# Example: Find h00's interface that connects to s30
h00_s30_intf = None
h00_s30_mac = None
for intf in h00.intfList():
    if intf.name != 'lo' and intf.link:
        if intf.link.intf1.node == s30 or intf.link.intf2.node == s30:
            h00_s30_intf = intf.name    # e.g., "h00-eth4"
            h00_s30_mac = intf.MAC()    # e.g., "ba:c1:c1:00:53:03"
            break
```

**What this does:**
- Iterates through all interfaces on h00
- Checks if the interface's link connects to switch s30
- Stores the interface name (h00-eth4) and MAC address

---

## Linux Networking Commands

### 1. `ip route add` - Static Route Configuration

#### Command Format
```bash
ip route add <destination_ip> dev <interface_name>
```

#### Example from Code (Line 82)
```python
h00.cmd(f'ip route add {h40.IP()} dev {h00_s30_intf}')
# Executes: ip route add 10.0.0.9 dev h00-eth4
```

#### What It Does

**Purpose:** Tells the Linux kernel which network interface to use when sending packets to a specific destination.

**Before the command:**
```bash
h00 routing table:
10.0.0.0/8 dev h00-eth0  <- Default: ALL traffic goes via eth0
```

**After the command:**
```bash
h00 routing table:
10.0.0.0/8 dev h00-eth0  <- Default route
10.0.0.9 dev h00-eth4    <- Specific route for h40's IP
```

**How it works:**
1. h00 wants to ping h40 (10.0.0.9)
2. Kernel checks routing table
3. Finds specific route: `10.0.0.9 dev h00-eth4`
4. Sends packet via **h00-eth4** (not eth0!)
5. Packet goes to s30 (which h00-eth4 connects to)

#### Why It's Critical

Without this command:
```
h00 ping h40 → Uses eth0 → Goes to s00 → BLOCKED (no flow rules on s00)
```

With this command:
```
h00 ping h40 → Uses eth4 → Goes to s30 → SUCCESS (flow rules on s30!)
```

---

### 2. `arp -s` - Static ARP Entry

#### Command Format
```bash
arp -s <ip_address> <mac_address>
```

#### Example from Code (Line 84)
```python
h00.cmd(f'arp -s {h40.IP()} {h40_s30_mac}')
# Executes: arp -s 10.0.0.9 ba:c1:c1:00:53:03
```

#### What It Does

**Purpose:** Creates a **permanent** mapping between an IP address and a MAC address, bypassing normal ARP discovery.

#### Understanding ARP (Address Resolution Protocol)

**Normal ARP Process:**
```
Step 1: h00 wants to send packet to 10.0.0.9
Step 2: h00 checks ARP cache - no entry found
Step 3: h00 broadcasts: "Who has 10.0.0.9? Tell 10.0.0.1"
Step 4: h40 responds: "10.0.0.9 is at MAC aa:bb:cc:dd:ee:ff"
Step 5: h00 caches this and sends packet
```

**Problem in BCube:**
```
h00 (eth4, no IP) → s30 → h40 (eth4, no IP)
                ↓
ARP broadcast goes via eth0 (default interface)
                ↓
h40's eth0 responds (wrong interface!)
                ↓
h00 gets wrong MAC address
                ↓
Packet sent to wrong MAC → FAILS
```

**Solution with Static ARP:**
```python
h00.cmd(f'arp -s {h40.IP()} {h40_s30_mac}')
```

**Effect:**
```bash
# Before:
h00> arp -a
# (empty or wrong entry)

# After:
h00> arp -a
? (10.0.0.9) at ba:c1:c1:00:53:03 [ether] PERM on h00-eth4
```

#### How It Works in Packet Flow

```
1. h00 wants to send to 10.0.0.9
2. Routing table says: use h00-eth4
3. Need MAC address for 10.0.0.9
4. Check ARP cache: Found! ba:c1:c1:00:53:03 (h40's eth4 MAC)
5. Build Ethernet frame:
   - Source MAC: h00-eth4's MAC
   - Dest MAC: h40-eth4's MAC (ba:c1:c1:00:53:03)
   - IP Packet inside
6. Send frame via h00-eth4
7. Frame arrives at s30
8. OpenFlow rule matches, forwards to h40-eth4
9. h40-eth4 receives (MAC matches!)
```

#### Why Both IP Route AND ARP Are Needed

**IP route alone:**
```
✓ Tells which interface to use
✗ Doesn't know destination MAC address
✗ Would try ARP broadcast (goes via wrong interface)
```

**ARP alone:**
```
✓ Knows destination MAC address
✗ Doesn't know which interface to use
✗ Would use default interface (eth0)
```

**Both together:**
```
✓ IP route: Use h00-eth4
✓ ARP: Destination MAC is ba:c1:c1:00:53:03
✓ Complete: Send frame via h00-eth4 to that MAC
✓ Success!
```

---

### 3. `sysctl -w` - Kernel Parameter Configuration

#### Command Format
```bash
sysctl -w parameter.name=value
```

#### Examples from Code (Lines 89-94)

```python
# 1. Enable IP forwarding
h40.cmd('sysctl -w net.ipv4.ip_forward=1')

# 2. Disable reverse path filtering
h40.cmd('sysctl -w net.ipv4.conf.all.rp_filter=0')
h40.cmd('sysctl -w net.ipv4.conf.default.rp_filter=0')
h40.cmd(f'sysctl -w net.ipv4.conf.{intf.name}.rp_filter=0')
```

#### 3a. `net.ipv4.ip_forward=1` - Enable IP Forwarding

**Purpose:** Allows a host to forward packets between its network interfaces (act as a router).

**Default Behavior (ip_forward=0):**
```
Packet arrives at h40-eth4 destined for h50:
1. h40 checks: Is this packet for me? (10.0.0.9)
2. No, it's for 10.0.0.11 (h50)
3. ip_forward=0 → DROP the packet
4. h50 never receives it
```

**With ip_forward=1:**
```
Packet arrives at h40-eth4 destined for h50:
1. h40 checks: Is this packet for me?
2. No, it's for 10.0.0.11
3. ip_forward=1 → Check routing table
4. Route says: 10.0.0.11 dev h40-eth1
5. Forward packet via h40-eth1 → s14 → h50
6. Success!
```

**Why Needed for GREEN Path:**
```
h00 → s30 → h40-eth4 [h40 RELAY] h40-eth1 → s14 → h50
              ↑
    Must forward between eth4 and eth1
    Only possible with ip_forward=1
```

#### 3b. `net.ipv4.conf.all.rp_filter=0` - Disable Reverse Path Filtering

**Purpose:** Allows packets to arrive and leave via different interfaces.

**What is Reverse Path Filtering?**

Reverse Path Filter (RPF) is a security feature that prevents IP spoofing:

```
When packet arrives on interface X:
1. Check packet's source IP
2. Lookup route to that source IP
3. Would we send reply via interface X?
4. If NO → DROP packet (anti-spoofing)
```

**Problem in BCube GREEN Path:**

```
Packet flow: h00 → s30 → h40-eth4
├── Source IP: 10.0.0.1 (h00)
├── Dest IP: 10.0.0.11 (h50)
└── Arrives at h40-eth4

RPF Check on h40:
1. Packet from 10.0.0.1 arrived on eth4
2. Check routing table for 10.0.0.1
3. Route says: 10.0.0.1 dev h40-eth4 ✓
4. Reply WOULD go via eth4 → PASS

Forward to eth1:
├── Now packet leaves via h40-eth1
└── No problem here

Return packet: h50 → s14 → h40-eth1
├── Source IP: 10.0.0.11 (h50)
├── Dest IP: 10.0.0.1 (h00)
└── Arrives at h40-eth1

RPF Check on h40:
1. Packet from 10.0.0.11 arrived on eth1
2. Check routing table for 10.0.0.11
3. Route says: 10.0.0.11 dev h40-eth1 ✓
4. Reply WOULD go via eth1 → PASS

Forward to eth4:
├── Packet should leave via h40-eth4
├── Destination is 10.0.0.1
└── Should work... BUT

RPF on eth4 (outgoing check):
1. Sending to 10.0.0.1 via eth4
2. Original packet came from 10.0.0.11
3. If reply comes, it would arrive via eth1
4. rp_filter=1 → Different interfaces!
5. DROP packet → FAIL
```

**With rp_filter=0:**
```
All RPF checks disabled
Packets can enter/exit any interface
h40 forwards between eth4 ↔ eth1
GREEN path works!
```

#### Summary of sysctl Parameters

| Parameter | Value | Effect | Needed For |
|-----------|-------|--------|------------|
| `net.ipv4.ip_forward` | 1 | Enable packet forwarding | GREEN path (h40 relay) |
| `net.ipv4.conf.all.rp_filter` | 0 | Disable RPF globally | GREEN path (multi-interface) |
| `net.ipv4.conf.default.rp_filter` | 0 | Disable RPF for new interfaces | GREEN path |
| `net.ipv4.conf.{intf}.rp_filter` | 0 | Disable RPF per interface | GREEN path |

---

## OpenFlow Rules

The [`add_flows_bcube()`](assignment1.py:166) function configures switch behavior using OpenFlow.

### OpenFlow Rule Structure

```bash
ovs-ofctl add-flow <switch> 'priority=<num>,<match_fields>,actions=<actions>'
```

### Three Types of Rules

#### 1. Default Drop (Priority 0)
```python
os.system(f"ovs-ofctl add-flow {sw.name} 'priority=0,actions=drop'")
```

**What it does:**
- **Lowest priority** (0)
- Matches **any packet**
- **Drops** packet
- Acts as a **firewall** - blocks all communication not explicitly allowed

#### 2. ARP Allow (Priority 200)
```python
os.system(f"ovs-ofctl add-flow {sw.name} 'priority=200,arp,actions=normal'")
```

**What it does:**
- **Highest priority** (200)
- Matches **ARP packets** only
- action=**normal** means use normal L2 switching
- **Critical:** Without this, hosts can't discover MAC addresses

#### 3. Specific Path Rules (Priority 100)
```python
os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h40.IP()},actions=output:{p_h40}'")
```

**What it does:**
- **Medium priority** (100)
- Matches **IP packets**
- **nw_src**: Source IP must be h00 (10.0.0.1)
- **nw_dst**: Destination IP must be h40 (10.0.0.9)
- **actions=output:{p_h40}**: Forward to port connected to h40

### Priority System

```
Priority 200: ARP (Allow for address resolution)
Priority 100: Specific paths (Only these 5 paths)
Priority 0:   Default DROP (Block everything else)
```

### Example: RED Path Flow Rules

```python
# s30 forwards h00 ↔ h40
os.system(f"ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.9,actions=output:2'")
os.system(f"ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.9,nw_dst=10.0.0.1,actions=output:1'")
```

**Meaning:**
- Rule 1: Packets from h00 to h40 → output port 2 (h40's port)
- Rule 2: Packets from h40 to h00 → output port 1 (h00's port)
- **Bidirectional** communication established

---

## Complete Data Flow Example

### RED Path: h00 ping h40

Let's trace a complete ping request:

#### Phase 1: Configuration (Done by script)

```python
# 1. Static route on h00
h00.cmd('ip route add 10.0.0.9 dev h00-eth4')

# 2. Static ARP on h00
h00.cmd('arp -s 10.0.0.9 ba:c1:c1:00:53:03')  # h40-eth4's MAC

# 3. OpenFlow rule on s30
os.system("ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.9,actions=output:2'")
```

#### Phase 2: Ping Execution

```bash
mininet> h00 ping -c 1 h40
```

#### Phase 3: Packet Journey

**Step 1: h00 creates ICMP packet**
```
ICMP Echo Request:
├── Source IP: 10.0.0.1 (h00)
├── Dest IP: 10.0.0.9 (h40)
└── Type: Echo Request
```

**Step 2: h00 checks routing table**
```
Query: Where to send 10.0.0.9?
Answer: Route found → 10.0.0.9 dev h00-eth4
Decision: Use interface h00-eth4
```

**Step 3: h00 checks ARP cache**
```
Query: What is MAC address for 10.0.0.9?
Answer: ARP cache → ba:c1:c1:00:53:03
Decision: Use this MAC as destination
```

**Step 4: h00 builds Ethernet frame**
```
Ethernet Frame:
├── Source MAC: d2:57:a3:c5:19:d8 (h00-eth4)
├── Dest MAC: ba:c1:c1:00:53:03 (h40-eth4)
└── Payload: IP packet with ICMP
```

**Step 5: Frame sent via h00-eth4**
```
h00-eth4 → wire → s30-eth1
```

**Step 6: s30 receives frame**
```
Frame arrives at s30 port 1
OpenFlow Table Lookup:
├── Check priority 200 (ARP): No match (this is IP)
├── Check priority 100 rules:
│   ├── Match: ip, nw_src=10.0.0.1, nw_dst=10.0.0.9
│   └── Action: output:2
└── Execute: Forward frame to port 2
```

**Step 7: s30 forwards to h40**
```
s30-eth2 → wire → h40-eth4
```

**Step 8: h40 receives frame**
```
h40-eth4 receives frame
Check destination MAC: ba:c1:c1:00:53:03
Match! This is for me
Check destination IP: 10.0.0.9
Match! This is my IP
Extract ICMP: Echo Request
Generate ICMP Echo Reply
```

**Step 9: Return path (h40 → h00)**
```
Same process in reverse:
1. h40 checks route for 10.0.0.1 → use h40-eth4
2. h40 checks ARP for 10.0.0.1 → d2:57:a3:c5:19:d8
3. Build frame, send via h40-eth4
4. s30 matches reverse flow rule
5. Forward to port 1 (h00)
6. h00-eth4 receives reply
7. PING successful!
```

---

## GREEN Path: Special Case with Relay

### Configuration Differences

```python
# h00 configuration
h00.cmd('ip route add 10.0.0.11 dev h00-eth4')
h00.cmd('arp -s 10.0.0.11 ba:c1:c1:00:53:03')  # h40-eth4 MAC, NOT h50's!

# h40 configuration (RELAY)
h40.cmd('sysctl -w net.ipv4.ip_forward=1')  # Enable forwarding
h40.cmd('sysctl -w net.ipv4.conf.all.rp_filter=0')  # Disable RPF
h40.cmd('ip route add 10.0.0.11 dev h40-eth1')  # h50 via eth1
h40.cmd('ip route add 10.0.0.1 dev h40-eth4')   # h00 via eth4
h40.cmd('arp -s 10.0.0.11 <h50-eth1-mac>')  # Real h50 MAC
h40.cmd('arp -s 10.0.0.1 <h00-eth4-mac>')   # Real h00 MAC

# h50 configuration
h50.cmd('ip route add 10.0.0.1 dev h50-eth1')
h50.cmd('arp -s 10.0.0.1 <h40-eth1-mac>')  # h40-eth1 MAC, NOT h00's!
```

### The Trick: MAC Address Deception

**Key insight:** h00 thinks h50 is at h40's MAC address!

```
h00's view:
├── IP 10.0.0.11 (h50) is at MAC ba:c1:c1:00:53:03
└── But that's actually h40-eth4's MAC!

h50's view:
├── IP 10.0.0.1 (h00) is at MAC <h40-eth1-mac>
└── But that's actually h40-eth1's MAC!

h40's view:
├── IP 10.0.0.1 is at <real-h00-eth4-mac>
├── IP 10.0.0.11 is at <real-h50-eth1-mac>
└── h40 knows the truth and forwards correctly
```

### Packet Journey: h00 ping h50

**Step 1-3:** Same as RED path (routing, ARP)

**Step 4:** h00 builds frame with **h40's MAC** as destination
```
Frame destined for h50 but MAC is h40's!
Source MAC: h00-eth4
Dest MAC: h40-eth4 (thinks this is h50!)
IP: 10.0.0.1 → 10.0.0.11
```

**Step 5-7:** s30 forwards to h40 (same as RED)

**Step 8:** h40 receives and forwards
```
h40-eth4 receives frame
MAC matches → Accept
IP is 10.0.0.11 → Not for me
ip_forward=1 → Forward it
Check routing: 10.0.0.11 dev h40-eth1
Check ARP: 10.0.0.11 at <real-h50-mac>
Build NEW frame:
├── Source MAC: h40-eth1 (changed!)
├── Dest MAC: real-h50-eth1-mac (changed!)
└── IP packet unchanged: 10.0.0.1 → 10.0.0.11
Send via h40-eth1
```

**Step 9:** s14 forwards to h50
```
Frame arrives at s14 from h40-eth1
OpenFlow rule matches: nw_src=10.0.0.1, nw_dst=10.0.0.11
Action: output to h50's port
Forward to h50-eth1
```

**Step 10:** h50 receives
```
h50-eth1 receives
MAC matches → Accept
IP matches → Process
ICMP Echo Request → Generate Reply
```

**Step 11-15:** Return path via same relay

---

## Why This Approach Works

### 1. Solves Multi-Interface Problem
- **Problem:** Hosts have 8 interfaces, only eth0 has IP
- **Solution:** Static routes force traffic via correct interface

### 2. Solves ARP Problem
- **Problem:** ARP broadcasts go via default interface
- **Solution:** Static ARP entries bypass broadcast

### 3. Enables Host-Based Relay
- **Problem:** h40 needs to forward between interfaces
- **Solution:** ip_forward=1 + rp_filter=0

### 4. Implements SDN Security
- **Problem:** Only 5 paths should work
- **Solution:** OpenFlow default-drop + specific rules

### 5. Works with BCube Architecture
- **Problem:** Non-traditional topology with multi-homed hosts
- **Solution:** Combination of static host config + OpenFlow

---

## Command Reference

### Check Configuration

```bash
# View routing table
mininet> h00 ip route

# View ARP cache  
mininet> h00 arp -a

# Check forwarding enabled
mininet> h40 sysctl net.ipv4.ip_forward

# Check RPF setting
mininet> h40 sysctl net.ipv4.conf.all.rp_filter

# View OpenFlow rules
mininet> sh ovs-ofctl dump-flows s30

# Check interface details
mininet> h00 ifconfig h00-eth4
```

### Manual Configuration Example

```bash
# Configure RED path manually
mininet> h00 ip route add 10.0.0.9 dev h00-eth4
mininet> h00 arp -s 10.0.0.9 <h40-eth4-mac>
mininet> h40 ip route add 10.0.0.1 dev h40-eth4  
mininet> h40 arp -s 10.0.0.1 <h00-eth4-mac>
mininet> sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.9,actions=output:2'
mininet> sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.9,nw_dst=10.0.0.1,actions=output:1'
mininet> h00 ping -c 3 h40
```

---

## Summary

This implementation combines three layers of networking:

1. **Layer 2 (Data Link):** Static ARP controls MAC address resolution
2. **Layer 3 (Network):** Static routes control interface selection  
3. **Layer 2.5 (SDN):** OpenFlow controls switch forwarding

Together, they create precise, controlled paths in a complex multi-interface BCube topology where traditional routing would fail.