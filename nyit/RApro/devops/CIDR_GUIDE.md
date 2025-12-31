# CIDR Strategy Guide

## The 1230.md Pattern Explained

The diagram you referenced shows **per-AZ subnets**, which is an advanced pattern:

```
VPC CIDR:       10.0.0.0/14
├─ Nodes:       10.0.0.0/16  ← AZ-1 subnet
├─ Nodes:       10.1.0.0/16  ← AZ-2 subnet
├─ Nodes:       10.2.0.0/16  ← AZ-3 subnet
│
├─ Pod CIDR:    10.32.0.0/13
│  ├─ Node 1:   10.32.0.0/23
│  ├─ Node 2:   10.32.2.0/23
│  └─ ...
│
└─ Service CIDR:10.96.0.0/12
```

## Two Subnet Strategies

### Strategy 1: Single Subnet (Azure Standard)

**What we implemented initially:**

```python
vnet = VirtualNetwork(address_prefixes=["10.0.0.0/14"])
node_subnet = Subnet(address_prefix="10.0.0.0/16")  # Spans all AZs

platform = AKSPlatform(
    subnet_id=node_subnet.id,
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)
```

**IP Allocation:**
```
VPC: 10.0.0.0/14
└── Single Subnet: 10.0.0.0/16
    ├── AZ-1 Nodes: 10.0.0.5, 10.0.0.8, ...
    ├── AZ-2 Nodes: 10.0.0.6, 10.0.0.9, ...
    └── AZ-3 Nodes: 10.0.0.7, 10.0.0.10, ...
```

**Pros:**
- ✅ Simpler configuration
- ✅ Azure best practice
- ✅ Automatic AZ distribution
- ✅ Easier IP management

**Cons:**
- ❌ Can't tell AZ from IP alone
- ❌ No per-AZ network policies

---

### Strategy 2: Per-AZ Subnets (1230.md Pattern)

**Advanced pattern:**

```python
vnet = VirtualNetwork(address_prefixes=["10.0.0.0/14"])

# Create 3 subnets, one per AZ
subnet_az1 = Subnet(address_prefix="10.0.0.0/16")
subnet_az2 = Subnet(address_prefix="10.1.0.0/16")
subnet_az3 = Subnet(address_prefix="10.2.0.0/16")

# Create node pools with specific subnets
# (requires extending platform.py)
```

**IP Allocation:**
```
VPC: 10.0.0.0/14
├── AZ-1 Subnet: 10.0.0.0/16
│   └── Nodes: 10.0.x.x
├── AZ-2 Subnet: 10.1.0.0/16
│   └── Nodes: 10.1.x.x
└── AZ-3 Subnet: 10.2.0.0/16
    └── Nodes: 10.2.x.x
```

**Pros:**
- ✅ Clear AZ identification (IP → AZ mapping)
- ✅ Per-AZ network policies possible
- ✅ Easier troubleshooting
- ✅ Better blast radius control

**Cons:**
- ❌ More complex setup
- ❌ Requires careful CIDR planning
- ❌ Need per-AZ node pool configuration

---

## Why Pod CIDR and Service CIDR Are Separate

### Node Subnet vs Pod CIDR

```
Node Subnet:    10.0.0.0/16   ← VM/host IPs (VNet)
Pod CIDR:       10.32.0.0/13  ← Container IPs (overlay)
Service CIDR:   10.96.0.0/12  ← Virtual IPs (not routed)
```

**Key Differences:**

| Component | Type | Routed in VNet? | Size |
|-----------|------|----------------|------|
| Node Subnet | Physical | ✅ Yes | Need enough for VMs |
| Pod CIDR | Overlay | ❌ No (overlay) | Need HUGE (500k pods) |
| Service CIDR | Virtual | ❌ No (iptables/eBPF) | Can be reused across clusters |

### Why This Matters

**Azure CNI Overlay Mode:**
```
Node IP:    10.0.0.5     (from node subnet, VNet routable)
Pod IP:     10.32.1.25   (from pod CIDR, overlay network)
Service IP: 10.96.10.50  (virtual, eBPF routing)
```

Pod IPs are **encapsulated** and don't need VNet routing, so they can use a massive CIDR without consuming VNet IPs.

---

## Updated platform.py - Now Configurable!

### New Parameters

```python
platform = AKSPlatform(
    "production",
    pod_cidr="10.32.0.0/13",      # NEW: Configurable pod CIDR
    service_cidr="10.96.0.0/12",  # NEW: Configurable service CIDR
    dns_service_ip="10.96.0.10",  # NEW: Optional (auto-calculated if omitted)
)
```

### Auto-Calculated DNS Service IP

If you don't specify `dns_service_ip`, it's automatically calculated as the 10th IP in your service CIDR:

```python
service_cidr="10.96.0.0/12"  → dns_service_ip="10.96.0.10"
service_cidr="10.100.0.0/16" → dns_service_ip="10.100.0.10"
```

