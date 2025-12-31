#!/usr/bin/env python3
"""
Custom_BCube_Topo.py
====================
Builds a BCube data center topology using Mininet
with host-level bridges for realistic multi-interface connectivity.

Each host connects to multiple BCube layers (switches).
A bridge (br0) inside every host ties all interfaces together,
so the host acts like a mini-switch (as in the true BCube design).

"""

# --------------------------------------------------------------
# Imports
# --------------------------------------------------------------
from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import OVSSwitch, RemoteController
from mininet.link import TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel, info


# --------------------------------------------------------------
# Helper Function – Add a pair of hosts to a switch
# --------------------------------------------------------------
def add_pair_links(topo, hA, hB, sw):
    """Connect two hosts to the same switch."""
    topo.addLink(hA, sw, **topo.linkopts)
    topo.addLink(hB, sw, **topo.linkopts)


# --------------------------------------------------------------
# BCube Topology Definition
# --------------------------------------------------------------
class BCubeTopo(Topo):
    def build(self):
        """
        Constructs a 4-layer BCube topology.
        Each layer has 8 switches; each host connects to one switch per layer.
        """
        self.linkopts = dict(bw=8, delay='4ms')  # link characteristics

        # ---- Create Hosts ----
        host_names = [
            'h00','h01','h10','h11','h20','h21','h30','h31',
            'h40','h41','h50','h51','h60','h61','h70','h71'
        ]
        hosts = {n: self.addHost(n) for n in host_names}

        # ---- Create Switches with stable DPIDs ----
        # Each switch gets an OpenFlow 1.0 DPID for FlowVisor slicing
        def mk(name, dpid_hex):
            return self.addSwitch(
                name,
                cls=OVSSwitch,
                protocols='OpenFlow10',
                failmode='secure',
                dpid=f'{dpid_hex:016x}'
            )

        # Layers 0–3, each with 8 switches
        self.s0 = [mk(f's_0{i}', 0x100 + i) for i in range(8)]
        self.s1 = [mk(f's_1{i}', 0x110 + i) for i in range(8)]
        self.s2 = [mk(f's_2{i}', 0x120 + i) for i in range(8)]
        self.s3 = [mk(f's_3{i}', 0x130 + i) for i in range(8)]

        info('*** Creating BCube layer links\n')

        # ---- Layer 0 connections ----
        add_pair_links(self, hosts['h00'], hosts['h01'], self.s0[0])
        add_pair_links(self, hosts['h10'], hosts['h11'], self.s0[1])
        add_pair_links(self, hosts['h20'], hosts['h21'], self.s0[2])
        add_pair_links(self, hosts['h30'], hosts['h31'], self.s0[3])
        add_pair_links(self, hosts['h40'], hosts['h41'], self.s0[4])
        add_pair_links(self, hosts['h50'], hosts['h51'], self.s0[5])
        add_pair_links(self, hosts['h60'], hosts['h61'], self.s0[6])
        add_pair_links(self, hosts['h70'], hosts['h71'], self.s0[7])

        # ---- Layer 1 connections ----
        add_pair_links(self, hosts['h00'], hosts['h10'], self.s1[0])  # s_10
        add_pair_links(self, hosts['h01'], hosts['h11'], self.s1[1])
        add_pair_links(self, hosts['h20'], hosts['h30'], self.s1[2])
        add_pair_links(self, hosts['h21'], hosts['h31'], self.s1[3])
        add_pair_links(self, hosts['h40'], hosts['h50'], self.s1[4])  # s_14
        add_pair_links(self, hosts['h41'], hosts['h51'], self.s1[5])  # s_15
        add_pair_links(self, hosts['h60'], hosts['h70'], self.s1[6])
        add_pair_links(self, hosts['h61'], hosts['h71'], self.s1[7])

        # ---- Layer 2 connections ----
        add_pair_links(self, hosts['h00'], hosts['h20'], self.s2[0])
        add_pair_links(self, hosts['h01'], hosts['h21'], self.s2[1])
        add_pair_links(self, hosts['h10'], hosts['h30'], self.s2[2])
        add_pair_links(self, hosts['h11'], hosts['h31'], self.s2[3])
        add_pair_links(self, hosts['h40'], hosts['h60'], self.s2[4])  # s_24
        add_pair_links(self, hosts['h41'], hosts['h61'], self.s2[5])
        add_pair_links(self, hosts['h50'], hosts['h70'], self.s2[6])
        add_pair_links(self, hosts['h51'], hosts['h71'], self.s2[7])

        # ---- Layer 3 connections ----
        add_pair_links(self, hosts['h00'], hosts['h40'], self.s3[0])  # s_30
        add_pair_links(self, hosts['h01'], hosts['h41'], self.s3[1])  # s_31
        add_pair_links(self, hosts['h10'], hosts['h50'], self.s3[2])  # s_32
        add_pair_links(self, hosts['h11'], hosts['h51'], self.s3[3])  # s_33
        add_pair_links(self, hosts['h20'], hosts['h60'], self.s3[4])
        add_pair_links(self, hosts['h21'], hosts['h61'], self.s3[5])
        add_pair_links(self, hosts['h30'], hosts['h70'], self.s3[6])
        add_pair_links(self, hosts['h31'], hosts['h71'], self.s3[7])


