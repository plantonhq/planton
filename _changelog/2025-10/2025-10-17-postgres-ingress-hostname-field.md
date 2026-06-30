# Postgres Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Enhancement  
**Component**: PostgresKubernetes

## Summary

Refactored the Postgres Kubernetes ingress configuration from a shared `IngressSpec` (with `enabled` and `dns_domain` fields) to a custom `PostgresKubernetesIngress` message with `enabled` and `hostname` fields. This change gives users full control over the ingress hostname instead of auto-constructing it from resource ID and DNS domain. The Terraform and Pulumi modules now use the user-supplied hostname directly, eliminating internal hostname construction logic.

## Motivation

### The Problem

The previous implementation used the shared `IngressSpec` from `dev.planton.shared.kubernetes`:

```yaml
ingress:
  enabled: true
  dns_domain: "planton.live"
```

This approach had several limitations:

1. **Hostname Auto-Construction**: The system automatically constructed hostnames as `{resource-id}.{dns-domain}`, giving users no control over the exact hostname pattern.

2. **Internal vs. External Hostnames**: The implementation created both "external" and "internal" hostnames (`{resource-id}.{dns-domain}` and `{resource-id}-internal.{dns-domain}`), with the internal hostname created as a separate LoadBalancer service. However, internal hostnames were never actually used in practice.

3. **Inflexibility**: Users couldn't specify custom subdomains like `postgres.example.com` or `postgres-prod.example.com` - they were locked into the resource-id-based pattern.

4. **Module Complexity**: Both Terraform and Pulumi modules contained hostname construction logic and managed internal LoadBalancer services that were never utilized.

5. **Shared Spec Limitations**: The generic `IngressSpec` was designed for multiple resource types, forcing Postgres to inherit patterns that didn't match its specific needs.

### The Solution

Replace `IngressSpec` with a custom `PostgresKubernetesIngress` message:

```yaml
ingress:
  enabled: true
  hostname: "postgres.example.com"
```

This approach:
- ✅ Gives users complete control over hostnames
- ✅ Eliminates unused internal hostname concept
- ✅ Simplifies Terraform and Pulumi modules (removed hostname construction logic)
- ✅ Provides clearer, more intuitive API
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)

## What's New

### 1. Custom PostgresKubernetesIngress Message

**Before (Shared IngressSpec)**:
```protobuf
// From project/planton/shared/kubernetes/kubernetes.proto
message IngressSpec {
  bool enabled = 1;
  string dns_domain = 2;
}

message PostgresKubernetesSpec {
  dev.planton.shared.kubernetes.IngressSpec ingress = 2;
}
```

**After (Custom Message)**:
```protobuf
message PostgresKubernetesIngress {
  // Flag to enable or disable ingress.
  // When enabled, creates a LoadBalancer service with external-dns annotations.
  bool enabled = 1;

  // The full hostname for external access (e.g., "postgres.example.com").
  // This hostname will be configured automatically via external-dns.
  // Required when enabled is true.
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}

message PostgresKubernetesSpec {
  PostgresKubernetesIngress ingress = 2;
}
```

### 2. Updated YAML Syntax

**Before**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: production-db
spec:
  container:
    replicas: 2
    disk_size: 100Gi
  ingress:
    enabled: true
    dns_domain: example.com
  # Resulting hostname: production-db.example.com (auto-constructed)
```

**After**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: production-db
spec:
  container:
    replicas: 2
    disk_size: 100Gi
  ingress:
    enabled: true
    hostname: postgres.example.com
  # Hostname: exactly as specified by user
```

### 3. Simplified Terraform Module

**Before** (Hostname Construction):
```hcl
# locals.tf
locals {
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")

  # External hostname (null if ingress is not enabled or domain is empty)
  ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}.${local.ingress_dns_domain}" : null

  # Internal hostname (null if ingress is not enabled or domain is empty)
  ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-internal.${local.ingress_dns_domain}" : null
}
```

**After** (Direct Usage):
```hcl
# locals.tf
locals {
  # Ingress configuration
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)
}
```

