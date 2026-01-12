# Cluster Update Strategies: In-Place vs Recreation

This guide helps you decide when to update your cluster in-place versus recreating it from scratch.

## TL;DR - Quick Decision Guide

| Change Type | Recommended Approach | Why |
|-------------|---------------------|-----|
| **Cilium settings** | In-place update | Safe, supported, no data loss |
| **K8s minor version** (1.30 → 1.31) | In-place upgrade | Standard procedure, preserves state |
| **K8s major architecture** | Recreate | Clean slate, less risk |
| **Instance type changes** | Recreate (blue/green) | Infrastructure change |
| **Adding workers** | In-place | No disruption |
| **Infrastructure code** | Recreate (dev), In-place (prod) | Testing vs stability |

---

## Philosophy: Mutable vs Immutable Infrastructure

### Mutable (In-Place Updates)

**Concept**: Update running infrastructure
```
Cluster v1.30 → Update → Cluster v1.31
(Same instances, updated software)
```

**Pros:**
- ✅ Faster (no resource recreation)
- ✅ Preserves state and data
- ✅ Standard Kubernetes approach
- ✅ Less AWS cost (no double resources)

**Cons:**
- ⚠️ Configuration drift over time
- ⚠️ Harder to reproduce exact state
- ⚠️ Risk of incomplete updates

### Immutable (Recreation/Blue-Green)

**Concept**: Create new, migrate, destroy old
```
Old Cluster → New Cluster (created fresh) → Migrate → Delete old
```

**Pros:**
- ✅ Clean, reproducible state
- ✅ Easy rollback (keep old cluster)
- ✅ Test before cutover
- ✅ Infrastructure as code guaranteed

**Cons:**
- ⚠️ Slower (full recreation)
- ⚠️ Requires migration planning
- ⚠️ Double AWS cost during migration
- ⚠️ Stateful apps need careful handling

---

## Decision Matrix

### Update Cilium Settings

**Example**: Change Hubble settings, update Cilium version, modify network mode

**Recommended: In-Place Update** ✅

```bash
# SSH to control plane
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@CP_IP

# Update Cilium
cilium upgrade --version 1.16.5

# Or modify settings
kubectl edit ciliumconfig -n kube-system
```

**Why in-place:**
- Cilium designed for zero-downtime updates
- No data loss
- No instance recreation needed
- Industry standard approach

### Update Kubernetes Version

**Example**: 1.30 → 1.31

**Recommended: In-Place Upgrade** ✅

Follow [`UPGRADING_CLUSTER.md`](./UPGRADING_CLUSTER.md):

```bash
# 1. Upgrade control plane
ssh to CP1 → kubeadm upgrade apply v1.31.0

# 2. Upgrade workers one-by-one
kubectl drain worker-1
ssh to worker-1 → kubeadm upgrade node
kubectl uncordon worker-1
```

**Why in-place:**
- Official Kubernetes upgrade path
- Preserves etcd data
- Minimal downtime
- Battle-tested procedure

**When to recreate instead:**
- Upgrading from very old version (2+ years)
- Moving to different infrastructure
- Dev/test environment (faster than sequential upgrades)

### Change Infrastructure Configuration

**Example**: Instance types, network topology, region change

**Recommended: Recreate (Blue-Green)** ✅

```bash
# 1. Update Pulumi config
pulumi config set worker_instance_type t3.large

# 2. Create new stack (blue-green)
pulumi stack init prod-v2
pulumi config set ...
pulumi up  # New cluster created

# 3. Migrate workloads
# 4. Switch traffic
# 5. Destroy old
pulumi stack select prod
pulumi destroy
```

**Why recreate:**
- Infrastructure changes require new instances anyway
- Clean state guaranteed
- Can test before cutover
- Easy rollback if issues

### Update Pulumi/Terraform Code

**Example**: Changed security group rules, updated user-data scripts

**Recommended Approach Varies:**

| Environment | Approach | Reason |
|-------------|----------|--------|
| **Dev** | Recreate | Fast, no data loss concern |
| **Staging** | In-place (if safe) | Test update path |
| **Production** | Blue-Green | Zero downtime, rollback capability |

---

## In-Place Update Procedures

### 1. Cilium Configuration Changes

```bash
# Method 1: Cilium CLI
ssh ubuntu@CP_IP
cilium config set hubble-enabled true
cilium config set cluster-pool-ipv4-cidr 10.32.0.0/13

# Method 2: Helm (if installed via Helm)
helm upgrade cilium cilium/cilium \
  --namespace kube-system \
  --set hubble.enabled=true

# Method 3: Direct edit
kubectl edit ciliumconfig default -n kube-system
```

**Validation:**
```bash
cilium status
kubectl rollout status daemonset/cilium -n kube-system
```

