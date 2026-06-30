# Deploying Raw Kubernetes Manifests: From kubectl apply to IaC Abstraction

## Introduction

Every Kubernetes practitioner knows the fundamental command: `kubectl apply -f manifest.yaml`. It's the gateway to Kubernetes—simple, direct, and powerful. Yet this simplicity masks significant operational challenges when deploying at scale: no state tracking, no rollback mechanism, no dependency ordering for CRDs, and no integration with broader infrastructure-as-code workflows.

The KubernetesManifest component addresses a specific gap in the Planton ecosystem: the need to deploy arbitrary Kubernetes resources that don't fit into more specialized components. While components like KubernetesDeployment and KubernetesHelmRelease excel at their specific use cases, there's always a need for a "raw" deployment mechanism—an escape hatch that provides all the benefits of IaC (state management, drift detection, dependency tracking) without imposing any abstraction on the manifest content itself.

This document explores the landscape of Kubernetes manifest deployment methods, explains why Planton's approach matters, and details the design decisions behind the KubernetesManifest component.

## The Problem Space

### Why Raw Manifests Still Matter

In an ideal world, every Kubernetes deployment would be neatly packaged—either as a purpose-built API resource or a Helm chart. Reality is messier:

1. **Operator Custom Resources**: Most operators require deploying Custom Resources alongside CRDs. These CRs often don't warrant dedicated components.

2. **Vendor-Provided Manifests**: Third-party vendors frequently provide raw YAML for installation. Wrapping these in Helm charts adds unnecessary complexity.

3. **Infrastructure Resources**: NetworkPolicies, ResourceQuotas, LimitRanges, and PriorityClasses are infrastructure concerns that don't fit application-focused components.

4. **Migration Paths**: Teams migrating from kubectl-based workflows need a bridge to IaC without rewriting all their manifests.

5. **Rapid Prototyping**: During development, engineers often need to deploy test resources quickly without creating new API resources.

### The kubectl apply Trap

The `kubectl apply` command is deceptively simple:

```bash
kubectl apply -f my-resources.yaml
```

But this simplicity hides critical operational gaps:

| Capability | kubectl apply | IaC Solution |
|------------|---------------|--------------|
| State tracking | None (cluster is source of truth) | Full state file |
| Rollback | Manual (apply previous version) | Built-in |
| Drift detection | Manual (`kubectl diff`) | Automatic |
| CI/CD integration | Script-based | Native |
| Dependency ordering | Manual | Automatic |
| Multi-cluster | Manual context switching | Provider-based |
| Audit trail | Kubernetes audit logs only | Full IaC history |

## Evolution of Kubernetes Manifest Deployment

### Phase 1: kubectl and Shell Scripts (2015-2017)

The earliest Kubernetes deployments were entirely kubectl-based:

```bash
#!/bin/bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

**Problems**:
- No error handling
- No idempotency guarantees
- Manual ordering required
- No rollback mechanism

### Phase 2: Kustomize (2017-2019)

Kustomize introduced "template-free" customization:

```yaml
# kustomization.yaml
resources:
  - deployment.yaml
  - service.yaml
patchesStrategicMerge:
  - production-patch.yaml
```

**Improvements**:
- Environment-specific overlays
- No templating language required
- Native kubectl integration (`kubectl apply -k`)

**Remaining gaps**:
- Still no state tracking
- No dependency management
- No programmatic access

### Phase 3: Helm (2016-present)

Helm introduced package management:

```bash
helm install my-release ./my-chart -f values.yaml
```

**Improvements**:
- Release versioning
- Rollback support
- Templating for reuse
- Dependency declaration

**Trade-offs**:
- Go templating complexity
- Chart abstraction overhead
- Not suitable for one-off resources

### Phase 4: GitOps Controllers (2019-present)

ArgoCD and Flux brought continuous reconciliation:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  source:
    repoURL: https://github.com/org/repo
    path: manifests/
```

**Improvements**:
- Git as source of truth
- Continuous drift detection
- Self-healing clusters

**Trade-offs**:
- Additional cluster components
- Git-centric workflow required
- Complex for simple use cases

### Phase 5: IaC Integration (2020-present)

Terraform and Pulumi brought manifest deployment into broader IaC:

```hcl
# Terraform
resource "kubernetes_manifest" "my_resource" {
  manifest = yamldecode(file("resource.yaml"))
}
```

```go
// Pulumi
yaml.NewConfigFile(ctx, "resources", &yaml.ConfigFileArgs{
    File: pulumi.String("resources.yaml"),
})
```

