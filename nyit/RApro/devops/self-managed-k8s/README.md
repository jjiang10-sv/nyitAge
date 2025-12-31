# Self-Managed Kubernetes on EC2 with Cilium

## Overview

This is a **completely self-managed Kubernetes cluster** on EC2 instances with **pure Cilium CNI**.

**NO EKS, NO AWS CNI, NO COMPROMISES** - just raw Kubernetes + Cilium.

---

## Why This Approach?

###  Avoids ALL EKS-Cilium Traps! ✅

| EKS Trap | Self-Managed Solution |
|----------|----------------------|
| **No AWS Support for Cilium** | ✅ You don't need AWS support - full control |
| **EKS Upgrades Break Cilium** | ✅ You control both K8s AND Cilium versions |
| **AWS VPC CNI Conflicts** | ✅ No AWS VPC CNI at all! |
| **Version Compatibility Hell** | ✅ Pick any compatible combo you want |
| **ENI Limits** | ✅ Use overlay mode, no limits |
| **IAM Complex Permissions** | ✅ Simpler - just EC2 + ENI permissions |
| **Must Remove VPC CNI** | ✅ Never installed in the first place |
| **No Cluster Mesh** | ✅ Full Cilium features available |
| **No BGP** | ✅ All Cilium features work |
| **Debugging Hard** | ✅ You own everything, full access |

**Result:** You have **100% control** over your Kubernetes environment!

---

## Architecture

### Infrastructure Layout

```
┌─────────────────────────────────────────────┐
│ AWS VPC (10.0.0.0/16)                        │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │ Control Plane (t3.medium)               │ │
│  │ ├─ kubeadm                              │ │
│  │ ├─ containerd                           │ │
│  │ ├─ etcd                                 │ │
│  │ ├─ kube-apiserver                       │ │
│  │ ├─ kube-controller-manager              │ │
│  │ └─ kube-scheduler                       │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │ Worker 1 (t3.large)                     │ │
│  │ ├─ kubelet                              │ │
│  │ └─ containerd                           │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │ Worker 2 (t3.large)                     │ │
│  │ ├─ kubelet                              │ │
│  │ └─ containerd                           │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │ Worker 3 (t3.large)                     │ │
│  │ ├─ kubelet                              │ │
│  │ └─ containerd                           │ │
│  └────────────────────────────────────────┘ │
│                                              │
│  ┌────────────────────────────────────────┐ │
│  │ Cilium (DaemonSet on all nodes)        │ │
│  │ ├─ Pure CNI (NO AWS VPC CNI!)           │ │
│  │ ├─ Native routing OR overlay            │ │
│  │ ├─ eBPF dataplane                       │ │
│  │ ├─ Hubble observability                │ │
│  │ └─ All features unlocked!               │ │
│  └────────────────────────────────────────┘ │
└─────────────────────────────────────────────┘
```

---

## Quick Start

### 1. Deploy Infrastructure

```bash
cd self-managed-k8s/

# Configure
pulumi config set cluster_name production
pulumi config set k8s_version 1.29.0
pulumi config set cilium_version 1.15.1

# Deploy
pulumi up
```

**This creates:**
- 1× Control plane node (t3.medium)
- 3× Worker nodes (t3.large)
- Security groups
- IAM roles
- All networking

**Time:** ~5 minutes

---

### 2. Initialize Kubernetes Cluster

```bash
# Get control plane IP
CP_IP=$(pulumi stack output control_plane_public_ips)

# SSH to control plane
ssh ubuntu@$CP_IP

# Initialize cluster (this runs kubeadm init + installs Cilium)
sudo /root/init-cluster.sh
```

**What this does:**
1. Runs `kubeadm init` with custom flags
2. Skips kube-proxy (Cilium replaces it)
3. Installs Cilium CLI
4. Installs Cilium with native routing
5. Enables Hubble

**Output:**
```
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, run:
  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config

Then you can join any number of worker nodes by running:
  kubeadm join 10.0.0.10:6443 --token abc123... \\
    --discovery-token-ca-cert-hash sha256:xyz789...
```

**Save the join command!**

---

### 3. Join Worker Nodes