### 2. Add/Remove Workers

```bash
# Add workers: Update Pulumi config
pulumi config set on_demand_worker_count 2  # Was 1
pulumi up  # Only creates new worker

# Remove workers: Drain then destroy
kubectl drain worker-3 --ignore-daemonsets
pulumi config set spot_worker_count 2  # Was 3
pulumi up
```

### 3. Update Application Settings

```bash
# Update ConfigMap
kubectl edit configmap app-config -n production

# OR: GitOps approach (recommended)
git commit -m "Update config"
git push
# ArgoCD/FluxCD auto-applies
```

---

## Recreation (Blue-Green) Procedures

### Full Cluster Recreation

**Use case**: Major infrastructure changes, testing, dev environments

**Step 1: Backup Current State**
```bash
# Backup etcd
ssh ubuntu@CP_IP
sudo /root/backup-etcd.sh

# Export current configs
kubectl get all --all-namespaces -o yaml > cluster-backup.yaml
```

**Step 2: Update Configuration**
```bash
# Update Pulumi config
pulumi config set k8s_version 1.31.0
pulumi config set control_plane_instance_type t3.medium
```

**Step 3: Destroy Old Cluster**
```bash
pulumi destroy
# Confirm with 'yes'
```

**Step 4: Create New Cluster**
```bash
pulumi up
# Review changes, confirm with 'yes'
```

**Step 5: Initialize**
```bash
# SSH to new control plane
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@NEW_CP_IP
sudo /root/init-cluster.sh
```

**Step 6: Restore Applications**
```bash
# Restore from GitOps or backup
kubectl apply -f cluster-backup.yaml
```

**Timeline:**
- Destroy: 2-3 minutes
- Create: 5-10 minutes
- Initialize: 10-15 minutes
- **Total: ~20-30 minutes**

### Blue-Green Deployment (Production)

**Use case**: Production upgrades with zero downtime

**Step 1: Create New Cluster (Blue)**
```bash
# Create separate stack
pulumi stack init prod-blue
pulumi config set cluster_name prod-blue-k8s
pulumi config set k8s_version 1.31.0
pulumi up
```

**Step 2: Deploy Applications to Blue**
```bash
# Point kubectl to blue cluster
export KUBECONFIG=~/.kube/config-blue

# Deploy apps
argocd app sync --all
# OR
kubectl apply -f manifests/
```

**Step 3: Test Blue Cluster**
```bash
# Run smoke tests
kubectl run test --image=curlimages/curl --rm -it -- curl http://app-service
```

**Step 4: Switch Traffic**
```bash
# Update DNS/Load Balancer to point to blue cluster
# OR: Update ingress controller
```

**Step 5: Monitor**
```bash
# Watch for errors
kubectl logs -f deployment/app
```

**Step 6: Destroy Green (Old)**
```bash
# After 24-48 hours of stability
pulumi stack select prod-green
pulumi destroy
```

**Timeline:**
- Create blue: 20-30 minutes
- Deploy/test: 1-2 hours
- Traffic switch: 5 minutes
- **Total: ~2-3 hours** (safe, zero downtime)

---

## Our Setup: What's Easy to Change

### ✅ Easy In-Place Updates

| Component | Update Method | Downtime |
|-----------|--------------|----------|
| Cilium version | `cilium upgrade` | None |
| Cilium settings | `kubectl edit` | None |
| Add workers | `pulumi up` | None |
| ConfigMaps | `kubectl edit` | None (apps restart) |
| Secrets | `kubectl edit` | None (apps restart) |

### ⚠️ Requires Careful Planning

| Component | Update Method | Downtime |
|-----------|--------------|----------|
| K8s version | Sequential upgrade | Minimal (rolling) |
| Remove workers | Drain + destroy | None (if >1 worker) |
| Control plane count | Recreate recommended | Full cluster |

### 🔴 Should Recreate

| Component | Reason |
|-----------|--------|
| Instance types | New instances needed anyway |
| Region change | Can't update in-place |
| VPC/networking | Infrastructure-level change |
| Init scripts | Already executed (baked in) |

---

## GitOps Approach (Recommended)

**Best practice**: Combine both approaches with GitOps

```
Infrastructure Code (Pulumi)
├─ Recreate when changed
└─ Stored in Git

Application Manifests (YAML)
├─ Auto-applied via ArgoCD
└─ Stored in Git
```

### Setup

```bash
# 1. Store Pulumi code in Git
git add example_usage.py Pulumi.*.yaml
git commit -m "Infrastructure as code"

# 2. Install ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# 3. Connect app repo
argocd app create myapp \
  --repo https://github.com/myorg/myapp \
  --path k8s \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace production
```

### Workflow

**Infrastructure changes (Pulumi):**
```bash
git commit -m "Update to t3.medium"
pulumi up  # Manual review and deploy
```

