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
            # counter = 0
            # prev_idx = 0
            # jump = False
            # old_stage = 0
            for cube in range(n ** k):
                sw = f's{lvl}{cube}'
                #counter +=1
                # goes to the server in each cube
                serverIdx = cube % n
                # the increment in jumping
                jumpIncrement = n ** lvl
                # how many jumps to made
                div = cube // jumpIncrement
                # to adjust within the jump increment
                reminder = cube % jumpIncrement
                # go to the correct jump span
                cubeIdx = div * jumpIncrement
                # finetune within the jump span
                cubeIdx += (reminder // n)
                # if idx1 > old_stage:
                #     jump = True
                #     old_stage = idx1
                # else:
                #     jump = False
                
                for h in range(n):
                    # increment within the n
                    cubeIdx += ((h%n ) * jumpIncrement//n)
                    host = f'h{cubeIdx}{serverIdx}'
                    self.addLink(host, sw, bw=bw, delay=delay)

    
    # Layer 0
        # add_pair_links(self, hosts['h00'], hosts['h01'], s0[0])
        # add_pair_links(self, hosts['h10'], hosts['h11'], s0[1])
        # add_pair_links(self, hosts['h20'], hosts['h21'], s0[2])
        # add_pair_links(self, hosts['h30'], hosts['h31'], s0[3])
        # add_pair_links(self, hosts['h40'], hosts['h41'], s0[4])
        # add_pair_links(self, hosts['h50'], hosts['h51'], s0[5])
        # add_pair_links(self, hosts['h60'], hosts['h61'], s0[6])
        # add_pair_links(self, hosts['h70'], hosts['h71'], s0[7])

        # # Layer 1
        # add_pair_links(self, hosts['h00'], hosts['h10'], s1[0])
        # add_pair_links(self, hosts['h01'], hosts['h11'], s1[1])
        # add_pair_links(self, hosts['h20'], hosts['h30'], s1[2])
        # add_pair_links(self, hosts['h21'], hosts['h31'], s1[3])
        # add_pair_links(self, hosts['h40'], hosts['h50'], s1[4])
        # add_pair_links(self, hosts['h41'], hosts['h51'], s1[5])
        # add_pair_links(self, hosts['h60'], hosts['h70'], s1[6])
        # add_pair_links(self, hosts['h61'], hosts['h71'], s1[7])

        # # Layer 2
        # add_pair_links(self, hosts['h00'], hosts['h20'], s2[0])
        # add_pair_links(self, hosts['h01'], hosts['h21'], s2[1])
        # add_pair_links(self, hosts['h10'], hosts['h30'], s2[2])
        # add_pair_links(self, hosts['h11'], hosts['h31'], s2[3])
        # add_pair_links(self, hosts['h40'], hosts['h60'], s2[4])
        # add_pair_links(self, hosts['h41'], hosts['h61'], s2[5])
        # add_pair_links(self, hosts['h50'], hosts['h70'], s2[6])
        # add_pair_links(self, hosts['h51'], hosts['h71'], s2[7])

        # # Layer 3
        # add_pair_links(self, hosts['h00'], hosts['h40'], s3[0])
        # add_pair_links(self, hosts['h01'], hosts['h41'], s3[1])
        # add_pair_links(self, hosts['h10'], hosts['h50'], s3[2])
        # add_pair_links(self, hosts['h11'], hosts['h51'], s3[3])
        # add_pair_links(self, hosts['h20'], hosts['h60'], s3[4])
        # add_pair_links(self, hosts['h21'], hosts['h61'], s3[5])
        # add_pair_links(self, hosts['h30'], hosts['h70'], s3[6])
        # add_pair_links(self, hosts['h31'], hosts['h71'], s3[7])
    # level 2 switch s22 connects to h10 and h30
    # level 2 switch s23 connects to h11 and h31

    # h01  s11 s21 s31

    # level 3 switch s32 connects to h10 and h60
    # level 3 switch s34 connects to h21 and h60
    # level 3 switch s36 connects to h30 and h70

def configure_host_routing(net):
    """Configure host routing for BCube paths."""
    print("\n=== Configuring host routing ===")
    
    h00 = net.get('h00'); h40 = net.get('h40'); h50 = net.get('h50')
    h20 = net.get('h20'); h30 = net.get('h30')
    h60 = net.get('h60'); h61 = net.get('h61'); h70 = net.get('h70')
    
    s30 = net.get('s30'); s14 = net.get('s14'); s12 = net.get('s12')
    s06 = net.get('s06'); s16 = net.get('s16')
    
    # RED PATH: h00 ↔ h40 via s30
    h00_s30_intf = None; h00_s30_mac = None
    for intf in h00.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s30 or intf.link.intf2.node == s30:
                h00_s30_intf = intf.name; h00_s30_mac = intf.MAC(); break
    
    h40_s30_intf = None; h40_s30_mac = None
    for intf in h40.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s30 or intf.link.intf2.node == s30:
                h40_s30_intf = intf.name; h40_s30_mac = intf.MAC(); break
    
    if h00_s30_intf and h40_s30_intf:
        h00.cmd(f'ip route add {h40.IP()} dev {h00_s30_intf}')
        h40.cmd(f'ip route add {h00.IP()} dev {h40_s30_intf}')
        h00.cmd(f'arp -s {h40.IP()} {h40_s30_mac}')
        h40.cmd(f'arp -s {h00.IP()} {h00_s30_mac}')
        print(f"✓ RED: h00→s30→h40")
    
    # GREEN PATH: h00 ↔ h50 via h40 relay
    h40.cmd('sysctl -w net.ipv4.ip_forward=1')
    h40.cmd('sysctl -w net.ipv4.conf.all.rp_filter=0')
    h40.cmd('sysctl -w net.ipv4.conf.default.rp_filter=0')
    for intf in h40.intfList():
        if intf.name != 'lo':
            h40.cmd(f'sysctl -w net.ipv4.conf.{intf.name}.rp_filter=0')
    
    h40_s14_intf = None; h40_s14_mac = None
    for intf in h40.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s14 or intf.link.intf2.node == s14:
                h40_s14_intf = intf.name; h40_s14_mac = intf.MAC(); break
    
    h50_s14_intf = None; h50_s14_mac = None
    for intf in h50.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s14 or intf.link.intf2.node == s14:
                h50_s14_intf = intf.name; h50_s14_mac = intf.MAC(); break
    
    if h00_s30_intf:
        h00.cmd(f'ip route add {h50.IP()} dev {h00_s30_intf}')
        h00.cmd(f'arp -s {h50.IP()} {h40_s30_mac}')
    if h50_s14_intf:
        h50.cmd(f'ip route add {h00.IP()} dev {h50_s14_intf}')
        h50.cmd(f'arp -s {h00.IP()} {h40_s14_mac}')
    if h40_s30_intf and h40_s14_intf:
        h40.cmd(f'ip route add {h50.IP()} dev {h40_s14_intf}')
        h40.cmd(f'ip route add {h00.IP()} dev {h40_s30_intf}')
        h40.cmd(f'arp -s {h00.IP()} {h00_s30_mac}')
        h40.cmd(f'arp -s {h50.IP()} {h50_s14_mac}')
        print(f"✓ GREEN: h00→s30→h40→s14→h50 (relay)")
    
    # BLUE PATH: h20 ↔ h30 via s12
    h20_s12_intf = None; h20_s12_mac = None
    for intf in h20.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s12 or intf.link.intf2.node == s12:
                h20_s12_intf = intf.name; h20_s12_mac = intf.MAC(); break
    
    h30_s12_intf = None; h30_s12_mac = None
    for intf in h30.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s12 or intf.link.intf2.node == s12:
                h30_s12_intf = intf.name; h30_s12_mac = intf.MAC(); break
    
    if h20_s12_intf and h30_s12_intf:
        h20.cmd(f'ip route add {h30.IP()} dev {h20_s12_intf}')
        h30.cmd(f'ip route add {h20.IP()} dev {h30_s12_intf}')
        h20.cmd(f'arp -s {h30.IP()} {h30_s12_mac}')
        h30.cmd(f'arp -s {h20.IP()} {h20_s12_mac}')
        print(f"✓ BLUE: h20→s12→h30")
    
    # PURPLE PATH: h60 ↔ h61 via s06
    print(f"✓ PURPLE: h60↔h61 (s06)")
    
    # BLACK PATH: h60 ↔ h70 via s16
    h60_s16_intf = None; h60_s16_mac = None
    for intf in h60.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s16 or intf.link.intf2.node == s16:
                h60_s16_intf = intf.name; h60_s16_mac = intf.MAC(); break
    
    h70_s16_intf = None; h70_s16_mac = None
    for intf in h70.intfList():
        if intf.name != 'lo' and intf.link:
            if intf.link.intf1.node == s16 or intf.link.intf2.node == s16:
                h70_s16_intf = intf.name; h70_s16_mac = intf.MAC(); break
    
    if h60_s16_intf and h70_s16_intf:
        h60.cmd(f'ip route add {h70.IP()} dev {h60_s16_intf}')
        h70.cmd(f'ip route add {h60.IP()} dev {h70_s16_intf}')
        h60.cmd(f'arp -s {h70.IP()} {h70_s16_mac}')
        h70.cmd(f'arp -s {h60.IP()} {h60_s16_mac}')
        print(f"✓ BLACK: h60→s16→h70")
    print()


def add_flows_bcube(net):
    """Add OpenFlow rules for the 5 paths."""
    h00 = net.get('h00'); h40 = net.get('h40'); h50 = net.get('h50')
    h20 = net.get('h20'); h30 = net.get('h30')
    h60 = net.get('h60'); h61 = net.get('h61'); h70 = net.get('h70')
    s30 = net.get('s30'); s14 = net.get('s14'); s12 = net.get('s12')
    s06 = net.get('s06'); s16 = net.get('s16')

    print("=== Configuring OpenFlow rules ===\n")
    
    for sw in [s30, s14, s12, s06, s16]:
        os.system(f"ovs-ofctl del-flows {sw.name}")
        os.system(f"ovs-ofctl add-flow {sw.name} 'priority=0,actions=drop'")
        os.system(f"ovs-ofctl --protocols=OpenFlow13 add-flow {sw.name} 'priority=50,idle_timeout=0,arp,actions=FLOOD'")
        #ovs-ofctl --protocols=OpenFlow13 add-flow "$s" ""
    
    # RED PATH
    h00_conn = s30.connectionsTo(h00); h40_conn = s30.connectionsTo(h40)
    if h00_conn and h40_conn:
        p_h00 = s30.ports[h00_conn[0][0]]; p_h40 = s30.ports[h40_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h40.IP()},actions=output:{p_h40}'")
        os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h40.IP()},nw_dst={h00.IP()},actions=output:{p_h00}'")
        print(f"[RED] s30 flows added")

    # GREEN PATH
    os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h50.IP()},actions=output:{p_h40}'")
    os.system(f"ovs-ofctl add-flow {s30.name} 'priority=100,ip,nw_src={h50.IP()},nw_dst={h00.IP()},actions=output:{p_h00}'")
    h40_s14 = s14.connectionsTo(h40); h50_s14 = s14.connectionsTo(h50)
    if h40_s14 and h50_s14:
        p_h40_s14 = s14.ports[h40_s14[0][0]]; p_h50_s14 = s14.ports[h50_s14[0][0]]
        os.system(f"ovs-ofctl add-flow {s14.name} 'priority=100,ip,nw_src={h00.IP()},nw_dst={h50.IP()},actions=output:{p_h50_s14}'")
        os.system(f"ovs-ofctl add-flow {s14.name} 'priority=100,ip,nw_src={h50.IP()},nw_dst={h00.IP()},actions=output:{p_h40_s14}'")
        print(f"[GREEN] s30,s14 flows added")

    # BLUE PATH
    h20_conn = s12.connectionsTo(h20); h30_conn = s12.connectionsTo(h30)
    if h20_conn and h30_conn:
        p_h20 = s12.ports[h20_conn[0][0]]; p_h30 = s12.ports[h30_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s12.name} 'priority=100,ip,nw_src={h20.IP()},nw_dst={h30.IP()},actions=output:{p_h30}'")
        os.system(f"ovs-ofctl add-flow {s12.name} 'priority=100,ip,nw_src={h30.IP()},nw_dst={h20.IP()},actions=output:{p_h20}'")
        print(f"[BLUE] s12 flows added")

    # PURPLE PATH
    h60_conn = s06.connectionsTo(h60); h61_conn = s06.connectionsTo(h61)
    if h60_conn and h61_conn:
        p_h60 = s06.ports[h60_conn[0][0]]; p_h61 = s06.ports[h61_conn[0][0]]
        os.system(f"ovs-ofctl add-flow {s06.name} 'priority=100,ip,nw_src={h60.IP()},nw_dst={h61.IP()},actions=output:{p_h61}'")
        os.system(f"ovs-ofctl add-flow {s06.name} 'priority=100,ip,nw_src={h61.IP()},nw_dst={h60.IP()},actions=output:{p_h60}'")
        print(f"[PURPLE] s06 flows added")

    # BLACK PATH
    h60_s16 = s16.connectionsTo(h60); h70_s16 = s16.connectionsTo(h70)
    if h60_s16 and h70_s16:
        p_h60_s16 = s16.ports[h60_s16[0][0]]; p_h70_s16 = s16.ports[h70_s16[0][0]]
        os.system(f"ovs-ofctl add-flow {s16.name} 'priority=100,ip,nw_src={h60.IP()},nw_dst={h70.IP()},actions=output:{p_h70_s16}'")
        os.system(f"ovs-ofctl add-flow {s16.name} 'priority=100,ip,nw_src={h70.IP()},nw_dst={h60.IP()},actions=output:{p_h60_s16}'")
        print(f"[BLACK] s16 flows added")
    print()

