# Azure Networking and Multi-AZ/Multi-Region Architecture

## Understanding Azure Availability Zones vs Regions

### Key Concepts

**Availability Zone (AZ)**
- Physical location within an Azure region
- Each region has 3+ AZs
- AZs are connected via high-speed fiber
- Latency: < 2ms between AZs
- Example: Canada Central has zones 1, 2, 3

**Region**
- Geographical area containing multiple datacenters
- Regions are far apart (100s-1000s of km)
- Independent power, cooling, networking
- Example: Canada Central, East US, West Europe

---

## How Subnets Work with Availability Zones

### Critical Fact

> **In Azure, subnets automatically span ALL availability zones within a region.**

You **cannot** assign a subnet to a specific AZ. The subnet is a logical construct that exists across all zones.

### Example

```
Canada Central Region
├── VNet: 10.0.0.0/16
    ├── Subnet: 10.0.1.0/24  ← This exists in ALL 3 zones
        ├── AZ 1 (physical datacenter 1)
        ├── AZ 2 (physical datacenter 2)
        └── AZ 3 (physical datacenter 3)
```

When you deploy a VM or AKS node pool:
- The **subnet** is the same
- The **physical placement** is determined by `availability_zones` parameter

---

## Single-Region Multi-AZ Deployment

This is what our default `example_usage.py` does:

```python
# One subnet spans all AZs
node_subnet = network.Subnet(
    "aks-node-subnet",
    address_prefix="10.0.0.0/16",  # Available in zones 1, 2, 3
)

# AKS node pools specify which zones to use
agent_pool_profiles=[
    ManagedClusterAgentPoolProfileArgs(
        name="system",
        availability_zones=["1", "2", "3"],  # Nodes distributed across zones
        vnet_subnet_id=subnet_id,  # Same subnet
    ),
]
```

### Result

```
Single Subnet: 10.0.0.0/16
├── Zone 1: Node IPs 10.0.0.5, 10.0.0.8, ...
├── Zone 2: Node IPs 10.0.0.6, 10.0.0.9, ...
└── Zone 3: Node IPs 10.0.0.7, 10.0.0.10, ...

All nodes share the subnet, but are in different physical locations.
```

### Benefits
- **99.99% SLA** (vs 99.9% for single-zone)
- Automatic failover between zones
- No routing complexity
- Seamless pod communication across zones

### Limitations
- Single region only
- Cannot survive regional outage
- All zones share same region's network

---

## Multi-Region Deployment

For true disaster recovery and global distribution:

### Architecture

```
Region 1: Canada Central
├── VNet: 10.0.0.0/14
│   └── Subnet: 10.0.0.0/16
│       ├── AZ 1, 2, 3
│       └── AKS Cluster 1

Region 2: East US
├── VNet: 10.4.0.0/14  ← Non-overlapping!
│   └── Subnet: 10.4.0.0/16
│       ├── AZ 1, 2, 3
│       └── AKS Cluster 2

Region 3: West Europe
├── VNet: 10.8.0.0/14  ← Non-overlapping!
    └── Subnet: 10.8.0.0/16
        ├── AZ 1, 2, 3
        └── AKS Cluster 3

Azure Front Door (Global)
├── Routes traffic to nearest region
└── Provides failover between regions
```

### CIDR Allocation

| Region | VNet CIDR | Node Subnet | Pod CIDR | Service CIDR |
|--------|-----------|-------------|----------|--------------|
| Canada Central | 10.0.0.0/14 | 10.0.0.0/16 | 10.32.0.0/13 | 10.96.0.0/12 |
| East US | 10.4.0.0/14 | 10.4.0.0/16 | 10.40.0.0/13 | 10.96.0.0/12 |
| West Europe | 10.8.0.0/14 | 10.8.0.0/16 | 10.48.0.0/13 | 10.96.0.0/12 |

**Important**: Service CIDR can be the same across regions because it's virtual (not routed).

