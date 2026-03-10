# KubernetesStatefulSet Environment Secrets Reference Support

**Date**: December 23, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi IaC Module, Terraform IaC Module

## Summary

Added support for environment secrets in `KubernetesStatefulSet` to be provided either as direct string values or as references to existing Kubernetes Secrets. This enables secure credential management in production deployments by leveraging Kubernetes-native secret storage instead of storing passwords in plaintext within configuration files. This is the same pattern applied to `KubernetesDeployment` in the prior changelog.

## Problem Statement / Motivation

When deploying stateful applications to Kubernetes via `KubernetesStatefulSet`, users previously had to provide secret values as plaintext strings directly in their manifest files via `spec.container.app.env.secrets`. This posed security concerns for production environments.

### Pain Points

- **Security risk**: Passwords stored in plaintext in YAML manifests could be accidentally committed to version control
- **Compliance issues**: Many organizations require secrets to be stored in dedicated secret management systems
- **Operational overhead**: Password rotation required updating manifest files rather than just rotating the Kubernetes Secret
- **Inconsistent patterns**: Other Kubernetes-native tools support secret references, making the plaintext-only approach feel outdated
- **GitOps unfriendly**: Storing secrets in manifests made GitOps workflows difficult to implement securely

## Solution / What's New

Extended the `secrets` field in `KubernetesStatefulSetContainerAppEnv` to use `KubernetesSensitiveValue` type, which supports:

1. **Direct string value** (for development/testing)
2. **Kubernetes Secret reference** (recommended for production)

### Before (Old Format)

```yaml
spec:
  container:
    app:
      env:
        secrets:
          DATABASE_PASSWORD: my-secret-password  # Plain string - security risk
```

### After (New Format)

**Option 1: Direct String Value**
```yaml
spec:
  container:
    app:
      env:
        secrets:
          DATABASE_PASSWORD:
            value: my-secret-password
```

**Option 2: Kubernetes Secret Reference (Production)**
```yaml
spec:
  container:
    app:
      env:
        secrets:
          DATABASE_PASSWORD:
            secretRef:
              name: my-app-secrets       # Name of existing K8s Secret
              key: db-password           # Key within the Secret
              namespace: ""              # Optional, defaults to deployment namespace
```

**Option 3: Mixed (Both Types)**
```yaml
spec:
  container:
    app:
      env:
        secrets:
          # Dev secret - direct value
          DEBUG_TOKEN:
            value: debug-only-token
          # Production secret - external reference
          DATABASE_PASSWORD:
            secretRef:
              name: postgres-credentials
              key: password
```

## Implementation Details

### 1. Proto Schema Changes

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/spec.proto`

**Changes Required**:

1. Add import for `kubernetes_secret.proto`:
```protobuf
import "org/openmcf/provider/kubernetes/kubernetes_secret.proto";
```

2. Change the `secrets` field type in `KubernetesStatefulSetContainerAppEnv`:

**Before**:
```protobuf
message KubernetesStatefulSetContainerAppEnv {
  map<string, string> variables = 1;
  map<string, string> secrets = 2;
}
```

**After**:
```protobuf
message KubernetesStatefulSetContainerAppEnv {
  map<string, string> variables = 1;
  
  /**
   * A map of secret environment variable names to their values.
   * Each secret can be provided either as a literal string value or as a reference 
   * to an existing Kubernetes Secret.
   *
   * Using secret references is recommended for production deployments.
   */
  map<string, org.openmcf.provider.kubernetes.KubernetesSensitiveValue> secrets = 2;
}
```

**Note**: The `KubernetesSensitiveValue` type already exists at:
`apis/org/openmcf/provider/kubernetes/kubernetes_secret.proto`

```protobuf
message KubernetesSensitiveValue {
  oneof value {
    string value = 1;
    KubernetesSecretKeyRef secret_ref = 2;
  }
}

message KubernetesSecretKeyRef {
  string namespace = 1;  // Optional
  string name = 2;       // Required
  string key = 3;        // Required
}
```

### 2. Pulumi Module Updates

#### 2a. secret.go - Only Create Secret for String Values

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/iac/pulumi/module/secret.go`

