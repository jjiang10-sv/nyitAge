# Blue-Green Deployment for Production Kubernetes Clusters

A comprehensive guide for zero-downtime cluster updates using blue-green deployment strategy.

## Understanding Pulumi Updates

### ❓ Do I Need to Destroy First?

**Short answer: NO**

When you change Pulumi code and run `pulumi up`, Pulumi automatically:
- ✅ **Creates** new resources that don't exist
- ✅ **Updates** existing resources in-place (when possible)
- ✅ **Replaces** resources that require recreation
- ✅ **Deletes** resources you removed from code

### Pulumi Update Behavior

```bash
# You change this in Pulumi code:
worker_instance_type="t3.medium"  # Was t3.small

# Then run:
pulumi up

# Pulumi will:
# 1. Show you the plan:
#    ~ 3 to update (replace)
#    
# 2. Replace workers (one by one or all at once)
# 3. Keep everything else unchanged
```

### When Pulumi Updates In-Place vs Replaces

| Change | Pulumi Behavior | Requires Destroy? |
|--------|----------------|-------------------|
| Add new worker | ✅ Creates new instance | No |
| Change instance type | 🔄 Replaces instances | No |
| Change security group rules | ✅ Updates in-place | No |
| Change user-data script | 🔄 Replaces instances | No |
| Add tags | ✅ Updates in-place | No |
| Change VPC | ❌ Can't update | Yes (or blue-green) |

### Example: Adding a Worker

```bash
# 1. Change config
pulumi config set on_demand_worker_count 2  # Was 1

# 2. Preview
pulumi preview
# Output:
#   + 1 to create (new worker)
#   ~ 0 to update
#   - 0 to delete

# 3. Apply
pulumi up
# Creates one new worker, keeps everything else
```

### Example: Changing Instance Type

```bash
# 1. Change config
pulumi config set worker_instance_type t3.large  # Was t3.medium

# 2. Preview
pulumi preview
# Output:
#   ~ 3 to replace (workers)
#   +- (replace instances)

# 3. Apply
pulumi up
# Pulumi will:
# - Create new t3.large instances
# - Delete old t3.medium instances
# - Keep control plane unchanged
```

**Note**: This causes **brief downtime** as workers are replaced. For zero downtime, use blue-green deployment.

---

## Blue-Green Deployment Strategy

### What is Blue-Green?

```
Current State (Green):
├─ Production traffic → Green Cluster
└─ Blue Cluster doesn't exist yet

Transition:
├─ Create Blue Cluster (new)
├─ Deploy apps to Blue
├─ Test Blue cluster
├─ Switch traffic: Green → Blue
└─ Keep Green as backup

Final State:
├─ Production traffic → Blue Cluster ✅
└─ Green Cluster (standby for rollback)

Cleanup (after validation):
└─ Delete Green Cluster
```

### Benefits

- ✅ **Zero downtime** - Traffic switch is instant
- ✅ **Easy rollback** - Switch traffic back to green
- ✅ **Test in prod** - Blue is identical to final state
- ✅ **Safe updates** - Green stays alive as backup
- ✅ **Confidence** - Validate before committing

### Cost

- ⚠️ **Double infrastructure cost** during deployment (typically 2-4 hours)
- **Green cluster** (~$360/month) + **Blue cluster** (~$360/month) = ~$720/month
- After cleanup: Back to ~$360/month

---

## Complete Blue-Green Deployment Guide

### Prerequisites

```bash
# 1. Working green (current) cluster
pulumi stack select prod
pulumi stack  # Verify it's deployed

# 2. Backup current state
ssh ubuntu@$(pulumi stack output control_plane_public_ips)
sudo etcdctl snapshot save /backup/etcd-$(date +%Y%m%d).db

# 3. Export app manifests
kubectl get all --all-namespaces -o yaml > green-cluster-backup.yaml

# 4. Git commit current state
git add .
git commit -m "Pre-blue-green backup"
git push
```

---

