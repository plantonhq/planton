# Kubernetes Gateway API CRDs: Research Documentation

## Introduction

The Kubernetes Gateway API represents the next generation of ingress and service mesh traffic routing APIs for Kubernetes. Unlike its predecessor, the Ingress API, Gateway API was designed from the ground up with role-oriented design, expressiveness, and extensibility as core principles.

This research document explores the Gateway API landscape, explains why CRD management matters, and justifies the design decisions in Planton's KubernetesGatewayApiCrds component.

## The Evolution of Kubernetes Traffic Management

### The Ingress Era (2015-2020)

Kubernetes Ingress was introduced in 2015 as a simple way to route HTTP traffic to services. While widely adopted, Ingress had significant limitations:

- **Single resource type**: One Ingress resource handled everything
- **Annotation explosion**: Advanced features required vendor-specific annotations
- **No role separation**: The same person defining routes also configured infrastructure
- **Limited protocol support**: HTTP/HTTPS only, no TCP/UDP
- **Provider lock-in**: Annotations differed between NGINX, Traefik, HAProxy

```yaml
# Ingress example with vendor-specific annotations
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    # ^ Provider-specific, non-portable
```

### Gateway API Genesis (2019-Present)

The Gateway API project started in 2019 as a Kubernetes SIG-Network initiative to address Ingress limitations. Key milestones:

| Year | Milestone |
|------|-----------|
| 2019 | Project inception as "Service APIs" |
| 2020 | Renamed to Gateway API |
| 2021 | Alpha release (v0.3.0) |
| 2022 | Beta release (v0.5.0) |
| 2023 | GA release (v1.0.0) |
| 2024 | v1.2.0 with enhanced features |

## Gateway API Architecture

### Role-Oriented Design

Gateway API separates concerns across three personas:

```
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure Provider                   │
│  Creates GatewayClass (defines capabilities, defaults)       │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Cluster Operator                          │
│  Creates Gateway (where traffic enters, TLS, listeners)      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Application Developer                     │
│  Creates HTTPRoute/GRPCRoute (how traffic routes)           │
└─────────────────────────────────────────────────────────────┘
```

### Core Resources

#### GatewayClass

Defines a class of Gateways with shared behavior. Typically created by infrastructure providers:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: istio
spec:
  controllerName: istio.io/gateway-controller
  description: "Istio Gateway Controller"
```

#### Gateway

Defines where and how traffic enters the cluster:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: prod-gateway
  namespace: gateway-system
spec:
  gatewayClassName: istio
  listeners:
  - name: https
    port: 443
    protocol: HTTPS
    hostname: "*.example.com"
    tls:
      mode: Terminate
      certificateRefs:
      - name: wildcard-cert
```

#### HTTPRoute

Routes HTTP traffic to backend services:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: app-route
  namespace: my-app
spec:
  parentRefs:
  - name: prod-gateway
    namespace: gateway-system
  hostnames:
  - "app.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /api
    backendRefs:
    - name: api-service
      port: 8080
```

### Standard vs Experimental Channels

Gateway API maintains two release channels:

#### Standard Channel

Contains stable, GA resources with strong backward compatibility guarantees:

| Resource | Status | Use Case |
|----------|--------|----------|
| GatewayClass | GA | Define gateway types |
| Gateway | GA | Define entry points |
| HTTPRoute | GA | HTTP routing |
| ReferenceGrant | GA | Cross-namespace refs |

#### Experimental Channel

Contains resources still maturing:

| Resource | Status | Use Case |
|----------|--------|----------|
| TCPRoute | Beta | Raw TCP routing |
| UDPRoute | Alpha | UDP routing |
| TLSRoute | Alpha | TLS passthrough |
| GRPCRoute | Beta | Native gRPC |

## Installation Landscape

### Manual Installation

The simplest approach—apply manifests directly:

```bash
# Standard channel
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/standard-install.yaml

