# ClickHouse Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Breaking Change, Enhancement  
**Component**: ClickHouseKubernetes

## Summary

Refactored the ClickHouse Kubernetes ingress configuration from a shared `IngressSpec` (with `enabled` and `dns_domain` fields) to a custom `ClickHouseKubernetesIngress` message with `enabled` and `hostname` fields. This change gives users full control over the ingress hostname instead of auto-constructing it from resource ID and DNS domain. The Terraform and Pulumi modules now use the user-supplied hostname directly, eliminating internal hostname construction logic.

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

2. **Internal vs. External Hostnames**: The implementation created both "external" and "internal" hostnames (`{resource-id}.{dns-domain}` and `{resource-id}-internal.{dns-domain}`), but internal hostnames were never actually used or needed for ClickHouse.

3. **Inflexibility**: Users couldn't specify custom subdomains like `analytics.example.com` or `clickhouse-prod.example.com` - they were locked into the resource-id-based pattern.

4. **Module Complexity**: Both Terraform and Pulumi modules contained hostname construction logic that could be eliminated if users specified the full hostname directly.

5. **Shared Spec Limitations**: The generic `IngressSpec` was designed for multiple resource types, forcing ClickHouse to inherit patterns that didn't match its specific needs.

### The Solution

Replace `IngressSpec` with a custom `ClickHouseKubernetesIngress` message:

```yaml
ingress:
  enabled: true
  hostname: "clickhouse.example.com"
```

This approach:
- ✅ Gives users complete control over hostnames
- ✅ Eliminates unused internal hostname concept
- ✅ Simplifies Terraform and Pulumi modules (removed hostname construction logic)
- ✅ Provides clearer, more intuitive API
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)

## What's New

### 1. Custom ClickHouseKubernetesIngress Message

**Before (Shared IngressSpec)**:
```protobuf
// From project/planton/shared/kubernetes/kubernetes.proto
message IngressSpec {
  bool enabled = 1;
  string dns_domain = 2;
}

message ClickHouseKubernetesSpec {
  dev.planton.shared.kubernetes.IngressSpec ingress = 3;
}
```

**After (Custom Message)**:
```protobuf
message ClickHouseKubernetesIngress {
  // Flag to enable or disable ingress.
  // When enabled, creates a LoadBalancer service with external-dns annotations.
  bool enabled = 1;

  // The full hostname for external access (e.g., "clickhouse.example.com").
  // This hostname will be configured automatically via external-dns.
  // Required when enabled is true.
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}

message ClickHouseKubernetesSpec {
  ClickHouseKubernetesIngress ingress = 3;
}
```

### 2. Updated YAML Syntax

**Before**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: public-clickhouse
spec:
  ingress:
    enabled: true
    dns_domain: example.com
  # Resulting hostname: public-clickhouse.example.com (auto-constructed)
```

**After**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: public-clickhouse
spec:
  ingress:
    enabled: true
    hostname: clickhouse.example.com
  # Hostname: exactly as specified by user
```

### 3. Simplified Terraform Module

**Before** (Hostname Construction):
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
```

**After** (Direct Usage):
```hcl
# locals.tf
locals {
  ingress_is_enabled        = try(var.spec.ingress.enabled, false)
  ingress_external_hostname = try(var.spec.ingress.hostname, null)
}
```

**Removed**:
- 8 lines of hostname construction logic
- Internal hostname variable (never used)
- DNS domain parsing and validation

### 4. Simplified Pulumi Module

**Before** (Hostname Construction):
```go
// locals.go
type Locals struct {
    IngressExternalHostname     string
    IngressInternalHostname     string  // Never used
    // ... other fields
}

func initializeLocals(ctx *pulumi.Context, stackInput *clickhousekubernetesv1.ClickHouseKubernetesStackInput) *Locals {
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
    IngressExternalHostname     string
    // IngressInternalHostname removed
    // ... other fields
}

func initializeLocals(ctx *pulumi.Context, stackInput *clickhousekubernetesv1.ClickHouseKubernetesStackInput) *Locals {
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
  description = "The external hostname for ClickHouse if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal hostname for ClickHouse if ingress is enabled."
  value       = local.ingress_internal_hostname
}
```

**After**:
```hcl
output "external_hostname" {
  description = "The external hostname for ClickHouse if ingress is enabled."
  value       = local.ingress_external_hostname
}

# internal_hostname output removed
```

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/clickhousekubernetes/v1/spec.proto`

