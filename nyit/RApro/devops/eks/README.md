# EKS Platform with Pure Cilium

## Why EKS + Cilium is SIMPLER than AKS

**No CNI Split! Cilium does everything:**

| Aspect | EKS Pure Cilium | AKS Hybrid |
|--------|----------------|------------|
| **Control Plane** | Cilium | Azure CNI |
| **Dataplane** | Cilium | Cilium |
| **IPAM** | Cilium | Azure CNI |
| **Overlay** | Cilium | Azure CNI |
| **Complexity** | ⭐ Simple | ⭐⭐ Split architecture |

---

## Quick Start

### Prerequisites

```bash
# Install tools
brew install pulumi awscli

# Configure AWS
aws configure

# Login to Pulumi
pulumi login
```

### Deploy Single Region

```bash
cd eks/
pulumi config set gitops_repo https://github.com/your-org/gitops
pulumi config set aws:region us-west-2
pulumi config set cilium_mode eni  # Native routing (fastest!)
pulumi up
```

### Deploy Multi-Region

```bash
pulumi config set multi_region true
pulumi up
```

---

## Cilium Modes

### ENI Mode (Recommended - FASTEST)

```bash
pulumi config set cilium_mode eni
```

**How it works:**
- Pods get real VPC IPs from AWS ENIs
- No overlay/encapsulation
- Native AWS routing
- Zero overhead

**Pros:**
- ✅ Fastest performance
- ✅ No encapsulation
- ✅ Works with AWS security groups
- ✅ Simple architecture

**Cons:**
- ❌ Uses VPC IP space
- ❌ ENI limits per instance type

---

### Overlay Mode (Maximum Scale)

```bash
pulumi config set cilium_mode overlay
```

**How it works:**
- Pods get IPs from pod CIDR (10.32.0.0/13)
- Geneve encapsulation
- Separate from VPC IPs

**Pros:**
- ✅ Massive scale (500k+ pods)
- ✅ Doesn't consume VPC IPs
- ✅ No ENI limits

**Cons:**
- ❌ ~5% performance overhead
- ❌ Can't use security groups per pod

---

## Features

### What You Get

✅ **Pure Cilium** - No AWS VPC CNI interference
✅ **eBPF Dataplane** - 10x faster than iptables
✅ **Hubble** - Network observability
✅ **Network Policies** - eBPF-based (fastest)
✅ **Gateway API** - Modern ingress
✅ **Argo CD ApplicationSet** - Advanced GitOps
✅ **SPIFFE/SPIRE** - Workload identity
✅ **BGP Support** - Available (not in AKS!)
✅ **Cluster Mesh** - Full support (limited in AKS!)

### Advantages Over AKS

| Feature | EKS | AKS |
|---------|-----|-----|
| **Pure Cilium** | ✅ Yes | ❌ No (hybrid) |
| **Custom IPAM** | ✅ Yes | ❌ No |
| **BGP** | ✅ Full | ⚠️ Limited |
| **Cluster Mesh** | ✅ Full | ⚠️ Limited |
| **Native Routing** | ✅ ENI mode | ❌ Always overlay |
| **Official Support** | ❌ Community | ✅ Microsoft |

---

## Architecture

### ENI Mode (Recommended)

```
┌────────────────────────────────┐
│ EC2 Worker Node                │
│                                │
│ Cilium Agent (eBPF)            │
│   ├─ Manages ENIs              │
│   ├─ Native routing            │
│   └─ No encapsulation          │
│                                │
│ Primary ENI: 10.0.1.5          │
│ Secondary ENI IPs:             │
│   ├─ Pod 1: 10.0.1.50          │
│   ├─ Pod 2: 10.0.1.51          │
│   └─ Pod 3: 10.0.1.52          │
│                                │
│ All real VPC IPs! ✅           │
└────────────────────────────────┘
```

### Overlay Mode

```
┌────────────────────────────────┐
│ EC2 Worker Node                │
│                                │
│ Node VPC IP: 10.0.1.5          │
│                                │
│ Cilium Overlay                 │
│   ├─ Pod 1: 10.32.1.10         │
│   ├─ Pod 2: 10.32.1.20         │
│   └─ Pod 3: 10.32.1.30         │
│                                │
│ Geneve encapsulation           │
└────────────────────────────────┘
```

---

## Comparison: EKS vs AKS Setup Complexity

### EKS (This Repo)

