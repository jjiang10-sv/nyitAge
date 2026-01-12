# Upgrading Kubernetes and Cilium

This guide covers upgrading your self-managed Kubernetes cluster to newer versions after initial deployment.

> [!WARNING]
> Always test upgrades in a non-production environment first. Kubernetes upgrades should be done incrementally (e.g., 1.29 → 1.30 → 1.31), not skipping minor versions.

## Overview

After creating your cluster with the initial versions (e.g., K8s 1.29.0 and Cilium 1.15.1), you may want to upgrade to newer versions (e.g., K8s 1.31.0 and Cilium 1.16.5).

**Upgrade Order:**
1. Backup your cluster
2. Upgrade control plane nodes
3. Upgrade worker nodes
4. Upgrade Cilium CNI
5. Verify cluster health

## Prerequisites

```bash
# SSH into control plane
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@$(pulumi stack output control_plane_public_ips)

# Verify current versions
kubectl version --short
cilium version
```

## Step 1: Backup Critical Components

### Backup etcd

```bash
# On control plane node
sudo ETCDCTL_API=3 etcdctl \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key \
  snapshot save /root/etcd-backup-$(date +%Y%m%d-%H%M%S).db

# Verify backup
sudo ETCDCTL_API=3 etcdctl --write-out=table snapshot status /root/etcd-backup-*.db
```

### Backup Important Resources

```bash
# Export critical manifests
kubectl get all --all-namespaces -o yaml > /root/all-resources-backup.yaml
kubectl get pv,pvc --all-namespaces -o yaml > /root/storage-backup.yaml
kubectl get configmap,secret --all-namespaces -o yaml > /root/configs-backup.yaml

# Download backups to local machine
scp -i ~/.ssh/k8s-dev-key.pem ubuntu@$(pulumi stack output control_plane_public_ips):/root/*backup* ~/k8s-backups/
```

## Step 2: Upgrade Control Plane

### Check Upgrade Path

```bash
# Check available versions for upgrade
sudo apt-cache madison kubeadm | grep 1.31
```

> [!IMPORTANT]
> Kubernetes version upgrades must be sequential. You cannot skip minor versions.
> - ✅ Correct: 1.29 → 1.30 → 1.31
> - ❌ Wrong: 1.29 → 1.31

### Upgrade kubeadm First

```bash
# Unhold packages
sudo apt-mark unhold kubeadm

# Update package lists for new version
# For 1.30:
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.30/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-1-30-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-1-30-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.30/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes-1.30.list

# For 1.31:
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-1-31-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-1-31-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes-1.31.list

# Update apt and install specific version
sudo apt-get update
sudo apt-get install -y kubeadm=1.31.0-1.1

# Verify kubeadm version
kubeadm version

# Hold kubeadm at new version
sudo apt-mark hold kubeadm
```

### Plan and Apply Upgrade

```bash
# Check what will be upgraded
sudo kubeadm upgrade plan

# Apply upgrade (use exact version)
sudo kubeadm upgrade apply v1.31.0

# Expected output:
# [upgrade/successful] SUCCESS! Your cluster was upgraded to "v1.31.0". Enjoy!
```

### Upgrade kubelet and kubectl on Control Plane

```bash
# Drain control plane node (mark as unschedulable)
kubectl drain $(hostname) --ignore-daemonsets --delete-emptydir-data

# Unhold kubelet and kubectl
sudo apt-mark unhold kubelet kubectl

# Upgrade kubelet and kubectl
sudo apt-get update
sudo apt-get install -y kubelet=1.31.0-1.1 kubectl=1.31.0-1.1

# Hold at new version
sudo apt-mark hold kubelet kubectl

# Restart kubelet
sudo systemctl daemon-reload
sudo systemctl restart kubelet

# Uncordon control plane node
kubectl uncordon $(hostname)

# Verify
kubectl get nodes
```

## Step 3: Upgrade Worker Nodes

**Repeat these steps for EACH worker node:**

### SSH to Worker Node

