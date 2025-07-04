172.21.0.3
john@johnjiangs-MacBook-Pro reconnaissance % docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 0d49f6b6384a
172.21.0.2

Here's a comprehensive guide to nmap commands and options:

## Basic nmap Syntax
```bash
nmap [Scan Type] [Options] {target specification}
```

## Target Specification
```bash
# Single IP
nmap 172.21.0.3

# IP range
nmap 172.21.0.3-254
nmap 172.21.0.0/24

# Multiple targets
nmap 172.21.0.3 172.21.0.5 172.21.0.100

# Hostname
nmap scanme.nmap.org

# Input from file
nmap -iL targets.txt

# Exclude targets
nmap 172.21.0.0/24 --exclude 172.21.0.3
nmap 172.21.0.0/24 --excludefile exclude.txt
```

## Scan Types
```bash
# TCP SYN scan (default, requires root)
nmap -sS 172.21.0.3

# TCP connect scan (no root required)
nmap -sT 172.21.0.3

# UDP scan
nmap -sU 172.21.0.3

# TCP ACK scan (firewall detection)
nmap -sA 172.21.0.3

# TCP Window scan
nmap -sW 172.21.0.3

# TCP Maimon scan
nmap -sM 172.21.0.3

# TCP Null scan
nmap -sN 172.21.0.3

# TCP FIN scan
nmap -sF 172.21.0.3

# TCP Xmas scan
nmap -sX 172.21.0.3

# IP protocol scan
nmap -sO 172.21.0.3

# Ping scan only (no port scan)
nmap -sn 172.21.0.0/24

# List scan (just list targets)
nmap -sL 172.21.0.0/24
```

## Port Specification
```bash
# Specific ports
nmap -p 22,80,443 172.21.0.3

# Port range
nmap -p 1-1000 172.21.0.3

# All ports
nmap -p- 172.21.0.3

# Top ports
nmap --top-ports 100 172.21.0.3

# Fast scan (100 most common ports)
nmap -F 172.21.0.3

# Specific port protocols
nmap -p U:53,T:21-25,80 172.21.0.3
```

## Host Discovery Options
```bash
# Skip ping (treat all hosts as up)
nmap -Pn 172.21.0.3

# ICMP ping
nmap -PE 172.21.0.3

# TCP SYN ping
nmap -PS22,80,443 172.21.0.3

# TCP ACK ping
nmap -PA22,80,443 172.21.0.3

# UDP ping
nmap -PU 172.21.0.3

# ARP ping (local network)
nmap -PR 172.21.0.3

# No ping
nmap -PN 172.21.0.3
```

## Service and Version Detection
```bash
# Service version detection
nmap -sV 172.21.0.3

# OS detection
nmap -O 172.21.0.3

# Aggressive service detection
nmap -sV --version-intensity 9 172.21.0.3

# Light service detection
nmap -sV --version-intensity 0 172.21.0.3

# Enable OS detection, version detection, script scanning, and traceroute
nmap -A 172.21.0.3
```

## Script Engine (NSE)
```bash
# Default scripts
nmap -sC 172.21.0.3
nmap --script=default 172.21.0.3

# Vulnerability scripts
nmap --script=vuln 172.21.0.3

# Specific script
nmap --script=http-title 172.21.0.3

# Multiple scripts
nmap --script="http-*" 172.21.0.3

# Script categories
nmap --script=auth,safe 172.21.0.3

# Script help
nmap --script-help=http-title

# Update script database
nmap --script-updatedb
```

## Timing and Performance
```bash
# Timing templates (0-5, paranoid to insane)
nmap -T0 172.21.0.3  # Paranoid (very slow)
nmap -T1 172.21.0.3  # Sneaky
nmap -T2 172.21.0.3  # Polite
nmap -T3 172.21.0.3  # Normal (default)
nmap -T4 172.21.0.3  # Aggressive
nmap -T5 172.21.0.3  # Insane (very fast)

# Parallel host scan groups
nmap --min-hostgroup 50 172.21.0.0/24

# Parallel port scan groups
nmap --min-parallelism 100 172.21.0.3

# Packet rate control
nmap --max-rate 100 172.21.0.3
nmap --min-rate 10 172.21.0.3

# Scan delay
nmap --scan-delay 1s 172.21.0.3
```

