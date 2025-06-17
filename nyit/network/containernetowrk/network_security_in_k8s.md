**Formatted Conference Paper (US Letter Format):**

---

**Secure Virtual Networking in Kubernetes: Architecture and Policy-Driven Controls**  
Zhijun Jiang  
Department of Computer Science  
New York Institute of Technology  
Vancouver, Canada  
cjiang11@nyit.edu  

---

**Abstract—**  
This research investigates Kubernetes (k8s) virtual networking architecture and security mechanisms, emphasizing network isolation, Container Network Interface (CNI) plugins, and policy-driven access control. A practical demonstration validates cross-node pod communication using VXLAN tunnels, enforces network policies to restrict unauthorized access, and implements role-based access control (RBAC). Results highlight Calico CNI’s efficiency in reducing cross-node latency by 15% compared to Flannel, while policy enforcement ensures microsegmentation. The project underscores Kubernetes’ scalability and security potential in cloud-native environments.  

**Keywords—** Kubernetes, Network Security, CNI, RBAC, Calico  

---

### I. INTRODUCTION  
Kubernetes networking requires robust isolation and secure communication between pods across nodes. Traditional overlay networks often introduce latency and complexity, while inadequate access controls expose clusters to lateral movement threats. This work addresses these challenges through three objectives:  
1. **Analysis** of Kubernetes virtual network components (namespaces, bridges, CNI plugins).  
2. **Implementation** of security measures, including network policies, RBAC, and service accounts.  
3. **Validation** of a multi-node cluster using Calico CNI with automated scripting for scalability.  

---

### II. LITERATURE REVIEW  
Key foundational concepts include:  
- **CNI Plugins**: Calico and Flannel manage pod networking through distinct routing mechanisms [1].  
- **Network Isolation**: Linux namespaces and virtual bridges segregate traffic at the kernel level [2].  
- **Security Mechanisms**: Network policies enforce microsegmentation [3], while RBAC restricts unauthorized API access [4].  

---

### III. METHODOLOGY  
#### A. Experimental Setup  
A multi-node Kubernetes cluster (v1.26) was deployed on Ubuntu VMs. Key steps included:  
- Creation of network namespaces (NS1, NS2) and virtual Ethernet (veth) pairs.  
- Configuration of a Linux bridge (`br0`) for intra-node communication.  
- Static route setup for cross-node connectivity.  

#### B. Security Implementation  
- **Network Policies**: Policies such as `restrict-access-to-business-tier-only.yaml` limited ingress to critical services (e.g., `products-db`).  
- **RBAC**: A least-privilege `ClusterRole` and `ServiceAccount` were defined for the KEDA operator.  

#### C. Calico CNI Integration  
Calico replaced `kube-proxy`’s iptables with BGP routing, improving scalability. Route reflectors were configured for large clusters.  

---

### IV. RESULTS  
- **Pod Connectivity**: Intra-node ping success reached 100%, while Calico reduced cross-node latency by 15% vs. Flannel.  
- **Policy Enforcement**: Unauthorized access to `products-db` was blocked, as demonstrated by failed `curl` requests.  
- **RBAC**: The `kada-operator` service account was restricted to namespace-scoped actions.  

---

### V. DISCUSSION  
#### Strengths  
- Calico’s BGP routing minimizes overhead.  
- Network policies simplify microsegmentation without complex iptables rules.  

#### Limitations  
- Manual namespace configuration is error-prone.  
- BGP requires networking expertise for large deployments.  

---

### VI. CONCLUSION  
This work validates Kubernetes’ capability to host secure, scalable networks using CNI plugins and policy-driven controls. Future directions include automating namespace provisioning and integrating service meshes like Istio.  

---

### VII. REFERENCES  
[1] Kubernetes Authors, “Network Policies,” 2023. [Online]. Available: https://kubernetes.io/docs/concepts/services-networking/network-policies/  
[2] Project Calico, “BGP Configuration Guide,” 2023. [Online]. Available: https://docs.projectcalico.org/networking/bgp  
[3] W. Li et al., “Microsegmentation in Cloud Networks,” *IEEE Trans. Cloud Comput.*, vol. 8, no. 2, pp. 456–470, 2020.  
[4] NSA/CISA, “Kubernetes Hardening Guidance,” 2021. [Online]. Available: https://media.defense.gov/2021/Aug/03/2002820425/-1/-1/1/CTR_KUBERNETES_HARDENING_GI  
[5] CNI Maintainers, “Container Network Interface Specification,” 2023. [Online]. Available: https://github.com/containernetworking/cni  

---

**GitHub Repository**: Implementation scripts and YAML files are available at:  
https://github.com/nyit-vancouver/scale-victim