---

## Common CIDR Configurations

### Small Development Cluster

```python
platform = AKSPlatform(
    "dev",
    pod_cidr="10.40.0.0/16",      # 65,536 pods
    service_cidr="10.100.0.0/16", # 65,536 services
)
```

### Production Large Scale

```python
platform = AKSPlatform(
    "prod",
    pod_cidr="10.32.0.0/13",      # 524,288 pods
    service_cidr="10.96.0.0/12",  # 1,048,576 services
)
```

### Multi-Region (Non-Overlapping)

```python
# Region 1: Canada Central
region1 = AKSPlatform(
    "can-central",
    vnet_cidr="10.0.0.0/14",
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",  # Can reuse!
)

# Region 2: East US
region2 = AKSPlatform(
    "us-east",
    vnet_cidr="10.4.0.0/14",      # Different!
    pod_cidr="10.40.0.0/13",      # Different!
    service_cidr="10.96.0.0/12",  # Same is OK (virtual)
)

# Region 3: West Europe
region3 = AKSPlatform(
    "eu-west",
    vnet_cidr="10.8.0.0/14",      # Different!
    pod_cidr="10.48.0.0/13",      # Different!
    service_cidr="10.96.0.0/12",  # Same is OK (virtual)
)
```

---

## CIDR Planning Best Practices

### 1. Always Over-Allocate

```python
# BAD: Too small, will run out
pod_cidr="10.32.0.0/20"  # Only 4,096 pods

# GOOD: Room to grow
pod_cidr="10.32.0.0/13"  # 524,288 pods
```

### 2. Keep Service CIDR Large

```python
# Service CIDR should be HUGE (it's free, it's virtual)
service_cidr="10.96.0.0/12"  # 1M services, why not?
```

### 3. Document Your Allocation

```python
# Good practice: Add comments
platform = AKSPlatform(
    pod_cidr="10.32.0.0/13",      # 524k pods, /23 per node
    service_cidr="10.96.0.0/12",  # 1M services
    # Supports 2,000 nodes × 250 pods
)
```

### 4. Multi-Region CIDR Table

Keep a table of your allocations:

| Region | VNet | Nodes | Pods | Services |
|--------|------|-------|------|----------|
| Can Central | 10.0.0.0/14 | 10.0.0.0/16 | 10.32.0.0/13 | 10.96.0.0/12 |
| US East | 10.4.0.0/14 | 10.4.0.0/16 | 10.40.0.0/13 | 10.96.0.0/12 |
| EU West | 10.8.0.0/14 | 10.8.0.0/16 | 10.48.0.0/13 | 10.96.0.0/12 |

---

## How to Deploy

### Single Subnet (Default)

```bash
pulumi config set gitops_repo https://github.com/org/gitops
pulumi up
```

### Per-AZ Subnets (Advanced)

```bash
pulumi config set use_per_az_subnets true
pulumi up
```

### Custom CIDRs

Edit `example_usage.py`:
```python
platform = AKSPlatform(
    "production",
    pod_cidr="10.50.0.0/13",      # Your custom CIDR
    service_cidr="10.120.0.0/12", # Your custom CIDR
)
```

---

## Which Strategy Should You Choose?

### Use Single Subnet If:
- ✅ Standard production workload
- ✅ Want simplicity
- ✅ Following Azure best practices
- ✅ Don't need per-AZ policies

**→ This is recommended for 95% of use cases**

### Use Per-AZ Subnets If:
- ✅ Need strict AZ isolation
- ✅ Compliance requires AZ separation
- ✅ Want IP-based AZ identification
- ✅ Have complex per-AZ policies

**→ Only for advanced use cases**

---

## Verification

### Check Applied CIDRs

```bash
# Get cluster network profile
az aks show -g <rg> -n <cluster> --query networkProfile

# Should show your configured CIDRs:
{
  "podCidr": "10.32.0.0/13",
  "serviceCidr": "10.96.0.0/12",
  "dnsServiceIP": "10.96.0.10"
}
```

### Verify Pod IPs

```bash
kubectl get pods -A -o wide

# Pod IPs should be in your pod_cidr range
# e.g., 10.32.x.x
```

### Verify Service IPs

```bash
kubectl get svc -A

# Service IPs should be in your service_cidr range
# e.g., 10.96.x.x
```

---

## Summary

✅ **Updated `platform.py`** to support configurable `pod_cidr` and `service_cidr`
✅ **Added automatic DNS service IP calculation**
✅ **Explained both subnet strategies** (single vs per-AZ)
✅ **Provided examples** for dev, prod, and multi-region

The 1230.md diagram shows the **per-AZ subnet pattern**, which is more advanced. Your current implementation uses the simpler **single subnet** approach, which is Azure's standard recommendation and works great for most use cases!
