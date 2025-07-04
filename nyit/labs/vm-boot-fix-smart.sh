#!/bin/bash

echo "🔧 VM Boot Fix - Smart Testing (Stops on First Success)"
echo ""

# Stop any running VMs
docker stop security-lab-vm security-lab-console security-lab-working 2>/dev/null || true
docker rm security-lab-vm security-lab-console security-lab-working 2>/dev/null || true

echo "Testing configurations until one works..."
echo ""

# Function to test configuration
test_config() {
    local config_name="$1"
    local docker_cmd="$2"
    
    echo "=== Testing: $config_name ==="
    if eval "$docker_cmd"; then
        echo "✅ SUCCESS: $config_name works!"
        echo ""
        echo "🎯 Working Configuration Found:"
        echo "Use this configuration for your final setup."
        return 0  # Success - stop testing
    else
        echo "❌ FAILED: $config_name didn't work"
        echo ""
        return 1  # Failed - continue testing
    fi
}

# Test 1: SATA/AHCI (Original VBox Config)
test_config "SATA/AHCI Controller" \
'docker run -it --rm --name vm-sata-test --privileged -v $(pwd)/local:/data alpine:latest sh -c "
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
"' && exit 0

# Test 2: Different Machine Type (only if Test 1 failed)
test_config "PC-i440fx Machine Type" \
'docker run -it --rm --name vm-machine-test --privileged -v $(pwd)/local:/data alpine:latest sh -c "
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
"' && exit 0

# Test 3: Original VMDK (only if Tests 1&2 failed)
test_config "Original VMDK File" \
'docker run -it --rm --name vm-vmdk-test --privileged -v $(pwd)/local:/data alpine:latest sh -c "
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
"' && exit 0

echo "❌ All configurations failed"
echo "The VM image may be corrupted or incompatible"
echo "Consider:"
echo "  1. Re-converting the OVA file"
echo "  2. Using VirtualBox directly"
echo "  3. Checking original VM requirements" 