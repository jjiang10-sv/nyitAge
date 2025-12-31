from constructs import Construct
from cdktf import TerraformStack, TerraformOutput

# These imports exist after: 
#   cdktf provider add hashicorp/azurerm
#   cdktf provider add hashicorp/kubernetes
#   cdktf get
from imports.azurerm import AzurermProvider, ResourceGroup, KubernetesCluster


class AzureStack(TerraformStack):
    def __init__(self, scope: Construct, id: str):
        super().__init__(scope, id)

        # -------------------------
        # Provider
        # -------------------------
        AzurermProvider(self, "azurerm", features={})

        # -------------------------
        # Resource Group
        # -------------------------
        rg = ResourceGroup(
            self,
            "rg",
            name="cdktf-cilium-rg",
            location="eastus",
        )

        # -------------------------
        # AKS Cluster WITH Cilium Addon
        # -------------------------
        aks = KubernetesCluster(
            self,
            "aks",
            name="cdktf-cilium-aks",
            location=rg.location,
            resource_group_name=rg.name,
            dns_prefix="cdktfaks",
            identity=[{
                "type": "SystemAssigned"
            }],
            default_node_pool=[{
                "name": "default",
                "node_count": 2,
                "vm_size": "Standard_D2s_v3",
                "os_disk_size_gb": 30
            }],
            network_profile=[{
                "network_plugin": "azure",
                "network_plugin_mode": "overlay",  # REQUIRED for Cilium addon
                "pod_cidr": "10.244.0.0/16",        # Overlay requires a pod CIDR
            }],
            azure_policy_enabled=True,
            workload_identity_enabled=True,
            addon_profile=[{
                "cilium": [{
                    "enabled": True,
                    "version": "1.14.0",     # or omit to auto-select latest supported
                }]
            }],
        )

        # -------------------------
        # Outputs
        # -------------------------
        TerraformOutput(self, "aks_name", value=aks.name)
        TerraformOutput(self, "kubeconfig", value=aks.kube_config_raw)
