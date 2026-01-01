# SPIFFE/SPIRE + Cilium mTLS Guide

## Overview

This guide covers the integration of **SPIFFE/SPIRE** with **Cilium mTLS** to provide:
- ✅ **Workload Identity** - Cryptographic identity for every pod
- ✅ **Mutual TLS** - Automatic encryption between services
- ✅ **Zero Trust** - No network perimeter, authenticate everything
- ✅ **No Code Changes** - Transparent to applications

---

## Table of Contents

1. [What is SPIFFE/SPIRE](#what-is-spiffespire)
2. [What is Cilium mTLS](#what-is-cilium-mtls)
3. [How They Work Together](#how-they-work-together)
4. [Architecture](#architecture)
5. [Verification](#verification)
6. [Using SPIRE Identities](#using-spire-identities)
7. [Network Policies with mTLS](#network-policies-with-mtls)
8. [Troubleshooting](#troubleshooting)
9. [Best Practices](#best-practices)

---

## What is SPIFFE/SPIRE?

### SPIFFE (Secure Production Identity Framework For Everyone)

**A standard for workload identity:**
```
SPIFFE ID format: spiffe://trust-domain/path

Examples:
- spiffe://cluster.local/ns/default/sa/my-app
- spiffe://cluster.local/ns/production/sa/frontend
```

**Key Concepts:**
- **Trust Domain**: `cluster.local` (your cluster)
- **Workload**: Any running process (pod/container)
- **SVID**: SPIFFE Verifiable Identity Document (X.509 cert)

### SPIRE (SPIFFE Runtime Environment)

**Implementation of SPIFFE:**
```
├─ SPIRE Server: Issues identities
├─ SPIRE Agent: Attests workloads
└─ Workload API: Provides SVIDs to apps
```

**What SPIRE Does:**
1. **Attests** workloads (verifies what they are)
2. **Issues** cryptographic identities (X.509 certs)
3. **Rotates** certificates automatically
4. **Federates** across clusters/environments

---

## What is Cilium mTLS?

### Mutual TLS with Cilium

**Transparent encryption between pods:**
```
Pod A → Cilium → [Encrypted Tunnel] → Cilium → Pod B

Application sees: Plain HTTP
Network sees: Encrypted TLS
```

**Key Features:**
- ✅ **Transparent** - No app changes needed
- ✅ **Automatic** - cilium manages certs
- ✅ **Identity-based** - Uses SPIRE identities
- ✅ **Policy-driven** - Control what can connect

---

## How They Work Together

### Integration Flow

```
┌──────────────────────────────────────────────────────┐
│                    SPIRE Server                       │
│  ├─ Issues X.509 certificates                        │
│  ├─ Trust Domain: cluster.local                      │
│  └─ Stores identities                                │
└─────────────────┬────────────────────────────────────┘
                  │
                  ↓
┌──────────────────────────────────────────────────────┐
│                    SPIRE Agent                        │
│  ├─ Runs on every node (DaemonSet)                   │
│  ├─ Attests pods                                     │
│  └─ Provides SVIDs via Workload API                  │
└─────────────────┬────────────────────────────────────┘
                  │
                  ↓
┌──────────────────────────────────────────────────────┐
│                    Cilium Agent                       │
│  ├─ Reads SVIDs from SPIRE                           │
│  ├─ Establishes mTLS tunnels                         │
│  └─ Enforces identity-based policies                 │
└─────────────────┬────────────────────────────────────┘
                  │
                  ↓
┌──────────────────────────────────────────────────────┐
│              Service-to-Service Traffic               │
│  Frontend ←─[mTLS]─→ Backend ←─[mTLS]─→ Database    │
│  (Encrypted & Authenticated automatically)           │
└──────────────────────────────────────────────────────┘
```

### Step-by-Step

**1. Pod Starts:**
```
kubectl run frontend --image=nginx
```

**2. SPIRE Agent Attests:**
```
- Reads pod metadata from kubelet
- Verifies: namespace, service account, labels
- Requests SVID from SPIRE Server
```

**3. SPIRE Server Issues Identity:**
```
SPIFFE ID: spiffe://cluster.local/ns/default/sa/default
Certificate: X.509 cert with 1-hour TTL
Key: Private key for mTLS
```

**4. Cilium Gets Identity:**
```
- Cilium reads SVID from SPIRE Workload API
- Associates SVID with pod's network endpoint
- Enables mTLS for pod
```

**5. Service Call:**
```
Frontend → Backend:
  1. Cilium on frontend node encrypts with backend's cert
  2. Network sees only encrypted traffic
  3. Cilium on backend node decrypts with SVID
  4. Backend receives plaintext
  5. All transparent to application!
```

---

## Architecture

### Our Implementation

```
┌─────────────────────────────────────────────────────────┐
│ Namespace: spire                                         │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ SPIRE Server (StatefulSet)                          │ │
│  │ ├─ Image: ghcr.io/spiffe/spire-server:1.8.0        │ │
│  │ ├─ Storage: EBS (1Gi)                               │ │
│  │ ├─ Port: 8081 (gRPC)                                │ │
│  │ ├─ Trust Domain: cluster.local                      │ │
│  │ └─ DataStore: SQLite                                │ │
│  └────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ SPIRE Agent (DaemonSet)                             │ │
│  │ ├─ Image: ghcr.io/spiffe/spire-agent:1.8.0         │ │
│  │ ├─ Runs on: Every node                              │ │
│  │ ├─ Socket: /run/spire/sockets                       │ │
│  │ ├─ Attestor: k8s_psat (Projected Service Token)    │ │
│  │ └─ Workload Attestor: k8s (Kubernetes)              │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│ Namespace: kube-system                                   │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ Cilium (DaemonSet)                                  │ │
│  │ ├─ mTLS: Enabled                                    │ │
│  │ ├─ SPIRE Address: spire-server.spire:8081           │ │
│  │ ├─ Trust Domain: cluster.local                      │ │
│  │ └─ Policy Enforcement: identity-based               │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Configuration

**Cilium mTLS Settings (in platform.py):**
```bash
cilium install \
  --set authentication.mutual.spire.enabled=true \
  --set authentication.mutual.spire.install.enabled=false \
  --set authentication.mutual.spire.serverAddress=spire-server.spire:8081 \
  --set authentication.mutual.spire.trustDomain=cluster.local
```

**What These Do:**
- `spire.enabled=true` - Enable SPIRE integration
- `install.enabled=false` - We install SPIRE separately (more control)
- `serverAddress` - Where Cilium finds SPIRE
- `trustDomain` - Must match SPIRE server config

---

## Verification

### 1. Check SPIRE Server

```bash
# Check SPIRE server is running
kubectl get pods -n spire -l app=spire-server

# Should see:
# NAME              READY   STATUS    RESTARTS   AGE
# spire-server-0    1/1     Running   0          5m

# Check logs
kubectl logs -n spire spire-server-0

# Should see:
# msg="Server is now available" subsystem_name=endpoints
```

### 2. Check SPIRE Agents

```bash
# Check agents on all nodes
kubectl get pods -n spire -l app=spire-agent -o wide

# Should have one pod per node:
# NAME                READY   STATUS    NODE
# spire-agent-abc12   1/1     Running   node1
# spire-agent-def34   1/1     Running   node2
# spire-agent-ghi56   1/1     Running   node3

# Check agent logs
kubectl logs -n spire -l app=spire-agent --tail=50
```

### 3. Check Cilium mTLS Status

```bash
# Exec into Cilium pod
kubectl exec -n kube-system ds/cilium -- cilium status

# Look for:
# Authentication:
#   Mode: mutual
#   SPIRE:
#     Server Address: spire-server.spire:8081
#     Trust Domain: cluster.local

# Or check more details:
kubectl exec -n kube-system ds/cilium -- cilium config view | grep -i spire
```

### 4. Test Pod Identity

```bash
# Create test pod
kubectl run test-pod --image=busybox:latest --command -- sleep 3600

# Check if SPIRE issued identity
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server entry show

# Should see entry for test-pod with SPIFFE ID like:
# spiffe://cluster.local/ns/default/sa/default
```

### 5. Verify mTLS is Active

```bash
# Create two test services
kubectl run frontend --image=nginx
kubectl run backend --image=nginx

kubectl expose pod frontend --port=80
kubectl expose pod backend --port=80

# Check Cilium sees identities
kubectl exec -n kube-system ds/cilium -- cilium endpoint list

# Should show SPIFFE IDs associated with endpoints
```

---

## Using SPIRE Identities

### Automatic Identity Assignment

**Every pod automatically gets an identity:**

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: production
spec:
  serviceAccountName: my-app-sa
  containers:
  - name: app
    image: my-app:latest
```

**Resulting SPIFFE ID:**
```
spiffe://cluster.local/ns/production/sa/my-app-sa
```

**No code changes needed!** SPIRE + Cilium handle everything.

---

### Custom Identity Registration

**For more granular control, register entries manually:**

```bash
# Register identity for specific deployment
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server entry create \
  -spiffeID spiffe://cluster.local/ns/production/deployment/frontend \
  -parentID spiffe://cluster.local/spire/agent/k8s_psat/cluster.local/node/worker-1 \
  -selector k8s:ns:production \
  -selector k8s:sa:frontend-sa \
  -selector k8s:container-name:frontend

# This creates an identity specifically for frontend deployment
```

---

## Network Policies with mTLS

### Identity-Based Policies

**Traditional Network Policy (IP-based):**
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-frontend
spec:
  podSelector:
    matchLabels:
      app: backend
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend  # ← Can be spoofed!
```

**Cilium Network Policy (Identity-based with mTLS):**
```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-frontend-mtls
spec:
  endpointSelector:
    matchLabels:
      app: backend
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    authentication:
      mode: required  # ← Requires valid SPIFFE ID + mTLS!
```

**Why Better:**
- ✅ Can't be spoofed (cryptographic identity)
- ✅ Automatically encrypted
- ✅ Identity verified before allowing traffic

---

### Example: Three-Tier App with mTLS

```yaml
---
# Frontend → Backend (mTLS required)
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: backend-policy
  namespace: production
spec:
  endpointSelector:
    matchLabels:
      tier: backend
  ingress:
  - fromEndpoints:
    - matchLabels:
        tier: frontend
    authentication:
      mode: required  # mTLS + valid identity required
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
---
# Backend → Database (mTLS required)
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: database-policy
  namespace: production
spec:
  endpointSelector:
    matchLabels:
      tier: database
  ingress:
  - fromEndpoints:
    - matchLabels:
        tier: backend
    authentication:
      mode: required  # mTLS + valid identity required
    toPorts:
    - ports:
      - port: "5432"
        protocol: TCP
```

**What This Achieves:**
- Frontend can only talk to backend with valid SPIFFE ID
- Backend can only talk to database with valid SPIFFE ID
- All traffic encrypted automatically
- Identity verified cryptographically

---

## Troubleshooting

### Issue 1: SPIRE Server Not Starting

**Symptom:**
```bash
kubectl get pods -n spire
# spire-server-0  0/1  CrashLoopBackOff
```

**Debug:**
```bash
kubectl logs -n spire spire-server-0

# Common issues:
# - PVC not bound (check: kubectl get pvc -n spire)
# - Config error
# - Port already in use
```

**Fix:**
```bash
# Check PVC
kubectl get pvc -n spire
# If Pending, check StorageClass exists

# Check config
kubectl get configmap -n spire spire-server -o yaml

# Delete and recreate if needed
kubectl delete pod -n spire spire-server-0
```

---

### Issue 2: SPIRE Agent Can't Reach Server

**Symptom:**
```bash
kubectl logs -n spire -l app=spire-agent
# Error: connection refused to spire-server
```

**Debug:**
```bash
# Check service
kubectl get svc -n spire spire-server

# Test connectivity from agent pod
kubectl exec -n spire spire-agent-xxx -- \
  nc -zv spire-server.spire.svc.cluster.local 8081
```

**Fix:**
```bash
# Ensure spire-server service exists and is correct
# Delete agents to recreate
kubectl rollout restart daemonset/spire-agent -n spire
```

---

### Issue 3: Cilium Not Getting Identities

**Symptom:**
```bash
cilium status
# SPIRE: Connection failed
```

**Debug:**
```bash
# Check Cilium can reach SPIRE
kubectl exec -n kube-system ds/cilium -- \
  curl -v http://spire-server.spire:8081/

# Check Cilium config
kubectl exec -n kube-system ds/cilium -- cilium config view | grep spire
```

**Fix:**
```bash
# Restart Cilium
kubectl rollout restart daemonset/cilium -n kube-system

# Or reinstall with correct settings
cilium upgrade --set authentication.mutual.spire.enabled=true ...
```

---

### Issue 4: mTLS Not Working

**Symptom:**
```
Pods can communicate but traffic not encrypted
```

**Debug:**
```bash
# Check if policy enforcement is enabled
kubectl exec -n kube-system ds/cilium -- cilium config view | grep -i policy

# Check endpoint has identity
kubectl exec -n kube-system ds/cilium -- cilium endpoint list

# Capture traffic (should be encrypted)
kubectl exec -n kube-system ds/cilium -- \
  tcpdump -i any -n port 8080 -X | head -100
```

**Fix:**
```yaml
# Ensure CiliumNetworkPolicy has authentication.mode: required
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
spec:
  ingress:
  - authentication:
      mode: required  # ← This line!
```

---

## Best Practices

### 1. Separate SPIRE Server Per Environment

```
Production Cluster:
└─ SPIRE Trust Domain: prod.company.com

Staging Cluster:
└─ SPIRE Trust Domain: staging.company.com

Development Cluster:
└─ SPIRE Trust Domain: dev.company.com
```

**Why:** Different trust boundaries for different environments.

---

### 2. Use ServiceAccounts for Identity

```yaml
# Create specific ServiceAccount per microservice
apiVersion: v1
kind: ServiceAccount
metadata:
  name: frontend-sa
  namespace: production
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  template:
    spec:
      serviceAccountName: frontend-sa  # ← Explicit SA
```

**SPIFFE ID becomes:**
```
spiffe://cluster.local/ns/production/sa/frontend-sa
```

**Why:** Clear identity per service, easier to manage policies.

---

### 3. Enable Policy Enforcement Gradually

**Phase 1: Monitor Only**
```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
spec:
  ingress:
  - authentication:
      mode: "optional"  # ← Log but don't block
```

**Phase 2: Enforce**
```yaml
spec:
  ingress:
  - authentication:
      mode: "required"  # ← Block invalid identities
```

**Why:** See what would be blocked before breaking prod.

---

### 4. Monitor SPIRE Certificate Rotation

```bash
# SPIRE automatically rotates certs (default: 1 hour TTL)
# Monitor rotation is happening:

kubectl logs -n spire -l app=spire-agent --tail=100 | grep -i rotation

# Should see regular rotation messages
```

---

### 5. Back Up SPIRE Server Data

```bash
# SPIRE server stores trust relationships in sqlite
# Back up the PVC regularly

# Create backup
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server info

# PVC is backed by EBS, ensure EBS snapshots are enabled
```

---

## Summary

### What You Get

✅ **Automatic Workload Identity**
- Every pod gets cryptographic identity
- Based on Kubernetes metadata
- No application changes

✅ **Transparent mTLS**
- Service-to-service encryption
- Handled by Cilium
- No code changes

✅ **Identity-Based Policies**
- Can't be spoofed (crypto-verified)
- Fine-grained access control
- Better than IP-based

✅ **Zero Trust Architecture**
- Authenticate everything
- Encrypt everything
- Trust nothing by default

---

### Architecture Summary

```
Application Layer:
├─ Apps don't know about SPIRE/mTLS
└─ Just make normal HTTP calls

Cilium Layer:
├─ Intercepts traffic
├─ Gets identity from SPIRE
├─ Establishes mTLS tunnels
└─ Enforces policies

SPIRE Layer:
├─ Attests workloads
├─ Issues identities (SVIDs)
├─ Rotates certificates
└─ Federates trust

Result: Zero-trust, encrypted, identity-based networking! 🔐
```

---

## Next Steps

1. **Deploy your applications** - They'll automatically get identities
2. **Create CiliumNetworkPolicies** - Use authentication.mode: required
3. **Monitor with Hubble** - See mTLS connections in Hubble UI
4. **Enable federation** (optional) - Connect to other clusters

**You now have enterprise-grade identity and encryption!** 🚀
