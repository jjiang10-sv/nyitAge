# Kubernetes Cluster Best Practices

A comprehensive guide for setting up production-grade self-managed Kubernetes clusters.

## 📚 Table of Contents

1. [Control Plane Best Practices](#control-plane-best-practices)
2. [Worker Node Configuration](#worker-node-configuration)
3. [High Availability](#high-availability)
4. [Networking and Security](#networking-and-security)
5. [Storage Configuration](#storage-configuration)
6. [Resource Management](#resource-management)
7. [Monitoring and Observability](#monitoring-and-observability)
8. [Backup and Disaster Recovery](#backup-and-disaster-recovery)
9. [Security Hardening](#security-hardening)
10. [Operational Excellence](#operational-excellence)

---

## Control Plane Best Practices

### 1. Taint Control Plane Nodes

**Rule**: Control plane nodes should NOT run user workloads.

```bash
# Automatically applied in our setup
kubectl taint nodes --all node-role.kubernetes.io/control-plane:NoSchedule
```

**Why:**
- ✅ Isolates control plane from user workload interference
- ✅ Protects etcd, API server, scheduler from resource contention
- ✅ Prevents user pods from crashing control plane
- ✅ Industry standard for production clusters

**What can still run on control plane:**
- System DaemonSets (Cilium, kube-proxy)
- Monitoring agents (node-exporter, fluent-bit)
- CNI plugins (required tolerations)

### 2. Use Odd Number of Control Planes

**Rule**: Always use 1, 3, or 5 control planes (never 2, 4, 6).

| Count | Use Case | Fault Tolerance | Cost |
|-------|----------|----------------|------|
| **1** | Dev/Testing | None (0) | Low |
| **3** | Production ✅ | Survives 1 failure | Medium |
| **5** | Mission-critical | Survives 2 failures | High |
| ❌ 2 | Never use | Split-brain risk | - |
| ❌ 4 | Never use | Wastes resources | - |

**Why odd numbers:**
- etcd requires quorum: (n/2)+1 nodes
- 3 CPs = 2 nodes needed for quorum (survives 1 failure)
- 2 CPs = 2 nodes needed for quorum (survives 0 failures - waste!)

### 3. Separate etcd (Optional)

**For very large clusters (>100 nodes):**

```
Option 1: Co-located etcd (our setup)
├─ Control Plane 1: API + Scheduler + Controller + etcd
├─ Control Plane 2: API + Scheduler + Controller + etcd
└─ Control Plane 3: API + Scheduler + Controller + etcd

Option 2: Separate etcd (advanced)
├─ etcd 1, etcd 2, etcd 3  (dedicated)
└─ CP 1, CP 2, CP 3 (API/Scheduler/Controller only)
```

**When to separate:**
- 100+ nodes
- High write traffic to etcd
- Need independent scaling
- Compliance requirements

### 4. Control Plane Resources

**Minimum requirements:**

| Cluster Size | Instance Type | vCPU | RAM |
|-------------|---------------|------|-----|
| < 10 nodes | t3.small | 2 | 2 GB |
| 10-50 nodes | t3.medium ✅ | 2 | 4 GB |
| 50-100 nodes | t3.large | 2 | 8 GB |
| 100+ nodes | m5.xlarge | 4 | 16 GB |

---

## Worker Node Configuration

### 1. Node Labels

**Apply descriptive labels:**

```bash
# Environment
kubectl label nodes worker-1 environment=production

# Workload type
kubectl label nodes worker-1 workload=compute-intensive
kubectl label nodes worker-2 workload=memory-intensive

# Availability zone
kubectl label nodes worker-1 topology.kubernetes.io/zone=us-west-2a

# Instance type
kubectl label nodes worker-1 node.kubernetes.io/instance-type=t3.large
```

**Use in pod scheduling:**

```yaml
nodeSelector:
  workload: compute-intensive
  environment: production
```

### 2. Resource Reservations

**Reserve resources for system daemons:**

```yaml
# kubelet config
systemReserved:
  cpu: 100m
  memory: 256Mi
  ephemeral-storage: 1Gi
kubeReserved:
  cpu: 100m
  memory: 256Mi
  ephemeral-storage: 1Gi
```

**Why:**
- Prevents pods from consuming all node resources
- Ensures kubelet/OS have resources
- Prevents node instability

### 3. Pod Density Limits

**Set maximum pods per node:**

```bash
# In kubelet config
maxPods: 110  # Default, adjust based on instance size
```

**Guidelines:**

| Instance Type | Recommended Max Pods |
|---------------|---------------------|
| t3.small | 20-30 |
| t3.medium | 40-50 |
| t3.large | 70-90 |
| m5.xlarge | 110 (default) |

---

## High Availability

### 1. Multi-AZ Deployment

**Spread across availability zones:**

```python
# Ensure workers spread across AZs
subnet_ids = [
    subnet_us_west_2a,
    subnet_us_west_2b,
    subnet_us_west_2c,
]

# Workers automatically distributed
```

**Anti-affinity for critical workloads:**

```yaml
affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchLabels:
          app: critical-app
      topologyKey: topology.kubernetes.io/zone
```

### 2. Application Replicas

**Never run single replica in production:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp
spec:
  replicas: 3  # ✅ Minimum for HA
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
```

### 3. Pod Disruption Budgets

**Protect against voluntary disruptions:**

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: webapp-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: webapp
```

**Why:**
- Prevents draining too many pods during node maintenance
- Ensures minimum replica count during updates
- Required for production workloads

---

## Networking and Security

### 1. Network Policies (Cilium)

**Implement zero-trust networking:**

```yaml
# Default deny all ingress
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: default-deny
spec:
  endpointSelector: {}
  ingress:
  - {}  # Deny all
---
# Allow specific traffic
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: allow-frontend-to-backend
spec:
  endpointSelector:
    matchLabels:
      app: backend
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "8080"
        protocol: TCP
```

### 2. Service Mesh (Optional)

**Our setup includes SPIRE for mTLS:**

```bash
# Verify mTLS is working
cilium status
# Should show: Mutual Authentication: SPIRE Enabled

# Check SPIRE agents
kubectl get pods -n spire
```

**Benefits:**
- Automatic pod-to-pod encryption
- Workload identity (SPIFFE)
- Zero-trust communication

### 3. Security Groups

**Defense in depth:**

```python
# Control Plane Security Group
- Port 6443: API server (from workers + admins)
- Port 2379-2380: etcd (from control planes only)
- Port 10250: kubelet (from control planes)

# Worker Security Group
- Port 10250: kubelet (from control planes)
- Port 30000-32767: NodePort services (from LB)
- All ports: from other workers (pod communication)
```

### 4. Private API Server (Production)

**Use internal load balancer for API:**

```
Public Internet
     ↓
VPN/Bastion Host
     ↓
Internal LB → Control Plane 1, 2, 3
```

**Not exposed to public internet directly.**

---

## Storage Configuration

### 1. StorageClass Best Practices

**Our setup (already implemented):**

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ebs-sc
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ebs.csi.aws.com
parameters:
  type: gp3  # ✅ Latest generation
  encrypted: "true"  # ✅ Always encrypt
  iops: "3000"
  throughput: "125"
volumeBindingMode: WaitForFirstConsumer  # ✅ Topology-aware
allowVolumeExpansion: true  # ✅ Allow growth
reclaimPolicy: Delete  # For dev; use Retain for prod data
```

### 2. Backup Strategy

**For stateful workloads:**

1. **Snapshot-based:**
   ```bash
   # Create VolumeSnapshot
   kubectl create -f snapshot.yaml
   ```

2. **Application-level:**
   ```bash
   # PostgreSQL example
   pg_dump database > backup.sql
   ```

3. **Velero for cluster backup:**
   ```bash
   velero backup create cluster-backup --include-namespaces production
   ```

### 3. Volume Types

| Type | Use Case | IOPS | Cost |
|------|----------|------|------|
| **gp3** | General purpose ✅ | 3,000-16,000 | Medium |
| gp2 | Legacy | Up to 3,000 | Medium |
| io2 | High performance DB | 100,000+ | High |
| st1 | Cold storage | Low | Low |

---

## Resource Management

### 1. Resource Requests and Limits

**Always set both:**

```yaml
resources:
  requests:  # Scheduler uses this
    cpu: 100m
    memory: 128Mi
  limits:  # OOM killer uses this
    cpu: 200m
    memory: 256Mi
```

**Guidelines:**
- Requests: Realistic average usage
- Limits: Prevent runaway processes
- CPU is compressible (throttled)
- Memory is not (OOM kill)

### 2. QoS Classes

**Kubernetes assigns QoS based on resources:**

| Class | Condition | Eviction Priority |
|-------|-----------|------------------|
| **Guaranteed** | requests = limits | Low (last) |
| **Burstable** | requests < limits | Medium |
| **BestEffort** | No requests/limits | High (first) |

**Production pods should be Guaranteed or Burstable.**

### 3. LimitRanges

**Set defaults per namespace:**

```yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: default-limits
  namespace: production
spec:
  limits:
  - default:  # Default limits
      cpu: 500m
      memory: 512Mi
    defaultRequest:  # Default requests
      cpu: 100m
      memory: 128Mi
    type: Container
```

### 4. ResourceQuotas

**Prevent namespace resource hogging:**

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: production-quota
  namespace: production
spec:
  hard:
    requests.cpu: "20"
    requests.memory: "40Gi"
    persistentvolumeclaims: "10"
    pods: "50"
```

---

## Monitoring and Observability

### 1. Essential Metrics

**Control Plane:**
- API server latency
- etcd leader changes
- Scheduler queue depth
- Controller loop duration

**Nodes:**
- CPU/Memory/Disk usage
- Network throughput
- Container restarts
- OOM kills

**Pods:**
- Restart count
- Resource usage vs requests
- Health check failures

### 2. Monitoring Stack

**Recommended setup:**

```
Prometheus → Metrics collection
Grafana → Visualization
AlertManager → Alerting
Loki → Log aggregation
```

**Quick start:**

```bash
# Install kube-prometheus-stack
helm install prometheus prometheus-community/kube-prometheus-stack
```

### 3. Logging Best Practices

**Structured logging:**

```json
{
  "timestamp": "2026-01-01T00:00:00Z",
  "level": "ERROR",
  "service": "api-server",
  "message": "Failed to process request",
  "request_id": "abc123",
  "user_id": "user456"
}
```

**Log aggregation:**
- Fluent Bit → lightweight forwarder
- Loki → log storage
- Grafana → log queries

### 4. Distributed Tracing

**For microservices:**

```
Jaeger or Zipkin
├─ Trace requests across services
├─ Identify bottlenecks
└─ Debug latency issues
```

---

## Backup and Disaster Recovery

### 1. etcd Backup

**Critical: Backup etcd regularly!**

```bash
# Automated backup script
#!/bin/bash
ETCDCTL_API=3 etcdctl snapshot save \
  /backup/etcd-$(date +%Y%m%d-%H%M%S).db \
  --endpoints=https://127.0.0.1:2379 \
  --cacert=/etc/kubernetes/pki/etcd/ca.crt \
  --cert=/etc/kubernetes/pki/etcd/server.crt \
  --key=/etc/kubernetes/pki/etcd/server.key

# Verify backup
ETCDCTL_API=3 etcdctl snapshot status /backup/etcd-*.db
```

**Schedule:**
- Production: Every 6 hours
- Development: Daily

### 2. Disaster Recovery Plan

**Test recovery procedure:**

1. **Restore etcd:**
   ```bash
   etcdctl snapshot restore snapshot.db
   ```

2. **Reinitialize cluster:**
   ```bash
   kubeadm init --config=etcd-restore.yaml
   ```

3. **Rejoin workers:**
   ```bash
   kubeadm join --token=... --discovery-token-ca-cert-hash=...
   ```

4. **Verify:**
   ```bash
   kubectl get nodes
   kubectl get pods -A
   ```

**RTO/RPO targets:**
- RTO (Recovery Time): < 1 hour
- RPO (Data Loss): < 6 hours

### 3. Application Backup

**Use Velero:**

```bash
# Full cluster backup
velero backup create full-backup

# Namespace backup
velero backup create app-backup --include-namespaces production

# Scheduled backups
velero schedule create daily-backup --schedule="0 2 * * *"
```

---

## Security Hardening

### 1. RBAC (Role-Based Access Control)

**Principle of least privilege:**

```yaml
# ServiceAccount for app
apiVersion: v1
kind: ServiceAccount
metadata:
  name: app-sa
  namespace: production
---
# Role with minimal permissions
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: app-role
  namespace: production
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["get", "list"]
---
# Bind role to ServiceAccount
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: app-rolebinding
  namespace: production
subjects:
- kind: ServiceAccount
  name: app-sa
roleRef:
  kind: Role
  name: app-role
  apiGroup: rbac.authorization.k8s.io
```

### 2. Pod Security Standards

**Enforce security policies:**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: production
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

**Restricted policy prevents:**
- Running as root
- Privileged containers
- Host network/PID/IPC
- Unsafe sysctls

### 3. Secrets Management

**Never store secrets in Git:**

```bash
# Use external secret management
# - AWS Secrets Manager
# - HashiCorp Vault
# - Sealed Secrets

# Example: External Secrets Operator
kubectl apply -f https://raw.githubusercontent.com/external-secrets/external-secrets/main/deploy/crds/bundle.yaml
```

### 4. Image Security

**Scan images for vulnerabilities:**

```yaml
# Admission controller
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingWebhookConfiguration
metadata:
  name: image-scanner
webhooks:
- name: scan.images.io
  # Blocks deployment of vulnerable images
```

**Best practices:**
- Use specific image tags (not `latest`)
- Scan images in CI/CD
- Use private registries
- Sign images (cosign)

### 5. Network Policies

**Implement microsegmentation:**

```yaml
# Allow only necessary traffic
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: api-allow-frontend
spec:
  podSelector:
    matchLabels:
      app: api
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: frontend
    ports:
    - protocol: TCP
      port: 8080
```

---

## Operational Excellence

### 1. GitOps Deployment

**Use declarative configs:**

```
Git Repository (Source of Truth)
    ↓
ArgoCD/FluxCD (Continuous Deployment)
    ↓
Kubernetes Cluster (Live State)
```

**Benefits:**
- Version control for cluster state
- Rollback capability
- Audit trail
- Automated deployments

### 2. Rolling Updates

**Zero-downtime deployments:**

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1  # Max extra pods during update
      maxUnavailable: 0  # Always maintain replicas
```

### 3. Health Checks

**Liveness and Readiness probes:**

```yaml
livenessProbe:  # Restart if unhealthy
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:  # Remove from service if not ready
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

### 4. Graceful Shutdown

**Handle SIGTERM properly:**

```yaml
lifecycle:
  preStop:
    exec:
      command: ["/bin/sh", "-c", "sleep 15"]
terminationGracePeriodSeconds: 30
```

**Why:**
- Allows connections to drain
- Prevents connection errors
- Clean shutdown

### 5. Autoscaling

**Horizontal Pod Autoscaler:**

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: webapp-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: webapp
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

**Cluster Autoscaler (Cloud):**
- Automatically adds nodes when needed
- Removes underutilized nodes
- Integrated with cloud providers

---

## Our Implementation

### ✅ What We've Implemented

| Best Practice | Status | Details |
|--------------|--------|---------|
| **Control Plane Taint** | ✅ | Automatic in init script |
| **HA Control Plane** | ✅ | Configurable (1 or 3 CPs) |
| **Cilium CNI** | ✅ | Native routing, no tunnels |
| **mTLS with SPIRE** | ✅ | Zero-trust network |
| **Storage CSI** | ✅ | AWS EBS with encryption |
| **Default StorageClass** | ✅ | gp3, encrypted, expandable |
| **Security Groups** | ✅ | Least-privilege rules |
| **Multi-AZ Support** | ✅ | Worker distribution |
| **Instance Tagging** | ✅ | Role and cluster tags |
| **Spot Instance Support** | ✅ | Cost optimization |

### 🔄 Recommended Additions

**For production, consider adding:**

1. **Monitoring Stack**
   ```bash
   helm install prometheus prometheus-community/kube-prometheus-stack
   ```

2. **Backup Solution**
   ```bash
   helm install velero vmware-tanzu/velero --set configuration.provider=aws
   ```

3. **GitOps**
   ```bash
   kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
   ```

4. **External Secrets**
   ```bash
   helm install external-secrets external-secrets/external-secrets
   ```

5. **Ingress Controller**
   ```bash
   helm install nginx-ingress nginx-stable/nginx-ingress
   ```

---

## Quick Reference Checklist

### Pre-Production Checklist

- [ ] 3 control plane nodes (odd number)
- [ ] Control planes tainted (no workloads)
- [ ] Multiple availability zones
- [ ] All on-demand instances (no spot for critical)
- [ ] Resource requests/limits set
- [ ] RBAC configured
- [ ] Network policies enabled
- [ ] Secrets management solution
- [ ] Monitoring stack deployed
- [ ] Logging aggregation configured
- [ ] Backup strategy implemented
- [ ] Disaster recovery tested
- [ ] Health checks on all pods
- [ ] Pod disruption budgets set
- [ ] Autoscaling configured
- [ ] CI/CD pipeline integrated

### Security Checklist

- [ ] RBAC enabled with least privilege
- [ ] Pod security policies/standards enforced
- [ ] Images scanned for vulnerabilities
- [ ] Secrets not in Git
- [ ] Network policies blocking default traffic
- [ ] mTLS between pods
- [ ] API server not publicly accessible
- [ ] Audit logging enabled
- [ ] Regular security updates
- [ ] Penetration testing performed

### Operational Checklist

- [ ] GitOps deployed (ArgoCD/FluxCD)
- [ ] Automated backups running
- [ ] Monitoring alerts configured
- [ ] Runbooks documented
- [ ] On-call rotation established
- [ ] Incident response plan
- [ ] Change management process
- [ ] Capacity planning done
- [ ] Cost monitoring enabled
- [ ] Performance baselines established

---

## Related Documentation

- [`PRODUCTION_STACK_SETUP.md`](./PRODUCTION_STACK_SETUP.md) - Production deployment guide
- [`SPOT_INSTANCE_BEST_PRACTICES.md`](./SPOT_INSTANCE_BEST_PRACTICES.md) - Spot instance configuration
- [`UPGRADING_CLUSTER.md`](./UPGRADING_CLUSTER.md) - Upgrade procedures
- [`COST_OPTIMIZATION.md`](./COST_OPTIMIZATION.md) - Cost management strategies

---

## Additional Resources

**Official Documentation:**
- [Kubernetes Best Practices](https://kubernetes.io/docs/setup/best-practices/)
- [Production Checklist](https://kubernetes.io/docs/setup/best-practices/cluster-large/)
- [Security Hardening](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/)

**Tools:**
- [kube-bench](https://github.com/aquasecurity/kube-bench) - CIS benchmark
- [Polaris](https://github.com/FairwindsOps/polaris) - Best practices validation
- [kubesec](https://kubesec.io/) - Security risk analysis

**Learning:**
- [Kubernetes the Hard Way](https://github.com/kelseyhightower/kubernetes-the-hard-way)
- [CNCF Landscape](https://landscape.cncf.io/)
- [Kubernetes Patterns](https://k8spatterns.io/)

---

**Remember**: Production readiness is a journey, not a destination. Continuously improve your cluster based on monitoring insights and industry best practices! 🚀
