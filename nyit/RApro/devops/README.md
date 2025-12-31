# AKS vs EKS Platform Guide - Documentation Index

This directory contains comprehensive documentation comparing AKS and EKS platforms with Cilium CNI.

## 📚 Core Documentation

### Platform Implementation
- **[aks/platform.py](aks/platform.py)** - AKS platform with Azure CNI + Cilium hybrid (533 lines)
- **[aks/example_usage.py](aks/example_usage.py)** - AKS multi-region deployment example
- **[eks/platform.py](eks/platform.py)** - EKS platform with pure Cilium (390 lines, 26% simpler!)
- **[eks/example_usage.py](eks/example_usage.py)** - EKS multi-region deployment example

---

## 🎯 Quick Start Guides

### AKS Quick Start
- **[aks/README.md](aks/README.md)** - Complete AKS setup guide
- **[MULTI_REGION_QUICKSTART.md](MULTI_REGION_QUICKSTART.md)** - AKS multi-region deployment

### EKS Quick Start
- **[eks/README.md](eks/README.md)** - Complete EKS setup guide (simpler than AKS!)

---

## 📊 Comparison Documents

### Architecture & Complexity
- **[AKS_VS_EKS_COMPLEXITY.md](AKS_VS_EKS_COMPLEXITY.md)** ⭐ **START HERE**
  - Why EKS is 26% simpler (630 vs 849 lines)
  - Architecture comparison
  - Feature availability

### Cost Analysis
- **[COST_ANALYSIS.md](COST_ANALYSIS.md)** 💰 **CRITICAL FOR DECISION**
  - Why EKS is 5-6x cheaper ($470 vs $2,650/month)
  - Azure Firewall deep dive
  - 3-year TCO comparison
  - Decision framework

### Support & Risk
- **[EKS_CILIUM_TRAPS.md](EKS_CILIUM_TRAPS.md)** ⚠️ **READ BEFORE EKS**
  - 10 major traps with EKS + Cilium
  - No official AWS support
  - Upgrade risks
  - Mitigation strategies

---

## 🔧 Technical Deep Dives

### Networking Fundamentals
- **[NETWORKING_DEEP_DIVE.md](NETWORKING_DEEP_DIVE.md)**
  - Overlay networking explained (Geneve/VXLAN)
  - Virtual service IPs (eBPF magic)
  - How packets flow
  - Performance analysis

### CNI Architecture
- **[AZURE_CNI_CILIUM_EXPLAINED.md](AZURE_CNI_CILIUM_EXPLAINED.md)**
  - AKS hybrid architecture
  - Control plane vs dataplane
  - IPAM ownership
  - Feature comparison

- **[EKS_VS_AKS_CNI.md](EKS_VS_AKS_CNI.md)**
  - CNI flexibility comparison
  - Pure Cilium vs hybrid
  - When to use each

### CIDR Planning
- **[CIDR_GUIDE.md](CIDR_GUIDE.md)**
  - Pod CIDR vs Service CIDR
  - Single subnet vs per-AZ subnets
  - CIDR allocation best practices
  - Avoid IP exhaustion

### VNet/VPC Connectivity
- **[VNET_PEERING_GUIDE.md](VNET_PEERING_GUIDE.md)**
  - Can pods ping across VNets?
  - When you need peering
  - Overlay network limitations
  - Cilium Cluster Mesh solution

---

## 🌍 Multi-Region Deployment

- **[MULTI_REGION_GUIDE.md](MULTI_REGION_GUIDE.md)**
  - Complete multi-region architecture
  - Azure Front Door / CloudFront
  - CIDR allocation across regions
  - Disaster recovery

---

## 📖 Reference Materials

### Quick References
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)**
  - Common commands
  - CIDR cheat sheet
  - Troubleshooting

### Original Research
- **[1230.md](1230.md)**
  - CIDR concepts
  - VPC/VNet planning
  - Per-AZ subnet pattern
  - Pulumi implementation examples

---

## 🎯 Decision Matrix

Use this flowchart to choose your platform:

```
Do you need official vendor support?
├─ Yes → AKS
│  ├─ Microsoft officially supports Cilium
│  ├─ Tested configurations
│  ├─ Enterprise SLA
│  └─ Cost: $2,650/month per region
│
└─ No → Consider EKS
   │
   Do you have Cilium/eBPF expertise?
   ├─ Yes → EKS
   │  ├─ Pure Cilium (all features)
   │  ├─ 5-6x cheaper
   │  ├─ More flexibility
   │  └─ Cost: $470/month per region
   │
   └─ No → Can you get Cilium Enterprise support?
      ├─ Yes → EKS + Cilium Enterprise
      │  ├─ Still cheaper than AKS
      │  ├─ Professional support
      │  └─ Total: ~$25k/year + infra
      │
      └─ No → AKS (safer choice)
         └─ Official support included
```

