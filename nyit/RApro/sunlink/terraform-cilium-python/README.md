# terraform-cilium-python

Python wrapper + Terraform to provision EKS (AWS) or AKS (Azure) and install Cilium via Helm.

This repo provides:
- `main.py` — Python CLI to run Terraform commands (init/plan/apply/destroy)
- Terraform configurations for AWS (EKS) and Azure (AKS) under `terraform/`
- A simple GitHub Actions workflow example

**Warning:** Always review and test infra code in throwaway accounts/resource groups. Adjust variables (VPC, subnets, regions), enable remote state, and harden permissions for production.
