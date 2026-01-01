# AWS VPC Networking: Subnets, Security Groups, and Communication

## TL;DR - The Misconception

**Common Belief (WRONG ❌):**
```
"Different subnets = isolated networks"
"Nodes in different subnets can't talk to each other"
```

**Reality (CORRECT ✅):**
```
"Subnets in the same VPC can freely communicate"
"Security Groups control what traffic is allowed"
"Subnets are NOT isolated by default"
```

---

## Table of Contents

1. [The Confusion Explained](#the-confusion-explained)
2. [VPC Networking Fundamentals](#vpc-networking-fundamentals)
3. [How Nodes Communicate](#how-nodes-communicate)
4. [Security Groups vs Subnets](#security-groups-vs-subnets)
5. [What Actually Controls Isolation](#what-actually-controls-isolation)
6. [Real-World Examples](#real-world-examples)
7. [When Subnets ARE Isolated](#when-subnets-are-isolated)
8. [Best Practices](#best-practices)

---

## The Confusion Explained

### What You Might Think

```
┌─────────────────────────────────────┐
│ VPC: 10.0.0.0/16                     │
│                                      │
│  ┌──────────────┐  ┌──────────────┐ │
│  │ Subnet 1     │  │ Subnet 2     │ │
│  │ 10.0.0.0/20  │  │ 10.0.16.0/20 │ │
│  │              │  │              │ │
│  │  Node A      │  │  Node B      │ │
│  └──────────────┘  └──────────────┘ │
│         ❌              ❌            │
│    Can't communicate (WRONG!)       │
└─────────────────────────────────────┘
```

### What Actually Happens

```
┌─────────────────────────────────────┐
│ VPC: 10.0.0.0/16                     │
│         VPC Router (implicit)        │
│              ↕                       │
│  ┌──────────────┐  ┌──────────────┐ │
│  │ Subnet 1     │  │ Subnet 2     │ │
│  │ 10.0.0.0/20  │  │ 10.0.16.0/20 │ │
│  │              │←→│              │ │
│  │  Node A      │  │  Node B      │ │
│  └──────────────┘  └──────────────┘ │
│         ✅              ✅            │
│    Can communicate freely!          │
└─────────────────────────────────────┘
```

---

## VPC Networking Fundamentals

### What is a VPC?

A VPC (Virtual Private Cloud) is **one large network** with:
- A single CIDR block (e.g., 10.0.0.0/16)
- An implicit router that connects all subnets
- Route tables that control routing
- Security groups that control traffic

**Key Concept:** All subnets in a VPC are part of the **same network**.

---

### What is a Subnet?

A subnet is **a subdivision of a VPC** for:
- Logical organization
- Availability zone placement
- Route table assignment
- Network ACL assignment

**Key Concept:** Subnets are **NOT isolated networks**. They're just **segments of the same VPC**.

---

### The VPC Router (Implicit)

Every VPC has an **implicit router** that:
- Connects all subnets automatically
- Routes traffic between subnets
- Cannot be seen or configured directly
- Always allows intra-VPC traffic (unless blocked by security groups/NACLs)

```
VPC: 10.0.0.0/16
├─ VPC Router (automatic, invisible)
│  ├─ Routes to Subnet 1 (10.0.0.0/20)
│  ├─ Routes to Subnet 2 (10.0.16.0/20)
│  └─ Routes to Subnet 3 (10.0.32.0/20)
```

**This is why nodes in different subnets can talk to each other!**

---

## How Nodes Communicate

### Scenario: Node in Subnet A → Node in Subnet B

```
┌──────────────────────────────────────────────────────┐
│ VPC: 10.0.0.0/16                                      │
│                                                       │
│  ┌─────────────────┐              ┌─────────────────┐│
│  │ Subnet A        │              │ Subnet B        ││
│  │ 10.0.0.0/20     │              │ 10.0.16.0/20    ││
│  │                 │              │                 ││
│  │  Node A         │              │  Node B         ││
│  │  10.0.5.100 ────┼──────────────┼───→ 10.0.20.50  ││
│  └─────────────────┘              └─────────────────┘│
│                                                       │
│  Step-by-step:                                       │
│  1. Node A sends packet to 10.0.20.50                │
│  2. Packet hits VPC router                           │
│  3. Router checks route table                        │
│  4. Route table says: "10.0.0.0/16 → local"          │
│  5. Router forwards to Subnet B                      │
│  6. Packet arrives at Node B                         │
│  7. Node B's security group checks: Allow?           │
│  8. If allowed, packet delivered ✅                   │
└──────────────────────────────────────────────────────┘
```

### The Route Table Entry

**Every route table in a VPC has this entry by default:**

```
Destination: 10.0.0.0/16 (VPC CIDR)
Target: local

This means: "All traffic within the VPC stays within the VPC"
```

**This is the magic that allows cross-subnet communication!**

### What Zones Actually Separate

**Subnets in different AZs can still communicate!**

```
┌────────────────────────────────────────────────────┐
│ VPC: 10.0.0.0/16                                    │
│                                                     │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────┐ │
│  │ Subnet 1     │  │ Subnet 2     │  │ Subnet 3 │ │
│  │ us-west-2a   │  │ us-west-2b   │  │ us-west-2c│ │
│  │ 10.0.0.0/20  │  │ 10.0.16.0/20 │  │10.0.32.0/20│
│  │              │  │              │  │          │ │
│  │  Node A ─────┼──┼─→ Node B ────┼──┼─→ Node C │ │
│  └──────────────┘  └──────────────┘  └──────────┘ │
│                                                     │
│  All can communicate via VPC router ✅              │
└────────────────────────────────────────────────────┘
```

**Availability Zones provide fault isolation, NOT network isolation.**

---

## Security Groups vs Subnets

### What Security Groups Control

Security Groups are **virtual firewalls** that control:
- ✅ What IPs/ports can connect to an instance
- ✅ What protocols are allowed
- ✅ Which security groups can communicate
- ✅ Inbound and outbound traffic rules

**Security Groups operate at the instance/ENI level.**

### What Subnets Control

Subnets control:
- ✅ Which AZ an instance is in
- ✅ Which route table is used
- ✅ Which Network ACL applies
- ❌ NOT whether instances can communicate!

**Subnets are organizational, not security boundaries.**

---

### Example: Our Kubernetes Cluster

```python
# From platform.py
# Worker security group allows worker-to-worker communication

aws.ec2.SecurityGroupRule(
    f"{name}-worker-to-worker",
    type="ingress",
    from_port=0,
    to_port=65535,
    protocol="-1",
    security_group_id=self.worker_sg.id,
    source_security_group_id=self.worker_sg.id,  # ← Key!
    description="Allow worker to worker",
)
```

**What this means:**
- ✅ Any instance with `worker_sg` can talk to any other instance with `worker_sg`
- ✅ This works **across all subnets** in the VPC
- ✅ Subnet doesn't matter, security group does!

---

## What Actually Controls Isolation

### Layer 1: Route Tables

**Route tables determine where traffic goes.**

```
Example route table for private subnet:
┌──────────────────┬─────────────────┐
│ Destination      │ Target          │
├──────────────────┼─────────────────┤
│ 10.0.0.0/16      │ local           │ ← Intra-VPC traffic
│ 0.0.0.0/0        │ nat-xxxxxx      │ ← Internet via NAT
└──────────────────┴─────────────────┘
```

**The "local" route allows all subnets to communicate.**

To block inter-subnet communication, you'd have to:
1. Remove the "local" route (impossible, it's required!)
2. Use Network ACLs (see below)

---

### Layer 2: Security Groups (Primary Control)

**Security Groups are the main isolation mechanism.**

Example: Isolate database from workers

```python
# Database security group
db_sg = aws.ec2.SecurityGroup(
    "db-sg",
    ingress=[
        # Only allow from app_sg
        aws.ec2.SecurityGroupIngressArgs(
            from_port=5432,
            to_port=5432,
            protocol="tcp",
            source_security_group_id=app_sg.id,  # Only app tier!
            description="PostgreSQL from app tier only",
        ),
    ],
)

# Database could be in ANY subnet
# App could be in ANY subnet
# Security group controls access, not subnet!
```

---

### Layer 3: Network ACLs (Rarely Used)

**Network ACLs are stateless firewalls at the subnet level.**

```python
# Subnet-level isolation (uncommon, but possible)
nacl = aws.ec2.NetworkAcl(
    "restricted-nacl",
    vpc_id=vpc.id,
)

# Deny traffic from another subnet
aws.ec2.NetworkAclRule(
    "deny-subnet-2",
    network_acl_id=nacl.id,
    rule_number=100,
    protocol="-1",
    rule_action="deny",
    cidr_block="10.0.16.0/20",  # Block Subnet 2
)
```

**Why NACLs are rarely used:**
- ❌ Stateless (must allow both request and response)
- ❌ Complex to manage
- ❌ Easy to misconfigure
- ✅ Security Groups are better for most use cases

---

## Real-World Examples

### Example 1: Kubernetes Nodes Across Subnets

**Setup:**
```
Subnet 1 (AZ-1): 10.0.16.0/20
├─ Control Plane: 10.0.16.10
├─ Worker 1: 10.0.16.20
└─ Worker 2: 10.0.16.30

Subnet 2 (AZ-2): 10.0.48.0/20
├─ Worker 3: 10.0.48.20
└─ Worker 4: 10.0.48.30

Subnet 3 (AZ-3): 10.0.80.0/20
├─ Worker 5: 10.0.80.20
└─ Worker 6: 10.0.80.30
```

**Communication:**
```bash
# From Worker 1 (Subnet 1) to Worker 5 (Subnet 3)
ping 10.0.80.20

# Step 1: Packet leaves Worker 1
# Step 2: Hits VPC router
# Step 3: Router checks: "10.0.80.20 is in 10.0.0.0/16 (VPC CIDR)"
# Step 4: Router forwards to Subnet 3
# Step 5: Worker 5's security group checks
# Step 6: Security group allows (worker-to-worker rule)
# Step 7: Ping succeeds! ✅
```

**Why it works:**
- ✅ Both in same VPC
- ✅ Route table has "local" route
- ✅ Security group allows worker-to-worker
- ✅ No Network ACL blocking

---

### Example 2: Cilium Overlay Network

**Even with Cilium, nodes still use VPC routing!**

```
Pod on Worker 1 (10.32.5.100) → Pod on Worker 5 (10.32.10.50)

Step 1: Cilium on Worker 1 encapsulates pod packet
Step 2: Outer packet: 10.0.16.20 → 10.0.80.20 (node IPs)
Step 3: VPC router forwards based on node IPs
Step 4: Packet arrives at Worker 5 (10.0.80.20)
Step 5: Cilium decapsulates and delivers to pod
```

**Both layers use VPC routing!**
- Node-to-node: VPC routes between subnets
- Pod-to-pod: Cilium overlay on top

---

## When Subnets ARE Isolated

### Scenario 1: Different VPCs

```
VPC-1: 10.0.0.0/16
└─ Subnet: 10.0.0.0/20

VPC-2: 10.1.0.0/16
└─ Subnet: 10.1.0.0/20

❌ Cannot communicate without VPC peering or Transit Gateway
```

**Different VPCs are truly isolated.**

---

### Scenario 2: Network ACLs Block Traffic

```python
# Network ACL that blocks specific subnet
nacl = aws.ec2.NetworkAcl("restrictive-nacl")

aws.ec2.NetworkAclRule(
    "deny-subnet-2",
    network_acl_id=nacl.id,
    rule_number=100,
    protocol="-1",
    rule_action="deny",
    cidr_block="10.0.16.0/20",
)

# Associate with Subnet 1
aws.ec2.NetworkAclAssociation(
    "nacl-assoc",
    network_acl_id=nacl.id,
    subnet_id=subnet1.id,
)
```

**Now Subnet 1 cannot communicate with Subnet 2.**

But this is **very rare** in practice!

---

## Best Practices

### 1. Use Security Groups for Isolation ✅

**Good:**
```python
# App tier security group
app_sg = aws.ec2.SecurityGroup(
    "app-sg",
    ingress=[
        # Only from load balancer
        aws.ec2.SecurityGroupIngressArgs(
            source_security_group_id=lb_sg.id,
        ),
    ],
)

# Database security group
db_sg = aws.ec2.SecurityGroup(
    "db-sg",
    ingress=[
        # Only from app tier
        aws.ec2.SecurityGroupIngressArgs(
            from_port=5432,
            to_port=5432,
            protocol="tcp",
            source_security_group_id=app_sg.id,
        ),
    ],
)
```

**Benefits:**
- ✅ Clear security boundaries
- ✅ Works across subnets
- ✅ Easy to understand
- ✅ Stateful (easier to manage)

---

### 2. Use Subnets for Organization ✅

**Good:**
```
Public Subnets (per AZ):
├─ NAT Gateways
├─ Load Balancers
└─ Bastion hosts

Private Subnets (per AZ):
├─ Kubernetes nodes
├─ Databases
└─ Application servers
```

**Benefits:**
- ✅ Logical separation
- ✅ Different route tables
- ✅ Fault isolation (multi-AZ)

---

### 3. Avoid Network ACLs ✅

**Unless you have a specific reason:**
- Security Groups are better
- NACLs are stateless (complex)
- Hard to troubleshoot

**When to use NACLs:**
- Compliance requirement for subnet-level filtering
- Block specific IP ranges
- Extra layer of defense

---

### 4. Design for Multi-AZ ✅

**Good:**
```python
# Distribute nodes across subnets/AZs
for i in range(worker_count):
    instance = aws.ec2.Instance(
        f"worker-{i}",
        subnet_id=subnet_ids[i % len(subnet_ids)],  # Round-robin
        # ... other config
    )
```

**Benefits:**
- ✅ High availability
- ✅ Fault tolerance
- ✅ Even distribution

---

## Testing Communication

### Test 1: Cross-Subnet Ping

```bash
# From node in Subnet 1
NODE1_IP=10.0.16.20

# To node in Subnet 2
NODE2_IP=10.0.48.30

# SSH to Node 1
ssh ubuntu@$NODE1_IP

# Ping Node 2
ping $NODE2_IP
# Should work! ✅

# Why it works:
# 1. Same VPC (10.0.0.0/16)
# 2. Route table has "local" route
# 3. Security group allows (worker-to-worker)
```

### Test 2: View Route Table

```bash
# On any node
ip route show

# Output:
# default via 10.0.16.1 dev eth0  ← Default gateway (VPC router)
# 10.0.16.0/20 dev eth0 scope link  ← Local subnet
# 10.0.0.0/16 via 10.0.16.1 dev eth0  ← Other subnets via VPC router

# The last route shows all VPC traffic goes through VPC router
```

### Test 3: Trace Route

```bash
# From Node 1 to Node 2 (different subnet)
traceroute 10.0.48.30

# Output:
# 1  10.0.16.1  ← VPC router (default gateway)
# 2  10.0.48.30  ← Destination

# Only 1 hop! VPC router directly connects subnets.
```

---

## Common Misconceptions Debunked

### Myth 1: "Different subnets can't talk"

**Reality:**
- ✅ All subnets in a VPC can communicate
- ✅ VPC router connects them automatically
- ✅ Only blocked by security groups or NACLs

---

### Myth 2: "Subnets provide security isolation"

**Reality:**
- ❌ Subnets do NOT provide security isolation
- ✅ Security Groups provide isolation
- ✅ Use security groups, not separate subnets, for security

---

### Myth 3: "AZ placement isolates networks"

**Reality:**
- ❌ AZs provide fault isolation, NOT network isolation
- ✅ Subnets in different AZs can still communicate
- ✅ AZ placement is for availability, not security

---

### Myth 4: "Private subnets can't talk to public subnets"

**Reality:**
- ✅ Private and public subnets can communicate freely
- ✅ "Public" just means route to Internet Gateway
- ✅ "Private" just means route to NAT Gateway
- ✅ Both can talk to each other within VPC

---

## Summary

### Key Takeaways

1. **Subnets in a VPC are NOT isolated by default**
   - They're connected by the VPC router
   - All have a "local" route to each other

2. **Security Groups control what can communicate**
   - Not subnets
   - Use security groups for security boundaries

3. **Subnets are for organization and routing**
   - Logical separation
   - AZ placement
   - Route table assignment

4. **Our Kubernetes cluster works because:**
   - All nodes in same VPC ✅
   - Security group allows worker-to-worker ✅
   - Route table has "local" route ✅
   - No Network ACLs blocking ✅

5. **To actually isolate networks:**
   - Use different VPCs (real isolation)
   - Or use Security Groups (preferred)
   - Or use Network ACLs (if you must)

---

## Practical Answer to Your Question

**"Can nodes in different subnets talk to each other?"**

**YES! ✅** Because:

1. **They're in the same VPC** (10.0.0.0/16)
2. **VPC router connects all subnets** automatically
3. **Route table has "local" route** for all VPC traffic
4. **Security group allows worker-to-worker** communication

**"Is this controlled by Security Group?"**

**YES! ✅** Security Groups are the **primary control mechanism**, not subnets.

**"I thought different subnets are isolated?"**

**Common misconception!** ❌ 
- Different **VPCs** are isolated
- Different **subnets in the same VPC** are NOT isolated
- Use **Security Groups** for isolation, not subnets

**"What makes nodes in the same subnet talk to each other?"**

**Same as cross-subnet:**
1. VPC router (same mechanism)
2. Security Groups allow it
3. No difference between same-subnet and cross-subnet!

---

## Visual Summary

```
┌──────────────────────────────────────────────────────┐
│ What Actually Controls Communication                  │
├──────────────────────────────────────────────────────┤
│                                                       │
│  ✅ Security Groups (primary control)                 │
│     ├─ Defined rules for ingress/egress              │
│     └─ Instance/ENI level                            │
│                                                       │
│  ✅ Route Tables                                      │
│     ├─ "local" route allows intra-VPC                │
│     └─ Subnet level                                  │
│                                                       │
│  ⚠️  Network ACLs (rarely used)                       │
│     ├─ Stateless subnet-level filtering             │
│     └─ Use only if required                          │
│                                                       │
│  ❌ Subnets (NOT a control mechanism)                │
│     ├─ Just organizational/routing                   │
│     └─ No isolation by default                       │
└──────────────────────────────────────────────────────┘
```

**Remember: In AWS, Security Groups are your friend!** 🛡️
