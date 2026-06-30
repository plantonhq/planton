# Civo Kubernetes Cluster

Deploy production-grade Kubernetes clusters on Civo Cloud using K3s - the lightweight, certified Kubernetes distribution.

## Overview

The `CivoKubernetesCluster` resource provisions managed Kubernetes clusters on Civo Cloud. Built on K3s (a CNCF-certified lightweight Kubernetes), Civo clusters provide full Kubernetes compatibility with faster startup times and lower resource overhead than standard distributions.

**Key features:**

- **Managed control plane** - Civo handles master node management, upgrades, and HA
- **K3s-based** - Lightweight, fast, fully Kubernetes-compliant
- **Built-in marketplace** - 80+ pre-packaged applications (Traefik, cert-manager, Prometheus)
- **Flexible CNI** - Flannel (default) or Cilium for NetworkPolicy support
- **Auto-scaling** - Optional cluster autoscaler integration
- **Fast provisioning** - Clusters ready in 90-120 seconds
- **Cost-effective** - Start at ~$11/month for development clusters

## Prerequisites

- Civo account with API access
- Civo API token ([get one here](https://dashboard.civo.com/security))
- Existing Civo network (VPC) - use `CivoVpc` resource
- Planton CLI installed
- `kubectl` installed locally

## Quick Start

### 1. Development Cluster (Single Node)

Minimal cluster for local development and testing:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoKubernetesCluster
metadata:
  name: dev-cluster
spec:
  clusterName: dev-k8s
  region: lon1
  kubernetesVersion: "1.29.0+k3s1"
  network:
    value: "your-network-id"
  autoUpgrade: true
  tags:
    - environment:dev
  defaultNodePool:
    size: g4s.kube.small
    nodeCount: 1
```

### 2. Staging Cluster (Multi-Node)

Production-like setup for testing:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoKubernetesCluster
metadata:
  name: staging-cluster
spec:
  clusterName: staging-k8s
  region: lon1
  kubernetesVersion: "1.29.0+k3s1"
  network:
    value: "your-network-id"
  autoUpgrade: false  # Manual upgrades for staging
  tags:
    - environment:staging
    - team:platform
  defaultNodePool:
    size: g4s.kube.medium
    nodeCount: 3
```

### 3. Production Cluster (HA)

High-availability cluster for production workloads:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoKubernetesCluster
metadata:
  name: production-cluster
spec:
  clusterName: prod-k8s
  region: fra1
  kubernetesVersion: "1.29.0+k3s1"
  network:
    value: "your-network-id"
  highlyAvailable: true  # HA control plane
  autoUpgrade: false  # Manual upgrades only
  disableSurgeUpgrade: false
  tags:
    - environment:production
    - critical:true
  defaultNodePool:
    size: g4s.kube.large
    nodeCount: 5
```

## Deploy with Planton CLI

```bash
# Create the cluster
planton apply -f cluster.yaml

# Wait for cluster to be ready (90-120 seconds)
planton get civokubernetesclusters

# Get kubeconfig
planton kubeconfig civokubernetesclusters/dev-cluster > ~/.kube/config

# Test access
kubectl get nodes
kubectl get pods --all-namespaces
```

## Configuration Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `clusterName` | string | Yes | Unique cluster name (alphanumeric, hyphens) |
| `region` | enum | Yes | Civo region: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1` |
| `kubernetesVersion` | string | Yes | K8s version (e.g., `1.29.0+k3s1`) |
| `network` | StringValueOrRef | Yes | Network (VPC) for cluster |
| `highlyAvailable` | bool | No | Enable HA control plane (default: `false`) |
| `autoUpgrade` | bool | No | Enable automatic patch upgrades (default: `false`) |
| `disableSurgeUpgrade` | bool | No | Disable surge upgrades (default: `false`) |
| `tags` | array[string] | No | Resource tags for organization |
| `defaultNodePool` | object | Yes | Default node pool configuration |

### Default Node Pool Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `size` | string | Yes | Node instance size (e.g., `g4s.kube.medium`) |
| `nodeCount` | uint32 | Yes | Number of nodes (must be > 0) |

### Available Node Sizes

| Size | CPU | RAM | Storage | Monthly Cost* |
|------|-----|-----|---------|---------------|
| `g4s.kube.small` | 1 | 2 GB | 30 GB | ~$11 |
| `g4s.kube.medium` | 2 | 4 GB | 50 GB | ~$22 |
| `g4s.kube.large` | 4 | 8 GB | 100 GB | ~$43 |
| `g4s.kube.xlarge` | 6 | 16 GB | 150 GB | ~$87 |

*Approximate pricing as of 2025. Check [Civo pricing](https://www.civo.com/pricing) for current rates.

### Available Regions

- `lon1` - London (UK)
- `lon2` - London 2 (UK)
- `fra1` - Frankfurt (Germany)
- `nyc1` - New York (US)
- `phx1` - Phoenix (US)
- `mum1` - Mumbai (India)

## Stack Outputs

After provisioning, the following outputs are available:

- `cluster_id` - Civo's unique cluster identifier
- `cluster_name` - Cluster name
- `kubeconfig` - Base64-encoded kubeconfig for cluster access
- `api_endpoint` - Kubernetes API server URL
- `master_ip` - Control plane public IP

Access outputs via:

```bash
planton outputs civokubernetesclusters/dev-cluster
```

## Best Practices

### 1. Network Isolation

Always create clusters in dedicated networks:

```yaml
# First create network
apiVersion: civo.planton.dev/v1
kind: CivoVpc
metadata:
  name: k8s-network
spec:
  label: Kubernetes cluster network
  region: lon1

# Then reference in cluster
spec:
  network:
    valueFrom:
      kind: CivoVpc
      name: k8s-network
      fieldPath: "status.outputs.network_id"
```

### 2. Version Pinning

Pin Kubernetes versions for stability:

```yaml
kubernetesVersion: "1.29.0+k3s1"  # Explicit version
autoUpgrade: false  # Manual upgrades only
```

**Upgrade strategy:**
1. Test new version in dev cluster
2. Upgrade staging
3. Wait 1-2 weeks
4. Upgrade production during maintenance window

### 3. Node Count Planning

| Environment | Recommended Nodes | Rationale |
|-------------|-------------------|-----------|
| Dev | 1 | Cost optimization |
| Staging | 3 | Test HA patterns |
| Production | 5+ | Allow graceful node rotation |

### 4. High Availability

Enable HA for production:

```yaml
highlyAvailable: true  # Multiple control plane nodes
```

**Note**: HA control plane has additional cost. Use for production only.

### 5. Resource Tags

Use consistent tagging:

```yaml
tags:
  - "environment:production"
  - "team:platform"
  - "cost-center:engineering"
  - "managed-by:planton"
```

Tags enable:
- Cost allocation and reporting
- Resource filtering and discovery
- Compliance and governance

### 6. Cluster Sizing

Right-size nodes for workload:

- **Dev/Test**: `g4s.kube.small` (1 core, 2 GB)
- **Staging**: `g4s.kube.medium` (2 cores, 4 GB)
- **Production**: `g4s.kube.large` (4 cores, 8 GB) or larger

**Tip**: Start smaller and scale up if needed. Kubernetes handles multi-node well.

## Post-Deployment Steps

### 1. Configure kubectl

```bash
# Get kubeconfig
planton kubeconfig civokubernetesclusters/dev-cluster > ~/.kube/dev-config

# Set context
export KUBECONFIG=~/.kube/dev-config

# Verify access
kubectl get nodes
kubectl cluster-info
```

### 2. Install Essential Apps

```bash
# cert-manager for TLS automation
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Verify installation
kubectl get pods -n cert-manager
```

### 3. Set Up Ingress

If using default Traefik:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
spec:
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: my-app
                port:
                  number: 80
```

## Common Use Cases

### NetworkPolicy Support

For production clusters requiring network segmentation, use Cilium CNI:

**Note**: CNI must be selected during cluster creation (cannot be changed later). Currently requires direct IaC module access.

### Reserved IP for LoadBalancer

Use with `CivoIpAddress` for stable LoadBalancer IPs:

```yaml
# First reserve IP
apiVersion: civo.planton.dev/v1
kind: CivoIpAddress
metadata:
  name: ingress-ip
spec:
  region: lon1
  description: Ingress LoadBalancer IP

# Then in Service
apiVersion: v1
kind: Service
metadata:
  name: nginx-ingress
spec:
  type: LoadBalancer
  loadBalancerIP: "74.220.24.88"  # From CivoIpAddress outputs
  ports:
    - port: 80
```

### Custom Firewall

Restrict cluster access with `CivoFirewall`:

```yaml
apiVersion: civo.planton.dev/v1
kind: CivoFirewall
metadata:
  name: k8s-firewall
spec:
  name: k8s-cluster-fw
  networkId:
    value: "same-network-as-cluster"
  inboundRules:
    - protocol: tcp
      portRange: "6443"
      cidrs: ["203.0.113.0/24"]  # Office IP range
      label: API server from office
    - protocol: tcp
      portRange: "80"
      cidrs: ["0.0.0.0/0"]
      label: HTTP ingress
    - protocol: tcp
      portRange: "443"
      cidrs: ["0.0.0.0/0"]
      label: HTTPS ingress
```

## Limitations

### K3s vs Standard Kubernetes

Civo uses K3s, which has some differences:

- **containerd only** - No Docker runtime option
- **SQLite embedded** - Default etcd replacement (HA uses etcd)
- **Smaller footprint** - ~50% less memory than standard K8s
- **Some features excluded** - Legacy in-tree cloud providers removed

For most workloads, these differences are invisible. See [K3s documentation](https://docs.k3s.io/) for details.

### CNI Selection

- **Cannot change CNI** after cluster creation
- NetworkPolicy requires Cilium (Flannel doesn't support it)
- Plan CNI choice carefully before deployment

### Node Pool Limitations

Currently only default node pool is configurable via this API. For:
- Multiple node pools with different sizes
- GPU nodes
- Spot instances (when available)

Use direct Civo API/CLI or IaC modules.

## Troubleshooting

### Cluster Creation Stuck

If cluster shows "Building" for >5 minutes:

```bash
# Check status
planton get civokubernetesclusters/dev-cluster

# View Civo console
# Visit https://dashboard.civo.com/kubernetes
```

Common causes:
- Network quota exceeded
- Insufficient IP addresses in network
- Region capacity issues

### Cannot Connect with kubectl

```bash
# Verify kubeconfig
kubectl config view

# Test API endpoint
curl -k https://api-endpoint:6443

# Check firewall rules (API port 6443 must be open)
planton get civofirewalls
```

### Nodes Not Ready

```bash
# Check node status
kubectl get nodes
kubectl describe node <node-name>

# Check kubelet logs (requires SSH to node)
civo kubernetes show dev-k8s
# Get node IP from output
ssh civo@<node-ip>
sudo journalctl -u k3s-agent -f
```

### Marketplace App Installation Failed

View installation logs:

```bash
# Check deployed apps
kubectl get pods --all-namespaces

# Check specific app
kubectl logs -n <namespace> <pod-name>
```

## Security Considerations

### API Server Access

By default, API server (port 6443) is publicly accessible. For production:

1. Create restrictive firewall
2. Limit API access to office/VPN IPs
3. Use bastion host for cluster administration

### Network Policies

Requires Cilium CNI:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all-ingress
spec:
  podSelector: {}
  policyTypes:
    - Ingress
```

### RBAC

Configure role-based access control:

```bash
# Create service account for application
kubectl create serviceaccount app-sa

# Bind minimal permissions
kubectl create rolebinding app-binding \
  --clusterrole=view \
  --serviceaccount=default:app-sa
```

## More Information

- **Deep Dive** - See [docs/README.md](docs/README.md) for comprehensive research on K3s, deployment methods, CNI decisions, and production patterns
- **Examples** - Check [examples.md](examples.md) for more cluster configurations
- **Pulumi Module** - See [iac/pulumi/README.md](iac/pulumi/README.md) for direct Pulumi usage
- **Civo Kubernetes API** - [Official API documentation](https://www.civo.com/api/kubernetes)
- **K3s Documentation** - [k3s.io](https://docs.k3s.io/)

## Support

- Issues & Feature Requests: [Planton GitHub](https://github.com/plantonhq/planton/issues)
- Civo Support: [support@civo.com](mailto:support@civo.com)
- Community: [Planton Discord](#)

## Related Resources

- `CivoVpc` - Create networks for clusters
- `CivoFirewall` - Restrict cluster access
- `CivoIpAddress` - Reserve static IPs for LoadBalancers
- `CivoDnsZone` - Configure DNS for cluster services

