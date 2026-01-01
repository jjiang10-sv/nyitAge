# Production Network Architecture for Self-Managed Kubernetes on AWS

## Overview

This guide covers **production-grade networking** for self-managed Kubernetes clusters on AWS, implementing AWS best practices for security, high availability, and scalability.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Design Principles](#design-principles)
3. [Network Layout](#network-layout)
4. [Subnet Sizing](#subnet-sizing)
5. [Complete Implementation](#complete-implementation)
6. [Security Hardening](#security-hardening)
7. [High Availability](#high-availability)
8. [Cost Analysis](#cost-analysis)
9. [Migration from Dev](#migration-from-dev)
10. [Monitoring & Troubleshooting](#monitoring--troubleshooting)

---

## Architecture Overview

### Production Network Design

```
┌───────────────────────────────────────────────────────────────────┐
│ VPC: 10.0.0.0/16 (65,536 IPs)                                     │
│                                                                    │
│  ┌────────────────── AZ-1 (us-west-2a) ──────────────────┐        │
│  │                                                        │        │
│  │  ┌─────────────────────────────────────────────────┐  │        │
│  │  │ Public Subnet: 10.0.0.0/20 (4,096 IPs)          │  │        │
│  │  │ ├─ Internet Gateway (shared)                    │  │        │
│  │  │ ├─ NAT Gateway #1 ⭐                             │  │        │
│  │  │ ├─ Bastion Host (optional)                      │  │        │
│  │  │ └─ Public Load Balancers                        │  │        │
│  │  └─────────────────────────────────────────────────┘  │        │
│  │                          ↓                             │        │
│  │  ┌─────────────────────────────────────────────────┐  │        │
│  │  │ Private Subnet: 10.0.16.0/20 (4,096 IPs)        │  │        │
│  │  │ ├─ Kubernetes Control Plane                     │  │        │
│  │  │ ├─ Kubernetes Worker Nodes                      │  │        │
│  │  │ ├─ Internal Load Balancers                      │  │        │
│  │  │ └─ Routes via NAT Gateway #1                    │  │        │
│  │  └─────────────────────────────────────────────────┘  │        │
│  └────────────────────────────────────────────────────────┘        │
│                                                                    │
│  ┌────────────────── AZ-2 (us-west-2b) ──────────────────┐        │
│  │  Public Subnet: 10.0.32.0/20  + NAT Gateway #2        │        │
│  │  Private Subnet: 10.0.48.0/20 + K8s Nodes             │        │
│  └────────────────────────────────────────────────────────┘        │
│                                                                    │
│  ┌────────────────── AZ-3 (us-west-2c) ──────────────────┐        │
│  │  Public Subnet: 10.0.64.0/20  + NAT Gateway #3        │        │
│  │  Private Subnet: 10.0.80.0/20 + K8s Nodes             │        │
│  └────────────────────────────────────────────────────────┘        │
└───────────────────────────────────────────────────────────────────┘
```

---

## Design Principles

### 1. Defense in Depth

**Layered Security:**
```
Internet
  ↓
Internet Gateway (public only)
  ↓
Public Subnets (NAT, LBs only)
  ↓
NAT Gateways (single purpose)
  ↓
Private Subnets (K8s nodes)
  ↓
Security Groups (locked down)
  ↓
Network Policies (Cilium)
  ↓
Pod Security Standards
```

### 2. High Availability

**Multi-AZ Deployment:**
- ✅ 3 Availability Zones
- ✅ NAT Gateway per AZ (no single point of failure)
- ✅ Nodes distributed across AZs
- ✅ EBS volumes in same AZ as nodes
- ✅ Load balancers cross-zone

### 3. Scalability

**Generous Subnet Sizing:**
```
/20 subnets = 4,096 IPs each
- Room for 100+ nodes per AZ
- Supports large cluster growth
- No re-IPing needed later
```

### 4. Cost Optimization

**Right-Sized Resources:**
- NAT Gateways: One per AZ (necessary for HA)
- Subnets: Free, size generously
- Route tables: Minimal
- Security groups: Shared where possible

---

## Network Layout

### VPC CIDR Block

```yaml
VPC: 10.0.0.0/16
Total IPs: 65,536
Reserved for: Kubernetes infrastructure only
```

### Subnet Allocation

#### AZ-1 (us-west-2a)
```yaml
Public Subnet:
  CIDR: 10.0.0.0/20
  IPs: 4,096
  Purpose: NAT Gateway, Bastion, Public LBs
  
Private Subnet:
  CIDR: 10.0.16.0/20
  IPs: 4,096
  Purpose: Kubernetes nodes, Internal LBs
```

#### AZ-2 (us-west-2b)
```yaml
Public Subnet:
  CIDR: 10.0.32.0/20
  IPs: 4,096
  
Private Subnet:
  CIDR: 10.0.48.0/20
  IPs: 4,096
```

#### AZ-3 (us-west-2c)
```yaml
Public Subnet:
  CIDR: 10.0.64.0/20
  IPs: 4,096
  
Private Subnet:
  CIDR: 10.0.80.0/20
  IPs: 4,096
```

### Reserved Space for Future

```yaml
10.0.96.0/19 → 10.0.127.255
Total: 8,192 IPs reserved
Use for: Database subnets, application subnets, etc.
```

---

## Subnet Sizing

### Why /20 Subnets?

**Capacity Planning:**
```
/24 subnet = 256 IPs:
  - AWS reserves 5
  = 251 usable
  ≈ 50 EC2 instances max
  ❌ Too small for production!

/20 subnet = 4,096 IPs:
  - AWS reserves 5
  = 4,091 usable
  ≈ 800+ EC2 instances
  ✅ Production-ready!
```

### IP Allocation Examples

**100-Node Cluster:**
```
Control Plane: 3 nodes
Workers: 97 nodes
NAT Gateways: 3
Bastion: 1
Load Balancers: ~10
Total IPs used: ~120

Available in /20: 4,091
Utilization: 3%
Room to grow: 97%! ✅
```

**500-Node Cluster:**
```
Control Plane: 3 nodes
Workers: 497 nodes
Infrastructure: ~20
Total IPs used: ~520

Available in /20: 4,091
Utilization: 13%
Room to grow: 87% ✅
```

---

## Complete Implementation

### Production example_usage.py

```python
"""
Production-Grade Network Architecture
- Public + Private subnets per AZ
- NAT Gateway per AZ for HA
- Kubernetes nodes in private subnets
- Proper tagging for Kubernetes discovery
"""

import pulumi
import pulumi_aws as aws
from platform import SelfManagedK8sCluster

# Configuration
config = pulumi.Config()
cluster_name = config.get("cluster_name") or "prod-k8s"
kubernetes_version = config.get("k8s_version") or "1.29.0"
cilium_version = config.get("cilium_version") or "1.15.1"

# ========================================
# VPC
# ========================================

vpc = aws.ec2.Vpc(
    "prod-vpc",
    cidr_block="10.0.0.0/16",
    enable_dns_hostnames=True,
    enable_dns_support=True,
    tags={
        "Name": "prod-k8s-vpc",
        f"kubernetes.io/cluster/{cluster_name}": "shared",
    },
)

# ========================================
# Internet Gateway
# ========================================

igw = aws.ec2.InternetGateway(
    "prod-igw",
    vpc_id=vpc.id,
    tags={"Name": "prod-k8s-igw"},
)

# ========================================
# Get Availability Zones
# ========================================

azs = aws.get_availability_zones(state="available")

# ========================================
# Create Subnets
# ========================================

public_subnets = []
private_subnets = []
nat_gateways = []
elastic_ips = []

for i, az in enumerate(azs.names[:3]):
    # Calculate CIDR blocks
    # Public:  10.0.{i*32}.0/20
    # Private: 10.0.{i*32+16}.0/20
    
    base = i * 32
    
    # -------------------------
    # Public Subnet
    # -------------------------
    public_subnet = aws.ec2.Subnet(
        f"public-subnet-{i}",
        vpc_id=vpc.id,
        cidr_block=f"10.0.{base}.0/20",
        availability_zone=az,
        map_public_ip_on_launch=True,
        tags={
            "Name": f"prod-public-{az}",
            "Type": "public",
            "kubernetes.io/role/elb": "1",  # For public load balancers
            f"kubernetes.io/cluster/{cluster_name}": "shared",
        },
    )
    public_subnets.append(public_subnet)
    
    # -------------------------
    # Private Subnet
    # -------------------------
    private_subnet = aws.ec2.Subnet(
        f"private-subnet-{i}",
        vpc_id=vpc.id,
        cidr_block=f"10.0.{base + 16}.0/20",
        availability_zone=az,
        map_public_ip_on_launch=False,
        tags={
            "Name": f"prod-private-{az}",
            "Type": "private",
            "kubernetes.io/role/internal-elb": "1",  # For internal LBs
            f"kubernetes.io/cluster/{cluster_name}": "shared",
        },
    )
    private_subnets.append(private_subnet)
    
    # -------------------------
    # Elastic IP for NAT Gateway
    # -------------------------
    eip = aws.ec2.Eip(
        f"nat-eip-{i}",
        vpc=True,
        tags={"Name": f"prod-nat-eip-{az}"},
    )
    elastic_ips.append(eip)
    
    # -------------------------
    # NAT Gateway (one per AZ for HA)
    # -------------------------
    nat = aws.ec2.NatGateway(
        f"nat-{i}",
        subnet_id=public_subnet.id,
        allocation_id=eip.id,
        tags={"Name": f"prod-nat-{az}"},
    )
    nat_gateways.append(nat)

# ========================================
# Route Tables
# ========================================

# Public Route Table (via Internet Gateway)
public_rt = aws.ec2.RouteTable(
    "public-rt",
    vpc_id=vpc.id,
    routes=[
        aws.ec2.RouteTableRouteArgs(
            cidr_block="0.0.0.0/0",
            gateway_id=igw.id,
        )
    ],
    tags={"Name": "prod-public-rt"},
)

# Associate public subnets with public route table
for i, subnet in enumerate(public_subnets):
    aws.ec2.RouteTableAssociation(
        f"public-rta-{i}",
        subnet_id=subnet.id,
        route_table_id=public_rt.id,
    )

# Private Route Tables (one per AZ, via respective NAT Gateway)
for i, (subnet, nat) in enumerate(zip(private_subnets, nat_gateways)):
    private_rt = aws.ec2.RouteTable(
        f"private-rt-{i}",
        vpc_id=vpc.id,
        routes=[
            aws.ec2.RouteTableRouteArgs(
                cidr_block="0.0.0.0/0",
                nat_gateway_id=nat.id,
            )
        ],
        tags={"Name": f"prod-private-rt-{azs.names[i]}"},
    )
    
    aws.ec2.RouteTableAssociation(
        f"private-rta-{i}",
        subnet_id=subnet.id,
        route_table_id=private_rt.id,
    )

# ========================================
# Bastion Host (Optional but Recommended)
# ========================================

# Bastion security group
bastion_sg = aws.ec2.SecurityGroup(
    "bastion-sg",
    vpc_id=vpc.id,
    description="Security group for bastion host",
    ingress=[
        aws.ec2.SecurityGroupIngressArgs(
            from_port=22,
            to_port=22,
            protocol="tcp",
            cidr_blocks=["YOUR_IP/32"],  # RESTRICT THIS!
            description="SSH from trusted IPs only",
        ),
    ],
    egress=[
        aws.ec2.SecurityGroupEgressArgs(
            from_port=0,
            to_port=0,
            protocol="-1",
            cidr_blocks=["0.0.0.0/0"],
        ),
    ],
    tags={"Name": "prod-bastion-sg"},
)

# Get latest Amazon Linux 2 AMI
bastion_ami = aws.ec2.get_ami(
    most_recent=True,
    owners=["amazon"],
    filters=[
        aws.ec2.GetAmiFilterArgs(
            name="name",
            values=["amzn2-ami-hvm-*-x86_64-gp2"],
        ),
    ],
)

# Bastion host
bastion = aws.ec2.Instance(
    "bastion",
    ami=bastion_ami.id,
    instance_type="t3.nano",  # Cheap!
    subnet_id=public_subnets[0].id,
    vpc_security_group_ids=[bastion_sg.id],
    key_name="YOUR_KEY_PAIR",  # CREATE THIS FIRST!
    tags={"Name": "prod-bastion"},
)

# ========================================
# Kubernetes Cluster
# ========================================

cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in private_subnets],  # PRIVATE subnets!
    kubernetes_version=kubernetes_version,
    cilium_version=cilium_version,
    control_plane_count=3,  # HA control plane
    worker_count=6,  # 2 per AZ
    control_plane_instance_type="t3.medium",
    worker_instance_type="t3.large",
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)

# Update cluster security group to allow SSH from bastion only
aws.ec2.SecurityGroupRule(
    "bastion-to-cluster",
    type="ingress",
    from_port=22,
    to_port=22,
    protocol="tcp",
    security_group_id=cluster.control_plane_sg.id,
    source_security_group_id=bastion_sg.id,
    description="SSH from bastion only",
)

# ========================================
# Outputs
# ========================================

pulumi.export("vpc_id", vpc.id)
pulumi.export("vpc_cidr", "10.0.0.0/16")
pulumi.export("public_subnet_ids", [s.id for s in public_subnets])
pulumi.export("private_subnet_ids", [s.id for s in private_subnets])
pulumi.export("nat_gateway_ips", [eip.public_ip for eip in elastic_ips])
pulumi.export("bastion_ip", bastion.public_ip)
pulumi.export("cluster_name", cluster_name)

# Connection instructions
pulumi.export("ssh_bastion", pulumi.Output.concat(
    "ssh -i YOUR_KEY.pem ec2-user@", bastion.public_ip
))

pulumi.export("ssh_control_plane_via_bastion", pulumi.Output.concat(
    "ssh -i YOUR_KEY.pem -J ec2-user@", bastion.public_ip,
    " ubuntu@", cluster.control_plane_instances[0].private_ip
))
```

---

## Security Hardening

### 1. Network Segmentation

**Private Subnets:**
```
✅ Nodes have NO public IPs
✅ Not directly accessible from internet
✅ Must go through NAT for outbound
✅ Inbound only via load balancers
```

**Security Group Updates:**

```python
# Restrict SSH to bastion only (add to platform.py)

# Remove this from control_plane_sg:
aws.ec2.SecurityGroupIngressArgs(
    from_port=22,
    to_port=22,
    protocol="tcp",
    cidr_blocks=["0.0.0.0/0"],  # ❌ Too permissive!
)

# Replace with:
# (SSH ingress rule added in example_usage.py from bastion SG)
```

### 2. API Server Access

**Option A: Private API (Most Secure)**
```python
# In platform.py, update control plane security group
# Remove public API access, only allow from VPC
aws.ec2.SecurityGroupIngressArgs(
    from_port=6443,
    to_port=6443,
    protocol="tcp",
    cidr_blocks=["10.0.0.0/16"],  # VPC only
)
```

**Option B: Restricted Public API**
```python
# Allow specific IPs only
aws.ec2.SecurityGroupIngressArgs(
    from_port=6443,
    to_port=6443,
    protocol="tcp",
    cidr_blocks=["YOUR_OFFICE_IP/32"],  # Whitelist
)
```

### 3. Network Policies

Enable Cilium network policies by default:

```yaml
# Add to Cilium installation in platform.py
--set policyEnforcementMode=default \
--set policyAuditMode=false
```

### 4. Encryption

**In Transit:**
```yaml
Cilium: transparentEncryption.enabled=true (WireGuard)
API Server: Always TLS
etcd: TLS
```

**At Rest:**
```yaml
EBS volumes: encrypted=true (already set)
etcd: encryption-config
```

---

## High Availability

### NAT Gateway HA

**Why One NAT Per AZ:**
```
Scenario: NAT in AZ-1 fails
├─ AZ-1 nodes: Lose internet ❌
├─ AZ-2 nodes: Still working ✅
└─ AZ-3 nodes: Still working ✅

Cluster stays up! 66% capacity maintained.
```

**vs Single NAT:**
```
Scenario: Single NAT fails
├─ AZ-1 nodes: Lose internet ❌
├─ AZ-2 nodes: Lose internet ❌
└─ AZ-3 nodes: Lose internet ❌

Cluster down! ❌
```

### Control Plane HA

```python
control_plane_count=3  # Across 3 AZs

# etcd quorum:
# 3 nodes = can tolerate 1 failure
# 5 nodes = can tolerate 2 failures
```

### Worker Node Distribution

```python
# Ensure workers spread across AZs
worker_count=6  # 2 per AZ minimum

# Or use Auto Scaling Groups (future enhancement)
```

---

## Cost Analysis

### Production Network Costs

**NAT Gateways (Biggest Cost):**
```
3 NAT Gateways:
├─ Fixed: $0.045/hour × 3 = $0.135/hour
│  = $98.55/month
│
└─ Data processing: $0.045/GB
   ├─ 10TB/month = $461.25
   ├─ 20TB/month = $922.50
   └─ Estimate: ~$500/month

Total NAT cost: ~$600/month
```

**Other Resources:**
```
Subnets: Free
Route tables: Free
Internet Gateway: Free
Elastic IPs (attached): Free
Bastion (t3.nano): $3.80/month
Security Groups: Free

Total infrastructure: ~$604/month
```

### Cost Optimization Strategies

#### Strategy 1: Single NAT (Dev/Staging)

```python
# Use only 1 NAT Gateway (in AZ-1)
# Route all private subnets through it

Savings: ~$65/month (2 NAT Gateways)
Risk: Single point of failure
Use for: Dev/Staging only
```

#### Strategy 2: VPC Endpoints

```python
# For AWS services, use VPC endpoints instead of NAT
# S3, ECR, STS, etc.

vpc_endpoint_s3 = aws.ec2.VpcEndpoint(
    "s3-endpoint",
    vpc_id=vpc.id,
    service_name="com.amazonaws.us-west-2.s3",
    route_table_ids=[private_rt.id for private_rt in private_rts],
)

# Reduces NAT data transfer costs significantly!
```

#### Strategy 3: Reserved NAT Capacity

```
Reserved Commitment (1 year):
- Save up to 36% on NAT Gateway hours
- Good for production (always running)
```

---

## Migration from Dev

### Step-by-Step Migration

#### Phase 1: Create Production Network

```bash
# 1. Create new directory
cp -r self-managed-k8s self-managed-k8s-prod
cd self-managed-k8s-prod

# 2. Use production example_usage.py (from above)
# Replace example_usage.py with production version

# 3. Configure
pulumi stack init production
pulumi config set cluster_name prod-k8s

# 4. Deploy network only (comment out cluster)
pulumi up
```

#### Phase 2: Deploy Production Cluster

```bash
# 5. Uncomment cluster in example_usage.py
pulumi up

# 6. Initialize control plane
ssh-add YOUR_KEY.pem
ssh -J ec2-user@BASTION_IP ubuntu@CP_PRIVATE_IP
sudo /root/init-cluster.sh

# 7. Join workers (via bastion)
for WORKER_IP in WORKER_IPS; do
  ssh -J ec2-user@BASTION_IP ubuntu@$WORKER_IP "sudo $JOIN_CMD"
done
```

#### Phase 3: Migrate Workloads

```bash
# 8. Download kubeconfig via bastion
ssh -J ec2-user@BASTION_IP ubuntu@CP_IP "sudo cat /etc/kubernetes/admin.conf" > ~/.kube/config-prod

# 9. Deploy workloads
kubectl --kubeconfig ~/.kube/config-prod apply -f apps/

# 10. Test thoroughly

# 11. Update DNS to point to new cluster

# 12. Decommission dev cluster
cd ../self-managed-k8s
pulumi destroy
```

---

## Monitoring & Troubleshooting

### Network Connectivity Tests

#### Test 1: Internet Access from Private Subnet

```bash
# SSH to node via bastion
ssh -J ec2-user@BASTION ubuntu@NODE_PRIVATE_IP

# Test internet
curl -I https://google.com
# Should work via NAT Gateway ✅

# Check route
ip route get 8.8.8.8
# Should show NAT Gateway IP
```

#### Test 2: Cross-AZ Communication

```bash
# From node in AZ-1
ping NODE_IN_AZ2_PRIVATE_IP
# Should work (via VPC routing) ✅
```

#### Test 3: Pod-to-Pod

```bash
# Create test pods
kubectl run test-1 --image=busybox --command -- sleep 3600
kubectl run test-2 --image=busybox --command -- sleep 3600

# Get IPs
POD1_IP=$(kubectl get pod test-1 -o jsonpath='{.status.podIP}')
POD2_IP=$(kubectl get pod test-2 -o jsonpath='{.status.podIP}')

# Test
kubectl exec test-1 -- ping -c 3 $POD2_IP
# Should work via Cilium ✅
```

### Common Issues

#### Issue 1: NAT Gateway Connectivity

**Symptom:**
```
Pods can't reach internet
Image pulls fail
```

**Debug:**
```bash
# Check NAT Gateway status
aws ec2 describe-nat-gateways

# Check route table
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=VPC_ID"

# Verify private subnet has route to NAT
```

**Fix:**
```
Ensure route table has:
Destination: 0.0.0.0/0
Target: nat-xxxxxxxx
```

#### Issue 2: Bastion Access

**Symptom:**
```
Can't SSH to bastion
```

**Debug:**
```bash
# Check security group
aws ec2 describe-security-groups --group-ids BASTION_SG_ID

# Verify your IP is whitelisted
```

**Fix:**
```python
# Update bastion security group with your current IP
cidr_blocks=["YOUR_NEW_IP/32"]
```

#### Issue 3: LoadBalancer Not Working

**Symptom:**
```
LoadBalancer service stuck in Pending
```

**Debug:**
```bash
kubectl describe svc YOUR_SERVICE

# Check events
kubectl get events -n NAMESPACE
```

**Fix:**
```
Ensure subnets have correct tags:
- Public subnets: kubernetes.io/role/elb: "1"
- Private subnets: kubernetes.io/role/internal-elb: "1"
```

---

## Best Practices Summary

### ✅ Do's

1. **Use private subnets** for all Kubernetes nodes
2. **One NAT Gateway per AZ** for high availability
3. **Generous subnet sizing** (/20 or larger)
4. **Proper subnet tagging** for Kubernetes discovery
5. **Bastion host** for secure access
6. **Security group hardening** (principle of least privilege)
7. **Enable VPC Flow Logs** for auditing
8. **Use VPC endpoints** to reduce NAT costs
9. **Monitor NAT Gateway** metrics (bytes, connections)
10. **Regular security audits**

### ❌ Don'ts

1. **Don't use /24 or smaller** subnets (too small)
2. **Don't skip NAT Gateway redundancy** (single point of failure)
3. **Don't allow SSH from 0.0.0.0/0** (security risk)
4. **Don't run nodes in public subnets** (production)
5. **Don't forget subnet tags** (breaks Kubernetes integration)
6. **Don't ignore costs** (monitor NAT data transfer)
7. **Don't skip bastion** (direct node access is risky)
8. **Don't mix workload types** in same subnet (separation needed)

---

## Production Checklist

Before going to production, ensure:

- [ ] **3 Availability Zones** configured
- [ ] **Public + Private subnets** per AZ
- [ ] **NAT Gateway per AZ** (HA)
- [ ] **/20 subnets** (4,096 IPs each)
- [ ] **Proper subnet tags** for Kubernetes
- [ ] **Bastion host** deployed
- [ ] **SSH restricted** to bastion only
- [ ] **Security groups** hardened
- [ ] **Control plane HA** (3 nodes)
- [ ] **VPC Flow Logs** enabled
- [ ] **CloudWatch alarms** for NAT Gateway
- [ ] **Backup strategy** for etcd
- [ ] **Disaster recovery** plan documented
- [ ] **Cost monitoring** configured
- [ ] **Network policies** enabled (Cilium)

---

## Conclusion

This production network architecture provides:

✅ **Security**: Defense in depth with private subnets
✅ **High Availability**: NAT Gateway per AZ, multi-AZ nodes
✅ **Scalability**: /20 subnets support 800+ nodes per AZ
✅ **Cost-Effective**: Right-sized for production workloads
✅ **Best Practices**: Follows AWS and Kubernetes guidelines

**Estimated cost:** ~$600/month for network infrastructure (3 NAT Gateways)
**Comparison:** This is still **90% cheaper** than AKS and provides **better control**!

Deploy with confidence! 🚀
