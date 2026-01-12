# Quick Reference: CIDR Strategies

## Strategy Comparison

### 1️⃣ Single Subnet (Default) ✅

```
┌─────────────────────────────────────────────────────┐
│ VNet: 10.0.0.0/14                                   │
│                                                     │
│  ┌───────────────────────────────────────────────┐ │
│  │ Subnet: 10.0.0.0/16 (spans all AZs)           │ │
│  │                                                │ │
│  │  AZ-1  AZ-2  AZ-3                             │ │
│  │   │     │     │                                │ │
│  │  Node  Node  Node                             │ │
│  │  .0.5  .0.6  .0.7                             │ │
│  └───────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘

Pod CIDR: 10.32.0.0/13 (overlay, not in VNet)
Service CIDR: 10.96.0.0/12 (virtual IPs)
```

**Deploy:**
```python
platform = AKSPlatform(
    "prod",
    subnet_id=node_subnet.id,  # One subnet
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)
```

---

### 2️⃣ Per-AZ Subnets (1230.md Pattern) 🔧

```
┌──────────────────────────────────────────────────────┐
│ VNet: 10.0.0.0/14                                    │
│                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │ Subnet AZ-1  │  │ Subnet AZ-2  │  │ Subnet AZ-3│ │
│  │ 10.0.0.0/16  │  │ 10.1.0.0/16  │  │10.2.0.0/16 │ │
│  │              │  │              │  │            │ │
│  │  Nodes:      │  │  Nodes:      │  │  Nodes:    │ │
│  │  10.0.x.x    │  │  10.1.x.x    │  │  10.2.x.x  │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
└──────────────────────────────────────────────────────┘

Pod CIDR: 10.32.0.0/13 (overlay, not in VNet)
Service CIDR: 10.96.0.0/12 (virtual IPs)
```

**Deploy:**
```bash
pulumi config set use_per_az_subnets true
```

---

## Complete CIDR Map

### Single Region Setup

| Layer | CIDR | Size | Purpose |
|-------|------|------|---------|
| **VNet** | `10.0.0.0/14` | 262,144 IPs | Virtual network |
| **Nodes** | `10.0.0.0/16` | 65,536 IPs | AKS VMs (all AZs) |
| **Firewall** | `10.0.128.0/26` | 64 IPs | Azure Firewall |
| **Pods** | `10.32.0.0/13` | 524,288 IPs | Container overlay |
| **Services** | `10.96.0.0/12` | 1M IPs | Virtual IPs |
| **DNS** | `10.96.0.10` | 1 IP | CoreDNS service |

### Multi-Region Setup

| Region | VNet | Nodes | Pods |
|--------|------|-------|------|
| **Canada** | 10.0.0.0/14 | 10.0.0.0/16 | 10.32.0.0/13 |
| **US East** | 10.4.0.0/14 | 10.4.0.0/16 | 10.40.0.0/13 |
| **EU West** | 10.8.0.0/14 | 10.8.0.0/16 | 10.48.0.0/13 |

**Note:** Service CIDR `10.96.0.0/12` can be **reused** across regions (it's virtual).

---

## Usage Examples

### Standard Production

```python
from platform import AKSPlatform

platform = AKSPlatform(
    "production",
    vnet_id=vnet.name,
    subnet_id=node_subnet.id,
    location="canadacentral",
    gitops_repo="https://github.com/org/gitops",
    pod_cidr="10.32.0.0/13",      # 524k pods
    service_cidr="10.96.0.0/12",  # 1M services
    # dns_service_ip auto = 10.96.0.10
)
```

### Small Dev Environment

```python
platform_dev = AKSPlatform(
    "dev",
    pod_cidr="10.40.0.0/16",      # 65k pods (smaller)
    service_cidr="10.100.0.0/16", # 65k services
)
```

### Multi-Region Production

```python
# Region 1
canada = AKSPlatform(
    "can-prod",
    location="canadacentral",
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)

# Region 2
us_east = AKSPlatform(
    "us-prod",
    location="eastus",
    pod_cidr="10.40.0.0/13",     # Different!
    service_cidr="10.96.0.0/12", # Same OK
)
```

---

## Visual: How IPs Are Allocated

### Per-Node Pod IP Allocation

```
Cluster Pod CIDR: 10.32.0.0/13
│
├── Node 1 gets: 10.32.0.0/23   (512 IPs, fits 250 pods)
├── Node 2 gets: 10.32.2.0/23   (512 IPs)
├── Node 3 gets: 10.32.4.0/23   (512 IPs)
├── Node 4 gets: 10.32.6.0/23   (512 IPs)
└── ... (can support 1,024 nodes)
```

### Service IP Allocation

```
Service CIDR: 10.96.0.0/12
│
├── 10.96.0.1    - Kubernetes API
├── 10.96.0.10   - CoreDNS
├── 10.96.10.5   - Your app service
├── 10.96.20.8   - Another service
└── ... (1M possible services)
```

---

## Configuration Commands

### View Current Config

```bash
pulumi config
```

### Set GitOps Repo

```bash
pulumi config set gitops_repo https://github.com/org/gitops
```

### Enable Per-AZ Subnets

```bash
pulumi config set use_per_az_subnets true
```

### Preview Changes

```bash
pulumi preview
```

### Deploy

```bash
pulumi up
```

### View Outputs

```bash
pulumi stack output

# Example output:
# pod_cidr: 10.32.0.0/13
# service_cidr: 10.96.0.0/12
# dns_service_ip: 10.96.0.10
```

---

## Verification Commands

### Check Cluster Network Config

```bash
az aks show \
  -g <resource-group> \
  -n <cluster-name> \
  --query networkProfile
```

### View Pod IPs

```bash
kubectl get pods -A -o wide | awk '{print $7}' | grep "^10"
# Should show IPs from your pod_cidr
```

### View Service IPs

```bash
kubectl get svc -A -o wide | awk '{print $4}' | grep "^10"
# Should show IPs from your service_cidr
```

### Check DNS Service

```bash
kubectl get svc -n kube-system kube-dns -o yaml | grep clusterIP
# Should match your dns_service_ip
```

---

## Troubleshooting

### CIDR Overlap Error

```
Error: CIDR ranges overlap
```

**Solution:** Ensure pod_cidr doesn't overlap with VNet or other clusters.

### DNS Resolution Fails

```bash
# Check DNS service IP is in service CIDR
kubectl get svc -n kube-system kube-dns

# Should be within your service_cidr range
```

### Pod CIDR Exhaustion

```
Error: No IP addresses available
```

**Solution:** Increase pod_cidr size:
```python
pod_cidr="10.32.0.0/12"  # Double the size
```

---

## Key Takeaways

1. **Subnets span all AZs** - No need for per-AZ subnets unless you want strict isolation
2. **Pod CIDR is separate** - Overlay network, doesn't consume VNet IPs
3. **Service CIDR is virtual** - Can reuse across regions
4. **DNS IP is auto-calculated** - 10th IP in service CIDR (e.g., 10.96.0.10)
5. **Always over-allocate** - CIDR changes are painful, plan for growth

---

## Files Reference

- **[platform.py](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/platform.py)** - Main component (now supports custom CIDRs)
- **[example_usage.py](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/example_usage.py)** - Both subnet strategies
- **[CIDR_GUIDE.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/CIDR_GUIDE.md)** - Detailed explanation
- **[ARCHITECTURE.md](file:///Users/john/Documents/projects/base/go-algorithms-master/nyit/RApro/devops/ARCHITECTURE.md)** - AZ vs Region concepts