---

## 📋 Summary Comparison

| Aspect | AKS | EKS |
|--------|-----|-----|
| **Architecture** | Hybrid (Azure CNI + Cilium) | Pure Cilium |
| **Complexity** | 849 lines | 630 lines (26% simpler) |
| **Cost (single region)** | $2,650/month | $470/month (5.6x cheaper) |
| **Cost (3 regions)** | $8,073/month | $1,332/month (6.1x cheaper) |
| **Official Support** | ✅ Microsoft | ❌ Community only |
| **Cilium Features** | ⚠️ Dataplane only | ✅ Full control |
| **IPAM Control** | ❌ Azure controls | ✅ Cilium controls |
| **Cluster Mesh** | ⚠️ Limited | ✅ Full support |
| **BGP** | ⚠️ Limited | ✅ Full support |
| **Native Routing** | ❌ Always overlay | ✅ ENI mode available |
| **Setup Difficulty** | Medium | Medium |
| **Operational Risk** | Low (vendor support) | Higher (DIY) |

---

## 🚀 Getting Started

### New to Kubernetes + Cilium?
1. Read **[AKS_VS_EKS_COMPLEXITY.md](AKS_VS_EKS_COMPLEXITY.md)** first
2. Then **[COST_ANALYSIS.md](COST_ANALYSIS.md)**
3. If choosing EKS, read **[EKS_CILIUM_TRAPS.md](EKS_CILIUM_TRAPS.md)**
4. Follow platform-specific README in `aks/` or `eks/`

### Want to Understand Networking?
1. **[NETWORKING_DEEP_DIVE.md](NETWORKING_DEEP_DIVE.md)** - Core concepts
2. **[AZURE_CNI_CILIUM_EXPLAINED.md](AZURE_CNI_CILIUM_EXPLAINED.md)** - AKS specifics
3. **[VNET_PEERING_GUIDE.md](VNET_PEERING_GUIDE.md)** - Cross-region connectivity

### Planning Deployment?
1. **[CIDR_GUIDE.md](CIDR_GUIDE.md)** - Plan your IP ranges
2. **[MULTI_REGION_GUIDE.md](MULTI_REGION_GUIDE.md)** - Multi-region architecture
3. Platform-specific `example_usage.py`

---

## 💡 Key Takeaways

### AKS Advantages
- ✅ **Official Microsoft support** (biggest advantage)
- ✅ Managed Cilium updates
- ✅ Lower operational risk
- ✅ Faster time to production
- ✅ Better for enterprise/regulated industries

### EKS Advantages
- ✅ **5-6x cheaper** (biggest advantage)
- ✅ Pure Cilium architecture (simpler)
- ✅ Full Cilium feature set
- ✅ More networking flexibility
- ✅ Better for cost-conscious orgs

### The Real Trade-off
```
AKS: Pay $2,180/month extra for official support & lower risk
EKS: Save $26,160/year but need Cilium expertise & accept higher risk
```

---

## 🛠️ Tools & Technologies

Both platforms use:
- **Pulumi** - Infrastructure as Code
- **Cilium** - eBPF-based CNI
- **Argo CD** - GitOps deployment
- **SPIFFE/SPIRE** - Workload identity
- **Gateway API** - Modern ingress
- **Hubble** - Network observability

Platform-specific:
- **AKS**: Azure Firewall, Azure Front Door, Azure CNI
- **EKS**: NAT Gateway, CloudFront, Pure Cilium

---

## 📞 Support

### AKS
- **Official**: Azure Support tickets
- **Community**: Azure Kubernetes Slack
- **Docs**: Microsoft Learn

### EKS
- **Official**: AWS Support (VPC CNI only, not Cilium!)
- **Community**: Cilium Slack, eBPF Slack
- **Enterprise**: Cilium Enterprise Support ($10k-50k/year)
- **Docs**: Cilium.io, AWS EKS docs

---

## 🎓 Learning Resources

1. **Cilium Official Docs**: https://docs.cilium.io
2. **eBPF.io**: https://ebpf.io
3. **Kubernetes Networking**: https://kubernetes.io/docs/concepts/services-networking/
4. **This Repository**: All guides above! 📚

---

## ✅ Final Recommendation

**For most organizations:**
- **Start with AKS** if you need official support
- **Consider EKS** if you have expertise and budget constraints

**Best of both worlds:**
- **EKS + Cilium Enterprise support** ($25k/year)
- Still cheaper than AKS
- Get professional Cilium support
- Access to full features

**Both platforms are production-ready and excellent!** Your choice depends on your organization's needs, budget, and risk tolerance. 🚀

---

*Last Updated: 2025-12-30*
*Total Documentation: 15+ comprehensive guides*
*Combined Length: 10,000+ lines of detailed analysis*