```bash
# Get join command from control plane
JOIN_CMD=$(ssh ubuntu@$CP_IP "sudo kubeadm token create --print-join-command")

# Get worker IPs
WORKER_IPS=$(pulumi stack output worker_public_ips | jq -r '.[]')

# Join each worker
for WORKER_IP in $WORKER_IPS; do
  echo "Joining $WORKER_IP..."
  ssh ubuntu@$WORKER_IP "sudo $JOIN_CMD"
done
```

**Time:** ~2 minutes per node

---

### 4. Download kubeconfig

```bash
# Copy kubeconfig from control plane
scp ubuntu@$CP_IP:/etc/kubernetes/admin.conf ~/.kube/config

# Or use this script:
ssh ubuntu@$CP_IP "sudo cat /etc/kubernetes/admin.conf" > ~/.kube/config
chmod 600 ~/.kube/config
```

---

### 5. Verify Cluster

```bash
# Check nodes
kubectl get nodes

# Output:
# NAME                STATUS   ROLES           AGE   VERSION
# ip-10-0-0-10        Ready    control-plane   5m    v1.29.0
# ip-10-0-1-20        Ready    <none>          3m    v1.29.0
# ip-10-0-2-30        Ready    <none>          3m    v1.29.0

# Check Cilium
kubectl get pods -n kube-system -l k8s-app=cilium

# Run Cilium status
kubectl exec -n kube-system ds/cilium -- cilium status

# Run connectivity test
kubectl exec -n kube-system ds/cilium -- cilium connectivity test
```

**Expected:** All nodes Ready, all Cilium pods Running

---

## Adding New Nodes

### Method 1: Via Pulumi (Automated)

```python
# Edit example_usage.py
cluster = SelfManagedK8sCluster(
    cluster_name,
    worker_count=5,  # Increase from 3 to 5
    # ... rest of config
)

# Update infrastructure
pulumi up

# New nodes will be created with worker init script
# Then join them as described below
```

---

### Method 2: Manual EC2 Instance

```bash
# 1. Create new EC2 instance with:
#    - Same AMI (Ubuntu 22.04)
#    - Same security group (worker_sg)
#    - Same IAM instance profile
#    - User data: worker_init_script

# 2. Wait for instance to be ready (~3 minutes)

# 3. Get join command
JOIN_CMD=$(ssh ubuntu@$CP_IP "sudo kubeadm token create --print-join-command")

# 4. SSH to new node and join
ssh ubuntu@$NEW_NODE_IP "sudo $JOIN_CMD"

# 5. Verify
kubectl get nodes
```

---

### Method 3: Using Pulumi Resource

```python
# Add to example_usage.py
new_worker = aws.ec2.Instance(
    "new-worker-4",
    instance_type="t3.large",
    ami=ubuntu_ami.id,
    subnet_id=subnets[0].id,
    vpc_security_group_ids=[cluster.worker_sg.id],
    iam_instance_profile=cluster.instance_profile.name,
    user_data=worker_init_script,
    tags={
        "Name": "k8s-worker-4",
        "Role": "worker",
    },
)

# Deploy
pulumi up

# Then join as usual
```

---

## Node Management

### Removing a Node

```bash
# 1. Drain node
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data

# 2. Delete from cluster
kubectl delete node <node-name>

# 3. Terminate EC2 instance
aws ec2 terminate-instances --instance-ids <instance-id>
```

### Upgrading Nodes

```bash
# 1. Upgrade control plane first
ssh ubuntu@$CP_IP

# Upgrade kubeadm
sudo apt-mark unhold kubeadm
sudo apt-get update
sudo apt-get install -y kubeadm=1.30.0-1.1
sudo apt-mark hold kubeadm

# Upgrade control plane
sudo kubeadm upgrade plan
sudo kubeadm upgrade apply v1.30.0

# Upgrade kubelet
sudo apt-mark unhold kubelet kubectl
sudo apt-get install -y kubelet=1.30.0-1.1 kubectl=1.30.0-1.1
sudo apt-mark hold kubelet kubectl
sudo systemctl daemon-reload
sudo systemctl restart kubelet

# 2. Upgrade each worker
# Drain, upgrade, uncordon
kubectl drain <worker-name> --ignore-daemonsets
ssh ubuntu@$WORKER_IP "sudo apt-get update && sudo apt-get install -y ..."
kubectl uncordon <worker-name>
```

