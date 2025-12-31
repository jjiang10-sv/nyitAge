"""
Self-Managed Kubernetes Cluster on AWS EC2 with Cilium CNI

This creates a Kubernetes cluster from scratch on EC2 instances.
NO EKS, just raw EC2 + kubeadm + Cilium.

Advantages:
- Full control over Kubernetes version
- No EKS-Cilium compatibility issues
- Pure Cilium from day 1
- No AWS VPC CNI conflicts
- Complete flexibility

Trade-offs:
- You manage the control plane
- More operational overhead
- No managed service SLA
"""

import pulumi
import pulumi_aws as aws
from platform import SelfManagedK8sCluster

# ========================================
# Configuration
# ========================================
config = pulumi.Config()
cluster_name = config.get("cluster_name") or "k8s-cluster"
kubernetes_version = config.get("k8s_version") or "1.29.0"
cilium_version = config.get("cilium_version") or "1.15.1"

# ========================================
# VPC Setup
# ========================================

# Create VPC
vpc = aws.ec2.Vpc(
    "k8s-vpc",
    cidr_block="10.0.0.0/16",
    enable_dns_hostnames=True,
    enable_dns_support=True,
    tags={"Name": "k8s-vpc"},
)

# Internet Gateway
igw = aws.ec2.InternetGateway(
    "k8s-igw",
    vpc_id=vpc.id,
    tags={"Name": "k8s-igw"},
)

# Get availability zones
azs = aws.get_availability_zones(state="available")

# Create subnets across 3 AZs
subnets = []
for i, az in enumerate(azs.names[:3]):
    subnet = aws.ec2.Subnet(
        f"k8s-subnet-{i}",
        vpc_id=vpc.id,
        cidr_block=f"10.0.{i}.0/24",
        availability_zone=az,
        map_public_ip_on_launch=True,
        tags={"Name": f"k8s-subnet-{i}"},
    )
    subnets.append(subnet)

# Route table
route_table = aws.ec2.RouteTable(
    "k8s-rt",
    vpc_id=vpc.id,
    routes=[
        aws.ec2.RouteTableRouteArgs(
            cidr_block="0.0.0.0/0",
            gateway_id=igw.id,
        )
    ],
    tags={"Name": "k8s-rt"},
)

# Associate route table with subnets
for i, subnet in enumerate(subnets):
    aws.ec2.RouteTableAssociation(
        f"k8s-rta-{i}",
        subnet_id=subnet.id,
        route_table_id=route_table.id,
    )

# ========================================
# Create Kubernetes Cluster
# ========================================

cluster = SelfManagedK8sCluster(
    cluster_name,
    vpc_id=vpc.id,
    subnet_ids=[s.id for s in subnets],
    kubernetes_version=kubernetes_version,
    cilium_version=cilium_version,
    control_plane_count=1,  # For production, use 3
    worker_count=3,
    control_plane_instance_type="t3.medium",
    worker_instance_type="t3.large",
    pod_cidr="10.32.0.0/13",
    service_cidr="10.96.0.0/12",
)

# ========================================
# Outputs
# ========================================

pulumi.export("vpc_id", vpc.id)
pulumi.export("control_plane_public_ips", cluster.control_plane_instances[0].public_ip)
pulumi.export("worker_public_ips", [w.public_ip for w in cluster.worker_instances])
pulumi.export("kubernetes_version", kubernetes_version)
pulumi.export("cilium_version", cilium_version)

# Export connection instructions
pulumi.export("instructions", pulumi.Output.concat(
    "\\n=== Cluster Setup Instructions ===\\n",
    "\\n1. SSH to control plane:\\n",
    "   ssh ubuntu@", cluster.control_plane_instances[0].public_ip, "\\n",
    "\\n2. Initialize cluster:\\n",
    "   sudo /root/init-cluster.sh\\n",
    "\\n3. Get join command:\\n",
    "   sudo kubeadm token create --print-join-command\\n",
    "\\n4. SSH to each worker and run the join command\\n",
    "\\n5. Download kubeconfig:\\n",
    "   scp ubuntu@", cluster.control_plane_instances[0].public_ip, ":/etc/kubernetes/admin.conf ~/.kube/config\\n",
    "\\n6. Verify cluster:\\n",
    "   kubectl get nodes\\n",
    "   kubectl get pods -n kube-system\\n",
))

# ========================================
# Deployment Instructions
# ========================================
#
# Deploy cluster:
#   pulumi config set cluster_name production
#   pulumi up
#
# After deployment:
#   1. SSH to control plane node
#   2. Run: sudo /root/init-cluster.sh
#   3. Copy join command
#   4. SSH to each worker and run join command
#   5. Download kubeconfig
#   6. Use cluster!
#
# Add new nodes:
#   1. Deploy new EC2 instance with worker user data
#   2. Get join command: kubeadm token create --print-join-command
#   3. SSH to new node and run join command
#
