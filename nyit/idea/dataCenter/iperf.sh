#!/usr/bin/env bash
# Static flows for BCube hybrid topology (secure mode)

set -euo pipefail

echo "=========================================="
echo "Setting up static OVS flows for BCube paths (secure mode)"
echo "=========================================="

SWITCHES=(
  s00 s01 s02 s03 s04 s05 s06 s07
  s10 s11 s12 s13 s14 s15 s16 s17
  s20 s21 s22 s23 s24 s25 s26 s27
  s30 s31 s32 s33 s34 s35 s36 s37
)

PATH_SWITCHES=(s_30 s14 s12 s06 s16)

echo "Setting fail-mode to secure..."
for s in "${SWITCHES[@]}"; do
  if ovs-vsctl br-exists "$s"; then
    ovs-vsctl set-fail-mode "$s" secure
  fi
done

sleep 2
echo "Clearing old flows..."
for s in "${SWITCHES[@]}"; do
  if ovs-vsctl br-exists "$s"; then
    ovs-ofctl --protocols=OpenFlow13 del-flows "$s" || true
    echo "  Cleared flows on $s"
  fi
done

echo "Adding ARP flood rules on path switches..."
for s in "${PATH_SWITCHES[@]}"; do
  ovs-ofctl --protocols=OpenFlow13 add-flow "$s" "priority=50,idle_timeout=0,arp,actions=FLOOD"
done

# ========= Path-specific rules =========
# RED: h00 <-> h40 via s30
ovs-ofctl add-flow s30 "priority=100,idle_timeout=0,in_port=1,actions=output:2"
ovs-ofctl add-flow s30 "priority=100,idle_timeout=0,in_port=2,actions=output:1"

# GREEN: h00 <-> h50 via s30 → h40 → s14
ovs-ofctl add-flow s14 "priority=100,idle_timeout=0,in_port=1,actions=output:2"
ovs-ofctl add-flow s14 "priority=100,idle_timeout=0,in_port=2,actions=output:1"

# BLUE: h20 <-> h30 via s12
ovs-ofctl add-flow s12 "priority=100,idle_timeout=0,in_port=1,actions=output:2"
ovs-ofctl add-flow s12 "priority=100,idle_timeout=0,in_port=2,actions=output:1"

# PURPLE: h60 <-> h61 via s06
ovs-ofctl add-flow s06 "priority=100,idle_timeout=0,in_port=1,actions=output:2"
ovs-ofctl add-flow s06 "priority=100,idle_timeout=0,in_port=2,actions=output:1"

# BLACK: h60 <-> h70 via s16
ovs-ofctl add-flow s16 "priority=100,idle_timeout=0,in_port=1,actions=output:2"
ovs-ofctl add-flow s16 "priority=100,idle_timeout=0,in_port=2,actions=output:1"

echo "=========================================="
echo "All static flows installed under secure mode."
echo "Test in Mininet CLI with:"
echo "  iperf h00 h40"
echo "  iperf h00 h50"
echo "  iperf h20 h30"
echo "  iperf h60 h61"
echo "  iperf h60 h70"
echo "=========================================="

