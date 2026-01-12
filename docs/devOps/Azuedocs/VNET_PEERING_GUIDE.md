# VNet Peering and Pod Communication Guide

## Quick Answer

**Without VNet Peering:**
- ❌ Pods in different regions **CANNOT** ping each other
- ❌ Node VMs **CANNOT** communicate across regions
- ✅ Users access apps via Azure Front Door

**With VNet Peering:**
- ✅ Node VMs **CAN** communicate across regions
- ⚠️ **Pods still CANNOT ping directly** (requires additional config)
- ✅ Enables shared services (databases, caches)

**Why the limitation?** We use **Azure CNI Overlay** mode, where pod IPs are in an overlay network, not directly routed in the VNet.

---

## Understanding the Network Layers

### Without VNet Peering (Default)

```
┌─────────────────────┐    ┌─────────────────────┐
│ Canada Central      │    │ East US             │
│                     │    │                     │
│ VNet: 10.0.0.0/14   │    │ VNet: 10.4.0.0/14   │
│                     │    │                     │
│ ┌─────────────────┐ │    │ ┌─────────────────┐ │
│ │ Node: 10.0.0.5  │ │    │ │ Node: 10.4.0.5  │ │
│ │                 │ │    │ │                 │ │
│ │ Pod: 10.32.1.10 │ │    │ │ Pod: 10.40.1.10 │ │
│ │   (overlay)     │ │    │ │   (overlay)     │ │
│ └─────────────────┘ │    │ └─────────────────┘ │
└─────────────────────┘    └─────────────────────┘
         ❌                          ❌
    No connectivity          No connectivity
```

**Result:**
- Nodes: 10.0.0.5 ❌ Cannot reach → 10.4.0.5
- Pods: 10.32.1.10 ❌ Cannot reach → 10.40.1.10

---

### With VNet Peering (Basic)

```
┌─────────────────────┐    ┌─────────────────────┐
│ Canada Central      │    │ East US             │
│                     │    │                     │
│ VNet: 10.0.0.0/14   │════│ VNet: 10.4.0.0/14   │
│                     │    │                     │
│ ┌─────────────────┐ │    │ ┌─────────────────┐ │
│ │ Node: 10.0.0.5  │─┼────┼─│ Node: 10.4.0.5  │ │
│ │      ✅         │ │    │ │      ✅         │ │
│ │ Pod: 10.32.1.10 │ │    │ │ Pod: 10.40.1.10 │ │
│ │      ❌         │ │    │ │      ❌         │ │
│ └─────────────────┘ │    │ └─────────────────┘ │
└─────────────────────┘    └─────────────────────┘
```

**Result:**
- Nodes: 10.0.0.5 ✅ Can reach → 10.4.0.5
- Pods: 10.32.1.10 ❌ Cannot reach → 10.40.1.10 (overlay not routed)

---

## Why Pods Can't Ping (By Default)

### The Issue: Overlay Networking

Our platform uses **Azure CNI Overlay** mode:

```python
network_profile=ContainerServiceNetworkProfileArgs(
    network_plugin="azure",
    network_plugin_mode="overlay",  # ← This is key!
    network_dataplane="cilium",
    pod_cidr="10.32.0.0/13",        # Overlay network
)
```

**What This Means:**

1. **Node IPs** (10.0.0.x) are in the VNet → **VNet peering works**
2. **Pod IPs** (10.32.x.x) are in an overlay → **VNet peering doesn't help**

### The Packet Flow

#### Node-to-Node (With VNet Peering) ✅

```
Canada Node (10.0.0.5)
  │
  │ Packet: src=10.0.0.5, dst=10.4.0.5
  ↓
VNet Peering
  ↓
East US Node (10.4.0.5) ✅ Receives packet
```

#### Pod-to-Pod (With VNet Peering Only) ❌

```
Canada Pod (10.32.1.10)
  │
  │ Packet: src=10.32.1.10, dst=10.40.1.10
  ↓
Canada Node (10.0.0.5)
  │
  │ "Where is 10.40.1.10?"
  │ VNet routing: No route! (overlay IP)
  ❌ Packet dropped
```

The VNet peering only knows about **VNet CIDRs** (10.0.0.0/14, 10.4.0.0/14), not **overlay pod CIDRs** (10.32.0.0/13, 10.40.0.0/13).

---

## How to Enable Pod-to-Pod Communication

### Option 1: Cilium Cluster Mesh (Recommended)

**What It Is:** Cilium's built-in multi-cluster connectivity

```bash
# Install Cilium cluster mesh
cilium clustermesh enable --context canada
cilium clustermesh enable --context eastus

# Connect clusters
cilium clustermesh connect --context canada --destination-context eastus
```

**How It Works:**
- Creates tunnels between clusters
- Pods get DNS resolution across clusters
- Service discovery across regions
- Load balancing across clusters

