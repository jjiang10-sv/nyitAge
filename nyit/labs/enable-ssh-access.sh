#!/bin/bash

echo "🔐 Starting VM with console access to enable SSH..."
echo ""
echo "Steps to enable SSH:"
echo "1. VM will boot to console"
echo "2. Login with existing credentials"
echo "3. Enable SSH service"
echo "4. Exit and restart in background mode"
echo ""

# Stop existing VM
docker stop security-lab-vm 2>/dev/null || true
docker rm security-lab-vm 2>/dev/null || true

echo "Starting VM with interactive console..."
echo ""

# Start with console access
docker run -it --rm \
  --name security-lab-setup \
  --privileged \
  -v $(pwd):/data \
  alpine:latest sh -c "
  apk add --no-cache qemu-system-x86_64 && 
  echo 'VM starting... You will see the boot process and login prompt.' &&
  echo 'Once logged in, run these commands to enable SSH:' &&
  echo '  sudo systemctl start ssh' &&
  echo '  sudo systemctl enable ssh' &&
  echo '  sudo ufw allow ssh' &&
  echo '' &&
  qemu-system-x86_64 \
    -machine pc \
    -accel tcg \
    -cpu qemu64 \
    -m 3324M \
    -smp 2 \
    -drive file=/data/server.qcow2,format=qcow2,if=ide \
    -netdev user,id=net0,hostfwd=tcp::2222-:22 \
    -device rtl8139,netdev=net0 \
    -nographic
"

echo ""
echo "After enabling SSH, restart with: docker-compose up -d" 