---

## Cilium Management

### Upgrade Cilium

```bash
# SSH to control plane
ssh ubuntu@$CP_IP

# Upgrade Cilium
cilium upgrade --version 1.16.0

# Verify
cilium status
kubectl rollout status daemonset/cilium -n kube-system
```

### Enable Hubble UI

```bash
# Already enabled by default!
# Access it:
kubectl port-forward -n kube-system svc/hubble-ui 8080:80

# Open browser: http://localhost:8080
```

### Enable Cluster Mesh

```bash
# On first cluster
cilium clustermesh enable --context cluster1
cilium clustermesh status

# On second cluster
cilium clustermesh enable --context cluster2

# Connect clusters
cilium clustermesh connect --context cluster1 --destination-context cluster2

# Verify
cilium clustermesh status
```

**This works because you have full control!** No AKS limitations.

---

## Cost Comparison

### Self-Managed vs EKS

**Self-Managed (this approach):**
```
1× t3.medium control plane:    $30/month
3× t3.large workers:            $180/month
NAT Gateway (optional):         $45/month
Data transfer:                  ~$20/month
────────────────────────────────────────
Total:                          $275/month
```

**EKS (from earlier analysis):**
```
EKS control plane:              $73/month
3× t3.large workers:            $180/month
NAT Gateway:                    $45/month
────────────────────────────────────────
Total:                          $298/month
```

**AKS (from earlier analysis):**
```
Total:                          $2,650/month
```

### Savings

 - **vs EKS:** $23/month (8% cheaper)
- **vs AKS:** $2,375/month (90% cheaper!)

**Plus: Full control, no vendor lock-in!**

---

## Advantages Over EKS

### 1. Full Version Control ✅

```
YOU choose:
├─ Kubernetes version (any!)
├─ Cilium version (any!)
├─ Upgrade timing
└─ No forced upgrades
```

**vs EKS:** AWS controls versions, forces upgrades

---

### 2. Pure Cilium Features ✅

```
Available (no limitations):
├─ Cluster Mesh (full support)
├─ BGP (works perfectly)
├─ Custom IPAM modes
├─ Advanced routing
├─ Service mesh
└─ All experimental features
```

**vs EKS:** No official support, compatibility issues

---

### 3. No AWS CNI Conflicts ✅

```
No conflicts because:
└─ AWS VPC CNI never installed!
```

**vs EKS:** Must manually remove VPC CNI

---

### 4. Flexible Networking ✅

```
Choose any networking:
├─ Native routing (fastest)
├─ Overlay (most flexible)
├─ BGP
└─ Custom configurations
```

**vs EKS:** Limited by AWS networking

---

### 5. Debug Anywhere ✅

```
Full access to:
├─ Control plane logs
├─ etcd
├─ All system components
└─ Every config file
```

**vs EKS:** Control plane is black box

---

## Trade-offs vs EKS

### What You Give Up

#### 1. No Managed Control Plane

**Self-Managed:**
- You run etcd
- You monitor control plane
- You back up etcd
- You handle control plane HA (need 3 nodes)

**EKS:**
- AWS manages all of this
- Auto-scaling control plane
- Automated backups

**Mitigation:**
```bash
# For production, use 3 control plane nodes:
control_plane_count=3  # In platform.py

# Set up etcd backups:
# (Add to control plane user data)
*/30 * * * * etcdctl snapshot save /backup/etcd-$(date +\%Y\%m\%d-\%H\%M).db
```

---

#### 2. More Operational Work

**Self-Managed:**
- Manual upgrades
- Manual security patches
- Monitor node health
- Plan for failures

**EKS:**
- One-click upgrades
- Automated patches
- Auto-recovery

**Mitigation:**
```bash
# Automate with scripts
# Use monitoring (Prometheus, Grafana)
# Set up alerts
# Document procedures
```

---

#### 3. No AWS SLA

**Self-Managed:**
- You're responsible for uptime
- No SLA from AWS

**EKS:**
- 99.95% uptime SLA

**Mitigation:**
- Multi-AZ deployment
- Regular backups
- Good monitoring
- Disaster recovery plan

---

