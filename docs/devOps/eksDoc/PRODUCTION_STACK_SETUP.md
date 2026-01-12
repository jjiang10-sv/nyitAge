# Production Stack Setup Guide

This guide shows how to create and configure a production Kubernetes stack.

## 🏭 Production Stack Configuration

### Step 1: Create Production Stack

```bash
cd /Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s

# Create prod stack
pulumi stack init prod

# Verify stack created
pulumi stack ls
# Should show: dev, prod
```

### Step 2: Configure Production Stack

```bash
# Switch to prod stack
pulumi stack select prod

# Configure AWS region (use region with 3+ AZs)
pulumi config set aws:region us-west-2

# Set cluster name
pulumi config set cluster_name prod-k8s

# Set Kubernetes and Cilium versions
pulumi config set k8s_version 1.30.0
pulumi config set cilium_version 1.16.5

# Control Plane: High Availability (NO SPOT)
pulumi config set control_plane_count 3
pulumi config set control_plane_instance_type t3.medium

# Workers: All on-demand for production
pulumi config set on_demand_worker_count 3
pulumi config set spot_worker_count 0  # No spot in production
pulumi config set worker_instance_type t3.large
```

### Step 3: Verify Configuration

```bash
pulumi config

# Expected output:
# KEY                           VALUE
# aws:region                    us-west-2
# cilium_version                1.16.5
# cluster_name                  prod-k8s
# control_plane_count           3
# control_plane_instance_type   t3.medium
# k8s_version                   1.30.0
# on_demand_worker_count        3
# spot_worker_count             0
# worker_instance_type          t3.large
```

### Step 4: Deploy Production Cluster

```bash
# Preview what will be created
pulumi preview

# Expected resources for prod:
# - 3 control plane nodes (on-demand)
# - 3 worker nodes (on-demand)
# - No spot instances
# Total: ~29 resources

# Deploy
pulumi up
```

## 📊 Stack Comparison

| Setting | Dev Stack | Prod Stack |
|---------|-----------|------------|
| **Region** | us-west-1 (2 AZs) | us-west-2 (4 AZs) |
| **Control Planes** | 1 × t3.small | 3 × t3.medium |
| **Workers (On-demand)** | 1 × t3.small | 3 × t3.large |
| **Workers (Spot)** | 3 × t3.small | 0 (none) |
| **Total Instances** | 5 | 6 |
| **Monthly Cost** | ~$35-44 | ~$360 |
| **Availability** | 99% | 99.95% |
| **Use Case** | Development/Testing | Production |

## 🔐 Production Best Practices

### Why No Spot in Production?

```yaml
# ❌ Not in production config
spot_worker_count: 0
```

**Reasons:**
1. **Predictable performance**: No interruptions
2. **SLA compliance**: Guaranteed capacity
3. **Incident response**: Always able to scale
4. **Customer trust**: No unexpected downtime

### Why 3 Control Planes?

```yaml
# ✅ Production config
control_plane_count: 3
```

**Benefits:**
- **High Availability**: Cluster survives 1 CP failure
- **etcd Quorum**: 2/3 CPs = cluster operational
- **Zero Downtime**: Rolling updates without outage
- **Production Grade**: Industry standard

### Why Larger Instances?

```yaml
# Production uses larger instances
control_plane_instance_type: t3.medium  # vs t3.small
worker_instance_type: t3.large          # vs t3.small
```

**Justification:**
- t3.medium: More headroom for API server load
- t3.large: Handle production workload density
- Better performance under load
- Fewer hosts to manage

## 🚀 Quick Stack Switching

### Work on Dev

```bash
pulumi stack select dev
pulumi preview
pulumi up
```

### Deploy to Prod

```bash
pulumi stack select prod
pulumi preview
pulumi up
```

### Check Current Stack

```bash
pulumi stack
# Shows: Currently selected stack is prod
```

## 💰 Cost Management

### Dev Stack (~$35-44/month)

```
1 × t3.small CP (on-demand):     $15/mo
1 × t3.small worker (on-demand): $15/mo
3 × t3.small workers (spot):     $5-14/mo
EBS storage (250GB):             $20/mo
──────────────────────────────────────
Total:                           $55-64/mo
```

### Prod Stack (~$360/month)

```
3 × t3.medium CP (on-demand):    $90/mo
3 × t3.large workers (on-demand): $240/mo
EBS storage (450GB):             $36/mo
──────────────────────────────────────
Total:                           $366/mo
```

**Cost Optimization Options:**
- Use Reserved Instances: Save 30-60%
- Savings Plans: Flexible discounts
- Right-size after monitoring: Adjust instance types

## 📝 Complete Configuration Files

### Pulumi.dev.yaml (Development)

```yaml
config:
  aws:region: us-west-1
  cluster_name: dev-k8s
  k8s_version: 1.30.0
  cilium_version: 1.16.5
  
  # Control Plane Configuration
  control_plane_count: 1
  control_plane_instance_type: t3.small
  
  # Worker Configuration (mixed)
  on_demand_worker_count: 1
  spot_worker_count: 3
  worker_instance_type: t3.small
```

### Pulumi.prod.yaml (Production)

> [!NOTE]
> This file is gitignored. Create it manually after running `pulumi stack init prod` and use the commands above to configure it.

```yaml
config:
  aws:region: us-west-2
  cluster_name: prod-k8s
  k8s_version: 1.30.0
  cilium_version: 1.16.5
  
  # Control Plane: HA
  control_plane_count: 3
  control_plane_instance_type: t3.medium
  
  # Workers: All on-demand
  on_demand_worker_count: 3
  spot_worker_count: 0
  worker_instance_type: t3.large
```

## 🔍 Validation Checklist

Before deploying to production:

- [ ] Stack name is `prod` (not `dev`)
- [ ] Region has 3+ availability zones
- [ ] `control_plane_count: 3` (HA)
- [ ] `spot_worker_count: 0` (no spot)
- [ ] Instance types sized appropriately
- [ ] SSH key exists in target region
- [ ] AWS credentials have necessary permissions
- [ ] Backup plan for etcd data
- [ ] Monitoring/alerting configured
- [ ] Cost budget approved

## 🛡️ Production Deployment Workflow

```bash
# 1. Switch to prod stack
pulumi stack select prod

# 2. Verify configuration
pulumi config
pulumi stack

# 3. Preview changes
pulumi preview > preview.txt
# Review preview.txt carefully

# 4. Deploy during maintenance window
pulumi up --yes

# 5. Verify deployment
ssh -i ~/.ssh/k8s-prod-key.pem ubuntu@$(pulumi stack output control_plane_public_ips | head -n1)
kubectl get nodes
kubectl get pods -A

# 6. Monitor for issues
kubectl top nodes
kubectl top pods -A
```

## 📚 Related Documentation

- [`SPOT_INSTANCE_BEST_PRACTICES.md`](./SPOT_INSTANCE_BEST_PRACTICES.md) - Why no spot in prod
- [`COST_OPTIMIZATION.md`](./COST_OPTIMIZATION.md) - Cost comparison
- [`UPGRADING_CLUSTER.md`](./UPGRADING_CLUSTER.md) - Production upgrade procedure

---

**Remember**: Production requires planning, testing in staging, and proper monitoring. Never experiment in prod! 🚨
