resource "helm_release" "cilium" {
  name       = "cilium"
  repository = "https://helm.cilium.io/"
  chart      = "cilium"
  version    = "1.14.0"
  namespace = "kube-system"
  create_namespace = false

  set {
    name  = "global.kubeProxyReplacement"
    value = "partial"
  }
  set {
    name  = "global.cni.chainingMode"
    value = "none"
  }

  depends_on = [azurerm_kubernetes_cluster.aks]
}
