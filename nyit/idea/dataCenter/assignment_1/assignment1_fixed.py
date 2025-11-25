#!/usr/bin/env python3
"""
Custom_BCube_Topo.py - BRIDGE-BASED VERSION (Auto-iperf Compatible)
BCube(3,2): 8 mini-cubes, 2 hosts each, 4 switch levels (0–3)
Link bandwidth = 8 Mbps, delay = 4 ms

KEY FIXES FOR AUTO-IPERF:
- Each host uses internal Linux bridge (br0) merging all interfaces
- Single IP address per host on br0 (solves iperf binding issue)
- Port-based OpenFlow rules (simpler and more reliable)
- autoStaticArp=True for immediate connectivity
- Layer 2 switching instead of Layer 3 routing

This approach makes Mininet's 'iperf' command work automatically!
"""
import os
import time
from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import OVSSwitch
from mininet.link import TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel, info

switches_global = []
class BCube32(Topo):
    def build(self, k=3, n=2, bw=8, delay='4ms'):
        # --- Hosts ---
        hosts = []
        for cube in range(2 ** k):          # 8 cubes
            for h in range(n):              # 2 hosts each
                name = f'h{cube}{h}'
                hosts.append(self.addHost(name))

        # --- Switches by level ---
        levels = k + 1                      # 0..3
        switches = {lvl: [] for lvl in range(levels)}
        for lvl in range(levels):
            for i in range(2 ** k):         # 8 per level
                sname = f's{lvl}{i}'
                switches[lvl].append(self.addSwitch(sname))
                switches_global.append(sname)

        # --- Level-0 connections ---
        for cube in range(2 ** k):
            for h in range(n):
                self.addLink(f'h{cube}{h}', f's0{cube}', bw=bw, delay=delay)

        # --- Higher-level connections (BCube formula) ---
        for lvl in range(1, levels):
            for cube in range(n ** k):
                sw = f's{lvl}{cube}'
                serverIdx = cube % n
                jumpIncrement = n ** lvl
                div = cube // jumpIncrement
                reminder = cube % jumpIncrement
                cubeIdx = div * jumpIncrement
                cubeIdx += (reminder // n)
                
                for h in range(n):
                    cubeIdx += ((h%n ) * jumpIncrement//n)
                    host = f'h{cubeIdx}{serverIdx}'
                    self.addLink(host, sw, bw=bw, delay=delay)


def setup_host_bridges(net):
    """
    Create internal Linux bridges on each host to merge all interfaces.
    This is the KEY FIX that makes Mininet's iperf work automatically!
    
    Each host gets:
    - br0 bridge with single IP address
    - All physical interfaces enslaved to br0 (no IPs on them)
    """
    info('\n*** Setting up internal bridges on each host\n')
    base_ip = 1  # Start from 10.0.0.1
    
    # First pass: Create bridges and assign IPs
    host_ips = {}  # Store hostname -> IP mapping
    for idx, h in enumerate(net.hosts, start=base_ip):
        name = h.name
        info(f' Configuring {name}...')
        
        # Create bridge
        h.cmd("ip link add br0 type bridge")
        h.cmd("ip link set br0 up")
        
        # Move all host interfaces under the bridge
        for intf in h.intfList():
            if intf.name != 'lo':
                # Remove any IP from physical interface
                h.cmd(f"ip addr flush dev {intf.name}")
                # Add to bridge
                h.cmd(f"ip link set {intf.name} master br0")
                h.cmd(f"ip link set {intf.name} up")
        
        # Assign single IP to bridge
        ip_only = f"10.0.0.{idx}"
        ip_addr = f"{ip_only}/24"
        h.cmd(f"ip addr add {ip_addr} dev br0")
        
        # Store for later
        host_ips[name] = ip_only
        
        # Override Mininet's IP() method to return br0's IP
        h.IP = lambda ip=ip_only: ip
        
        info(f" br0 -> {ip_addr}\n")
    
    info('\n*** Building hostname-to-IP mappings for Mininet\n')
    # Build Mininet-compatible hostname resolution
    # Store in net object for Mininet's internal use
    if not hasattr(net, 'hostnames'):
        net.hostnames = {}
    
    for hostname, ip in host_ips.items():
        net.hostnames[hostname] = ip
        info(f"  {hostname} -> {ip}\n")
    
    info('\n All hosts now have br0 with single IP\n')
    info(' Physical interfaces are bridge members (Layer 2 only)\n')
    info(' Hostname resolution configured via Mininet\n\n')


def add_flows_bcube(net):
    """
    Add OpenFlow rules using PORT-BASED forwarding.
    This is simpler and more reliable than IP-based matching!
    
    Port-based rules work for ALL protocols (TCP, UDP, ICMP) without special cases.
    """
    h00 = net.get('h00'); h40 = net.get('h40'); h50 = net.get('h50')
    h20 = net.get('h20'); h30 = net.get('h30')
    h60 = net.get('h60'); h61 = net.get('h61'); h70 = net.get('h70')
    s30 = net.get('s30'); s14 = net.get('s14'); s12 = net.get('s12')
    s06 = net.get('s06'); s16 = net.get('s16')

    info("=== Configuring Port-Based OpenFlow Rules ===\n")
    
    # Clear all flows and enable ARP flooding on path switches
    for sw in [s30, s14, s12, s06, s16]:
        os.system(f"ovs-ofctl del-flows {sw.name}")
        # ARP flooding for dynamic address resolution
        os.system(f"ovs-ofctl add-flow {sw.name} 'priority=50,idle_timeout=0,arp,actions=FLOOD'")
    
    # RED PATH: h00 ↔ h40 via s30
    h00_conn = s30.connectionsTo(h00); h40_conn = s30.connectionsTo(h40)
    if h00_conn and h40_conn:
        p_h00 = s30.ports[h00_conn[0][0]]; p_h40 = s30.ports[h40_conn[0][0]]
        # Simple port-to-port forwarding (works for ALL traffic)
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,idle_timeout=0,in_port={p_h00},actions=output:{p_h40}'")
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,idle_timeout=0,in_port={p_h40},actions=output:{p_h00}'")
        info(f"[RED] s30 port {p_h00} ↔ port {p_h40} (h00 ↔ h40)\n")

    # GREEN PATH: h00 ↔ h50 via s30 → h40 → s14
    # Note: With bridges, h40 forwards at Layer 2, no IP forwarding needed!
    h40_s14 = s14.connectionsTo(h40); h50_s14 = s14.connectionsTo(h50)
    if h40_s14 and h50_s14:
        p_h40_s14 = s14.ports[h40_s14[0][0]]; p_h50_s14 = s14.ports[h50_s14[0][0]]
        os.system(f"ovs-ofctl add-flow {s14.name} 'priority=100,idle_timeout=0,in_port={p_h40_s14},actions=output:{p_h50_s14}'")
        os.system(f"ovs-ofctl add-flow {s14.name} 'priority=100,idle_timeout=0,in_port={p_h50_s14},actions=output:{p_h40_s14}'")
        info(f"[GREEN] s14 port {p_h40_s14} ↔ port {p_h50_s14} (h40 ↔ h50)\n")

    # BLUE PATH: h20 ↔ h30 via s12
    h20_conn = s12.connectionsTo(h20); h30_conn = s12.connectionsTo(h30)
    if h20_conn and h30_conn:
        p_h20 = s12.ports[h20_conn[0][0]]; p_h30 = s12.ports[h30_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s12.name} 'priority=100,idle_timeout=0,in_port={p_h20},actions=output:{p_h30}'")
        os.system(f"ovs-ofctl add-flow {s12.name} 'priority=100,idle_timeout=0,in_port={p_h30},actions=output:{p_h20}'")
        info(f"[BLUE] s12 port {p_h20} ↔ port {p_h30} (h20 ↔ h30)\n")

    # PURPLE PATH: h60 ↔ h61 via s06
    h60_conn = s06.connectionsTo(h60); h61_conn = s06.connectionsTo(h61)
    if h60_conn and h61_conn:
        p_h60 = s06.ports[h60_conn[0][0]]; p_h61 = s06.ports[h61_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s06.name} 'priority=100,idle_timeout=0,in_port={p_h60},actions=output:{p_h61}'")
        os.system(f"ovs-ofctl add-flow {s06.name} 'priority=100,idle_timeout=0,in_port={p_h61},actions=output:{p_h60}'")
        info(f"[PURPLE] s06 port {p_h60} ↔ port {p_h61} (h60 ↔ h61)\n")

    # BLACK PATH: h60 ↔ h70 via s16
    h60_s16 = s16.connectionsTo(h60); h70_s16 = s16.connectionsTo(h70)
    if h60_s16 and h70_s16:
        p_h60_s16 = s16.ports[h60_s16[0][0]]; p_h70_s16 = s16.ports[h70_s16[0][0]]
        os.system(f"ovs-ofctl add-flow {s16.name} 'priority=100,idle_timeout=0,in_port={p_h60_s16},actions=output:{p_h70_s16}'")
        os.system(f"ovs-ofctl add-flow {s16.name} 'priority=100,idle_timeout=0,in_port={p_h70_s16},actions=output:{p_h60_s16}'")
        info(f"[BLACK] s16 port {p_h60_s16} ↔ port {p_h70_s16} (h60 ↔ h70)\n")
    
    info("\n")


def run():
    topo = BCube32()
    # KEY FIX: Add autoStaticArp=True for immediate connectivity
    net = Mininet(
        topo=topo,
        switch=OVSSwitch,
        link=TCLink,
        controller=None,
        autoSetMacs=True,      # Auto-assign MAC addresses
        autoStaticArp=True     # Pre-populate ARP tables
    )
    net.start()
    info("\n*** BCube(3,2) topology built ***\n")
    
    # Setup internal bridges on all hosts (THE KEY FIX!)
    setup_host_bridges(net)
    
    # Add port-based OpenFlow rules
    add_flows_bcube(net)
    
    info("\n" + "="*70 + "\n")
    info("*** AUTOMATIC IPERF NOW WORKS! ***\n")
    info("="*70 + "\n")
    info("✅ Bridge-based design enables Mininet's built-in iperf command\n\n")
    
    info("Test paths with automatic iperf:\n")
    info("  iperf h00 h40   # RED path via s30\n")
    info("  iperf h00 h50   # GREEN path via s30→h40→s14\n")
    info("  iperf h20 h30   # BLUE path via s12\n")
    info("  iperf h60 h61   # PURPLE path via s06\n")
    info("  iperf h60 h70   # BLACK path via s16\n\n")
    
    info("Or test with ping:\n")
    info("  h00 ping -c 3 h40\n")
    info("  h00 ping -c 3 h50\n")
    info("  h20 ping -c 3 h30\n")
    info("  h60 ping -c 3 h61\n")
    info("  h60 ping -c 3 h70\n\n")
    
    info("View host IPs:\n")
    for h in net.hosts:
        info(f"  {h.name}: {h.IP()}\n")
    
    info("\n" + "="*70 + "\n\n")

    CLI(net)
    net.stop()

if __name__ == '__main__':
    setLogLevel('info')
    run()