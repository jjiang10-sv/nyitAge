import pulumi
import pulumi_aws as aws
from pulumi import ComponentResource, ResourceOptions, Output
from typing import List, Optional
import base64

class SelfManagedK8sCluster(ComponentResource):
    """
    Self-managed Kubernetes cluster on EC2 instances with Cilium CNI.
    
    ADVANTAGES over EKS:
    - Full control over Kubernetes version
    - No AWS VPC CNI conflicts
    - Pure Cilium from day 1
    - No EKS upgrade surprises
    - No AWS support dependency
    
    TRADE-OFFS:
    - You manage control plane
    - You handle upgrades
    - More operational overhead
    - No EKS SLA
    
    Uses kubeadm for cluster bootstrapping.
    """

    def __init__(
        self,
        name: str,
        vpc_id: pulumi.Input[str],
        subnet_ids: pulumi.Input[List[str]],
        kubernetes_version: str = "1.29.0",
        cilium_version: str = "1.15.1",
        control_plane_count: int = 1,  # 1 for dev, 3 for prod
        worker_count: int = 3,
        control_plane_instance_type: str = "t3.medium",
        worker_instance_type: str = "t3.large",
        pod_cidr: str = "10.32.0.0/13",
        service_cidr: str = "10.96.0.0/12",
        opts: ResourceOptions = None,
    ):
        super().__init__("custom:SelfManagedK8sCluster", name, {}, opts)

        self.name = name
        self.k8s_version = kubernetes_version
        self.cilium_version = cilium_version
        self.pod_cidr = pod_cidr
        self.service_cidr = service_cidr

        # -------------------------
        # Security Groups
        # -------------------------
        
        # Control plane security group
        self.control_plane_sg = aws.ec2.SecurityGroup(
            f"{name}-cp-sg",
            vpc_id=vpc_id,
            description="Security group for K8s control plane",
            ingress=[
                # Kubernetes API server
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=6443,
                    to_port=6443,
                    protocol="tcp",
                    cidr_blocks=["0.0.0.0/0"],  # Adjust for production
                    description="Kubernetes API",
                ),
                # etcd
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=2379,
                    to_port=2380,
                    protocol="tcp",
                    self=True,
                    description="etcd",
                ),
                # Kubelet API
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=10250,
                    to_port=10250,
                    protocol="tcp",
                    self=True,
                    description="Kubelet API",
                ),
                # SSH (for management)
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=22,
                    to_port=22,
                    protocol="tcp",
                    cidr_blocks=["0.0.0.0/0"],  # Restrict in production
                    description="SSH",
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
            tags={"Name": f"{name}-cp-sg"},
            opts=ResourceOptions(parent=self),
        )

        # Worker security group
        self.worker_sg = aws.ec2.SecurityGroup(
            f"{name}-worker-sg",
            vpc_id=vpc_id,
            description="Security group for K8s workers",
            ingress=[
                # Kubelet API
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=10250,
                    to_port=10250,
                    protocol="tcp",
                    self=True,
                    description="Kubelet API",
                ),
                # NodePort Services
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=30000,
                    to_port=32767,
                    protocol="tcp",
                    cidr_blocks=["0.0.0.0/0"],
                    description="NodePort Services",
                ),
                # SSH
                aws.ec2.SecurityGroupIngressArgs(
                    from_port=22,
                    to_port=22,
                    protocol="tcp",
                    cidr_blocks=["0.0.0.0/0"],
                    description="SSH",
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
            tags={"Name": f"{name}-worker-sg"},
            opts=ResourceOptions(parent=self),
        )

        # Allow control plane <-> worker communication
        aws.ec2.SecurityGroupRule(
            f"{name}-cp-to-worker",
            type="ingress",
            from_port=0,
            to_port=65535,
            protocol="-1",
            security_group_id=self.worker_sg.id,
            source_security_group_id=self.control_plane_sg.id,
            description="Allow control plane to workers",
            opts=ResourceOptions(parent=self),
        )

        aws.ec2.SecurityGroupRule(
            f"{name}-worker-to-cp",
            type="ingress",
            from_port=0,
            to_port=65535,
            protocol="-1",
            security_group_id=self.control_plane_sg.id,
            source_security_group_id=self.worker_sg.id,
            description="Allow workers to control plane",
            opts=ResourceOptions(parent=self),
        )

        # Allow worker <-> worker communication (for Cilium)
        aws.ec2.SecurityGroupRule(
            f"{name}-worker-to-worker",
            type="ingress",
            from_port=0,
            to_port=65535,
            protocol="-1",
            security_group_id=self.worker_sg.id,
            source_security_group_id=self.worker_sg.id,
            description="Allow worker to worker",
            opts=ResourceOptions(parent=self),
        )

        # -------------------------
        # IAM Role for EC2 instances
        # -------------------------
        
        self.instance_role = aws.iam.Role(
            f"{name}-instance-role",
            assume_role_policy="""{
                "Version": "2012-10-17",
                "Statement": [{
                    "Effect": "Allow",
                    "Principal": {"Service": "ec2.amazonaws.com"},
                    "Action": "sts:AssumeRole"
                }]
            }""",
            opts=ResourceOptions(parent=self),
        )

        # Attach policies for Cilium ENI management
        aws.iam.RolePolicyAttachment(
            f"{name}-ssm-policy",
            role=self.instance_role.name,
            policy_arn="arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore",
            opts=ResourceOptions(parent=self),
        )

        # Custom policy for Cilium
        cilium_policy = aws.iam.Policy(
            f"{name}-cilium-policy",
            policy=pulumi.Output.json_dumps({
                "Version": "2012-10-17",
                "Statement": [{
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
                        "ec2:UnassignPrivateIpAddresses",
                        "ec2:DescribeSubnets",
                        "ec2:DescribeVpcs",
                        "ec2:DescribeTags",
                    ],
                    "Resource": "*"
                }]
            }),
            opts=ResourceOptions(parent=self),
        )

        aws.iam.RolePolicyAttachment(
            f"{name}-cilium-policy-attach",
            role=self.instance_role.name,
            policy_arn=cilium_policy.arn,
            opts=ResourceOptions(parent=self),
        )

        # Instance profile
        self.instance_profile = aws.iam.InstanceProfile(
            f"{name}-instance-profile",
            role=self.instance_role.name,
            opts=ResourceOptions(parent=self),
        )

        # -------------------------
        # Get latest Ubuntu AMI
        # -------------------------
        
        ubuntu_ami = aws.ec2.get_ami(
            most_recent=True,
            owners=["099720109477"],  # Canonical
            filters=[
                aws.ec2.GetAmiFilterArgs(
                    name="name",
                    values=["ubuntu/images/hvm-ssd/ubuntu-jammy-22.04-amd64-server-*"],
                ),
                aws.ec2.GetAmiFilterArgs(
                    name="virtualization-type",
                    values=["hvm"],
                ),
            ],
        )

        # -------------------------
        # User Data Scripts
        # -------------------------

        # Control plane init script
        control_plane_init_script = f"""#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y apt-transport-https ca-certificates curl software-properties-common

# Install containerd
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update
apt-get install -y containerd.io

# Configure containerd for Kubernetes
mkdir -p /etc/containerd
containerd config default | tee /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
systemctl restart containerd
systemctl enable containerd

# Load kernel modules
cat <<EOF | tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

modprobe overlay
modprobe br_netfilter

# Sysctl params
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sysctl --system

# Install Kubernetes components
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list
apt-get update
apt-get install -y kubelet={kubernetes_version}-1.1 kubeadm={kubernetes_version}-1.1 kubectl={kubernetes_version}-1.1
apt-mark hold kubelet kubeadm kubectl

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

# Initialize cluster with kubeadm (will be run manually after instance creation)
# This is saved as a script for manual execution
cat <<'INITSCRIPT' > /root/init-cluster.sh
#!/bin/bash
kubeadm init \\
  --pod-network-cidr={pod_cidr} \\
  --service-cidr={service_cidr} \\
  --skip-phases=addon/kube-proxy \\
  --apiserver-advertise-address=$(hostname -I | awk '{{print $1}}')

# Setup kubectl for root
mkdir -p /root/.kube
cp /etc/kubernetes/admin.conf /root/.kube/config

# Install Cilium
CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
CLI_ARCH=amd64
curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${{CILIUM_CLI_VERSION}}/cilium-linux-${{CLI_ARCH}}.tar.gz{{,.sha256sum}}
sha256sum --check cilium-linux-${{CLI_ARCH}}.tar.gz.sha256sum
tar xzvfC cilium-linux-${{CLI_ARCH}}.tar.gz /usr/local/bin
rm cilium-linux-${{CLI_ARCH}}.tar.gz{{,.sha256sum}}

# Install Cilium with native routing
cilium install \\
  --version {cilium_version} \\
  --set ipam.mode=cluster-pool \\
  --set ipam.operator.clusterPoolIPv4PodCIDRList={pod_cidr} \\
  --set tunnel=disabled \\
  --set ipv4NativeRoutingCIDR={pod_cidr} \\
  --set kubeProxyReplacement=strict \\
  --set hubble.relay.enabled=true \\
  --set hubble.ui.enabled=true

echo "Cluster initialized! Join command:"
kubeadm token create --print-join-command
INITSCRIPT

chmod +x /root/init-cluster.sh

echo "Control plane node ready. Run: sudo /root/init-cluster.sh"
"""

        # Worker node init script (join command will be added manually)
        worker_init_script = f"""#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y apt-transport-https ca-certificates curl software-properties-common

# Install containerd
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update
apt-get install -y containerd.io

# Configure containerd
mkdir -p /etc/containerd
containerd config default | tee /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
systemctl restart containerd
systemctl enable containerd

# Load kernel modules
cat <<EOF | tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

modprobe overlay
modprobe br_netfilter

# Sysctl params
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sysctl --system

# Install Kubernetes components
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list
apt-get update
apt-get install -y kubelet={kubernetes_version}-1.1 kubeadm={kubernetes_version}-1.1 kubectl={kubernetes_version}-1.1
apt-mark hold kubelet kubeadm kubectl

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

echo "Worker node ready. Run join command from control plane."
"""

        # -------------------------
        # Create Control Plane Instances
        # -------------------------
        
        self.control_plane_instances = []
        for i in range(control_plane_count):
            instance = aws.ec2.Instance(
                f"{name}-cp-{i}",
                instance_type=control_plane_instance_type,
                ami=ubuntu_ami.id,
                subnet_id=subnet_ids[i % len(subnet_ids)] if isinstance(subnet_ids, list) else subnet_ids,
                vpc_security_group_ids=[self.control_plane_sg.id],
                iam_instance_profile=self.instance_profile.name,
                user_data=control_plane_init_script,
                tags={
                    "Name": f"{name}-control-plane-{i}",
                    "Role": "control-plane",
                    "Cluster": name,
                },
                root_block_device=aws.ec2.InstanceRootBlockDeviceArgs(
                    volume_size=50,
                    volume_type="gp3",
                ),
                opts=ResourceOptions(parent=self),
            )
            self.control_plane_instances.append(instance)

        # -------------------------
        # Create Worker Instances
        # -------------------------
        
        self.worker_instances = []
        for i in range(worker_count):
            instance = aws.ec2.Instance(
                f"{name}-worker-{i}",
                instance_type=worker_instance_type,
                ami=ubuntu_ami.id,
                subnet_id=subnet_ids[i % len(subnet_ids)] if isinstance(subnet_ids, list) else subnet_ids,
                vpc_security_group_ids=[self.worker_sg.id],
                iam_instance_profile=self.instance_profile.name,
                user_data=worker_init_script,
                tags={
                    "Name": f"{name}-worker-{i}",
                    "Role": "worker",
                    "Cluster": name,
                },
                root_block_device=aws.ec2.InstanceRootBlockDeviceArgs(
                    volume_size=100,
                    volume_type="gp3",
                ),
                opts=ResourceOptions(parent=self),
            )
            self.worker_instances.append(instance)

        # -------------------------
        # Outputs
        # -------------------------
        
        self.register_outputs({
            "control_plane_ips": [i.private_ip for i in self.control_plane_instances],
            "control_plane_public_ips": [i.public_ip for i in self.control_plane_instances],
            "worker_ips": [i.private_ip for i in self.worker_instances],
            "worker_public_ips": [i.public_ip for i in self.worker_instances],
            "kubernetes_version": kubernetes_version,
            "cilium_version": cilium_version,
            "pod_cidr": pod_cidr,
            "service_cidr": service_cidr,
        })
