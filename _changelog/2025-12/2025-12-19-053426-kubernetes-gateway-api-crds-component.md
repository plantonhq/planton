# New Component: KubernetesGatewayApiCrds

**Date**: December 19, 2025  
**Type**: Feature  
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Provider Framework

## Summary

Forged a complete new deployment component `KubernetesGatewayApiCrds` that installs Kubernetes Gateway API Custom Resource Definitions (CRDs) on any Kubernetes cluster. This component enables declarative management of Gateway API CRD installation across multiple clusters with version pinning and channel selection (standard vs experimental).

## Problem Statement / Motivation

The Kubernetes Gateway API is the next-generation API for managing ingress and service mesh traffic, replacing the legacy Ingress API. However, installing Gateway API CRDs across multiple clusters presents challenges:

### Pain Points

- **Version consistency**: Different clusters may have different Gateway API versions, leading to compatibility issues with Gateway implementations
- **Manual installation**: Teams typically run `kubectl apply -f` manually, which is error-prone and not auditable
- **Channel confusion**: Choosing between standard (stable) and experimental CRDs requires understanding the differences
- **Upgrade complexity**: Coordinating CRD upgrades across clusters is difficult without proper tooling
- **No IaC support**: Gateway API CRD installation wasn't available as a declarative, infrastructure-as-code resource

## Solution / What's New

Created `KubernetesGatewayApiCrds` as a Planton deployment component that provides declarative CRD installation.

### Component Capabilities

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesGatewayApiCrds
metadata:
  name: gateway-api-crds
spec:
  target_cluster:
    cluster_kind: GcpGkeCluster
    cluster_name: prod-cluster
  version: v1.2.1           # Pin to specific version
  install_channel:
    channel: standard        # or experimental
```

### Channels Supported

| Channel | CRDs Installed |
|---------|----------------|
| **Standard** | GatewayClass, Gateway, HTTPRoute, ReferenceGrant |
| **Experimental** | Standard + TCPRoute, UDPRoute, TLSRoute, GRPCRoute |

## Implementation Details

### Proto API (4 files)

**spec.proto** - Configuration schema with validations:
```protobuf
message KubernetesGatewayApiCrdsSpec {
  KubernetesClusterSelector target_cluster = 1;
  optional string version = 2 [
    (dev.planton.shared.options.default) = "v1.2.1",
    (buf.validate.field).string = {
      min_len: 1
      pattern: "^v[0-9]+\\.[0-9]+\\.[0-9]+(-[a-zA-Z0-9]+)?$"
    }
  ];
  InstallChannel install_channel = 3;
}
```

**Files created**:
- `api.proto` - KRM envelope (metadata, spec, status)
- `spec.proto` - Configuration with semver validation
- `stack_input.proto` - IaC module inputs
- `stack_outputs.proto` - Deployment outputs (version, channel, CRD list)

### Registry Entry

Added to `cloud_resource_kind.proto`:
```protobuf
KubernetesGatewayApiCrds = 837 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8sgwcrds"
}];
```

### Validation Tests (15 tests)

Comprehensive test coverage for:
- Version format validation (semver with `v` prefix)
- Invalid version patterns (missing prefix, malformed)
- Channel selection (standard, experimental, unspecified)
- Target cluster configurations (GKE, EKS, AKS)
- Complete configuration scenarios

### Pulumi Module

```
iac/pulumi/
├── main.go           # Entrypoint
├── Pulumi.yaml       # Project config
├── Makefile          # Build automation
└── module/
    ├── main.go       # ConfigFile application
    ├── locals.go     # Version/channel computation
    ├── outputs.go    # Stack exports
    └── vars.go       # Manifest URL constants
```

Key implementation - fetches official manifests from GitHub releases:
```go
crds, err := pulumiyaml.NewConfigFile(ctx, locals.ResourceName,
    &pulumiyaml.ConfigFileArgs{
        File: locals.ManifestURL, // e.g., https://github.com/.../v1.2.1/standard-install.yaml
    },
    pulumi.Provider(kubernetesProvider))
```

### Terraform Module

```
iac/tf/
├── variables.tf      # Mirrors spec.proto
├── locals.tf         # URL computation
├── provider.tf       # kubernetes + kubectl providers
├── main.tf           # http data + kubectl_manifest
├── outputs.tf        # version, channel, CRD list
└── README.md         # Usage docs
```

### Documentation

| File | Purpose | Lines |
|------|---------|-------|
| `README.md` | User-facing overview | ~150 |
| `examples.md` | 8 YAML examples for different scenarios | ~200 |
| `docs/README.md` | Research: Gateway API landscape, architecture | ~430 |
| `iac/pulumi/README.md` | Pulumi usage guide | ~80 |
| `iac/tf/README.md` | Terraform usage guide | ~120 |

## Benefits

### For Platform Teams
- **Declarative CRD management** - Manage Gateway API like any other infrastructure
- **Version pinning** - Ensure consistency across all clusters
- **Audit trail** - CRD installation tracked in IaC state
- **Multi-cluster deployment** - Same manifest works everywhere

### For Application Teams
- **Prerequisite clarity** - Clear dependency for Gateway implementations
- **Self-service** - Request CRD installation via manifest
- **Version awareness** - Know exactly which Gateway API features are available

### Operational Benefits
- **Upgrade path** - Change version in manifest, apply to all clusters
- **Rollback capability** - IaC state enables reverting to previous versions
- **Consistency** - No more "works on my cluster" issues

## File Summary

| Category | Files | Description |
|----------|-------|-------------|
| Proto definitions | 4 | api, spec, stack_input, stack_outputs |
| Generated stubs | 4 | .pb.go files |
| Tests | 1 | 15 validation tests |
| Documentation | 5 | README, examples, research, module docs |
| Pulumi | 7 | Entrypoint + module + supporting files |
| Terraform | 6 | Full module with docs |
| Test manifest | 1 | iac/hack/manifest.yaml |
| **Total** | **28** | Complete component |

## Impact

### Users
- Can now deploy Gateway API CRDs via `planton pulumi up`
- Version and channel selection via simple YAML configuration
- Works with any Kubernetes cluster (GKE, EKS, AKS, DOKS, Civo)

### Ecosystem
- Enables adoption of Gateway API in Planton workflows
- Foundation for future Gateway implementations (Istio, Envoy Gateway integrations)
- Consistent with existing Kubernetes addon components

### Codebase
- Follows established deployment component patterns
- Reuses existing infrastructure (KubernetesClusterSelector, provider config)
- No breaking changes to existing components

## Related Work

This component serves as a prerequisite for Gateway API implementations:
- Future: `KubernetesIstioGateway` - Istio Gateway controller
- Future: `KubernetesEnvoyGateway` - Envoy Gateway controller
- Existing: `KubernetesIngressNginx` - Legacy ingress (Gateway API is the successor)

---

**Status**: ✅ Production Ready  
**Timeline**: Single session (~1 hour)  
**Tests**: 15/15 passing  
**Build**: ✅ Compiles successfully