**Changes Made**:
1. **Removed Import**: No longer imports `project/planton/shared/kubernetes/kubernetes.proto`
2. **Updated Field Type**: Line 50 changed from `dev.planton.shared.kubernetes.IngressSpec` to `ClickHouseKubernetesIngress`
3. **Added Message**: New `ClickHouseKubernetesIngress` message with CEL validation (lines 346-369)

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Terraform Module Updates

**Files Modified**:
1. **`iac/tf/locals.tf`**: 
   - Simplified ingress variables (lines 52-54)
   - Removed hostname construction logic
   - Removed internal hostname variable

2. **`iac/tf/ingress.tf`**: 
   - Already correctly references `local.ingress_external_hostname` (no changes needed)

3. **`iac/tf/outputs.tf`**:
   - Removed `internal_hostname` output

4. **`iac/tf/README.md`**:
   - Updated ingress feature description
   - Updated example code with new syntax
   - Removed internal hostname from outputs table

5. **`iac/tf/examples.md`**:
   - Updated Example 6 with new ingress format

### Pulumi Module Updates

**Files Modified**:
1. **`iac/pulumi/module/locals.go`**:
   - Removed `IngressInternalHostname` field from `Locals` struct
   - Simplified ingress hostname logic
   - Removed internal hostname export

2. **`iac/pulumi/module/outputs.go`**:
   - Removed `OpInternalHostname` constant

3. **`iac/pulumi/module/ingress.go`**:
   - Already correctly references `locals.IngressExternalHostname` (no changes needed)

4. **`iac/pulumi/README.md`**:
   - Updated ingress integration description

5. **`iac/pulumi/examples.md`**:
   - Updated Examples 7 and 8 with new ingress format

### Documentation Updates

All documentation files updated with correct ingress syntax:

1. **`v1/examples.md`**: 
   - Updated "Example w/ Ingress Enabled" section
   - Changed from `isEnabled`/`dnsDomain` to `enabled`/`hostname`

2. **`v1/README.md`**:
   - Enhanced "Networking and Ingress" section with detailed feature list
   - Updated configuration examples

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for existing ClickHouseKubernetes resources with ingress enabled.

