#!/bin/bash

echo "🔍 VM Boot Debug - Detailed Analysis"
echo ""

# Stop any running VMs
docker stop security-lab-vm security-lab-console security-lab-working vm-sata-test vm-machine-test vm-vmdk-test 2>/dev/null || true
docker rm security-lab-vm security-lab-console security-lab-working vm-sata-test vm-machine-test vm-vmdk-test 2>/dev/null || true

echo "Starting VM with maximum debug information..."
echo "This will show:"
echo "  - BIOS POST messages"
echo "  - Boot sector loading"
echo "  - Bootloader messages"
echo "  - Kernel loading (if it gets there)"
echo ""

docker run -it --rm \
  --name vm-debug \
  --privileged \
  -v $(pwd)/local:/data \
  alpine:latest sh -c "
  apk add --no-cache qemu-system-x86_64 && 
  echo 'Starting VM with debug output...' &&
  qemu-system-x86_64 \
    -machine pc \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -drive file=/data/server.qcow2,format=qcow2,if=ide,cache=writeback \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -device rtl8139,netdev=net0 \
    -nographic \
    -boot c \
    -d cpu_reset,int,pcall \
    -D /tmp/qemu-debug.log
"

echo ""
echo "Debug Information:"
echo "  If VM shows more than 'Booting from Hard Disk...', note what appears"
echo "  Look for error messages, kernel panic, or filesystem errors"
echo "  Press Ctrl+C if it hangs for more than 2 minutes" 