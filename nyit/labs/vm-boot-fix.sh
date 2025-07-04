#!/bin/bash

echo "🔧 VM Boot Fix - Matching Original Hardware Configuration"
echo ""

# Stop any running VMs
docker stop security-lab-vm security-lab-console security-lab-working 2>/dev/null || true
docker rm security-lab-vm security-lab-console security-lab-working 2>/dev/null || true

echo "Trying different hardware configurations..."
echo ""

# Configuration 1: SATA/AHCI (matches original VirtualBox)
echo "=== Attempt 1: SATA/AHCI Controller (Original VBox Config) ==="
docker run -it --rm \
  --name vm-sata-test \
  --privileged \
  -v $(pwd)/local:/data \
  alpine:latest sh -c "
  apk add --no-cache qemu-system-x86_64 && 
  timeout 30 qemu-system-x86_64 \
    -machine pc \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -drive file=/data/server.qcow2,format=qcow2,if=none,id=hd0 \
    -device ahci,id=ahci \
    -device ide-hd,drive=hd0,bus=ahci.0 \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -device rtl8139,netdev=net0 \
    -nographic \
    -boot c
" && echo "SATA test completed" || echo "SATA test failed/timeout"

echo ""
echo "=== Attempt 2: Different Machine Type ==="
docker run -it --rm \
  --name vm-machine-test \
  --privileged \
  -v $(pwd)/local:/data \
  alpine:latest sh -c "
  apk add --no-cache qemu-system-x86_64 && 
  timeout 30 qemu-system-x86_64 \
    -machine pc-i440fx-2.12 \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -hda /data/server.qcow2 \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -device rtl8139,netdev=net0 \
    -nographic \
    -boot c
" && echo "Machine type test completed" || echo "Machine type test failed/timeout"

echo ""
echo "=== Attempt 3: Original VMDK File ==="
docker run -it --rm \
  --name vm-vmdk-test \
  --privileged \
  -v $(pwd)/local:/data \
  alpine:latest sh -c "
  apk add --no-cache qemu-system-x86_64 && 
  timeout 30 qemu-system-x86_64 \
    -machine pc \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -drive file=/data/vm-disk001.vmdk,format=vmdk,if=ide \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -device rtl8139,netdev=net0 \
    -nographic \
    -boot c
" && echo "VMDK test completed" || echo "VMDK test failed/timeout"

echo ""
echo "📊 Boot Test Results:"
echo "Check above for which configuration shows more boot progress" 