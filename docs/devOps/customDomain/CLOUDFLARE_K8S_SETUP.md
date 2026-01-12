# Cloudflare DNS/CDN with AWS Kubernetes: Complete Setup Guide

## Architecture Overview

```
User Request
    ↓
Cloudflare DNS (ns1.cloudflare.com)
    ↓ [Resolves to:]
Cloudflare CDN (if proxy enabled)
    ↓
AWS Load Balancer (ALB/NLB)
    ↓
Kubernetes Gateway (GatewayClass)
    ↓
HTTPRoute (routing rules)
    ↓
Kubernetes Services
    ↓
Pods
```

**Key Points:**
- Cloudflare DNS resolves your domain to AWS load balancer
- Cloudflare CDN (optional) sits between users and AWS
- Kubernetes services exposed via **Gateway API** (successor to Ingress)
- Traffic flow: Cloudflare → AWS LB → Gateway → HTTPRoute → Service → Pods
- Gateway API provides better control, observability, and extensibility

---

## Setup: Three Main Approaches

### Approach 1: DNS Only (Gray Cloud) ☁️

**Best for:** APIs, WebSocket servers, dynamic content

```
Cloudflare: DNS only (gray cloud)
    ↓ [DNS resolves to AWS LB IP]
User → AWS Load Balancer → K8s
```

**Pros:**
- ✅ Lower latency (no extra hop)
- ✅ WebSocket works perfectly
- ✅ Real client IPs preserved

**Cons:**
- ❌ No Cloudflare CDN caching
- ❌ No Cloudflare DDoS protection
- ❌ AWS load balancer exposed

### Approach 2: Proxy Enabled (Orange Cloud) ☁️

**Best for:** Static sites, public APIs, websites

```
Cloudflare: Proxy enabled (orange cloud)
    ↓
User → Cloudflare Edge → AWS LB → K8s
```

**Pros:**
- ✅ Cloudflare DDoS protection
- ✅ Cloudflare WAF
- ✅ CDN caching
- ✅ Origin IP hidden

**Cons:**
- ❌ Extra latency hop
- ❌ WebSocket requires special setup
- ❌ Client IPs need restoration

### Approach 3: Hybrid (Recommended)

**Best for:** Production applications

```
Static assets: Cloudflare CDN (orange cloud)
    → S3 or CloudFront origin

Dynamic API: Cloudflare DNS only (gray cloud)
    → AWS Load Balancer → K8s
```

---

## Step-by-Step Setup

### Prerequisites

```bash
# Tools needed
- kubectl (Kubernetes CLI)
- helm (Kubernetes package manager)
- aws CLI (AWS command line)
- Cloudflare account with domain added
```

### Step 1: Deploy Kubernetes Cluster (EKS)

```bash
# Create EKS cluster
eksctl create cluster \
  --name production-cluster \
  --region us-east-1 \
  --nodegroup-name standard-workers \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 1 \
  --nodes-max 4 \
  --managed

# Verify cluster
kubectl get nodes
```

### Step 2: Install Gateway API CRDs and Gateway Controller

#### Install Gateway API CRDs

```bash
# Install Gateway API CRDs (v1.0.0 or latest)
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml

# Verify CRDs are installed
kubectl get crd | grep gateway
# Expected output:
# gatewayclasses.gateway.networking.k8s.io
# gateways.gateway.networking.k8s.io
# httproutes.gateway.networking.k8s.io
```

#### Install NGINX Gateway Fabric

```yaml
# nginx-gateway-values.yaml
nginxGateway:
  service:
    annotations:
      # Create AWS Network Load Balancer
      service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
      service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
      
      # Enable Proxy Protocol to preserve client IPs
      service.beta.kubernetes.io/aws-load-balancer-proxy-protocol: "*"
    
    type: LoadBalancer
    
  config:
    # Trust Cloudflare IPs (if using orange cloud)
    use-forwarded-headers: "true"
    compute-full-forwarded-for: "true"
    use-proxy-protocol: "true"
```

