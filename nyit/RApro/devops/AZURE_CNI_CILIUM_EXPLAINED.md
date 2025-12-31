# Azure CNI Overlay vs Cilium Overlay - The Truth

## Quick Answer

Your configuration uses: **Azure CNI Overlay with Cilium Dataplane** (HYBRID)

```python
network_profile=ContainerServiceNetworkProfileArgs(
    network_plugin="azure",           # ← Azure CNI (IPAM + overlay)
    network_plugin_mode="overlay",    # ← Azure manages overlay
    network_dataplane="cilium",       # ← Cilium handles packets (eBPF)
)
```

**This is NOT:**
- ❌ Pure Cilium overlay
- ❌ Pure Azure CNI overlay

**This IS:**
- ✅ **Azure CNI Overlay** (control plane + IPAM)
- ✅ **Cilium** (dataplane + eBPF packet processing)
- ✅ Best of both worlds!

---

## Understanding the Split

### Control Plane vs Dataplane

```
┌─────────────────────────────────────────────┐
│ CONTROL PLANE (Who decides what)            │
│ - IP address allocation (IPAM)              │
│ - Pod CIDR management                       │
│ - Overlay network setup                     │
│ - Route distribution                        │
│                                             │
│ Owner: Azure CNI                            │
└─────────────────┬───────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────┐
│ DATAPLANE (Who moves packets)               │
│ - Packet forwarding                         │
│ - Load balancing (services)                 │
│ - Network policies                          │
│ - Encapsulation/decapsulation               │
│                                             │
│ Owner: Cilium (eBPF)                        │
└─────────────────────────────────────────────┘
```

---

## Your Configuration Breakdown

### Parameter 1: `network_plugin="azure"`

**What It Means:**
```
Uses Azure CNI as the CNI plugin
```

**Azure CNI Responsibilities:**
- Allocates IP addresses to pods
- Manages the pod CIDR (10.32.0.0/13)
- Sets up the overlay network infrastructure
- Integrates with Azure VNet
- Handles pod-to-pod routing metadata

**NOT Cilium CNI!**

---

### Parameter 2: `network_plugin_mode="overlay"`

**What It Means:**
```
Azure CNI operates in overlay mode
(not traditional mode where pods get VNet IPs)
```

**Azure CNI Overlay:**
- Creates a **separate pod network** (10.32.0.0/13)
- Pods get IPs from this overlay range
- Azure manages the encapsulation infrastructure
- Uses **Geneve** protocol by default

**This is the key**: Azure, not Cilium, sets up the overlay.

---

### Parameter 3: `network_dataplane="cilium"`

**What It Means:**
```
Replace default dataplane with Cilium's eBPF dataplane
```

**Cilium Dataplane Responsibilities:**
- **Packet forwarding** using eBPF (not iptables)
- **Service load balancing** using eBPF maps
- **Network policies** using eBPF programs
- **Observability** (Hubble)
- **Advanced features** (BGP, service mesh, etc.)

**Cilium does NOT:**
- Manage IP allocation (Azure CNI does this)
- Set up overlay infrastructure (Azure does this)
- Control the pod CIDR (Azure does this)

---

## The Three Modes of AKS Networking

### Mode 1: Azure CNI (Traditional)

```python
network_profile=ContainerServiceNetworkProfileArgs(
    network_plugin="azure",
    # No network_plugin_mode specified
    # No network_dataplane specified
)
```

**Architecture:**
```
Pods get IPs from VNet directly
Pod CIDR = VNet subnet CIDR
Dataplane = Linux kernel + iptables
```

**Pros:**
- ✅ Simple integration with VNet
- ✅ Direct routing

**Cons:**
- ❌ Consumes VNet IPs rapidly
- ❌ IP exhaustion risk
- ❌ iptables performance issues at scale

---

### Mode 2: Azure CNI Overlay (without Cilium)

```python
network_profile=ContainerServiceNetworkProfileArgs(
    network_plugin="azure",
    network_plugin_mode="overlay",
    # network_dataplane not specified = default Linux kernel
)
```

**Architecture:**
```
Control Plane: Azure CNI
- IPAM
- Overlay setup (Geneve)
- Pod CIDR management

Dataplane: Linux kernel
- iptables for services
- Linux routing
- Standard kernel networking
```

**Pros:**
- ✅ Separate pod IP space
- ✅ No VNet IP exhaustion

**Cons:**
- ❌ iptables performance limits
- ❌ No advanced observability
- ❌ Limited network policy features

---

### Mode 3: Azure CNI Overlay + Cilium Dataplane (YOUR SETUP)

```python
network_profile=ContainerServiceNetworkProfileArgs(
    network_plugin="azure",
    network_plugin_mode="overlay",
    network_dataplane="cilium",  # 🎯 The difference!
)
```

