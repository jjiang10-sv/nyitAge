# Kubernetes Persistent Storage Guide for Self-Managed Clusters on AWS

## Do You Need Storage? YES! ✅

**Without storage, you can't run:**
- Databases (PostgreSQL, MySQL, MongoDB)
- Message queues (RabbitMQ, Kafka)
- CI/CD (Jenkins, GitLab)
- Monitoring (Prometheus, Grafana)
- Any stateful application

---

## Storage Options for AWS

### Option 1: AWS EBS CSI Driver ⭐ **RECOMMENDED**

**Best for:** Most use cases, production workloads

```yaml
Pros:
✅ AWS-native (reliable, fast)
✅ Automated by CSI driver
✅ SSD performance (gp3, io2)
✅ Snapshots and backups
✅ Encryption at rest
✅ Simple to set up

Cons:
❌ Single-AZ only (volume attached to one node)
❌ ReadWriteOnce only (one pod at a time)
```

**Use for:**
- Databases
- Most stateful apps
- High-performance workloads

---

### Option 2: AWS EFS CSI Driver

**Best for:** Shared storage across zones

```yaml
Pros:
✅ Multi-AZ (replicated across zones)
✅ ReadWriteMany (multiple pods)
✅ Shared file system
✅ Auto-scaling storage

Cons:
❌ Slower than EBS
❌ More expensive
❌ Network latency
```

**Use for:**
- Shared data between pods
- Multi-AZ applications
- File sharing

---

### Option 3: Local Path Provisioner

**Best for:** Development, testing

```yaml
Pros:
✅ Very simple
✅ Fast (local disk)
✅ No cloud costs

Cons:
❌ Data lost if node fails
❌ Not for production
❌ Single node only
```

**Use for:**
- Development clusters
- Testing
- Temporary data

---

### Option 4: NFS (What You Showed)

**Best for:** Legacy apps, specific use cases

```yaml
Pros:
✅ Works anywhere
✅ Well-understood
✅ Simple protocol

Cons:
❌ Not AWS-native
❌ You manage NFS server
❌ Performance overhead
❌ Single point of failure
```

**Use for:**
- Legacy applications that require NFS
- When you already have NFS expertise

---

## Recommended: AWS EBS CSI Driver

### Installation Steps

#### 1. Install AWS EBS CSI Driver

Add to your control plane init script:

```bash
# After Cilium installation, add:

# Install AWS EBS CSI Driver
kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.26"

# Create StorageClass
cat <<EOF | kubectl apply -f -
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
volumeBindingMode: WaitForFirstConsumer
allowVolumeExpansion: true
EOF
```

#### 2. Update IAM Role

Your nodes need these permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateSnapshot",
        "ec2:AttachVolume",
        "ec2:DetachVolume",
        "ec2:ModifyVolume",
        "ec2:DescribeAvailabilityZones",
        "ec2:DescribeInstances",
        "ec2:DescribeSnapshots",
        "ec2:DescribeTags",
        "ec2:DescribeVolumes",
        "ec2:DescribeVolumesModifications"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateTags"
      ],
      "Resource": [
        "arn:aws:ec2:*:*:volume/*",
        "arn:aws:ec2:*:*:snapshot/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DeleteTags"
      ],
      "Resource": [
        "arn:aws:ec2:*:*:volume/*",
        "arn:aws:ec2:*:*:snapshot/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateVolume"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "aws:RequestTag/ebs.csi.aws.com/cluster": "true"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateVolume"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "aws:RequestTag/CSIVolumeName": "*"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DeleteVolume"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "ec2:ResourceTag/ebs.csi.aws.com/cluster": "true"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DeleteVolume"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "ec2:ResourceTag/CSIVolumeName": "*"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DeleteVolume"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "ec2:ResourceTag/kubernetes.io/created-for/pvc/name": "*"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DeleteSnapshot"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "ec2:ResourceTag/CSIVolumeSnapshotName": "*"
        }
      }
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DeleteSnapshot"
      ],
      "Resource": "*",
      "Condition": {
        "StringLike": {
          "ec2:ResourceTag/ebs.csi.aws.com/cluster": "true"
        }
      }
    }
  ]
}
```

#### 3. Test Storage

```bash
# Create PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-pvc
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: ebs-sc
  resources:
    requests:
      storage: 10Gi
EOF

# Create pod using PVC
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: app
    image: ubuntu
    command: ["/bin/sh"]
    args: ["-c", "echo 'Hello' > /data/test.txt && sleep 3600"]
    volumeMounts:
    - name: data
      mountPath: /data
  volumes:
  - name: data
    persistentVolumeClaim:
      claimName: test-pvc
EOF

# Verify
kubectl get pvc
kubectl get pv
kubectl exec test-pod -- cat /data/test.txt
```

---

## Complete Setup for Each Option

### Setup 1: AWS EBS CSI (Automated via Pulumi)

I'll update `platform.py` to include this option.

### Setup 2: AWS EFS CSI

```bash
# Create EFS file system
aws efs create-file-system \
  --performance-mode generalPurpose \
  --throughput-mode bursting \
  --encrypted \
  --tags Key=Name,Value=k8s-efs