## Firewall/IDS Evasion
```bash
# Fragment packets
nmap -f 172.21.0.3

# Specify MTU
nmap --mtu 32 172.21.0.3

# Decoy scans
nmap -D RND:10 172.21.0.3
nmap -D 172.21.0.101,172.21.0.102,ME 172.21.0.3

# Idle zombie scan
nmap -sI zombie_host 172.21.0.3

# Source port specification
nmap --source-port 53 172.21.0.3

# Randomize target order
nmap --randomize-hosts 172.21.0.0/24

# Spoof MAC address
nmap --spoof-mac 0 172.21.0.3
```

## Output Options
```bash
# Normal output
nmap -oN output.txt 172.21.0.3

# XML output
nmap -oX output.xml 172.21.0.3

# Grepable output
nmap -oG output.gnmap 172.21.0.3

# All formats
nmap -oA output_files 172.21.0.3

# Verbose output
nmap -v 172.21.0.3
nmap -vv 172.21.0.3  # More verbose

# Debug output
nmap -d 172.21.0.3
nmap -dd 172.21.0.3  # More debug info
```

## Common Command Combinations
```bash
# Quick scan
nmap -T4 -F 172.21.0.0/24

# Comprehensive scan
nmap -T4 -A -v 172.21.0.3

# Stealth scan
nmap -sS -T2 -f 172.21.0.3

# UDP scan with version detection
nmap -sU -sV 172.21.0.3

# Network discovery
nmap -sn 172.21.0.0/24

# Vulnerability assessment
nmap -sV --script=vuln 172.21.0.3

# Web server enumeration
nmap -p 80,443 --script=http-* 172.21.0.3
```

## Advanced Examples
```bash
# Scan for specific service
nmap -p 22 --open -sV 172.21.0.0/24

# Find live hosts and open ports
nmap -sn 172.21.0.0/24 | grep -E "Nmap scan report"

# Detect load balancers
nmap --script=http-load-balancer 172.21.0.3

# SSL/TLS information
nmap -p 443 --script=ssl-* 172.21.0.3

# SMB enumeration
nmap -p 445 --script=smb-* 172.21.0.3
```

These commands cover most nmap functionality. Start with basic scans (`nmap -sV -sC target`) and gradually add more options as needed for your specific reconnaissance requirements.


Here's how to discover hosts/IPs in your network using nmap:

## Step 1: Find Your Network Range

First, determine your network range:

### On macOS:
```bash
# Find your network interface and IP
ifconfig | grep "inet "

# Find your default gateway/network
netstat -rn | grep default

# Common result might be: 172.21.0.x with subnet 172.21.0.0/24
```

### On Linux:
```bash
# Find your IP and network
ip addr show
ip route show
```

## Step 2: Host Discovery Methods

### Basic Ping Sweep (Most Common)
```bash
# Discover all live hosts in your network
nmap -sn 172.21.0.0/24

# Example output shows live hosts:
# Nmap scan report for 172.21.0.3
# Host is up (0.0012s latency).
# Nmap scan report for 172.21.0.5
# Host is up (0.0034s latency).
```

### Different Network Ranges
```bash
# Common home networks
nmap -sn 172.21.0.0/24    # 172.21.0.3-254
nmap -sn 192.168.0.0/24    # 192.168.0.1-254
nmap -sn 10.0.0.0/24       # 10.0.0.1-254
nmap -sn 172.16.0.0/24     # 172.16.0.1-254

# Larger network ranges
nmap -sn 192.168.0.0/16    # All 192.168.x.x addresses
nmap -sn 10.0.0.0/8        # All 10.x.x.x addresses
```

### ARP Scan (Local Network Only)
```bash
# More reliable for local network discovery
nmap -sn -PR 172.21.0.0/24

# ARP scan is faster and more accurate for local hosts
```

### Advanced Discovery Options
```bash
# Skip DNS resolution (faster)
nmap -sn -n 172.21.0.0/24

# Include reason for host status
nmap -sn --reason 172.21.0.0/24

# Verbose output
nmap -sn -v 172.21.0.0/24

# Save results to file
nmap -sn 172.21.0.0/24 -oG - | grep Up
```

## Step 3: Enhanced Discovery with Port Scanning

