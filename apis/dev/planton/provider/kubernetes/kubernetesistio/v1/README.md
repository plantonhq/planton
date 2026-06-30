# Kubernetes Istio

## Overview

The **KubernetesIstio** API resource provides a streamlined interface for deploying the Istio service mesh on Kubernetes clusters. This resource simplifies the installation of Istio's core components (base, istiod control plane, and ingress gateway) using the official Helm charts, with sensible defaults and configurable resource allocations for the control plane.

## Why We Created This API Resource

Deploying a production-ready service mesh presents several challenges:

1. **Complex Installation**: Istio requires multiple components (base CRDs, control plane, and gateways) to be installed in the correct order
2. **Resource Management**: Tuning control plane resources (CPU, memory) for different cluster sizes requires deep Istio knowledge
3. **Version Management**: Keeping Istio components in sync and managing upgrades is error-prone
4. **Configuration Overhead**: Managing Helm values across environments leads to configuration drift
5. **Deployment Methods**: Choosing between `istioctl`, Helm charts, and operators adds decision complexity

The KubernetesIstio resource solves these problems by providing a single, consistent API that:

- **Simplifies Installation**: Handles the complete installation sequence automatically (base → istiod → gateway)
- **Manages Resources**: Allows easy configuration of control plane resources through a simple container specification
- **Ensures Consistency**: Uses pinned, tested Helm chart versions for reproducible deployments
- **Reduces Complexity**: Abstracts away Helm value file management
- **Production Ready**: Built-in best practices for service mesh deployment

## Key Features

### Automated Component Installation

Deploys all essential Istio components in the correct order:

- **Istio Base**: Custom Resource Definitions (CRDs) and foundational resources
- **Istiod Control Plane**: Unified control plane combining Pilot, Citadel, and Galley
- **Ingress Gateway**: Gateway for handling external traffic entering the mesh

### Resource Configuration

Easy control over control plane resources:

- **CPU and Memory Limits**: Configure resource limits for the istiod deployment
- **Resource Requests**: Set resource requests to ensure QoS
- **Default Values**: Sensible defaults (1000m CPU, 1Gi memory limits) that work for most clusters
- **Scalability**: Resource configuration supports both small dev clusters and large production environments

### Version Management

- **Pinned Chart Versions**: Uses tested, stable Helm chart versions (currently 1.22.3)
- **Upgrade Path**: Clear path to upgrade Istio versions through chart version updates
- **Consistency**: All three components (base, istiod, gateway) use the same chart version

### Namespace Management

The KubernetesIstio component manages two namespaces:

1. **istio-system**: Main namespace for control plane components (istiod, base CRDs)
2. **istio-ingress**: Dedicated namespace for ingress gateway

**Namespace Creation Options**:

- **Automatic Creation** (`create_namespace: true`): The component creates both namespaces automatically
- **Use Existing** (`create_namespace: false`): Both namespaces must already exist in the cluster

**Important**: When using existing namespaces (`create_namespace: false`), ensure both `istio-system` and `istio-ingress` are created before deploying this component.

**Security Best Practice**: Separating control plane and ingress gateway into distinct namespaces provides better isolation and follows Istio's recommended architecture.

## How It Works

When you create a KubernetesIstio resource, the following happens:

1. **Namespace Creation**: Creates `istio-system` and `istio-ingress` namespaces with proper labels
2. **Base Installation**: Deploys Istio base Helm chart (CRDs and core resources)
3. **Control Plane Deployment**: Installs istiod with configured resource allocations
4. **Gateway Deployment**: Deploys ingress gateway configured as ClusterIP service
5. **Output Generation**: Exports connection endpoints and utility commands

All deployments use atomic Helm releases with automatic rollback on failure, ensuring safe installations.

## Benefits

### For Platform Engineers

- **Simplified Operations**: One resource type manages the entire service mesh installation
- **Reproducible Deployments**: Declarative configuration ensures consistent mesh deployments
- **Resource Control**: Easy tuning of control plane resources without complex Helm values
- **Clean Architecture**: Proper namespace separation and component organization