def setup_static_flows_secure():
    """
    Setup static OVS flows for BCube paths (secure mode).
    Converted from iperf.sh bash script.
    """
    print("=" * 42)
    print("Setting up static OVS flows for BCube paths (secure mode)")
    print("=" * 42)
    
    # Define all switches
    switches = [
        's00', 's01', 's02', 's03', 's04', 's05', 's06', 's07',
        's10', 's11', 's12', 's13', 's14', 's15', 's16', 's17',
        's20', 's21', 's22', 's23', 's24', 's25', 's26', 's27',
        's30', 's31', 's32', 's33', 's34', 's35', 's36', 's37'
    ]
    
    path_switches = ['s30', 's14', 's12', 's06', 's16']
    
    # Set fail-mode to secure
    print("Setting fail-mode to secure...")
    for s in switches:
        cmd = f"ovs-vsctl br-exists {s} && ovs-vsctl set-fail-mode {s} secure"
        os.system(cmd)
    
    time.sleep(2)
    
    # Clear old flows
    print("Clearing old flows...")
    for s in switches:
        cmd = f"ovs-vsctl br-exists {s} && ovs-ofctl --protocols=OpenFlow13 del-flows {s} 2>/dev/null || true"
        os.system(cmd)
        print(f"  Cleared flows on {s}")
    
    # Add ARP flood rules on path switches
    print("Adding ARP flood rules on path switches...")
    for s in path_switches:
        cmd = f"ovs-ofctl --protocols=OpenFlow13 add-flow {s} 'priority=50,idle_timeout=0,arp,actions=FLOOD'"
        os.system(cmd)
    
    # ========= Path-specific rules =========
    # RED: h00 <-> h40 via s30
    print("Adding path-specific rules...")
    os.system(f"ovs-ofctl add-flow s30 'priority=100,idle_timeout=0,in_port=1,actions=output:2'")
    os.system(f"ovs-ofctl add-flow s30 'priority=100,idle_timeout=0,in_port=2,actions=output:1'")
    
    # GREEN: h00 <-> h50 via s30 → h40 → s14
    os.system(f"ovs-ofctl add-flow s14 'priority=100,idle_timeout=0,in_port=1,actions=output:2'")
    os.system(f"ovs-ofctl add-flow s14 'priority=100,idle_timeout=0,in_port=2,actions=output:1'")
    
    # BLUE: h20 <-> h30 via s12
    os.system(f"ovs-ofctl add-flow s12 'priority=100,idle_timeout=0,in_port=1,actions=output:2'")
    os.system(f"ovs-ofctl add-flow s12 'priority=100,idle_timeout=0,in_port=2,actions=output:1'")
    
    # PURPLE: h60 <-> h61 via s06
    os.system(f"ovs-ofctl add-flow s06 'priority=100,idle_timeout=0,in_port=1,actions=output:2'")
    os.system(f"ovs-ofctl add-flow s06 'priority=100,idle_timeout=0,in_port=2,actions=output:1'")
    
    # BLACK: h60 <-> h70 via s16
    os.system(f"ovs-ofctl add-flow s16 'priority=100,idle_timeout=0,in_port=1,actions=output:2'")
    os.system(f"ovs-ofctl add-flow s16 'priority=100,idle_timeout=0,in_port=2,actions=output:1'")
    
    print("=" * 42)
    print("All static flows installed under secure mode.")
    print("Test in Mininet CLI with:")
    print("  iperf h00 h40")
    print("  iperf h00 h50")
    print("  iperf h20 h30")
    print("  iperf h60 h61")
    print("  iperf h60 h70")
    print("=" * 42)
    print()