**Architecture:**
```
Control Plane: Azure CNI
- IPAM (IP allocation)
- Overlay infrastructure (Geneve)
- Pod CIDR: 10.32.0.0/13
- Integration with Azure

Dataplane: Cilium
- eBPF packet processing
- eBPF service load balancing
- eBPF network policies
- Hubble observability
- Advanced features (BGP, mesh, etc.)
```

**Pros:**
- ✅ Separate pod IP space (no VNet exhaustion)
- ✅ eBPF performance (10x faster than iptables)
- ✅ Advanced network policies
- ✅ Hubble observability
- ✅ Azure integration
- ✅ Best of both worlds!

**Cons:**
- ⚠️ Slightly more complex
- ⚠️ Cilium learning curve

---

## Why This Hybrid Approach?

### What Microsoft Did

Microsoft wanted to leverage **Cilium's eBPF dataplane** without completely replacing Azure CNI:

```
┌───────────────────────────────────────────┐
│ Azure CNI                                 │
│ (What Microsoft knows well)               │
│ - Azure VNet integration                  │
│ - IPAM that works with Azure              │
│ - Overlay that integrates with SDN        │
└──────────────┬────────────────────────────┘
               │
               ├─────────────────────────────┐
               │                             │
               ↓                             ↓
┌──────────────────────────┐  ┌──────────────────────────┐
│ Option 1: Linux Kernel   │  │ Option 2: Cilium eBPF   │
│ - iptables               │  │ - eBPF programs         │
│ - Slow at scale          │  │ - 10x faster            │
│ - Limited features       │  │ - Advanced features     │
└──────────────────────────┘  └──────────────────────────┘
                                        ↑
                              You chose this! ✅
```

**Result:**
- Azure handles the **"Azure stuff"** (VNet, IPAM, overlay)
- Cilium handles the **"fast packet stuff"** (eBPF, policies, observability)

---

## Pure Cilium Overlay (For Comparison)

If you used **pure Cilium** (like in non-AKS Kubernetes):

```yaml
# Cilium ConfigMap (pure Cilium)
ipam:
  mode: kubernetes              # Cilium controls IP allocation
tunnel: geneve                 # Cilium manages overlay
datapath-mode: vxlan           # Cilium does encapsulation
```

**Differences:**
- Cilium manages **everything** (IPAM + dataplane)
- No Azure CNI involvement
- More Cilium-native features
- Less Azure integration

**In AKS, you can't do this!** Azure CNI is always the control plane.

---

## What Actually Happens in Your Setup

### Pod Creation Flow

```
1. Pod is scheduled to Node
   ↓
2. Azure CNI Plugin Called
   ├─ Allocates IP from pod CIDR (10.32.x.x)
   ├─ Sets up overlay interface
   ├─ Configures routes
   └─ Tells Cilium about the pod
   ↓
3. Cilium Takes Over
   ├─ Installs eBPF programs
   ├─ Sets up efficient packet handling
   ├─ Configures service load balancing
   └─ Enables Hubble monitoring
```

### Packet Flow (Pod to Pod)

```
┌─────────────────────────────────────────┐
│ Pod A (10.32.1.10) sends packet         │
└────────────────┬────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│ Cilium eBPF Program (on Node A)         │
│ - Looks up destination                  │
│ - Finds Pod B on Node B                 │
│ - Decides to encapsulate                │
└────────────────┬────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│ Azure CNI Overlay Infrastructure        │
│ - Uses Geneve encapsulation             │
│ - Outer: Node A IP → Node B IP          │
│ - Inner: Pod A IP → Pod B IP            │
└────────────────┬────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│ Azure VNet Routes Packet                │
│ - Sees Node A → Node B                  │
│ - Routes via VNet                       │
└────────────────┬────────────────────────┘
                 ↓
┌─────────────────────────────────────────┐
│ Cilium eBPF Program (on Node B)         │
│ - Decapsulates packet                   │
│ - Delivers to Pod B                     │
└─────────────────────────────────────────┘
```

**Key Point:** 
- **Encapsulation format**: Azure CNI's Geneve
- **Packet processing**: Cilium's eBPF

---

## Feature Ownership Table

| Feature | Azure CNI | Cilium | Who Wins? |
|---------|-----------|--------|-----------|
| **IP Address Allocation** | ✅ | ❌ | Azure CNI |
| **Pod CIDR Management** | ✅ | ❌ | Azure CNI |
| **Overlay Setup** | ✅ | ❌ | Azure CNI |
| **Encapsulation Protocol** | ✅ (Geneve) | ❌ | Azure CNI |
| **Packet Forwarding** | ❌ | ✅ (eBPF) | Cilium |
| **Service Load Balancing** | ❌ | ✅ (eBPF) | Cilium |
| **Network Policies** | ❌ | ✅ (eBPF) | Cilium |
| **Observability (Hubble)** | ❌ | ✅ | Cilium |
| **BGP Support** | ❌ | ✅ | Cilium |
| **Service Mesh** | ❌ | ✅ | Cilium |