**Result:**
```
Canada Pod (10.32.1.10) ⟷ Cilium Mesh ⟷ East US Pod (10.40.1.10) ✅
```

**VNet Peering Requirement:** ✅ Yes (for control plane communication)

---

### Option 2: Service Mesh (Istio/Linkerd)

**Setup:**
```bash
# Install Istio with multi-cluster
istioctl install --set values.global.meshID=global-mesh

# Join clusters
istioctl create-remote-secret --context=canada | \
  kubectl apply -f - --context=eastus
```

**Result:**
- mTLS between pods across regions
- Service discovery
- Traffic routing

**VNet Peering Requirement:** ✅ Yes

---

### Option 3: Azure Virtual WAN (Enterprise)

**What It Is:** Hub-and-spoke global transit network

```
        ┌──────────────┐
        │ Virtual WAN  │
        │     Hub      │
        └──────┬───────┘
               │
    ┏━━━━━━━━━┻━━━━━━━━━┓
    ↓          ↓         ↓
 Canada     East US   W Europe
```

**Features:**
- Full routing between all regions
- Can route pod CIDRs (with UDRs)
- Global transit for all Azure resources

**Cost:** $$$ (more expensive than VNet peering)

---

### Option 4: Public Services (Simple)

**Don't need pod-to-pod?** Use services!

```yaml
# Expose service in East US
apiVersion: v1
kind: Service
metadata:
  name: shared-api
  namespace: eastus
spec:
  type: LoadBalancer
  selector:
    app: api
  ports:
  - port: 443
```

```yaml
# Access from Canada pods
apiVersion: v1
kind: Pod
metadata:
  name: client
  namespace: canada
spec:
  containers:
  - name: app
    env:
    - name: API_URL
      value: "https://<eastus-lb-ip>"
```

**Result:**
- Canada pods → Internet → East US LoadBalancer → East US pods ✅
- No VNet peering needed
- Traffic goes through Front Door or public LB

---

## VNet Peering Use Cases

### ✅ When You SHOULD Enable VNet Peering

#### 1. **Shared Databases**

```
┌────────────┐    VNet Peering    ┌────────────┐
│ Canada AKS │════════════════════│ East US    │
│   Pods     │                    │ PostgreSQL │
│            │────────────────────▶│ Private IP │
└────────────┘    10.4.5.10       └────────────┘
```

Pods in Canada can access private database in East US.

#### 2. **Shared Services (Redis, Kafka)**

```
All Regions ═══▶ Central Region
                 ├─ Redis Cluster
                 ├─ Kafka
                 └─ Shared Storage
```

#### 3. **Disaster Recovery**

```
Canada (Primary)    East US (DR)
    │                   │
    │ Replicate data    │
    └──────────────────▶│
       VNet Peering
```

Continuous data replication between regions.

#### 4. **Hybrid Cloud**

```
AKS ════ VNet Peering ════ On-Premises
                           ├─ AD
                           ├─ Legacy Apps
                           └─ Databases
```

Connect to on-prem resources.

---

### ❌ When You DON'T Need VNet Peering

#### 1. **Stateless Apps Only**

If all apps are stateless and independent:
```
Front Door → Region 1 (complete stack)
          → Region 2 (complete stack)
          → Region 3 (complete stack)
```

Each region is self-contained.

#### 2. **Using Public Services**

```
App in Canada → Cosmos DB (global)
App in East US → Cosmos DB (same global instance)
```

No peering needed for globally replicated services.

#### 3. **Front Door Routing Only**

```
Users → Front Door → Nearest Region → Complete App
```

If regions don't need to talk to each other.

---

## Real-World Examples

### Example 1: E-Commerce (No Peering Needed)

```
Architecture:
├─ Canada: Frontend + Backend + Cosmos DB
├─ East US: Frontend + Backend + Cosmos DB
└─ Europe: Frontend + Backend + Cosmos DB

Cosmos DB: Global replication
Front Door: Routes users to nearest region
```

**No VNet peering needed** - each region is complete.

---

### Example 2: Shared Database (Peering Required)

```
Architecture:
├─ Canada: Application Pods
├─ East US: Application Pods
└─ Shared: PostgreSQL (10.100.0.0/16)
    └─ Accessed from all regions

VNet Peering: Required for private DB access
```

---

### Example 3: Microservices Mesh (Peering + Cluster Mesh)

```
Architecture:
├─ Canada: Frontend + Auth Service
├─ East US: API Gateway + User Service
└─ Europe: Analytics Service

VNet Peering: ✅ Enabled
Cilium Mesh: ✅ Enabled

Result: Services can call each other across regions
```

---

## Testing Pod Communication

### Test 1: Without VNet Peering

