#!/bin/bash

# Security Lab VM Management Script

case "$1" in
  start)
    echo "🚀 Starting Security Lab VM..."
    docker-compose up -d security-lab
    echo "✅ VM started! Check status with: $0 status"
    echo "💡 If SSH fails, first run: $0 setup-ssh"
    ;;
    
  setup-ssh)
    echo "🔧 Starting VM in console mode to setup SSH..."
    echo "After VM boots, login and run:"
    echo "  sudo systemctl start ssh"
    echo "  sudo systemctl enable ssh"
    echo "  sudo passwd root"
    echo ""
    docker-compose --profile console up security-lab-console
    ;;
    
  stop)
    echo "🛑 Stopping Security Lab VM..."
    docker-compose down
    echo "✅ VM stopped!"
    ;;
    
  restart)
    echo "🔄 Restarting Security Lab VM..."
    docker-compose restart security-lab
    ;;
    
  status)
    echo "📊 VM Status:"
    docker-compose ps
    echo ""
    echo "🌐 Port Mappings:"
    docker port security-lab-vm 2>/dev/null || echo "VM not running"
    ;;
    
  ssh)
    echo "🔐 Connecting via SSH..."
    ssh -p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null root@localhost
    ;;
    
  console)
    echo "💻 Accessing VM console..."
    docker exec -it security-lab-vm sh
    ;;
    
  logs)
    echo "📋 VM Logs:"
    docker-compose logs -f security-lab
    ;;
    
  *)
    echo "Security Lab VM Control"
    echo ""
    echo "Usage: $0 {start|setup-ssh|stop|restart|status|ssh|console|logs}"
    echo ""
    echo "Commands:"
    echo "  start      - Start the VM in background"
    echo "  setup-ssh  - Start VM in console mode to enable SSH"
    echo "  stop       - Stop the VM"
    echo "  restart    - Restart the VM"
    echo "  status     - Show VM status and ports"
    echo "  ssh        - Connect via SSH (port 2222)"
    echo "  console    - Access Docker container console"
    echo "  logs       - Show VM logs"
    ;;
esac 