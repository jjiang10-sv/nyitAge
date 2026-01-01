#!/bin/bash
# Script to add SPIRE and mTLS to platform.py
# Run this to update the control plane init script

FILE="/Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/self-managed-k8s/platform.py"

# Find the line with "Install Cilium with native routing" and update the cilium install command
sed -i.bak '428,436s/--set hubble.ui.enabled=true/--set hubble.ui.enabled=true \\\
  --set authentication.mutual.spire.enabled=true \\\
  --set authentication.mutual.spire.install.enabled=false \\\
  --set authentication.mutual.spire.serverAddress=spire-server.spire:8081 \\\
  --set authentication.mutual.spire.trustDomain=cluster.local/' "$FILE"

echo "✅ Updated Cilium installation with mTLS/SPIRE support"
echo "Now you need to manually add SPIRE installation after line 468 (after STORAGECLASS)"
echo "See SPIRE_MTLS_GUIDE.md for the complete SPIRE server/agent configuration"
