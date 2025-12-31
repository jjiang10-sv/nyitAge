import pulumi
from pulumi import ComponentResource, ResourceOptions, Output
from pulumi_azure_native import (
    containerservice,
    network,
    managedidentity,
    keyvault,
    cdn,
    resources,
)
from pulumi_kubernetes import Provider, helm
from pulumi_kubernetes.apiextensions import CustomResource
from typing import List, Optional

class AKSPlatform(ComponentResource):
    """
    Production-grade AKS platform with:
    - Multi-AZ deployment
    - Private cluster with Azure Firewall egress control
    - Cilium CNI with eBPF dataplane
    - Argo CD with ApplicationSet support
    - SPIFFE/SPIRE for workload identity
    - Gateway API for modern ingress
    - Azure Front Door for global traffic management
    """

    def __init__(
        self,
        name: str,
        vnet_id: pulumi.Input[str],
        subnet_id: pulumi.Input[str],
        location: str,
        gitops_repo: str,
        resource_group_name: Optional[pulumi.Input[str]] = None,
        pod_cidr: str = "10.32.0.0/13",
        service_cidr: str = "10.96.0.0/12",
        dns_service_ip: Optional[str] = None,
        enable_multi_region: bool = False,
        additional_regions: Optional[List[str]] = None,
        enable_spire: bool = True,
        enable_gateway_api: bool = True,
        enable_front_door: bool = False,
        opts: ResourceOptions = None,
    ):
        super().__init__("platform:AKSPlatform", name, {}, opts)

        self.location = location
        self.gitops_repo = gitops_repo
        self.pod_cidr = pod_cidr
        self.service_cidr = service_cidr
        
        # Calculate DNS service IP (10th IP in service CIDR by default)
        # e.g., 10.96.0.0/12 → 10.96.0.10
        if dns_service_ip is None:
            # Extract base IP from service CIDR
            base_ip = service_cidr.split('/')[0]
            octets = base_ip.split('.')
            # Set 4th octet to 10 (Kubernetes convention)
            self.dns_service_ip = f"{octets[0]}.{octets[1]}.{octets[2]}.10"
        else:
            self.dns_service_ip = dns_service_ip
        
        # -------------------------
        # Azure Firewall Infrastructure
        # -------------------------
        # Firewall requires a subnet with exact name "AzureFirewallSubnet"
        firewall_subnet = network.Subnet(
            f"{name}-fw-subnet",
            resource_group_name=resource_group_name,
            virtual_network_name=vnet_id,
            address_prefix="10.0.128.0/26",
            subnet_name="AzureFirewallSubnet",
            opts=ResourceOptions(parent=self),
        )

        # Public IP for Azure Firewall
        fw_public_ip = network.PublicIPAddress(
            f"{name}-fw-pip",
            resource_group_name=resource_group_name,
            location=location,
            sku=network.PublicIPAddressSkuArgs(name="Standard"),
            public_ip_allocation_method=network.IPAllocationMethod.STATIC,
            opts=ResourceOptions(parent=self),
        )

        # Azure Firewall for egress control
        firewall = network.AzureFirewall(
            f"{name}-firewall",
            resource_group_name=resource_group_name,
            location=location,
            sku=network.AzureFirewallSkuArgs(
                name="AZFW_VNet",
                tier="Standard"
            ),
            ip_configurations=[
                network.AzureFirewallIPConfigurationArgs(
                    name="fw-ipconfig",
                    subnet=network.SubResourceArgs(id=firewall_subnet.id),
                    public_ip_address=network.SubResourceArgs(id=fw_public_ip.id),
                )
            ],
            opts=ResourceOptions(parent=self),
        )

        # Get firewall private IP for routing
        fw_private_ip = firewall.ip_configurations.apply(
            lambda configs: configs[0].private_ip_address if configs else "10.0.128.4"
        )

        # Route table to force all egress through firewall
        route_table = network.RouteTable(
            f"{name}-egress-rt",
            resource_group_name=resource_group_name,
            location=location,
            routes=[
                network.RouteArgs(
                    name="default-route",
                    address_prefix="0.0.0.0/0",
                    next_hop_type=network.RouteNextHopType.VIRTUAL_APPLIANCE,
                    next_hop_ip_address=fw_private_ip,
                )
            ],
            opts=ResourceOptions(parent=self),
        )

        # Associate route table with AKS subnet
        network.SubnetRouteTableAssociation(
            f"{name}-subnet-rt-assoc",
            route_table_id=route_table.id,
            subnet_id=subnet_id,
            opts=ResourceOptions(parent=self),
        )

        # -------------------------
        # AKS Cluster (Private + Cilium + Multi-AZ)
        # -------------------------
        self.cluster = containerservice.ManagedCluster(
            f"{name}-aks",
            resource_group_name=resource_group_name,
            location=location,
            dns_prefix=name,
            # System-assigned managed identity
            identity=containerservice.ManagedClusterIdentityArgs(
                type=containerservice.ResourceIdentityType.SYSTEM_ASSIGNED
            ),
            # Private cluster configuration
            api_server_access_profile=containerservice.ManagedClusterAPIServerAccessProfileArgs(
                enable_private_cluster=True,
                private_dns_zone="System",
            ),
            # Enhanced networking with Cilium
            network_profile=containerservice.ContainerServiceNetworkProfileArgs(
                network_plugin=containerservice.NetworkPlugin.AZURE,
                network_plugin_mode=containerservice.NetworkPluginMode.OVERLAY,
                network_dataplane=containerservice.NetworkDataplane.CILIUM,
                pod_cidr=self.pod_cidr,
                service_cidr=self.service_cidr,
                dns_service_ip=self.dns_service_ip,
                outbound_type=containerservice.OutboundType.USER_DEFINED_ROUTING,  # Use firewall
            ),
            # Multi-AZ node pools
            agent_pool_profiles=[
                # System pool - runs critical Kubernetes components
                containerservice.ManagedClusterAgentPoolProfileArgs(
                    name="system",
                    mode=containerservice.AgentPoolMode.SYSTEM,
                    vm_size="Standard_D4s_v5",
                    node_count=3,
                    availability_zones=["1", "2", "3"],
                    vnet_subnet_id=subnet_id,
                    max_pods=250,
                    os_type=containerservice.OSType.LINUX,
                    type=containerservice.AgentPoolType.VIRTUAL_MACHINE_SCALE_SETS,
                ),
                # Workload pool - runs application workloads
                containerservice.ManagedClusterAgentPoolProfileArgs(
                    name="workload",
                    mode=containerservice.AgentPoolMode.USER,
                    vm_size="Standard_D8s_v5",
                    node_count=6,
                    availability_zones=["1", "2", "3"],
                    vnet_subnet_id=subnet_id,
                    max_pods=250,
                    os_type=containerservice.OSType.LINUX,
                    type=containerservice.AgentPoolType.VIRTUAL_MACHINE_SCALE_SETS,
                ),
            ],
            opts=ResourceOptions(parent=self),
        )

        # -------------------------
        # Kubernetes Provider
        # -------------------------
        kubeconfig = Output.all(
            self.cluster.resource_group_name,
            self.cluster.name
        ).apply(lambda args:
            containerservice.list_managed_cluster_user_credentials(
                resource_group_name=args[0],
                resource_name=args[1]
            ).kubeconfigs[0].value.apply(lambda v: v.decode())
        )

        self.k8s_provider = Provider(
            f"{name}-k8s",
            kubeconfig=kubeconfig,
            opts=ResourceOptions(parent=self),
        )

        # -------------------------
        # Cilium with Hubble + Gateway API
        # -------------------------
        cilium_values = {
            "hubble": {
                "enabled": True,
                "ui": {"enabled": True},
                "relay": {"enabled": True},
                "metrics": {
                    "enabled": ["dns", "drop", "tcp", "flow", "port-distribution", "icmp"]
                }
            },
            "prometheus": {"enabled": True},
            "operator": {
                "prometheus": {"enabled": True}
            },
        }

        if enable_gateway_api:
            cilium_values["gatewayAPI"] = {
                "enabled": True,
            }

        helm.v3.Chart(
            f"{name}-cilium",
            helm.v3.ChartOpts(
                chart="cilium",
                namespace="kube-system",
                fetch_opts=helm.v3.FetchOpts(
                    repo="https://helm.cilium.io"
                ),
                values=cilium_values,
            ),
            opts=ResourceOptions(
                provider=self.k8s_provider,
                parent=self,
            ),
        )

        # -------------------------
        # Gateway API CRDs (if enabled)
        # -------------------------
        if enable_gateway_api:
            helm.v3.Chart(
                f"{name}-gateway-api",
                helm.v3.ChartOpts(
                    chart="gateway-api",
                    namespace="gateway-system",
                    fetch_opts=helm.v3.FetchOpts(
                        repo="https://gateway-api.github.io/gateway-api"
                    ),
                    values={
                        "crds": {
                            "gatewayclass": True,
                            "gateway": True,
                            "httproute": True,
                        }
                    }
                ),
                opts=ResourceOptions(
                    provider=self.k8s_provider,
                    parent=self,
                ),
            )

        # -------------------------
        # Argo CD with ApplicationSet
        # -------------------------
        argocd_chart = helm.v3.Chart(
            f"{name}-argocd",
            helm.v3.ChartOpts(
                chart="argo-cd",
                namespace="argocd",
                fetch_opts=helm.v3.FetchOpts(
                    repo="https://argoproj.github.io/argo-helm"
                ),
                values={
                    "server": {
                        "service": {"type": "ClusterIP"}  # Private cluster
                    },
                    "applicationSet": {
                        "enabled": True,  # Enable ApplicationSet controller
                    },
                    "dex": {
                        "enabled": False,  # Disable Dex for now, use built-in auth
                    },
                }
            ),
            opts=ResourceOptions(
                provider=self.k8s_provider,
                parent=self,
            ),
        )

        # Create bootstrap ApplicationSet for GitOps
        bootstrap_app_set = CustomResource(
            f"{name}-bootstrap-appset",
            api_version="argoproj.io/v1alpha1",
            kind="ApplicationSet",
            metadata={
                "name": "cluster-bootstrap",
                "namespace": "argocd",
            },
            spec={
                "generators": [
                    {
                        "git": {
                            "repoURL": gitops_repo,
                            "revision": "main",
                            "directories": [
                                {"path": "apps/*"},
                                {"path": "platform/*"},
                            ]
                        }
                    }
                ],
                "template": {
                    "metadata": {
                        "name": "{{path.basename}}",
                    },
                    "spec": {
                        "project": "default",
                        "source": {
                            "repoURL": gitops_repo,
                            "targetRevision": "main",
                            "path": "{{path}}",
                        },
                        "destination": {
                            "server": "https://kubernetes.default.svc",
                            "namespace": "{{path.basename}}",
                        },
                        "syncPolicy": {
                            "automated": {
                                "prune": True,
                                "selfHeal": True,
                            },
                            "syncOptions": ["CreateNamespace=true"],
                        },
                    },
                },
            },
            opts=ResourceOptions(
                provider=self.k8s_provider,
                parent=self,
                depends_on=[argocd_chart],
            ),
        )

        # -------------------------
        # SPIFFE/SPIRE (if enabled)
        # -------------------------
        if enable_spire:
            # Install SPIRE server
            spire_server = helm.v3.Chart(
                f"{name}-spire-server",
                helm.v3.ChartOpts(
                    chart="spire-server",
                    namespace="spire",
                    fetch_opts=helm.v3.FetchOpts(
                        repo="https://spiffe.github.io/helm-charts-hardened"
                    ),
                    values={
                        "global": {
                            "spire": {
                                "clusterName": name,
                                "trustDomain": f"{name}.local",
                            }
                        },
                        "server": {
                            "ca": {
                                "subject": {
                                    "country": "US",
                                    "organization": "Platform",
                                    "commonName": f"{name}-spire",
                                }
                            }
                        }
                    }
                ),
                opts=ResourceOptions(
                    provider=self.k8s_provider,
                    parent=self,
                ),
            )

            # Install SPIRE agent (DaemonSet on all nodes)
            helm.v3.Chart(
                f"{name}-spire-agent",
                helm.v3.ChartOpts(
                    chart="spire-agent",
                    namespace="spire",
                    fetch_opts=helm.v3.FetchOpts(
                        repo="https://spiffe.github.io/helm-charts-hardened"
                    ),
                    values={
                        "global": {
                            "spire": {
                                "clusterName": name,
                                "trustDomain": f"{name}.local",
                            }
                        },
                    }
                ),
                opts=ResourceOptions(
                    provider=self.k8s_provider,
                    parent=self,
                    depends_on=[spire_server],
                ),
            )

        # -------------------------
        # Workload Identity + Key Vault
        # -------------------------
        identity = managedidentity.UserAssignedIdentity(
            f"{name}-wi",
            resource_group_name=resource_group_name,
            location=location,
            opts=ResourceOptions(parent=self),
        )

        vault = keyvault.Vault(
            f"{name}-kv",
            resource_group_name=resource_group_name,
            location=location,
            properties=keyvault.VaultPropertiesArgs(
                tenant_id=identity.tenant_id,
                sku=keyvault.SkuArgs(
                    family=keyvault.SkuFamily.A,
                    name=keyvault.SkuName.STANDARD,
                ),
                access_policies=[],
                enable_rbac_authorization=True,
                network_acls=keyvault.NetworkRuleSetArgs(
                    bypass=keyvault.NetworkRuleBypassOptions.AZURE_SERVICES,
                    default_action=keyvault.NetworkRuleAction.DENY,
                ),
            ),
            opts=ResourceOptions(parent=self),
        )

        # -------------------------
        # Azure Front Door (if enabled)
        # -------------------------
        if enable_front_door:
            # Create Front Door profile
            fd_profile = cdn.Profile(
                f"{name}-frontdoor",
                resource_group_name=resource_group_name,
                location="Global",
                sku=cdn.SkuArgs(name=cdn.SkuName.PREMIUM_AZURE_FRONT_DOOR),
                opts=ResourceOptions(parent=self),
            )

            # Create endpoint
            fd_endpoint = cdn.AFDEndpoint(
                f"{name}-fd-endpoint",
                resource_group_name=resource_group_name,
                profile_name=fd_profile.name,
                endpoint_name=f"{name}-endpoint",
                location="Global",
                enabled_state=cdn.EnabledState.ENABLED,
                opts=ResourceOptions(parent=self),
            )

            # WAF policy for Front Door
            fd_waf = cdn.Policy(
                f"{name}-fd-waf",
                resource_group_name=resource_group_name,
                policy_name=f"{name}wafpolicy",
                sku=cdn.SkuArgs(name=cdn.SkuName.PREMIUM_AZURE_FRONT_DOOR),
                managed_rules=cdn.ManagedRuleSetListArgs(
                    managed_rule_sets=[
                        cdn.ManagedRuleSetArgs(
                            rule_set_type="Microsoft_DefaultRuleSet",
                            rule_set_version="2.1",
                        ),
                        cdn.ManagedRuleSetArgs(
                            rule_set_type="Microsoft_BotManagerRuleSet",
                            rule_set_version="1.0",
                        ),
                    ]
                ),
                policy_settings=cdn.PolicySettingsArgs(
                    enabled_state=cdn.PolicyEnabledState.ENABLED,
                    mode=cdn.PolicyMode.PREVENTION,
                ),
                opts=ResourceOptions(parent=self),
            )

            self.front_door_profile = fd_profile
            self.front_door_endpoint = fd_endpoint
            self.front_door_waf = fd_waf

        # -------------------------
        # Multi-Region Support (if enabled)
        # -------------------------
        self.regional_clusters = {}
        if enable_multi_region and additional_regions:
            for idx, region in enumerate(additional_regions):
                # Note: In production, you'd create separate VNets per region
                # with non-overlapping CIDR ranges
                # This is a simplified example showing the structure
                self.regional_clusters[region] = {
                    "location": region,
                    "note": f"Create separate VNet and AKSPlatform instance for {region}",
                    # In practice: AKSPlatform(f"{name}-{region}", ..., location=region)
                }

        # -------------------------
        # Outputs
        # -------------------------
        self.register_outputs({
            "cluster_name": self.cluster.name,
            "cluster_id": self.cluster.id,
            "key_vault_name": vault.name,
            "key_vault_uri": vault.properties.apply(lambda p: p.vault_uri),
            "firewall_public_ip": fw_public_ip.ip_address,
            "kubeconfig": pulumi.Output.secret(kubeconfig),
            "gitops_repo": gitops_repo,
            "pod_cidr": self.pod_cidr,
            "service_cidr": self.service_cidr,
            "dns_service_ip": self.dns_service_ip,
        })
