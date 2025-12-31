"""
Production-ready Multi-Region AKS Platform Deployment

This example demonstrates:
1. Single-region deployment (default)
2. Multi-region deployment with separate clusters
3. Azure Front Door for global load balancing
4. Optional VNet peering for cluster-to-cluster communication
5. Proper CIDR allocation to avoid conflicts
"""

import pulumi
from pulumi_azure_native import resources, network, cdn
from platform import AKSPlatform
from typing import List, Dict

# ========================================
# Configuration
# ========================================
config = pulumi.Config()
gitops_repo = config.require("gitops_repo")
deploy_multi_region = config.get_bool("multi_region") or False
enable_vnet_peering = config.get_bool("enable_vnet_peering") or False

# ========================================
# Region Configuration
# ========================================
# Define all regions with non-overlapping CIDR ranges
REGIONS = [
    {
        "name": "canadacentral",
        "short_name": "can",
        "vnet_cidr": "10.0.0.0/14",      # 262k IPs
        "node_subnet": "10.0.0.0/16",    # 65k IPs
        "pod_cidr": "10.32.0.0/13",      # 524k pods
        "service_cidr": "10.96.0.0/12",  # Virtual (can reuse)
        "is_primary": True,
    },
    {
        "name": "eastus",
        "short_name": "eus",
        "vnet_cidr": "10.4.0.0/14",      # Non-overlapping!
        "node_subnet": "10.4.0.0/16",
        "pod_cidr": "10.40.0.0/13",      # Non-overlapping!
        "service_cidr": "10.96.0.0/12",  # Same OK (virtual)
        "is_primary": False,
    },
    {
        "name": "westeurope",
        "short_name": "euw",
        "vnet_cidr": "10.8.0.0/14",      # Non-overlapping!
        "node_subnet": "10.8.0.0/16",
        "pod_cidr": "10.48.0.0/13",      # Non-overlapping!
        "service_cidr": "10.96.0.0/12",  # Same OK (virtual)
        "is_primary": False,
    },
]

# Use only primary region if multi-region is disabled
regions_to_deploy = REGIONS if deploy_multi_region else [r for r in REGIONS if r["is_primary"]]

# ========================================
# Deploy Regional Infrastructure
# ========================================
regional_deployments: List[Dict] = []

for region_config in regions_to_deploy:
    region_name = region_config["name"]
    short_name = region_config["short_name"]
    
    # Create resource group per region
    rg = resources.ResourceGroup(
        f"platform-{short_name}-rg",
        location=region_name,
        tags={
            "environment": "production",
            "region": region_name,
            "managed-by": "pulumi",
        }
    )
    
    # Create VNet per region with non-overlapping CIDR
    vnet = network.VirtualNetwork(
        f"platform-{short_name}-vnet",
        resource_group_name=rg.name,
        location=rg.location,
        address_space=network.AddressSpaceArgs(
            address_prefixes=[region_config["vnet_cidr"]]
        ),
        tags={
            "region": region_name,
        }
    )
    
    # Create AKS node subnet (spans all 3 AZs automatically)
    node_subnet = network.Subnet(
        f"aks-{short_name}-subnet",
        resource_group_name=rg.name,
        virtual_network_name=vnet.name,
        address_prefix=region_config["node_subnet"],
    )
    
    # Deploy AKS Platform in this region
    platform = AKSPlatform(
        f"prod-{short_name}",
        vnet_id=vnet.name,
        subnet_id=node_subnet.id,
        resource_group_name=rg.name,
        location=region_name,
        gitops_repo=gitops_repo,
        pod_cidr=region_config["pod_cidr"],
        service_cidr=region_config["service_cidr"],
        enable_front_door=False,  # Deploy Front Door globally once
        enable_spire=True,
        enable_gateway_api=True,
    )
    
    # Store deployment info
    regional_deployments.append({
        "region": region_name,
        "short_name": short_name,
        "resource_group": rg,
        "vnet": vnet,
        "subnet": node_subnet,
        "platform": platform,
        "config": region_config,
    })
    
    # Export per-region outputs
    pulumi.export(f"{short_name}_cluster_name", platform.cluster.name)
    pulumi.export(f"{short_name}_vnet_cidr", region_config["vnet_cidr"])
    pulumi.export(f"{short_name}_pod_cidr", platform.pod_cidr)
    pulumi.export(f"{short_name}_service_cidr", platform.service_cidr)

