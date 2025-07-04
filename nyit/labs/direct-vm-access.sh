#!/bin/bash

echo "🖥️ Direct VM Console Access"
echo "This will show you the actual VM boot and login process"
echo ""

# Clean up any existing containers
echo "Cleaning up..."
docker ps -q --filter "name=security-lab" | xargs -r docker stop
docker ps -aq --filter "name=security-lab" | xargs -r docker rm

echo "Starting VM with direct console access..."
echo ""
echo "You should see:"
echo "  1. VM BIOS messages"
echo "  2. Linux boot process"  
echo "  3. Login prompt"
echo ""
echo "Once you see a login prompt, try common credentials:"
echo "  - Username: root, Password: (empty/enter)"
echo "  - Username: ubuntu, Password: ubuntu"
echo "  - Username: admin, Password: admin"
echo ""

# Simple direct VM launch
docker run -it --rm \
  --name vm-direct \
  --privileged \
  -p 2222:22 \
  -v $(pwd)/local:/data \
  alpine:latest sh -c '
  apk add --no-cache qemu-system-x86_64 && 
  echo "=== VM Starting Now ===" &&
  qemu-system-x86_64 \
    -machine pc \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -drive file=/data/vm-disk001.vmdk,format=vmdk,if=ide \
    -netdev user,id=net0,hostfwd=tcp::22-:22 \
    -device rtl8139,netdev=net0 \
    -nographic \
    -boot c
'

echo ""
echo "VM session ended. If you saw a login prompt, you can now:"
echo "1. Start the background VM: ./vm-control-working.sh start"
echo "2. Try SSH access: ssh -p 2222 username@localhost" 