```bash
# From your local machine
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@<WORKER_IP>
```

### Upgrade Worker Node

```bash
# Drain node from control plane first
# (Run this on control plane, not worker)
kubectl drain <worker-hostname> --ignore-daemonsets --delete-emptydir-data

# On worker node, unhold kubeadm
sudo apt-mark unhold kubeadm

# Add new Kubernetes repo (same as control plane)
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-1-31-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-1-31-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes-1.31.list

# Upgrade kubeadm
sudo apt-get update
sudo apt-get install -y kubeadm=1.31.0-1.1
sudo apt-mark hold kubeadm

# Upgrade node
sudo kubeadm upgrade node

# Upgrade kubelet and kubectl
sudo apt-mark unhold kubelet kubectl
sudo apt-get install -y kubelet=1.31.0-1.1 kubectl=1.31.0-1.1
sudo apt-mark hold kubelet kubectl

# Restart kubelet
sudo systemctl daemon-reload
sudo systemctl restart kubelet

# From control plane, uncordon the node
kubectl uncordon <worker-hostname>
```

### Verify Worker Upgrade

```bash
# From control plane
kubectl get nodes

# Should show all nodes at v1.31.0
# NAME          STATUS   ROLES           AGE   VERSION
# ip-10-0-0-x   Ready    control-plane   1h    v1.31.0
# ip-10-0-1-x   Ready    <none>          1h    v1.31.0
# ip-10-0-2-x   Ready    <none>          1h    v1.31.0
```

## Step 4: Upgrade Cilium

### Check Current Cilium Version

```bash
cilium version
```

### Upgrade Cilium CLI (if needed)

```bash
# Get latest Cilium CLI
CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
CLI_ARCH=amd64

curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
sudo tar xzvfC cilium-linux-${CLI_ARCH}.tar.gz /usr/local/bin
rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

cilium version
```

### Upgrade Cilium

```bash
# Check upgrade compatibility
cilium upgrade --version 1.16.5

# Apply upgrade
cilium upgrade --version 1.16.5

# Monitor rollout
kubectl rollout status -n kube-system daemonset/cilium

# Verify Cilium health
cilium status --wait

# Expected output:
#     /¯¯\
#  /¯¯\__/¯¯\    Cilium:             OK
#  \__/¯¯\__/    Operator:           OK
#  /¯¯\__/¯¯\    Envoy DaemonSet:    disabled (using embedded mode)
#  \__/¯¯\__/    Hubble Relay:       OK
#     \__/       ClusterMesh:        disabled
```

### Verify Cilium Connectivity

```bash
# Run connectivity tests
cilium connectivity test

# Check pod networking
kubectl get pods -A -o wide
kubectl exec -n kube-system <cilium-pod> -- cilium status
```

## Step 5: Post-Upgrade Verification

### System Health Checks

```bash
# Check all nodes
kubectl get nodes -o wide

# Check system pods
kubectl get pods -n kube-system

# Check workload pods
kubectl get pods -A

# Verify Cilium endpoints
kubectl get ciliumnodes -A
kubectl get ciliumendpoints -A

# Check SPIRE (if using mTLS)
kubectl get pods -n spire
```

### Functional Tests

```bash
# Deploy test pod
kubectl run test-nginx --image=nginx:latest --port=80

# Check pod is running
kubectl get pod test-nginx

# Expose and test connectivity
kubectl expose pod test-nginx --port=80 --target-port=80
kubectl run test-client --rm -it --image=busybox -- wget -O- test-nginx

# Cleanup
kubectl delete pod test-nginx
kubectl delete svc test-nginx
```

### Storage Tests

```bash
# Create test PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: ebs-sc
  resources:
    requests:
      storage: 1Gi
EOF

# Verify PVC is bound
kubectl get pvc test-pvc

# Create pod using PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-storage-pod
spec:
  containers:
  - name: test
    image: nginx
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: test-pvc
EOF

# Verify pod is running
kubectl get pod test-storage-pod

# Cleanup
kubectl delete pod test-storage-pod
kubectl delete pvc test-pvc
```