# Experimental channel
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.1/experimental-install.yaml
```

**Pros:**
- Simple, no tools required
- Official manifests from upstream

**Cons:**
- Manual process per cluster
- No version tracking
- No rollback mechanism
- Inconsistent across clusters

### Helm-Based Installation

Some implementations bundle CRDs with their Helm charts:

```bash
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm install istio-base istio/base -n istio-system --create-namespace
# CRDs installed as part of istio-base
```

**Pros:**
- Bundled with implementation
- Helm release tracking

**Cons:**
- CRD version tied to implementation version
- Potential conflicts if multiple implementations
- CRD ownership issues during uninstall

### GitOps Installation

Apply CRD manifests via ArgoCD/Flux:

```yaml
# ArgoCD Application
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: gateway-api-crds
spec:
  source:
    repoURL: https://github.com/kubernetes-sigs/gateway-api
    targetRevision: v1.2.1
    path: config/crd/standard
  destination:
    server: https://kubernetes.default.svc
```

**Pros:**
- GitOps workflow
- Version controlled
- Audit trail

**Cons:**
- Requires GitOps tooling
- Complex for multi-cluster

### IaC-Based Installation

Use Terraform or Pulumi to manage CRD lifecycle:

```hcl
# Terraform
resource "kubernetes_manifest" "gateway_api_crds" {
  for_each = fileset("${path.module}/crds", "*.yaml")
  manifest = yamldecode(file("${path.module}/crds/${each.value}"))
}
```

**Pros:**
- Infrastructure-as-code approach
- State tracking
- Multi-cluster deployment

**Cons:**
- Additional tooling
- State management overhead

## Why CRD Management Matters

### Version Consistency

Gateway API implementations have specific CRD version requirements:

| Implementation | Minimum CRD Version | Recommended |
|----------------|---------------------|-------------|
| Istio 1.21+ | v1.0.0 | v1.1.0+ |
| Envoy Gateway 1.0 | v1.0.0 | v1.1.0+ |
| NGINX Gateway Fabric | v1.0.0 | v1.0.0+ |
| GKE Gateway Controller | v1.0.0 | v1.0.0+ |

Running mismatched versions causes:
- Validation failures
- Missing features
- Unexpected behavior
- Controller crashes

### Upgrade Considerations

CRD upgrades require care:

1. **Backward compatibility**: New CRD versions must support existing resources
2. **Conversion webhooks**: Some versions require webhooks for migration
3. **Implementation compatibility**: Controllers must support new CRD version
4. **Rollback complexity**: Downgrading CRDs can break existing resources

### Multi-Cluster Challenges

Organizations running multiple clusters face:

- **Version drift**: Different clusters have different CRD versions
- **Upgrade coordination**: Upgrading CRDs across clusters is error-prone
- **Testing complexity**: Hard to test upgrades before production
- **Audit requirements**: No record of when/what was installed

## Planton's Approach

### Design Principles

1. **Declarative Management**: CRD installation is a manifest, not a command
2. **Version Pinning**: Explicit version selection, no "latest" surprises
3. **Channel Selection**: Easy switch between standard and experimental
4. **Multi-Cluster Ready**: Same manifest works across all clusters
5. **Audit Trail**: Installation tracked like any infrastructure change

### Implementation Strategy

Our KubernetesGatewayApiCrds component:

1. **Fetches official manifests** from kubernetes-sigs/gateway-api releases
2. **Applies CRDs** using kubernetes provider
3. **Tracks installed version** in stack outputs
4. **Supports both channels** via simple configuration

### 80/20 Scoping

We intentionally keep this component focused:

**In Scope:**
- CRD installation from official releases
- Version selection
- Channel selection
- Multi-cluster deployment

**Out of Scope:**
- Gateway implementation deployment (handled by separate components)
- Custom CRD modifications
- CRD conversion webhook management
- Implementation-specific CRD extensions

This separation ensures:
- Clean dependency management
- Faster updates
- Reduced complexity
- Clear ownership

## Implementation Comparison

### Implementations Overview

| Implementation | Maintainer | Best For |
|----------------|------------|----------|
| Istio | Google/Community | Service mesh users |
| Envoy Gateway | Envoy Proxy | Standalone Envoy |
| NGINX Gateway Fabric | F5/NGINX | NGINX users |
| Traefik | Traefik Labs | Simple setups |
| Contour | VMware | Enterprise deployments |
| Kong Gateway | Kong | API management |
| GKE Gateway Controller | Google | GKE native |
| AWS Gateway API Controller | AWS | EKS/ALB native |

### Feature Matrix

| Feature | Istio | Envoy GW | NGINX GW | Traefik |
|---------|-------|----------|----------|---------|
| HTTPRoute | ✅ | ✅ | ✅ | ✅ |
| GRPCRoute | ✅ | ✅ | ❌ | ✅ |
| TCPRoute | ✅ | ✅ | ✅ | ✅ |
| TLSRoute | ✅ | ✅ | ❌ | ✅ |
| Traffic Splitting | ✅ | ✅ | ✅ | ✅ |
| Header Modification | ✅ | ✅ | ✅ | ✅ |
| Rate Limiting | ✅ | ✅ | ✅ | ✅ |
| mTLS | ✅ | ✅ | ❌ | ✅ |

## Best Practices

### Version Management

1. **Pin to specific versions**: Never use "latest" in production
2. **Test upgrades in staging**: Verify compatibility before production
3. **Document version requirements**: Track which implementations need which CRD versions
4. **Coordinate upgrades**: Upgrade CRDs before implementations that need them

### Channel Selection

1. **Start with standard**: Use standard channel unless you need experimental features
2. **Test experimental thoroughly**: Experimental resources may change
3. **Monitor deprecations**: Stay informed about graduation timelines
4. **Plan migrations**: Have a plan for when experimental becomes standard

### Multi-Cluster Strategy

1. **Centralized version management**: Use Planton to ensure consistency
2. **Environment progression**: dev → staging → production
3. **Canary upgrades**: Upgrade one cluster first, monitor, then roll out
4. **Rollback plan**: Know how to revert if issues arise

## Common Pitfalls

### CRD Ownership Conflicts

**Problem:** Multiple Helm releases try to manage the same CRDs.

**Solution:** Install CRDs separately from implementations using KubernetesGatewayApiCrds.

### Version Mismatches

**Problem:** Implementation requires newer CRD version than installed.

**Solution:** Always check implementation requirements, upgrade CRDs first.

### Experimental Feature Breakage

**Problem:** Experimental resource API changed between versions.

**Solution:** Monitor changelog, test in non-production first, have migration plan.

### Cross-Namespace Reference Issues

**Problem:** Routes in namespace A can't reference backends in namespace B.

**Solution:** Create appropriate ReferenceGrant resources.

## Conclusion

The Kubernetes Gateway API represents a significant improvement over Ingress, providing a more expressive, role-oriented, and portable API for traffic management. Proper CRD management is essential for production deployments.

Planton's KubernetesGatewayApiCrds component provides:

- **Declarative CRD management** instead of manual kubectl commands
- **Version control** for consistent deployments
- **Multi-cluster support** from a single manifest
- **Clear separation** between CRD installation and implementation deployment

This approach aligns with infrastructure-as-code best practices and enables organizations to confidently adopt Gateway API across their Kubernetes fleet.

## References

- [Gateway API Official Documentation](https://gateway-api.sigs.k8s.io/)
- [Gateway API GitHub Repository](https://github.com/kubernetes-sigs/gateway-api)
- [Gateway API Releases](https://github.com/kubernetes-sigs/gateway-api/releases)
- [Gateway API Implementations](https://gateway-api.sigs.k8s.io/implementations/)
- [Gateway API GEPs (Enhancement Proposals)](https://gateway-api.sigs.k8s.io/geps/overview/)
- [SIG-Network Gateway API Meetings](https://github.com/kubernetes/community/tree/master/sig-network)