```bash
# Add NGINX Gateway Fabric Helm repository
helm repo add nginx-stable https://helm.nginx.com/stable
helm repo update

# Install NGINX Gateway Fabric
helm install nginx-gateway nginx-stable/nginx-gateway \
  --namespace nginx-gateway \
  --create-namespace \
  --values nginx-gateway-values.yaml

# Verify installation
kubectl get pods -n nginx-gateway
kubectl get gatewayclass

# Get Load Balancer hostname
kubectl get svc -n nginx-gateway nginx-gateway

# Output example:
# NAME            TYPE           EXTERNAL-IP
# nginx-gateway   LoadBalancer   a1b2c3-123456789.us-east-1.elb.amazonaws.com
```

#### Alternative: AWS Load Balancer Controller with Gateway API

```bash
# Install AWS Load Balancer Controller (for ALB support)
helm repo add eks https://aws.github.io/eks-charts
helm repo update

# Create IAM policy and service account (prerequisite)
# See: https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html

helm install aws-load-balancer-controller eks/aws-load-balancer-controller \
  -n kube-system \
  --set clusterName=production-cluster \
  --set serviceAccount.create=false \
  --set serviceAccount.name=aws-load-balancer-controller
```

### Step 3: Create DNS Records in Cloudflare

#### Get Load Balancer IP/Hostname

```bash
# Get LB hostname
LB_HOSTNAME=$(kubectl get svc -n nginx-gateway nginx-gateway \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo "Load Balancer: $LB_HOSTNAME"

# For NLB, you'll get an IP address instead
LB_IP=$(kubectl get svc -n nginx-gateway nginx-gateway \
  -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
```

#### Option A: CNAME Record (if using ALB/ELB)

```
Type: CNAME
Name: api (or @ for apex domain)
Target: a1b2c3-123456789.us-east-1.elb.amazonaws.com
Proxy: OFF (gray cloud) - recommended for K8s
TTL: Auto
```

**Note:** Cloudflare doesn't allow CNAME on apex (@) with proxy OFF. Use CNAME Flattening or A record.

#### Option B: A Record (if using NLB with static IP)

```
Type: A
Name: @ (apex) or subdomain
Target: 52.45.123.45 (NLB IP)
Proxy: OFF (gray cloud)
TTL: Auto
```

#### Option C: CNAME Flattening (Cloudflare feature)

```
Type: CNAME
Name: @ (apex domain)
Target: a1b2c3-123456789.us-east-1.elb.amazonaws.com
Proxy: ON (orange cloud)

Cloudflare automatically flattens CNAME to A record
```

### Step 4: Deploy Sample Application

```yaml
# app-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-app
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: hello
  template:
    metadata:
      labels:
        app: hello
    spec:
      containers:
      - name: hello
        image: gcr.io/google-samples/hello-app:1.0
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: hello-service
  namespace: default
spec:
  selector:
    app: hello
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP  # Internal only, not exposed
```

```bash
kubectl apply -f app-deployment.yaml
```

### Step 5: Create Gateway and HTTPRoute Resources

#### Create Gateway

```yaml
# gateway.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: production-gateway
  namespace: default
spec:
  gatewayClassName: nginx  # Or 'aws-load-balancer' for AWS LB Controller
  listeners:
  - name: http
    protocol: HTTP
    port: 80
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: api-tls-secret  # SSL cert (optional if Cloudflare proxy ON)
  addresses:
  - type: NamedAddress
    value: api.yourdomain.com
```

```bash
kubectl apply -f gateway.yaml

# Check Gateway status
kubectl get gateway production-gateway
kubectl describe gateway production-gateway
```

#### Create HTTPRoute

```yaml
# httproute.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: hello-route
  namespace: default
spec:
  parentRefs:
  - name: production-gateway
    namespace: default
  hostnames:
  - "api.yourdomain.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /
    backendRefs:
    - name: hello-service
      port: 80
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        set:
        - name: X-Forwarded-Proto
          value: https
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        set:
        - name: X-Content-Type-Options
          value: nosniff
        - name: X-Frame-Options
          value: DENY
```

```bash
kubectl apply -f httproute.yaml

# Check HTTPRoute status
kubectl get httproute hello-route
kubectl describe httproute hello-route
```

#### Advanced HTTPRoute with Multiple Rules