## Rollback Procedure

If the upgrade fails, you can rollback:

### Rollback Control Plane

```bash
# Downgrade packages
sudo apt-mark unhold kubeadm kubelet kubectl
sudo apt-get install -y kubeadm=1.29.0-1.1 kubelet=1.29.0-1.1 kubectl=1.29.0-1.1
sudo apt-mark hold kubeadm kubelet kubectl

# Restart kubelet
sudo systemctl daemon-reload
sudo systemctl restart kubelet
```

### Restore from etcd Backup

```bash
# Stop API server and etcd
sudo mv /etc/kubernetes/manifests /etc/kubernetes/manifests.backup
sleep 30

# Restore etcd snapshot
sudo ETCDCTL_API=3 etcdctl snapshot restore /root/etcd-backup-<timestamp>.db \
  --data-dir=/var/lib/etcd-restore

# Update etcd manifest to use new data directory
sudo sed -i 's|/var/lib/etcd|/var/lib/etcd-restore|g' /etc/kubernetes/manifests.backup/etcd.yaml

# Start API server and etcd
sudo mv /etc/kubernetes/manifests.backup /etc/kubernetes/manifests

# Wait for cluster to recover
kubectl get nodes
```

## Troubleshooting

### Kubeadm Upgrade Fails

```bash
# Check logs
sudo journalctl -xeu kubelet

# Check pod status
kubectl get pods -n kube-system

# Force cleanup if needed
sudo kubeadm reset
# Then re-init cluster with original version
```

### Cilium Upgrade Issues

```bash
# Check Cilium operator logs
kubectl logs -n kube-system deployment/cilium-operator

# Check Cilium agent logs
kubectl logs -n kube-system daemonset/cilium

# Rollback Cilium
helm rollback cilium -n kube-system
```

### Nodes Not Ready

```bash
# Check node conditions
kubectl describe node <node-name>

# Check kubelet status
sudo systemctl status kubelet
sudo journalctl -xeu kubelet

# Restart kubelet
sudo systemctl restart kubelet
```

## Best Practices

1. **Always backup before upgrading** - etcd, manifests, configs
2. **Upgrade incrementally** - Do not skip minor versions
3. **Test in staging first** - Never upgrade production directly
4. **Upgrade during maintenance windows** - Minimize user impact
5. **Monitor closely** - Watch logs and metrics during upgrade
6. **Have rollback plan** - Know how to rollback before starting
7. **Document everything** - Keep notes of what you did
8. **Verify at each step** - Don't proceed if something fails

## Quick Reference

### Target Versions Example

| Component | Old Version | New Version |
|-----------|-------------|-------------|
| Kubernetes | 1.29.0 | 1.31.0 |
| Cilium | 1.15.1 | 1.16.5 |
| containerd | 1.7.x | 1.7.x (usually no change) |

### Upgrade Commands Summary

```bash
# Control Plane
sudo kubeadm upgrade apply v1.31.0
sudo apt-get install -y kubelet=1.31.0-1.1 kubectl=1.31.0-1.1

# Workers
sudo kubeadm upgrade node
sudo apt-get install -y kubelet=1.31.0-1.1

# Cilium
cilium upgrade --version 1.16.5
```

### Verification Commands

```bash
kubectl get nodes                  # Check node versions
kubectl get pods -A                # Check all pods
cilium status                      # Check Cilium health
kubectl version --short            # Verify K8s version
```

## Related Documentation

- [Official Kubernetes Upgrade Guide](https://kubernetes.io/docs/tasks/administer-cluster/kubeadm/kubeadm-upgrade/)
- [Cilium Upgrade Guide](https://docs.cilium.io/en/stable/operations/upgrade/)
- [Backing Up an etcd Cluster](https://kubernetes.io/docs/tasks/administer-cluster/configure-upgrade-etcd/)

---

**Remember:** Upgrades should be planned, tested, and executed carefully. Always have a rollback plan ready.
