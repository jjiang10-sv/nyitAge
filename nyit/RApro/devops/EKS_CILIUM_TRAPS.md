# EKS + Cilium: Real-World Traps & Solutions

## The Fundamental Issue

**AWS does NOT officially support Cilium.** This means:
- ❌ No AWS support tickets for Cilium issues
- ❌ No tested compatibility guarantees
- ❌ No AWS documentation
- ❌ You're responsible for everything

**vs AKS where Microsoft officially supports it** ✅

---

## Critical Traps (Must Know!)

### 🔴 Trap #1: You Own All Cilium Problems

**The Reality:**
```
You: "Cilium pods are crashing after EKS upgrade"
AWS Support: "We don't support Cilium. Use AWS VPC CNI."
You: 😱
```

**What This Means:**
- AWS won't help debug Cilium issues
- You must rely on community/Cilium support
- Can't open AWS support tickets for networking
- No SLA for Cilium-related outages

**Solution:**
1. ✅ Get Cilium Enterprise support (paid)
2. ✅ Build in-house expertise
3. ✅ Have AWS VPC CNI rollback plan
4. ✅ Thoroughly test before production

---

### 🔴 Trap #2: EKS Upgrades Can Break Cilium

**The Problem:**
```bash
# Before upgrade
EKS 1.28 + Cilium 1.14 = ✅ Working

# After upgrade
aws eks update-cluster-version --name my-cluster --kubernetes-version 1.29
# Result: Cilium stops working! ❌
```

**Why It Happens:**
- AWS tests with VPC CNI, not Cilium
- Kubernetes API changes may break Cilium
- Node AMI changes may affect eBPF
- No compatibility guarantee

**Real Example:**
```
EKS 1.28 → 1.29 upgrade:
- New kernel version
- Different eBPF verifier
- Cilium 1.14 fails to load eBPF programs
- Need Cilium 1.15+ (didn't exist yet!)
- Production outage! 🔥
```

**Solution:**
```bash
# ALWAYS test upgrades in staging first!

# 1. Create test cluster with new EKS version
eksctl create cluster --version 1.29 --name test-cluster

# 2. Install Cilium
helm install cilium cilium/cilium

# 3. Run connectivity tests
cilium connectivity test

# 4. If passes, check Cilium compatibility matrix
# https://docs.cilium.io/en/stable/network/kubernetes/compatibility/

# 5. Only then upgrade production
```

**Prevention:**
- ✅ Subscribe to Cilium release notes
- ✅ Check compatibility before EVERY EKS upgrade
- ✅ Test in staging cluster first
- ✅ Have rollback plan ready

---

### 🔴 Trap #3: Must Completely Remove AWS VPC CNI

**The Problem:**
```bash
# EKS installs AWS VPC CNI by default
kubectl get pods -n kube-system | grep aws-node
aws-node-xxxxx  1/1  Running  ← Still there!

# If you don't remove it, you get:
# - Two CNIs fighting for control
# - IP conflicts
# - Pods get wrong IPs
# - Networking breaks randomly
```

**The Fix:**
```python
# In Pulumi
cluster = eks.Cluster(
    default_addons_to_remove=["vpc-cni"],  # CRITICAL!
)
```

**But here's the trap:**
```bash
# Even after removing, sometimes VPC CNI comes back!

# Why? EKS managed addons auto-install
# Need to explicitly disable:
aws eks delete-addon --cluster-name my-cluster --addon-name vpc-cni

# And prevent re-installation:
aws eks describe-addon-versions --addon-name vpc-cni
# Ensure it's not in managed addons
```

**Verification:**
```bash
# Must verify VPC CNI is completely gone
kubectl get daemonset -n kube-system | grep aws-node
# Should return NOTHING

# If it's still there:
kubectl delete daemonset aws-node -n kube-system
kubectl delete serviceaccount aws-node -n kube-system
kubectl delete clusterrole aws-node
kubectl delete clusterrolebinding aws-node
```

---

### 🔴 Trap #4: ENI Limits Can Kill Your Cluster

**The Problem (ENI Mode):**
```
Instance Type: t3.medium
Max ENIs: 3
IPs per ENI: 6
Max pods: (3 × 6) - 3 = 15 pods only! 😱
```

**Real-World Disaster:**
```bash
# You deploy normally
kubectl apply -f app.yaml
# replica count: 20

# Pods start scheduling
kubectl get pods
# 15 running
# 5 pending forever!

# Why?
kubectl describe pod pending-pod-xxx
# Error: "no available IP addresses"
# ENI limit hit!
```