```yaml
# advanced-httproute.yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: advanced-route
  namespace: default
spec:
  parentRefs:
  - name: production-gateway
  hostnames:
  - "api.yourdomain.com"
  rules:
  # Rule 1: API v1 with rate limiting
  - matches:
    - path:
        type: PathPrefix
        value: /api/v1
    backendRefs:
    - name: api-v1-service
      port: 80
    filters:
    - type: URLRewrite
      urlRewrite:
        path:
          type: ReplacePrefixMatch
          replacePrefixMatch: /v1
  
  # Rule 2: API v2
  - matches:
    - path:
        type: PathPrefix
        value: /api/v2
    backendRefs:
    - name: api-v2-service
      port: 80
  
  # Rule 3: Redirect HTTP to HTTPS
  - matches:
    - headers:
      - type: Exact
        name: X-Forwarded-Proto
        value: http
    filters:
    - type: RequestRedirect
      requestRedirect:
        scheme: https
        statusCode: 301
```

### Step 6: SSL Certificate Setup

#### Option A: Cloudflare Proxy ON (Recommended)

```
Cloudflare handles SSL automatically!

SSL Mode: Full (strict)
- User → Cloudflare: Cloudflare SSL cert
- Cloudflare → Origin: Origin needs valid cert

No need for cert-manager in K8s!
```

#### Option B: Cloudflare Proxy OFF (Need cert-manager)

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create Let's Encrypt ClusterIssuer
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        gatewayHTTPRoute:
          parentRefs:
          - name: production-gateway
            namespace: default
EOF

# Create Certificate resource
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: api-tls-cert
  namespace: default
spec:
  secretName: api-tls-secret
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - api.yourdomain.com
EOF
```

---

## DNS Record Patterns for Different Services

### Pattern 1: Single Service (Simple)

```yaml
# DNS Record
Type: A
Name: api
Value: <NLB_IP>
Proxy: OFF

# Gateway + HTTPRoute
Gateway: production-gateway
HTTPRoute hostname: api.yourdomain.com
  → Service: hello-service
```

### Pattern 2: Multiple Subdomains

```yaml
# DNS Records
api.yourdomain.com → CNAME → <LB_hostname>
admin.yourdomain.com → CNAME → <LB_hostname>
www.yourdomain.com → CNAME → <LB_hostname>

# HTTPRoute (single route, multiple hostnames)
hostnames:
- api.yourdomain.com → api-service
- host: admin.yourdomain.com
  paths: [/] → admin-service  
- host: www.yourdomain.com
  paths: [/] → web-service
```

### Pattern 3: Path-Based Routing

```yaml
# DNS Record (single)
app.yourdomain.com → A → <NLB_IP>

# Ingress (single host, multiple paths)
host: app.yourdomain.com
  paths:
  - /api → backend-service
  - /web → frontend-service
  - /admin → admin-service
```

### Pattern 4: Wildcard Subdomain

```yaml
# DNS Record
Type: A
Name: *
Value: <NLB_IP>
Proxy: OFF

# Ingress with wildcard
rules:
- host: "*.apps.yourdomain.com"
  paths: [/] → dynamic-service

# Requests
tenant1.apps.yourdomain.com → works
tenant2.apps.yourdomain.com → works
```

---

## Restoring Real Client IPs (Cloudflare Proxy ON)

### Problem

```
Without configuration:
User (1.2.3.4) → Cloudflare → K8s sees Cloudflare IP (104.16.x.x)

Application logs: 104.16.132.45 (Cloudflare, not user!)
```

### Solution 1: Use X-Forwarded-For Header

```yaml
# Ingress annotation
nginx.ingress.kubernetes.io/use-forwarded-headers: "true"
nginx.ingress.kubernetes.io/compute-full-forwarded-for: "true"

# In your application
const clientIP = req.headers['cf-connecting-ip'] ||  // Cloudflare specific
                 req.headers['x-forwarded-for']?.split(',')[0] ||
                 req.connection.remoteAddress;
```

### Solution 2: Trust Cloudflare IPs

```yaml
# ConfigMap for NGINX
apiVersion: v1
kind: ConfigMap
metadata:
  name: nginx-configuration
  namespace: ingress-nginx
data:
  # Trust Cloudflare IP ranges
  use-forwarded-headers: "true"
  proxy-real-ip-cidr: "173.245.48.0/20,103.21.244.0/22,103.22.200.0/22,103.31.4.0/22,141.101.64.0/18,108.162.192.0/18,190.93.240.0/20,188.114.96.0/20,197.234.240.0/22,198.41.128.0/17,162.158.0.0/15,104.16.0.0/13,104.24.0.0/14,172.64.0.0/13,131.0.72.0/22"