def setup_host_bridges(net):
    """Merge all host interfaces into br0 and assign a single IP."""
    info('*** Setting up internal bridges on each host\n')
    base_ip = 10  # start IP offset

    for idx, h in enumerate(net.hosts, start=base_ip):
        name = h.name
        info(f' Configuring {name}\n')

        h.cmd("ip link add br0 type bridge")
        h.cmd("ip link set br0 up")

        # Move all host interfaces under the bridge
        for intf in h.intfList():
            h.cmd(f"ip addr flush dev {intf.name}")
            h.cmd(f"ip link set {intf.name} master br0")
            h.cmd(f"ip link set {intf.name} up")

        # Single IP per host (10.0.0.X)
        ip_addr = f"10.0.0.{idx}/24"
        h.cmd(f"ip addr add {ip_addr} dev br0")
        info(f"  {name}: br0 -> {ip_addr}\n")
    
    for h in net.hosts:
        ip = h.cmd("ip -4 addr show br0 | grep 'inet ' | awk '{print $2}' | cut -d/ -f1").strip()
        if ip:
            h.IP = lambda ip=ip: ip  # override Mininet's default IP() lookup
        else:
            h.IP = lambda: None

def test_connectivity(net):
    """Test connectivity with ping."""
    print("\n" + "="*70)
    print("AUTOMATED CONNECTIVITY TESTS (PING)")
    print("="*70)
    
    h00 = net.get('h00'); h40 = net.get('h40'); h50 = net.get('h50')
    h20 = net.get('h20'); h30 = net.get('h30')
    h60 = net.get('h60'); h61 = net.get('h61'); h70 = net.get('h70')
    
    tests = [
        ('RED', h00, h40, 's30'),
        ('GREEN', h00, h50, 's30→h40→s14'),
        ('BLUE', h20, h30, 's12'),
        ('PURPLE', h60, h61, 's06'),
        ('BLACK', h60, h70, 's16')
    ]
    
    for name, src, dst, path in tests:
        print(f"\n[{name}] {src.name} ping {dst.name} (via {path})")
        result = src.cmd(f'ping -c 3 -W 2 {dst.IP()}')
        
        if '3 received' in result or '3 packets transmitted, 3 received' in result:
            import re
            rtt = re.search(r'rtt min/avg/max/mdev = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+)', result)
            if rtt:
                print(f"  ✓ PASS - 3/3 packets, avg RTT: {rtt.group(2)} ms")
            else:
                print(f"  ✓ PASS - 3/3 packets received")
        elif '0 received' in result:
            print(f"  ✗ FAIL - 0/3 packets received")
        else:
            print(f"  ⚠ PARTIAL - Some packet loss")
    
    print("\n" + "="*70)
    print("CONNECTIVITY TESTS COMPLETE")
    print("="*70 + "\n")


def run():
    topo = BCube32()
    net = Mininet(topo=topo, switch=OVSSwitch, link=TCLink, controller=None)
    net.start()
    info("\n*** BCube(3,2) topology built ***\n")
    setup_host_bridges(net)
    setup_static_flows_secure()
    #configure_host_routing(net)
    #add_flows_bcube(net)
    #test_connectivity(net)
    
    info("\n*** Manual Testing Commands ***\n")
    info("="*70 + "\n")
    info("IPERF bandwidth tests:\n")
    info("  iperf h00 h40  # RED path\n")
    info("  iperf h00 h50  # GREEN path\n")
    info("  iperf h20 h30  # BLUE path\n")
    info("  iperf h60 h61  # PURPLE path\n")
    info("  iperf h60 h70  # BLACK path\n")
    info("\nPING tests:\n")
    info("  h00 ping -c 3 h40\n")
    info("  h00 ping -c 3 h50\n")
    info("  h20 ping -c 3 h30\n")
    info("  h60 ping -c 3 h61\n")
    info("  h60 ping -c 3 h70\n")
    info("="*70 + "\n\n")

    CLI(net)
    net.stop()

if __name__ == '__main__':
    setLogLevel('info')
    run()