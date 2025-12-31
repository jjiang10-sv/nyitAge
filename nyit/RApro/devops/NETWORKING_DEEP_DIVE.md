# AKS Overlay and Virtual IP Deep Dive

## Quick Answer

### Does AKS Use VXLAN?

**With Cilium (our setup):** 
- Default: **Geneve** (not VXLAN)
- Alternative: **VXLAN** (configurable)
- Best: **Native routing** with eBPF (no encapsulation)

**With Azure CNI Overlay (Microsoft's implementation):**
- Uses **Geneve** encapsulation

---

## Part 1: Overlay IPs Explained

### What is an Overlay Network?

An **overlay network** creates a virtual network **on top of** the physical network (underlay).

```
┌────────────────────────────────────────┐
│  Overlay Network (Virtual)             │
│  Pod IPs: 10.32.0.0/13                 │
│  ┌─────┐  ┌─────┐  ┌─────┐             │
│  │ Pod │  │ Pod │  │ Pod │             │
│  │.1.10│  │.1.20│  │.2.30│             │
│  └─────┘  └─────┘  └─────┘             │
└────────────────────────────────────────┘
              ▲ Encapsulation
              │ (Geneve/VXLAN)
              ▼
┌────────────────────────────────────────┐
│  Underlay Network (Physical VNet)      │
│  Node IPs: 10.0.0.0/16                 │
│  ┌──────┐  ┌──────┐  ┌──────┐          │
│  │ Node │  │ Node │  │ Node │          │
│  │ .0.5 │  │ .0.6 │  │ .0.7 │          │
│  └──────┘  └──────┘  └──────┘          │
└────────────────────────────────────────┘
```

### Why Use Overlay?

**Problem with Direct Routing:**
```bash
# If we used VNet IPs for pods directly:
Pod 1: 10.0.0.100
Pod 2: 10.0.0.101
Pod 3: 10.0.0.102
...
Pod 65,000: 10.0.255.255  ← VNet exhausted!
```

**Solution with Overlay:**
```bash
# Pods use separate IP space:
Pod 1: 10.32.0.1    ← Not in VNet
Pod 2: 10.32.0.2    ← Not in VNet
...
Pod 500,000: 10.39.255.255  ← Still not in VNet!

# VNet only sees nodes:
Node 1: 10.0.0.5    ← In VNet (only need a few)
Node 2: 10.0.0.6    ← In VNet
```

**Benefits:**
1. **Massive IP space** - 524k pod IPs without consuming VNet IPs
2. **IP mobility** - Pods can move between nodes
3. **Isolation** - Pod network separate from node network
4. **Flexibility** - Can use any CIDR for pods

---

## Part 2: How Overlay Works (Geneve/VXLAN)

### Encapsulation Process

#### Step 1: Pod sends packet

```
Pod A (10.32.1.10) → Pod B (10.32.2.20)

Original Packet:
┌─────────────────────────────────┐
│ Src: 10.32.1.10                 │
│ Dst: 10.32.2.20                 │
│ Data: "Hello"                   │
└─────────────────────────────────┘
```

#### Step 2: Node A encapsulates (Geneve/VXLAN)

```
Node A knows:
- "Pod B (10.32.2.20) is on Node B (10.0.0.6)"

Encapsulated Packet:
┌────────────────────────────────────────────┐
│ Outer Header (VNet routing)                │
│ Src: 10.0.0.5 (Node A)                     │
│ Dst: 10.0.0.6 (Node B)                     │
│ Protocol: Geneve (UDP 6081)                │
│                                            │
│  ┌───────────────────────────────────────┐│
│  │ Inner Header (Overlay)                ││
│  │ Src: 10.32.1.10 (Pod A)               ││
│  │ Dst: 10.32.2.20 (Pod B)               ││
│  │ Data: "Hello"                         ││
│  └───────────────────────────────────────┘│
└────────────────────────────────────────────┘
```

#### Step 3: VNet routes to Node B

```
Azure VNet sees:
- Packet from 10.0.0.5 to 10.0.0.6
- Routes using VNet routing table
- Doesn't see pod IPs at all!
```

#### Step 4: Node B decapsulates

```
Node B receives packet:
1. Removes outer header
2. Extracts inner packet
3. Delivers to Pod B

Pod B receives:
┌─────────────────────────────────┐
│ Src: 10.32.1.10                 │
│ Dst: 10.32.2.20                 │
│ Data: "Hello"                   │
└─────────────────────────────────┘
```

---

## Part 3: Geneve vs VXLAN

### VXLAN (Virtual Extensible LAN)

**Protocol:**
- UDP port 4789
- 24-bit VNI (Virtual Network ID)
- Header: 8 bytes

**Packet Structure:**
```
┌────────────────────────────┐
│ Outer Ethernet Header      │
├────────────────────────────┤
│ Outer IP Header            │
│ Src: Node A IP             │
│ Dst: Node B IP             │
├────────────────────────────┤
│ Outer UDP Header           │
│ Dst Port: 4789             │
├────────────────────────────┤
│ VXLAN Header               │
│ VNI: 1234 (24 bits)        │
├────────────────────────────┤
│ Inner Ethernet Frame       │
│ (Original pod packet)      │
└────────────────────────────┘
```

**Pros:**
- ✅ Widely supported
- ✅ Hardware offload available
- ✅ Mature technology

**Cons:**
- ❌ Limited metadata (only VNI)
- ❌ Fixed header format
- ❌ Limited extensibility

---

### Geneve (Generic Network Virtualization Encapsulation)

**Protocol:**
- UDP port 6081
- 24-bit VNI (compatible with VXLAN)
- Variable-length options
- Header: 8+ bytes (extensible)

**Packet Structure:**
```
┌────────────────────────────┐
│ Outer Ethernet Header      │
├────────────────────────────┤
│ Outer IP Header            │
│ Src: Node A IP             │
│ Dst: Node B IP             │
├────────────────────────────┤
│ Outer UDP Header           │
│ Dst Port: 6081             │
├────────────────────────────┤
│ Geneve Header              │
│ VNI: 1234                  │
│ Options: (metadata)        │
│  - Security labels         │
│  - QoS info                │
│  - Custom fields           │
├────────────────────────────┤
│ Inner Ethernet Frame       │
│ (Original pod packet)      │
└────────────────────────────┘
```

**Pros:**
- ✅ Extensible (can add metadata)
- ✅ Better for cloud-native
- ✅ Supports security labels, QoS
- ✅ Future-proof

**Cons:**
- ❌ Newer (less hardware offload)
- ❌ Slightly larger headers

---

## Part 4: AKS with Cilium - The Reality

### Our Configuration

```python
network_profile=ContainerServiceNetworkProfileArgs(
    network_plugin="azure",
    network_plugin_mode="overlay",    # Enables overlay
    network_dataplane="cilium",       # Uses Cilium
    pod_cidr="10.32.0.0/13",
)
```

### What Cilium Actually Uses

**Default: Native Routing (No Encapsulation!)**

Cilium with eBPF tries to avoid encapsulation entirely:

```
┌────────────────────────────────────┐
│ Cilium eBPF Routing (Best Case)   │
│                                    │
│ Pod A → eBPF program               │
│         ↓                          │
│         Direct routing             │
│         ↓                          │
│         Pod B                      │
│                                    │
│ No encapsulation!                  │
│ Uses kernel routing table          │
└────────────────────────────────────┘
```

**Fallback: Geneve/VXLAN (When Needed)**

Cilium uses encapsulation when:
- Cross-node communication
- Network doesn't support native routing
- Specific security policies require it

**Configuration:**
```yaml
# Cilium ConfigMap
tunnel: "disabled"     # No encapsulation (best)
tunnel: "geneve"       # Geneve encapsulation
tunnel: "vxlan"        # VXLAN encapsulation
```

**In AKS with Azure CNI Overlay:**
- Azure manages the overlay
- Uses **Geneve** by default
- Cilium handles the dataplane (eBPF)

---

## Part 5: Service IPs - The "Virtual" Magic

### What is a Service IP?

A Service IP is **completely virtual** - it doesn't exist on any network interface!

```bash
# Create a service
kubectl create service clusterip my-svc --tcp=80:8080

# Service gets IP
NAME     TYPE        CLUSTER-IP     PORT(S)
my-svc   ClusterIP   10.96.10.50    80/TCP

# But this IP doesn't exist anywhere!
ping 10.96.10.50  # ❌ Won't work (not a real interface)
```

### How Service IPs Work

#### Traditional (kube-proxy with iptables)

```
┌──────────────────────────────────────────┐
│ 1. Pod tries to connect to service       │
│    curl 10.96.10.50:80                   │
└────────────────┬─────────────────────────┘
                 ↓
┌──────────────────────────────────────────┐
│ 2. Packet hits iptables rules            │
│    (kube-proxy sets these up)            │
│                                          │
│    iptables -t nat -A PREROUTING         │
│    -d 10.96.10.50 -p tcp --dport 80      │
│    -j DNAT --to-destination 10.32.1.10   │
│                  (backend pod IP)        │
└────────────────┬─────────────────────────┘
                 ↓
┌──────────────────────────────────────────┐
│ 3. Packet rewritten                      │
│    Before: dst=10.96.10.50:80            │
│    After:  dst=10.32.1.10:8080           │
└────────────────┬─────────────────────────┘
                 ↓
┌──────────────────────────────────────────┐
│ 4. Delivered to backend pod              │
│    Pod receives packet                   │
│    Thinks it came from service IP        │
└──────────────────────────────────────────┘
```

**The service IP (10.96.10.50) never actually exists!**

---

#### Modern (Cilium with eBPF)

```
┌──────────────────────────────────────────┐
│ 1. Pod attempts connection               │
│    connect(10.96.10.50:80)               │
└────────────────┬─────────────────────────┘
                 ↓
┌──────────────────────────────────────────┐
│ 2. eBPF program intercepts               │
│    (attached to network interface)       │
│                                          │
│    if (dst == 10.96.10.50) {             │
│      dst_ip = select_backend()           │
│      dst_ip = 10.32.1.10                 │
│    }                                     │
│                                          │
│    All in kernel, no iptables!           │
└────────────────┬─────────────────────────┘
                 ↓
┌──────────────────────────────────────────┐
│ 3. Direct to backend                     │
│    Packet sent to 10.32.1.10:8080        │
│    Much faster than iptables!            │
└──────────────────────────────────────────┘
```

**Why "Virtual"?**
- Not assigned to any interface
- Exists only in iptables/eBPF rules
- Load balanced across backend pods
- Stable even when pods restart

---

## Part 6: The Complete Flow

### Scenario: Pod A calls Service → Pod B

```
┌─────────────────────────────────────────────────────────┐
│ Node 1                          Node 2                  │
│                                                         │
│ ┌──────────────┐                ┌──────────────┐       │
│ │ Pod A        │                │ Pod B        │       │
│ │ 10.32.1.10   │                │ 10.32.2.20   │       │
│ │              │                │ (backend)    │       │
│ │ curl service │                │              │       │
│ └──────┬───────┘                └───────▲──────┘       │
│        │                                 │              │
│        │ 1. Connect to                   │ 5. Receives  │
│        │    10.96.10.50:80               │    packet    │
│        ↓                                 │              │
│ ┌─────────────────────────────────────┐ │              │
│ │ Cilium eBPF (on Node 1)             │ │              │
│ │                                     │ │              │
│ │ 2. Service lookup:                  │ │              │
│ │    10.96.10.50 → backends:          │ │              │
│ │    - 10.32.2.20 (Node 2)            │ │              │
│ │    - 10.32.3.30 (Node 3)            │ │              │
│ │                                     │ │              │
│ │ 3. Load balance → 10.32.2.20        │ │              │
│ └─────────────┬───────────────────────┘ │              │
│               ↓                          │              │
│        4a. Encapsulate (if cross-node)  │              │
│            Outer: 10.0.0.5 → 10.0.0.6   │              │
│            Inner: 10.32.1.10→10.32.2.20 │              │
│               │                          │              │
│               └──────Azure VNet──────────┘              │
│                                          │              │
│                        4b. Decapsulate   │              │
│                            on Node 2 ────┘              │
└─────────────────────────────────────────────────────────┘
```

**Key Points:**
1. Service IP (10.96.10.50) = Virtual (eBPF lookup)
2. Pod IPs (10.32.x.x) = Overlay (Geneve encapsulation)
3. Node IPs (10.0.x.x) = Underlay (VNet routing)

---

## Part 7: Azure CNI Overlay Architecture

### Full Stack

```
┌──────────────────────────────────────────────────┐
│ Application Layer                                │
│ Pod: 10.32.1.10                                  │
└────────────────┬─────────────────────────────────┘
                 ↓
┌──────────────────────────────────────────────────┐
│ Service Layer (Virtual IPs)                      │
│ Service: 10.96.10.50 (eBPF maps to backend)      │
└────────────────┬─────────────────────────────────┘
                 ↓
┌──────────────────────────────────────────────────┐
│ Overlay Layer (Pod Network)                      │
│ CNI: Cilium                                      │
│ Encapsulation: Geneve (UDP 6081)                │
│ Pod CIDR: 10.32.0.0/13                           │
└────────────────┬─────────────────────────────────┘
                 ↓
┌──────────────────────────────────────────────────┐
│ Underlay Layer (Node Network)                    │
│ Azure VNet: 10.0.0.0/14                          │
│ Routing: Azure SDN                               │
│ Firewall: Azure Firewall / NSGs                  │
└────────────────┬─────────────────────────────────┘
                 ↓
┌──────────────────────────────────────────────────┐
│ Physical Layer                                   │
│ Azure Infrastructure                             │
└──────────────────────────────────────────────────┘
```

---

## Part 8: Verification Commands

### Check Overlay Configuration

```bash
# Check Cilium tunnel mode
kubectl -n kube-system exec -ti ds/cilium -- cilium status | grep -i tunnel

# Output examples:
# "Encapsulation: Geneve"   ← Using Geneve
# "Encapsulation: VXLAN"    ← Using VXLAN  
# "Encapsulation: Disabled" ← Native routing
```

### Inspect Geneve Tunnels

```bash
# On AKS node (requires SSH access)
ip -d link show  # Look for genev_ interfaces

# Example output:
# genev_sys_6081: <BROADCAST,MULTICAST,UP,LOWER_UP>
#     link/ether 9a:f2:44:17:26:e9
#     geneve id 1 remote 10.0.0.6 ttl auto dstport 6081
```

### Monitor Service Mappings (eBPF)

```bash
# View Cilium service mappings
kubectl -n kube-system exec -ti ds/cilium -- cilium service list

# Output:
# ID   Frontend           Backend
# 1    10.96.0.1:443      10.32.1.5:6443
# 2    10.96.10.50:80     10.32.2.20:8080
#                         10.32.3.30:8080
```

### Capture Encapsulated Traffic

```bash
# On node (requires access)
tcpdump -i eth0 'udp port 6081' -vv

# You'll see Geneve packets:
# IP node1.6081 > node2.6081: Geneve, Flags [C]
#   vni 0x1, proto TEB (0x6558)
#   IP pod1 > pod2: ICMP echo request
```

---

## Part 9: Performance Comparison

### Encapsulation Overhead

| Method | Overhead | Latency | Throughput | CPU |
|--------|----------|---------|------------|-----|
| **Native Routing** | 0 bytes | Lowest | Highest | Lowest |
| **Geneve** | ~50 bytes | +5-10% | -5-10% | +10-15% |
| **VXLAN** | ~50 bytes | +5-10% | -5-10% | +10-15% |

### eBPF vs iptables (Service Routing)

| Metric | iptables | eBPF (Cilium) |
|--------|----------|---------------|
| **Latency** | ~100µs | ~10µs (10x faster) |
| **Rule Scale** | O(n) | O(1) |
| **CPU Usage** | High | Low |
| **Connection Tracking** | Limited | Advanced |

**With 10,000 services:**
- **iptables**: 100,000+ rules, slow
- **eBPF**: Hash map lookup, fast

---

## Part 10: Why This Design?

### Overlay Benefits

1. **IP Exhaustion Solved**
   ```
   VNet: 65k IPs → only for nodes
   Pods: 524k IPs → separate overlay
   ```

2. **Mobility**
   ```
   Pod IP remains same when:
   - Moving between nodes
   - Node failure
   - Scaling
   ```

3. **Isolation**
   ```
   Pod network: 10.32.0.0/13
   Node network: 10.0.0.0/16
   No conflicts!
   ```

### Virtual Service IPs Benefits

1. **Stability**
   ```
   Service IP: 10.96.10.50 (never changes)
   Backend pods: Can restart, scale, move
   Clients don't notice
   ```

2. **Load Balancing**
   ```
   One service IP → multiple backend pods
   eBPF distributes traffic automatically
   ```

3. **Decoupling**
   ```
   Clients → Service IP (stable)
   Backends → Pod IPs (ephemeral)
   ```

---

## Summary

### Overlay IPs (Pod CIDR: 10.32.0.0/13)

**What:** Virtual network on top of physical network
**How:** Geneve/VXLAN encapsulation (or native routing with eBPF)
**Why:** Avoid VNet IP exhaustion, enable mobility
**AKS:** Uses Geneve by default with Azure CNI Overlay

### Virtual IPs (Service CIDR: 10.96.0.0/12)

**What:** Load balancer IPs that don't exist on any interface
**How:** eBPF or iptables rewrite packets
**Why:** Stable frontend for ephemeral backends
**AKS with Cilium:** eBPF (much faster than iptables)

### Does AKS Use VXLAN?

**Answer:** No, with Azure CNI Overlay + Cilium, it uses **Geneve** by default.

- **Geneve**: Modern, extensible, better for cloud-native
- **VXLAN**: Available as alternative, more hardware offload support
- **Native**: Best performance, when possible (no encapsulation)

---

## Quick Reference

```bash
# Check encapsulation method
kubectl exec -n kube-system ds/cilium -- cilium status

# View service mappings
kubectl exec -n kube-system ds/cilium -- cilium service list

# Monitor overlay traffic
tcpdump -i eth0 'udp port 6081'  # Geneve
tcpdump -i eth0 'udp port 4789'  # VXLAN
```

**Bottom Line:** Overlay IPs use encapsulation (Geneve) to create a virtual network, while Service IPs use eBPF for lightning-fast load balancing without any real network interface! 🚀