# ========================================
# Azure Front Door (Multi-Region Only)
# ========================================
if deploy_multi_region:
    # Get primary region for Front Door resource group
    primary_region = [r for r in regional_deployments if r["config"]["is_primary"]][0]
    
    # Create Front Door Profile (global resource)
    fd_profile = cdn.Profile(
        "global-frontdoor",
        resource_group_name=primary_region["resource_group"].name,
        location="Global",
        sku=cdn.SkuArgs(name=cdn.SkuName.PREMIUM_AZURE_FRONT_DOOR),
        tags={
            "purpose": "global-load-balancing",
        }
    )
    
    # Create Front Door Endpoint
    fd_endpoint = cdn.AFDEndpoint(
        "global-endpoint",
        resource_group_name=primary_region["resource_group"].name,
        profile_name=fd_profile.name,
        endpoint_name="aks-global",
        location="Global",
        enabled_state=cdn.EnabledState.ENABLED,
    )
    
    # Create WAF Policy
    fd_waf = cdn.Policy(
        "global-waf",
        resource_group_name=primary_region["resource_group"].name,
        policy_name="aksglobalwaf",
        sku=cdn.SkuArgs(name=cdn.SkuName.PREMIUM_AZURE_FRONT_DOOR),
        managed_rules=cdn.ManagedRuleSetListArgs(
            managed_rule_sets=[
                # OWASP Top 10 protection
                cdn.ManagedRuleSetArgs(
                    rule_set_type="Microsoft_DefaultRuleSet",
                    rule_set_version="2.1",
                ),
                # Bot protection
                cdn.ManagedRuleSetArgs(
                    rule_set_type="Microsoft_BotManagerRuleSet",
                    rule_set_version="1.0",
                ),
            ]
        ),
        policy_settings=cdn.PolicySettingsArgs(
            enabled_state=cdn.PolicyEnabledState.ENABLED,
            mode=cdn.PolicyMode.PREVENTION,  # Block attacks
        ),
    )
    
    # Create Origin Group with all regional clusters
    fd_origin_group = cdn.AFDOriginGroup(
        "global-origin-group",
        resource_group_name=primary_region["resource_group"].name,
        profile_name=fd_profile.name,
        origin_group_name="aks-clusters",
        load_balancing_settings=cdn.LoadBalancingSettingsParametersArgs(
            sample_size=4,
            successful_samples_required=3,
            additional_latency_in_milliseconds=50,
        ),
        health_probe_settings=cdn.HealthProbeParametersArgs(
            probe_interval_in_seconds=30,
            probe_path="/health",
            probe_protocol=cdn.ProbeProtocol.HTTPS,
            probe_request_type=cdn.HealthProbeRequestType.GET,
        ),
    )
    
    # Add origins (one per region)
    # Note: In production, you'd configure Private Link to AKS internal load balancers
    for idx, deployment in enumerate(regional_deployments):
        cdn.AFDOrigin(
            f"origin-{deployment['short_name']}",
            resource_group_name=primary_region["resource_group"].name,
            profile_name=fd_profile.name,
            origin_group_name=fd_origin_group.name,
            origin_name=f"aks-{deployment['short_name']}",
            host_name=f"{deployment['short_name']}.example.com",  # Replace with actual
            http_port=80,
            https_port=443,
            priority=1 if deployment["config"]["is_primary"] else 2,
            weight=1000,
            enabled_state=cdn.EnabledState.ENABLED,
        )
    
    # Export Front Door outputs
    pulumi.export("front_door_endpoint", fd_endpoint.host_name)
    pulumi.export("front_door_url", fd_endpoint.host_name.apply(lambda h: f"https://{h}"))
    pulumi.export("waf_policy", fd_waf.name)

# ========================================
# VNet Peering (Optional)
# ========================================
if deploy_multi_region and enable_vnet_peering:
    # Create full-mesh VNet peering between all regions
    # This allows cluster-to-cluster communication
    
    for i, source_deployment in enumerate(regional_deployments):
        for j, target_deployment in enumerate(regional_deployments):
            if i != j:  # Don't peer to self
                # Create bidirectional peering
                network.VirtualNetworkPeering(
                    f"peer-{source_deployment['short_name']}-to-{target_deployment['short_name']}",
                    resource_group_name=source_deployment["resource_group"].name,
                    virtual_network_name=source_deployment["vnet"].name,
                    virtual_network_peering_name=f"to-{target_deployment['short_name']}",
                    remote_virtual_network=network.SubResourceArgs(
                        id=target_deployment["vnet"].id
                    ),
                    allow_virtual_network_access=True,
                    allow_forwarded_traffic=True,
                    allow_gateway_transit=False,
                    use_remote_gateways=False,
                )
    
    pulumi.export("vnet_peering_enabled", True)

# ========================================
# Summary Exports
# ========================================
pulumi.export("deployment_mode", "multi-region" if deploy_multi_region else "single-region")
pulumi.export("total_regions", len(regional_deployments))
pulumi.export("regions", [d["region"] for d in regional_deployments])
pulumi.export("gitops_repo", gitops_repo)

# Export CIDR allocation table
cidr_allocation = {
    d["region"]: {
        "vnet": d["config"]["vnet_cidr"],
        "nodes": d["config"]["node_subnet"],
        "pods": d["config"]["pod_cidr"],
        "services": d["config"]["service_cidr"],
    }
    for d in regional_deployments
}
pulumi.export("cidr_allocation", cidr_allocation)

# ========================================
# Deployment Instructions
# ========================================
# 
# Single Region (default):
#   pulumi config set gitops_repo https://github.com/org/gitops
#   pulumi up
#
# Multi-Region:
#   pulumi config set multi_region true
#   pulumi up
#
# Multi-Region + VNet Peering:
#   pulumi config set multi_region true
#   pulumi config set enable_vnet_peering true
#   pulumi up
#
# ========================================
# CIDR Allocation Map
# ========================================
#
# Canada Central:
#   VNet:     10.0.0.0/14
#   Nodes:    10.0.0.0/16
#   Pods:     10.32.0.0/13
#   Services: 10.96.0.0/12
#
# East US:
#   VNet:     10.4.0.0/14
#   Nodes:    10.4.0.0/16
#   Pods:     10.40.0.0/13
#   Services: 10.96.0.0/12 (reused)
#
# West Europe:
#   VNet:     10.8.0.0/14
#   Nodes:    10.8.0.0/16
#   Pods:     10.48.0.0/13
#   Services: 10.96.0.0/12 (reused)
#
