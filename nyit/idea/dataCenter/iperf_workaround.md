# Mininet iperf Limitation & Workaround

## The Problem

Even with bridges properly configured, Mininet's built-in `iperf h00 h40` command hangs because:

1. Mininet's iperf implementation has internal hostname resolution that's hard to override
2. The bridge setup changes IPs after Mininet initializes
3. Mininet's internal data structures don't get properly updated

## The Working Solution

**Use IP addresses directly instead of hostnames:**

```bash
# In Mininet CLI - these WORK:
mininet> h40 iperf -s -p 5001 &
mininet> h00 iperf -c 10.0.0.9 -t 10

# This is MORE RELIABLE than trying to make 'iperf h00 h40' work!
```

## Complete Test Script

```bash
# RED PATH (h00 ↔ h40)
mininet> h40 iperf -s -p 5001 &
mininet> h00 iperf -c 10.0.0.9 -t 10
mininet> sh killall iperf

# GREEN PATH (h00 ↔ h50) 
mininet> h50 iperf -s -p 5001 &
mininet> h00 iperf -c 10.0.0.11 -t 10
mininet> sh killall iperf

# BLUE PATH (h20 ↔ h30)
mininet> h30 iperf -s -p 5001 &
mininet> h20 iperf -c 10.0.0.7 -t 10
mininet> sh killall iperf

# PURPLE PATH (h60 ↔ h61)
mininet> h61 iperf -s -p 5001 &
mininet> h60 iperf -c 10.0.0.13 -t 10
mininet> sh killall iperf

# BLACK PATH (h60 ↔ h70)
mininet> h70 iperf -s -p 5001 &
mininet> h60 iperf -c 10.0.0.15 -t 10
mininet> sh killall iperf
```

## Why This Works

1. ✅ **Explicit server start** - No reliance on Mininet's auto-start
2. ✅ **Direct IP addressing** - No hostname resolution needed  
3. ✅ **Full control** - You see exactly what's running
4. ✅ **More reliable** - Works with any network configuration

## Host IP Reference

From assignment1_fixed.py, hosts have these IPs on br0:

| Host | IP | Host | IP |
|------|-------|------|-------|
| h00 | 10.0.0.1 | h40 | 10.0.0.9 |
| h01 | 10.0.0.2 | h41 | 10.0.0.10 |
| h10 | 10.0.0.3 | h50 | 10.0.0.11 |
| h11 | 10.0.0.4 | h51 | 10.0.0.12 |
| h20 | 10.0.0.5 | h60 | 10.0.0.13 |
| h21 | 10.0.0.6 | h61 | 10.0.0.14 |
| h30 | 10.0.0.7 | h70 | 10.0.0.15 |
| h31 | 10.0.0.8 | h71 | 10.0.0.16 |

## Verify IP Addresses

```bash
# Check any host's IP:
mininet> h00 ip addr show br0 | grep inet
mininet> py h40.IP()
```

## Conclusion

**The manual iperf method with IP addresses is more reliable than trying to make Mininet's automatic `iperf h00 h40` work with modified network configurations.**

This is actually the **standard approach** for Mininet testing when you've customized the network setup beyond Mininet's defaults.