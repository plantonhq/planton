# KubernetesTekton: Manifest-Based Tekton Deployment Component

**Date**: December 20, 2025
**Type**: Feature
**Components**: Kubernetes Provider, API Definitions, Pulumi CLI Integration, IAC Stack Runner

## Summary

Created a new `KubernetesTekton` deployment component that deploys Tekton Pipelines and Dashboard using official release manifests directly (kubectl apply style), as an alternative to the existing operator-based `KubernetesTektonOperator`. This component provides a simpler deployment model with CloudEvents integration for pipeline notifications and Gateway API-based dashboard ingress.

## Problem Statement / Motivation

Users deploying Tekton on Kubernetes had only one option: the operator-based `KubernetesTektonOperator`. While the operator provides lifecycle management and automated upgrades, some users prefer:

### Pain Points

- **Simplicity**: Direct manifest deployment is easier to understand and debug
- **Transparency**: Users can see exactly what's being deployed without operator abstraction
- **Speed**: No operator overhead, faster initial deployment
- **Flexibility**: Direct control over configuration via ConfigMap patches
- **CloudEvents Integration**: Previously required manual ConfigMap editing to configure `default-cloud-events-sink`
- **Dashboard Access**: No built-in ingress support for the Tekton Dashboard

## Solution / What's New

Introduced `KubernetesTekton` as a new deployment component that:

1. Deploys Tekton Pipelines and Dashboard using official release manifests
2. Provides declarative CloudEvents sink configuration
3. Exposes the dashboard via Kubernetes Gateway API with TLS

### Deployment Approaches Comparison

| Feature | KubernetesTekton | KubernetesTektonOperator |
|---------|------------------|-------------------------|
| Deployment Method | Direct manifests | Operator + TektonConfig CRD |
| Complexity | Simpler | More abstraction |
| Lifecycle Management | Manual | Automated |
| Configuration | ConfigMap patches | TektonConfig CR |
| Best For | Simple deployments | Production automation |

### API Design

The spec uses a clean nested structure for dashboard configuration:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: my-tekton
spec:
  pipeline_version: "v0.65.2"
  dashboard:
    enabled: true
    version: "v0.53.0"
    ingress:
      enabled: true
      hostname: "tekton-dashboard.example.com"
  cloud_events:
    sink_url: "http://receiver.ns.svc.cluster.local/events"
```

## Implementation Details

### Proto Schema

**File**: `apis/dev/planton/provider/kubernetes/kubernetestekton/v1/spec.proto`

```protobuf
message KubernetesTektonSpec {
  KubernetesClusterSelector target_cluster = 1;
  string pipeline_version = 2;
  KubernetesTektonDashboard dashboard = 3;
  KubernetesTektonCloudEvents cloud_events = 4;
}

message KubernetesTektonDashboard {
  bool enabled = 1;
  string version = 2;
  KubernetesTektonDashboardIngress ingress = 3;
}

message KubernetesTektonDashboardIngress {
  bool enabled = 1;
  string hostname = 2;
  // CEL validation: hostname required when enabled
}

message KubernetesTektonCloudEvents {
  string sink_url = 1;
  // CEL validation: must be HTTP/HTTPS URL
}
```

### Pulumi Module

**Deployment Flow** (`iac/pulumi/module/main.go`):

```
1. Deploy Tekton Pipelines manifests
   └── Creates tekton-pipelines namespace, CRDs, controllers

2. Deploy Tekton Dashboard manifests (if enabled)
   └── Adds web UI service on port 9097

3. Configure CloudEvents (if sink_url specified)
   └── Patches config-defaults ConfigMap

4. Create Dashboard Ingress (if enabled)
   └── Certificate → Gateway → HTTPRoutes
```

**Manifest Deployment** (`iac/pulumi/module/tekton.go`):

Uses `yamlv2.NewConfigFile` to apply official release manifests:

```go
pipelineManifests, err := yamlv2.NewConfigFile(ctx, "tekton-pipelines", &yamlv2.ConfigFileArgs{
    File: pulumi.String(locals.PipelineManifestURL),
}, pulumi.Provider(kubernetesProvider))
```

**CloudEvents Configuration** (`iac/pulumi/module/config.go`):

Patches the `config-defaults` ConfigMap to set the sink URL:

```go
_, err := corev1.NewConfigMapPatch(ctx, "tekton-config-defaults-patch", &corev1.ConfigMapPatchArgs{
    Metadata: metav1.ObjectMetaPatchArgs{
        Name:      pulumi.String("config-defaults"),
        Namespace: pulumi.String(locals.Namespace),
    },
    Data: pulumi.StringMap{
        "default-cloud-events-sink": pulumi.String(locals.CloudEventsSinkURL),
    },
}, ...)
```

**Ingress Pattern** (`iac/pulumi/module/ingress.go`):

Follows the established Planton pattern (same as Solr, Temporal):

```
Certificate (cert-manager)
    ↓
Gateway (istio gatewayClassName)
    ├── HTTPS listener (port 443, TLS terminate)
    └── HTTP listener (port 80, for redirect)
    ↓
HTTPRoute (http-redirect)
    └── 301 redirect to HTTPS
    ↓
HTTPRoute (https-backend)
    └── Routes to tekton-dashboard:9097
