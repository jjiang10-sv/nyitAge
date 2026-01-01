# Quick Reference: Adding SPIRE + mTLS to platform.py

## What Needs to Be Added

### 1. Update Cilium Installation (line ~428-436)

**Change FROM:**
```bash
cilium install \\
  --version {cilium_version} \\
  --set ipam.mode=cluster-pool \\
  --set ipam.operator.clusterPoolIPv4PodCIDRList={pod_cidr} \\
  --set tunnel=disabled \\
  --set ipv4NativeRoutingCIDR={pod_cidr} \\
  --set kubeProxyReplacement=strict \\
  --set hubble.relay.enabled=true \\
  --set hubble.ui.enabled=true
```

**Change TO:**
```bash
cilium install \\
  --version {cilium_version} \\
  --set ipam.mode=cluster-pool \\
  --set ipam.operator.clusterPoolIPv4PodCIDRList={pod_cidr} \\
  --set tunnel=disabled \\
  --set ipv4NativeRoutingCIDR={pod_cidr} \\
  --set kubeProxyReplacement=strict \\
  --set hubble.relay.enabled=true \\
  --set hubble.ui.enabled=true \\
  --set authentication.mutual.spire.enabled=true \\
  --set authentication.mutual.spire.install.enabled=false \\
  --set authentication.mutual.spire.serverAddress=spire-server.spire:8081 \\
  --set authentication.mutual.spire.trustDomain=cluster.local
```

### 2. Add SPIRE Installation (after line ~468, after STORAGECLASS)

**Add this section after the STORAGECLASS creation and before the final success message:**

```bash
# Install SPIFFE/SPIRE for workload identity
echo "Installing SPIFFE/SPIRE..."
kubectl create namespace spire || true
kubectl create serviceaccount spire-server -n spire || true
kubectl create serviceaccount spire-agent -n spire || true

# Install SPIRE Server
kubectl apply -f - <<'SPIRESERVER'
apiVersion: v1
kind: ConfigMap
metadata:
  name: spire-server
  namespace: spire
data:
  server.conf: |
    server {
      bind_address = "0.0.0.0"
      bind_port = "8081"
      trust_domain = "cluster.local"
      data_dir = "/run/spire/data"
      log_level = "INFO"
      ca_subject = {
        country = ["US"]
        organization = ["Kubernetes"]
        common_name = "SPIRE Server"
      }
    }
    plugins {
      DataStore "sql" {
        plugin_data {
          database_type = "sqlite3"
          connection_string = "/run/spire/data/datastore.sqlite3"
        }
      }
      NodeAttestor "k8s_psat" {
        plugin_data {
          clusters = {
            "cluster.local" = {
              service_account_allow_list = ["spire:spire-agent"]
            }
          }
        }
      }
      KeyManager "disk" {
        plugin_data {
          keys_path = "/run/spire/data/keys.json"
        }
      }
      Notifier "k8sbundle" {
        plugin_data {}
      }
    }
    health_checks {
      listener_enabled = true
      bind_address = "0.0.0.0"
      bind_port = "8080"
      live_path = "/live"
      ready_path = "/ready"
    }
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: spire-server
  namespace: spire
spec:
  replicas: 1
  selector:
    matchLabels:
      app: spire-server
  serviceName: spire-server
  template:
    metadata:
      labels:
        app: spire-server
    spec:
      serviceAccountName: spire-server
      containers:
      - name: spire-server
        image: ghcr.io/spiffe/spire-server:1.8.0
        args: ["-config", "/run/spire/config/server.conf"]
        ports:
        - containerPort: 8081
        volumeMounts:
        - name: spire-config
          mountPath: /run/spire/config
          readOnly: true
        - name: spire-data
          mountPath: /run/spire/data
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 15
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
      volumes:
      - name: spire-config
        configMap:
          name: spire-server
  volumeClaimTemplates:
  - metadata:
      name: spire-data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: ebs-sc
      resources:
        requests:
          storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: spire-server
  namespace: spire
spec:
  ports:
  - name: grpc
    port: 8081
  selector:
    app: spire-server
SPIRESERVER

echo "Waiting for SPIRE Server..."
sleep 20
kubectl wait --for=condition=ready pod -l app=spire-server -n spire --timeout=300s || true

# Install SPIRE Agent  
kubectl apply -f - <<'SPIREAGENT'
apiVersion: v1
kind: ConfigMap
metadata:
  name: spire-agent
  namespace: spire
data:
  agent.conf: |
    agent {
      data_dir = "/run/spire"
      log_level = "INFO"
      server_address = "spire-server"
      server_port = "8081"
      trust_domain = "cluster.local"
    }
    plugins {
      NodeAttestor "k8s_psat" {
        plugin_data {
          cluster = "cluster.local"
        }
      }
      KeyManager "memory" {
        plugin_data {}
      }
      WorkloadAttestor "k8s" {
        plugin_data {
          skip_kubelet_verification = true
        }
      }
    }
    health_checks {
      listener_enabled = true
      bind_address = "0.0.0.0"
      bind_port = "8080"
      live_path = "/live"
      ready_path = "/ready"
    }
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: spire-agent
  namespace: spire
spec:
  selector:
    matchLabels:
      app: spire-agent
  template:
    metadata:
      labels:
        app: spire-agent
    spec:
      hostPID: true
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      serviceAccountName: spire-agent
      initContainers:
      - name: wait-spire-server
        image: busybox:1.35
        command: ['sh', '-c', 'until nslookup spire-server.spire.svc.cluster.local; do sleep 2; done']
      containers:
      - name: spire-agent
        image: ghcr.io/spiffe/spire-agent:1.8.0
        args: ["-config", "/run/spire/config/agent.conf"]
        volumeMounts:
        - name: spire-config
          mountPath: /run/spire/config
          readOnly: true
        - name: spire-agent-socket
          mountPath: /run/spire/sockets
        - name: spire-token
          mountPath: /var/run/secrets/tokens
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
      volumes:
      - name: spire-config
        configMap:
          name: spire-agent
      - name: spire-agent-socket
        hostPath:
          path: /run/spire/sockets
          type: DirectoryOrCreate
      - name: spire-token
        projected:
          sources:
          - serviceAccountToken:
              path: spire-agent
              expirationSeconds: 7200
              audience: spire-server
SPIREAGENT

echo "Waiting for SPIRE Agents..."
sleep 15
```

