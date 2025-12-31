from constructs import Construct
from cdktf import TerraformStack, TerraformOutput
# After running `cdktf get` these imports will exist
from imports.aws import AwsProvider, Vpc, Subnet, InternetGateway, RouteTable, Route, RouteTableAssociation, SecurityGroup, SecurityGroupRule, IamRole, IamRolePolicyAttachment, EksCluster, EksNodeGroup
from imports.kubernetes import KubernetesProvider
from imports.helm import HelmProvider, Release as HelmRelease

class AwsStack(TerraformStack):
    def __init__(self, scope: Construct, id: str):
        super().__init__(scope, id)

        # Provider
        AwsProvider(self, "aws", region = "us-west-2")

        # --- Networking: simple VPC ---
        vpc = Vpc(self, "vpc",
                  cidr_block = "10.0.0.0/16",                      enable_dns_hostnames = True,                      enable_dns_support = True,                      tags={"Name": "cdktf-cilium-vpc"})

        igw = InternetGateway(self, "igw", vpc_id = vpc.id, tags={"Name": "cdktf-igw"})

        # create two public subnets
        subnet1 = Subnet(self, "subnet1", vpc_id = vpc.id, cidr_block = "10.0.1.0/24", availability_zone = "us-west-2a", map_public_ip_on_launch=True)
        subnet2 = Subnet(self, "subnet2", vpc_id = vpc.id, cidr_block = "10.0.2.0/24", availability_zone = "us-west-2b", map_public_ip_on_launch=True)

        # route table and route
        rtable = RouteTable(self, "rtable", vpc_id = vpc.id, tags={"Name": "cdktf-rt"})
        Route(self, "route", route_table_id = rtable.id, destination_cidr_block = "0.0.0.0/0", gateway_id = igw.id)
        RouteTableAssociation(self, "rta1", subnet_id = subnet1.id, route_table_id = rtable.id)
        RouteTableAssociation(self, "rta2", subnet_id = subnet2.id, route_table_id = rtable.id)

        # Security group for node group and control plane access
        sg = SecurityGroup(self, "nodes_sg", name = "cdktf-nodes-sg", vpc_id = vpc.id, description = "Allow SSH and all egress" )
        SecurityGroupRule(self, "sg_ssh", type_ = "ingress", from_port = 22, to_port = 22, protocol = "tcp", cidr_blocks=["0.0.0.0/0"], security_group_id = sg.id)
        SecurityGroupRule(self, "sg_all_egress", type_ = "egress", from_port = 0, to_port = 0, protocol = "-1", cidr_blocks=["0.0.0.0/0"], security_group_id = sg.id)

        # IAM role for EKS cluster and node group
        eks_role = IamRole(self, "eks_role",
                           assume_role_policy = '''{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}''')

        # Attach AmazonEKSClusterPolicy to role (managed policy)
        IamRolePolicyAttachment(self, "eks_role_attach", role = eks_role.name, policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy")

        node_role = IamRole(self, "node_role",
                            assume_role_policy = '''{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}''')

        IamRolePolicyAttachment(self, "node_attach1", role = node_role.name, policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy")
        IamRolePolicyAttachment(self, "node_attach2", role = node_role.name, policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly")
        IamRolePolicyAttachment(self, "node_attach3", role = node_role.name, policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy")

        # EKS Cluster
        eks = EksCluster(self, "eks_cluster",
                         name = "cdktf-cilium-eks",
                         role_arn = eks_role.arn,
                         vpc_config = {
                             "subnet_ids": [subnet1.id, subnet2.id],
                             "endpoint_private_access": False,
                             "endpoint_public_access": True,
                         },
                         kubernetes_network_config = None,
                         version = "1.26")

        # Managed node group
        EksNodeGroup(self, "node_group",
                     cluster_name = eks.name,
                     node_group_name = "default-ng",
                     node_role_arn = node_role.arn,
                     subnet_ids = [subnet1.id, subnet2.id],
                     scaling_config = {"desired_size": 2, "max_size": 3, "min_size": 1},
                     instance_types = ["t3.medium"]) 

        # Kubernetes and Helm providers using cluster data (CDKTF will wire outputs after apply)

        kubernetes_provider = KubernetesProvider(self, "k8s", host = eks.endpoint, token = eks.certificate_authority)  # placeholder: will be replaced by proper token/ca after generation
        helm_provider = HelmProvider(self, "helm", kubernetes = {"host": eks.endpoint})

        # Cilium Helm release
        HelmRelease(self, "cilium",
                    name = "cilium",
                    repository = "https://helm.cilium.io/",
                    chart = "cilium",
                    version = "1.14.0",
                    namespace = "kube-system",
                    create_namespace = False,
                    set = [
                        {"name": "global.kubeProxyReplacement", "value": "partial"},
                        {"name": "global.cni.chainingMode", "value": "none"}
                    ])

        # Outputs
        TerraformOutput(self, "cluster_name", value = eks.name)
        TerraformOutput(self, "cluster_endpoint", value = eks.endpoint)