```

### Terraform Module

Feature parity with Pulumi implementation:

- `variables.tf`: Nested dashboard object structure
- `locals.tf`: Computed values (manifest URLs, dashboard_enabled, ingress config)
- `main.tf`: kubectl_manifest resources for pipeline/dashboard deployment
- `ingress.tf`: kubernetes_manifest resources for Gateway API
- `outputs.tf`: Stack outputs (namespace, versions, endpoints)

### Registry Entry

**File**: `apis/dev/planton/shared/cloudresourcekind/cloud_resource_kind.proto`

```protobuf
KubernetesTekton = 839 [(kind_meta) = {
  provider: kubernetes
  version: v1
  id_prefix: "k8stktn"
}];
```

## Files Created/Modified

```
apis/dev/planton/provider/kubernetes/kubernetestekton/
└── v1/
    ├── api.proto                    # KRM wrapper (KubernetesTekton message)
    ├── spec.proto                   # Spec with nested dashboard config
    ├── stack_input.proto            # IaC stack input
    ├── stack_outputs.proto          # Deployment outputs
    ├── spec_test.go                 # 8 validation tests
    ├── BUILD.bazel                  # Bazel build config
    ├── README.md                    # User documentation
    ├── examples.md                  # 5 usage examples
    ├── docs/README.md               # Research documentation
    └── iac/
        ├── hack/manifest.yaml       # Test manifest
        ├── pulumi/
        │   ├── main.go              # Entrypoint
        │   ├── Pulumi.yaml          # Project config
        │   ├── Makefile             # Build helpers
        │   ├── README.md            # Module docs
        │   └── module/
        │       ├── vars.go          # Constants
        │       ├── locals.go        # Computed values
        │       ├── main.go          # Orchestration
        │       ├── tekton.go        # Manifest deployment
        │       ├── config.go        # CloudEvents config
        │       ├── ingress.go       # Gateway API resources
        │       └── outputs.go       # Output documentation
        └── tf/
            ├── variables.tf         # Input variables
            ├── locals.tf            # Local values
            ├── main.tf              # Manifest deployment
            ├── ingress.tf           # Gateway API resources
            ├── outputs.tf           # Module outputs
            ├── provider.tf          # Provider requirements
            └── README.md            # Module docs
```

## Benefits

### For Users

- **Simpler Mental Model**: Deploy Tekton with a single manifest, no operator to understand
- **CloudEvents Out-of-the-Box**: Declarative sink configuration without manual ConfigMap editing
- **Dashboard Access**: Production-ready ingress with TLS, no manual Gateway/Route creation
- **Version Pinning**: Explicit control over pipeline and dashboard versions

### For Operators

- **Faster Debugging**: Direct manifest deployment makes troubleshooting easier
- **No Operator Maintenance**: One less controller to monitor and upgrade
- **Predictable Behavior**: What you deploy is what you get

### For Developers

- **Clean API Design**: Nested dashboard configuration is intuitive
- **Consistent Patterns**: Follows established Planton component structure
- **Full IaC Parity**: Both Pulumi and Terraform implementations

## Stack Outputs

| Output | Description |
|--------|-------------|
| `namespace` | Always `tekton-pipelines` |
| `pipeline_version` | Deployed Tekton Pipelines version |
| `dashboard_version` | Deployed Dashboard version (if enabled) |
| `dashboard_internal_endpoint` | `tekton-dashboard.tekton-pipelines.svc.cluster.local:9097` |
| `dashboard_external_hostname` | External hostname (if ingress enabled) |
| `port_forward_dashboard_command` | `kubectl port-forward -n tekton-pipelines service/tekton-dashboard 9097:9097` |
| `cloud_events_sink_url` | Configured CloudEvents sink URL |

## Usage Examples

### Minimal Deployment

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: tekton-minimal
spec:
  pipeline_version: "latest"
```

### Production Setup

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTekton
metadata:
  name: tekton-prod
  org: my-org
  env: prod
spec:
  pipeline_version: "v0.65.2"
  dashboard:
    enabled: true
    version: "v0.53.0"
    ingress:
      enabled: true
      hostname: "tekton-dashboard.planton.live"
  cloud_events:
    sink_url: "http://service-hub.platform.svc.cluster.local/tekton/cloud-event"
```

## Impact

### Component Ecosystem

- Provides alternative to `KubernetesTektonOperator` for different use cases
- Expands Kubernetes provider coverage for CI/CD tooling
- Demonstrates CloudEvents integration pattern reusable by other components

### User Choice

Users can now choose based on their needs:
- **KubernetesTekton**: Simple deployments, debugging, development
- **KubernetesTektonOperator**: Production with automated lifecycle management

## Related Work

- **KubernetesTektonOperator**: Operator-based Tekton deployment (existing component)
- **Solr/Temporal Ingress**: Gateway API pattern reused from these components
- **CloudEvents**: Pattern applicable to other event-emitting Kubernetes workloads

## Testing

### Validation Tests

8 test cases covering:
- Basic configuration validation
- CloudEvents sink URL format (HTTP/HTTPS)
- Empty sink URL (allowed)
- Dashboard ingress hostname requirement
- Disabled ingress without hostname (allowed)
- Invalid URL scheme rejection

```bash
go test ./apis/dev/planton/provider/kubernetes/kubernetestekton/v1/... -v
# 8/8 tests passing
```

### Build Verification

```bash
make protos  # Proto generation ✅
go vet ./...  # No errors ✅
make build    # Full build ✅
```

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation
