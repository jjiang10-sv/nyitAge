#!/usr/bin/env python3
"""
bcube_hybrid_bridge_topo.py

Hybrid BCube topology using internal host bridges (br0)
=======================================================

- 16 hosts (h00–h71)
- 32 switches (s_00–s_37), arranged in 4 layers (8 each)
- Each switch connects exactly two hosts
- Each host merges all its interfaces into a single internal Linux bridge (br0)
- Each host has ONE IP (10.0.0.X/24)
- No IP forwarding (pure Layer 2 bridging)
- Switches are controlled via external OVS flow script (fail-mode=secure)
- ARP floods are still enabled via switch flow rules

Usage:
  sudo python3 bcube_hybrid_bridge_topo.py
  # Then run your flow setup shell script (e.g., BcubeFlow.sh)
"""

from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import OVSSwitch
from mininet.link import TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel, info


# ------------------------------------------------------------
# Helper Functions
# ------------------------------------------------------------
def add_pair_links(topo, hA, hB, sw):
    """Connect two hosts to a switch with link options."""
    topo.addLink(hA, sw, **topo.linkopts)
    topo.addLink(hB, sw, **topo.linkopts)


# ------------------------------------------------------------
# Topology Definition
# ------------------------------------------------------------
class BCubeBridgeTopo(Topo):
    def build(self):
        self.linkopts = dict(bw=8, delay='4ms')

        # --- Hosts (16) ---
        host_names = [
            'h00','h01','h10','h11','h20','h21','h30','h31',
            'h40','h41','h50','h51','h60','h61','h70','h71'
        ]
        hosts = {name: self.addHost(name) for name in host_names}

        # --- Switches (32) ---
        s0 = [self.addSwitch(f's_0{i}') for i in range(8)]
        s1 = [self.addSwitch(f's_1{i}') for i in range(8)]
        s2 = [self.addSwitch(f's_2{i}') for i in range(8)]
        s3 = [self.addSwitch(f's_3{i}') for i in range(8)]

        info('*** Creating BCube layer links\n')

        # Layer 0
        add_pair_links(self, hosts['h00'], hosts['h01'], s0[0])
        add_pair_links(self, hosts['h10'], hosts['h11'], s0[1])
        add_pair_links(self, hosts['h20'], hosts['h21'], s0[2])
        add_pair_links(self, hosts['h30'], hosts['h31'], s0[3])
        add_pair_links(self, hosts['h40'], hosts['h41'], s0[4])
        add_pair_links(self, hosts['h50'], hosts['h51'], s0[5])
        add_pair_links(self, hosts['h60'], hosts['h61'], s0[6])
        add_pair_links(self, hosts['h70'], hosts['h71'], s0[7])

        # Layer 1
        add_pair_links(self, hosts['h00'], hosts['h10'], s1[0])
        add_pair_links(self, hosts['h01'], hosts['h11'], s1[1])
        add_pair_links(self, hosts['h20'], hosts['h30'], s1[2])
        add_pair_links(self, hosts['h21'], hosts['h31'], s1[3])
        add_pair_links(self, hosts['h40'], hosts['h50'], s1[4])
        add_pair_links(self, hosts['h41'], hosts['h51'], s1[5])
        add_pair_links(self, hosts['h60'], hosts['h70'], s1[6])
        add_pair_links(self, hosts['h61'], hosts['h71'], s1[7])

        # Layer 2
        add_pair_links(self, hosts['h00'], hosts['h20'], s2[0])
        add_pair_links(self, hosts['h01'], hosts['h21'], s2[1])
        add_pair_links(self, hosts['h10'], hosts['h30'], s2[2])
        add_pair_links(self, hosts['h11'], hosts['h31'], s2[3])
        add_pair_links(self, hosts['h40'], hosts['h60'], s2[4])
        add_pair_links(self, hosts['h41'], hosts['h61'], s2[5])
        add_pair_links(self, hosts['h50'], hosts['h70'], s2[6])
        add_pair_links(self, hosts['h51'], hosts['h71'], s2[7])

        # Layer 3
        add_pair_links(self, hosts['h00'], hosts['h40'], s3[0])
        add_pair_links(self, hosts['h01'], hosts['h41'], s3[1])
        add_pair_links(self, hosts['h10'], hosts['h50'], s3[2])
        add_pair_links(self, hosts['h11'], hosts['h51'], s3[3])
        add_pair_links(self, hosts['h20'], hosts['h60'], s3[4])
        add_pair_links(self, hosts['h21'], hosts['h61'], s3[5])
        add_pair_links(self, hosts['h30'], hosts['h70'], s3[6])
        add_pair_links(self, hosts['h31'], hosts['h71'], s3[7])


# ------------------------------------------------------------
# Network Configuration Functions
# ------------------------------------------------------------
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


# ------------------------------------------------------------
# Main Execution
# ------------------------------------------------------------
def run():
    setLogLevel('info')
    info('*** Building BCube Hybrid Bridge Topology (no controller)\n')

    topo = BCubeBridgeTopo()
    net = Mininet(
        topo=topo,
        controller=None,
        switch=OVSSwitch,
        link=TCLink,
        autoSetMacs=True,
        autoStaticArp=True
    )

    net.start()
    setup_host_bridges(net)
    # Patch Mininet's .IP() method to use br0 address
    for h in net.hosts:
        ip = h.cmd("ip -4 addr show br0 | grep 'inet ' | awk '{print $2}' | cut -d/ -f1").strip()
        if ip:
            h.IP = lambda ip=ip: ip  # override Mininet's default IP() lookup
        else:
            h.IP = lambda: None
    
    info('\n*** Host IP Assignments (from br0)\n')
    for h in net.hosts:
        ip = h.IP()
        info(f' {h.name}: {ip}\n')
    
    info('\n*** BCube Hybrid Bridge topology ready.\n')
    info(' Each host has br0 bridging all NICs with single IP.\n')
    info(' Switch control and ARP flood handled by your flow script.\n')
    info(' Try ping/iperf across allowed paths only.\n\n')

    CLI(net)
    net.stop()


if __name__ == '__main__':
    run()

