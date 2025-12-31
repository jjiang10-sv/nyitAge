# terraform-cilium-cdktf-typed

This repo contains **typed CDK for Terraform (Python)** stacks that provision EKS/AKS and install Cilium via Helm.

IMPORTANT steps before running:
1. Install Terraform CLI and Node.js, and install cdktf CLI (`npm i -g cdktf-cli@latest`).
2. Create and activate Python venv, install requirements: `pip install -r requirements.txt`.
3. Add providers and generate bindings:
   cdktf provider add hashicorp/aws@~>4.0
   cdktf provider add hashicorp/kubernetes@~>2.0
   cdktf provider add hashicorp/helm@~>2.0
   cdktf provider add hashicorp/azurerm@~>3.0
   cdktf get
4. Then run `cdktf synth` and `cdktf deploy`.

Notes:
- The stacks assume minimal default values; review variables and security settings before running in production.
- For AWS EKS the stack creates a VPC, subnets, IAM roles, EKS cluster and a managed node group, then installs Cilium via Helm.
- For AKS the stack creates a resource group, AKS cluster with system-assigned identity and node pool, then installs Cilium via Helm.
