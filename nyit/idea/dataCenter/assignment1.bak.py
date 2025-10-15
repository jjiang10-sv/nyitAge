#!/usr/bin/env python3
"""
Custom_BCube_Topo.py
BCube(3,2): 8 mini-cubes, 2 hosts each, 4 switch levels (0–3)
Link bandwidth = 8 Mbps, delay = 4 ms
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
            counter = 0
            for cube in range(2 ** k):
                sw = f's{lvl}{cube}'
                counter +=1
                for h in range(n):
                    idx0 = cube % 2
                    idx1 = (cube // (2 ** lvl) ) * (2**lvl)
                    if counter == 2 and lvl > 1:
                        idx1 += 1
                        counter = 0
                    idx1 += h%2 * (2**(lvl-1))
                    host = f'h{idx1}{idx0}'
                    self.addLink(host, sw, bw=bw, delay=delay)


def configure_host_routing(net):
    """
    Configure host routing for BCube paths.
    """
    print("\n=== Configuring host routing ===")
    
    # Get all hosts
    h00 = net.get('h00'); h40 = net.get('h40'); h50 = net.get('h50')
    h20 = net.get('h20'); h30 = net.get('h30')
    h60 = net.get('h60'); h61 = net.get('h61'); h70 = net.get('h70')
    
    # Get switches
    s30 = net.get('s30'); s14 = net.get('s14'); s12 = net.get('s12')
    s06 = net.get('s06'); s16 = net.get('s16')
    
    # RED PATH: h00 ↔ h40 via s30
    h00_s30_intf = None
    h00_s30_mac = None
    for intf in h00.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s30 or intf.link.intf2.node == s30:
                h00_s30_intf = intf.name
                h00_s30_mac = intf.MAC()
                break
    
    h40_s30_intf = None
    h40_s30_mac = None
    for intf in h40.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s30 or intf.link.intf2.node == s30:
                h40_s30_intf = intf.name
                h40_s30_mac = intf.MAC()
                break
    
    if h00_s30_intf and h40_s30_intf:
        h00.cmd(f'ip route add {h40.IP()} dev {h00_s30_intf}')
        h40.cmd(f'ip route add {h00.IP()} dev {h40_s30_intf}')
        # Add static ARP entries
        h00.cmd(f'arp -s {h40.IP()} {h40_s30_mac}')
        h40.cmd(f'arp -s {h00.IP()} {h00_s30_mac}')
        print(f"✓ RED: h00→{h00_s30_intf}({h00_s30_mac})→s30→{h40_s30_intf}({h40_s30_mac})→h40")
    
    # GREEN PATH: h00 ↔ h50 (requires h40 as relay)
    # Enable IP forwarding on h40
    h40.cmd('sysctl -w net.ipv4.ip_forward=1')
    
    # Disable reverse path filtering on h40
    h40.cmd('sysctl -w net.ipv4.conf.all.rp_filter=0')
    h40.cmd('sysctl -w net.ipv4.conf.default.rp_filter=0')
    for intf in h40.intfList():
        if intf.name != 'lo':
            h40.cmd(f'sysctl -w net.ipv4.conf.{intf.name}.rp_filter=0')
    
    # Find h40's interface to s14
    h40_s14_intf = None
    h40_s14_mac = None
    for intf in h40.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s14 or intf.link.intf2.node == s14:
                h40_s14_intf = intf.name
                h40_s14_mac = intf.MAC()
                break
    
    # Find h50's interface to s14
    h50_s14_intf = None
    h50_s14_mac = None
    for intf in h50.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s14 or intf.link.intf2.node == s14:
                h50_s14_intf = intf.name
                h50_s14_mac = intf.MAC()
                break
    
    # Configure routes for green path
    if h00_s30_intf and h00_s30_mac:
        # h00: route to h50 via s30 interface, use h40's MAC
        h00.cmd(f'ip route add {h50.IP()} dev {h00_s30_intf}')
        h00.cmd(f'arp -s {h50.IP()} {h40_s30_mac}')  # Critical: h00 thinks h50 is at h40's MAC
    
    if h50_s14_intf and h50_s14_mac:
        # h50: route to h00 via s14 interface, use h40's MAC
        h50.cmd(f'ip route add {h00.IP()} dev {h50_s14_intf}')
        h50.cmd(f'arp -s {h00.IP()} {h40_s14_mac}')  # Critical: h50 thinks h00 is at h40's MAC
    
    if h40_s30_intf and h40_s14_intf:
        # h40: specific routes for forwarding
        h40.cmd(f'ip route add {h50.IP()} dev {h40_s14_intf}')
        h40.cmd(f'ip route add {h00.IP()} dev {h40_s30_intf}')
        
        # h40 needs to know the real MACs
        if h00_s30_mac:
            h40.cmd(f'arp -s {h00.IP()} {h00_s30_mac}')
        if h50_s14_mac:
            h40.cmd(f'arp -s {h50.IP()} {h50_s14_mac}')
        
        print(f"✓ GREEN: h00→{h00_s30_intf}→s30→{h40_s30_intf}→h40→{h40_s14_intf}→s14→{h50_s14_intf}→h50")
        print(f"  Static ARP: h00 maps h50→{h40_s30_mac}, h50 maps h00→{h40_s14_mac}")
    
    # BLUE PATH: h20 ↔ h30 via s12
    h20_s12_intf = None
    h20_s12_mac = None
    for intf in h20.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s12 or intf.link.intf2.node == s12:
                h20_s12_intf = intf.name
                h20_s12_mac = intf.MAC()
                break
    
    h30_s12_intf = None
    h30_s12_mac = None
    for intf in h30.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s12 or intf.link.intf2.node == s12:
                h30_s12_intf = intf.name
                h30_s12_mac = intf.MAC()
                break
    
    if h20_s12_intf and h30_s12_intf:
        h20.cmd(f'ip route add {h30.IP()} dev {h20_s12_intf}')
        h30.cmd(f'ip route add {h20.IP()} dev {h30_s12_intf}')
        h20.cmd(f'arp -s {h30.IP()} {h30_s12_mac}')
        h30.cmd(f'arp -s {h20.IP()} {h20_s12_mac}')
        print(f"✓ BLUE: h20→{h20_s12_intf}→s12→{h30_s12_intf}→h30")
    
    # PURPLE PATH: h60 ↔ h61 via s06 (already on level-0)
    print(f"✓ PURPLE: h60 ↔ h61 (level-0 switch s06)")
    
    # BLACK PATH: h60 ↔ h70 via s16
    h60_s16_intf = None
    h60_s16_mac = None
    for intf in h60.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s16 or intf.link.intf2.node == s16:
                h60_s16_intf = intf.name
                h60_s16_mac = intf.MAC()
                break
    
    h70_s16_intf = None
    h70_s16_mac = None
    for intf in h70.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s16 or intf.link.intf2.node == s16:
                h70_s16_intf = intf.name
                h70_s16_mac = intf.MAC()
                break
    
    if h60_s16_intf and h70_s16_intf:
        h60.cmd(f'ip route add {h70.IP()} dev {h60_s16_intf}')
        h70.cmd(f'ip route add {h60.IP()} dev {h70_s16_intf}')
        h60.cmd(f'arp -s {h70.IP()} {h70_s16_mac}')
        h70.cmd(f'arp -s {h60.IP()} {h60_s16_mac}')
        print(f"✓ BLACK: h60→{h60_s16_intf}→s16→{h70_s16_intf}→h70")
    
    print()


def add_flows_bcube(net):
    """
    Add OpenFlow rules for the 5 specific BCube paths.
    """
    # --- Get hosts ---
    h00 = net.get('h00'); h40 = net.get('h40'); h50 = net.get('h50')
    h20 = net.get('h20'); h30 = net.get('h30')
    h60 = net.get('h60'); h61 = net.get('h61'); h70 = net.get('h70')

    # --- Get switches ---
    s30 = net.get('s30'); s14 = net.get('s14'); s12 = net.get('s12')
    s06 = net.get('s06'); s16 = net.get('s16')

    path_switches = [s30, s14, s12, s06, s16]

    print("=== Configuring OpenFlow rules ===")
    
    # Clear all flows and set default drop
    for sw in path_switches:
        os.system(f"ovs-ofctl del-flows {sw.name}")
        os.system(f"ovs-ofctl add-flow {sw.name} 'priority=0,actions=drop'")
        # Allow ARP (critical for communication)
        os.system(f"ovs-ofctl add-flow {sw.name} 'priority=200,arp,actions=normal'")
    
    print(f"\n[INFO] Host IPs: h00={h00.IP()}, h40={h40.IP()}, h50={h50.IP()}")
    print(f"       h20={h20.IP()}, h30={h30.IP()}, h60={h60.IP()}, h61={h61.IP()}, h70={h70.IP()}")

    # ========== RED PATH: h00 ↔ h40 via s30 ==========
    print(f"\n[RED] h00({h00.IP()}) ↔ h40({h40.IP()}) via s30")
    h00_conn = s30.connectionsTo(h00)
    h40_conn = s30.connectionsTo(h40)
    
    if h00_conn and h40_conn:
        port_h00 = s30.ports[h00_conn[0][0]]
        port_h40 = s30.ports[h40_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h40.IP()},actions=output:{port_h40}'")
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h40.IP()},nw_dst={h00.IP()},actions=output:{port_h00}'")
        print(f"  ✓ s30: port {port_h00}(h00) ↔ port {port_h40}(h40)")

    # ========== GREEN PATH: h00 ↔ h50 via s30 → h40 → s14 ==========
    print(f"\n[GREEN] h00({h00.IP()}) ↔ h50({h50.IP()}) via s30→h40→s14")
    # s30: forward packets between h00 and h40
    os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h50.IP()},actions=output:{port_h40}'")
    os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h50.IP()},nw_dst={h00.IP()},actions=output:{port_h00}'")
    print(f"  ✓ s30: forwards h00↔h50 traffic via h40")
    
    # s14: forward packets between h40 and h50
    h40_s14_conn = s14.connectionsTo(h40)
    h50_s14_conn = s14.connectionsTo(h50)
    if h40_s14_conn and h50_s14_conn:
        port_s14_h40 = s14.ports[h40_s14_conn[0][0]]
        port_s14_h50 = s14.ports[h50_s14_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s14.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h50.IP()},actions=output:{port_s14_h50}'")
        os.system(f"ovs-ofctl add-flow {s14.name} 'priority=100,ip,nw_src={h50.IP()},nw_dst={h00.IP()},actions=output:{port_s14_h40}'")
        print(f"  ✓ s14: port {port_s14_h40}(h40) ↔ port {port_s14_h50}(h50)")

    # ========== BLUE PATH: h20 ↔ h30 via s12 ==========
    print(f"\n[BLUE] h20({h20.IP()}) ↔ h30({h30.IP()}) via s12")
    h20_conn = s12.connectionsTo(h20)
    h30_conn = s12.connectionsTo(h30)
    if h20_conn and h30_conn:
        port_h20 = s12.ports[h20_conn[0][0]]
        port_h30 = s12.ports[h30_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s12.name} 'priority=100,ip,nw_src={h20.IP()},nw_dst={h30.IP()},actions=output:{port_h30}'")
        os.system(f"ovs-ofctl add-flow {s12.name} 'priority=100,ip,nw_src={h30.IP()},nw_dst={h20.IP()},actions=output:{port_h20}'")
        print(f"  ✓ s12: port {port_h20}(h20) ↔ port {port_h30}(h30)")

    # ========== PURPLE PATH: h60 ↔ h61 via s06 ==========
    print(f"\n[PURPLE] h60({h60.IP()}) ↔ h61({h61.IP()}) via s06")
    h60_conn = s06.connectionsTo(h60)
    h61_conn = s06.connectionsTo(h61)
    if h60_conn and h61_conn:
        port_h60 = s06.ports[h60_conn[0][0]]
        port_h61 = s06.ports[h61_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s06.name} 'priority=100,ip,nw_src={h60.IP()},nw_dst={h61.IP()},actions=output:{port_h61}'")
        os.system(f"ovs-ofctl add-flow {s06.name} 'priority=100,ip,nw_src={h61.IP()},nw_dst={h60.IP()},actions=output:{port_h60}'")
        print(f"  ✓ s06: port {port_h60}(h60) ↔ port {port_h61}(h61)")

    # ========== BLACK PATH: h60 ↔ h70 via s16 ==========
    print(f"\n[BLACK] h60({h60.IP()}) ↔ h70({h70.IP()}) via s16")
    h60_s16_conn = s16.connectionsTo(h60)
    h70_s16_conn = s16.connectionsTo(h70)
    if h60_s16_conn and h70_s16_conn:
        port_h60 = s16.ports[h60_s16_conn[0][0]]
        port_h70 = s16.ports[h70_s16_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s16.name} 'priority=100,ip,nw_src={h60.IP()},nw_dst={h70.IP()},actions=output:{port_h70}'")
        os.system(f"ovs-ofctl add-flow {s16.name} 'priority=100,ip,nw_src={h70.IP()},nw_dst={h60.IP()},actions=output:{port_h60}'")
        print(f"  ✓ s16: port {port_h60}(h60) ↔ port {port_h70}(h70)")

    print("\n✅ All flows configured!\n")


def run():
    topo = BCube32()
    net = Mininet(topo=topo, switch=OVSSwitch, link=TCLink, controller=None)
    net.start()
    info("\n*** BCube(3,2) topology built ***\n")
    
    # Configure host routing FIRST
    configure_host_routing(net)
    
    # Then add OpenFlow rules
    add_flows_bcube(net)
    
    info("\n*** Test the 5 paths ***\n")
    info("=" * 60 + "\n")
    info("  h00 ping -c 3 h40  # RED path via s30\n")
    info("  h00 ping -c 3 h50  # GREEN path via s30→h40→s14\n")
    info("  h20 ping -c 3 h30  # BLUE path via s12\n")
    info("  h60 ping -c 3 h61  # PURPLE path via s06\n")
    info("  h60 ping -c 3 h70  # BLACK path via s16\n")
    info("=" * 60 + "\n")
    info("\nDebug:\n")
    info("  h00 arp -a            # Check ARP cache\n")
    info("  h40 arp -a            # Check ARP cache\n")
    info("  sh ovs-ofctl dump-flows s30 | grep n_packets\n")
    info("  sh ovs-ofctl dump-flows s14 | grep n_packets\n\n")

    CLI(net)
    net.stop()

if __name__ == '__main__':
    setLogLevel('info')
    run()