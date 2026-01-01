#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y apt-transport-https ca-certificates curl software-properties-common

# Install containerd
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
apt-get update
apt-get install -y containerd.io

# Configure containerd for Kubernetes
mkdir -p /etc/containerd
containerd config default | tee /etc/containerd/config.toml
sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml
systemctl restart containerd
systemctl enable containerd

# Load kernel modules
cat <<EOF | tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

modprobe overlay
modprobe br_netfilter

# Sysctl params
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF

sysctl --system

# Install Kubernetes components
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.29/deb/Release.key | gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.29/deb/ /' | tee /etc/apt/sources.list.d/kubernetes.list
apt-get update
apt-get install -y kubelet=__KUBERNETES_VERSION__-1.1 kubeadm=__KUBERNETES_VERSION__-1.1 kubectl=__KUBERNETES_VERSION__-1.1
apt-mark hold kubelet kubeadm kubectl

# Disable swap
swapoff -a
sed -i '/ swap / s/^/#/' /etc/fstab

# Initialize cluster with kubeadm (will be run manually after instance creation)
# This is saved as a script for manual execution
cat <<'INITSCRIPT' > /root/init-cluster.sh
#!/bin/bash
kubeadm init \
  --pod-network-cidr=__POD_CIDR__ \
  --service-cidr=__SERVICE_CIDR__ \
  --skip-phases=addon/kube-proxy \
  --apiserver-advertise-address=$(hostname -I | awk '{print $1}')

# Setup kubectl for root
mkdir -p /root/.kube
cp /etc/kubernetes/admin.conf /root/.kube/config

# Install Cilium
CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt)
CLI_ARCH=amd64
curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}
sha256sum --check cilium-linux-${CLI_ARCH}.tar.gz.sha256sum
tar xzvfC cilium-linux-${CLI_ARCH}.tar.gz /usr/local/bin
rm cilium-linux-${CLI_ARCH}.tar.gz{,.sha256sum}

# Install Cilium with native routing and mTLS
cilium install \
  --version __CILIUM_VERSION__ \
  --set ipam.mode=cluster-pool \
  --set ipam.operator.clusterPoolIPv4PodCIDRList=__POD_CIDR__ \
  --set tunnel=disabled \
  --set ipv4NativeRoutingCIDR=__POD_CIDR__ \
  --set kubeProxyReplacement=strict \
  --set hubble.relay.enabled=true \
  --set hubble.ui.enabled=true \
  --set authentication.mutual.spire.enabled=true \
  --set authentication.mutual.spire.install.enabled=false \
  --set authentication.mutual.spire.serverAddress=spire-server.spire:8081 \
  --set authentication.mutual.spire.trustDomain=cluster.local

echo "Waiting for Cilium to be ready..."
kubectl wait --for=condition=ready pod -l k8s-app=cilium -n kube-system --timeout=300s

# Install AWS EBS CSI Driver for persistent storage
echo "Installing AWS EBS CSI Driver..."
kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.26"

# Wait for CSI driver to be ready
echo "Waiting for EBS CSI Driver..."
sleep 30
kubectl wait --for=condition=ready pod -l app=ebs-csi-controller -n kube-system --timeout=300s

# Create default StorageClass using EBS gp3
echo "Creating default StorageClass..."
cat <<'STORAGECLASS' | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ebs-sc
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
provisioner: ebs.csi.aws.com
parameters:
  type: gp3
  encrypted: "true"
  iops: "3000"
  throughput: "125"
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
reclaimPolicy: Delete
STORAGECLASS

# Install SPIFFE/SPIRE for workload identity
echo "Installing SPIFFE/SPIRE..."
kubectl create namespace spire || true
kubectl create serviceaccount spire-server -n spire || true
kubectl create serviceaccount spire-agent -n spire || true

# Install SPIRE Server
kubectl apply -f - <<'SPIRESERVER'
apiVersion: v1
kind: ConfigMap
metadata:
  name: spire-server
  namespace: spire
data:
  server.conf: |
    server {
      bind_address = "0.0.0.0"
      bind_port = "8081"
      trust_domain = "cluster.local"
      data_dir = "/run/spire/data"
      log_level = "INFO"
      ca_subject = {
        country = ["US"]
        organization = ["Kubernetes"]
        common_name = "SPIRE Server"
      }
    }
    plugins {
      DataStore "sql" {
        plugin_data {
          database_type = "sqlite3"
          connection_string = "/run/spire/data/datastore.sqlite3"
        }
      }
      NodeAttestor "k8s_psat" {
        plugin_data {
          clusters = {
            "cluster.local" = {
              service_account_allow_list = ["spire:spire-agent"]
            }
          }
        }
      }
      KeyManager "disk" {
        plugin_data {
          keys_path = "/run/spire/data/keys.json"
        }
      }
      Notifier "k8sbundle" {
        plugin_data {}
      }
    }
    health_checks {
      listener_enabled = true
      bind_address = "0.0.0.0"
      bind_port = "8080"
      live_path = "/live"
      ready_path = "/ready"
    }
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: spire-server
  namespace: spire