```

---

## Monitoring & Troubleshooting

### Check DNS Resolution

```bash
# Test DNS
dig api.yourdomain.com

# Should return:
# Cloudflare proxy ON: 104.16.x.x (Cloudflare IP)
# Cloudflare proxy OFF: AWS LB IP/hostname
```

### Check Load Balancer

```bash
# Get LB details
kubectl describe svc -n ingress-nginx ingress-nginx-controller

# Test direct connection
curl http://<LB_HOSTNAME>
```

### Check Ingress

```bash
# List ingresses
kubectl get ingress -A

# Describe ingress
kubectl describe ingress hello-ingress

# Check logs
kubectl logs -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx
```

### Common Issues

**1. 502 Bad Gateway**
```bash
# Check pod health
kubectl get pods
kubectl logs <pod-name>

# Check service endpoints
kubectl get endpoints hello-service
```

**2. SSL Certificate Errors**
```bash
# Check cert-manager (if used)
kubectl get certificate
kubectl describe certificate api-tls-secret

# Check Cloudflare SSL mode
# Should be "Full" or "Full (strict)"
```

**3. Infinite Redirect Loop**
```
Problem: Cloudflare → HTTP → NGINX → redirects to HTTPS → loop

Solution: Disable SSL redirect in Ingress
nginx.ingress.kubernetes.io/ssl-redirect: "false"

Cloudflare already handles HTTPS termination
```

---

## Production Best Practices

### 1. Use Separate Subdomains

```
Static assets: cdn.yourdomain.com → S3/CloudFront
API: api.yourdomain.com → K8s (gray cloud)
Admin: admin.yourdomain.com → K8s + Cloudflare Access
```

### 2. Enable Cloudflare Argo Smart Routing

```
Cloudflare Dashboard → Traffic → Argo Smart Routing

Cost: $5/month + $0.10/GB
Benefit: 30% faster by routing through Cloudflare's optimized backbone
```

### 3. Configure Cloudflare Page Rules

```
Rule 1: Cache static paths
URL: api.yourdomain.com/static/*
Settings: Cache Everything, Edge Cache TTL: 1 month

Rule 2: Bypass cache for API
URL: api.yourdomain.com/api/*
Settings: Cache Level: Bypass
```

### 4. Set Up Health Checks

```yaml
# Kubernetes readiness probe
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

# AWS Target Group health check
# Automatically configured by ingress-nginx
```

### 5. Use Network Policies

```yaml
# Restrict traffic to ingress only
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-ingress-only
spec:
  podSelector:
    matchLabels:
      app: hello
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
```

---

## Cost Optimization

### Cloudflare Costs
```
DNS: FREE
CDN (if proxy ON): FREE
DDoS protection: FREE
Argo (optional): $5/month + $0.10/GB
```

### AWS Costs
```
NLB: $0.0225/hour (~$16/month) + $0.006/GB processed
ALB: $0.0225/hour (~$16/month) + $0.008/LCU-hour
EKS cluster: $0.10/hour (~$73/month)
EC2 nodes: t3.medium x3 = ~$90/month

Total: ~$180-200/month (small cluster)
```

### Savings Tips
1. Use Cloudflare proxy (FREE) instead of CloudFront ($$$)
2. Use NLB instead of ALB if you don't need Layer 7 features
3. Enable cluster autoscaler to scale nodes down during low traffic
4. Use Spot instances for non-critical workloads

---

## Summary: Setup Checklist

- [ ] Create EKS cluster
- [ ] Install NGINX Ingress Controller (creates AWS Load Balancer)
- [ ] Get Load Balancer hostname/IP
- [ ] Add DNS records in Cloudflare (CNAME or A record)
- [ ] Deploy application & service to K8s
- [ ] Create Ingress resource pointing to service
- [ ] Configure SSL (Cloudflare proxy OR cert-manager)
- [ ] Test: `curl https://api.yourdomain.com`
- [ ] Configure real IP restoration (if proxy ON)
- [ ] Set up monitoring and alerts

**Your application is now live with Cloudflare DNS/CDN + AWS Kubernetes!** 🎉
