# Multi-Region Deployment Guide

## Overview

The enhanced `example_usage.py` now supports **production-ready multi-region deployments** with:

- ✅ Separate VNets and AKS clusters per region
- ✅ Non-overlapping CIDR ranges
- ✅ Azure Front Door for global load balancing
- ✅ Optional VNet peering for cluster-to-cluster communication
- ✅ WAF protection
- ✅ Health probes and automatic failover

---

## Quick Start

### Single Region (Default)

```bash
# Configure GitOps repository
pulumi config set gitops_repo https://github.com/your-org/gitops

# Deploy to Canada Central only
pulumi up
```

**Result**: 1 cluster in Canada Central

---

### Multi-Region Deployment

```bash
# Enable multi-region mode
pulumi config set multi_region true

# Deploy
pulumi up
```

**Result**: 
- 3 clusters (Canada Central, East US, West Europe)
- Azure Front Door routes traffic to nearest healthy region
- Each region fully independent

---

### Multi-Region + VNet Peering

```bash
# Enable VNet peering for cluster-to-cluster communication
pulumi config set multi_region true
pulumi config set enable_vnet_peering true

# Deploy
pulumi up
```

**Result**:
- All regions deployed
- Full-mesh VNet peering
- Pods can communicate across regions

---

## Architecture

### Single Region

```
┌─────────────────────────────────────┐
│ Canada Central                      │
│                                     │
│  ┌───────────────────────────────┐ │
│  │ VNet: 10.0.0.0/14             │ │
│  │                                │ │
│  │  ┌──────────────────────────┐ │ │
│  │  │ AKS Cluster              │ │ │
│  │  │ - Nodes: 10.0.0.0/16     │ │ │
│  │  │ - Pods:  10.32.0.0/13    │ │ │
│  │  │ - 3 Availability Zones   │ │ │
│  │  └──────────────────────────┘ │ │
│  └───────────────────────────────┘ │
└─────────────────────────────────────┘
```

### Multi-Region

```
                  ┌─────────────────────┐
                  │  Azure Front Door   │
                  │  (Global WAF + LB)  │
                  └──────────┬──────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
          ▼                  ▼                  ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│ Canada Central   │ │ East US          │ │ West Europe      │
│                  │ │                  │ │                  │
│ VNet: 10.0.0.0   │ │ VNet: 10.4.0.0   │ │ VNet: 10.8.0.0   │
│ Pods: 10.32.0.0  │ │ Pods: 10.40.0.0  │ │ Pods: 10.48.0.0  │
│                  │ │                  │ │                  │
│ ┌──────────────┐ │ │ ┌──────────────┐ │ │ ┌──────────────┐ │
│ │ AKS Cluster  │ │ │ │ AKS Cluster  │ │ │ │ AKS Cluster  │ │
│ │ 3 AZs        │ │ │ │ 3 AZs        │ │ │ │ 3 AZs        │ │
│ └──────────────┘ │ │ └──────────────┘ │ │ └──────────────┘ │
└──────────────────┘ └──────────────────┘ └──────────────────┘
         ▲                   ▲                   ▲
         └───────────────────┴───────────────────┘
              VNet Peering (if enabled)
```

---

## CIDR Allocation

### Complete CIDR Map

| Region | VNet | Nodes | Pods | Services | Total Capacity |
|--------|------|-------|------|----------|----------------|
| **Canada Central** | 10.0.0.0/14 | 10.0.0.0/16 | 10.32.0.0/13 | 10.96.0.0/12 | 524k pods |
| **East US** | 10.4.0.0/14 | 10.4.0.0/16 | 10.40.0.0/13 | 10.96.0.0/12 | 524k pods |
| **West Europe** | 10.8.0.0/14 | 10.8.0.0/16 | 10.48.0.0/13 | 10.96.0.0/12 | 524k pods |

