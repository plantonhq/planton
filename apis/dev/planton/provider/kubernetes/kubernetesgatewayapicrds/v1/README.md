# Kubernetes Gateway API CRDs

Install Kubernetes Gateway API Custom Resource Definitions (CRDs) on any Kubernetes cluster to enable next-generation ingress and service mesh traffic management.

## Overview

**KubernetesGatewayApiCrds** is a Planton component that installs the [Kubernetes Gateway API](https://gateway-api.sigs.k8s.io/) CRDs on any Kubernetes cluster. The Gateway API is the next evolution of Kubernetes ingress, providing a more expressive, role-oriented API for managing HTTP, TLS, and TCP routing.

Once installed, users can create Gateway, HTTPRoute, GRPCRoute, and other Gateway API resources that work with any conformant implementation (Istio, Envoy Gateway, NGINX Gateway Fabric, Traefik, etc.).

## Why We Created This

Managing Gateway API CRD installation across multiple clusters presents challenges:

- **Version consistency**: Different clusters may have different Gateway API versions
- **Channel selection**: Standard vs experimental CRDs need careful consideration
- **Upgrade complexity**: Upgrading CRDs across many clusters is error-prone
- **Prerequisite management**: Gateway API implementations expect specific CRD versions

Our KubernetesGatewayApiCrds module solves these problems by providing:

1. **Declarative CRD management**: Specify version and channel in a manifest
2. **Multi-cluster consistency**: Same manifest deploys identical CRDs everywhere
3. **Version pinning**: Control exactly which Gateway API version is installed
4. **Channel flexibility**: Easy switch between standard and experimental features
5. **Audit trail**: CRD installation is tracked like any other infrastructure

## Key Features

### 🚀 Simple Installation
- One manifest installs all Gateway API CRDs
- No kubectl commands or manual YAML application
- Works with any Kubernetes cluster (GKE, EKS, AKS, etc.)

### 📦 Version Control
- Pin to specific Gateway API versions (v1.2.1, v1.3.0, etc.)
- Easily upgrade by changing version in manifest
- Audit trail of version changes

### 🔧 Channel Selection
- **Standard channel**: Stable resources (Gateway, GatewayClass, HTTPRoute, ReferenceGrant)
- **Experimental channel**: All standard resources plus TCPRoute, UDPRoute, TLSRoute, GRPCRoute

### 🌍 Multi-Cloud Ready
- Works on any Kubernetes cluster
- Cloud-agnostic installation
- Consistent behavior across providers

## Quick Start

See [examples.md](examples.md) for complete YAML manifests.

### Example: Install Standard Gateway API CRDs

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api-crds
spec:
  target_cluster:
    cluster_kind: GcpGkeCluster
    cluster_name: prod-gke-cluster
  version: v1.2.1
  install_channel:
    channel: standard
```

This installs:
- `gatewayclasses.gateway.networking.k8s.io`
- `gateways.gateway.networking.k8s.io`
- `httproutes.gateway.networking.k8s.io`
- `referencegrants.gateway.networking.k8s.io`

## How It Works

1. **Deploy**: Apply a KubernetesGatewayApiCrds manifest
2. **Fetch**: The module downloads the specified Gateway API version
3. **Install**: CRDs are applied to the target cluster
4. **Ready**: Gateway API resources can now be created

## CRD Channels

### Standard Channel (Recommended)

Includes stable, production-ready resources:

| CRD | Description |
|-----|-------------|
| `GatewayClass` | Defines a class of Gateways with shared configuration |
| `Gateway` | Defines where and how traffic enters the cluster |
| `HTTPRoute` | Routes HTTP traffic to backend services |
| `ReferenceGrant` | Allows cross-namespace references |

### Experimental Channel

Includes all standard CRDs plus experimental resources:

| CRD | Description |
|-----|-------------|
| `TCPRoute` | Routes TCP traffic |
| `UDPRoute` | Routes UDP traffic |
| `TLSRoute` | Routes TLS traffic based on SNI |
| `GRPCRoute` | Native gRPC routing (without HTTP tunneling) |

**Note:** Experimental resources may change between versions.

## Configuration Reference

| Field | Description | Default |
|-------|-------------|---------|
| `target_cluster` | Target Kubernetes cluster | Required |
| `version` | Gateway API version to install | `v1.2.1` |
| `install_channel.channel` | CRD channel (standard/experimental) | `standard` |

## Prerequisites

- Kubernetes cluster (any provider)
- Cluster admin permissions (to install CRDs)
- A Gateway API implementation (Istio, Envoy Gateway, etc.) if you want to use the CRDs

## Use Cases

- **Pre-requisite installation**: Install CRDs before deploying Gateway API implementations
- **Multi-cluster standardization**: Ensure all clusters have the same Gateway API version
- **Version upgrades**: Controlled upgrade of Gateway API across environments
- **Development environments**: Quick setup of Gateway API for testing

## Next Steps

1. Review the [examples](examples.md) for your cluster type
2. Create a KubernetesGatewayApiCrds manifest
3. Deploy via `planton deploy`
4. Install a Gateway API implementation (Istio, Envoy Gateway, etc.)
5. Create Gateway and HTTPRoute resources

## Documentation

- **[Examples](examples.md)**: Complete YAML manifests for different scenarios
- **[Research Documentation](docs/README.md)**: Deep dive into Gateway API landscape
- **[Pulumi Module](iac/pulumi/README.md)**: Using the Pulumi implementation directly
- **[Terraform Module](iac/tf/README.md)**: Using the Terraform implementation directly

## External Resources

- [Gateway API Official Documentation](https://gateway-api.sigs.k8s.io/)
- [Gateway API GitHub Repository](https://github.com/kubernetes-sigs/gateway-api)
- [Gateway API Implementations](https://gateway-api.sigs.k8s.io/implementations/)

## Support

For issues, questions, or contributions, see the main [Planton repository](https://github.com/plantonhq/planton).
