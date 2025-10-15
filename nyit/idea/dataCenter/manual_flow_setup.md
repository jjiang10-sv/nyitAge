# Manual Flow Setup Guide for BCube Topology

## Understanding BCube Architecture

In **BCube topology**, each host has **multiple network interfaces** connecting to different level switches. This is different from traditional tree topologies:

- Each host connects to **one level-0 switch** AND **multiple higher-level switches**
- Hosts can **relay traffic** between their interfaces (host-based routing)
- Switches only need flows for hosts **directly connected** to them

## The 5 Required Paths

### 1. RED Path: h00 ↔ h40 via s30
Both hosts connect directly to s30. **No intermediate switches needed.**

### 2. GREEN Path: h00 ↔ h50 via s30 → h40 → s14
h40 acts as a **relay host**, forwarding packets between its s30 and s14 interfaces.

### 3. BLUE Path: h20 ↔ h30 via s12
Both hosts connect directly to s12.

### 4. PURPLE Path: h60 ↔ h61 via s06
Both hosts connect directly to s06 (same level-0 switch).

### 5. BLACK Path: h60 ↔ h70 via s16
Both hosts connect directly to s16.

## Manual Setup Commands

### Step 1: Check Topology Connections

First, verify which hosts connect to which switches:

```bash
mininet> links
```

Look for connections like:
- h00-ethX ↔ s30-ethY (h00 connects to s30)
- h40-ethX ↔ s30-ethY (h40 also connects to s30)
- h40-ethZ ↔ s14-ethW (h40 also connects to s14)

### Step 2: Check Port Numbers

For each switch in your paths, identify port numbers:

```bash
mininet> sh ovs-ofctl show s30
mininet> sh ovs-ofctl show s14
mininet> sh ovs-ofctl show s12
mininet> sh ovs-ofctl show s06
mininet> sh ovs-ofctl show s16
```

Output shows port mappings like:
```
 1(s30-eth1): addr:... (connects to h00)
 2(s30-eth2): addr:... (connects to h40)
```

### Step 3: Get Host IP Addresses

```bash
mininet> h00 ifconfig
mininet> h40 ifconfig
mininet> h50 ifconfig
```

Typical IPs (Mininet default):
- h00 = 10.0.0.1
- h40 = 10.0.0.5  
- h50 = 10.0.0.6
- h20 = 10.0.0.3
- h30 = 10.0.0.4
- h60 = 10.0.0.7
- h61 = 10.0.0.8
- h70 = 10.0.0.9

### Step 4: Configure Each Path

#### RED Path (h00 ↔ h40 via s30)

```bash
# Clear existing flows
mininet> sh ovs-ofctl del-flows s30

# Allow ARP (essential!)
mininet> sh ovs-ofctl add-flow s30 'priority=200,arp,actions=normal'

# Default drop (security)
mininet> sh ovs-ofctl add-flow s30 'priority=0,actions=drop'

# Forward path: h00 → h40 (check port numbers with ovs-ofctl show s30)
mininet> sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.5,actions=output:PORT_TO_H40'

# Reverse path: h40 → h00
mininet> sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.5,nw_dst=10.0.0.1,actions=output:PORT_TO_H00'

# Test
mininet> h00 ping -c 3 h40
```

#### GREEN Path (h00 ↔ h50 via s30 → h40 → s14)

This path uses h40 as a relay, so we need to:
1. Enable IP forwarding on h40
2. Configure flows on s30 and s14
3. Add routing rules on h40

```bash
# Enable forwarding on h40
mininet> h40 sysctl -w net.ipv4.ip_forward=1

# Configure s30
mininet> sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.6,actions=output:PORT_TO_H40'
mininet> sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.6,nw_dst=10.0.0.1,actions=output:PORT_TO_H00'

# Configure s14
mininet> sh ovs-ofctl del-flows s14
mininet> sh ovs-ofctl add-flow s14 'priority=200,arp,actions=normal'
mininet> sh ovs-ofctl add-flow s14 'priority=0,actions=drop'
mininet> sh ovs-ofctl add-flow s14 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.6,actions=output:PORT_TO_H50'
mininet> sh ovs-ofctl add-flow s14 'priority=100,ip,nw_src=10.0.0.6,nw_dst=10.0.0.1,actions=output:PORT_TO_H40'

# Configure routing on h40 (find interface names first)
mininet> h40 ifconfig
# Look for two interfaces connecting to s30 and s14, e.g., h40-eth3 and h40-eth1

mininet> h40 ip route add 10.0.0.6 dev h40-ethX  # interface to s14
mininet> h40 ip route add 10.0.0.1 dev h40-ethY  # interface to s30

# Verify h40 routing table
mininet> h40 ip route

# Test
mininet> h00 ping -c 3 h50
```