**Key Points:**
- ✅ No CIDR overlaps between regions
- ✅ Service CIDR reused (it's virtual, not routed)
- ✅ Each region supports 2,000+ nodes and 500k+ pods
- ✅ Room for future region expansion

---

## Features

### 1. Azure Front Door

**Global Load Balancing:**
- Routes users to nearest healthy region
- Sub-second failover
- SSL/TLS termination
- Anycast IP for global reach

**WAF Protection:**
- OWASP Top 10 (DefaultRuleSet 2.1)
- Bot mitigation
- Prevention mode (blocks attacks)
- Custom rules support

**Health Probes:**
- 30-second intervals
- HTTPS `/health` endpoint
- 3/4 successful samples required
- Auto-removes unhealthy origins

### 2. Regional Isolation

**Each region gets:**
- Dedicated resource group
- Independent VNet
- Separate AKS cluster
- Own Azure Firewall
- Isolated Key Vault

**Benefits:**
- Blast radius containment
- Independent deployments
- Regional compliance
- Easier troubleshooting

### 3. VNet Peering (Optional)

**When Enabled:**
- Full-mesh connectivity
- Pod-to-pod across regions
- Shared services possible
- Low latency (Azure backbone)

**When to Use:**
- Multi-region databases
- Shared cache layers
- Cross-region data sync
- Disaster recovery

---

## Cost Estimates

### Single Region
- AKS Nodes: ~$1,400/month
- Azure Firewall: ~$1,200/month
- Other services: ~$200/month
- **Total: ~$2,800/month**

### Multi-Region (3 regions)
- AKS Nodes: ~$4,200/month
- Azure Firewalls: ~$3,600/month
- Front Door: ~$35/month + data transfer
- VNet Peering (optional): ~$20/TB
- **Total: ~$8,000+/month**

**Cost Optimization Tips:**
1. Start with 2 regions instead of 3
2. Use smaller node pools initially
3. Enable cluster autoscaling
4. Use Azure Reserved Instances

---

## Deployment Steps

### 1. Prerequisites

```bash
# Install tools
brew install pulumi azure-cli

# Login
az login
pulumi login
```

### 2. Create Pulumi Project

```bash
cd /path/to/devops
pulumi new azure-python --name aks-platform
```

### 3. Configure

```bash
# Required
pulumi config set gitops_repo https://github.com/your-org/gitops

# Optional
pulumi config set multi_region true
pulumi config set enable_vnet_peering true
```

### 4. Preview

```bash
pulumi preview
```

Expected resources for multi-region:
- 3 Resource Groups
- 3 VNets
- 3 Subnets
- 3 Azure Firewalls
- 3 AKS Clusters
- 1 Front Door Profile
- 1 WAF Policy
- 6 VNet Peerings (if enabled)

### 5. Deploy

```bash
pulumi up
```

⏱️ **Deployment time**: 30-45 minutes for multi-region

### 6. Verify

```bash
# List all outputs
pulumi stack output

# Get Front Door URL
pulumi stack output front_door_url

# Get cluster names
pulumi stack output can_cluster_name
pulumi stack output eus_cluster_name
pulumi stack output euw_cluster_name
```

---

## Post-Deployment

### Connect to Clusters

```bash
# Canada Central
az aks get-credentials \
  --resource-group $(pulumi stack output can_resource_group) \
  --name $(pulumi stack output can_cluster_name)

# East US
az aks get-credentials \
  --resource-group $(pulumi stack output eus_resource_group) \
  --name $(pulumi stack output eus_cluster_name) \
  --overwrite-existing

# West Europe
az aks get-credentials \
  --resource-group $(pulumi stack output euw_resource_group) \
  --name $(pulumi stack output euw_cluster_name) \
  --overwrite-existing
```

### Switch Between Clusters

```bash
# List contexts
kubectl config get-contexts

# Switch to Canada
kubectl config use-context <canada-context>

# Switch to East US
kubectl config use-context <eastus-context>
```

### Deploy Applications via GitOps

Applications are automatically deployed via Argo CD ApplicationSet to **all** regions:

```bash
# Push to GitOps repo
cd gitops/
mkdir -p apps/frontend
cat > apps/frontend/deployment.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: app
        image: nginx:latest
EOF

git add .
git commit -m "Add frontend app"
git push

# Argo CD syncs to all regions automatically
```

---

## Testing Multi-Region

### 1. Deploy Test Application

```bash
# Create test app in GitOps repo
mkdir -p gitops/apps/test

cat > gitops/apps/test/service.yaml <<EOF
apiVersion: v1
kind: Service
metadata:
  name: test-svc
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: test
EOF

git add . && git commit -m "Add test" && git push
```

### 2. Verify Front Door Routing

```bash
# Get Front Door endpoint
FD_URL=$(pulumi stack output front_door_url)

# Test from different locations
curl -I $FD_URL

# Check X-Azure-Ref header to see which region served request
```

### 3. Simulate Regional Failure

```bash
# Stop Canada cluster (simulated failure)
kubectl --context canada-context scale deployment --all --replicas=0

# Front Door automatically routes to East US or West Europe
curl $FD_URL  # Should still work!
```

### 4. Test VNet Peering (if enabled)

```bash
# Deploy pod in Canada
kubectl --context canada run test-can --image=busybox -- sleep 3600

# Deploy pod in East US
kubectl --context eastus run test-eus --image=busybox -- sleep 3600

# Get pod IPs
CAN_IP=$(kubectl --context canada get pod test-can -o jsonpath='{.status.podIP}')
EUS_IP=$(kubectl --context eastus get pod test-eus -o jsonpath='{.status.podIP}')

# Test cross-region connectivity
kubectl --context canada exec test-can -- ping -c 3 $EUS_IP
# Should work if VNet peering is enabled!
```

---

## Monitoring & Observability

### Check Cluster Health

```bash
# Per region
for ctx in canada eastus westeurope; do
  echo "=== $ctx ==="
  kubectl --context $ctx get nodes
  kubectl --context $ctx get pods -A
  cilium --context $ctx status
done
```

### Front Door Metrics

```bash
# View Front Door analytics in Azure Portal
az monitor metrics list \
  --resource $(pulumi stack output front_door_id) \
  --metric RequestCount,OriginHealthPercentage
```

### Application Logs

```bash
# View logs across all regions
for ctx in canada eastus westeurope; do
  echo "=== Logs from $ctx ==="
  kubectl --context $ctx logs -l app=frontend --tail=10
done
```

---

## Disaster Recovery

### Regional Failover

Front Door handles automatic failover:

1. **Health probe fails** in one region
2. **Front Door marks origin unhealthy** (after 3/4 failed probes)
3. **Traffic automatically routed** to healthy regions
4. **Recovery**: Once healthy, traffic gradually returns

### Database Considerations

For stateful applications:

```yaml
# Use Azure Cosmos DB (multi-region writes)
# or PostgreSQL Flexible Server with geo-replication

# Example: Cosmos DB connection
apiVersion: v1
kind: Secret
metadata:
  name: cosmos-connection
data:
  endpoint: <multi-region-cosmos-endpoint>
```

### GitOps Sync

All regions sync from **same GitOps repo**:
- Consistent configuration
- Single source of truth
- Atomic multi-region updates

---

## Troubleshooting

### Cluster Not Accessible

```bash
# Check private cluster access
# Must be on VNet or use bastion/VPN

# Option: Deploy jumpbox
az vm create \
  --resource-group platform-can-rg \
  --name jumpbox \
  --image UbuntuLTS \
  --vnet-name platform-can-vnet \
  --subnet aks-can-subnet
```

### Front Door Not Routing

```bash
# Check origin health
az afd endpoint show \
  --resource-group <rg> \
  --profile-name global-frontdoor \
  --endpoint-name global-endpoint

# Verify origins are configured
az afd origin list \
  --resource-group <rg> \
  --profile-name global-frontdoor \
  --origin-group-name aks-clusters
```

### VNet Peering Not Working

```bash
# Check peering status
az network vnet peering list \
  --resource-group platform-can-rg \
  --vnet-name platform-can-vnet

# Ensure both directions are "Connected"
```

---

## Migration from Single to Multi-Region

### Zero-Downtime Migration

1. **Deploy additional regions** (doesn't affect existing)
   ```bash
   pulumi config set multi_region true
   pulumi up  # Adds new regions
   ```

2. **Verify new regions** are healthy
   ```bash
   kubectl --context eastus get nodes
   kubectl --context westeurope get nodes
   ```

3. **Enable Front Door** (routes traffic)
   - Starts sending traffic to all regions
   - Existing region continues running

4. **Monitor metrics** for 24-48 hours

5. **Done!** You now have multi-region with zero downtime

---

## Best Practices

### 1. Always Use GitOps

```bash
# Don't kubectl apply directly
# Instead, commit to Git:
cd gitops/
git add apps/
git commit -m "Deploy new feature"
git push

# Argo CD syncs to all regions automatically
```

### 2. Test Regional Failover

```bash
# Regularly test failover scenarios
# Automate with chaos engineering tools
```

### 3. Monitor Cross-Region Latency

```bash
# Set up Prometheus queries for inter-region traffic
# Alert if latency spikes
```

### 4. Use Blue/Green Per Region

```bash
# Deploy to one region first
# Verify health
# Roll out to other regions
```

### 5. Plan for Data Sovereignty

```bash
# Keep EU user data in West Europe
# US data in East US/Canada
# Use Front Door routing rules
```

---

## Summary

✅ **Complete multi-region implementation**
- Separate infrastructure per region
- Non-overlapping CIDR ranges
- Azure Front Door with WAF
- Optional VNet peering
- GitOps-driven deployments

✅ **Production-ready features**
- Automatic failover
- Health monitoring
- Regional isolation
- Scalable to any number of regions

✅ **Easy deployment**
```bash
pulumi config set multi_region true
pulumi up
```

Your AKS platform is now globally distributed! 🌍