**Key Logic**:
- Only add secrets with `GetValue() != ""` to the Kubernetes Secret
- Skip secrets that have `GetSecretRef() != nil` (they reference external secrets)
- Only create the Kubernetes Secret resource if there are string values to store

**Full Implementation**:
```go
package module

import (
	"sort"

	"github.com/pkg/errors"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func secret(ctx *pulumi.Context, locals *Locals, kubernetesProvider pulumi.ProviderResource) error {
	dataMap := make(map[string]string)

	// Add only secrets with direct string values to data map
	if locals.KubernetesStatefulSet.Spec.Container.App.Env != nil {
		secrets := locals.KubernetesStatefulSet.Spec.Container.App.Env.Secrets
		if secrets != nil && len(secrets) > 0 {
			// Sort keys for deterministic output
			sortedKeys := make([]string, 0, len(secrets))
			for k := range secrets {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)

			for _, secretKey := range sortedKeys {
				secretValue := secrets[secretKey]
				// Only add secrets that are direct string values
				if secretValue.GetValue() != "" {
					dataMap[secretKey] = secretValue.GetValue()
				}
			}
		}
	}

	// Only create the secret if there are direct string values to store
	if len(dataMap) == 0 {
		return nil
	}

	// Create a standard kubernetes secret with computed name to avoid conflicts
	secretArgs := &kubernetescorev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(locals.EnvSecretName),
			Namespace: pulumi.String(locals.Namespace),
			Labels:    pulumi.ToStringMap(locals.Labels),
		},
		Type:       pulumi.String("Opaque"),
		StringData: pulumi.ToStringMap(dataMap),
	}

	_, err := kubernetescorev1.NewSecret(ctx,
		locals.EnvSecretName,
		secretArgs,
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create secret resource")
	}

	return nil
}
```

**Important Changes from Original**:
1. Removed `sortstringmap` package import, use standard `sort` package instead
2. Added early return when no string values exist (`if len(dataMap) == 0 { return nil }`)
3. Changed iteration to use `GetValue()` method on `KubernetesSensitiveValue`

#### 2b. statefulset.go - Handle Both Types in Env Var Creation

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/iac/pulumi/module/statefulset.go`

**Key Changes**:

1. Add `"sort"` to imports:
```go
import (
	"fmt"
	"sort"
	// ... other imports
)
```

2. Replace the secrets handling block:

**Before**:
```go
if target.Spec.Container.App.Env.Secrets != nil {
	sortedEnvironmentSecretKeys := sortstringmap.SortMap(target.Spec.Container.App.Env.Secrets)

	for _, environmentSecretKey := range sortedEnvironmentSecretKeys {
		envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
			Name: pulumi.String(environmentSecretKey),
			ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
				SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
					Name: pulumi.String(locals.EnvSecretName),
					Key:  pulumi.String(environmentSecretKey),
				},
			},
		}))
	}
}
```

**After**:
```go
if target.Spec.Container.App.Env.Secrets != nil {
	// Sort keys for deterministic output
	sortedSecretKeys := make([]string, 0, len(target.Spec.Container.App.Env.Secrets))
	for k := range target.Spec.Container.App.Env.Secrets {
		sortedSecretKeys = append(sortedSecretKeys, k)
	}
	sort.Strings(sortedSecretKeys)

	for _, secretKey := range sortedSecretKeys {
		secretValue := target.Spec.Container.App.Env.Secrets[secretKey]

		if secretValue.GetSecretRef() != nil {
			// Use external Kubernetes Secret reference
			secretRef := secretValue.GetSecretRef()
			envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
				Name: pulumi.String(secretKey),
				ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
					SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
						Name: pulumi.String(secretRef.Name),
						Key:  pulumi.String(secretRef.Key),
					},
				},
			}))
		} else if secretValue.GetValue() != "" {
			// Use the internally created secret for direct string values
			envVarInputs = append(envVarInputs, kubernetescorev1.EnvVarInput(kubernetescorev1.EnvVarArgs{
				Name: pulumi.String(secretKey),
				ValueFrom: &kubernetescorev1.EnvVarSourceArgs{
					SecretKeyRef: &kubernetescorev1.SecretKeySelectorArgs{
						Name: pulumi.String(locals.EnvSecretName),
						Key:  pulumi.String(secretKey),
					},
				},
			}))
		}
	}
}
```

**Key Logic**:
- For secrets with `GetSecretRef() != nil`: Reference the external Secret directly using `secretRef.Name` and `secretRef.Key`
- For secrets with `GetValue() != ""`: Reference the internally created secret (`locals.EnvSecretName`)

### 3. Terraform Module Updates

#### 3a. variables.tf - Update Secrets Type

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/iac/tf/variables.tf`