### Discover hosts by scanning common ports
```bash
# Scan for hosts with common open ports
nmap -PS22,80,443,3389 172.21.0.0/24

# TCP SYN ping on multiple ports
nmap -PS21,22,23,25,53,80,110,443,993,995,1723,3389,5900 172.21.0.0/24

# UDP ping (useful for devices that don't respond to TCP)
nmap -PU 172.21.0.0/24
```

### Quick Discovery Commands
```bash
# Fast host discovery
nmap -sn -T4 172.21.0.0/24

# Discover and show MAC addresses (if on same network segment)
sudo nmap -sn 172.21.0.0/24

# Discover hosts that respond to ping
nmap -sn -PE 172.21.0.0/24

# Discover hosts using different ping types
nmap -sn -PE -PP -PM 172.21.0.0/24
```

## Step 4: Get More Information About Discovered Hosts

### Basic information about live hosts
```bash
# Get hostname resolution
nmap -sn 172.21.0.0/24 | grep -E "Nmap scan report|Host is up"

# Get OS information for discovered hosts
nmap -O 172.21.0.3-50

# Get service information on discovered hosts
nmap -sV -p 21,22,23,25,53,80,110,443 172.21.0.0/24
```

### Format output for easier reading
```bash
# Extract just the IP addresses
nmap -sn 172.21.0.0/24 | grep "Nmap scan report" | cut -d " " -f 5

# Show live hosts with their status
nmap -sn 172.21.0.0/24 | awk '/Nmap scan report/{print $5} /Host is up/{print "  Status: UP"}'

# Save live IPs to a file
nmap -sn 172.21.0.0/24 | grep "Nmap scan report" | cut -d " " -f 5 > live_hosts.txt
```

## Practical Examples

### Discover Docker Network Hosts
```bash
# For Docker containers (common Docker network)
nmap -sn 172.17.0.0/16
nmap -sn 172.18.0.0/16

# Check your Docker network specifically
docker network inspect bridge | grep Subnet
# Then scan that subnet with nmap
```

### Complete Network Assessment
```bash
# Step 1: Discover all hosts
nmap -sn 172.21.0.0/24

# Step 2: Quick port scan on discovered hosts
nmap -F 172.21.0.3-50  # Fast scan on first 50 IPs

# Step 3: Detailed scan on interesting hosts
nmap -sV -sC 172.21.0.3
```

### Troubleshooting Discovery Issues
```bash
# If hosts don't show up, try different methods:

# 1. Skip ping (treat all hosts as up)
nmap -Pn -p 80,443 172.21.0.0/24

# 2. Use ARP scan (most reliable for local network)
nmap -sn -PR 172.21.0.0/24

# 3. Try different ping types
nmap -sn -PE -PA80,443 172.21.0.0/24

# 4. Increase timing (faster scan)
nmap -sn -T4 --min-parallelism 100 172.21.0.0/24
```

**Start with**: `nmap -sn 172.21.0.0/24` (replace with your actual network range) - this is the most common and effective method for discovering hosts in your network.

To check the container network IP range and determine if it's using a bridge interface, you can use several Docker commands:

## Check Docker Network Configuration

### 1. List all Docker networks:
```bash
docker network ls
```

### 2. Inspect the default bridge network:
```bash
docker network inspect bridge
```
This will show you the subnet, gateway, and IP range for the default bridge network.

### 3. Check all networks with detailed info:
```bash
docker network inspect $(docker network ls -q)
```

### 4. For a running container, check its network settings:
```bash
# Replace <container_name> with your container name
docker inspect <container_name> | grep -A 20 "NetworkSettings"
```

### 5. Get container IP address directly:
```bash
docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' <container_name>
```

### 6. Check network interface inside the container:
```bash
# Run this inside the container
docker exec -it <container_name> ip addr show
# or
docker exec -it <container_name> ifconfig
```

## Common Docker Network Information

**Default Bridge Network:**
- Usually uses IP range: `172.17.0.0/16`
- Gateway: `172.17.0.1`
- Interface: `docker0` on the host

**To check host bridge interface:**
```bash
# On the host system
ip addr show docker0
# or
ifconfig docker0
```

**To see which containers are using which networks:**
```bash
docker ps --format "table {{.Names}}\t{{.Networks}}"
```

The default behavior is that containers use the bridge network unless specified otherwise with `--network` flag when running the container.