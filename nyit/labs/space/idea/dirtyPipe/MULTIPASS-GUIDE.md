# Multipass VM Setup for Dirty Pipe (CVE-2022-0847) Testing

## Overview
This guide shows how to use Multipass to create a VM environment for testing the Dirty Pipe vulnerability safely.

## Prerequisites
- Multipass installed on your system
- Basic knowledge of Linux commands
- The exploit files from this directory

## Quick Start

### 1. Create and Setup VM
```bash
# Navigate to the dirty pipe directory
cd /path/to/dirtyPipe

# Create VM (already done)
multipass launch 22.04 --name dirty-pipe-vm --memory 2G --disk 15G --cpus 2

# Check VM status
multipass info dirty-pipe-vm

# Transfer exploit files
multipass transfer exploit_fixed.c dirty-pipe-vm:/home/ubuntu/
multipass transfer exploit.c dirty-pipe-vm:/home/ubuntu/
multipass transfer install-vulnerable-kernel.sh dirty-pipe-vm:/home/ubuntu/
```

### 2. Setup VM Environment
```bash
# Setup development environment and test user
multipass exec dirty-pipe-vm -- bash -c "
sudo apt update && sudo apt install -y gcc build-essential
sudo useradd -m -s /bin/bash testuser
echo 'testuser:password' | sudo chpasswd
sudo usermod -aG sudo testuser
cd /home/ubuntu && gcc exploit_fixed.c -o exploit_fixed
"
```

### 3. Connect to VM
```bash
# Shell into the VM
multipass shell dirty-pipe-vm

# Once inside VM, switch to testuser
su - testuser
# Password: password
```

### 4. Test Current Kernel
```bash
# Check kernel version
uname -r

# Test if current kernel is vulnerable
cd /home/ubuntu
./exploit_fixed

# Check if testuser can normally modify /etc/passwd (should fail)
echo "test" >> /etc/passwd
```

### 5. Install Vulnerable Kernel (if needed)
```bash
# If the current kernel is patched, try installing older kernel
chmod +x /home/ubuntu/install-vulnerable-kernel.sh
sudo /home/ubuntu/install-vulnerable-kernel.sh

# Reboot to use older kernel
sudo reboot

# Reconnect after reboot
# multipass shell dirty-pipe-vm
# su - testuser
```

## Current VM Status

**VM Created:** ✅ dirty-pipe-vm (Running)
- **IP:** 192.168.64.12
- **OS:** Ubuntu 22.04.5 LTS  
- **Kernel:** 5.15.0-142-generic (patched)
- **Users:** ubuntu (admin), testuser (test user with password: password)
- **Exploits:** Compiled and ready in /home/ubuntu/

## Testing Steps

### Step 1: Basic Environment Check
```bash
multipass shell dirty-pipe-vm
su - testuser
whoami && id
uname -r
```

### Step 2: Test File Permissions
```bash
# Test if permissions are enforced
echo "test" > /tmp/test.txt
chmod 444 /tmp/test.txt
echo "modified" > /tmp/test.txt  # Should fail with Permission denied
```

### Step 3: Test /etc/passwd Access
```bash
# Test if testuser can modify /etc/passwd normally (should fail)
echo "test" >> /etc/passwd  # Should fail with Permission denied
```

### Step 4: Run Dirty Pipe Exploit
```bash
cd /home/ubuntu
./exploit_fixed

# Check if passwd was modified
head -1 /etc/passwd

# Test new root password (if exploit worked)
su root
# Password: aaron
```

## Expected Results

### If Kernel is Patched (Current Status)
- ✅ File permissions enforced correctly
- ✅ testuser cannot modify /etc/passwd normally  
- ❌ Dirty Pipe exploit fails to modify files
- ❌ `su root` with "aaron" fails

### If Kernel is Vulnerable
- ✅ File permissions enforced correctly
- ✅ testuser cannot modify /etc/passwd normally
- ✅ Dirty Pipe exploit successfully modifies /etc/passwd
- ✅ `su root` with "aaron" succeeds

## Troubleshooting

### Kernel Too New
The current kernel (5.15.0-142-generic) is patched. To test the vulnerability:

1. **Install older kernel:**
   ```bash
   sudo /home/ubuntu/install-vulnerable-kernel.sh
   sudo reboot
   ```

2. **Use a different base image:**
   ```bash
   # Delete current VM and try older Ubuntu
   multipass delete dirty-pipe-vm
   multipass purge
   
   # Try Ubuntu 20.04 (if available)
   multipass launch 20.04 --name vulnerable-vm --memory 2G --disk 10G
   ```

### Exploit Not Working
If exploit compiles but doesn't work:
- Check kernel version: `uname -r`
- Verify you're running as testuser (not root)
- Ensure file permissions are enforced
- Try the simple file test first

## Cleanup

```bash
# Stop and delete VM
multipass stop dirty-pipe-vm
multipass delete dirty-pipe-vm
multipass purge

# List remaining VMs
multipass list
```

## Alternative: Using Older Ubuntu

If you need a definitely vulnerable system, try downloading older Ubuntu ISO manually:

```bash
# Create VM from older Ubuntu (requires manual ISO)
multipass launch --disk 10G --memory 2G --name old-ubuntu file://ubuntu-20.04.3-live-server-amd64.iso
```

## Security Note

⚠️ **Important:** Only use this setup for educational purposes in isolated environments. The Dirty Pipe vulnerability can be dangerous if exploited on production systems.

## Files in This Directory

- `exploit.c` - Original exploit
- `exploit_fixed.c` - Modified exploit with better output
- `install-vulnerable-kernel.sh` - Script to install older kernel
- `multipass-setup.sh` - Automated setup script
- `MULTIPASS-GUIDE.md` - This guide 