#!/bin/bash
# Diagnostic script to debug iperf hang issue between h00 and h40

echo "======================================"
echo "IPERF DEBUG DIAGNOSTIC SCRIPT"
echo "======================================"
echo ""

echo "=== Step 1: Test basic ping connectivity ==="
echo "Command: h00 ping -c 3 10.0.0.40"
echo "Expected: Should see 3 packets received"
echo "Run this in Mininet CLI and paste results"
echo ""

echo "=== Step 2: Check if iperf server can be manually started ==="
echo "Command: h40 iperf -s -p 5001 &"
echo "Expected: Should see 'Server listening on TCP port 5001'"
echo "Run this in Mininet CLI and paste results"
echo ""

echo "=== Step 3: Check OpenFlow rules on s30 ==="
echo "Command: sh ovs-ofctl dump-flows s30"
echo "Expected: Should see rules for h00<->h40 traffic"
echo "Run this in Mininet CLI and paste results"
echo ""

echo "=== Step 4: Check routing tables ==="
echo "Command: h00 ip route"
echo "Expected: Should see route to 10.0.0.40 via specific interface"
echo "Run this in Mininet CLI and paste results"
echo ""

echo "=== Step 5: Check ARP entries ==="
echo "Command: h00 arp -n"
echo "Expected: Should see ARP entry for 10.0.0.40"
echo "Run this in Mininet CLI and paste results"
echo ""

echo "=== Step 6: Test TCP connection manually ==="
echo "Command on h40: nc -l -p 5001"
echo "Command on h00: nc 10.0.0.40 5001"
echo "Expected: Should establish connection"
echo "Run these in Mininet CLI and report if connection establishes"
echo ""

echo "=== Step 7: Capture packets on s30 ==="
echo "Command: sh timeout 10 tcpdump -i s30 -nn -c 20"
echo "Then run: iperf h00 h40"
echo "Expected: Should see TCP SYN packets"
echo "Run this in Mininet CLI and paste packet capture"
echo ""

echo "======================================"
echo "Please run these steps in order and provide the output"
echo "======================================"