#### BLUE Path (h20 ↔ h30 via s12)

```bash
mininet> sh ovs-ofctl del-flows s12
mininet> sh ovs-ofctl add-flow s12 'priority=200,arp,actions=normal'
mininet> sh ovs-ofctl add-flow s12 'priority=0,actions=drop'
mininet> sh ovs-ofctl add-flow s12 'priority=100,ip,nw_src=10.0.0.3,nw_dst=10.0.0.4,actions=output:PORT_TO_H30'
mininet> sh ovs-ofctl add-flow s12 'priority=100,ip,nw_src=10.0.0.4,nw_dst=10.0.0.3,actions=output:PORT_TO_H20'

# Test
mininet> h20 ping -c 3 h30
```

#### PURPLE Path (h60 ↔ h61 via s06)

```bash
mininet> sh ovs-ofctl del-flows s06
mininet> sh ovs-ofctl add-flow s06 'priority=200,arp,actions=normal'
mininet> sh ovs-ofctl add-flow s06 'priority=0,actions=drop'
mininet> sh ovs-ofctl add-flow s06 'priority=100,ip,nw_src=10.0.0.7,nw_dst=10.0.0.8,actions=output:PORT_TO_H61'
mininet> sh ovs-ofctl add-flow s06 'priority=100,ip,nw_src=10.0.0.8,nw_dst=10.0.0.7,actions=output:PORT_TO_H60'

# Test
mininet> h60 ping -c 3 h61
```

#### BLACK Path (h60 ↔ h70 via s16)

```bash
mininet> sh ovs-ofctl del-flows s16
mininet> sh ovs-ofctl add-flow s16 'priority=200,arp,actions=normal'
mininet> sh ovs-ofctl add-flow s16 'priority=0,actions=drop'
mininet> sh ovs-ofctl add-flow s16 'priority=100,ip,nw_src=10.0.0.7,nw_dst=10.0.0.9,actions=output:PORT_TO_H70'
mininet> sh ovs-ofctl add-flow s16 'priority=100,ip,nw_src=10.0.0.9,nw_dst=10.0.0.7,actions=output:PORT_TO_H60'

# Test
mininet> h60 ping -c 3 h70
```

## Complete Manual Setup Script

Here's the full sequence (replace PORT_TO_XXX with actual port numbers):

```bash
# === RED PATH ===
sh ovs-ofctl del-flows s30
sh ovs-ofctl add-flow s30 'priority=200,arp,actions=normal'
sh ovs-ofctl add-flow s30 'priority=0,actions=drop'
sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.5,actions=output:PORT_TO_H40'
sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.5,nw_dst=10.0.0.1,actions=output:PORT_TO_H00'

# === GREEN PATH ===
h40 sysctl -w net.ipv4.ip_forward=1
sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.6,actions=output:PORT_TO_H40'
sh ovs-ofctl add-flow s30 'priority=100,ip,nw_src=10.0.0.6,nw_dst=10.0.0.1,actions=output:PORT_TO_H00'

sh ovs-ofctl del-flows s14
sh ovs-ofctl add-flow s14 'priority=200,arp,actions=normal'
sh ovs-ofctl add-flow s14 'priority=0,actions=drop'
sh ovs-ofctl add-flow s14 'priority=100,ip,nw_src=10.0.0.1,nw_dst=10.0.0.6,actions=output:PORT_TO_H50'
sh ovs-ofctl add-flow s14 'priority=100,ip,nw_src=10.0.0.6,nw_dst=10.0.0.1,actions=output:PORT_TO_H40'

h40 ip route add 10.0.0.6 dev h40-ethX
h40 ip route add 10.0.0.1 dev h40-ethY

# === BLUE PATH ===
sh ovs-ofctl del-flows s12
sh ovs-ofctl add-flow s12 'priority=200,arp,actions=normal'
sh ovs-ofctl add-flow s12 'priority=0,actions=drop'
sh ovs-ofctl add-flow s12 'priority=100,ip,nw_src=10.0.0.3,nw_dst=10.0.0.4,actions=output:PORT_TO_H30'
sh ovs-ofctl add-flow s12 'priority=100,ip,nw_src=10.0.0.4,nw_dst=10.0.0.3,actions=output:PORT_TO_H20'

# === PURPLE PATH ===
sh ovs-ofctl del-flows s06
sh ovs-ofctl add-flow s06 'priority=200,arp,actions=normal'
sh ovs-ofctl add-flow s06 'priority=0,actions=drop'
sh ovs-ofctl add-flow s06 'priority=100,ip,nw_src=10.0.0.7,nw_dst=10.0.0.8,actions=output:PORT_TO_H61'
sh ovs-ofctl add-flow s06 'priority=100,ip,nw_src=10.0.0.8,nw_dst=10.0.0.7,actions=output:PORT_TO_H60'

# === BLACK PATH ===
sh ovs-ofctl del-flows s16
sh ovs-ofctl add-flow s16 'priority=200,arp,actions=normal'
sh ovs-ofctl add-flow s16 'priority=0,actions=drop'
sh ovs-ofctl add-flow s16 'priority=100,ip,nw_src=10.0.0.7,nw_dst=10.0.0.9,actions=output:PORT_TO_H70'
sh ovs-ofctl add-flow s16 'priority=100,ip,nw_src=10.0.0.9,nw_dst=10.0.0.7,actions=output:PORT_TO_H60'
```

