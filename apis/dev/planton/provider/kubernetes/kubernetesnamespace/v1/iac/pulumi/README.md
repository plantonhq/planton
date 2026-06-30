# Kubernetes Namespace - Pulumi Module

## Overview

This Pulumi module provides automated deployment and management of Kubernetes namespaces with complete resource quotas, network policies, and service mesh integration. It implements the "Namespace-as-a-Service" pattern, transforming a basic namespace into a production-ready, multi-tenant environment.

## Key Features

- **Automated Namespace Provisioning**: Creates Kubernetes namespaces with all required configuration
- **Resource Quotas**: Implements CPU, memory, and object count quotas based on T-shirt sizes or custom values
- **LimitRanges**: Automatically injects default resource requests/limits into containers
- **Network Policies**: Enforces zero-trust networking with ingress/egress controls
- **Service Mesh Integration**: Automatic sidecar injection for Istio, Linkerd, or Consul
- **Pod Security Standards**: Enforces Kubernetes-native security policies (Privileged/Baseline/Restricted)
- **Cost Allocation**: Labels for tracking and governance

## Usage

### Basic Deployment

```bash
cd apis/dev/planton/provider/kubernetes/kubernetesnamespace/v1/iac/pulumi

# Preview changes
make preview manifest=../hack/manifest.yaml

# Deploy
make up manifest=../hack/manifest.yaml

# Destroy
make down manifest=../hack/manifest.yaml
```

### Using Planton CLI

```bash
# Validate the manifest
planton validate --manifest namespace.yaml

# Deploy with Pulumi
planton pulumi up --manifest namespace.yaml --stack myorg/myproject/dev

# Check outputs
planton pulumi stack output --manifest namespace.yaml --stack myorg/myproject/dev
```

## Module Architecture

### Resource Creation Flow

1. **Namespace**: Creates the base Kubernetes namespace resource with labels and annotations
2. **ResourceQuota**: Applies CPU, memory, and object count limits
3. **LimitRange**: Configures default container resource requests/limits
4. **NetworkPolicies**: Creates ingress and egress isolation policies
5. **Outputs**: Exports observable identifiers and configuration status

### Component Structure

```
module/
├── main.go              # Entry point and orchestration
├── locals.go            # Configuration and derived values
├── namespace.go         # Namespace creation
├── resource_quota.go    # ResourceQuota implementation
├── limit_range.go       # LimitRange implementation
├── network_policies.go  # NetworkPolicy creation
└── outputs.go           # Stack outputs
```

## Configuration Patterns

### T-Shirt Sizing

The module provides preset resource profiles:

- **SMALL**: 2-4 CPU, 4-8Gi memory, 20 pods
- **MEDIUM**: 4-8 CPU, 8-16Gi memory, 50 pods
- **LARGE**: 8-16 CPU, 16-32Gi memory, 100 pods
- **XLARGE**: 16-32 CPU, 32-64Gi memory, 200 pods

### Custom Quotas

For precise control, specify exact values for CPU, memory, and object counts.

### Network Isolation

- **Ingress Isolation**: Default-deny ingress with explicit allow lists
- **Egress Restriction**: Block external access except DNS, Kubernetes API, and specified CIDRs/domains

### Service Mesh

Automatic configuration for:
- **Istio**: Supports revision tags for canary upgrades
- **Linkerd**: Lightweight mesh injection
- **Consul**: HashiCorp service mesh integration

## Outputs

The module exports the following outputs:

| Output | Description |
|--------|-------------|
| `namespace` | The created namespace name |
| `namespace_id` | Namespace identifier |
| `resource_quotas_applied` | Whether quotas were configured |
| `limit_ranges_applied` | Whether default limits were set |
| `network_policies_applied` | Whether network policies were created |
| `service_mesh_enabled` | Service mesh injection status |
| `service_mesh_type` | The configured mesh type |
| `pod_security_standard` | Enforced security level |
| `labels_json` | Applied labels (JSON) |
| `annotations_json` | Applied annotations (JSON) |

## Prerequisites

- Kubernetes cluster (any distribution: GKE, EKS, AKS, self-hosted)
- Kubernetes credentials (kubeconfig)
- For network policies: CNI plugin that supports NetworkPolicy (Calico, Cilium, etc.)
- For service mesh: Pre-installed mesh control plane (Istio/Linkerd/Consul)

## Best Practices

1. **Start with Presets**: Use SMALL/MEDIUM/LARGE profiles initially
2. **Enable Network Isolation**: Always enable for production namespaces
3. **Use Meaningful Labels**: Add team, environment, cost-center labels
4. **Service Mesh Revision Tags**: Use revision tags instead of hardcoded versions
5. **Test Security Standards**: Start with BASELINE, move to RESTRICTED after validation

## Troubleshooting

### Namespace Stuck in Terminating

```bash
# Check for finalizers
kubectl get namespace <name> -o yaml | grep finalizers

# Remove finalizers if needed (use with caution)
kubectl patch namespace <name> -p '{"metadata":{"finalizers":[]}}' --type=merge
```

### Pods Can't Be Scheduled (Quota Exceeded)

```bash
# Check quota usage
kubectl describe resourcequota -n <namespace>

# Update manifest with higher quota and reapply
```

### Network Policies Blocking Traffic

```bash
# Check policies
kubectl get networkpolicy -n <namespace>
kubectl describe networkpolicy -n <namespace>

# Add allowed namespaces/CIDRs to manifest
```

## Examples

See [../../examples.md](../../examples.md) for complete examples including:
- Basic namespace creation
- Production namespace with network isolation
- Service mesh integration
- Custom resource quotas
- Ephemeral PR environments

## References

- [Kubernetes Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
- [Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Multi-Tenancy](https://kubernetes.io/docs/concepts/security/multi-tenancy/)