**Improvements**:
- Full state management
- Cross-provider orchestration
- Programmatic control
- Unified workflows

## Deployment Methods Landscape

### Level 0: kubectl apply (Manual)

**Workflow**:
```bash
# Direct application
kubectl apply -f manifest.yaml

# With namespace
kubectl apply -f manifest.yaml -n my-namespace

# Multi-file
kubectl apply -f ./manifests/
```

**Pros**:
- Zero tooling overhead
- Immediate feedback
- Universal availability

**Cons**:
- No state tracking
- Manual error recovery
- No dependency ordering
- No drift detection

**Verdict**: Suitable for ad-hoc debugging and learning. Not for production workflows.

### Level 1: Kustomize

**Workflow**:
```bash
# Build and apply
kubectl apply -k overlays/production/

# Preview output
kubectl kustomize overlays/production/
```

**Example structure**:
```
base/
  deployment.yaml
  service.yaml
  kustomization.yaml
overlays/
  production/
    kustomization.yaml
    replica-patch.yaml
```

**Pros**:
- Template-free customization
- Native kubectl support
- Clean separation of concerns

**Cons**:
- Still no state management
- Verbose for complex scenarios
- No programmatic access

**Verdict**: Good for environment-specific configuration. Insufficient for production operations.

### Level 2: Terraform kubernetes_manifest

**Example**:
```hcl
resource "kubernetes_manifest" "configmap" {
  manifest = {
    apiVersion = "v1"
    kind       = "ConfigMap"
    metadata = {
      name      = "my-config"
      namespace = "default"
    }
    data = {
      key = "value"
    }
  }
}

# Or from YAML file
resource "kubernetes_manifest" "from_file" {
  manifest = yamldecode(file("${path.module}/resources/my-resource.yaml"))
}
```

**Pros**:
- Full Terraform state management
- Cross-provider orchestration
- Plan/Apply workflow

**Cons**:
- One resource per manifest block
- Complex for multi-document YAML
- HCL conversion overhead

**Verdict**: Good for infrastructure resources. Awkward for complex applications.

### Level 3: Pulumi yaml/v2

**Example**:
```go
// Single file
_, err := yamlv2.NewConfigFile(ctx, "resources", &yamlv2.ConfigFileArgs{
    File: pulumi.String("resources.yaml"),
})

// Inline YAML
_, err := yamlv2.NewConfigGroup(ctx, "inline", &yamlv2.ConfigGroupArgs{
    Yaml: pulumi.StringPtr(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
data:
  key: value
`),
})
```

**Pros**:
- Full Pulumi state management
- CRD-aware ordering
- Multi-document support
- Programmatic control

**Cons**:
- Requires Pulumi infrastructure
- Go/TypeScript/Python knowledge needed

**Verdict**: Production-ready with proper dependency handling.

## Comparative Analysis

| Method | State Management | CRD Ordering | Multi-Doc | Rollback | CI/CD Integration |
|--------|-----------------|--------------|-----------|----------|-------------------|
| kubectl apply | ❌ | ❌ | ✅ | ❌ | ⚠️ Script-based |
| Kustomize | ❌ | ❌ | ✅ | ❌ | ⚠️ Script-based |
| Terraform | ✅ | ⚠️ Manual | ⚠️ Complex | ✅ | ✅ Native |
| Pulumi yaml/v2 | ✅ | ✅ Auto | ✅ Native | ✅ | ✅ Native |
| Planton | ✅ | ✅ Auto | ✅ Native | ✅ | ✅ Native |

## The Planton Approach

### Design Philosophy

KubernetesManifest follows Planton's core principle: **make the correct choice the easy choice**. This means:

1. **Zero abstraction**: The manifest YAML is applied exactly as written
2. **Automatic CRD ordering**: No manual dependency management required
3. **Unified experience**: Same API pattern as all other components
4. **State management**: Full Pulumi state tracking and drift detection

### Why Not Just Use Pulumi Directly?

While Pulumi's yaml/v2 is excellent, using it directly requires:
- Setting up Pulumi projects
- Managing provider configuration
- Writing Go/TypeScript/Python code
- Handling credentials and state backends

KubernetesManifest wraps this complexity in a declarative API:

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesManifest
metadata:
  name: my-resources
spec:
  namespace: my-namespace
  create_namespace: true
  manifest_yaml: |
    apiVersion: v1
    kind: ConfigMap
    ...
```

