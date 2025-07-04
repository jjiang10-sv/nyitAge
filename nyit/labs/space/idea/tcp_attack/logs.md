To check the default network interface and all available interfaces in your Ubuntu 24 VM, use these commands:

### 1. **Check Default Network Interface**
```bash
ip route show default | awk '/default/ {print $5}'
```
**Example output:**
```bash
enp0s5
```

### 2. **List All Network Interfaces**
```bash
ip -o link show | awk -F': ' '{print $2}'
```
**Example output:**
```bash
lo
enp0s5
docker0 (if Docker is installed)
```

---

### Alternative Methods:
#### **Using `ip addr` (Detailed Info)**
```bash
ip addr
```
Shows:
- `lo`: Loopback interface (127.0.0.1)
- `enp0s*`/`eth0`: Physical/Ethernet interfaces
- `wlp0s*`: Wireless interfaces (if applicable)
- `docker0`/`veth*`: Docker interfaces (if installed)

#### **Using `netstat` (Deprecated but useful)**
```bash
sudo apt install net-tools  # Install if needed
netstat -i
```

#### **Using `ifconfig` (Classic tool)**
```bash
sudo apt install net-tools  # Install if missing
ifconfig -a
```

---

### Key Notes:
1. **Default Interface Logic:**
   - The command `ip route show default` identifies the interface used for internet traffic.
   - The `awk` filter extracts the interface name (e.g., `enp0s5`).

2. **Common Interface Types:**
   - `lo`: Loopback (internal)
   - `en*`/`eth*`: Ethernet (physical)
   - `wl*`: Wireless
   - `docker0`/`br-*`: Docker/bridge interfaces

3. **Special Cases:**
   - **No default route?** Run `sudo dhclient enp0s5` (replace `enp0s5` with your interface) to get DHCP IP.
   - **VPNs/Tunnels:** Look for `tun0` (OpenVPN) or `wg0` (WireGuard).

---

### Example Workflow:
```bash
# 1. List interfaces
$ ip -o link show | awk -F': ' '{print $2}'
lo
enp0s5

# 2. Identify default interface
$ ip route show default
default via 192.168.64.1 dev enp0s5 proto static

# 3. Verify with IP configuration
$ ip addr show enp0s5
2: enp0s5: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc fq_codel state UP group default qlen 1000
    link/ether 00:16:3e:ab:cd:ef brd ff:ff:ff:ff:ff:ff
    inet 192.168.64.5/24 brd 192.168.64.255 scope global enp0s5
       valid_lft forever preferred_lft forever
```

> **Note:** In Multipass VMs, the primary interface is typically named `enp0s5` or `eth0`. The `lo` interface exists on all systems but isn't used for external traffic.

To capture the **last TCP packet** with specific details (like source/destination ports, sequence/acknowledgment numbers, etc.) using `tshark`, follow these steps. We'll focus on capturing packets until interrupted and then extract the last matching TCP packet.

---

### **1. Capture Packets and Save to a File**
Start capturing packets to a temporary file. Press `Ctrl+C` when you want to stop:
```bash
tshark -i eth0 -f "tcp" -w /tmp/capture.pcap
tshark -i eth0 -f "tcp" 
```
- `-i eth0`: Replace `eth0` with your interface (use `ip link` to list interfaces).
- `-f "tcp"`: Capture only TCP traffic.
- `-w /tmp/capture.pcap`: Save packets to a file.

---

### **2. Extract the Last TCP Packet from the Capture**
Read the saved file and display the **last packet** in verbose format:
```bash
tshark -r /tmp/capture.pcap -V -c 1 -Y "tcp" | tail -n 25
```
- `-r /tmp/capture.pcap`: Read the captured file.
- `-V`: Verbose output (shows full protocol details).
- `-c 1`: Display only the last packet.
- `-Y "tcp"`: Filter for TCP packets.
- `tail -n 25`: Show the last 25 lines (adjust as needed).

---