# Get file system ID
EFS_ID=$(aws efs describe-file-systems --query 'FileSystems[0].FileSystemId' --output text)

# Create mount targets (one per AZ)
for SUBNET in subnet-xxx subnet-yyy subnet-zzz; do
  aws efs create-mount-target \
    --file-system-id $EFS_ID \
    --subnet-id $SUBNET \
    --security-groups sg-xxx
done

# Install EFS CSI Driver
kubectl apply -k "github.com/kubernetes-sigs/aws-efs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.7"

# Create StorageClass
cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: efs-sc
provisioner: efs.csi.aws.com
parameters:
  provisioningMode: efs-ap
  fileSystemId: ${EFS_ID}
  directoryPerms: "700"
EOF
```

### Setup 3: Local Path Provisioner (Simple!)

```bash
# Install local-path-provisioner
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.26/deploy/local-path-storage.yaml

# Set as default
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

# Test
kubectl create -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/examples/pvc/pvc.yaml
kubectl create -f https://raw.githubusercontent.com/rancher/local-path-provisioner/master/examples/pod/pod.yaml
```

### Setup 4: NFS (What You Showed)

Your approach works, but here's the improved version:

```bash
# On a dedicated NFS server node (or control plane for dev)
sudo apt install nfs-kernel-server -y
sudo mkdir -p /export/volumes/dynamic
sudo bash -c 'echo "/export/volumes  *(rw,no_root_squash,no_subtree_check)" > /etc/exports'
sudo exportfs -a
sudo systemctl restart nfs-kernel-server

# On Kubernetes cluster
# Install NFS CSI driver (better than subdir-provisioner)
helm repo add csi-driver-nfs https://raw.githubusercontent.com/kubernetes-csi/csi-driver-nfs/master/charts
helm install csi-driver-nfs csi-driver-nfs/csi-driver-nfs \
  --namespace kube-system \
  --set kubeletDir=/var/lib/kubelet

# Create StorageClass
cat <<EOF | kubectl apply -f -
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: nfs-csi
provisioner: nfs.csi.k8s.io
parameters:
  server: NFS_SERVER_IP
  share: /export/volumes/dynamic
reclaimPolicy: Delete
volumeBindingMode: Immediate
mountOptions:
  - hard
  - nfsvers=4.1
EOF
```

---

## Which One Should You Use?

### Decision Matrix

| Use Case |  Recommendation | Why |
|----------|----------------|-----|
| **PostgreSQL** | EBS CSI | High performance, single-AZ is fine |
| **MySQL** | EBS CSI | Same as PostgreSQL |
| **MongoDB** | EBS CSI | Block storage, good IOPS |
| **Kafka** | EBS CSI | High throughput, local to brokers |
| **Prometheus** | EBS CSI | Time-series data, fast writes |
| **Shared Files** | EFS CSI | Multi-AZ, ReadWriteMany |
| **WordPress** | EFS CSI | Multiple pods need same files |
| **GitLab** | EFS CSI | Shared repositories |
| **Development** | Local Path | Fast, simple, free |
| **Testing** | Local Path | Don't need persistence |

### My Recommendation

**For Production:**
```bash
Primary: AWS EBS CSI Driver
Fallback: AWS EFS CSI Driver (for shared storage)
```

**For Development:**
```bash
Local Path Provisioner (sufficient for testing)
```

---

## Cost Comparison

### EBS Volumes (gp3)

```
10GB volume: $0.80/month
100GB volume: $8.00/month
1TB volume: $80.00/month

IOPS: 3,000 baseline (free)
Additional IOPS: $0.005/IOPS/month
Throughput: 125 MB/s baseline (free)
```

### EFS

```
Standard storage: $0.30/GB-month
Infrequent Access: $0.025/GB-month

Example:
100GB: $30/month
1TB: $300/month

More expensive than EBS!
```

### NFS (DIY)

```
NFS server EC2: ~$30-60/month (t3.medium)
EBS volume for NFS: $8-80/month
Total: $40-140/month

But: You manage it!
```

### Local Path

```
Free! (uses node disk)
```

---

## Updated platform.py

I'll add EBS CSI driver support to the platform:

```python
# In control plane init script, add after Cilium install:

# Install AWS EBS CSI Driver
kubectl apply -k "github.com/kubernetes-sigs/aws-ebs-csi-driver/deploy/kubernetes/overlays/stable/?ref=release-1.26"

# Wait for driver to be ready
kubectl wait --for=condition=ready pod -l app=ebs-csi-controller -n kube-system --timeout=300s

# Create default StorageClass
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
```

---

## Summary

### What You Need

**Minimum (Development):**
```bash
✅ Local Path Provisioner (free, simple)
```

**Recommended (Production):**
```bash
✅ AWS EBS CSI Driver (reliable, fast, AWS-native)
```

**For Shared Storage:**
```bash
✅ AWS EFS CSI Driver (multi-AZ, ReadWriteMany)
```

**NFS Option:**
```yaml
Status: Works but not recommended
Reason: Not AWS-native, you manage server
Use only if: Legacy requirement
```

### Next Steps

1. Choose your storage backend
2. I'll update `platform.py` to include it
3. Deploy and test

**What storage do you prefer? EBS (recommended), EFS, or both?**