**The Math:**
| Instance Type | Max ENIs | IPs per ENI | Max Pods |
|---------------|----------|-------------|----------|
| t3.small | 3 | 4 | 9 |
| t3.medium | 3 | 6 | 15 |
| t3.large | 3 | 12 | 33 |
| m5.large | 3 | 10 | 27 |
| m5.xlarge | 4 | 15 | 56 |

**Solutions:**

**Option 1: Use overlay mode** (no ENI limits)
```yaml
ipam:
  mode: cluster-pool  # Not ENI!
tunnel: geneve
```

**Option 2: Use larger instances**
```python
instance_types=["m5.xlarge"]  # Instead of t3.medium
```

**Option 3: Prefix delegation** (advanced)
```yaml
eni:
  enabled: true
  awsEnablePrefixDelegation: true  # Increases IPs per ENI
```

---

### 🔴 Trap #5: IAM Permissions Nightmare

**The Problem:**
Cilium needs IAM permissions to manage ENIs, but getting it right is tricky.

**What Cilium Needs:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:AttachNetworkInterface",
        "ec2:DeleteNetworkInterface",
        "ec2:DetachNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeInstances",
        "ec2:ModifyNetworkInterfaceAttribute",
        "ec2:AssignPrivateIpAddresses",
        "ec2:UnassignPrivateIpAddresses"
      ],
      "Resource": "*"
    }
  ]
}
```

**Common Mistakes:**

**Mistake 1: Forgetting to attach policy**
```bash
# Cilium operator pods crash
kubectl logs -n kube-system cilium-operator-xxx

# Error: "AccessDenied: Not authorized to create ENI"
```

**Mistake 2: Using instance profile instead of IRSA**
```python
# Bad: Instance profile (too broad)
# Good: IRSA (scoped)

# Create IRSA
cilium_role = iam.Role("cilium-operator-role",
    assume_role_policy=eks_oidc_provider.arn.apply(
        lambda arn: json.dumps({
            "Version": "2012-10-17",
            "Statement": [{
                "Effect": "Allow",
                "Principal": {
                    "Federated": arn
                },
                "Action": "sts:AssumeRoleWithWebIdentity",
            }]
        })
    )
)

# Attach to Cilium operator service account
```

**Mistake 3: Subnet permissions**
```bash
# Cilium needs to query subnet info
# Missing permission:
"ec2:DescribeSubnets",
"ec2:DescribeVpcs",
```

---

### 🔴 Trap #6: Cilium Version vs Kubernetes Version Hell

**The Compatibility Matrix:**
```
Kubernetes 1.28:
  ✅ Cilium 1.14.x
  ✅ Cilium 1.15.x
  ❌ Cilium 1.12.x (too old)

Kubernetes 1.29:
  ❌ Cilium 1.14.0-1.14.4 (broken)
  ✅ Cilium 1.14.5+ (fixed)
  ✅ Cilium 1.15.x

Kubernetes 1.30:
  ❌ Cilium 1.14.x (deprecated APIs)
  ❌ Cilium 1.15.0-1.15.2 (broken)
  ✅ Cilium 1.15.3+
```

**The Trap:**
```bash
# You're on EKS 1.28 + Cilium 1.14.3
# AWS auto-upgrades to 1.29 (you enabled auto-upgrade)
# Cilium breaks!
# No rollback option! 😱
```

**Solution:**
```bash
# NEVER enable auto-upgrade with Cilium
aws eks update-cluster-config \
  --name my-cluster \
  --no-enable-auto-update

# Always check compatibility FIRST
curl -s https://docs.cilium.io/en/stable/network/kubernetes/compatibility/ | \
  grep "1.29"

# Upgrade Cilium BEFORE EKS
helm upgrade cilium cilium/cilium --version 1.14.5
# Wait for rollout
# THEN upgrade EKS
aws eks update-cluster-version --kubernetes-version 1.29
```

---

### 🔴 Trap #7: Security Groups Don't Work Per-Pod

**The Problem:**
```python
# With AWS VPC CNI, you can do:
pod_security_group = ec2.SecurityGroup(...)

# And attach to specific pods
# But with Cilium in overlay mode: ❌ DOESN'T WORK!
```

**Why:**
- AWS security groups work at ENI level
- Overlay mode = pods don't have ENIs
- Security groups can't see pod IPs

**Workaround:**
```yaml
# Use Cilium Network Policies instead
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: restrict-access
spec:
  endpointSelector:
    matchLabels:
      app: frontend
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: backend
```

**Or use ENI mode** (but ENI limits apply):
```yaml
eni:
  enabled: true  # Now pods get ENIs
