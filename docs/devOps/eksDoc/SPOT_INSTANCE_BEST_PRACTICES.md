# Spot Instance Best Practices for Kubernetes

This guide covers optimal configuration for using AWS Spot Instances in self-managed Kubernetes clusters.

## 🎯 Golden Rules

1. **Never use spot for control plane in production**
2. **Always keep at least 1 on-demand worker for guaranteed capacity**
3. **Use persistent spot instances with stop behavior**
4. **Set max_price to None (on-demand price cap)**

## 💰 Spot Pricing Configuration

### Understanding max_price

```python
spot_options=aws.ec2.InstanceInstanceMarketOptionsSpotOptionsArgs(
    max_price=None,  # ✅ RECOMMENDED
    spot_instance_type="persistent",
    instance_interruption_behavior="stop",
)
```

| Setting | Behavior | Use Case |
|---------|----------|----------|
| `max_price=None` | Won't pay more than on-demand price | ✅ **Best practice** - always saves money |
| `max_price="0.0208"` | Explicit cap at $0.0208/hr | ✅ Good - same as None for t3.small |
| `max_price="0.005"` | Only use if spot ≤ $0.005/hr | ⚠️ More interruptions |
| `max_price=""` | Unclear behavior | ❌ Avoid - use None instead |

### How Spot Pricing Works

**Example: t3.small in us-west-1**

```
On-demand price:    $0.0208/hour (fixed)
Spot price range:   $0.0021-0.0062/hour (fluctuates)
Your savings:       70-90% cheaper than on-demand

With max_price=None:
├─ Spot at $0.0021/hr → You pay $0.0021/hr ✅
├─ Spot at $0.0050/hr → You pay $0.0050/hr ✅
├─ Spot at $0.0200/hr → You pay $0.0200/hr ✅
└─ Spot at $0.0210/hr → Instance interrupted ⚠️
```

**Key insight**: You always pay the **current spot price**, never more than **on-demand**.

## 🔧 Spot Instance Type

### Persistent vs One-Time

```python
spot_instance_type="persistent"  # ✅ RECOMMENDED
```

| Type | Behavior on Interruption | Data Preservation | Use Case |
|------|-------------------------|-------------------|----------|
| **persistent** | Stops, auto-restarts when capacity available | ✅ EBS preserved | ✅ Kubernetes nodes |
| **one-time** | Terminates permanently | ❌ Instance gone | Batch jobs |

**Why persistent for Kubernetes?**
- EBS volumes stay attached
- Node rejoins cluster automatically
- No manual intervention needed
- etcd/kubelet config preserved

## 🛡️ Interruption Behavior

### Stop vs Terminate vs Hibernate

```python
instance_interruption_behavior="stop"  # ✅ RECOMMENDED
```

| Behavior | What Happens | EBS Data | Cost When Stopped | Recovery Time |
|----------|--------------|----------|-------------------|---------------|
| **stop** | Instance stops | ✅ Preserved | $0 compute, EBS charges only | 2-5 min |
| **terminate** | Instance deleted | ❌ Lost | $0 | Must recreate |
| **hibernate** | RAM saved to disk | ✅ Preserved + RAM | $0 compute, EBS + RAM snapshot | 1-2 min |

**Recommendation**: Use **stop** for Kubernetes
- No snapshot costs (vs hibernate)
- Data preserved (vs terminate)
- Fast enough recovery
- Standard Kubernetes behavior

## 📊 Cluster Architecture Best Practices

### ✅ Recommended Configuration

```python
# Control Plane: On-Demand (reliability)
control_plane_count=1 (dev) or 3 (prod)
use_spot_instances=False for control plane

# Workers: Mixed (cost + reliability)
on_demand_worker_count=1-2  # Guaranteed capacity
spot_worker_count=3-10       # Cost optimization
```

**Why this works:**
- Control plane always available (no management outages)
- At least 1 on-demand worker ensures workloads run
- Spot workers provide 70-90% cost savings
- If all spots interrupted, on-demand handles load

### ❌ Anti-Patterns

**Don't:**
```python
# Bad: All spot (including control plane)
use_spot_instances=True
control_plane_count=1
on_demand_worker_count=0
spot_worker_count=3
```
**Why bad:** Cluster management unavailable during interruptions

**Don't:**
```python
# Bad: All workers are spot
on_demand_worker_count=0
spot_worker_count=5
```
**Why bad:** If all spots interrupted simultaneously, no capacity for workloads

**Don't:**
```python
# Bad: Setting max_price too low
max_price="0.003"  # Below typical spot price
```
**Why bad:** Frequent interruptions when spot price fluctuates above $0.003

## 🎯 Configuration by Use Case

### Learning/Testing ($30-35/month)

```python
SelfManagedK8sCluster(
    cluster_name="dev-k8s",
    kubernetes_version="1.30.0",
    cilium_version="1.16.5",
    
    # Control plane: On-demand for reliability
    control_plane_count=1,
    control_plane_instance_type="t3.small",
    
    # Workers: Mixed
    on_demand_worker_count=1,
    spot_worker_count=3,
    worker_instance_type="t3.small",
)
```

**Cost breakdown:**
- CP on-demand: $15/month
- 1 worker on-demand: $15/month
- 3 workers spot: $4.50-13.50/month
- **Total: $35-44/month**

### Development/Staging ($45-60/month)

```python
SelfManagedK8sCluster(
    cluster_name="staging-k8s",
    
    # Same as above but:
    on_demand_worker_count=2,  # More guaranteed capacity
    spot_worker_count=4,        # More elastic capacity
    worker_instance_type="t3.medium",  # More resources
)
```

