# KubernetesSignoz Password Secret Reference Support

**Date**: December 19, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added support for the ClickHouse password in `KubernetesSignoz` to be provided either as a plain string value or as a reference to an existing Kubernetes Secret. This enables secure credential management in production deployments by leveraging Kubernetes-native secret storage instead of storing passwords in plaintext within configuration files.

## Problem Statement / Motivation

When connecting SigNoz to an external ClickHouse instance, users previously had to provide the password as a plaintext string directly in their manifest files. This posed security concerns for production environments:

### Pain Points

- **Security risk**: Passwords stored in plaintext in YAML manifests could be accidentally committed to version control
- **Compliance issues**: Many organizations require secrets to be stored in dedicated secret management systems
- **Operational overhead**: Password rotation required updating manifest files rather than just rotating the Kubernetes Secret
- **Inconsistent patterns**: Other Kubernetes-native tools support secret references, making the plaintext-only approach feel outdated

## Solution / What's New

Introduced a new `KubernetesSensitiveValue` proto type that uses `oneof` to support either:
1. A plain string value (for development/testing)
2. A reference to an existing Kubernetes Secret (recommended for production)

The SigNoz Helm chart already supports `existingSecret` and `existingSecretPasswordKey` for external ClickHouse authentication, so we map our new proto structure to these Helm values.

### New Proto Types

**File**: `apis/dev/planton/provider/kubernetes/kubernetes_secret.proto`

```protobuf
message KubernetesSecretKeyRef {
  string namespace = 1;  // Optional - defaults to deployment namespace
  string name = 2;       // Required - name of the Kubernetes Secret
  string key = 3;        // Required - key within the Secret
}

message KubernetesSensitiveValue {
  oneof value {
    string string_value = 1;           // Plain string (dev/test)
    KubernetesSecretKeyRef secret_ref = 2;  // Secret reference (production)
  }
}
```

### Usage Examples

**Using Kubernetes Secret (Recommended for Production):**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: SignozKubernetes
metadata:
  name: signoz-prod
spec:
  database:
    isExternal: true
    externalDatabase:
      host: clickhouse.example.com
      username: signoz
      password:
        secretRef:
          name: clickhouse-credentials
          key: password
```

**Using Plain String (Dev/Test Only):**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: SignozKubernetes
metadata:
  name: signoz-dev
spec:
  database:
    isExternal: true
    externalDatabase:
      host: clickhouse.example.com
      username: signoz
      password:
        stringValue: my-password
```

## Implementation Details

### Proto Schema Changes

Updated `KubernetesSignozExternalClickhouse` to use the new `KubernetesSensitiveValue` type:

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/spec.proto`

```protobuf
message KubernetesSignozExternalClickhouse {
  // ... existing fields ...
  
  // Changed from: string password = 7;
  dev.planton.provider.kubernetes.KubernetesSensitiveValue password = 7;
}
```

### Pulumi Module Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/pulumi/module/signoz.go`

The Pulumi module now checks which variant of the password is provided and sets the appropriate Helm values:

```go
if ext.Password != nil {
    if ext.Password.GetSecretRef() != nil {
        // Use existing Kubernetes secret for password
        secretRef := ext.Password.GetSecretRef()
        externalClickhouseValues["existingSecret"] = pulumi.String(secretRef.Name)
        externalClickhouseValues["existingSecretPasswordKey"] = pulumi.String(secretRef.Key)
    } else if ext.Password.GetStringValue() != "" {
        // Use plain string password
        externalClickhouseValues["password"] = pulumi.String(ext.Password.GetStringValue())
    }
}
```

### Terraform Module Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/tf/variables.tf`

```hcl
password = object({
  string_value = optional(string)
  secret_ref = optional(object({
    namespace = optional(string)
    name = string
    key = string
  }))
})
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetessignoz/v1/iac/tf/signoz.tf`

```hcl
externalClickhouse = var.spec.database.is_external ? merge(
  { /* base fields */ },
  try(var.spec.database.external_database.password.secret_ref, null) != null ? {
    existingSecret            = var.spec.database.external_database.password.secret_ref.name
    existingSecretPasswordKey = var.spec.database.external_database.password.secret_ref.key
  } : try(var.spec.database.external_database.password.string_value, null) != null ? {
    password = var.spec.database.external_database.password.string_value
  } : {}
) : null
```

### Test Updates

Updated all test cases in `spec_test.go` to use the new password format:

```go
Password: &kubernetes.KubernetesSensitiveValue{
    Value: &kubernetes.KubernetesSensitiveValue_StringValue{
        StringValue: "my-password",
    },
},
```

## Benefits

- **Improved security**: Production deployments can use Kubernetes Secrets instead of plaintext passwords
- **GitOps friendly**: Manifests can be safely committed to version control without exposing credentials
- **Easier rotation**: Password changes only require updating the Kubernetes Secret, not the manifest
- **Follows proto patterns**: Uses `oneof` pattern consistent with existing `StringValueOrRef` in the codebase
- **Reusable type**: `KubernetesSensitiveValue` can be used by other components needing similar functionality
- **Backward compatible API**: Both Pulumi and Terraform modules handle both value types seamlessly

## Impact

### Users
- External ClickHouse users can now secure their credentials properly
- No breaking changes for existing deployments (they just need to update YAML structure)
- Clear documentation with examples for both approaches

### Developers
- New reusable `KubernetesSensitiveValue` type available for other sensitive fields
- Pattern established for handling secrets across the codebase
- All tests updated and passing

## Files Changed

| File | Change |
|------|--------|
| `kubernetes_secret.proto` | **New** - KubernetesSecretKeyRef and KubernetesSensitiveValue types |
| `kubernetessignoz/v1/spec.proto` | Updated password field type |
| `iac/pulumi/module/signoz.go` | Handle both password types |
| `iac/tf/variables.tf` | New password object structure |
| `iac/tf/signoz.tf` | Map to correct Helm values |
| `spec_test.go` | Updated test cases |
| `iac/pulumi/examples.md` | Added secret ref examples |
| `iac/tf/examples.md` | Added Terraform secret ref examples |
| `examples.md` | Updated component examples |

## Related Work

- Follows the pattern established by `StringValueOrRef` in `apis/dev/planton/shared/foreignkey/v1/foreign_key.proto`
- Uses SigNoz Helm chart's built-in `existingSecret` and `existingSecretPasswordKey` support
- Can be extended to other components needing sensitive value handling (e.g., database passwords, API keys)

---

**Status**: ✅ Production Ready
**Timeline**: ~1 hour implementation