spec:
  replicas: 1
  selector:
    matchLabels:
      app: spire-server
  serviceName: spire-server
  template:
    metadata:
      labels:
        app: spire-server
    spec:
      serviceAccountName: spire-server
      containers:
      - name: spire-server
        image: ghcr.io/spiffe/spire-server:1.8.0
        args: ["-config", "/run/spire/config/server.conf"]
        ports:
        - containerPort: 8081
        volumeMounts:
        - name: spire-config
          mountPath: /run/spire/config
          readOnly: true
        - name: spire-data
          mountPath: /run/spire/data
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
          initialDelaySeconds: 15
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
      volumes:
      - name: spire-config
        configMap:
          name: spire-server
  volumeClaimTemplates:
  - metadata:
      name: spire-data
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: ebs-sc
      resources:
        requests:
          storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: spire-server
  namespace: spire
spec:
  ports:
  - name: grpc
    port: 8081
  selector:
    app: spire-server
SPIRESERVER

echo "Waiting for SPIRE Server..."
sleep 20
kubectl wait --for=condition=ready pod -l app=spire-server -n spire --timeout=300s || true

# Install SPIRE Agent  
kubectl apply -f - <<'SPIREAGENT'
apiVersion: v1
kind: ConfigMap
metadata:
  name: spire-agent
  namespace: spire
data:
  agent.conf: |
    agent {
      data_dir = "/run/spire"
      log_level = "INFO"
      server_address = "spire-server"
      server_port = "8081"
      trust_domain = "cluster.local"
    }
    plugins {
      NodeAttestor "k8s_psat" {
        plugin_data {
          cluster = "cluster.local"
        }
      }
      KeyManager "memory" {
        plugin_data {}
      }
      WorkloadAttestor "k8s" {
        plugin_data {
          skip_kubelet_verification = true
        }
      }
    }
    health_checks {
      listener_enabled = true
      bind_address = "0.0.0.0"
      bind_port = "8080"
      live_path = "/live"
      ready_path = "/ready"
    }
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: spire-agent
  namespace: spire
spec:
  selector:
    matchLabels:
      app: spire-agent
  template:
    metadata:
      labels:
        app: spire-agent
    spec:
      hostPID: true
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet
      serviceAccountName: spire-agent
      initContainers:
      - name: wait-spire-server
        image: busybox:1.35
        command: ['sh', '-c', 'until nslookup spire-server.spire.svc.cluster.local; do sleep 2; done']
      containers:
      - name: spire-agent
        image: ghcr.io/spiffe/spire-agent:1.8.0
        args: ["-config", "/run/spire/config/agent.conf"]
        volumeMounts:
        - name: spire-config
          mountPath: /run/spire/config
          readOnly: true
        - name: spire-agent-socket
          mountPath: /run/spire/sockets
        - name: spire-token
          mountPath: /var/run/secrets/tokens
        livenessProbe:
          httpGet:
            path: /live
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
      volumes:
      - name: spire-config
        configMap:
          name: spire-agent
      - name: spire-agent-socket
        hostPath:
          path: /run/spire/sockets
          type: DirectoryOrCreate
      - name: spire-token
        projected:
          sources:
          - serviceAccountToken:
              path: spire-agent
              expirationSeconds: 7200
              audience: spire-server
SPIREAGENT

echo "Waiting for SPIRE Agents..."
sleep 15
kubectl wait --for=condition=ready pod -l app=spire-agent -n spire --timeout=300s || true

echo ""
echo "========================================"
echo "Cluster initialized successfully!"
echo "========================================"
echo ""
echo "Installed components:"
echo "- Kubernetes __KUBERNETES_VERSION__"
echo "- Cilium __CILIUM_VERSION__ (with mTLS support)"
echo "- AWS EBS CSI Driver"
echo "- Default StorageClass: ebs-sc"
echo "- SPIFFE/SPIRE (workload identity)"
echo ""
echo "SPIRE Configuration:"
echo "- Trust Domain: cluster.local"
echo "- Server: spire-server.spire:8081"
echo "- Cilium mTLS: ENABLED"
echo ""
echo "Join command:"
kubeadm token create --print-join-command
INITSCRIPT

chmod +x /root/init-cluster.sh

echo "Control plane node ready. Run: sudo /root/init-cluster.sh"
