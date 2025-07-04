#!/bin/bash

echo "🔧 Fixing SSH Access - Direct VM Console"
echo "This will show you the actual VM boot process and login prompt"
echo ""

# Stop current VM
docker stop security-lab-vm 2>/dev/null
docker rm security-lab-vm 2>/dev/null

echo "Starting VM with full console output..."
echo "Watch for:"
echo "  - Linux boot messages" 
echo "  - Login prompt"
echo "  - Error messages"
echo ""

# Run with full console access and proper terminal
docker run -it --rm \
  --name security-lab-debug \
  --privileged \
  -v $(pwd)/local:/data \
  alpine:latest sh -c "
  apk add --no-cache qemu-system-x86_64 && 
  echo '=== VM Starting with Console Access ===' &&
  qemu-system-x86_64 \
    -machine pc \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -drive file=/data/server.qcow2,format=qcow2,if=ide \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -device rtl8139,netdev=net0 \
    -nographic \
    -serial stdio
"

echo ""
echo "After VM boots and you login:"
echo "1. Check if SSH is running: systemctl status ssh"
echo "2. Start SSH if needed: sudo systemctl start ssh"
echo "3. Set root password: sudo passwd root"
echo "4. Test login: ssh -p 2222 root@<vm-ip>" 