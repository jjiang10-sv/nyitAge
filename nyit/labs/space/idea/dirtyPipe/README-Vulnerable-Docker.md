# Vulnerable Docker Environments for Dirty Pipe (CVE-2022-0847)

This directory contains multiple Docker configurations for testing the Dirty Pipe vulnerability in controlled environments.

## Overview

**Dirty Pipe (CVE-2022-0847)** is a local privilege escalation vulnerability that affects Linux kernels from 5.8 to 5.15.25 (and other versions before patches). It allows overwriting data in arbitrary read-only files.

## ⚠️ Security Warning

**FOR EDUCATIONAL USE ONLY**  
These containers may contain vulnerable kernels. Use only in isolated, non-production environments for security research and education.

## Available Configurations

### 1. Ubuntu 20.04 (Most Compatible)
**File:** `Dockerfile.vulnerable`  
**Compose:** `docker-compose.vulnerable.yml`

- **Base:** Ubuntu 20.04 LTS
- **Expected Kernel:** Various 5.4.x/5.8.x versions
- **Likelihood:** May or may not be vulnerable (depends on host kernel)
- **Best for:** General compatibility testing

### 2. Ubuntu 21.04 (Higher Chance of Vulnerability)
**File:** `Dockerfile.ubuntu2104`

- **Base:** Ubuntu 21.04 (Hirsute Hippo) 
- **Expected Kernel:** 5.11.x series
- **Likelihood:** Higher chance of vulnerable kernel
- **Note:** Uses old-releases.ubuntu.com (21.04 reached EOL)

### 3. Current Setup (Patched)
**File:** `docker-compose.yml`

- **Base:** iridium191/cve-2022-0847 image
- **Kernel:** 5.15.49-linuxkit-pr (patched)
- **Status:** ❌ Not vulnerable (demonstration only)

## Quick Start

### Option 1: Ubuntu 20.04 Environment
```bash
# Build and run
chmod +x build-vulnerable.sh
./build-vulnerable.sh

# Connect
docker exec -it dirty-pipe-vulnerable /bin/bash

# Test as normal user
su - normaluser  # password: password
cd /exploits
./test_dirtypipe
```

### Option 2: Ubuntu 21.04 Environment
```bash
# Build Ubuntu 21.04 version
docker build -f Dockerfile.ubuntu2104 -t dirty-pipe-2104 .

# Run container
docker run -it --name test-2104 dirty-pipe-2104

# In another terminal, connect
docker exec -it test-2104 /bin/bash
su - normaluser
cd /exploits
./test_dirtypipe
```

## Testing Process

### Step 1: Environment Check
```bash
# Check kernel version
uname -r

# Check OS version  
cat /etc/os-release

# Check users
id testuser && id normaluser
```

### Step 2: Permission Verification
```bash
# As non-root user, test basic file permissions
echo "test" > /tmp/readonly.txt
chmod 444 /tmp/readonly.txt
echo "modified" > /tmp/readonly.txt  # Should fail
```

### Step 3: Vulnerability Test
```bash
# Run the built-in test
./test_dirtypipe

# Or run specific exploits
./exploit_fixed  # If available
```

## Expected Results

### Vulnerable Environment ✅
- File permissions properly enforced
- Normal write to read-only files fails
- Dirty Pipe exploit succeeds in modifying read-only files
- Test shows: "🚨 SUCCESS: Dirty Pipe vulnerability CONFIRMED!"

### Patched Environment ❌
- File permissions properly enforced  
- Normal write to read-only files fails
- Dirty Pipe exploit fails to modify files
- Test shows: "✅ SAFE: Dirty Pipe exploit failed (kernel patched)"

### Over-Privileged Environment ⚠️
- File permissions NOT enforced (can write to "read-only" files normally)
- This indicates container is running with excessive privileges
- Dirty Pipe test is meaningless in this scenario

## Troubleshooting

### Container Uses Host Kernel
Docker containers share the host kernel. If your host is:
- **macOS/Windows:** Uses VM with potentially newer kernel
- **Linux with updated kernel:** Will be patched
- **Linux with older kernel:** May be vulnerable

### Building Issues
```bash
# If build fails with old Ubuntu repositories
docker build --no-cache -f Dockerfile.ubuntu2104 -t dirty-pipe-2104 .

# Check Docker BuildKit
docker buildx version
```

### Permission Issues
```bash
# Ensure files are readable
chmod 644 exploit*.c

# Check Docker daemon
docker info
```

## File Structure

```
dirtyPipe/
├── Dockerfile.vulnerable          # Ubuntu 20.04 setup
├── Dockerfile.ubuntu2104          # Ubuntu 21.04 setup  
├── docker-compose.vulnerable.yml  # Compose for Ubuntu 20.04
├── docker-compose.yml             # Original (patched) setup
├── build-vulnerable.sh            # Build script
├── exploit.c                      # Original exploit
├── exploit_fixed.c               # Modified exploit
├── install-vulnerable-kernel.sh   # Kernel downgrade script
├── multipass-setup.sh             # VM alternative
└── README-Vulnerable-Docker.md    # This file
```

## Alternative: Multipass VM

For guaranteed kernel control, use the Multipass setup:
```bash
# See MULTIPASS-GUIDE.md for VM-based testing
./multipass-setup.sh
```

## Security Research Notes

### Why Docker May Not Work
1. **Shared Kernel:** Container uses host kernel, which is likely patched
2. **Privileged Mode:** Can bypass normal permission checks
3. **Container Runtime:** May have additional protections

### Better Alternatives
1. **Virtual Machines:** Full kernel control (Multipass, VirtualBox, VMware)
2. **Physical Hardware:** Dedicated test machine with old kernel
3. **Cloud Instances:** Spin up older OS versions

## References

- [CVE-2022-0847 Details](https://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2022-0847)
- [Original Disclosure](https://dirtypipe.cm4all.com/)
- [Kernel Patches](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/commit/?id=9d2231c5d74e13b2a0546fee6737ee4446017903)

## License & Legal

This code is for educational purposes only. Users are responsible for complying with all applicable laws and regulations. Do not use on systems you do not own or have explicit permission to test. 