#### 4. Security Updates

**Self-Managed:**
- You apply OS patches
- You update Kubernetes
- You secure nodes

**EKS:**
- Managed AMIs with security patches

**Mitigation:**
```bash
# Automate OS updates
sudo apt-get update && sudo apt-get upgrade -y

# Use tools like:
# - AWS Systems Manager
# - Ansible
# - Chef/Puppet
```

---

## When to Use Self-Managed

### ✅ Use Self-Managed If:

1. **Want Full Control**
   - Need specific K8s versions
   - Want all Cilium features
   - Custom networking requirements

2. **Cost Sensitive**
   - Slightly cheaper than EKS
   - Much cheaper than AKS
   - Can handle operational overhead

3. **Have Expertise**
   - Team knows Kubernetes internals
   - Can manage control plane
   - Comfortable with kubeadm

4. **Avoid Vendor Lock-in**
   - Want portability
   - Can migrate to any cloud
   - No EKS-specific dependencies

5. **Need Bleeding Edge**
   - Latest K8s features
   - Latest Cilium features
   - Experimental configurations

---

### ❌ Use EKS Instead If:

1. **Want Managed Service**
   - Prefer hands-off control plane
   - Want AWS SLA
   - Limited DevOps resources

2. **Need Easy Upgrades**
   - One-click upgrades
   - Automated patches
   - Less operational work

3. **Enterprise Requirements**
   - Need SLA guarantees
   - Compliance requirements
   - Vendor support needed

4. **AWS Integration**
   - Deep AWS service integration
   - IAM roles for service accounts
   - AWS-specific features

---

## Comparison Summary

| Aspect | Self-Managed | EKS | AKS |
|--------|-------------|-----|-----|
| **Cost** | $275/mo | $298/mo | $2,650/mo |
| **Control** | ✅ Full | ⚠️ Limited | ⚠️ Limited |
| **Cilium** | ✅ Pure | ⚠️ Community | ✅ Hybrid |
| **Ops Overhead** | ❌ High | ✅ Low | ✅ Low |
| **Version Control** | ✅ Full | ⚠️ Limited | ⚠️ Limited |
| **SLA** | ❌ None | ✅ 99.95% | ✅ 99.95% |
| **Support** | ❌ DIY | ⚠️ AWS (not Cilium) | ✅ Microsoft |
| **Flexibility** | ✅ Highest | ⚠️ Medium | ⚠️ Low |

---

## Production Checklist

### Before Production

- [ ] **3× Control plane nodes** (HA)
- [ ] **etcd backup automation**
- [ ] **Monitoring** (Prometheus, Grafana)
- [ ] **Logging** (ELK, Loki)
- [ ] **Security hardening**
- [ ] **Disaster recovery plan**
- [ ] **Upgrade procedures documented**
- [ ] **Team training complete**
- [ ] **Load testing done**
- [ ] **Security audit passed**

---

## Conclusion

### The Bottom Line

**Self-Managed Kubernetes on EC2 with Cilium:**

✅ **Avoids ALL EKS-Cilium traps**
✅ **Full control over both K8s and Cilium**
✅ **All Cilium features available**
✅ **Slightly cheaper than EKS**
✅ **No vendor lock-in**

❌ **More operational overhead**
❌ **No managed control plane**
❌ **No AWS SLA**

**Best For:**
- Teams with K8s expertise
- Need full Cilium features
- Want maximum control
- Can handle operations

**Use EKS if you prefer managed services.**
**Use AKS if you need official support.**

**Deploy with confidence - you own everything!** 🚀

---

## Quick Commands Reference

```bash
# Deploy cluster
pulumi up

# Initialize cluster
ssh ubuntu@$CP_IP "sudo /root/init-cluster.sh"

# Join workers
ssh ubuntu@$WORKER_IP "sudo kubeadm join..."

# Add node
pulumi up  # with increased worker_count

# Remove node
kubectl drain <node> && kubectl delete node <node>

# Upgrade Cilium
cilium upgrade --version 1.16.0

# Backup etcd
etcdctl snapshot save backup.db

# Check cluster health
kubectl get nodes
kubectl get pods -A
cilium status
```

---

*This is true self-managed Kubernetes - you have complete control!*
