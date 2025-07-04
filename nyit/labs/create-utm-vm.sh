#!/bin/bash

# UTM VM Creation Script from OVA
# Usage: ./create-utm-vm.sh <ova-file> <vm-name>

set -e

OVA_FILE="$1"
VM_NAME="$2"

if [ -z "$OVA_FILE" ] || [ -z "$VM_NAME" ]; then
    echo "Usage: $0 <ova-file> <vm-name>"
    echo "Example: $0 Server.ova SecurityLab"
    exit 1
fi

echo "🚀 Creating UTM VM from OVA file..."

# Check if OVA file exists
if [ ! -f "$OVA_FILE" ]; then
    echo "❌ Error: OVA file '$OVA_FILE' not found"
    exit 1
fi

# Create working directory
WORK_DIR="/tmp/utm-vm-creation-$$"
mkdir -p "$WORK_DIR"
cd "$WORK_DIR"

echo "📦 Extracting OVA file..."
tar -xf "$OVA_FILE"

# Find VMDK file
VMDK_FILE=$(find . -name "*.vmdk" | head -1)
if [ -z "$VMDK_FILE" ]; then
    echo "❌ Error: No VMDK file found in OVA"
    exit 1
fi

echo "🔄 Converting VMDK to QCOW2..."
QCOW2_FILE="${VM_NAME}.qcow2"
qemu-img convert -O qcow2 "$VMDK_FILE" "$QCOW2_FILE"

# Get VM info from OVF file
OVF_FILE=$(find . -name "*.ovf" | head -1)
if [ -n "$OVF_FILE" ]; then
    echo "📋 Reading VM configuration from OVF..."
    # Extract memory and CPU info (basic parsing)
    MEMORY=$(grep -o 'rasd:VirtualQuantity>[0-9]*' "$OVF_FILE" | grep -o '[0-9]*' | head -1)
    CPU_COUNT=$(grep -o 'rasd:VirtualQuantity>[0-9]*' "$OVF_FILE" | grep -o '[0-9]*' | tail -1)
else
    echo "⚠️  No OVF file found, using default settings"
    MEMORY="4096"
    CPU_COUNT="2"
fi

# Move QCOW2 to UTM directory
UTM_VM_DIR="$HOME/Library/Group Containers/WDNLXAD4W8.com.utmapp.UTM/Data/Documents"
if [ ! -d "$UTM_VM_DIR" ]; then
    mkdir -p "$UTM_VM_DIR"
fi

FINAL_QCOW2="$UTM_VM_DIR/${VM_NAME}.qcow2"
mv "$QCOW2_FILE" "$FINAL_QCOW2"

echo "✅ VM disk created at: $FINAL_QCOW2"

# Create UTM VM launch script
cat > "${VM_NAME}-launch.sh" << EOF
#!/bin/bash
# Launch script for $VM_NAME

echo "🖥️  Starting $VM_NAME..."

qemu-system-aarch64 \\
  -machine virt,gic-version=3 \\
  -accel hvf \\
  -cpu host \\
  -m ${MEMORY}M \\
  -smp $CPU_COUNT \\
  -drive file="$FINAL_QCOW2",format=qcow2 \\
  -netdev vmnet-bridged,id=net0,ifname=bridge100 \\
  -device virtio-net-pci,netdev=net0 \\
  -nographic \\
  -serial stdio

EOF

chmod +x "${VM_NAME}-launch.sh"
mv "${VM_NAME}-launch.sh" "$HOME/Desktop/"

# Cleanup
cd /
rm -rf "$WORK_DIR"

echo "🎉 UTM VM creation complete!"
echo ""
echo "VM Details:"
echo "  Name: $VM_NAME"
echo "  Memory: ${MEMORY}MB"
echo "  CPUs: $CPU_COUNT"
echo "  Disk: $FINAL_QCOW2"
echo ""
echo "To launch the VM:"
echo "  ~/Desktop/${VM_NAME}-launch.sh"
echo ""
echo "Or import into UTM GUI and manage with:"
echo "  utmctl list"
echo "  utmctl start '$VM_NAME'" 