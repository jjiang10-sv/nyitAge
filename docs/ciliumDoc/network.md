To remove all Cilium NetworkPolicies, you can follow any of these methods depending on your setup:

---

### ✅ **1. Remove All Cilium NetworkPolicies in All Namespaces**

```bash
kubectl delete ciliumnetworkpolicies --all --all-namespaces
```

This removes all CiliumNetworkPolicy (`CNP`) resources across your cluster.

---

### ✅ **2. Remove CiliumClusterwideNetworkPolicies (if any)**

Cluster-wide policies are not tied to a namespace:

```bash
kubectl delete ciliumclusterwidenetworkpolicies --all
```

---

### ✅ **3. Delete All Policies in a Specific Namespace**

If you only want to clear policies in a certain namespace:

```bash
kubectl get ciliumnetworkpolicies -n <namespace> -o name | xargs kubectl delete -n <namespace>
```

Example:

```bash
kubectl get ciliumnetworkpolicies -n default -o name | xargs kubectl delete -n default
```

---

### ✅ **4. If You Want to Remove All Cilium-Related CRDs (Advanced / Reset)**

⚠️ This will remove all Cilium CRDs (Cluster config, identities, etc.). Use only if you want to completely reset Cilium setup.

```bash
kubectl delete crd ciliumnetworkpolicies.cilium.io
kubectl delete crd ciliumclusterwidenetworkpolicies.cilium.io
kubectl delete crd ciliumendpoints.cilium.io
kubectl delete crd ciliumidentities.cilium.io
```

---

### 🎯 **Verify All Policies are Removed**

```bash
kubectl get ciliumnetworkpolicies --all-namespaces
kubectl get ciliumclusterwidenetworkpolicies
```

If nothing is listed, all Cilium network policies are successfully removed.

---

### ✅ **Bonus: Clean via Cilium CLI**

If you are using the Cilium CLI:

```bash
cilium policy delete --all
```

This must be executed on a Cilium agent pod or via `cilium-cli`.

---

Let me know if you want a script for automation or how to disable policy enforcement completely.