**Application changes (GitOps):**
```bash
git commit -m "Update app to v2.0"
git push
# ArgoCD auto-deploys (or with approval)
```

---

## Anti-Patterns to Avoid

### ❌ Don't: Manual SSH Changes Without Code

```bash
# BAD: SSH and manually install something
ssh ubuntu@worker-1
sudo apt install some-package
```

**Problem**: Lost on next recreation, not reproducible

**Better**: Add to user-data script or Ansible playbook

### ❌ Don't: Skip Testing in Dev

```bash
# BAD: Test new version directly in prod
pulumi stack select prod
pulumi config set k8s_version 1.31.0
pulumi up
```

**Problem**: No rollback plan if it fails

**Better**: Test in dev first, blue-green in prod

### ❌ Don't: Update Init Scripts on Running Cluster

Init scripts (`control-plane-init.sh`, `worker-init.sh`) run **once** at instance creation.

**These changes require recreation:**
- Kubernetes repo version
- Package installations
- Sysctl settings
- Containerd configuration

**These can be changed in-place:**
- Cilium (via `cilium upgrade`)
- Applications (via kubectl)
- ConfigMaps/Secrets

---

## Recommendations by Scenario

### Scenario 1: Dev Environment, Testing Changes

**Approach**: Destroy and recreate
```bash
pulumi destroy && pulumi up
```
1. Pulumi Updates - No Need to Destroy! ✅

When you change code and run pulumi up, Pulumi automatically handles updates
Adds new resource
Updates existing (in-place when possible)
Replaces when needed (like instance type changes)
You only manually destroy for complete recreation

**Why**: 
- Fast (20 min total)
- Guaranteed clean state
- No data to preserve
- Perfect for testing IaC changes

### Scenario 2: Production, Kubernetes Upgrade

**Approach**: In-place sequential upgrade
```bash
# Follow UPGRADING_CLUSTER.md
kubeadm upgrade apply v1.31.0
```

**Why**:
- Official K8s procedure
- Preserves data
- Minimal downtime
- Well-tested path

### Scenario 3: Production, Major Infrastructure Change

**Approach**: Blue-green deployment
```bash
# Create new, migrate, cutover
pulumi stack init prod-v2
pulumi up
# Migrate and test
# Switch traffic
# Destroy old
```

**Why**:
- Zero downtime
- Easy rollback
- Test before cutover
- Production-grade change management

### Scenario 4: Cilium Configuration Tweak

**Approach**: In-place edit
```bash
kubectl edit ciliumconfig default -n kube-system
```

**Why**:
- Designed for this
- No downtime
- Immediate effect
- Safest for live cluster

---

## Quick Reference Commands

### Check What Changed
```bash
# See what Pulumi will change
pulumi preview

# See diff
pulumi preview --diff
```

### Safe Update Workflow
```bash
# 1. Preview changes
pulumi preview > changes.txt
cat changes.txt  # Review carefully

# 2. Backup
ssh ubuntu@CP_IP sudo /root/backup-etcd.sh

# 3. Apply
pulumi up

# 4. Verify
kubectl get nodes
kubectl get pods -A
```

### Rollback

```bash
# In-place: Revert config
pulumi config set k8s_version 1.30.0
pulumi up

# Recreation: Use Git
git revert HEAD
pulumi up

# Blue-green: Switch traffic back
# Update DNS/LB to old cluster
```

---

## Summary

**General Guidelines:**

1. **Dev/Test**: Recreate freely - it's faster and cleaner
2. **Production Applications**: In-place updates via GitOps
3. **Production Infrastructure**: Blue-green for major changes
4. **Kubernetes Versions**: In-place sequential upgrades
5. **Cilium/Apps**: Always in-place (designed for it)

**Decision Tree:**

```
Need to update something?
├─ Is it infrastructure code?
│  ├─ Dev environment? → Destroy and recreate
│  └─ Production? → Blue-green deployment
├─ Is it Kubernetes version?
│  └─ Follow UPGRADING_CLUSTER.md (in-place)
├─ Is it Cilium?
│  └─ cilium upgrade (in-place)
└─ Is it an application?
   └─ Use GitOps (rolling update)
```

**The Golden Rule**: *If you can recreate it in 30 minutes, consider recreation. If downtime matters, plan blue-green.*

---

## Related Documentation

- [`UPGRADING_CLUSTER.md`](./UPGRADING_CLUSTER.md) - Detailed K8s upgrade procedure
- [`K8S_CLUSTER_BEST_PRACTICES.md`](./K8S_CLUSTER_BEST_PRACTICES.md) - Operational best practices
- [`PRODUCTION_STACK_SETUP.md`](./PRODUCTION_STACK_SETUP.md) - Production deployment guide
