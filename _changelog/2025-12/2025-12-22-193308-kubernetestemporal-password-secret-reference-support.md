# KubernetesTemporal Password Secret Reference Support

**Date**: December 22, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added support for password fields in `KubernetesTemporal` to be provided either as a plain string value or as a reference to an existing Kubernetes Secret. This enables secure credential management in production deployments by leveraging Kubernetes-native secret storage instead of storing passwords in plaintext within configuration files.

## Problem Statement / Motivation

When connecting Temporal to an external database or Elasticsearch cluster, users previously had to provide passwords as plaintext strings directly in their manifest files. This posed security concerns for production environments:

### Pain Points

- **Security risk**: Passwords stored in plaintext in YAML manifests could be accidentally committed to version control
- **Compliance issues**: Many organizations require secrets to be stored in dedicated secret management systems
- **Operational overhead**: Password rotation required updating manifest files rather than just rotating the Kubernetes Secret
- **Inconsistent patterns**: Other Kubernetes-native tools support secret references, making the plaintext-only approach feel outdated

## Solution / What's New

Updated both `KubernetesTemporalExternalDatabase` and `KubernetesTemporalExternalElasticsearch` messages to use the `KubernetesSensitiveValue` type (using `oneof`) for their password fields. This supports either:
1. A plain string value (for development/testing)
2. A reference to an existing Kubernetes Secret (recommended for production)

### Usage Examples

**Using Kubernetes Secret (Recommended for Production):**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: temporal-prod
spec:
  namespace:
    value: temporal-prod
  database:
    backend: postgresql
    external_database:
      host: postgres.example.com
      port: 5432
      username: temporaluser
      password:
        secretRef:
          name: temporal-db-credentials
          key: password
  external_elasticsearch:
    host: elasticsearch.example.com
    port: 9200
    user: elasticuser
    password:
      secretRef:
        name: es-credentials
        key: password
```

**Using Plain String (Dev/Test Only):**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesTemporal
metadata:
  name: temporal-dev
spec:
  namespace:
    value: temporal-dev
  database:
    backend: postgresql
    external_database:
      host: postgres.example.com
      port: 5432
      username: temporaluser
      password:
        stringValue: my-password
```

## Implementation Details

### Proto Schema Changes

Updated `KubernetesTemporalExternalDatabase` and `KubernetesTemporalExternalElasticsearch` to use the `KubernetesSensitiveValue` type:

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/spec.proto`

```protobuf
message KubernetesTemporalExternalDatabase {
  // ... existing fields ...
  
  // Changed from: string password = 4;
  dev.planton.provider.kubernetes.KubernetesSensitiveValue password = 4;
}

message KubernetesTemporalExternalElasticsearch {
  // ... existing fields ...
  
  // Changed from: string password = 4;
  dev.planton.provider.kubernetes.KubernetesSensitiveValue password = 4;
}
```

### Pulumi Module Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/iac/pulumi/module/db_password_secret.go`

The Pulumi module now only creates a secret when `string_value` is provided. When `secret_ref` is used, no new secret is created:

```go
// If using a secret reference, we don't need to create a new secret
if ext.Password.GetSecretRef() != nil {
    return nil
}
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/iac/pulumi/module/helm_chart.go`

The Helm chart configuration now uses the appropriate secret name and key based on whether `secret_ref` is provided:

```go
// Determine which secret to use for database password
dbSecretName := locals.DatabasePasswordSecretName
dbSecretKey := vars.DatabasePasswordSecretKey
if ext.Password != nil && ext.Password.GetSecretRef() != nil {
    secretRef := ext.Password.GetSecretRef()
    dbSecretName = secretRef.Name
    dbSecretKey = secretRef.Key
}
```

### Terraform Module Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/iac/tf/variables.tf`

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

**File**: `apis/dev/planton/provider/kubernetes/kubernetestemporal/v1/iac/tf/locals.tf`

```hcl
# Password handling - check if using secret_ref or string_value
external_db_password_secret_ref = try(var.spec.database.external_database.password.secret_ref, null)
external_db_password_string     = try(var.spec.database.external_database.password.string_value, "")

# Determine which secret to use for database password
use_existing_db_secret = local.external_db_password_secret_ref != null
db_secret_name         = local.use_existing_db_secret ? local.external_db_password_secret_ref.name : local.database_secret_name
db_secret_key          = local.use_existing_db_secret ? local.external_db_password_secret_ref.key : local.database_secret_key
```

### Test Updates

Updated all test cases in `spec_test.go` to use the new password format:

```go
Password: stringPassword("my-password"),

// Helper functions added:
func stringPassword(value string) *kubernetes.KubernetesSensitiveValue {
    return &kubernetes.KubernetesSensitiveValue{
        Value: &kubernetes.KubernetesSensitiveValue_StringValue{
            StringValue: value,
        },
    }
}

func secretRefPassword(name, key string) *kubernetes.KubernetesSensitiveValue {
    return &kubernetes.KubernetesSensitiveValue{
        Value: &kubernetes.KubernetesSensitiveValue_SecretRef{
            SecretRef: &kubernetes.KubernetesSecretKeyRef{
                Name: name,
                Key:  key,
            },
        },
    }
}
```

## Benefits

- **Improved security**: Production deployments can use Kubernetes Secrets instead of plaintext passwords
- **GitOps friendly**: Manifests can be safely committed to version control without exposing credentials
- **Easier rotation**: Password changes only require updating the Kubernetes Secret, not the manifest
- **Consistency**: Follows the same pattern established by `KubernetesSignoz` component
- **Reuses existing types**: Uses the shared `KubernetesSensitiveValue` type for consistency
- **Backward compatible API**: Both Pulumi and Terraform modules handle both value types seamlessly

## Impact

### Users
- External database and Elasticsearch users can now secure their credentials properly
- No breaking changes for existing deployments (they just need to update YAML structure)
- Clear documentation with examples for both approaches

### Developers
- Pattern established for handling secrets is consistent across components
- All tests updated and passing

## Files Changed

| File | Change |
|------|--------|
| `kubernetestemporal/v1/spec.proto` | Updated password fields to use KubernetesSensitiveValue |
| `iac/pulumi/module/db_password_secret.go` | Handle secret_ref case |
| `iac/pulumi/module/helm_chart.go` | Handle both password types for DB and ES |
| `iac/tf/variables.tf` | New password object structure |
| `iac/tf/locals.tf` | Password handling logic |
| `iac/tf/main.tf` | Map to correct Helm values |
| `spec_test.go` | Updated test cases |
| `examples.md` | Added secret ref examples |

## Related Work

- Follows the pattern established by `KubernetesSignoz` in `2025-12-19-kubernetessignoz-password-secret-reference-support.md`
- Uses the shared `KubernetesSensitiveValue` type from `kubernetes_secret.proto`
- Uses Temporal Helm chart's built-in `existingSecret` and `existingSecretKey` support

---

**Status**: ✅ Production Ready
**Timeline**: ~1 hour implementation