### **3. Filter for Specific TCP Details (Optional)**
To get details **exactly like your example** (e.g., Telnet traffic with `PSH+ACK` flags), add filters:
```bash
tshark -r /tmp/capture.pcap -V -c 1 -Y "tcp.port==23 && tcp.flags==0x018" | grep -A 20 "Transmission Control Protocol"
```
- `tcp.port==23`: Filter for Telnet traffic (source or destination port 23).
- `tcp.flags==0x018`: Filter for packets with `PSH+ACK` flags.
- `grep -A 20`: Show 20 lines after the "Transmission Control Protocol" header.

---

### **Example Output**
You'll see details like:
```plaintext
Transmission Control Protocol, Src Port: 23, Dst Port: 59498, Seq: 199921802, Ack: 1577076274, Len: 1
    Source Port: 23
    Destination Port: 59498
    [Stream index: 0]
    [TCP Segment Len: 1]
    Sequence number: 199921802
    [Next sequence number: 199921803]
    Acknowledgment number: 1577076274
    Header Length: 32 bytes
    Flags: 0x018 (PSH, ACK)
    Window size value: 227
    [Calculated window size: 29056]
    [Window size scaling factor: 128]
    Checksum: 0x953c [unverified]
    [Checksum Status: Unverified]
    Urgent pointer: 0
    Options: (12 bytes), No-Operation (NOP), No-Operation (NOP), Timestamps
    [SEQ/ACK analysis]
```

---

### **4. One-Liner for Real-Time Capture**
Capture packets and display the **last TCP packet** immediately after stopping (combines steps 1 and 2):
```bash
tshark -i eth0 -f "tcp" -w - > /tmp/capture.pcap && tshark -r /tmp/capture.pcap -V -c 1 -Y "tcp" | tail -n 25
```
- `-w -`: Write to stdout (redirected to a file).
- Runs the capture and analysis sequentially after `Ctrl+C`.

---

### **Key Notes**
1. **Interface**: Replace `eth0` with your active interface (e.g., `enp0s5`).
2. **Telnet Filter**: Use `tcp.port==23` to focus on Telnet traffic.
3. **Flags**:
   - `PSH+ACK` = `0x018` (24 in decimal).
   - Filter with `tcp.flags==0x018`.
4. **Cleanup**: Remove the temporary capture file afterward:
   ```bash
   rm /tmp/capture.pcap
   ```

---

### **Alternative: Capture Specific Fields**
To extract only specific fields (e.g., ports, seq/ack numbers):
```bash
tshark -r /tmp/capture.pcap -T fields -e tcp.srcport -e tcp.dstport -e tcp.seq -e tcp.ack -e tcp.flags -e tcp.len -c 1
```
Output:
```plaintext
23  59498  199921802  1577076274  0x018  1
```

This gives you precise control over the details you need!

multipass shell nyit-lab

Here are the commands to remove all Docker images, containers, and networks:

## **Complete Docker Cleanup (Nuclear Option)**

**Remove everything at once:**
```bash
docker system prune -a --volumes
```
This removes:
- All stopped containers
- All networks not used by at least one container
- All images without at least one container associated to them
- All build cache
- All volumes

## **Step-by-Step Cleanup**

**1. Stop and remove all containers:**
```bash
# Stop all running containers
docker stop $(docker ps -aq)

# Remove all containers
docker rm $(docker ps -aq)
```

**2. Remove all images:**
```bash
# Remove all images
docker rmi $(docker images -q)

# Force remove all images (if some are in use)
docker rmi -f $(docker images -q)
```

**3. Remove all networks:**
```bash
# Remove all custom networks
docker network prune

# Or remove all networks (except default ones)
docker network rm $(docker network ls -q)
```

**4. Remove all volumes:**
```bash
docker volume prune
```

## **Docker Compose Specific Cleanup**

If you want to clean up only the current Docker Compose project:
```bash
# Stop and remove containers, networks, and volumes for current project
docker-compose down -v --remove-orphans

# Also remove images created by docker-compose
docker-compose down -v --rmi all --remove-orphans
```

## **Verify Cleanup**
```bash
# Check containers
docker ps -a

# Check images  
docker images

# Check networks
docker network ls

# Check volumes
docker volume ls
```

**Warning:** These commands will remove ALL Docker data. Make sure you don't have any important containers or data you want to keep!
