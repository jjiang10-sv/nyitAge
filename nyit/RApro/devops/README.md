# Enhanced AKS Platform

Production-grade Azure Kubernetes Service (AKS) platform with advanced cloud-native features.

## Features

### 🏗️ Infrastructure
- **Multi-AZ Deployment**: System and workload node pools across 3 availability zones
- **Private AKS Cluster**: API server accessible only via private network
- **Azure Firewall**: Centralized egress control with user-defined routing
- **Optimized CIDR Planning**: 
  - Pod CIDR: `10.32.0.0/13` (524,288 IPs)
  - Service CIDR: `10.96.0.0/12` (1,048,576 IPs)
  - Supports massive scale (2,000+ nodes, 500,000+ pods)

### 🌐 Networking
- **Cilium CNI**: eBPF-based dataplane for high-performance networking
- **Hubble**: Network observability and security monitoring
- **Gateway API**: Modern, extensible Kubernetes ingress (replaces traditional Ingress)
- **Azure Front Door**: Global load balancing with WAF protection

### 🔐 Security & Identity
- **SPIFFE/SPIRE**: Platform-agnostic workload identity and automatic mTLS
- **Azure Workload Identity**: Native Azure service authentication
- **Key Vault Integration**: Secrets management with RBAC
- **Network Policies**: eBPF-based microsegmentation

### 🚀 GitOps & Deployment
- **Argo CD**: Declarative GitOps continuous deployment
- **ApplicationSet**: Multi-cluster and multi-tenant application management
- **Automated Sync**: Self-healing deployments with drift detection

### 🌍 Multi-Region (Optional)
- Support for deploying across multiple Azure regions
- Global traffic management via Azure Front Door
- Regional failover capabilities

## Architecture

```
Internet
   ↓
Azure Front Door (WAF, Global LB)
   ↓
Private Link / Private Endpoint
   ↓
Azure Firewall (Egress Control)
   ↓
Private AKS Cluster
   ├── Cilium (eBPF CNI)
   ├── Gateway API (Ingress)
   ├── SPIRE (mTLS Identity)
   ├── Argo CD (GitOps)
   └── Multi-AZ Node Pools
       ├── System Pool (3 nodes)
       └── Workload Pool (6 nodes)
```

## Prerequisites

```bash
# Install Pulumi
brew install pulumi

# Install Azure CLI
brew install azure-cli

# Login to Azure
az login

# Install Python dependencies
pip install pulumi pulumi-azure-native pulumi-kubernetes
```

## Quick Start

### 1. Configure Pulumi

```bash
# Create new Pulumi project
pulumi new python

# Set Azure region
pulumi config set location canadacentral

# Set GitOps repository
pulumi config set gitops_repo https://github.com/your-org/gitops
```

### 2. Deploy Platform

```python
from platform import AKSPlatform

platform = AKSPlatform(
    "production",
    vnet_id=vnet.name,
    subnet_id=subnet_id,
    resource_group_name=rg.name,
    location="canadacentral",
    gitops_repo="https://github.com/your-org/gitops",
    enable_spire=True,
    enable_gateway_api=True,
    enable_front_door=True,
)
```

See [`example_usage.py`](./example_usage.py) for complete example.

### 3. Deploy Infrastructure

```bash
pulumi up
```

### 4. Access Cluster

```bash
# Get kubeconfig
az aks get-credentials \
  --resource-group <rg-name> \
  --name <cluster-name>

# Verify cluster
kubectl get nodes
kubectl get pods -A
```

## Feature Flags

| Parameter | Default | Description |
|-----------|---------|-------------|
| `enable_multi_region` | `False` | Deploy to multiple regions |
| `enable_spire` | `True` | Install SPIFFE/SPIRE for workload identity |
| `enable_gateway_api` | `True` | Enable Gateway API for modern ingress |
| `enable_front_door` | `False` | Deploy Azure Front Door for global traffic management |

## GitOps Repository Structure

Your GitOps repository should follow this structure:

```
gitops/
├── apps/
│   ├── frontend/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── backend/
│       ├── deployment.yaml
│       └── service.yaml
├── platform/
│   ├── monitoring/
│   ├── security/
│   └── networking/
└── clusters/
    ├── production/
    └── staging/
```

Argo CD ApplicationSet will automatically discover and deploy applications from the `apps/` and `platform/` directories.

## Verification

### Check Cilium Status

```bash
# Install Cilium CLI
brew install cilium-cli

# Check status
cilium status

# Run connectivity test
cilium connectivity test
```