## Verification Commands

```bash
# View flows on a switch
mininet> sh ovs-ofctl dump-flows s30

# Check if flows are being used (n_packets > 0)
mininet> sh ovs-ofctl dump-flows s30 | grep n_packets

# View host routing table
mininet> h40 ip route

# View host interfaces
mininet> h40 ifconfig

# Trace packet path
mininet> h00 traceroute h40
mininet> h00 traceroute h50

# Test connectivity
mininet> pingall  # Should show only 5 bidirectional paths work
```

## Common Issues and Solutions

### Issue 1: "Destination Host Unreachable"
**Cause:** Missing ARP flows  
**Solution:** Always add `priority=200,arp,actions=normal` on each switch

### Issue 2: GREEN Path Not Working (h00 → h50)
**Cause:** h40 forwarding disabled or routing not configured  
**Solutions:**
```bash
h40 sysctl -w net.ipv4.ip_forward=1
h40 ip route add 10.0.0.6 dev h40-ethX
h40 ip route add 10.0.0.1 dev h40-ethY
```

### Issue 3: Wrong Port Numbers
**Cause:** Using incorrect port numbers in flow rules  
**Solution:** Always check with `sh ovs-ofctl show <switch>` first

### Issue 4: One-Way Communication Only
**Cause:** Missing reverse flows  
**Solution:** Always add flows in **both directions** (src→dst AND dst→src)

### Issue 5: Other Hosts Can Communicate
**Cause:** Missing default drop rule  
**Solution:** Add `priority=0,actions=drop` on all switches

## Key Concepts

### BCube Topology Features:
1. **Multi-homed hosts**: Each host has multiple interfaces
2. **Host-based routing**: Hosts can relay traffic (like h40 in green path)
3. **Direct connections**: Hosts connect directly to multiple-level switches
4. **No traditional routing**: Switches use OpenFlow, not IP routing

### Flow Priority Rules:
- Priority 200: ARP (highest, allow all)
- Priority 100: Specific IP paths
- Priority 0: Default drop (lowest, catch-all)

### Required for Each Path:
1. ✓ ARP flows on all switches
2. ✓ Forward flow (src → dst)
3. ✓ Reverse flow (dst → src)
4. ✓ Correct port numbers
5. ✓ Host forwarding enabled (for relay hosts like h40)
6. ✓ Host routing rules (for relay hosts)

## Automated vs Manual

**Use automated script (recommended):**
```bash
sudo python3 nyit/idea/dataCenter/assignment1.py
```

The script automatically:
- Discovers port numbers
- Configures all switches
- Enables host forwarding
- Sets up routing on h40

**Use manual commands when:**
- Learning OpenFlow concepts
- Debugging specific paths
- Testing incremental changes
- Understanding BCube architecture

## Summary

The key difference in BCube topology is that **hosts connect directly to higher-level switches**, not through a hierarchy of level-0, level-1, etc. This means:
- **No intermediate switches** for most paths
- **Hosts can relay** traffic between their interfaces
- **Only configure switches** mentioned in the path specification
- **Enable IP forwarding** on relay hosts

For the green path specifically, h40 acts as a relay between s30 and s14, requiring both OpenFlow rules on switches AND routing configuration on h40 itself.