## Step-by-Step Deployment

### Phase 1: Create Blue Cluster

**Step 1.1: Create Blue Stack**

```bash
cd /Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s

# Create new stack for blue cluster
pulumi stack init prod-blue

# Verify stacks
pulumi stack ls
# Should show:
# - prod (green - current)
# - prod-blue (new)
```

**Step 1.2: Configure Blue Stack**

```bash
# Switch to blue
pulumi stack select prod-blue

# Copy config from green, with different cluster name
pulumi config set aws:region us-west-2
pulumi config set cluster_name prod-blue-k8s  # Different name!
pulumi config set k8s_version 1.31.0          # New version
pulumi config set cilium_version 1.16.5

# Infrastructure config
pulumi config set control_plane_count 3
pulumi config set control_plane_instance_type t3.medium
pulumi config set on_demand_worker_count 3
pulumi config set spot_worker_count 0
pulumi config set worker_instance_type t3.large

# Verify config
pulumi config
```

**Step 1.3: Preview Blue Cluster**

```bash
pulumi preview

# Expected output:
# + ~26 resources to create
# (completely new cluster)
```

**Step 1.4: Deploy Blue Cluster**

```bash
# Create blue cluster
pulumi up

# Confirm with 'yes'
# Wait ~10-15 minutes for infrastructure

# Save outputs
pulumi stack output control_plane_public_ips > blue-cp-ips.txt
pulumi stack output worker_public_ips > blue-worker-ips.txt
```

### Phase 2: Initialize Blue Cluster

**Step 2.1: SSH to Blue Control Plane**

```bash
# Get blue CP IP
BLUE_CP_IP=$(pulumi stack output control_plane_public_ips | head -n1)

# SSH
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@$BLUE_CP_IP
```

**Step 2.2: Initialize First Control Plane**

```bash
# On blue CP
sudo /root/init-cluster.sh

# Wait ~10 minutes for:
# - kubeadm init
# - Cilium installation
# - SPIRE installation
# - EBS CSI driver

# Verify
kubectl get nodes
# Should show: 1 control plane, NotReady (waiting for CNI)

kubectl get pods -n kube-system
# Should show: Cilium pods running
```

**Step 2.3: Join Other Control Planes (if 3 CPs)**

```bash
# On first CP, get join command
kubeadm token create --print-join-command --certificate-key $(kubeadm init phase upload-certs --upload-certs 2>/dev/null | tail -1)

# Copy output, SSH to CP2 and CP3:
ssh ubuntu@<CP2_IP>
sudo <paste join command with --control-plane>

# Repeat for CP3

# Verify
kubectl get nodes
# Should show: 3 control planes
```

**Step 2.4: Join Workers**

```bash
# On any CP, get worker join command
kubeadm token create --print-join-command

# SSH to each worker
ssh ubuntu@<WORKER_IP>
sudo <paste join command>

# Verify all nodes joined
kubectl get nodes
# Should show: 3 CPs + 3 workers, all Ready
```

### Phase 3: Deploy Applications to Blue

**Step 3.1: Configure kubectl for Blue**

```bash
# On local machine
# Copy kubeconfig from blue cluster
scp -i ~/.ssh/k8s-dev-key.pem \
  ubuntu@$BLUE_CP_IP:/etc/kubernetes/admin.conf \
  ~/.kube/config-blue

# Set environment variable
export KUBECONFIG=~/.kube/config-blue

# Test
kubectl get nodes
# Should show blue cluster nodes
```

**Step 3.2: Deploy Apps via GitOps (ArgoCD)**

If using ArgoCD:

```bash
# Install ArgoCD on blue cluster
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Wait for ArgoCD
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d

# Port forward to ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Access UI: https://localhost:8080
# Login: admin / <password from above>

# Sync all apps
argocd app sync --all
```

**Step 3.3: Deploy Apps Manually (if no GitOps)**