---

## Verification Commands

### Check Who's Managing What

```bash
# Check CNI plugin
kubectl get pods -n kube-system -o wide | grep azure-cni
# Should see azure-cni-* pods

# Check Cilium dataplane
kubectl get pods -n kube-system -o wide | grep cilium
# Should see cilium-* pods

# Check Cilium status
kubectl -n kube-system exec ds/cilium -- cilium status

# Output shows:
# KubeProxyReplacement: Strict      ← Cilium handles services
# Cilium:               OK
# IPAM:                 Azure        ← Azure handles IP allocation! 🎯
```

### Check IPAM Mode

```bash
kubectl -n kube-system exec ds/cilium -- cilium status | grep -i ipam

# Expected output:
# IPAM: Azure                    ← Azure CNI manages IPs
# (NOT "IPAM: Cluster Pool" which would be pure Cilium)
```

### Check Overlay

```bash
kubectl -n kube-system exec ds/cilium -- cilium status | grep -i tunnel

# Expected output:
# Encapsulation: Geneve          ← Azure CNI's overlay
# (Cilium is aware but not managing it)
```

---

## Common Misconceptions

### ❌ Misconception 1: "Using Cilium = Pure Cilium"

**Reality:**
```
In AKS, Cilium is the DATAPLANE only.
Azure CNI is still the control plane.
```

### ❌ Misconception 2: "Azure CNI Overlay doesn't use Cilium"

**Reality:**
```
You can choose:
- Azure CNI Overlay + Linux kernel (default)
- Azure CNI Overlay + Cilium dataplane (our choice ✅)
```

### ❌ Misconception 3: "Cilium manages the overlay"

**Reality:**
```
Azure CNI sets up the overlay (Geneve).
Cilium uses it but doesn't control it.
```

---

## Why This Matters

### 1. Troubleshooting

**IP Address Issues:**
```bash
# Check Azure CNI logs (IPAM problems)
kubectl logs -n kube-system -l component=azure-cni

# Check Cilium logs (packet forwarding problems)
kubectl logs -n kube-system -l k8s-app=cilium
```

### 2. Configuration

**Pod CIDR changes:**
```python
# This is Azure CNI configuration
pod_cidr="10.32.0.0/13"  # ← Azure CNI uses this
```

**Cilium features:**
```yaml
# Cilium Helm values
hubble:
  enabled: true  # ← Cilium feature, works fine
```

### 3. Limitations

**Can't do (Azure CNI limitations):**
- Custom IPAM modes
- Pure Cilium cluster mesh (need workarounds)
- Direct routing without overlay (Azure decides)

**Can do (Cilium features):**
- eBPF-based network policies
- Hubble observability
- L7 traffic management
- BGP (with limitations)

---

## Benefits of This Hybrid

### Why Microsoft Chose This

1. **Azure Integration** ✅
   - Works with Azure VNet
   - Compatible with Azure Firewall
   - Integrates with Azure Policy

2. **Performance** ✅
   - eBPF is 10x faster than iptables
   - Better service load balancing
   - Lower CPU usage

3. **Observability** ✅
   - Hubble for network visibility
   - Better than basic Azure monitoring

4. **Stability** ✅
   - Azure CNI is battle-tested in AKS
   - Gradual Cilium adoption = less risk

5. **Flexibility** ✅
   - Can swap dataplane (Cilium ↔ default)
   - Azure CNI stays stable

---

## Summary

### Your Configuration:

```python
network_plugin="azure"           # Azure CNI is the boss
network_plugin_mode="overlay"    # Azure CNI creates overlay
network_dataplane="cilium"       # Cilium processes packets
```

### Answer: **Azure CNI Overlay** (with Cilium Dataplane)

**Division of Labor:**

| Responsibility | Owner |
|----------------|-------|
| IP Allocation | Azure CNI |
| Overlay Setup | Azure CNI |
| Encapsulation | Azure CNI (Geneve) |
| VNet Integration | Azure CNI |
| Packet Forwarding | **Cilium eBPF** |
| Service LB | **Cilium eBPF** |
| Network Policies | **Cilium eBPF** |
| Observability | **Cilium Hubble** |

**Think of it as:**
- Azure CNI = The architect (designs the network)
- Cilium = The builder (moves the packets efficiently)

---

## Quick Test

```bash
# Verify it's hybrid mode
kubectl -n kube-system exec ds/cilium -- cilium status

# Key indicators:
# ✅ IPAM: Azure              (Azure CNI manages IPs)
# ✅ Encapsulation: Geneve    (Azure CNI's overlay)
# ✅ KubeProxyReplacement: Strict  (Cilium handles services)
# ✅ Cilium: OK               (Cilium dataplane active)
```

**Conclusion:** You're using **Azure CNI Overlay** for the control plane and **Cilium eBPF** for the dataplane - the best of both worlds! 🎯