**Removed**:
- 10+ lines of hostname construction logic
- Internal hostname variable (never used)
- DNS domain parsing and validation

### 4. Simplified Pulumi Module

**Before** (Hostname Construction):
```go
// locals.go
type Locals struct {
    IngressExternalHostname string
    IngressInternalHostname string  // Never used
    // ... other fields
}

func initializeLocals(ctx *pulumi.Context, stackInput *postgreskubernetesv1.PostgresKubernetesStackInput) *Locals {
    // ... other initialization

    if target.Spec.Ingress == nil ||
        !target.Spec.Ingress.Enabled ||
        target.Spec.Ingress.DnsDomain == "" {
        return locals
    }

    // Construct external hostname
    locals.IngressExternalHostname = fmt.Sprintf("%s.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)

    ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

    // Construct internal hostname (never used)
    locals.IngressInternalHostname = fmt.Sprintf("%s-internal.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)

    ctx.Export(OpInternalHostname, pulumi.String(locals.IngressInternalHostname))

    return locals
}
```

**After** (Direct Usage):
```go
// locals.go
type Locals struct {
    IngressExternalHostname string
    // IngressInternalHostname removed
    // ... other fields
}

func initializeLocals(ctx *pulumi.Context, stackInput *postgreskubernetesv1.PostgresKubernetesStackInput) *Locals {
    // ... other initialization

    if target.Spec.Ingress == nil ||
        !target.Spec.Ingress.Enabled ||
        target.Spec.Ingress.Hostname == "" {
        return locals
    }

    locals.IngressExternalHostname = target.Spec.Ingress.Hostname

    //export external hostname
    ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

    return locals
}
```

**Removed**:
- `IngressInternalHostname` field from `Locals` struct
- `OpInternalHostname` constant from outputs
- 10+ lines of hostname construction logic
- Internal hostname export (never used)

### 5. Terraform Outputs Simplified

**Before**:
```hcl
output "external_hostname" {
  description = "The external hostname for Postgres if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal hostname for Postgres if ingress is enabled."
  value       = local.ingress_internal_hostname
}
```

**After**:
```hcl
output "external_hostname" {
  description = "The external hostname for Postgres if ingress is enabled."
  value       = local.ingress_external_hostname
}

# internal_hostname output removed
```

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/postgreskubernetes/v1/spec.proto`

**Changes Made**:
1. **Updated Field Type**: Line 68 changed from `dev.planton.shared.kubernetes.IngressSpec` to `PostgresKubernetesIngress`
2. **Added Message**: New `PostgresKubernetesIngress` message with CEL validation

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Terraform Module Updates

**Files Modified**:
1. **`iac/tf/locals.tf`**: 
   - Simplified ingress variables (lines 52-53)
   - Removed hostname construction logic
   - Removed internal hostname variable

2. **`iac/tf/main.tf`**: 
   - Removed `internal_lb` resource (lines 40-71)
   - Updated `external_lb` condition

3. **`iac/tf/outputs.tf`**:
   - Removed `internal_hostname` output

4. **`iac/tf/variables.tf`**:
   - Updated ingress structure to use `enabled` and `hostname` fields

5. **`iac/tf/hack/manifest.yaml`**:
   - Updated with new ingress format

### Pulumi Module Updates

**Files Modified**:
1. **`iac/pulumi/module/locals.go`**:
   - Removed `IngressInternalHostname` field from `Locals` struct
   - Simplified ingress hostname logic
   - Removed internal hostname export

2. **`iac/pulumi/module/outputs.go`**:
   - Removed `OpInternalHostname` constant

3. **`iac/pulumi/module/main.go`**:
   - Updated ingress condition to check `Hostname` instead of `DnsDomain`

### Documentation Updates

All documentation files updated with correct ingress syntax:

1. **`v1/README.md`**: 
   - Enhanced "Ingress Configuration" section with detailed feature list

2. **`v1/iac/pulumi/README.md`**:
   - Removed internal hostname from outputs documentation

3. **`v1/iac/pulumi/examples.md`**:
   - Updated all 4 examples to use new syntax

4. **`v1/iac/tf/hack/manifest.yaml`**:
   - Updated with new field names and ingress structure

5. **`v1/api_test.go`**:
   - Updated test to use new `PostgresKubernetesIngress` message

### Stack Outputs Proto

**File**: `apis/project/planton/provider/kubernetes/workload/postgreskubernetes/v1/stack_outputs.proto`

Added deprecation comment to `internal_hostname` field (line 30) for backward compatibility.

## Examples

### Basic Ingress Configuration

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: basic-postgres
spec:
  container:
    replicas: 1
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    disk_size: 10Gi
  ingress:
    enabled: true
    hostname: postgres.example.com
```

