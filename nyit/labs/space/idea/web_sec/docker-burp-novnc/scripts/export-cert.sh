#!/bin/bash

echo "Exporting Burp Suite CA certificate..."

# Wait for Burp proxy to be ready
echo "Waiting for Burp proxy to start..."
while ! nc -z localhost 8082; do
    sleep 2
done

echo "Burp proxy is ready, downloading certificate..."

# Download the CA certificate
curl -x 127.0.0.1:8082 -k -o /tmp/certificates/burp-cert.der http://burp/cert 2>/dev/null

if [ -f "/tmp/certificates/burp-cert.der" ]; then
    echo "✅ Certificate downloaded successfully!"
    echo "📍 Certificate location: /tmp/certificates/burp-cert.der"
    
    # Convert to PEM format
    openssl x509 -inform DER -in /tmp/certificates/burp-cert.der -out /tmp/certificates/burp-cert.pem 2>/dev/null
    echo "🔄 Certificate converted to PEM format"
    
    echo ""
    echo "📋 To install certificate in Firefox:"
    echo "1. Copy from container: docker cp burp-suite-proxy:/tmp/certificates/burp-cert.der ."
    echo "2. Firefox: Settings > Privacy & Security > Certificates > View Certificates > Import"
    echo "3. Select the downloaded .der file and trust it for websites"
    
else
    echo "❌ Failed to download certificate!"
    echo "Make sure Burp Suite is fully started and proxy is running"
fi 