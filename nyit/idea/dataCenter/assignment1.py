#!/usr/bin/env python3
"""
Custom_BCube_Topo.py
BCube(3,2): 8 mini-cubes, 2 hosts each, 4 switch levels (0–3)
Link bandwidth = 8 Mbps, delay = 4 ms
"""

from mininet.topo import Topo
from mininet.net import Mininet
from mininet.node import OVSSwitch
from mininet.link import TCLink
from mininet.cli import CLI
from mininet.log import setLogLevel, info

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
                    # reminder = cube % (2^lvl)
                    # if reminder >= 2^(lvl-1):
                    #     idx1 += 1
                    if counter == 2 and lvl > 1:
                        idx1 += 1
                        counter = 0
                    idx1 += h%2 * (2**(lvl-1))
                    host = f'h{idx1}{idx0}'
                    if host == "h80":
                        print(f"lvl:{lvl}, cube:{cube}, idx0:{idx0}, idx1:{idx1}, host:{host}, sw:{sw}")
                    self.addLink(host, sw, bw=bw, delay=delay)
                #
                # self.addLink("00", sw, bw=bw, delay=delay) 
                # binary("00") + 2^(lvl)
                # self.addLink("01", sw, bw=bw, delay=delay)

                # level = 3; cube = 0. host:00, 40
                # level = 3; cube = 1. host:01, 41
                # level = 3; cube = 2. host:10, 50
                # level = 3; cube = 3. host:11, 51
                # level = 3; cube = 4. host:20, 60
                # level = 3; cube = 5. host:21, 61
                # level = 3; cube = 6. host:30, 70
                # level = 3; cube = 7. host:31, 71


                # levl = 2; cube = 0. host:00, 20. idx1 = 0 * 4 = 0
                # levl = 2; cube = 1. host:01, 21. idx1 = 0 * 4 = 0
                # levl = 2; cube = 2. host:10, 30. idx1 = 1*4
                # levl = 2; cube = 3. host:11, 31. idx1 = 1*4
                # levl = 2; cube = 4. host:40, 60   
                # levl = 2; cube = 5. host:41, 61
                # levl = 2; cube = 6. host:50, 70
                # levl = 2; cube = 7. host:51, 71


                # levl = 1; cube = 0. host:00, 10
                # levl = 1; cube = 1. host:01, 11
                # levl = 1; cube = 2. host:20, 30
                # level =1; cube = 3. host:21, 31
                # level =1; cube = 4. host:40, 50
                # level =1; cube = 5. host:41, 51
                # level =1; cube = 6. host:60, 70
                # level =1; cube = 7. host:61, 71

                # for h in range(n):
                #     # sw 10 -> h00, h10.   
                #     # sw 11 -> h01, h11
                #     host = f'h{cube}{h}'
                #     # flip bit (lvl-1) of cube index to get target switch index
                #     target = cube ^ (1 << (lvl - 1))
                #     sw = f's{lvl}{target}'
                #     self.addLink(host, sw, bw=bw, delay=delay)

def run():
    topo = BCube32()
    net = Mininet(topo=topo, switch=OVSSwitch, link=TCLink, controller=None)
    net.start()
    info("\n*** BCube(3,2) topology built (8 × 2 hosts, 4 levels) ***\n")
    CLI(net)
    net.stop()

if __name__ == '__main__':
    setLogLevel('info')
    run()