### Production with Custom Hostname

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: prod-postgres
spec:
  container:
    replicas: 3
    resources:
      requests:
        cpu: 1000m
        memory: 4Gi
      limits:
        cpu: 4000m
        memory: 16Gi
    disk_size: 200Gi
  ingress:
    enabled: true
    hostname: postgres-prod.company.com
```

### Using with External-DNS

The `hostname` field works seamlessly with external-dns annotation:

```yaml
# Postgres manifest
spec:
  ingress:
    enabled: true
    hostname: postgres.example.com
```

This creates a LoadBalancer service with:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-external-lb
  annotations:
    external-dns.alpha.kubernetes.io/hostname: postgres.example.com
spec:
  type: LoadBalancer
  # ... service configuration
```

External-DNS then automatically:
1. Detects the annotation
2. Waits for LoadBalancer IP assignment
3. Creates DNS A record: `postgres.example.com` → `<LoadBalancer-IP>`
4. Creates TXT record for ownership tracking

## Benefits

### 1. User Control Over Hostnames

**Before**: System decides hostname pattern
```yaml
# User specifies:
ingress:
  dns_domain: "example.com"

# System creates: my-resource-id.example.com (no control)
```

**After**: User decides exact hostname
```yaml
# User specifies:
ingress:
  hostname: "postgres.example.com"

# System uses: postgres.example.com (exact match)
```

### 2. Simplified Module Implementation

**Code Reduction**:
- Terraform: 15+ lines removed (hostname construction logic + internal LB)
- Pulumi: 12+ lines removed (hostname construction and internal hostname)
- Total: ~27 lines of code eliminated

**Maintenance Benefits**:
- Fewer edge cases to handle
- No string manipulation or formatting logic
- Direct pass-through from manifest to Kubernetes

### 3. Clearer API

**Before** (Two-step mental model):
1. User provides DNS domain
2. System constructs hostname from resource ID + DNS domain

**After** (One-step mental model):
1. User provides exact hostname

### 4. Flexibility

Users can now use any hostname pattern:
- Subdomains: `postgres.analytics.example.com`
- Environments: `postgres-prod.example.com`, `postgres-staging.example.com`
- Descriptive: `database.example.com`, `primary-db.example.com`
- No pattern: Full freedom to match organizational DNS conventions

### 5. Removed Unused Features

**Internal Hostname and LoadBalancer**: Previously generated but never used in:
- LoadBalancer services (Pulumi never created it)
- Terraform created it but it was unused
- Documentation examples
- Any operational workflows

Removing it simplifies the codebase and API surface.

## Validation

### CEL Validation Rules

The new `PostgresKubernetesIngress` message includes built-in validation:

```protobuf
option (buf.validate.message).cel = {
  id: "spec.ingress.hostname.required"
  expression: "!this.enabled || size(this.hostname) > 0"
  message: "hostname is required when ingress is enabled"
};
```

**Validation Behavior**:

✅ **Valid** - Ingress disabled, no hostname:
```yaml
ingress:
  enabled: false
```

✅ **Valid** - Ingress enabled with hostname:
```yaml
ingress:
  enabled: true
  hostname: "postgres.example.com"
```

❌ **Invalid** - Ingress enabled without hostname:
```yaml
ingress:
  enabled: true
  # Error: hostname is required when ingress is enabled
```

❌ **Invalid** - Empty hostname with ingress enabled:
```yaml
ingress:
  enabled: true
  hostname: ""
  # Error: hostname is required when ingress is enabled
```

