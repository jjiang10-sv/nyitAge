# AKS vs EKS Platform Complexity Comparison

## Quick Answer: EKS is SIMPLER! ✅

**EKS with Pure Cilium:**
- ✅ **One CNI** - Cilium does everything
- ✅ **No hybrid** - Simpler architecture
- ✅ **More flexibility** - Full Cilium control
- ✅ **Lower cost** - ~6x cheaper

---

## Code Complexity Comparison

### Lines of Code

| File | AKS | EKS | Winner |
|------|-----|-----|--------|
| **platform.py** | 533 lines | 390 lines | EKS (-27%) |
| **example_usage.py** | 316 lines | 240 lines | EKS (-24%) |
| **Total** | **849 lines** | **630 lines** | **EKS (-26%)** |

**EKS has 26% less code!**

---

## Architecture Complexity

### AKS (Hybrid Azure CNI + Cilium)

```
┌──────────────────────────────┐
│ Azure CNI Control Plane      │
│ ├─ IPAM                      │
│ ├─ Overlay infrastructure    │
│ ├─ Pod CIDR management       │
│ └─ VNet integration          │
└─────────────┬────────────────┘
              │
              ↓
┌──────────────────────────────┐
│ Azure Firewall (Required)    │  ← Extra complexity!
│ ├─ Firewall subnet           │
│ ├─ Public IP                 │
│ ├─ Route tables              │
│ └─ UDR configuration         │
└─────────────┬────────────────┘
              │
              ↓
┌──────────────────────────────┐
│ Cilium Dataplane             │
│ ├─ eBPF packet processing    │
│ ├─ Service load balancing    │
│ └─ Network policies          │
└──────────────────────────────┘
```

**Components:** 3 layers (Azure CNI + Firewall + Cilium)

---

### EKS (Pure Cilium)

```
┌──────────────────────────────┐
│ Cilium (Does Everything!)    │
│ ├─ IPAM (ENI or cluster-pool)│
│ ├─ Overlay (or native)       │
│ ├─ Pod CIDR management       │
│ ├─ eBPF dataplane            │
│ ├─ Service load balancing    │
│ └─ Network policies          │
└──────────────────────────────┘
```

**Components:** 1 layer (just Cilium!)

---

## Setup Complexity

### AKS Setup Steps

```python
# 1. Create VNet
vnet = network.VirtualNetwork(...)

# 2. Create node subnet
node_subnet = network.Subnet(...)

# 3. Create firewall subnet (required!)
fw_subnet = network.Subnet(
    subnet_name="AzureFirewallSubnet",  # Must be exact name!
    address_prefix="10.0.128.0/26",
)

# 4. Create firewall public IP
fw_ip = network.PublicIPAddress(...)

# 5. Create Azure Firewall
firewall = network.AzureFirewall(...)

# 6. Get firewall private IP
fw_private_ip = firewall.ip_configurations[0].private_ip_address

# 7. Create route table
route_table = network.RouteTable(
    routes=[{
        "next_hop_type": "VirtualAppliance",
        "next_hop_ip_address": fw_private_ip,  # Route to firewall
    }]
)

# 8. Associate route table with subnet
network.SubnetRouteTableAssociation(...)

# 9. Create AKS cluster (with hybrid CNI)
cluster = containerservice.ManagedCluster(
    network_profile={
        "network_plugin": "azure",       # Azure CNI required
        "network_plugin_mode": "overlay",
        "network_dataplane": "cilium",   # Cilium as dataplane only
        "outbound_type": "USER_DEFINED_ROUTING",  # Use firewall
    }
)

# 10. Install Cilium (limited features)
helm.Chart("cilium", ...)
```

**Steps:** 10 major components
**Azure Firewall:** Required (~$1,200/month)
**Complexity:** High

---

### EKS Setup Steps

```python
# 1. Create VPC
vpc = ec2.Vpc(...)

# 2. Create subnets
subnets = [ec2.Subnet(...) for _ in range(3)]

# 3. Create internet gateway
igw = ec2.InternetGateway(...)

# 4. Create route table (simple!)
route_table = ec2.RouteTable(
    routes=[{"cidr_block": "0.0.0.0/0", "gateway_id": igw.id}]
)

# 5. Create EKS cluster (remove VPC CNI!)
cluster = eks.Cluster(
    default_addons_to_remove=["vpc-cni"],  # Remove AWS CNI!
)

# 6. Install Cilium (full features!)
helm.Chart("cilium", values={
    "ipam": {"mode": "eni"},       # Native or overlay
    "tunnel": "disabled",          # Or "geneve"
    # Full Cilium control!
})
```

**Steps:** 6 major components
**Firewall:** Not required (optional NAT Gateway ~$45/month)
**Complexity:** Low

---

## Feature Availability

### Control & Flexibility

| Feature | AKS | EKS |
|---------|-----|-----|
| **IPAM Control** | ❌ Azure controls | ✅ Cilium controls |
| **Overlay Control** | ❌ Azure controls | ✅ Cilium controls |
| **Native Routing** | ❌ Not available | ✅ ENI mode |
| **Custom Pod CIDR** | ⚠️ Limited | ✅ Full control |
| **BGP** | ⚠️ Limited | ✅ Full support |
| **Cluster Mesh** | ⚠️ Limited | ✅ Full support |

### What You Get

| Feature | AKS | EKS |
|---------|-----|-----|
| **eBPF Dataplane** | ✅ | ✅ |
| **Hubble** | ✅ | ✅ |
| **Network Policies** | ✅ | ✅ |
| **Service Mesh** | ✅ | ✅ |
| **Gateway API** | ✅ | ✅ |
| **Official Support** | ✅ Microsoft | ❌ Community |