This provides:
- Declarative YAML interface
- Automatic provider setup
- Consistent credential handling
- Integration with Planton's ecosystem

### API Design Decisions

**1. manifest_yaml as a string field**

We chose a single string field rather than structured YAML for several reasons:
- Preserves exact formatting and comments
- Supports any Kubernetes resource type
- No schema maintenance for new resource types
- Easy copy-paste from existing manifests

**2. Required namespace field**

The namespace field is required even though manifests can specify their own namespaces:
- Provides a default for resources without explicit namespaces
- Enables namespace creation when needed
- Follows the pattern of other Kubernetes components

**3. Optional target_cluster**

For consistency with other components, target_cluster is optional:
- Defaults to the current cluster context
- Enables multi-cluster deployments when specified
- Supports all Kubernetes cluster types (GKE, EKS, AKS, etc.)

### Implementation Details

**CRD Ordering with yaml/v2**

The critical implementation choice is using Pulumi's yaml/v2 module:

```go
_, err := yamlv2.NewConfigGroup(ctx, "manifest", &yamlv2.ConfigGroupArgs{
    Yaml: pulumi.StringPtr(locals.ManifestYAML),
}, opts...)
```

This provides automatic CRD ordering that older yaml.ConfigFile lacks. When a manifest contains:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
...
---
apiVersion: custom.example.com/v1
kind: MyResource
...
```

yaml/v2 ensures the CRD is fully registered before creating the Custom Resource.

**Namespace Dependency**

When `create_namespace: true`, the manifest depends on the namespace:

```go
if namespaceResource != nil {
    opts = append(opts, pulumi.DependsOn([]pulumi.Resource{namespaceResource}))
}
```

This ensures proper creation order and correct deletion order (resources before namespace).

## Production Best Practices

### When to Use KubernetesManifest

✅ **Good Use Cases**:
- Deploying operator Custom Resources
- Infrastructure resources (NetworkPolicy, ResourceQuota)
- RBAC configurations
- One-off or test resources
- Migrating from kubectl workflows
- Vendor-provided manifests

❌ **Better Alternatives**:
- Microservices → KubernetesDeployment
- Stateful applications → KubernetesStatefulSet
- Helm charts → KubernetesHelmRelease
- Operators → Dedicated operator components

### Manifest Organization

For large deployments, organize manifests logically:

```yaml
# Option 1: Multiple KubernetesManifest resources
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesManifest
metadata:
  name: rbac-resources
spec:
  namespace: my-app
  manifest_yaml: |
    # RBAC resources only
    ...
---
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesManifest
metadata:
  name: network-policies
spec:
  namespace: my-app
  manifest_yaml: |
    # NetworkPolicy resources only
    ...
```

### Security Considerations

1. **Secret handling**: Avoid embedding secrets in manifest_yaml. Use external secret management (External Secrets Operator, Sealed Secrets).

2. **RBAC scope**: Apply principle of least privilege when deploying RBAC resources.

3. **Network policies**: Consider deploying NetworkPolicies via KubernetesManifest to enforce isolation.

### Common Pitfalls

1. **Namespace mismatch**: Resources in the manifest that specify their own namespace won't use the spec.namespace value.

2. **CRD timing**: Even with yaml/v2, very complex CRD hierarchies may need multiple KubernetesManifest resources with explicit ordering.

3. **Large manifests**: Manifests over 1000 lines may hit API limits. Split into multiple resources.

## Conclusion

The KubernetesManifest component fills an essential gap in the Planton ecosystem: deploying arbitrary Kubernetes resources with full IaC benefits. By wrapping Pulumi's yaml/v2 in a declarative API, it provides:

- **Simplicity**: Declare what you want, the platform handles the rest
- **Flexibility**: Deploy any Kubernetes resource type
- **Safety**: Automatic CRD ordering prevents timing issues
- **Consistency**: Same patterns as all other Planton components

For teams migrating from kubectl-based workflows, KubernetesManifest provides a smooth transition path. For teams already using Planton, it's the escape hatch that handles everything the specialized components don't.

The right tool for the job isn't always the most specialized one—sometimes it's the most flexible one that integrates seamlessly with your existing workflow.

## References

- [Kubernetes Documentation: Managing Resources](https://kubernetes.io/docs/concepts/cluster-administration/manage-deployment/)
- [Pulumi Kubernetes yaml/v2 Documentation](https://www.pulumi.com/blog/kubernetes-yaml-v2/)
- [Kustomize Documentation](https://kustomize.io/)
- [Terraform kubernetes_manifest Resource](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest)