### For Development Teams

- **Transparent Service Mesh**: Istio runs seamlessly without requiring application changes
- **Traffic Management**: Advanced routing, retries, and circuit breaking without code changes
- **Observability**: Automatic distributed tracing and metrics for all services
- **Security**: mTLS encryption between services with automatic certificate management

### For Organizations

- **Security by Default**: Service-to-service encryption and strong identity for workloads
- **Compliance Ready**: Standardized service mesh deployment across all clusters
- **Production Proven**: Uses official Istio Helm charts following best practices
- **Future Ready**: Foundation for advanced traffic management and security policies

## Quick Start

### Basic Istio Installation

Deploy Istio with default resource allocations:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIstio
metadata:
  name: main-istio
spec:
  target_cluster:
    cluster_name: my-gke-cluster
  namespace:
    value: istio-system
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 50m
        memory: 100Mi
      limits:
        cpu: 1000m
        memory: 1Gi
```

### Production Istio with Higher Resources

For production clusters with high traffic:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIstio
metadata:
  name: prod-istio
spec:
  target_cluster:
    cluster_name: prod-gke-cluster
  namespace:
    value: istio-system
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 4000m
        memory: 8Gi
```

### Development Environment

Minimal resources for development clusters:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesIstio
metadata:
  name: dev-istio
spec:
  target_cluster:
    cluster_name: dev-gke-cluster
  namespace:
    value: istio-system
  create_namespace: true
  container:
    resources:
      requests:
        cpu: 25m
        memory: 64Mi
      limits:
        cpu: 500m
        memory: 256Mi
```

## Use Cases

1. **Microservices Communication**: Enable secure, reliable service-to-service communication
2. **Traffic Management**: Implement advanced routing (canary deployments, A/B testing)
3. **Security**: Enforce mTLS encryption between all services
4. **Observability**: Automatic distributed tracing and metrics collection
5. **API Gateway**: Use Istio ingress gateway as a modern API gateway
6. **Multi-Cluster**: Foundation for multi-cluster service mesh deployments

## Component Architecture

```
┌─────────────────────────────────────────────┐
│         KubernetesIstio Resource            │
│  (Declarative Service Mesh Configuration)   │
└─────────────────┬───────────────────────────┘
                  │
                  ├─────► istio-system namespace
                  │         ├─ Istio Base (CRDs)
                  │         └─ Istiod Control Plane
                  │            ├─ Pilot (traffic mgmt)
                  │            ├─ Citadel (security)
                  │            └─ Galley (config)
                  │
                  └─────► istio-ingress namespace
                            └─ Istio Gateway (ingress)
```

## Stack Outputs

After deployment, the following outputs are available:

- **namespace**: The namespace where Istio control plane is deployed (`istio-system`)
- **service**: The name of the istiod service (`istiod`)
- **port_forward_command**: Command to port-forward to istiod for debugging
- **kube_endpoint**: Kubernetes service endpoint for istiod
- **ingress_endpoint**: Kubernetes service endpoint for the ingress gateway

## Additional Resources

- [Official Istio Documentation](https://istio.io/latest/docs/)
- [Istio Helm Charts](https://github.com/istio/istio/tree/master/manifests/charts)
- [Istio Best Practices](https://istio.io/latest/docs/ops/best-practices/)
- [Detailed Research Documentation](docs/README.md) - Deep dive into deployment methods and architecture

## Next Steps

After deploying Istio:

1. **Enable Sidecar Injection**: Label namespaces with `istio-injection=enabled`
2. **Deploy Workloads**: Deploy your applications to namespaces with sidecar injection
3. **Configure Traffic Management**: Create VirtualServices and DestinationRules
4. **Set Up Observability**: Install Prometheus, Grafana, Jaeger, or Kiali for visibility
5. **Implement Security Policies**: Configure authorization policies and peer authentication

For comprehensive examples, see [examples.md](examples.md).

