# Troubleshooting: SSH Access to EC2 Instances

## Problem Description

After deploying the self-managed Kubernetes cluster with Pulumi, attempting to SSH into the EC2 instances failed with:

```bash
$ ssh ubuntu@18.144.170.120
ubuntu@18.144.170.120: Permission denied (publickey).
```

## Root Cause Analysis

### Initial Investigation

1. **Check EC2 Instance Details**
   ```bash
   aws ec2 describe-instances --region us-west-1 \
     --filters "Name=tag:Cluster,Values=dev-k8s" \
     --query 'Reservations[*].Instances[*].[InstanceId,KeyName]'
   ```
   
   **Result:** `KeyName` was `null` - no SSH key pair was associated with the instances.

2. **Check Available SSH Keys**
   ```bash
   aws ec2 describe-key-pairs --region us-west-1
   ```
   
   **Result:** No key pairs existed in the us-west-1 region.

3. **Review Code**
   
   Examined [`k8s_self_host.py`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/k8s_self_host.py):
   
   ```python
   instance = aws.ec2.Instance(
       f"{name}-cp-{i}",
       instance_type=control_plane_instance_type,
       ami=ubuntu_ami.id,
       subnet_id=subnet_ids[i % len(subnet_ids)],
       # ❌ MISSING: key_name parameter
       vpc_security_group_ids=[self.control_plane_sg.id],
       ...
   )
   ```

### Root Cause

**The EC2 instances were created without an associated SSH key pair**, making SSH authentication impossible. By default, Ubuntu EC2 instances only allow public key authentication, and without a key pair configured, there's no way to authenticate.

## Solution

### Step 1: Create SSH Key Pair in AWS

```bash
# Create key pair and save private key locally
aws ec2 create-key-pair \
  --key-name k8s-dev-key \
  --region us-west-1 \
  --query 'KeyMaterial' \
  --output text > ~/.ssh/k8s-dev-key.pem

# Set correct permissions (required for SSH)
chmod 400 ~/.ssh/k8s-dev-key.pem
```

**Why this step:** AWS key pairs must be created in AWS first, then the private key is saved locally for SSH client use.

### Step 2: Add SSH Key Parameter to Code

Modified [`k8s_self_host.py`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/k8s_self_host.py) to support SSH keys:

**A. Updated Class Constructor**

```python
def __init__(
    self,
    name: str,
    vpc_id: pulumi.Input[str],
    subnet_ids: pulumi.Input[List[str]],
    ssh_key_name: Optional[str] = None,  # ✅ NEW: SSH key parameter
    kubernetes_version: str = "1.29.0",
    cilium_version: str = "1.15.1",
    ...
):
    super().__init__("custom:SelfManagedK8sCluster", name, {}, opts)
    
    self.name = name
    self.ssh_key_name = ssh_key_name  # ✅ Store SSH key name
    ...
```

**B. Updated Control Plane Instance Creation**

```python
instance = aws.ec2.Instance(
    f"{name}-cp-{i}",
    instance_type=control_plane_instance_type,
    ami=ubuntu_ami.id,
    key_name=self.ssh_key_name,  # ✅ NEW: Add SSH key
    subnet_id=subnet_ids[i % len(subnet_ids)],
    vpc_security_group_ids=[self.control_plane_sg.id],
    ...
)
```

**C. Updated Worker Instance Creation**

```python
instance = aws.ec2.Instance(
    f"{name}-worker-{i}",
    instance_type=worker_instance_type,
    ami=ubuntu_ami.id,
    key_name=self.ssh_key_name,  # ✅ NEW: Add SSH key
    subnet_id=subnet_ids[i % len(subnet_ids)],
    vpc_security_group_ids=[self.worker_sg.id],
    ...
)
```

### Step 3: Update Usage Code

Modified [`example_usage.py`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/example_usage.py):

```python
cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in subnets],
    ssh_key_name="k8s-dev-key",  # ✅ NEW: Specify SSH key
    kubernetes_version=kubernetes_version,
    cilium_version=cilium_version,
    control_plane_count=1,
    worker_count=3,
    control_plane_instance_type="t3.medium",
    worker_instance_type="t3.medium",
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)
```

### Step 4: Apply Infrastructure Changes

