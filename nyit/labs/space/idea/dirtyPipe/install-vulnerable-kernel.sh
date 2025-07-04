#!/bin/bash

# Script to install a potentially vulnerable kernel for Dirty Pipe testing
# Run this inside the Ubuntu VM

echo "=== Installing Potentially Vulnerable Kernel ==="

# Check current kernel
echo "Current kernel: $(uname -r)"

# Update package lists
sudo apt update

# Install required tools
sudo apt install -y wget

# Download older kernel that might be vulnerable
# Using 5.15.0-25 which was released before major Dirty Pipe patches
KERNEL_VERSION="5.15.0-25.25"
KERNEL_URL="http://archive.ubuntu.com/ubuntu/pool/main/l/linux"

echo "Downloading kernel ${KERNEL_VERSION}..."

# Download kernel packages
wget ${KERNEL_URL}/linux-image-5.15.0-25-generic_${KERNEL_VERSION}_amd64.deb
wget ${KERNEL_URL}/linux-headers-5.15.0-25-generic_${KERNEL_VERSION}_amd64.deb
wget ${KERNEL_URL}/linux-headers-5.15.0-25_${KERNEL_VERSION}_all.deb

# Install the downloaded packages
echo "Installing older kernel packages..."
sudo dpkg -i linux-headers-5.15.0-25_${KERNEL_VERSION}_all.deb
sudo dpkg -i linux-headers-5.15.0-25-generic_${KERNEL_VERSION}_amd64.deb
sudo dpkg -i linux-image-5.15.0-25-generic_${KERNEL_VERSION}_amd64.deb

# Update grub to make sure older kernel is available at boot
sudo update-grub

echo "=== Kernel Installation Complete ==="
echo "Available kernels:"
ls /boot/vmlinuz-* | sort

echo ""
echo "To boot into the older kernel:"
echo "1. Reboot the system: sudo reboot"
echo "2. During boot, select 'Advanced options for Ubuntu'"
echo "3. Choose kernel 5.15.0-25-generic"
echo ""
echo "Or set it as default:"
echo "sudo sed -i 's/GRUB_DEFAULT=0/GRUB_DEFAULT=\"1>2\"/' /etc/default/grub"
echo "sudo update-grub"
echo "sudo reboot"

# Clean up downloaded files
rm -f *.deb

echo "Reboot required to use the new kernel!" 