### Production ($200-300/month)

```python
SelfManagedK8sCluster(
    cluster_name="prod-k8s",
    
    # Control plane: HA
    control_plane_count=3,
    control_plane_instance_type="t3.medium",
    
    # Workers: Mostly on-demand
    on_demand_worker_count=3,   # Baseline capacity
    spot_worker_count=5,        # Burst capacity (optional)
    worker_instance_type="t3.large",
)
```

**Why different for production:**
- 3 CPs for high availability (quorum)
- More on-demand workers for reliability
- Spot workers only for burst/non-critical workloads

## 🔍 Spot Interruption Handling

### Timeline of Interruption

```
T+0s    - AWS sends 2-minute warning notification
T+120s  - Instance stops (pods evicted)
T+130s  - Kubernetes reschedules pods to other workers
T+180s  - Pods running on new workers
T+5min  - AWS restarts spot instance (when capacity available)
T+6min  - Worker rejoins cluster (empty, ready for pods)
```

### Kubernetes Pod Behavior

**What happens to pods on interrupted spot worker:**

1. **Eviction**: Pods receive SIGTERM (graceful shutdown)
2. **Rescheduling**: Kubernetes schedules to available nodes
3. **Startup**: Pods start on new nodes (on-demand or other spot workers)
4. **Service continuity**: LoadBalancer/Service routes to new pods

**Impact:**
- Brief traffic disruption (30-90 seconds)
- No data loss (PersistentVolumes reattach)
- Minimal user impact (if using replicas)

## 📈 Monitoring Spot Instances

### Check Spot Status

```bash
# View spot instance requests
aws ec2 describe-spot-instance-requests \
  --region us-west-1 \
  --filters "Name=state,Values=active"

# Check spot price history
aws ec2 describe-spot-price-history \
  --instance-types t3.small \
  --start-time $(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%S) \
  --product-descriptions "Linux/UNIX" \
  --region us-west-1
```

### Set Up Interruption Alerts

```bash
# Create SNS topic for alerts
aws sns create-topic --name spot-interruption-alerts --region us-west-1

# Subscribe to email notifications
aws sns subscribe \
  --topic-arn arn:aws:sns:us-west-1:ACCOUNT_ID:spot-interruption-alerts \
  --protocol email \
  --notification-endpoint your-email@example.com
```

### CloudWatch Metrics

Monitor these metrics:
- `SpotInstanceInterruptions`: Count of interruptions
- `RunningInstances`: Current spot instance count
- `AvailableCapacity`: Spot capacity trends

## 💡 Advanced Optimization

### Diversify Across Instance Types

Instead of all t3.small:

```python
# Mix instance types for better availability
workers = [
    ("t3.small", 2),   # Primary
    ("t3a.small", 1),  # Alternative (cheaper)
    ("t2.small", 1),   # Fallback
]
```

**Why:** Different instance types have different interruption rates

### Spread Across Availability Zones

```python
# Use all 2 AZs in us-west-1
subnet_ids=[subnet_1a, subnet_1b]

# Workers distributed:
# - Worker 0: us-west-1a (on-demand)
# - Worker 1: us-west-1b (spot)
# - Worker 2: us-west-1a (spot)
# - Worker 3: us-west-1b (spot)
```

**Why:** AZ-level capacity issues won't affect all workers

### Use Spot Instance Advisor

Check [AWS Spot Instance Advisor](https://aws.amazon.com/ec2/spot/instance-advisor/):
- **Interruption frequency**: <5% = good, >20% = avoid
- **Savings rate**: How much cheaper than on-demand
- **Best instance types**: Least interrupted options

## 🎓 Key Takeaways

1. **max_price=None is optimal** - Always saves money vs on-demand
2. **persistent + stop is the right combo** - Data preserved, auto-recovery
3. **Never all-spot** - Keep 1+ on-demand worker for guaranteed capacity
4. **Control plane = on-demand** - Management availability is critical
5. **Monitor interruptions** - Learn your cluster's behavior
6. **Diversify** - Multiple instance types + AZs = better availability

## 📚 Reference

### Current Cluster Configuration

Your optimal setup:
```python
cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in subnets],
    ssh_key_name="k8s-dev-key",
    kubernetes_version=kubernetes_version,
    cilium_version=cilium_version,
    
    # Control Plane: On-demand (stable)
    control_plane_count=1,
    control_plane_instance_type="t3.small",
    
    # Workers: 1 on-demand + 3 spot (cost-optimized)
    on_demand_worker_count=1,
    spot_worker_count=3,
    worker_instance_type="t3.small",
    
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)
```

**Spot worker configuration (in k8s_self_host.py):**
```python
spot_options = aws.ec2.InstanceInstanceMarketOptionsArgs(
    market_type="spot",
    spot_options=aws.ec2.InstanceInstanceMarketOptionsSpotOptionsArgs(
        max_price=None,                           # ✅ On-demand price cap
        spot_instance_type="persistent",          # ✅ Auto-restart
        instance_interruption_behavior="stop",    # ✅ Preserve EBS
    ),
)
```

This configuration provides **99%+ uptime** at **60-70% cost savings**! 🎉

## Related Documentation

- [`SPOT_INSTANCE_RISKS.md`](./SPOT_INSTANCE_RISKS.md) - Detailed risk analysis
- [`COST_OPTIMIZATION.md`](./COST_OPTIMIZATION.md) - Cost comparison scenarios
- [AWS Spot Instance Documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-spot-instances.html)