### Deployment

```bash
# Enable multi-region mode
pulumi config set enable_multi_region true

# Deploy
pulumi up
```

### Benefits
- Survives entire region failure
- Global load distribution
- Lower latency for users worldwide
- Geographic compliance (data residency)

### Tradeoffs
- Higher cost (3x infrastructure)
- More complex networking
- Data replication challenges
- Stateful apps need special handling

---

## Comparison Table

| Feature | Single Zone | Single Region Multi-AZ | Multi-Region |
|---------|-------------|------------------------|--------------|
| **Subnets** | 1 | 1 (spans zones) | 3+ (one per region) |
| **VNets** | 1 | 1 | 3+ (one per region) |
| **AKS Clusters** | 1 | 1 | 3+ (one per region) |
| **Zones per Cluster** | 1 | 3 | 3 per cluster |
| **SLA** | 99.9% | 99.99% | 99.99%+ |
| **Survives** | Node failure | Zone failure | Region failure |
| **Latency** | Lowest | Low (<2ms) | Variable |
| **Cost** | $ | $$ | $$$ |
| **Complexity** | Low | Medium | High |

---

## Common Mistakes

### ❌ Trying to create subnets per AZ

```python
# WRONG! Azure doesn't support this
subnet_az1 = network.Subnet("subnet-az1", availability_zone="1")  # Not valid
```

**Why**: Subnets are regional, not zonal.

### ❌ Using same CIDR across regions

```python
# WRONG! CIDR conflict
region1_vnet = network.VirtualNetwork("vnet1", address_prefixes=["10.0.0.0/16"])
region2_vnet = network.VirtualNetwork("vnet2", address_prefixes=["10.0.0.0/16"])
```

**Why**: When you connect regions (VNet peering, VPN), overlapping IPs break routing.

### ❌ Forgetting to specify availability_zones

```python
# WRONG! Will deploy to single zone (poor availability)
agent_pool_profiles=[
    ManagedClusterAgentPoolProfileArgs(
        name="system",
        # availability_zones missing!
    ),
]
```

**Why**: Defaults to single zone, losing multi-AZ benefits.

---

## Recommendations

### Start Here: Single-Region Multi-AZ

```python
platform = AKSPlatform(
    "production",
    location="canadacentral",
    enable_multi_region=False,  # Start simple
)
```

**When to use**:
- Most production workloads
- 99.99% SLA sufficient
- Regional disaster acceptable
- Budget-conscious

### Scale to Multi-Region When...

```python
platform = AKSPlatform(
    "production",
    enable_multi_region=True,
    additional_regions=["eastus", "westeurope"],
)
```

**When to use**:
- Financial services (strict uptime)
- Global user base
- Regulatory requirements (data residency)
- Budget for 3x+ infrastructure

---

## Testing Multi-AZ

### Verify node distribution

```bash
kubectl get nodes -o json | jq '.items[] | {name: .metadata.name, zone: .metadata.labels["topology.kubernetes.io/zone"]}'
```

Expected output:
```json
{"name": "aks-system-12345", "zone": "canadacentral-1"}
{"name": "aks-system-67890", "zone": "canadacentral-2"}
{"name": "aks-system-23456", "zone": "canadacentral-3"}
```

### Simulate zone failure

```bash
# Cordon all nodes in zone 1
kubectl cordon -l topology.kubernetes.io/zone=canadacentral-1

# Pods should reschedule to zones 2 and 3
kubectl get pods -o wide
```

---

## Further Reading

- [Azure Availability Zones](https://learn.microsoft.com/en-us/azure/reliability/availability-zones-overview)
- [AKS Multi-AZ Best Practices](https://learn.microsoft.com/en-us/azure/aks/availability-zones)
- [Azure Virtual Network Design](https://learn.microsoft.com/en-us/azure/architecture/networking/guide/virtual-network-design)
