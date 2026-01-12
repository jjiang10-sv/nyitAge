# Spot Instance Risks and Recommendations

## Current Configuration Risk Level

**Your setup:**
- Control Plane: 1× spot (⚠️ HIGH RISK)
- Workers: 1× on-demand + 3× spot (✅ SAFE)

## The Traps

### 1. **Control Plane Interruption = Total Outage**

When a spot control plane is interrupted:
- ❌ kubectl stops working
- ❌ Can't deploy new apps
- ❌ Can't scale existing apps
- ❌ API server unavailable
- ❌ Cluster management disabled
- ✅ **Existing pods keep running** (but degraded)

**Recovery**: 5-30 minutes (when AWS restarts spot instance)

### 2. **No Auto-Recovery During Interruption**

Unlike managed services (EKS), you're responsible for:
- Monitoring spot interruptions
- Ensuring etcd data integrity
- Manually recovering if needed

### 3. **Persistent vs Terminate**

Our config uses `persistent` spot instances:
```python
spot_instance_type="persistent"
instance_interruption_behavior="stop"
```

**Good news**: Instance stops (not terminates), EBS volumes preserved  
**Bad news**: Still unavailable during interruption

## Recommendations by Use Case

### For Learning/Testing ✅ (Your Current Use Case)

**Current config is ACCEPTABLE**:
- Interruptions are rare enough for learning
- You can tolerate 5-30 min outages
- Cost savings justify the risk

**To deploy current config:**
```bash
pulumi up
```

### For Development/Staging ⚠️

**Upgrade to:**
```python
use_spot_instances=False,      # On-demand control plane
control_plane_count=1,
on_demand_worker_count=1,
spot_worker_count=3,
```

**Why**: Dev teams need reliable cluster access

### For Production ✅

**Minimum requirements:**
```python
use_spot_instances=False,      # Never spot for control plane
control_plane_count=3,         # High availability
on_demand_worker_count=2-3,    # Guaranteed capacity
spot_worker_count=0-10,        # Optional cost savings
```

## Severity of Interruptions

### Control Plane Spot Interruption

**Severity: 🔴 CRITICAL**

What breaks:
```bash
❌ kubectl get pods
❌ kubectl apply -f deployment.yaml
❌ kubectl scale deployment/app --replicas=5
❌ kubectl logs pod/app
❌ kubectl exec -it pod/app -- bash
```

What still works:
```bash
✅ Existing pods keep running
✅ Services still route traffic
✅ Persistent volumes still attached
✅ Node-to-node communication works
```

**Business Impact**: Can't respond to incidents, can't deploy fixes

### Worker Spot Interruption (with on-demand backup)

**Severity: 🟡 MINOR**

What happens:
```bash
⚠️ Pods on spot worker get evicted
✅ Kubernetes reschedules to on-demand worker
✅ Brief traffic disruption (30-60s)
✅ No data loss (PVs reattach)
```

**Business Impact**: Minimal - some 503 errors during rescheduling

## Real-World Scenarios

### Scenario 1: Single Spot CP Interrupted

```
Timeline:
00:00 - Spot interruption warning (2-minute notice)
00:02 - Control plane stops
00:02 - API server unreachable
00:05 - Pods still running, but no management
00:25 - AWS restarts spot instance
00:27 - Control plane back online
00:27 - kubectl works again

Impact: 25 minutes of cluster management outage
Cost of incident: $0 (learning/testing)
```

### Scenario 2: Spot Worker + On-Demand Backup

```
Timeline:
00:00 - Spot worker-2 interrupted
00:00 - Pods on worker-2 evicted
00:01 - Kubernetes schedules pods to worker-ondemand-0
00:01 - Pods starting on new node
00:02 - Pods running on on-demand worker
00:05 - Spot worker-2 restarts
00:06 - Worker-2 rejoins cluster (empty)

Impact: 1-2 minutes brief pod disruption
Cost of incident: Negligible
```

## Best Practice: Separate Control Plane Spot Decision

I recommend updating the code to allow independent control:

```python
cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in subnets],
    ssh_key_name="k8s-dev-key",
    
    # Separate control
    control_plane_spot=True,       # NEW: Control plane spot decision
    control_plane_count=1,
    
    # Workers always mixed
    on_demand_worker_count=1,
    spot_worker_count=3,
    ...
)
```

This would let you:
- Dev: CP spot=True (save money)
- Prod: CP spot=False (reliability)
- Workers: Always mixed for best of both worlds

## Decision Matrix

| Workload | CP Spot? | CP Count | On-Demand Workers | Spot Workers | Monthly Cost |
|----------|----------|----------|-------------------|--------------|--------------|
| **Learning** | ✅ Yes | 1 | 1 | 3 | ~$30 |
| **Testing** | ⚠️ Maybe | 1 | 1 | 3 | ~$35-45 |
| **Development** | ❌ No | 1 | 1 | 3-5 | ~$45-60 |
| **Staging** | ❌ No | 1 | 2 | 3-5 | ~$60-80 |
| **Production** | ❌ **NEVER** | 3 | 3 | 5-10 | ~$200-300 |

## My Recommendation for You

**Since you're learning:**

Keep current config ✅
```python
use_spot_instances=True           # Accept risk for learning
control_plane_count=1
on_demand_worker_count=1
spot_worker_count=3
```

**Benefits:**
- ~$30/month cost
- Learn about spot instance behavior
- Practice handling interruptions
- Understand Kubernetes resilience

**Risks to accept:**
- ~1-2 outages per month
- 5-30 min recovery time per outage
- Can't manage cluster during outage

**When to upgrade:**
- Moving from learning → development: Switch CP to on-demand
- Need >99% uptime: Switch to 3× on-demand CPs
- Running production: Never use spot for CP

Want me to create separate control plane spot configuration?
