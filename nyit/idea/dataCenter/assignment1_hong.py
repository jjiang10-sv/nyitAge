#!/usr/bin/env python3
"""
BCube(2,3) with Bridge-Based Host Configuration
================================================
- 16 hosts (h00-h71) with BCube connectivity
- Each host uses internal br0 bridge (single IP: 10.0.0.X/24)
- Port-based OpenFlow rules for h00 <-> h40 path via s30
- autoStaticArp for immediate connectivity
- Automatic iperf support!
"""

from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import OVSSwitch
from mininet.link import TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel, info
import os

class SimpleBCube(Topo):
    def build(self):
        n = 2          # ports per switch (base)
        k = 3          # k (levels = k+1 = 4)
        bw = 8
        delay = '4ms'

        levels = k + 1
        num_switches_per_level = n ** k  # 8 switches per level

        # --- Create switches ---
        switches = []
        for level in range(levels):
            level_switches = []
            for i in range(num_switches_per_level):
                sw = self.addSwitch(f"s{level}{i}")
                level_switches.append(sw)
            switches.append(level_switches)

        # --- Create hosts (no IP assigned here - will be set on br0) ---
        hosts = []
        for i in range(num_switches_per_level):
            for j in range(n):
                name = f"h{i}{j}"
                host = self.addHost(name)  # No IP yet - will assign to br0
                hosts.append(host)

        # --- Connect hosts to switches ---
        for i in range(num_switches_per_level):
            for j in range(n):
                h = hosts[i*n+j]

                sw0 = switches[0][i]
                sw1 = switches[1][(i//2)*2 + j]
                sw2 = switches[2][(i//4)*4 + (i%2)*2 + j]
                sw3 = switches[3][(i%4)*2 + j]

                self.addLink(h, sw0, bw=bw, delay=delay)
                self.addLink(h, sw1, bw=bw, delay=delay)
                self.addLink(h, sw2, bw=bw, delay=delay)
                self.addLink(h, sw3, bw=bw, delay=delay)


def setup_host_bridges(net):
    """
    Create internal bridges on each host (KEY FIX for iperf!)
    - Merge all interfaces into br0
    - Assign single IP to br0 (10.0.0.X/24)
    """
    info('*** Setting up internal bridges on each host\n')
    base_ip = 1  # Start from 10.0.0.1
    
    for idx, h in enumerate(net.hosts, start=base_ip):
        name = h.name
        info(f' Configuring {name}...')
        
        # Create bridge
        h.cmd("ip link add br0 type bridge")
        h.cmd("ip link set br0 up")
        
        # Move all host interfaces under the bridge
        for intf in h.intfList():
            if intf.name != 'lo':
                h.cmd(f"ip addr flush dev {intf.name}")
                h.cmd(f"ip link set {intf.name} master br0")
                h.cmd(f"ip link set {intf.name} up")
        
        # Assign single IP to bridge
        ip_addr = f"10.0.0.{idx}/24"
        h.cmd(f"ip addr add {ip_addr} dev br0")
        info(f" br0 -> {ip_addr}\n")
    
    # Update Mininet's IP() method to return br0's IP
    for h in net.hosts:
        ip = h.cmd("ip -4 addr show br0 | grep 'inet ' | awk '{print $2}' | cut -d/ -f1").strip()
        if ip:
            h.IP = lambda ip=ip: ip
        else:
            h.IP = lambda: None
    
    info('\n*** All hosts configured with br0 bridge\n')
    info(' Each host has single IP on br0\n')
    info(' Physical interfaces are bridge members (Layer 2)\n\n')


def add_flows_h00_h40(net):
    """
    Add OpenFlow rules for h00 <-> h40 path via s30 (Level 3, Switch 0)
    Uses simple port-based forwarding.
    """
    info('*** Configuring OpenFlow rules for h00 <-> h40 path\n')
    
    h00 = net.get('h00')
    h40 = net.get('h40')
    s30 = net.get('s30')  # Level 3, switch 0 connects h00 and h40
    
    # Clear existing flows
    os.system(f"ovs-ofctl del-flows {s30.name}")
    
    # Add ARP flooding for dynamic address resolution
    os.system(f"ovs-ofctl add-flow {s30.name} 'priority=50,idle_timeout=0,arp,actions=FLOOD'")
    
    # Get port numbers
    h00_conn = s30.connectionsTo(h00)
    h40_conn = s30.connectionsTo(h40)
    
    if h00_conn and h40_conn:
        # p_h00 = s30.ports[h00_conn[0][0]]
        # p_h40 = s30.ports[h40_conn[0][0]]
        p_h00 = 1
        p_h40 = 2
        
        # Simple port-to-port forwarding (works for ALL protocols)
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,idle_timeout=0,in_port={p_h00},actions=output:{p_h40}'")
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,idle_timeout=0,in_port={p_h40},actions=output:{p_h00}'")
        info(f'  s30: port {p_h00} (h00) <-> port {p_h40} (h40)\n')
        info('  Port-based forwarding configured\n\n')
    else:
        info('  ERROR: Could not find connections for h00 and h40 on s30\n\n')


def run():
    setLogLevel('info')
    info('*** Building BCube(2,3) with Bridge-Based Configuration\n')
    
    topo = SimpleBCube()
    net = Mininet(
        topo=topo,
        switch=OVSSwitch,
        link=TCLink,
        controller=None,
        autoSetMacs=True,      # Auto-assign MAC addresses
        autoStaticArp=True     # Pre-populate ARP tables (KEY!)
    )
    net.start()
    
    # Setup bridges on all hosts
    setup_host_bridges(net)
    
    # Add OpenFlow rules for h00 <-> h40
    add_flows_h00_h40(net)
    
    info('=' * 70 + '\n')
    info('*** BCube(2,3) Ready with Bridge Configuration\n')
    info('=' * 70 + '\n')
    info('Host IPs (on br0):\n')
    for h in net.hosts:
        info(f'  {h.name}: {h.IP()}\n')
    
    info('\nTest h00 <-> h40 path:\n')
    info('  Manual iperf (RECOMMENDED):\n')
    info('    mininet> h40 iperf -s -p 5001 &\n')
    info('    mininet> h00 iperf -c 10.0.0.9 -t 10\n')
    info('    mininet> sh killall iperf\n\n')
    info('  Or try automatic:\n')
    info('    mininet> iperf h00 h40\n\n')
    info('  Ping test:\n')
    info('    mininet> h00 ping -c 3 10.0.0.9\n')
    info('=' * 70 + '\n\n')

    CLI(net)
    net.stop()


if __name__ == '__main__':
    run()