```python
# Simple! Just install Cilium
cluster = Cluster(
    "my-cluster",
    default_addons_to_remove=["vpc-cni"],  # Remove AWS CNI
)

# Install Cilium - it does everything
helm.Chart("cilium", values={
    "ipam": {"mode": "eni"},
    "tunnel": "disabled",
})
```

**Lines of code:** ~350
**Moving parts:** Cilium (one thing)

---

### AKS (From aks/ directory)

```python
# More complex - hybrid architecture
cluster = ManagedCluster(
    network_profile={
        "network_plugin": "azure",     # Azure CNI required
        "network_plugin_mode": "overlay",
        "network_dataplane": "cilium",  # Cilium as dataplane only
    }
)

# Azure Firewall setup (required for private clusters)
firewall = AzureFirewall(...)
route_table = RouteTable(...)

# Cilium install (but limited - Azure controls IPAM)
helm.Chart("cilium", ...)
```

**Lines of code:** ~570
**Moving parts:** Azure CNI + Cilium + Firewall + Routes

---

## Verification

### Check Cilium is Running

```bash
# Get kubeconfig
aws eks update-kubeconfig --name prod-usw2-eks

# Check Cilium status
kubectl -n kube-system exec ds/cilium -- cilium status

# Expected output:
# KubeProxyReplacement: Strict  ✅
# IPAM: ENI (or cluster-pool)  ✅
# Cilium: OK                    ✅
```

### Check No AWS VPC CNI

```bash
kubectl get pods -n kube-system | grep aws-node
# Should return nothing! ✅
```

### Test Network Performance

```bash
# Install netperf
kubectl run netperf-server --image=networkstatic/netperf
kubectl run netperf-client --image=networkstatic/netperf

# Run test
kubectl exec netperf-client -- netperf -H <server-ip>

# ENI mode: ~9-10 Gbps
# Overlay mode: ~8-9 Gbps (still excellent!)
```

---

## Cost Comparison

### Single Region

- **EC2 Nodes (6 total)**: ~$350/month
- **EKS Control Plane**: $73/month
- **Data transfer**: Variable
- **NAT Gateway**: ~$45/month
- **Total**: ~$470/month

**vs AKS Single Region (~$2,800/month)**
- EKS is **much cheaper** (no Azure Firewall required!)

### Multi-Region (3 regions)

- **Total**: ~$1,400/month

**vs AKS Multi-Region (~$8,000/month)**
- EKS is **~6x cheaper!**

---

## Migration from AKS

If you have AKS and want pure Cilium:

```bash
# 1. Deploy EKS cluster
cd eks/
pulumi up

# 2. Use Velero for backup/restore
velero backup create aks-backup
velero restore create --from-backup aks-backup

# 3. Update DNS to point to EKS

# 4. Decommission AKS
cd ../aks/
pulumi destroy
```

---

## Best Practices

### 1. Start with ENI Mode

```bash
pulumi config set cilium_mode eni
```

- Fastest performance
- Simpler architecture
- Most AWS-native

### 2. Use Overlay for Massive Scale

```bash
pulumi config set cilium_mode overlay
```

- When you need >100k pods
- When VPC IPs are limited

### 3. Enable Cluster Mesh for Multi-Region

```bash
cilium clustermesh enable --context us-west-2
cilium clustermesh enable --context us-east-1
cilium clustermesh connect --context us-west-2 --destination-context us-east-1
```

- Full cross-region service discovery
- Works perfectly (unlike AKS!)

---

## Troubleshooting

### Cilium Not Installing

```bash
# Check Cilium pods
kubectl get pods -n kube-system | grep cilium

# Check logs
kubectl logs -n kube-system ds/cilium
```

### ENI Limits Hit

```bash
# Check ENI usage
aws ec2 describe-network-interfaces --filters "Name=attachment.instance-id,Values=<instance-id>"

# Solution: Use larger instance types or overlay mode
```

### Performance Issues

```bash
# Check Cilium status
kubectl exec -n kube-system ds/cilium -- cilium status

# Run connectivity test
cilium connectivity test
```

---

## Summary

### EKS + Pure Cilium is SIMPLER because:

1. ✅ **No CNI split** - Cilium does everything
2. ✅ **Fewer components** - No firewall/routing setup
3. ✅ **More flexibility** - Full Cilium features
4. ✅ **Lower cost** - ~6x cheaper than AKS
5. ✅ **Better performance** - ENI mode = zero overhead

### Trade-off:

- ❌ **No official AWS support** (community only)
- ✅ **AKS has official Microsoft support**

**For maximum Cilium features and lowest cost:** Use EKS
**For official enterprise support:** Use AKS

Both are excellent choices! 🚀
