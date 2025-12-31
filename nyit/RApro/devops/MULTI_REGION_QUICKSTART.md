# Multi-Region Quick Start Card

## 🚀 Deploy Single Region (Default)

```bash
pulumi config set gitops_repo https://github.com/your-org/gitops
pulumi up
```
**Result**: 1 cluster in Canada Central

---

## 🌍 Deploy Multi-Region

```bash
pulumi config set multi_region true
pulumi up
```
**Result**: 3 clusters (Canada, US, Europe) + Azure Front Door

---

## 🔗 Enable VNet Peering

```bash
pulumi config set multi_region true
pulumi config set enable_vnet_peering true
pulumi up
```
**Result**: Full-mesh connectivity between regions

---

## 📊 View Deployments

```bash
# List all outputs
pulumi stack output

# Key outputs
pulumi stack output deployment_mode          # "multi-region"
pulumi stack output regions                   # ["canadacentral", "eastus", ...]
pulumi stack output front_door_url           # https://... 
pulumi stack output cidr_allocation          # Full IP map
```

---

## 🔌 Connect to Clusters

```bash
# Canada
az aks get-credentials \
  --resource-group platform-can-rg \
  --name $(pulumi stack output can_cluster_name)

# East US  
az aks get-credentials \
  --resource-group platform-eus-rg \
  --name $(pulumi stack output eus_cluster_name)

# West Europe
az aks get-credentials \
  --resource-group platform-euw-rg \
  --name $(pulumi stack output euw_cluster_name)
```

---

## 🎯 Test Global Load Balancing

```bash
# Get Front Door URL
FD_URL=$(pulumi stack output front_door_url)

# Test from your location
curl -I $FD_URL

# Front Door routes to nearest healthy region
# Check 'X-Azure-Ref' header for region info
```

---

## 📦 Deploy App to All Regions

```bash
# Apps sync to ALL regions via Argo CD ApplicationSet
cd gitops/
mkdir -p apps/myapp

cat > apps/myapp/deployment.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: app
        image: nginx:latest
        ports:
        - containerPort: 80
EOF

git add . && git commit -m "Add myapp" && git push

# Argo CD syncs to Canada + East US + West Europe automatically!
```

---

## 🔍 Verify Multi-Region

```bash
# Check all clusters
for region in can eus euw; do
  echo "=== $region ==="
  kubectl --context ${region}-context get nodes
  kubectl --context ${region}-context get pods -A | grep myapp
done
```

---

## 💰 Cost Breakdown

### Single Region: ~$2,800/month
- 3 system nodes + 6 workload nodes
- Azure Firewall
- Key Vault, networking

### Multi-Region: ~$8,000/month
- 3× single-region infrastructure
- Azure Front Door (~$35)
- VNet peering (~$20/TB)

---

## 🗺️ CIDR Map

| Region | VNet | Pods | 
|--------|------|------|
| 🇨🇦 Canada | 10.0.0.0/14 | 10.32.0.0/13 |
| 🇺🇸 East US | 10.4.0.0/14 | 10.40.0.0/13 |
| 🇪🇺 W Europe | 10.8.0.0/14 | 10.48.0.0/13 |

**Services**: 10.96.0.0/12 (all regions, it's virtual)

---

## ✅ What You Get

### Single Region
✅ 1 AKS cluster (3 AZs)
✅ 99.99% SLA
✅ Private cluster + firewall
✅ Cilium + Hubble
✅ Argo CD + ApplicationSet
✅ SPIFFE/SPIRE
✅ Gateway API

### Multi-Region (all above +)
✅ 3 AKS clusters (9 total AZs)
✅ Azure Front Door
✅ WAF protection
✅ Global load balancing
✅ Auto-failover
✅ Optional VNet peering

---

## 🆘 Quick Troubleshooting

### Can't connect to cluster
```bash
# Private cluster requires VNet access
# Deploy jumpbox or use bastion
```

### Front Door not routing
```bash
# Check origins are healthy
az afd origin list \
  --profile-name global-frontdoor \
  --origin-group-name aks-clusters \
  --resource-group platform-can-rg
```

### Pod can't reach other region
```bash
# Ensure VNet peering is enabled
pulumi config set enable_vnet_peering true
pulumi up
```

---

## 📚 Full Docs

- **[MULTI_REGION_GUIDE.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/MULTI_REGION_GUIDE.md)** - Complete guide
- **[README.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/README.md)** - Platform overview
- **[CIDR_GUIDE.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/CIDR_GUIDE.md)** - CIDR planning
- **[ARCHITECTURE.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/ARCHITECTURE.md)** - AZ vs Region

---

## 🎓 Best Practices

1. **Always use GitOps** - Never kubectl apply directly
2. **Test failover** - Regularly simulate regional failures
3. **Monitor metrics** - Set up alerts for cross-region latency
4. **Start small** - Begin with 2 regions, add 3rd later
5. **Plan for data** - Use multi-region databases (Cosmos DB)

---

## Migration Path

Already have a single-region cluster? No problem:

```bash
# Step 1: Enable multi-region (keeps existing cluster)
pulumi config set multi_region true

# Step 2: Deploy (adds new regions, doesn't touch existing)
pulumi up

# Step 3: Front Door routes to all regions
# Zero downtime! 🎉
```

---

**Need help?** See [MULTI_REGION_GUIDE.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/MULTI_REGION_GUIDE.md) for detailed instructions.