```bash
# Deploy test pods in both regions
kubectl --context canada run test-can --image=busybox -- sleep 3600
kubectl --context eastus run test-eus --image=busybox -- sleep 3600

# Get pod IPs
CAN_POD_IP=$(kubectl --context canada get pod test-can -o jsonpath='{.status.podIP}')
EUS_POD_IP=$(kubectl --context eastus get pod test-eus -o jsonpath='{.status.podIP}')

# Try to ping (will fail)
kubectl --context canada exec test-can -- ping -c 3 $EUS_POD_IP
# ❌ Network unreachable or timeout
```

---

### Test 2: With VNet Peering (Node-to-Node)

```bash
# Enable VNet peering first
pulumi config set enable_vnet_peering true
pulumi up

# Get node IPs (not pod IPs!)
CAN_NODE_IP=$(kubectl --context canada get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')
EUS_NODE_IP=$(kubectl --context eastus get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="InternalIP")].address}')

# Ping node from pod (will work!)
kubectl --context canada exec test-can -- ping -c 3 $EUS_NODE_IP
# ✅ 64 bytes from 10.4.0.5: icmp_seq=1 ttl=64 time=15 ms
```

Nodes can communicate, but pods still can't (without Cilium Cluster Mesh).

---

### Test 3: With Cilium Cluster Mesh

```bash
# Enable cluster mesh
cilium clustermesh enable --context canada
cilium clustermesh enable --context eastus
cilium clustermesh connect --context canada --destination-context eastus

# Now pods CAN communicate
kubectl --context canada exec test-can -- ping -c 3 $EUS_POD_IP
# ✅ 64 bytes from 10.40.1.10: icmp_seq=1 ttl=64 time=25 ms
```

---

## Cost Comparison

### No VNet Peering: ~$8,000/month
- 3 regions × $2,800/month
- Azure Front Door: $35/month

### With VNet Peering: ~$8,100/month
- Same as above
- VNet Peering: ~$0.01/GB ingress + $0.01/GB egress
- Example: 10 TB/month = ~$100/month

### With Virtual WAN: ~$9,500/month
- Virtual WAN Hub: ~$0.25/hour × 3 = $540/month
- Data transfer: Higher rates
- More complex but more features

---

## Decision Tree

```
Do regions need to share resources?
│
├─ No
│  └─ ❌ Don't enable VNet peering
│     Each region is isolated
│     Use Front Door for user routing
│
└─ Yes
   │
   ├─ Shared databases/services?
   │  └─ ✅ Enable VNet peering
   │     Pods access via node IPs
   │
   └─ Pod-to-pod communication?
      └─ ✅ Enable VNet peering + Cilium Cluster Mesh
         Full service mesh across regions
```

---

## Updated Example Usage

### Enable Pod-to-Pod (Add to example_usage.py)

```python
# After VNet peering is created, configure Cilium Cluster Mesh

if deploy_multi_region and enable_vnet_peering:
    # VNet peering is created (existing code)
    
    # TODO: Add Cilium Cluster Mesh configuration
    # This would require:
    # 1. Expose Cilium cluster mesh API service
    # 2. Create shared secrets
    # 3. Configure cluster mesh connections
    
    # For now, this is done manually:
    pulumi.export("cilium_mesh_setup", 
        "Run: cilium clustermesh enable && cilium clustermesh connect")
```

---

## Summary

### VNet Peering Alone

| Communication Type | Without Peering | With Peering |
|-------------------|-----------------|--------------|
| **Node ↔ Node** | ❌ | ✅ |
| **Pod ↔ Node** | ❌ | ✅ (same region only) |
| **Pod ↔ Pod** | ❌ | ❌ (overlay issue) |
| **Pod ↔ DB** | ❌ | ✅ (if DB in VNet) |

### VNet Peering + Cilium Cluster Mesh

| Communication Type | Result |
|-------------------|--------|
| **Node ↔ Node** | ✅ |
| **Pod ↔ Node** | ✅ |
| **Pod ↔ Pod** | ✅ (via mesh) |
| **Pod ↔ Service** | ✅ (cross-cluster) |

---

## Recommendation

**For most production scenarios:**

1. **Start without VNet peering**
   - Simplest architecture
   - Lowest cost
   - Each region independent

2. **Add VNet peering when needed for:**
   - Shared databases
   - Central services
   - Disaster recovery

3. **Add Cilium Cluster Mesh only if:**
   - Microservices span regions
   - Need service discovery across regions
   - Complex multi-region workflows

**Our current implementation** enables VNet peering optionally, which is perfect for giving you the **flexibility** without forcing unnecessary complexity.

---

## Quick Commands

```bash
# No peering (default)
pulumi up

# With VNet peering (nodes can talk)
pulumi config set enable_vnet_peering true
pulumi up

# With pod-to-pod (manual setup after deployment)
cilium clustermesh enable --context canada
cilium clustermesh enable --context eastus
cilium clustermesh connect --context canada --destination-context eastus
```

**Bottom line:** VNet peering connects the VNets (nodes can talk), but pods need additional configuration (Cilium Cluster Mesh) to communicate across regions due to overlay networking. 🎯