### Check Hubble (Network Observability)

```bash
# Port-forward Hubble UI
kubectl port-forward -n kube-system svc/hubble-ui 8080:80

# Open http://localhost:8080
```

### Check SPIRE

```bash
# Check SPIRE server health
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server healthcheck

# List registered workloads
kubectl exec -n spire spire-server-0 -- \
  /opt/spire/bin/spire-server entry show
```

### Check Argo CD

```bash
# Port-forward Argo CD UI
kubectl port-forward -n argocd svc/argocd-server 8080:443

# Get admin password
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d

# Open https://localhost:8080
```

### Check Gateway API

```bash
# List gateways
kubectl get gateway -A

# List HTTP routes
kubectl get httproute -A
```

## CIDR Planning

The platform uses the following CIDR allocation:

| Component | CIDR | IPs | Purpose |
|-----------|------|-----|---------|
| VNet | `10.0.0.0/14` | 262,144 | Main virtual network |
| Node Subnet | `10.0.0.0/16` | 65,536 | AKS node VMs |
| Firewall Subnet | `10.0.128.0/26` | 64 | Azure Firewall (fixed) |
| Pod CIDR | `10.32.0.0/13` | 524,288 | Kubernetes pods |
| Service CIDR | `10.96.0.0/12` | 1,048,576 | Kubernetes services |

This allocation supports:
- **2,000+ nodes**
- **500,000+ pods**
- **1M+ services**

## Multi-Region Deployment

For multi-region deployments, use separate CIDR ranges per region:

| Region | VNet CIDR | Pod CIDR |
|--------|-----------|----------|
| Region 1 | `10.0.0.0/14` | `10.32.0.0/13` |
| Region 2 | `10.4.0.0/14` | `10.40.0.0/13` |
| Region 3 | `10.8.0.0/14` | `10.48.0.0/13` |

## Security Considerations

### Egress Control

All pod egress traffic is routed through Azure Firewall. Configure firewall rules to allow necessary destinations:

```python
# Example: Allow HTTPS egress
network.AzureFirewallApplicationRule(
    "allow-https",
    rule_collections=[
        {
            "name": "allow-https",
            "priority": 100,
            "action": "Allow",
            "rules": [
                {
                    "name": "https",
                    "protocols": [{"port": 443, "protocol_type": "Https"}],
                    "target_fqdns": ["*.microsoft.com", "*.github.com"],
                }
            ]
        }
    ]
)
```

### Network Policies

Use Cilium NetworkPolicies for microsegmentation:

```yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  name: backend-policy
spec:
  endpointSelector:
    matchLabels:
      app: backend
  ingress:
  - fromEndpoints:
    - matchLabels:
        app: frontend
    toPorts:
    - ports:
      - port: "8080"
```

### Workload Identity

SPIRE automatically provisions X.509 certificates to workloads. Example usage:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: app
spec:
  containers:
  - name: app
    volumeMounts:
    - name: spire-agent-socket
      mountPath: /run/spire/sockets
  volumes:
  - name: spire-agent-socket
    hostPath:
      path: /run/spire/sockets
      type: Directory
```

## Cost Optimization

- **System Pool**: 3x `Standard_D4s_v5` (~$280/month)
- **Workload Pool**: 6x `Standard_D8s_v5` (~$1,120/month)
- **Azure Firewall**: ~$1,200/month
- **Front Door**: ~$35/month + data transfer

**Total**: ~$2,635/month (baseline, excluding data transfer and storage)

## Troubleshooting

### Private Cluster Access

If you can't access the private cluster, ensure you're on the VNet or use Azure Bastion:

```bash
# Option 1: Deploy jumpbox in same VNet
# Option 2: Use Azure Bastion
# Option 3: VPN Gateway
```

### Firewall Blocking Traffic

Check firewall logs:

```bash
az monitor diagnostic-settings create \
  --resource <firewall-id> \
  --workspace <log-analytics-id> \
  --logs '[{"category": "AzureFirewallApplicationRule", "enabled": true}]'
```

## References

- [AKS Best Practices](https://learn.microsoft.com/en-us/azure/aks/best-practices)
- [Cilium Documentation](https://docs.cilium.io/)
- [Gateway API](https://gateway-api.sigs.k8s.io/)
- [SPIFFE/SPIRE](https://spiffe.io/docs/latest/)
- [Argo CD](https://argo-cd.readthedocs.io/)
- [Azure Front Door](https://learn.microsoft.com/en-us/azure/frontdoor/)

## License

MIT