```bash
# Apply manifests
kubectl apply -f green-cluster-backup.yaml

# Or deploy from Git
kubectl apply -f https://github.com/myorg/myapp/k8s/

# Verify deployments
kubectl get deployments --all-namespaces
kubectl get pods --all-namespaces
```

**Step 3.4: Configure Ingress/Load Balancer**

```bash
# Install ingress controller on blue
helm install nginx-ingress nginx-stable/nginx-ingress

# Get blue LB address
kubectl get svc nginx-ingress-controller
# Note the EXTERNAL-IP

# Update DNS (don't switch yet, just prepare)
# Create DNS record: blue.example.com → <BLUE_LB_IP>
```

### Phase 4: Validate Blue Cluster

**Step 4.1: Smoke Tests**

```bash
# Test pods are running
kubectl get pods --all-namespaces
# All pods should be Running

# Test services are accessible
kubectl run test-pod --image=curlimages/curl --rm -it -- \
  curl http://my-app-service.production.svc.cluster.local

# Test external access (using blue.example.com)
curl https://blue.example.com/health
```

**Step 4.2: Load Testing**

```bash
# Run load test against blue cluster
kubectl run loadtest --image=williamyeh/hey \
  -- -n 1000 -c 10 http://my-app-service.production

# Monitor
kubectl top nodes
kubectl top pods -n production
```

**Step 4.3: Integration Tests**

```bash
# Run full test suite
kubectl apply -f tests/integration-tests.yaml

# Check results
kubectl logs -f job/integration-tests
```

**Step 4.4: Validation Checklist**

- [ ] All pods running and ready
- [ ] All services responding
- [ ] Database connections working
- [ ] External API calls successful
- [ ] Monitoring/logging working
- [ ] Performance acceptable
- [ ] No errors in logs
- [ ] Cert-manager certificates valid
- [ ] Persistent volumes attached
- [ ] Network policies enforced

### Phase 5: Traffic Cutover

**Step 5.1: Pre-Cutover Verification**

```bash
# Green cluster (current production)
export KUBECONFIG=~/.kube/config-green
kubectl get nodes  # Should show green cluster

# Blue cluster (new)
export KUBECONFIG=~/.kube/config-blue
kubectl get nodes  # Should show blue cluster
```

**Step 5.2: Update DNS/Load Balancer**

**Option A: DNS Update (Gradual)**

```bash
# Update DNS to point to blue cluster LB
# production.example.com: <GREEN_LB> → <BLUE_LB>

# DNS TTL determines how fast (usually 60-300 seconds)
# Traffic gradually shifts as DNS propagates
```

**Option B: Load Balancer Target Update (Instant)**

```bash
# If using AWS ALB/NLB:
aws elbv2 modify-target-group \
  --target-group-arn <TG_ARN> \
  --health-check-enabled \
  --targets Id=<BLUE_INSTANCE_IDS>

# Traffic switches immediately
```

**Option C: Ingress Controller Update**

```bash
# Update ingress to point to blue services
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind:Ingress
metadata:
  name: production
spec:
  rules:
  - host: production.example.com
    http:
      paths:
      - path: /
        backend:
          service:
            name: app-service-blue  # Changed from app-service-green
            port:
              number: 80
EOF
```

**Step 5.3: Monitor Traffic Shift**

```bash
# Watch incoming requests on blue
export KUBECONFIG=~/.kube/config-blue
kubectl logs -f deployment/app --tail=100

# Watch green traffic drop
export KUBECONFIG=~/.kube/config-green
kubectl logs -f deployment/app --tail=100

# Monitor errors
kubectl get events --all-namespaces --sort-by='.lastTimestamp'
```

### Phase 6: Post-Cutover Monitoring

**Step 6.1: Active Monitoring (First 2 Hours)**

```bash
# Monitor error rates
kubectl top pods -n production

# Watch for crashloops
kubectl get pods -n production --watch

# Check logs
kubectl logs -f deployment/app | grep ERROR

# Monitor external access
while true; do
  curl -s -o /dev/null -w "%{http_code}\n" https://production.example.com
  sleep 5
done
```