# Can use security groups, but fewer pods per node
```

---

### 🔴 Trap #8: Cluster Mesh is Harder

**The Problem:**
```bash
# With AKS + Cilium:
# Microsoft manages some cluster mesh setup

# With EKS + Cilium:
# You do EVERYTHING manually
```

**What You Need to Do:**

**1. Expose Cilium cluster mesh API**
```yaml
# Need LoadBalancer service
apiVersion: v1
kind: Service
metadata:
  name: clustermesh-apiserver
spec:
  type: LoadBalancer  # Costs money on AWS!
  ports:
  - port: 2379
```

**2. Configure VPC peering yourself**
```bash
aws ec2 create-vpc-peering-connection \
  --vpc-id vpc-xxx \
  --peer-vpc-id vpc-yyy \
  --peer-region us-east-1
```

**3. Update route tables manually**
```bash
# For EACH region, add routes
aws ec2 create-route \
  --route-table-id rtb-xxx \
  --destination-cidr-block 10.1.0.0/16 \
  --vpc-peering-connection-id pcx-xxx
```

**4. Exchange cluster mesh secrets**
```bash
# Extract from cluster 1
cilium clustermesh connection extract --context cluster1 > cluster1.yaml

# Apply to cluster 2
kubectl apply -f cluster1.yaml --context cluster2
```

**It's doable, but lots of manual work!**

---

### 🔴 Trap #9: Debugging is Much Harder

**The Problem:**

**With AKS + Cilium:**
```bash
# Azure support can help
az support tickets create \
  --issue-type technical \
  --summary "Cilium pods crashing"

# They have internal tools
# They've seen similar issues
# Fast resolution
```

**With EKS + Cilium:**
```bash
# AWS support
aws support create-case \
  --subject "Cilium issue"

# Response: "Cilium is not supported. Use VPC CNI."
# You're on your own! 😱
```

**What You Need:**
1. Strong in-house expertise
2. Cilium Enterprise support ($$$)
3. Community Slack
4. Read Cilium source code
5. Trial and error

**Example Debug Session:**
```bash
# Pod can't reach service
# With VPC CNI: AWS support helps
# With Cilium: You must debug

# Check Cilium status
kubectl exec -n kube-system cilium-xxx -- cilium status

# Check eBPF programs
kubectl exec -n kube-system cilium-xxx -- cilium bpf lb list

# Check service backend
kubectl exec -n kube-system cilium-xxx -- cilium service list

# Check connectivity
kubectl exec -n kube-system cilium-xxx -- cilium monitor

# Read logs
kubectl logs -n kube-system cilium-xxx --tail=1000

# Check endpoints
kubectl exec -n kube-system cilium-xxx -- cilium endpoint list

# Hours of debugging later...
# Finally find issue: eBPF conntrack table full
# Increase limits in Cilium config
```

---

### 🔴 Trap #10: No Tested Upgrade Path

**The Problem:**
```
AWS tests: EKS 1.28 → 1.29 with VPC CNI ✅
AWS does NOT test: EKS 1.28 → 1.29 with Cilium ❌
```

**Real Example:**
```bash
# Upgrade EKS from 1.27 → 1.28
# Following AWS docs (which assume VPC CNI)

# Step 1: Upgrade control plane ✅
# Step 2: AWS doc says "update VPC CNI addon" 
#         (you skip this, you use Cilium)
# Step 3: Upgrade nodes ✅
# Step 4: Pods start crashing! ❌

# Why? Node OS update changed:
# - Kernel version
# - iptables implementation  
# - cgroup version
# - eBPF verifier strictness

# Cilium config needs update! But which settings?
# AWS doesn't know. You must figure it out.
```

**Solution:**
```bash
# Create detailed testing plan

# 1. Review Cilium release notes
curl https://github.com/cilium/cilium/releases

# 2. Check Cilium compatibility
# https://docs.cilium.io/en/stable/network/kubernetes/compatibility/

# 3. Create test cluster
eksctl create cluster --version 1.29 --name upgrade-test

# 4. Install exact Cilium version you use
helm install cilium cilium/cilium--version 1.14.5

# 5. Deploy all your critical apps

# 6. Run extensive tests
cilium connectivity test
kubectl run test --image=busybox -- sh
# Test everything!

