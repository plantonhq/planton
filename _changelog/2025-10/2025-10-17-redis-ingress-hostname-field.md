# Redis Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Breaking Change, Enhancement  
**Component**: RedisKubernetes

## Summary

Refactored the Redis Kubernetes ingress configuration from a shared `IngressSpec` (with `enabled` and `dns_domain` fields) to a custom `RedisKubernetesIngress` message with `enabled` and `hostname` fields. This change gives users full control over the ingress hostname instead of auto-constructing it from resource ID and DNS domain. The Terraform and Pulumi modules now use the user-supplied hostname directly, eliminating internal hostname construction logic and removing the unused internal LoadBalancer service.

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

2. **Internal vs. External Hostnames**: The implementation created both "external" and "internal" hostnames (`{resource-id}.{dns-domain}` and `{resource-id}-internal.{dns-domain}`), with separate LoadBalancer services. However, the internal hostname and LoadBalancer were never actually used in practice.

3. **Inflexibility**: Users couldn't specify custom subdomains like `cache.example.com` or `redis-prod.example.com` - they were locked into the resource-id-based pattern.

4. **Module Complexity**: Both Terraform and Pulumi modules contained hostname construction logic and managed two LoadBalancer services when only one was needed.

5. **Shared Spec Limitations**: The generic `IngressSpec` was designed for multiple resource types, forcing Redis to inherit patterns that didn't match its specific needs.

### The Solution

Replace `IngressSpec` with a custom `RedisKubernetesIngress` message:

```yaml
ingress:
  enabled: true
  hostname: "redis.example.com"
```

This approach:
- ✅ Gives users complete control over hostnames
- ✅ Eliminates unused internal hostname and LoadBalancer
- ✅ Simplifies Terraform and Pulumi modules (removed hostname construction logic)
- ✅ Provides clearer, more intuitive API
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)

## What's New

### 1. Custom RedisKubernetesIngress Message

**Before (Shared IngressSpec)**:
```protobuf
// From project/planton/shared/kubernetes/kubernetes.proto
message IngressSpec {
  bool enabled = 1;
  string dns_domain = 2;
}

message RedisKubernetesSpec {
  dev.planton.shared.kubernetes.IngressSpec ingress = 2;
}
```

**After (Custom Message)**:
```protobuf
message RedisKubernetesIngress {
  // Flag to enable or disable ingress.
  // When enabled, creates a LoadBalancer service with external-dns annotations.
  bool enabled = 1;

  // The full hostname for external access (e.g., "redis.example.com").
  // This hostname will be configured automatically via external-dns.
  // Required when enabled is true.
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}

message RedisKubernetesSpec {
  RedisKubernetesIngress ingress = 2;
}
```

### 2. Updated YAML Syntax

**Before**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: prod-redis
spec:
  ingress:
    enabled: true
    dns_domain: example.com
  # Resulting hostnames:
  # - External: prod-redis.example.com (auto-constructed)
  # - Internal: prod-redis-internal.example.com (auto-constructed, unused)
```

**After**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: prod-redis
spec:
  ingress:
    enabled: true
    hostname: redis.example.com
  # Hostname: exactly as specified by user
```

### 3. Simplified Terraform Module

**Before** (Hostname Construction + Internal LB):
```hcl
# locals.tf
locals {
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")

  # Construct external hostname
  ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}.${local.ingress_dns_domain}" : null

  # Construct internal hostname (never actually used)
  ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-internal.${local.ingress_dns_domain}" : null
}

# load_balancer_ingress.tf - Two LoadBalancer services
resource "kubernetes_service" "redis_external_lb" {
  count = var.spec.ingress.is_enabled && var.spec.ingress.dns_domain != "" ? 1 : 0
  # ... external LB configuration
}

resource "kubernetes_service" "redis_internal_lb" {
  count = var.spec.ingress.is_enabled && var.spec.ingress.dns_domain != "" ? 1 : 0
  # ... internal LB configuration with GCP-specific annotations
}
```

**After** (Direct Usage + Single LB):
```hcl
# locals.tf
locals {
  ingress_is_enabled = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)
}

# load_balancer_ingress.tf - One LoadBalancer service
resource "kubernetes_service" "redis_external_lb" {
  count = var.spec.ingress.enabled && var.spec.ingress.hostname != "" ? 1 : 0
  # ... external LB configuration
}
```

**Removed**:
- 8 lines of hostname construction logic
- Internal hostname variable (never used)
- Internal LoadBalancer service resource
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

func initializeLocals(ctx *pulumi.Context, stackInput *rediskubernetesv1.RedisKubernetesStackInput) *Locals {
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

func initializeLocals(ctx *pulumi.Context, stackInput *rediskubernetesv1.RedisKubernetesStackInput) *Locals {
    // ... other initialization

    if target.Spec.Ingress == nil ||
        !target.Spec.Ingress.Enabled ||
        target.Spec.Ingress.Hostname == "" {
        return locals
    }

    locals.IngressExternalHostname = target.Spec.Ingress.Hostname

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
  description = "The external hostname for Redis if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal hostname for Redis if ingress is enabled."
  value       = local.ingress_internal_hostname
}
```

**After**:
```hcl
output "external_hostname" {
  description = "The external hostname for Redis if ingress is enabled."
  value       = local.ingress_external_hostname
}

# internal_hostname output removed
```

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/rediskubernetes/v1/spec.proto`

**Changes Made**:
1. **Updated Field Type**: Line 39 changed from `dev.planton.shared.kubernetes.IngressSpec` to `RedisKubernetesIngress`
2. **Added Message**: New `RedisKubernetesIngress` message with CEL validation (lines 78-96)

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Terraform Module Updates

**Files Modified**:
1. **`iac/tf/locals.tf`**: 
   - Simplified ingress variables (lines 50-51)
   - Removed hostname construction logic
   - Removed internal hostname variable

2. **`iac/tf/load_balancer_ingress.tf`**: 
   - Updated external LB condition to check hostname instead of dns_domain
   - Removed entire `redis_internal_lb` resource

3. **`iac/tf/outputs.tf`**:
   - Removed `internal_hostname` output

4. **`iac/tf/variables.tf`**:
   - Updated ingress object to use `enabled` and `hostname` fields

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
   - Updated ingress condition to use new field names

### Documentation Updates

All documentation files updated with correct ingress syntax:

1. **`v1/README.md`**: 
   - Updated "Ingress Configuration" section with hostname control details

2. **`v1/examples.md`**: 
   - Updated Example 1 (basic deployment without ingress)
   - Updated Example 2 (persistent Redis with ingress enabled)

3. **`v1/iac/pulumi/examples.md`**:
   - Matched updates from root-level examples.md

4. **`v1/iac/pulumi/README.md`**:
   - Removed internal_hostname from RedisKubernetesStackOutputs section

### Test Updates

**File**: `v1/api_test.go`

Updated test input to use new `RedisKubernetesIngress` message with `Enabled` and `Hostname` fields.

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for existing RedisKubernetes resources with ingress enabled.

**Affected Users**: Users who have deployed Redis with ingress enabled (likely a small subset given the resource's usage patterns).

### Migration Steps

#### Step 1: Identify Affected Resources

Find all RedisKubernetes manifests with ingress configuration:

```bash
# Search for manifests with ingress enabled
grep -r "ingress:" -A 2 *.yaml | grep -E "(enabled|dns_domain)"
```

#### Step 2: Update Manifest Syntax

**Before Migration**:
```yaml
spec:
  ingress:
    enabled: true
    dns_domain: "planton.live"
    # System creates: {resource-id}.planton.live
```

**After Migration**:
```yaml
spec:
  ingress:
    enabled: true
    hostname: "my-redis.planton.live"
    # User controls exact hostname
```

**Field Name Changes**:
| Old Field     | New Field  | Notes |
|---------------|------------|-------|
| `enabled`     | `enabled`  | ✅ No change |
| `dns_domain`  | `hostname` | ⚠️ Changed: now specify full hostname |

#### Step 3: Determine Your Hostname

The old system constructed hostnames as `{resource-id}.{dns-domain}`. You need to replicate this or choose a new hostname:

**Option A - Keep Existing Hostname** (recommended for minimal disruption):
```yaml
# If your manifest had:
metadata:
  name: prod-redis
spec:
  ingress:
    dns_domain: "example.com"

# The old system created: prod-redis.example.com
# So use:
spec:
  ingress:
    hostname: "prod-redis.example.com"
```

**Option B - Choose New Hostname** (take advantage of flexibility):
```yaml
spec:
  ingress:
    hostname: "cache.example.com"  # Any hostname you want!
```

#### Step 4: Update CLI and Regenerate Code

```bash
# Update CLI
brew update && brew upgrade planton

# Or fresh install
brew install plantonhq/tap/planton

# Verify version
planton version

# For developers: regenerate protobuf stubs
cd apis
make
```

#### Step 5: Update DNS Records (if changing hostname)

If you chose a different hostname than the auto-constructed one:

1. **Before applying**: Note the current hostname
2. **Update manifest**: Apply new configuration
3. **Update external-dns**: The LoadBalancer service will get new annotations
4. **DNS propagation**: External-DNS will create the new record and remove the old one
5. **Verify**: Test the new hostname works before removing old DNS entries manually (if needed)

**Note**: If you keep the same hostname, DNS records won't change.

#### Step 6: Apply Changes

```bash
# Preview changes
planton pulumi preview --manifest redis.yaml

# Apply
planton pulumi up --manifest redis.yaml
```

### Automated Migration Script

For users with many manifests:

```bash
#!/bin/bash
# migrate-redis-ingress.sh

# Function to extract resource ID from metadata.name or metadata.id
get_resource_id() {
    local file=$1
    # Try to extract id first, fallback to name
    local id=$(yq eval '.metadata.id // .metadata.name' "$file")
    echo "$id"
}

# Function to update a single file
migrate_file() {
    local file=$1
    echo "Processing $file..."
    
    # Check if file has ingress configuration
    if ! grep -q "dns_domain:" "$file"; then
        echo "  No ingress configuration found, skipping"
        return
    fi
    
    # Extract resource ID and DNS domain
    local resource_id=$(get_resource_id "$file")
    local dns_domain=$(yq eval '.spec.ingress.dns_domain' "$file")
    
    if [[ "$dns_domain" == "null" ]]; then
        echo "  No dns_domain found, skipping"
        return
    fi
    
    # Construct hostname
    local hostname="${resource_id}.${dns_domain}"
    
    echo "  Resource ID: $resource_id"
    echo "  DNS Domain: $dns_domain"
    echo "  New Hostname: $hostname"
    
    # Replace dns_domain with hostname
    yq eval -i ".spec.ingress.hostname = \"$hostname\" | del(.spec.ingress.dns_domain)" "$file"
    
    echo "  ✅ Migrated successfully"
}

# Find all RedisKubernetes manifests
find . -name "*.yaml" -type f | while read file; do
    # Check if it's a RedisKubernetes resource
    if grep -q "kind: RedisKubernetes" "$file"; then
        migrate_file "$file"
    fi
done

echo ""
echo "✅ Migration complete!"
echo ""
echo "Next steps:"
echo "1. Review the changes with: git diff"
echo "2. Test with: planton pulumi preview --manifest <file>"
echo "3. Apply with: planton pulumi up --manifest <file>"
```

**Usage**:
```bash
chmod +x migrate-redis-ingress.sh
./migrate-redis-ingress.sh
```

## Examples

### Basic Redis Configuration Without Ingress

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: internal-cache
spec:
  container:
    replicas: 1
    resources:
      requests:
        cpu: 50m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
    is_persistence_enabled: false
```

### Production with Ingress and Persistence

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: prod-redis
spec:
  container:
    replicas: 3
    resources:
      requests:
        cpu: 100m
        memory: 512Mi
      limits:
        cpu: 2000m
        memory: 2Gi
    is_persistence_enabled: true
    disk_size: 20Gi
  ingress:
    enabled: true
    hostname: redis-prod.company.com
```

### Using with External-DNS

The `hostname` field works seamlessly with external-dns annotation:

```yaml
# Redis manifest
spec:
  ingress:
    enabled: true
    hostname: cache.example.com
```

This creates a LoadBalancer service with:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-external-lb
  annotations:
    external-dns.alpha.kubernetes.io/hostname: cache.example.com
spec:
  type: LoadBalancer
  # ... service configuration
```

External-DNS then automatically:
1. Detects the annotation
2. Waits for LoadBalancer IP assignment
3. Creates DNS A record: `cache.example.com` → `<LoadBalancer-IP>`
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
  hostname: "cache.example.com"

# System uses: cache.example.com (exact match)
```

### 2. Simplified Module Implementation

**Code Reduction**:
- Terraform: 15+ lines removed (hostname construction + internal LB)
- Pulumi: 12+ lines removed (hostname construction and internal hostname)
- Total: ~27 lines of code eliminated

**Maintenance Benefits**:
- Fewer edge cases to handle
- No string manipulation or formatting logic
- Direct pass-through from manifest to Kubernetes
- Single LoadBalancer service instead of two

### 3. Clearer API

**Before** (Two-step mental model):
1. User provides DNS domain
2. System constructs external and internal hostnames from resource ID + DNS domain

**After** (One-step mental model):
1. User provides exact hostname

### 4. Flexibility

Users can now use any hostname pattern:
- Subdomains: `cache.prod.example.com`
- Environments: `redis-prod.example.com`, `redis-staging.example.com`
- Descriptive: `session-cache.example.com`, `rate-limiter.example.com`
- No pattern: Full freedom to match organizational DNS conventions

### 5. Removed Unused Features

**Internal Hostname and LoadBalancer**: Previously generated but never used in:
- Production deployments
- Documentation examples
- Operational workflows
- Any Redis client connections

Removing them simplifies the codebase and API surface while eliminating confusion.

## Validation

### CEL Validation Rules

The new `RedisKubernetesIngress` message includes built-in validation:

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
  hostname: "redis.example.com"
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
cat > redis-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: test-redis
spec:
  container:
    replicas: 1
    is_persistence_enabled: true
    disk_size: 10Gi
  ingress:
    enabled: true
    hostname: test-redis.example.com
EOF

# Deploy
planton pulumi up --manifest redis-test.yaml

# Verify LoadBalancer service created with correct annotation
kubectl get svc -n test-redis ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show: test-redis.example.com
```

**Scenario 2: Update Existing Deployment**
```bash
# Update existing manifest from old to new format
# Before: ingress.dns_domain = "example.com"
# After: ingress.hostname = "existing-redis.example.com"

# Apply update
planton pulumi up --manifest redis-existing.yaml

# Verify hostname annotation updated
kubectl get svc -n existing-redis ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show new hostname
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > redis-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: invalid-redis
spec:
  ingress:
    enabled: true
    # Missing hostname - should fail validation
EOF

# Attempt deploy
planton pulumi up --manifest redis-invalid.yaml
# Expected error: hostname is required when ingress is enabled
```

## Performance Impact

**No runtime performance impact**:
- Ingress configuration is applied once at deployment time
- LoadBalancer services created/updated during Helm/Pulumi apply
- No ongoing hostname construction or manipulation
- Module simplification actually reduces deployment time slightly (fewer operations, one less LoadBalancer)

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

- **Redis Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/rediskubernetes/v1/`
- **External-DNS Integration**: See `ExternalDnsKubernetes` resource for DNS automation
- **Bitnami Redis Helm Chart**: https://github.com/bitnami/charts/tree/main/bitnami/redis
- **CEL Validation**: https://github.com/bufbuild/protovalidate

## Breaking Change Checklist

- [x] Migration guide provided with step-by-step instructions
- [x] Automated migration script included
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
✅ **Migration Script**: Automated migration script provided  
✅ **Validation**: CEL validation ensures hostname required when enabled  
✅ **Outputs**: Internal hostname output removed from both Terraform and Pulumi

**Ready for**: Protobuf regeneration and user migration

## Future Enhancements

1. **Multiple Hostnames**: Support array of hostnames for multi-domain access
   ```yaml
   ingress:
     enabled: true
     hostnames:
       - redis.example.com
       - cache.example.com
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
     hostname: redis.example.com
     tls:
       enabled: true
       secretName: redis-tls
   ```

4. **Internal LoadBalancer Option**: Support cloud-provider-specific internal LBs when needed
   ```yaml
   ingress:
     enabled: true
     hostname: redis.internal.example.com
     internal: true  # Cloud-provider-specific internal LB
   ```

## Support

For questions or issues with migration:
1. Review the [migration guide](#migration-guide) above
2. Use the [automated migration script](#automated-migration-script)
3. Check [examples](#examples) for reference configurations
4. Verify [validation rules](#validation) are met
5. Contact Planton support if issues persist

---

**Impact**: This change improves the Redis Kubernetes API by providing user control over ingress hostnames, simplifying implementation, removing unused features (internal hostname and LoadBalancer), and streamlining the deployment model. The migration path is straightforward with clear documentation and automation tools.