**Step 6.2: Validation Period (24-48 Hours)**

During this period:
- ✅ Monitor application metrics
- ✅ Check error rates vs green baseline
- ✅ Verify all features working
- ✅ Monitor cost (should be 2x during overlap)
- ⚠️ Keep green cluster running (rollback safety)

### Phase 7: Rollback (If Needed)

**If issues detected within 48 hours:**

```bash
# 1. Switch traffic back to green
# Update DNS: production.example.com → <GREEN_LB>

# 2. Verify green still works
export KUBECONFIG=~/.kube/config-green
kubectl get pods -A  # Should all be Running

# 3. Investigate blue cluster
export KUBECONFIG=~/.kube/config-blue
kubectl logs <problematic-pod>

# 4. Fix and retry later
pulumi stack select prod-blue
pulumi destroy  # Clean up blue cluster
```

**Rollback is instant** - just switch DNS/LB back!

### Phase 8: Cleanup (After Success)

**After 48 hours of stable blue operation:**

```bash
# 1. Verify blue is stable
export KUBECONFIG=~/.kube/config-blue
kubectl get pods -A
# All should be healthy

# 2. Backup blue (now production)
ssh ubuntu@$BLUE_CP_IP
sudo etcdctl snapshot save /backup/etcd-blue-stable.db

# 3. Destroy green cluster
pulumi stack select prod
pulumi destroy

# Confirm with 'yes'
# Wait ~2-3 minutes

# 4. Rename blue → prod
pulumi stack select prod-blue
pulumi stack rename prod

# 5. Update kubeconfig
mv ~/.kube/config-blue ~/.kube/config

# 6. Cleanup
rm blue-cp-ips.txt blue-worker-ips.txt

# Cost is now back to normal (~$360/month)
```

---

## Blue-Green Deployment Timeline

### Typical Production Deployment

```
Hour 0:00 - Create blue stack
Hour 0:15 - Blue infrastructure ready
Hour 0:30 - Blue cluster initialized
Hour 1:00 - Applications deployed
Hour 1:30 - Validation tests pass
Hour 2:00 - Traffic cutover
Hour 2:05 - Monitor traffic shift
Hour 2:30 - Initial stability confirmed
──────────────────────────────────────
Hour 24:00 - 24-hour validation period ends
Hour 48:00 - Destroy green cluster
```

**Total time commitment:** ~2-3 hours active work  
**Total validation:** 48 hours monitoring  
**Double cost period:** 48 hours (~$24 extra)

---

## Production Deployment Checklist

### Pre-Deployment

- [ ] Backup green cluster (etcd + manifests)
- [ ] Git commit current state
- [ ] Schedule maintenance window (optional, zero downtime)
- [ ] Notify team of deployment
- [ ] Prepare rollback procedure

### Blue Cluster Creation

- [ ] Create prod-blue stack
- [ ] Configure with new settings
- [ ] Deploy infrastructure (`pulumi up`)
- [ ] Initialize Kubernetes cluster
- [ ] Join all control planes and workers

### Application Deployment

- [ ] Deploy apps to blue
- [ ] Run smoke tests
- [ ] Run integration tests
- [ ] Load test
- [ ] Verify all features

### Traffic Cutover

- [ ] Update DNS/LB to blue
- [ ] Monitor traffic shift
- [ ] Check error rates
- [ ] Verify user experience

### Post-Deployment

- [ ] 2-hour active monitoring
- [ ] 24-hour passive monitoring
- [ ] 48-hour validation period
- [ ] Destroy green cluster
- [ ] Rename blue → prod
- [ ] Update documentation

---

## Advanced Scenarios

### Scenario 1: Gradual Migration (Canary)