# 7. Document any issues found
# 8. Create runbook for production upgrade
# 9. Schedule maintenance window
# 10. Upgrade with rollback plan ready
```

---

## Comparison: AKS vs EKS Support

| Scenario | AKS (Official Support) | EKS (No Support) |
|----------|----------------------|------------------|
| **Cilium crashes** | Open Azure ticket ✅ | Community Slack ❌ |
| **After upgrade** | Microsoft tested ✅ | You test ❌ |
| **Network issue** | Azure helps debug ✅ | You debug ❌ |
| **Security patch** | Microsoft applies ✅ | You apply ❌ |
| **New feature** | Tested by MS ✅ | You test ❌ |
| **SLA coverage** | Yes ✅ | No ❌ |
| **Docs** | Official docs ✅ | Community ❌ |
| **Tested configs** | MS provides ✅ | DIY ❌ |

---

## How to Mitigate Risks

### 1. Get Cilium Enterprise Support

```
Cost: $10k-50k+/year
Benefit: 
- Cilium team helps debug
- Priority bug fixes
- Tested configurations
- Upgrade assistance
```

### 2. Build Internal Expertise

```bash
# Train team on:
- eBPF fundamentals
- Cilium architecture
- Kubernetes networking
- AWS VPC networking

# Recommended:
- Cilium certification
- eBPF summit attendance
- Run Cilium in dev/staging first
```

### 3. Maintain Rollback Plan

```python
# Keep AWS VPC CNI ready
cluster = eks.Cluster(
    # Don't remove VPC CNI addon entirely
    # Just disable it
)

# Emergency rollback procedure:
# 1. helm uninstall cilium
# 2. Enable VPC CNI
# 3. Restart all pods
# 4. Back in business (with VPC CNI)
```

### 4. Extensive Testing

```bash
# Test matrix:
- Cilium version X × EKS version Y
- Security groups
- Load balancers  
- Network policies
- Cross-AZ traffic
- Cross-region (cluster mesh)
- Performance benchmarks
- Failure scenarios

# Automate tests
# Run in CI/CD
# Block production deployment if fails
```

### 5. Monitor Everything

```yaml
# Deploy comprehensive monitoring
prometheus:
  enabled: true
  
hubble:
  enabled: true
  metrics:
    enabled: ["dns", "drop", "tcp", "flow", "http"]
    
# Set up alerts
# - Cilium pod restarts
# - eBPF program failures
# - IP exhaustion
# - Conntrack table full
# - Service backend unavailable
```

---

## When to Use AKS Instead

### Use AKS if:

1. ✅ **Official support is critical**
   - Can't afford downtime
   - Need SLA coverage
   - Enterprise requirements

2. ✅ **Limited Cilium expertise**
   - Small team
   - No time to learn eBPF
   - Need vendor support

3. ✅ **Regulated industry**
   - Finance, healthcare, government
   - Need certified configurations
   - Compliance requirements

4. ✅ **Risk-averse**
   - Can't accept community support
   - Need guaranteed upgrades
   - Want tested configurations

### Use EKS if:

1. ✅ **Cost is primary concern**
   - 5-6x cheaper
   - Smaller budget
   - Startup/scale-up

2. ✅ **Have Cilium expertise**
   - Team knows eBPF
   - Can debug networking
   - Comfortable with OSS

3. ✅ **Need full Cilium features**
   - Cluster mesh
   - BGP
   - Custom IPAM
   - Advanced features

4. ✅ **Can invest in testing**
   - Good CI/CD
   - Multiple environments
   - Time for thorough testing

---

## Summary: The Real Traps

### Top 5 Traps:

1. **No AWS support** - You're completely on your own
2. **EKS upgrades can break Cilium** - Must test thoroughly
3. **ENI limits** - Easy to hit pod limits
4. **IAM complexity** - Getting permissions right is tricky
5. **Debugging is hard** - No AWS tools or help

### Risk Mitigation:

- ✅ Get Cilium Enterprise support
- ✅ Build internal expertise
- ✅ Keep rollback plan ready
- ✅ Test everything thoroughly
- ✅ Monitor comprehensively

### Bottom Line:

**EKS + Cilium is powerful but risky without:**
1. Cilium expertise
2. Enterprise support OR
3. Strong testing culture

**AKS + Cilium is safer because:**
1. Microsoft officially supports it
2. Tested configurations
3. SLA coverage
4. Reliable upgrades

**Choose based on your risk tolerance and expertise!** 🎯