```bash
# Preview changes
pulumi preview

# Expected output:
# +-  dev-k8s-cp-0      replace  [diff: ~key_name]
# +-  dev-k8s-worker-0  replace  [diff: ~key_name]
# +-  dev-k8s-worker-1  replace  [diff: ~key_name]
# +-  dev-k8s-worker-2  replace  [diff: ~key_name]

# Apply changes
pulumi up --yes
```

**Note:** Changing the SSH key requires **replacing** the instances because `key_name` is an immutable property. Pulumi will:
1. Create new instances with the SSH key
2. Delete the old instances
3. Update all outputs with new IP addresses

### Step 5: Verify SSH Access

```bash
# Get new control plane IP
CONTROL_PLANE_IP=$(pulumi stack output control_plane_public_ips)

# Test SSH connection
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@$CONTROL_PLANE_IP "echo 'SSH connection successful!'"
```

**Expected output:**
```
Warning: Permanently added '13.57.3.190' (ED25519) to the list of known hosts.
SSH connection successful!
```

## Verification Checklist

- [x] SSH key pair exists in AWS
- [x] Private key saved locally with correct permissions (400)
- [x] Code updated with `ssh_key_name` parameter
- [x] All instances recreated with SSH key
- [x] SSH connection successful to control plane
- [x] SSH connection successful to workers

## Additional Enhancements

### 1. Created SSH Helper Script

[`ssh-to-cluster.sh`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/ssh-to-cluster.sh) - Interactive menu to connect to any node:

```bash
./ssh-to-cluster.sh

# Output:
# === Kubernetes Cluster SSH Helper ===
# 
# 1) Control Plane: 13.57.3.190
# 
# Workers:
#   2) Worker 0: 13.52.76.199
#   3) Worker 1: 18.144.48.134
#   4) Worker 2: 18.144.28.109
# 
# Select node (1-4):
```

### 2. Created SSH Access Documentation

[`SSH_ACCESS.md`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/SSH_ACCESS.md) - Complete reference for SSH operations

## Common Issues and Solutions

### Issue 1: Permission Denied (Still)

**Symptoms:**
```bash
$ ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@13.57.3.190
Permission denied (publickey).
```

**Solutions:**

1. **Check key permissions:**
   ```bash
   ls -la ~/.ssh/k8s-dev-key.pem
   # Should show: -r-------- (400)
   
   chmod 400 ~/.ssh/k8s-dev-key.pem
   ```

2. **Verify correct key being used:**
   ```bash
   aws ec2 describe-instances \
     --region us-west-1 \
     --instance-ids $(pulumi stack output control_plane_instance_id) \
     --query 'Reservations[0].Instances[0].KeyName'
   ```

3. **Check SSH agent conflicts:**
   ```bash
   # Use -o flag to specify key explicitly
   ssh -o IdentitiesOnly=yes -i ~/.ssh/k8s-dev-key.pem ubuntu@13.57.3.190
   ```

### Issue 2: Key Pair Already Exists

**Symptoms:**
```bash
$ aws ec2 create-key-pair --key-name k8s-dev-key ...
An error occurred (InvalidKeyPair.Duplicate): The key pair 'k8s-dev-key' already exists.
```

**Solutions:**

**Option A: Use existing key** (if you have the private key)
```bash
# Just update the code to use the existing key name
# No need to create a new one
```

**Option B: Delete and recreate**
```bash
# Delete existing key pair
aws ec2 delete-key-pair --key-name k8s-dev-key --region us-west-1

# Create new one
aws ec2 create-key-pair \
  --key-name k8s-dev-key \
  --region us-west-1 \
  --query 'KeyMaterial' \
  --output text > ~/.ssh/k8s-dev-key.pem

chmod 400 ~/.ssh/k8s-dev-key.pem

# Recreate instances
pulumi up
```

### Issue 3: Security Group Blocks SSH

**Symptoms:**
```bash
$ ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@13.57.3.190
# Connection times out (no response)
```

**Solutions:**

1. **Check security group rules:**
   ```bash
   aws ec2 describe-security-groups \
     --region us-west-1 \
     --filters "Name=tag:Name,Values=dev-k8s-cp-sg" \
     --query 'SecurityGroups[0].IpPermissions[?FromPort==`22`]'
   ```