**Affected Users**: Users who have deployed ClickHouse with ingress enabled (likely a small subset given the resource's recent introduction).

### Migration Steps

#### Step 1: Identify Affected Resources

Find all ClickHouseKubernetes manifests with ingress configuration:

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
    hostname: "my-clickhouse.planton.live"
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
  name: prod-clickhouse
spec:
  ingress:
    dns_domain: "example.com"

# The old system created: prod-clickhouse.example.com
# So use:
spec:
  ingress:
    hostname: "prod-clickhouse.example.com"
```

**Option B - Choose New Hostname** (take advantage of flexibility):
```yaml
spec:
  ingress:
    hostname: "analytics.example.com"  # Any hostname you want!
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
make protos
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
planton pulumi preview --manifest clickhouse.yaml

# Apply
planton pulumi up --manifest clickhouse.yaml
```

### Automated Migration Script

For users with many manifests:

```bash
#!/bin/bash
# migrate-clickhouse-ingress.sh

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

# Find all ClickHouseKubernetes manifests
find . -name "*.yaml" -type f | while read file; do
    # Check if it's a ClickHouseKubernetes resource
    if grep -q "kind: ClickHouseKubernetes" "$file"; then
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
chmod +x migrate-clickhouse-ingress.sh
./migrate-clickhouse-ingress.sh
```

## Examples

### Basic Ingress Configuration

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: analytics-clickhouse
spec:
  clusterName: analytics-cluster
  container:
    replicas: 1
    isPersistenceEnabled: true
    diskSize: 100Gi
    resources:
      requests:
        cpu: 1000m
        memory: 4Gi
      limits:
        cpu: 4000m
        memory: 16Gi
  ingress:
    enabled: true
    hostname: clickhouse.example.com
```

### Production with Custom Hostname

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: prod-analytics
spec:
  clusterName: production-cluster
  version: "24.8"
  container:
    replicas: 1
    isPersistenceEnabled: true
    diskSize: 200Gi
    resources:
      requests:
        cpu: 2000m
        memory: 8Gi
      limits:
        cpu: 8000m
        memory: 32Gi
  ingress:
    enabled: true
    hostname: analytics-prod.company.com
```

### Clustered Deployment with Ingress

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: signoz-backend
spec:
  clusterName: cluster
  version: "24.8"
  container:
    isPersistenceEnabled: true
    diskSize: 50Gi
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        cpu: 2000m
        memory: 4Gi
  cluster:
    isEnabled: true
    shardCount: 2
    replicaCount: 2
  coordination:
    type: keeper
    keeperConfig:
      replicas: 3
  ingress:
    enabled: true
    hostname: signoz-clickhouse.example.com
```

### Using with External-DNS

The `hostname` field works seamlessly with external-dns annotation:

```yaml
# ClickHouse manifest
spec:
  ingress:
    enabled: true
    hostname: clickhouse.example.com
```

This creates a LoadBalancer service with:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-external-lb
  annotations:
    external-dns.alpha.kubernetes.io/hostname: clickhouse.example.com
spec:
  type: LoadBalancer
  # ... service configuration
```

External-DNS then automatically:
1. Detects the annotation
2. Waits for LoadBalancer IP assignment
3. Creates DNS A record: `clickhouse.example.com` → `<LoadBalancer-IP>`
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
  hostname: "analytics.example.com"

# System uses: analytics.example.com (exact match)
```

### 2. Simplified Module Implementation

**Code Reduction**:
- Terraform: 10+ lines removed (hostname construction logic)
- Pulumi: 15+ lines removed (hostname construction and internal hostname)
- Total: ~25 lines of code eliminated

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
- Subdomains: `clickhouse.analytics.example.com`
- Environments: `clickhouse-prod.example.com`, `clickhouse-staging.example.com`
- Descriptive: `analytics-db.example.com`, `observability-clickhouse.example.com`
- No pattern: Full freedom to match organizational DNS conventions

### 5. Removed Unused Features

**Internal Hostname**: Previously generated but never used in:
- LoadBalancer services
- Ingress resources
- Documentation examples
- Any operational workflows

Removing it simplifies the codebase and API surface.

## Validation

### CEL Validation Rules

The new `ClickHouseKubernetesIngress` message includes built-in validation:

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
  hostname: "clickhouse.example.com"
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
cat > clickhouse-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: test-clickhouse
spec:
  container:
    replicas: 1
    isPersistenceEnabled: true
    diskSize: 20Gi
  ingress:
    enabled: true
    hostname: test-clickhouse.example.com
EOF

# Deploy
planton pulumi up --manifest clickhouse-test.yaml

# Verify LoadBalancer service created with correct annotation
kubectl get svc -n test-clickhouse ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show: test-clickhouse.example.com
```

**Scenario 2: Update Existing Deployment**
```bash
# Update existing manifest from old to new format
# Before: ingress.dns_domain = "example.com"
# After: ingress.hostname = "existing-clickhouse.example.com"

# Apply update
planton pulumi up --manifest clickhouse-existing.yaml

# Verify hostname annotation updated
kubectl get svc -n existing-clickhouse ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show new hostname
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > clickhouse-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: ClickHouseKubernetes
metadata:
  name: invalid-clickhouse
spec:
  ingress:
    enabled: true
    # Missing hostname - should fail validation
EOF

# Attempt deploy
planton pulumi up --manifest clickhouse-invalid.yaml
# Expected error: hostname is required when ingress is enabled
```

## Performance Impact

**No runtime performance impact**:
- Ingress configuration is applied once at deployment time
- LoadBalancer services created/updated during Helm/Pulumi apply
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

- **ClickHouse Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/clickhousekubernetes/v1/`
- **External-DNS Integration**: See `ExternalDnsKubernetes` resource for DNS automation
- **Altinity ClickHouse Operator**: https://github.com/Altinity/clickhouse-operator
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
✅ **Terraform Module**: Hostname construction removed, direct usage implemented  
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
       - clickhouse.example.com
       - analytics.example.com
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
     hostname: clickhouse.example.com
     tls:
       enabled: true
       secretName: clickhouse-tls
   ```

4. **Load Balancer Type**: Support internal vs external LoadBalancer
   ```yaml
   ingress:
     enabled: true
     hostname: clickhouse.internal.example.com
     internal: true  # Internal LoadBalancer
   ```

## Support

For questions or issues with migration:
1. Review the [migration guide](#migration-guide) above
2. Use the [automated migration script](#automated-migration-script)
3. Check [examples](#examples) for reference configurations
4. Verify [validation rules](#validation) are met
5. Contact Planton support if issues persist

---

**Impact**: This change improves the ClickHouse Kubernetes API by providing user control over ingress hostnames, simplifying implementation, and removing unused features. The migration path is straightforward with clear documentation and automation tools.