---

## Cost Comparison

### Single Region

**AKS:**
```
System nodes (3):        $350/month
Workload nodes (6):      $1,050/month
Azure Firewall:          $1,200/month  ← Expensive!
NAT Gateway:             $45/month
Key Vault:               $5/month
──────────────────────────────────────
Total:                   ~$2,650/month
```

**EKS:**
```
System nodes (3):        $90/month
Workload nodes (6):      $260/month
EKS control plane:       $73/month
NAT Gateway:             $45/month
──────────────────────────────────────
Total:                   ~$470/month
```

**EKS is 5.6x cheaper!** 💰

---

### Multi-Region (3 regions)

**AKS:**
```
3 regions × $2,650 =     $7,950/month
Azure Front Door:        $35/month
VNet peering:            $100/month
──────────────────────────────────────
Total:                   ~$8,085/month
```

**EKS:**
```
3 regions × $470 =       $1,410/month
CloudFront:              $50/month
VPC peering:             $40/month
──────────────────────────────────────
Total:                   ~$1,500/month
```

**EKS is 5.4x cheaper!** 💰💰

---

## Deployment Comparison

### AKS

```bash
cd aks/
pulumi config set gitops_repo https://github.com/org/gitops
pulumi up

# Wait 30-45 minutes (firewall is slow)
# Cost: $2,650/month

# Get kubeconfig
az aks get-credentials --resource-group ... --name ...

# Verify hybrid setup
kubectl exec -n kube-system ds/cilium -- cilium status
# IPAM: Azure  ← Azure controls IPs
# Encapsulation: Geneve  ← Azure overlay
```

---

### EKS

```bash
cd eks/
pulumi config set gitops_repo https://github.com/org/gitops
pulumi config set cilium_mode eni  # Native routing
pulumi up

# Wait 15-20 minutes (no firewall!)
# Cost: $470/month

# Get kubeconfig
aws eks update-kubeconfig --name prod-usw2-eks

# Verify pure Cilium
kubectl exec -n kube-system ds/cilium -- cilium status
# IPAM: ENI  ← Cilium controls IPs!
# Encapsulation: Disabled  ← Native routing!
```

---

## Why EKS is Simpler

### 1. No CNI Split

**AKS:**
- Azure CNI for control plane
- Cilium for dataplane
- Need to understand both

**EKS:**
- Cilium for everything
- Simpler mental model

---

### 2. No Mandatory Firewall

**AKS:**
- Azure Firewall required for private clusters
- Complex UDR setup
- Expensive ($1,200/month)

**EKS:**
- Optional NAT Gateway
- Simple routing
- Cheap ($45/month)

---

### 3. More Cilium Features

**AKS:**
- Limited to dataplane features
- Can't use custom IPAM
- Limited Cluster Mesh

**EKS:**
- Full Cilium feature set
- All advanced features
- Complete control

---

### 4. Flexible Networking

**AKS:**
- Always uses overlay
- Always uses Geneve
- Can't optimize

**EKS:**
- ENI mode = native routing (fastest!)
- Overlay mode = massive scale
- You choose!

---

## When to Use Each

### Use AKS if:

✅ You need **official Microsoft support**
✅ You're **committed to Azure** ecosystem
✅ You want **managed Cilium updates**
✅ You need **Azure integration** (Firewall, Policy, Monitor)
✅ Enterprise SLA is critical

**Trade-offs:**
- More expensive
- Less flexibility
- More complexity

---

### Use EKS if:

✅ You want **pure Cilium** with all features
✅ You need **lower cost** (5-6x cheaper!)
✅ You value **simplicity** (26% less code)
✅ You want **maximum performance** (ENI mode)
✅ You need **full Cluster Mesh**

**Trade-offs:**
- No official AWS support (community only)
- You manage Cilium updates

---

## Migration Path

### From AKS to EKS (If You Want Pure Cilium)

```bash
# 1. Deploy EKS cluster
cd eks/
pulumi config set cilium_mode eni
pulumi up

# 2. Backup AKS workloads
velero backup create aks-backup --include-namespaces '*'

# 3. Restore to EKS
velero restore create --from-backup aks-backup

# 4. Update DNS/Front Door

# 5. Decommission AKS
cd ../aks/
pulumi destroy

# Save $2,180/month per region! 💰
```

---

## Summary

### Complexity

| Metric | AKS | EKS | Winner |
|--------|-----|-----|--------|
| **Lines of Code** | 849 | 630 | EKS (-26%) |
| **Components** | 3 layers | 1 layer | EKS |
| **Setup Steps** | 10 | 6 | EKS |
| **Mental Model** | Hybrid | Pure | EKS |

### Cost

| Deployment | AKS | EKS | Savings |
|------------|-----|-----|---------|
| **Single Region** | $2,650 | $470 | **$2,180** (5.6x) |
| **Multi-Region** | $8,085 | $1,500 | **$6,585** (5.4x)|

### Features

- **Same eBPF performance** ✅
- **Same Hubble observability** ✅
- **EKS has more Cilium features** ✅
- **AKS has official support** ✅

---

## Bottom Line

**For Pure Cilium Setup:**
→ **EKS is SIMPLER, CHEAPER, and MORE FLEXIBLE** ✅

**For Enterprise Support:**
→ **AKS has official Microsoft support** ✅

**Your choice depends on:**
- Need official support? → AKS
- Want lowest cost & simplicity? → EKS
- Need full Cilium features? → EKS
- Already on Azure? → AKS

**Both are excellent platforms!** The EKS implementation demonstrates that pure Cilium is actually simpler than the hybrid approach. 🚀