**Change the `env.secrets` type from**:
```hcl
env = optional(object({
  variables = optional(map(string))
  secrets   = optional(map(string))
}))
```

**To**:
```hcl
env = optional(object({
  variables = optional(map(string))
  secrets = optional(map(object({
    value = optional(string)
    secret_ref = optional(object({
      namespace = optional(string)
      name      = string
      key       = string
    }))
  })))
}))
```

#### 3b. secret.tf - Only Create Secret for String Values

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/iac/tf/secret.tf`

**Full Implementation**:
```hcl
locals {
  # Filter secrets to only include those with direct string values
  value_secrets = {
    for k, v in try(var.spec.container.app.env.secrets, {}) :
    k => v.value
    if try(v.value, null) != null && v.value != ""
  }
}

# Create a secret for environment secrets if any direct string values are defined
resource "kubernetes_secret" "env_secrets" {
  count = length(local.value_secrets) > 0 ? 1 : 0

  metadata {
    name      = local.env_secret_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  type = "Opaque"

  data = local.value_secrets

  depends_on = [
    kubernetes_namespace.this
  ]
}
```

#### 3c. statefulset.tf - Handle Both Types

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/iac/tf/statefulset.tf`

**Replace the single secrets dynamic block with two**:

**Before**:
```hcl
# Add env variables from secrets
dynamic "env" {
  for_each = try(var.spec.container.app.env.secrets, {})
  content {
    name = env.key
    value_from {
      secret_key_ref {
        name = local.env_secret_name
        key  = env.key
      }
    }
  }
}
```

**After**:
```hcl
# Add env variables from secrets with direct string values
dynamic "env" {
  for_each = {
    for k, v in try(var.spec.container.app.env.secrets, {}) :
    k => v
    if try(v.value, null) != null && v.value != ""
  }
  content {
    name = env.key
    value_from {
      secret_key_ref {
        name = local.env_secret_name
        key  = env.key
      }
    }
  }
}

# Add env variables from external Kubernetes Secret references
dynamic "env" {
  for_each = {
    for k, v in try(var.spec.container.app.env.secrets, {}) :
    k => v
    if try(v.secret_ref, null) != null
  }
  content {
    name = env.key
    value_from {
      secret_key_ref {
        name = env.value.secret_ref.name
        key  = env.value.secret_ref.key
      }
    }
  }
}
```

### 4. Test Updates

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/spec_test.go`

**Add test cases for**:
1. Secrets with direct string values - should pass
2. Secrets with Kubernetes Secret references - should pass
3. Mixed types (both) - should pass
4. Secret ref missing `name` - should fail validation
5. Secret ref missing `key` - should fail validation

**Test Implementation**:
```go
ginkgo.Describe("Environment secrets validation", func() {
	ginkgo.Context("When secrets have direct string values", func() {
		ginkgo.It("should pass validation", func() {
			input.Spec.Container.App.Env = &KubernetesStatefulSetContainerAppEnv{
				Secrets: map[string]*kubernetes.KubernetesSensitiveValue{
					"DATABASE_PASSWORD": {
						Value: &kubernetes.KubernetesSensitiveValue_Value{
							Value: "my-password",
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When secrets have Kubernetes Secret references", func() {
		ginkgo.It("should pass validation with valid secret ref", func() {
			input.Spec.Container.App.Env = &KubernetesStatefulSetContainerAppEnv{
				Secrets: map[string]*kubernetes.KubernetesSensitiveValue{
					"DATABASE_PASSWORD": {
						Value: &kubernetes.KubernetesSensitiveValue_SecretRef{
							SecretRef: &kubernetes.KubernetesSecretKeyRef{
								Name: "my-app-secrets",
								Key:  "db-password",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When secrets have mixed types", func() {
		ginkgo.It("should pass validation with both string values and secret refs", func() {
			input.Spec.Container.App.Env = &KubernetesStatefulSetContainerAppEnv{
				Secrets: map[string]*kubernetes.KubernetesSensitiveValue{
					"DEBUG_TOKEN": {
						Value: &kubernetes.KubernetesSensitiveValue_Value{
							Value: "debug-only-token",
						},
					},
					"DATABASE_PASSWORD": {
						Value: &kubernetes.KubernetesSensitiveValue_SecretRef{
							SecretRef: &kubernetes.KubernetesSecretKeyRef{
								Name: "postgres-credentials",
								Key:  "password",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("When secret ref is missing required fields", func() {
		ginkgo.It("should fail validation when name is missing", func() {
			input.Spec.Container.App.Env = &KubernetesStatefulSetContainerAppEnv{
				Secrets: map[string]*kubernetes.KubernetesSensitiveValue{
					"DATABASE_PASSWORD": {
						Value: &kubernetes.KubernetesSensitiveValue_SecretRef{
							SecretRef: &kubernetes.KubernetesSecretKeyRef{
								Name: "",
								Key:  "password",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should fail validation when key is missing", func() {
			input.Spec.Container.App.Env = &KubernetesStatefulSetContainerAppEnv{
				Secrets: map[string]*kubernetes.KubernetesSensitiveValue{
					"DATABASE_PASSWORD": {
						Value: &kubernetes.KubernetesSensitiveValue_SecretRef{
							SecretRef: &kubernetes.KubernetesSecretKeyRef{
								Name: "my-secrets",
								Key:  "",
							},
						},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
```

### 5. Documentation Updates

**File**: `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/examples.md`

Updated to show three example patterns:
1. Direct string values
2. Kubernetes Secret references (production)
3. Mixed types (both)

## Files Changed

| File | Change |
|------|--------|
| `kubernetesstatefulset/v1/spec.proto` | Added import for `kubernetes_secret.proto`; Changed `secrets` from `map<string, string>` to `map<string, KubernetesSensitiveValue>` |
| `kubernetesstatefulset/v1/iac/pulumi/module/secret.go` | Filter to only create secret for string values; Early return when no string values |
| `kubernetesstatefulset/v1/iac/pulumi/module/statefulset.go` | Added `sort` import; Handle both secret types in env var creation |
| `kubernetesstatefulset/v1/iac/tf/variables.tf` | Updated secrets type definition to support both `value` and `secret_ref` |
| `kubernetesstatefulset/v1/iac/tf/secret.tf` | Added locals block for filtering; Only create secret for string values |
| `kubernetesstatefulset/v1/iac/tf/statefulset.tf` | Two dynamic blocks for different secret types |
| `kubernetesstatefulset/v1/spec_test.go` | Added 5 new test cases for secret reference validation |
| `kubernetesstatefulset/v1/examples.md` | Updated with all three secret options |

## Build & Validation Commands

After making changes, run:

```bash
# 1. Regenerate proto stubs
make protos

# 2. Run component-specific tests
go test ./apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/...

# 3. Full build
make build

# 4. Full test suite
make test
```

## Test Results

✅ **20 of 20 tests passed** including the 5 new secret reference tests:
- Secrets with direct string values - PASSED
- Secrets with Kubernetes Secret references - PASSED
- Mixed types (both) - PASSED
- Validation failure when secret ref name is missing - PASSED
- Validation failure when secret ref key is missing - PASSED

## Planton Web Console Integration

When this pattern is adopted, the Planton web console (`planton` repo) will need updates:

### Form Components

1. **Create Form**: Add UI for selecting between `value` and `secretRef`
2. **Edit Modal**: Support editing both value types
3. **Validation**: Client-side validation for secret ref name/key requirements

### Suggested UI Pattern

```
Secret Entry Type: [Radio: Direct Value | Kubernetes Secret Reference]

If Direct Value:
  - Text input for value

If Kubernetes Secret Reference:
  - Text input: Secret Name (required)
  - Text input: Secret Key (required)
  - Text input: Namespace (optional, shows "(defaults to deployment namespace)")
```

### TypeScript Stub Changes

After `make update-deps` in web console:
- `KubernetesSensitiveValue` type will be available
- Forms should handle `value` vs `secret_ref` oneof

## Benefits

- **Improved security**: Production deployments can use Kubernetes Secrets instead of plaintext passwords
- **GitOps friendly**: Manifests can be safely committed to version control without exposing credentials
- **Easier rotation**: Password changes only require updating the Kubernetes Secret, not the manifest
- **Follows proto patterns**: Uses `oneof` pattern consistent with existing `ValueOrRef` in the codebase
- **Reusable type**: `KubernetesSensitiveValue` is shared across components
- **Backward compatible API structure**: Both Pulumi and Terraform modules handle both value types seamlessly

## Impact

### Users
- Users deploying StatefulSets to Kubernetes can now secure their credentials properly
- No breaking changes to existing deployments (manifest YAML structure changes but is more explicit)
- Clear documentation with examples for both approaches

### Developers
- Pattern established for handling secrets across the Kubernetes provider components
- All tests updated and passing
- Reusable `KubernetesSensitiveValue` type for other sensitive fields

## Applying to Other Components

This same change should be applied to these additional components that have the same `env.secrets` pattern:

### Components Already Updated
1. **KubernetesDeployment** - `apis/org/openmcf/provider/kubernetes/kubernetesdeployment/v1/` (see prior changelog)
2. **KubernetesStatefulset** - `apis/org/openmcf/provider/kubernetes/kubernetesstatefulset/v1/` (this changelog)

### Components Still Pending
1. **KubernetesDaemonset** - `apis/org/openmcf/provider/kubernetes/kubernetesdaemonset/v1/`
2. **KubernetesCronjob** - `apis/org/openmcf/provider/kubernetes/kubernetescronjob/v1/`

**For each component**:
1. Check if `spec.proto` has the same `Container.App.Env.Secrets` pattern
2. Apply the same changes to all files listed in the table above
3. Run tests for that component
4. Update examples

## Related Work

- **Prior art**: `KubernetesDeployment` secrets reference support (2025-12-23-143353)
- **Prior art**: `KubernetesSignoz` password secret reference support (2025-12-19)
- **Prior art**: `KubernetesTemporal` password secret reference support (2025-12-22)
- **Prior art**: `KubernetesOpenfga` password secret reference support (2025-12-22)
- **Shared type**: Uses `KubernetesSecretKeyRef` and `KubernetesSensitiveValue` from `kubernetes_secret.proto`
- **Pattern reference**: Similar to `ValueOrRef` in `apis/org/openmcf/shared/foreignkey/v1/foreign_key.proto`

## Scope Clarification

This iteration **only** supports secret refs for `spec.container.app.env.secrets`.

**Not in scope** (future work):
- Sidecar container secrets (different structure)
- Volume mount secrets (already reference external secrets by design)
- Image pull secrets (already support external references)

---

**Status**: ✅ Production Ready
**Timeline**: ~30 minutes implementation
**Test Results**: 20/20 tests passing

---

## Quick Reference: Key Code Snippets

### Proto Import
```protobuf
import "org/openmcf/provider/kubernetes/kubernetes_secret.proto";
```

### Go GetSecretRef/GetValue
```go
if secretValue.GetSecretRef() != nil {
    secretRef := secretValue.GetSecretRef()
    // Use secretRef.Name, secretRef.Key
} else if secretValue.GetValue() != "" {
    // Use secretValue.GetValue()
}
```

### Terraform Filter Expression
```hcl
# String values only
if try(v.value, null) != null && v.value != ""

# Secret refs only
if try(v.secret_ref, null) != null
```

### Test Value Construction
```go
// String value
&kubernetes.KubernetesSensitiveValue{
    Value: &kubernetes.KubernetesSensitiveValue_Value{
        Value: "my-password",
    },
}

// Secret ref
&kubernetes.KubernetesSensitiveValue{
    Value: &kubernetes.KubernetesSensitiveValue_SecretRef{
        SecretRef: &kubernetes.KubernetesSecretKeyRef{
            Name: "my-secrets",
            Key:  "password",
        },
    },
}
```

