#!/bin/bash

# Multipass Setup Script for Dirty Pipe (CVE-2022-0847) Testing
# This script sets up a VM environment for testing the vulnerability

echo "=== Setting up Multipass VM for Dirty Pipe Testing ==="

# Step 1: Create VM with older Ubuntu if possible
echo "Creating VM..."
multipass launch 22.04 --name dirty-pipe-lab --memory 2G --disk 15G --cpus 2

# Step 2: Wait for VM to be ready
echo "Waiting for VM to boot..."
sleep 10

# Step 3: Transfer exploit files to VM
echo "Transferring exploit files..."
multipass transfer exploit_fixed.c dirty-pipe-lab:/home/ubuntu/
multipass transfer exploit.c dirty-pipe-lab:/home/ubuntu/
multipass transfer compile.sh dirty-pipe-lab:/home/ubuntu/

# Step 4: Setup VM environment
echo "Setting up VM environment..."
multipass exec dirty-pipe-lab -- bash -c '
# Update system
sudo apt update

# Install required packages
sudo apt install -y gcc build-essential linux-headers-generic

# Check current kernel
echo "Current kernel version:"
uname -r

# Check available kernels
echo "Available kernels:"
apt list --installed | grep linux-image

# Try to install older kernel packages (if available)
echo "Attempting to install older kernel..."
sudo apt install -y linux-image-5.15.0-25-generic linux-headers-5.15.0-25-generic 2>/dev/null || echo "Older kernel not available in repos"

# Create test user for privilege escalation testing
sudo useradd -m -s /bin/bash testuser
echo "testuser:password" | sudo chpasswd
sudo usermod -aG sudo testuser

# Set permissions
chmod +x /home/ubuntu/compile.sh
chown ubuntu:ubuntu /home/ubuntu/*.c /home/ubuntu/*.sh

# Compile exploits
cd /home/ubuntu
gcc exploit.c -o exploit
gcc exploit_fixed.c -o exploit_fixed

echo "=== VM Setup Complete ==="
echo "Kernel: $(uname -r)"
echo "Test user: testuser (password: password)"
echo "Exploit files ready in /home/ubuntu/"
'

echo "=== VM Information ==="
multipass info dirty-pipe-lab

echo ""
echo "=== Usage Instructions ==="
echo "1. Connect to VM:"
echo "   multipass shell dirty-pipe-lab"
echo ""
echo "2. Switch to test user:"
echo "   su - testuser"
echo "   # Password: password"
echo ""
echo "3. Test exploit:"
echo "   cd /home/ubuntu"
echo "   ./exploit_fixed"
echo ""
echo "4. If kernel is patched, you can try downloading a vulnerable kernel:"
echo "   # Inside VM as ubuntu user:"
echo "   wget http://archive.ubuntu.com/ubuntu/pool/main/l/linux/linux-image-5.15.0-25-generic_5.15.0-25.25_amd64.deb"
echo "   sudo dpkg -i linux-image-5.15.0-25-generic_5.15.0-25.25_amd64.deb"
echo "   sudo reboot"
echo ""
echo "5. Clean up when done:"
echo "   multipass delete dirty-pipe-lab"
echo "   multipass purge" 