```bash
# Route 10% traffic to blue, 90% to green
# AWS ALB weighted target groups:
aws elbv2 modify-listener \
  --listener-arn <ARN> \
  --default-actions '{
    "Type": "forward",
    "ForwardConfig": {
      "TargetGroups": [
        {"TargetGroupArn": "<GREEN_TG>", "Weight": 90},
        {"TargetGroupArn": "<BLUE_TG>", "Weight": 10}
      ]
    }
  }'

# Gradually increase blue: 10% → 25% → 50% → 100%
```

### Scenario 2: Database Migration

**For stateful apps:**

```bash
# 1. Replicate database to blue
pg_dump green_db | psql blue_db

# 2. Set up replication (green → blue)
# 3. Deploy apps on blue (read-only mode)
# 4. Stop writes on green
# 5. Final sync
# 6. Cut over to blue
# 7. Enable writes on blue
```

### Scenario 3: Multi-Region Blue-Green

```bash
# Create blue clusters in all regions simultaneously
pulumi stack init prod-blue-us-west-2
pulumi stack init prod-blue-us-east-1
pulumi stack init prod-blue-eu-west-1

# Deploy to all
for stack in prod-blue-*; do
  pulumi stack select $stack
  pulumi up
done

# Global load balancer switches regions together
```

---

## Cost Analysis

### Standard Blue-Green Deployment

```
Green Cluster (Running):     $360/month
Blue Cluster (Created):      $360/month (for 48 hours)
───────────────────────────────────────
Total for 48 hours:          $360 + ($360 ÷ 30 × 2) = $384
Additional cost:             $24 for safety/testing
```

### Cost Optimization

```bash
# Option 1: Faster cleanup (12 hours instead of 48)
# Save: $12

# Option 2: Use smaller blue for testing first
pulumi config set worker_instance_type t3.medium  # Cheaper
# Then upgrade after validation

# Option 3: Schedule during low-traffic
# Less risk = faster cleanup
```

---

## Troubleshooting

### Issue: Blue Cluster Won't Initialize

```bash
# Check control plane logs
ssh ubuntu@$BLUE_CP_IP
sudo journalctl -u kubelet -f

# Common issues:
# - Incorrect user-data
# - Network connectivity
# - AWS IAM permissions
```

### Issue: Apps Not Starting on Blue

```bash
# Check pod events
kubectl describe pod <pod-name>

# Check persistent volumes
kubectl get pv
kubectl get pvc

# Verify Cilium
cilium status
```

### Issue: High Error Rate After Cutover

```bash
# Rollback immediately
# Update DNS: blue → green

# Investigate
kubectl logs deployment/app | grep ERROR

# Common causes:
# - Database connection issues
# - Missing environment variables
# - Network policy blocks
```

---

## Related Documentation

- [`CLUSTER_UPDATE_STRATEGIES.md`](./CLUSTER_UPDATE_STRATEGIES.md) - Update strategy decision guide
- [`UPGRADING_CLUSTER.md`](./UPGRADING_CLUSTER.md) - In-place upgrade procedures
- [`K8S_CLUSTER_BEST_PRACTICES.md`](./K8S_CLUSTER_BEST_PRACTICES.md) - Production best practices
- [`PRODUCTION_STACK_SETUP.md`](./PRODUCTION_STACK_SETUP.md) - Initial production setup

---

## Summary

**Key Points:**

1. **Pulumi updates DON'T require manual destroy** - `pulumi up` handles updates intelligently
2. **Blue-green is for zero-downtime production** - Worth the ~$24 extra cost
3. **Keep green running for 48 hours** - Safety margin for rollback
4. **Test blue thoroughly before cutover** - Validation is critical
5. **Traffic switch is instant** - Rollback is just as fast

**When to use blue-green:**
- ✅ Production deployments
- ✅ Major infrastructure changes
- ✅ Kubernetes version updates
- ✅ Critical system updates

**When NOT needed:**
- ❌ Development environments
- ❌ Application updates (use rolling updates)
- ❌ Config changes (use kubectl edit)
- ❌ Adding resources (pulumi up is fine)

**The blue-green deployment is the gold standard for production infrastructure updates!** 🚀