## Testing

### Test Scenarios

**Scenario 1: New Deployment with Ingress**
```bash
# Create manifest with new syntax
cat > postgres-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: test-postgres
spec:
  container:
    replicas: 1
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    disk_size: 20Gi
  ingress:
    enabled: true
    hostname: test-postgres.example.com
EOF

# Deploy
planton pulumi up --manifest postgres-test.yaml

# Verify LoadBalancer service created with correct annotation
kubectl get svc -n test-postgres ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show: test-postgres.example.com
```

**Scenario 2: Update Existing Deployment**
```bash
# Update existing manifest from old to new format
# Before: ingress.dns_domain = "example.com"
# After: ingress.hostname = "existing-postgres.example.com"

# Apply update
planton pulumi up --manifest postgres-existing.yaml

# Verify hostname annotation updated
kubectl get svc -n existing-postgres ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show new hostname
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > postgres-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: invalid-postgres
spec:
  ingress:
    enabled: true
    # Missing hostname - should fail validation
EOF

# Attempt deploy
planton pulumi up --manifest postgres-invalid.yaml
# Expected error: hostname is required when ingress is enabled
```

## Performance Impact

**No runtime performance impact**:
- Ingress configuration is applied once at deployment time
- LoadBalancer services created/updated during Pulumi apply
- No ongoing hostname construction or manipulation
- Module simplification actually reduces deployment time slightly (fewer operations)

## Security Considerations

**No security impact**:
- Hostnames are public information (used in DNS records)
- No changes to authentication, authorization, or encryption
- LoadBalancer and external-dns security model unchanged
- Validation ensures hostname cannot be empty when ingress is enabled

**Operational Security**:
- Users must ensure hostname ownership before use
- DNS domain should be under organization's control
- External-DNS credentials still required for DNS record creation

## Related Documentation

- **Postgres Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/postgreskubernetes/v1/`
- **External-DNS Integration**: See `ExternalDnsKubernetes` resource for DNS automation
- **Zalando PostgreSQL Operator**: https://github.com/zalando/postgres-operator
- **CEL Validation**: https://github.com/bufbuild/protovalidate

## Implementation Checklist

- [x] Documentation updated (examples, README, API docs)
- [x] Validation added to catch invalid configurations
- [x] Clear error messages for validation failures
- [x] Before/after comparison provided
- [x] Testing instructions included
- [x] Terraform and Pulumi modules both updated
- [x] All examples updated to new syntax

## Deployment Status

✅ **Protobuf Contract**: Updated with custom ingress message and CEL validation  
✅ **Terraform Module**: Hostname construction removed, internal LB eliminated, direct usage implemented  
✅ **Pulumi Module**: Hostname construction removed, internal hostname eliminated  
✅ **Documentation**: All examples and READMEs updated  
✅ **Validation**: CEL validation ensures hostname required when enabled  
✅ **Outputs**: Internal hostname output removed/deprecated in both Terraform and Pulumi  
✅ **Stack Outputs**: Deprecation comment added for backward compatibility

**Status**: Complete and ready for protobuf regeneration

## Future Enhancements

1. **Multiple Hostnames**: Support array of hostnames for multi-domain access
   ```yaml
   ingress:
     enabled: true
     hostnames:
       - postgres.example.com
       - database.example.com
   ```

2. **Hostname Validation**: Add regex validation for DNS compliance
   ```protobuf
   string hostname = 2 [
     (buf.validate.field).string.pattern = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"
   ];
   ```

3. **TLS Configuration**: Add optional TLS cert configuration
   ```yaml
   ingress:
     enabled: true
     hostname: postgres.example.com
     tls:
       enabled: true
       secretName: postgres-tls
   ```

4. **Internal LoadBalancer Option**: Support cloud-provider-specific internal LBs when needed
   ```yaml
   ingress:
     enabled: true
     hostname: postgres.internal.example.com
     internal: true  # Cloud-provider-specific internal LB
   ```

---

**Impact**: This change improves the Postgres Kubernetes API by providing user control over ingress hostnames, simplifying implementation, and removing unused features.

