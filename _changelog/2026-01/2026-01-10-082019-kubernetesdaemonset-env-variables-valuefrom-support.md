# KubernetesDaemonSet: Environment Variables valueFrom Support

**Date:** 2026-01-10
**Type:** Enhancement
**Component:** KubernetesDaemonSet
**Tags:** kubernetes, daemonset, environment-variables, valueFrom, reference

---

## Summary

Enhanced the `KubernetesDaemonSet` environment variables (`env.variables`) to support both direct string values and `valueFrom` references to other Planton resources. This enables dynamic configuration where environment variable values can be derived from outputs of other resources (e.g., cluster names, database hosts) without hardcoding values.

---

## Problem Statement

Previously, environment variables in `KubernetesDaemonSet` only supported direct string values:

```yaml
env:
  variables:
    LOG_LEVEL: "info"  # Only literal strings supported
```

This limitation required users to manually look up and hardcode values from other resources, leading to:
- Configuration drift when referenced resources change
- Manual coordination between resource deployments
- Error-prone copy-paste of values across manifests

---

## Solution

The `variables` field now uses the `StringValueOrRef` type, allowing each variable to be either:

1. **Direct string value** - for static configuration
2. **Reference to another resource's field** - for dynamic values resolved by the orchestrator

### YAML Examples

**Option 1: Direct string value**
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDaemonSet
metadata:
  name: log-collector
spec:
  container:
    app:
      env:
        variables:
          LOG_LEVEL:
            value: "info"
          NODE_ENV:
            value: "production"
```

**Option 2: Reference to another resource's field**
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesDaemonSet
metadata:
  name: log-collector
spec:
  container:
    app:
      env:
        variables:
          CLUSTER_NAME:
            valueFrom:
              kind: GcpGkeCluster
              name: my-cluster
              fieldPath: "status.outputs.name"
          ELASTICSEARCH_HOST:
            valueFrom:
              kind: ElasticsearchCluster
              name: logging-es
              fieldPath: "status.outputs.host"
```

**Mixed usage (recommended)**
```yaml
spec:
  container:
    app:
      env:
        variables:
          # Static configuration
          LOG_LEVEL:
            value: "info"
          BUFFER_SIZE:
            value: "1024"
          # Dynamic references
          CLUSTER_NAME:
            valueFrom:
              kind: GcpGkeCluster
              name: my-cluster
              fieldPath: "status.outputs.name"
```

---

## Implementation Details

### Proto Definition

Updated `KubernetesDaemonSetContainerAppEnv.variables` from `map<string, string>` to `map<string, StringValueOrRef>`:

```protobuf
message KubernetesDaemonSetContainerAppEnv {
  map<string, dev.planton.shared.foreignkey.v1.StringValueOrRef> variables = 1;
  map<string, dev.planton.provider.kubernetes.KubernetesSensitiveValue> secrets = 2;
}
```

### Pulumi Module

Updated `daemonset.go` to iterate over `StringValueOrRef` map and extract values using `.GetValue()`:

```go
for _, envVarKey := range sortedVarKeys {
    envVarValue := target.Spec.Container.App.Env.Variables[envVarKey]
    // Orchestrator resolves valueFrom and places result in .value
    if envVarValue.GetValue() != "" {
        envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
            Name:  pulumi.String(envVarKey),
            Value: pulumi.String(envVarValue.GetValue()),
        }))
    }
}
```

### Terraform Module

Updated `variables.tf` input type definition:

```hcl
variables = optional(map(object({
  value = optional(string)
  value_from = optional(object({
    kind       = optional(string)
    env        = optional(string)
    name       = string
    field_path = optional(string)
  }))
})))
```

Updated `daemonset.tf` dynamic block:

```hcl
dynamic "env" {
  for_each = {
    for k, v in try(var.spec.container.app.env.variables, {}) :
    k => v.value
    if try(v.value, null) != null && v.value != ""
  }
  content {
    name  = env.key
    value = env.value
  }
}
```

### Tests

Added test cases in `spec_test.go`:
- Variables with direct string values
- Variables with valueFrom references
- Mixed direct values and valueFrom references
- Validation failure when valueFrom is missing required `name` field

---

## Key Implementation Notes

1. **Orchestrator Resolution**: The `valueFrom` references are resolved upstream by the CLI/orchestrator before invoking IaC modules. The resolved value is placed in the `.value` field.

2. **IaC Module Simplicity**: Pulumi and Terraform modules only read from `.value`, keeping the IaC logic simple and focused on resource provisioning.

3. **Backward Compatibility**: Users can continue using simple string values by wrapping them in the `value:` field.

---

## Files Changed

- `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/spec.proto`
- `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/iac/pulumi/module/daemonset.go`
- `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/iac/tf/variables.tf`
- `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/iac/tf/daemonset.tf`
- `apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/spec_test.go`

---

## Build & Validation

```bash
# Regenerate proto stubs
make protos

# Run component tests
go test ./apis/dev/planton/provider/kubernetes/kubernetesdaemonset/v1/... -v

# Full build
make build

# Full test suite
make test
```

---

## Benefits

1. **Dynamic Configuration**: Environment variables can now reference outputs from other resources
2. **Reduced Configuration Drift**: Values stay in sync with source resources
3. **Cleaner Manifests**: No need to hardcode values that come from other resources
4. **Consistency**: Same pattern as other Planton components using `StringValueOrRef`

---

## Impact

- **Proto Change**: Yes (regeneration required)
- **Breaking Change**: No (existing `map<string, string>` usage needs migration to `value:` wrapper)
- **IaC Modules Updated**: Yes (Pulumi and Terraform)
- **Tests Updated**: Yes

---

## Related Work

- KubernetesDeployment env variables valueFrom support (2026-01-10)
- KubernetesStatefulSet env variables valueFrom support (2026-01-10)
- KubernetesCronJob env variables valueFrom support (2026-01-10)