### 3. Update Success Message (line ~470-482)

**Change FROM:**
```bash
echo ""
echo "========================================"
echo "Cluster initialized successfully!"
echo "========================================"
echo ""
echo "Installed components:"
echo "- Kubernetes {kubernetes_version}"
echo "- Cilium {cilium_version}"
echo "- AWS EBS CSI Driver"
echo "- Default StorageClass: ebs-sc"
echo ""
echo "Join command:"
```

**Change TO:**
```bash
echo ""
echo "========================================"
echo "Cluster initialized successfully!"
echo "========================================"
echo ""
echo "Installed components:"
echo "- Kubernetes {kubernetes_version}"
echo "- Cilium {cilium_version} (with mTLS support)"
echo "- AWS EBS CSI Driver"
echo "- Default StorageClass: ebs-sc"
echo "- SPIFFE/SPIRE (workload identity)"
echo ""
echo "SPIRE Configuration:"
echo "- Trust Domain: cluster.local"
echo "- Server: spire-server.spire:8081"
echo "- Cilium mTLS: ENABLED"
echo ""
echo "Join command:"
```

---

## Manual Update Instructions

1. **Open platform.py** in your editor
2. **Find line 428-436** (Cilium installation)
   - Add the 4 new `--set authentication.mutual.spire.*` lines
3. **Find line 468** (end of STORAGECLASS)
   - Add the entire SPIRE installation section (server + agent)
4. **Find line 470-482** (success message)
   - Update to mention SPIRE and mTLS
5. **Save the file**
6. **Validate:** `python3 -m py_compile platform.py`

---

## Quick Test After Update

```bash
# Deploy the cluster
cd self-managed-k8s/
pulumi up

# SSH to control plane
ssh ubuntu@<CP_IP>

# Run init script
sudo /root/init-cluster.sh

# Verify SPIRE is installed
kubectl get pods -n spire

# Should see:
# spire-server-0   1/1 Running
# spire-agent-xxx  1/1 Running (one per node)

# Verify Cilium mTLS
kubectl exec -n kube-system ds/cilium -- cilium status | grep -i spire
# Should show SPIRE integration enabled
```

---

## Why Manual Update is Needed

The automated file editing failed due to special characters and multi-line string complexity in the bash script within Python f-strings. Manual editing is more reliable for this type of complex nested script.

## Complete Reference

See **SPIRE_MTLS_GUIDE.md** for:
- Complete architecture explanation
- How SPIRE + Cilium mTLS works
- Verification procedures
- Troubleshooting guide
- Network policy examples

---

*Happy securing! 🔐*
