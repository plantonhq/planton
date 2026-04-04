# Overview

The **KubernetesManifest** Pulumi module provides a straightforward way to deploy raw Kubernetes manifests to a cluster. Unlike more specialized components (KubernetesDeployment, KubernetesStatefulSet), this module takes arbitrary YAML and applies it directly, making it ideal for custom resources, operator CRDs, or any Kubernetes configuration that doesn't fit into predefined patterns.

## Module Architecture

```
KubernetesManifestStackInput
├── Target (KubernetesManifest resource)
│   ├── Metadata (name, labels)
│   └── Spec
│       ├── TargetCluster (optional cluster selector)
│       ├── Namespace (required)
│       ├── CreateNamespace (bool)
│       └── ManifestYaml (raw YAML string)
└── ProviderConfig (Kubernetes credentials)
```

### File Organization

```
iac/pulumi/
├── main.go          # Entrypoint - loads stack input, calls module
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build and dependency management
└── module/
    ├── main.go      # Resource orchestration
    ├── locals.go    # Input transformation and label setup
    └── outputs.go   # Output constant definitions
```

## Key Components

### Controller (module/main.go)

The main controller orchestrates resource creation:

1. **Initialize Locals**: Transform input spec into internal structures
2. **Create Provider**: Set up Kubernetes provider with credentials
3. **Create Namespace** (conditional): If `create_namespace: true`
4. **Apply Manifest**: Use yaml/v2.ConfigGroup to apply the YAML

### Locals (module/locals.go)

Transforms the input spec into:

- **Labels**: Standard OpenMCF resource labels
- **Namespace**: Extracted from spec
- **ManifestYAML**: Raw YAML string for deployment

### Outputs (module/outputs.go)

Defines exported values:

- `namespace`: The namespace where resources were deployed

## Resource Flow

```
StackInput (spec + credentials)
    ↓
initializeLocals()
    ↓
    ├── Extract namespace from spec
    ├── Build resource labels
    └── Store manifest YAML
    ↓
Resources()
    ↓
    ├── Create Kubernetes Provider
    ├── (Optional) Create Namespace
    └── Apply Manifest via yamlv2.ConfigGroup
    ↓
Outputs (namespace)
```

## Design Decisions

### Why yaml/v2?

The older `yaml.ConfigFile` has known issues with CRD timing—if a manifest contains both a CRD and Custom Resources that use it, the Custom Resources may fail because the CRD isn't registered yet.

`yaml/v2.ConfigGroup` solves this by:
- Analyzing the manifest for CRDs
- Applying CRDs first and waiting for registration
- Then applying dependent resources

### Why ConfigGroup over ConfigFile?

- **ConfigGroup** accepts a YAML string directly
- **ConfigFile** requires a file path
- Since our manifest comes from the API spec as a string, ConfigGroup is the natural choice

### Namespace Dependency

When `create_namespace: true`, the ConfigGroup depends on the namespace resource. This ensures:
1. Namespace exists before resources are created
2. Pulumi correctly tracks the dependency graph
3. Deletion happens in reverse order (resources before namespace)

## Customization Guide

### Adding Labels to Manifest Resources

The module does NOT modify the manifest YAML. If you need labels on resources, add them in the manifest itself:

```yaml
metadata:
  labels:
    app.kubernetes.io/managed-by: openmcf
```

### Changing Default Namespace

Resources in the manifest that don't specify a namespace will be created in the cluster's default namespace (usually `default`), NOT the namespace specified in the spec. To use the spec namespace:

```yaml
# In manifest_yaml
metadata:
  namespace: {{ use the same value as spec.namespace }}
```

### Handling Large Manifests

For very large manifests (1000+ lines), consider:
1. Splitting into multiple KubernetesManifest resources
2. Using KubernetesHelmRelease if the manifest is from a Helm chart

## Common Patterns

### Naming Conventions

- Resource names follow the format: `{metadata.name}-{purpose}`
- Pulumi resource names are kept short for readability

### Label Management

Standard OpenMCF labels are applied to the namespace (if created):
- `planton.ai/resource: true`
- `planton.ai/resource-name: {metadata.name}`
- `planton.ai/resource-kind: KubernetesManifest`

### Error Handling

The module uses `errors.Wrap` for contextual error messages:
- Provider creation failures
- Namespace creation failures
- Manifest application failures

Each error includes context about what was being attempted.