# --------------------------------------------------------------
# Setup Function – Create bridges on hosts
# --------------------------------------------------------------
def setup_host_bridges(net):
    """
    Creates an internal bridge (br0) inside every host.
    - Binds all interfaces (eth0, eth1, eth2, ...) to br0
    - Assigns a single IP address to br0
    """
    info('*** Setting up host bridges for BCube realism\n')
    base_ip = 10  # starting IP: 10.0.0.10

    for idx, host in enumerate(net.hosts, start=base_ip):
        host.cmd("ip link add br0 type bridge")
        host.cmd("ip link set br0 up")

        for intf in host.intfList():
            host.cmd(f"ip addr flush dev {intf.name}")
            host.cmd(f"ip link set {intf.name} master br0")
            host.cmd(f"ip link set {intf.name} up")

        ip_addr = f"10.0.0.{idx}/24"
        host.cmd(f"ip addr add {ip_addr} dev br0")
        info(f"{host.name} assigned {ip_addr}\n")


# --------------------------------------------------------------
# Utility Function – Print switch DPIDs
# --------------------------------------------------------------
def print_switch_dpids(net):
    """Displays all switch names with their DPIDs (useful for FlowVisor flowspace rules)."""
    info('\n*** Switch DPID Table ***\n')
    for sw in net.switches:
        dpid = sw.dpid
        info(f"{sw.name}: {dpid}\n")


# --------------------------------------------------------------
# Runner Function
# --------------------------------------------------------------
def run():
    """Builds and runs the topology, connects to FlowVisor, and opens the Mininet CLI."""
    setLogLevel('info')
    info('*** Launching BCube topology (with host bridges) ***\n')

    topo = BCubeTopo()
    net = Mininet(
        topo=topo,
        controller=None,            # FlowVisor will control the switches
        switch=OVSSwitch,
        link=TCLink,
        autoSetMacs=True,
        autoStaticArp=True
    )

    # Point all switches to FlowVisor (default 127.0.0.1:6633)
    net.addController('fv', controller=RemoteController, ip='127.0.0.1', port=6633)

    net.start()
    setup_host_bridges(net)

    # Print tables for convenience
    info('\n==================== SUMMARY ====================\n')
    info('*** Host IP Table ***\n')
    for host in net.hosts:
    	ip_addr = host.cmd("ip -4 addr show br0 | grep -oP '(?<=inet\\s)\\d+(\\.\\d+){3}'").strip()
        info(f"{host.name}: {ip_addr}\n")

    print_switch_dpids(net)

    info('=================================================\n')
    info('*** BCube topology is active. Switches connected to FlowVisor on 127.0.0.1:6633 ***\n')
    CLI(net)
    net.stop()


# --------------------------------------------------------------
#  Main Entry Point
# --------------------------------------------------------------
if __name__ == '__main__':
    run()