2. **Verify your public IP is allowed:**
   ```bash
   curl ifconfig.me
   # Compare with allowed CIDR blocks in security group
   ```

3. **Update security group if needed:**
   
   The code already allows SSH from `0.0.0.0/0`. For production, restrict to your IP:
   
   ```python
   aws.ec2.SecurityGroupIngressArgs(
       from_port=22,
       to_port=22,
       protocol="tcp",
       cidr_blocks=["YOUR_IP/32"],  # Replace with your IP
       description="SSH",
   ),
   ```

### Issue 4: Wrong Username

**Symptoms:**
```bash
$ ssh -i ~/.ssh/k8s-dev-key.pem admin@13.57.3.190
Permission denied (publickey).
```

**Solution:**

Ubuntu AMIs use the `ubuntu` user, not `admin` or `ec2-user`:
```bash
ssh -i ~/.ssh/k8s-dev-key.pem ubuntu@13.57.3.190
```

## Prevention for Future Deployments

### 1. Make SSH Key Required

Update the constructor to make it non-optional:

```python
def __init__(
    self,
    name: str,
    vpc_id: pulumi.Input[str],
    subnet_ids: pulumi.Input[List[str]],
    ssh_key_name: str,  # ✅ Required parameter (no default)
    ...
):
```

### 2. Add Validation

```python
def __init__(self, ...):
    if not ssh_key_name:
        raise ValueError("ssh_key_name is required for SSH access to instances")
    
    super().__init__("custom:SelfManagedK8sCluster", name, {}, opts)
```

### 3. Document in README

Add SSH key setup to deployment prerequisites:

```markdown
## Prerequisites

1. AWS CLI configured
2. Pulumi installed
3. **SSH key pair created in target region:**
   ```bash
   aws ec2 create-key-pair --key-name k8s-dev-key ...
   ```
```

### 4. Use Pulumi Config

Make the key name configurable:

```python
# In example_usage.py
config = pulumi.Config()
ssh_key_name = config.require("ssh_key_name")  # Force user to set it

cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in subnets],
    ssh_key_name=ssh_key_name,
    ...
)
```

Then in deployment:
```bash
pulumi config set ssh_key_name k8s-dev-key
pulumi up
```

## Lessons Learned

1. **Always include SSH access in infrastructure code** - Essential for troubleshooting and cluster initialization
2. **Test SSH immediately after deployment** - Catch issues early before investing time in configuration
3. **Key pairs are region-specific** - Must create in each AWS region where you deploy
4. **Document access methods clearly** - Save time for future operations
5. **Consider using AWS Systems Manager (SSM)** - Alternative to SSH that doesn't require key management (already configured via IAM policy)

## Alternative: AWS Systems Manager Session Manager

If SSH becomes problematic, you can use SSM (already configured in the IAM role):

```bash
# Connect without SSH keys
aws ssm start-session \
  --target $(pulumi stack output control_plane_instance_id) \
  --region us-west-1
```

**Advantages:**
- No SSH keys needed
- Works even without public IPs
- Audit logs in CloudTrail
- Fine-grained IAM control

**Disadvantages:**
- Requires AWS CLI and SSM plugin
- Slightly more latency
- Different shell experience

## Summary

**Problem:** EC2 instances deployed without SSH key pairs  
**Root Cause:** Missing `key_name` parameter in Pulumi EC2 instance configuration  
**Solution:** Create SSH key pair in AWS and add to instance configuration  
**Result:** Full SSH access to all cluster nodes  
**Time to Fix:** ~10 minutes (including instance replacement)  
**Resources Affected:** 4 EC2 instances (1 control plane + 3 workers)  

## References

- [AWS EC2 Key Pairs Documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-key-pairs.html)
- [Pulumi AWS EC2 Instance](https://www.pulumi.com/registry/packages/aws/api-docs/ec2/instance/)
- [`SSH_ACCESS.md`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/SSH_ACCESS.md) - Quick reference guide
- [`ssh-to-cluster.sh`](file:///Users/john/Documents/projects/aa/tutorials/k8s/devSecOps/sunlink/selfManagedK8s/ssh-to-cluster.sh) - Helper script
