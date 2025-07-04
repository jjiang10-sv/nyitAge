#!/bin/bash

echo "Starting Burp Suite in Docker with noVNC..."

# Function to check if X server is running
check_x_server() {
    local retries=0
    local max_retries=30
    
    while [ $retries -lt $max_retries ]; do
        if xdpyinfo -display :0 >/dev/null 2>&1; then
            echo "X server is ready!"
            return 0
        fi
        echo "Waiting for X server... (attempt $((retries + 1))/$max_retries)"
        sleep 2
        retries=$((retries + 1))
    done
    
    echo "ERROR: X server failed to start after $max_retries attempts"
    return 1
}

# Wait for X server to be ready
if ! check_x_server; then
    echo "ERROR: Cannot start Burp Suite without X server"
    exit 1
fi

echo "X server is ready, starting Burp Suite..."

# Wait a bit more to ensure everything is stable
sleep 5

# Set up X11 forwarding for burp user
export DISPLAY=:0

# Start Burp Suite with better error handling (removed --project-file for Community Edition)
java -Xmx2g \
     -Djava.awt.headless=false \
     -Djava.net.useSystemProxies=false \
     -jar /home/burp/burpsuite_community.jar &

BURP_PID=$!
echo "Burp Suite started with PID: $BURP_PID"

# Wait for Burp to start and verify it's running
sleep 15

if ! kill -0 $BURP_PID 2>/dev/null; then
    echo "ERROR: Burp Suite failed to start"
    exit 1
fi

# Export certificate
echo "Exporting Burp certificate..."
if [ -f "/home/burp/export-cert.sh" ]; then
    /home/burp/export-cert.sh
else
    echo "WARNING: export-cert.sh not found"
fi

echo "Burp Suite is ready!"
echo "Access GUI via browser: http://localhost:6080"
echo "Proxy available at: localhost:8082"

# Keep the script running and monitor Burp process
wait $BURP_PID 