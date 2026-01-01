# Self-Managed Kubernetes Deployment Guide

## Quick Start

### 1. Prerequisites

```bash
# Install Pulumi
brew install pulumi  # macOS
# or curl -fsSL https://get.pulumi.com | sh

# Install AWS CLI
brew install awscli

# Configure AWS credentials
aws configure
# Enter your AWS Access Key ID, Secret Key, and default region
```

### 2. Set Up Python Environment

```bash
cd self-managed-k8s/

# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r ../requirements.txt
# or manually:
pip install pulumi pulumi-aws pulumi-kubernetes
```

### 3. Initialize Pulumi Stack

```bash
# Login to Pulumi (choose backend)
pulumi login  # Uses pulumi.com (free for individuals)
# or
pulumi login --local  # Uses local file system

# Initialize stack
pulumi stack init dev

# Verify configuration
pulumi config
```

### 4. Configure Your Stack

```bash
# Set AWS region
pulumi config set aws:region us-west-1

# Set cluster name
pulumi config set cluster_name dev-k8s

# (Optional) Set custom versions
pulumi config set k8s_version 1.29.0
pulumi config set cilium_version 1.15.1
```

### 5. Preview Deployment

```bash
# See what will be created
pulumi preview

# Expected resources:
# - 1 VPC
# - 3 Subnets (one per AZ)
# - 1 Internet Gateway
# - 1 Route Table
# - Security Groups
# - IAM Roles and Policies
# - 1 Control Plane EC2 instance
# - 3 Worker EC2 instances
```

### 6. Deploy!

```bash
pulumi up

# Review the changes, then type "yes" to proceed
# Deployment takes ~5 minutes
```

### 7. Initialize Cluster

```bash
# Get control plane IP
pulumi stack output control_plane_public_ips

# SSH to control plane
ssh ubuntu@<CONTROL_PLANE_IP>

# Run initialization script
sudo /root/init-cluster.sh

# This will:
# - Initialize Kubernetes with kubeadm
# - Install Cilium with mTLS
# - Install AWS EBS CSI driver
# - Install SPIRE server and agents
# - Display join command for workers
```

### 8. Join Worker Nodes

```bash
# Get join command from control plane
JOIN_CMD="<kubeadm join command from step 7>"

# Get worker IPs
pulumi stack output worker_public_ips

# Join each worker
ssh ubuntu@<WORKER_1_IP> "sudo $JOIN_CMD"
ssh ubuntu@<WORKER_2_IP> "sudo $JOIN_CMD"
ssh ubuntu@<WORKER_3_IP> "sudo $JOIN_CMD"
```

### 9. Download kubeconfig

```bash
# Copy kubeconfig from control plane
scp ubuntu@<CONTROL_PLANE_IP>:/etc/kubernetes/admin.conf ~/.kube/config

# Or via SSH:
ssh ubuntu@<CONTROL_PLANE_IP> "sudo cat /etc/kubernetes/admin.conf" > ~/.kube/config
chmod 600 ~/.kube/config
```

### 10. Verify Cluster

```bash
# Check nodes
kubectl get nodes

# Should show:
# NAME          STATUS   ROLES           AGE   VERSION
# ip-10-0-x-x   Ready    control-plane   5m    v1.29.0
# ip-10-0-x-x   Ready    <none>          3m    v1.29.0
# ip-10-0-x-x   Ready    <none>          3m    v1.29.0
# ip-10-0-x-x   Ready    <none>          3m    v1.29.0

# Check Cilium
kubectl get pods -n kube-system -l k8s-app=cilium

# Check SPIRE
kubectl get pods -n spire

# Check storage
kubectl get storageclass
```

---

## Stack Outputs

After deployment, these outputs are available:

```bash
# View all outputs
pulumi stack output

# Specific outputs:
pulumi stack output vpc_id
pulumi stack output control_plane_public_ips
pulumi stack output worker_public_ips
pulumi stack output kubernetes_version
pulumi stack output cilium_version
```

---

## Configuration Options

### Cluster Sizing

Edit `example_usage.py`:

```python
cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in subnets],
    control_plane_count=3,      # 1 for dev, 3 for prod
    worker_count=6,              # Scale as needed
    control_plane_instance_type="t3.medium",
    worker_instance_type="t3.medium",  # or t3.large for production
)
```

### Network Configuration

Default CIDRs (can modify in `example_usage.py`):

```python
vpc_cidr = "10.0.0.0/16"
pod_cidr = "10.32.0.0/13"
service_cidr = "10.96.0.0/12"
```

---

## Multiple Stacks

### Create Production Stack

```bash
# Create new stack
pulumi stack init production

# Configure
pulumi config set aws:region us-east-1
pulumi config set cluster_name prod-k8s

# Deploy
pulumi up
```

### Switch Between Stacks

```bash
# List stacks
pulumi stack ls

# Switch to dev
pulumi stack select dev

# Switch to production  
pulumi stack select production
```

---

## Cost Estimates

### Development (Current Config)

```
1× t3.medium control plane:    $30/month
3× t3.medium workers:           $90/month
100GB EBS (gp3):                $8/month
Data transfer:                  ~$10/month
────────────────────────────────────────
Total:                          ~$138/month
```

### Production (Recommended Config)

```
3× t3.medium control plane:    $90/month
6× t3.large workers:            $360/month
500GB EBS (gp3):               $40/month
Data transfer:                  ~$30/month
────────────────────────────────────────
Total:                          ~$520/month
```

---

## Troubleshooting

### Issue: Pulumi can't find AWS credentials

```bash
# Check credentials
aws sts get-caller-identity

# Reconfigure if needed
aws configure
```

### Issue: EC2 instances not starting

```bash
# Check AWS limits
aws service-quotas list-service-quotas \
  --service-code ec2 \
  --query 'Quotas[?QuotaName==`Running On-Demand Standard (A, C, D, H, I, M, R, T, Z) instances`]'

# Request increase if needed
```

### Issue: Can't SSH to instances

```bash
# Check security group
# Ensure port 22 is open from your IP

# Get your public IP
curl ifconfig.me

# Update security group if needed
```

### Issue: Cluster initialization fails

```bash
# SSH to control plane
ssh ubuntu@<CP_IP>

# Check logs
sudo journalctl -u kubelet -f

# Manually run init if needed
sudo /root/init-cluster.sh
```

---

## Cleanup

### Destroy Everything

```bash
# Destroy all resources
pulumi destroy

# Confirm by typing "yes"

# Remove stack
pulumi stack rm dev
```

### Destroy Only Cluster (Keep VPC)

Edit `example_usage.py` to comment out cluster creation, then:

```bash
pulumi up  # Will destroy cluster but keep VPC
```

---

## Next Steps

1. **Deploy Applications** - Use `kubectl` to deploy your apps
2. **Set Up GitOps** - Configure Argo CD from the cluster
3. **Enable Monitoring** - Install Prometheus/Grafana
4. **Configure Backups** - Set up etcd backups
5. **Review Security** - See `PRODUCTION_NETWORK.md` for hardening

---

## Additional Documentation

- **SPIRE_MTLS_GUIDE.md** - SPIRE/mTLS architecture and usage
- **PRODUCTION_NETWORK.md** - Production networking setup
- **STORAGE_GUIDE.md** - Persistent storage options
- **VPC_NETWORKING_EXPLAINED.md** - AWS VPC fundamentals
- **README.md** - General architecture overview

---

**Ready to deploy!** 🚀

Run `pulumi up` to get started!
