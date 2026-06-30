# MongoDB Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Breaking Change, Enhancement  
**Component**: MongodbKubernetes

## Summary

Refactored the MongoDB Kubernetes ingress configuration from a shared `IngressSpec` (with `enabled` and `dns_domain` fields) to a custom `MongodbKubernetesIngress` message with `enabled` and `hostname` fields. This change gives users full control over the ingress hostname instead of auto-constructing it from resource ID and DNS domain. The Terraform and Pulumi modules now use the user-supplied hostname directly, eliminating internal hostname construction logic. Additionally, renamed `is_persistence_enabled` to `persistence_enabled` for consistency across the API.

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

2. **Internal vs. External Hostnames**: The implementation created both "external" and "internal" hostnames (`{resource-id}.{dns-domain}` and `{resource-id}-internal.{dns-domain}`), but internal hostnames were never actually used or needed for MongoDB.

3. **Inflexibility**: Users couldn't specify custom subdomains like `mongodb.example.com` or `mongo-prod.example.com` - they were locked into the resource-id-based pattern.

4. **Module Complexity**: Both Terraform and Pulumi modules contained hostname construction logic that could be eliminated if users specified the full hostname directly.

5. **Shared Spec Limitations**: The generic `IngressSpec` was designed for multiple resource types, forcing MongoDB to inherit patterns that didn't match its specific needs.

6. **Inconsistent Field Naming**: The `is_persistence_enabled` field didn't follow modern Go/protobuf naming conventions.

### The Solution

Replace `IngressSpec` with a custom `MongodbKubernetesIngress` message:

```yaml
ingress:
  enabled: true
  hostname: "mongodb.example.com"
```

This approach:
- ✅ Gives users complete control over hostnames
- ✅ Eliminates unused internal hostname concept
- ✅ Simplifies Terraform and Pulumi modules (removed hostname construction logic)
- ✅ Provides clearer, more intuitive API
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)
- ✅ Improves field naming consistency (`persistence_enabled` instead of `is_persistence_enabled`)

## What's New

### 1. Custom MongodbKubernetesIngress Message

**Before (Shared IngressSpec)**:
```protobuf
// From project/planton/shared/kubernetes/kubernetes.proto
message IngressSpec {
  bool enabled = 1;
  string dns_domain = 2;
}

message MongodbKubernetesSpec {
  dev.planton.shared.kubernetes.IngressSpec ingress = 2;
}
```

**After (Custom Message)**:
```protobuf
message MongodbKubernetesIngress {
  // Flag to enable or disable ingress.
  // When enabled, creates a LoadBalancer service with external-dns annotations.
  bool enabled = 1;

  // The full hostname for external access (e.g., "mongodb.example.com").
  // This hostname will be configured automatically via external-dns.
  // Required when enabled is true.
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}

message MongodbKubernetesSpec {
  MongodbKubernetesIngress ingress = 2;
}
```

### 2. Updated YAML Syntax

**Before**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: MongodbKubernetes
metadata:
  name: prod-mongodb
spec:
  container:
    replicas: 1
    isPersistenceEnabled: true
    diskSize: 10Gi
  ingress:
    enabled: true
    dns_domain: example.com
  # Resulting hostname: prod-mongodb.example.com (auto-constructed)
```

**After**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: MongodbKubernetes
metadata:
  name: prod-mongodb
spec:
  container:
    replicas: 1
    persistenceEnabled: true
    diskSize: 10Gi
  ingress:
    enabled: true
    hostname: mongodb.example.com
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

func initializeLocals(ctx *pulumi.Context, stackInput *mongodbkubernetesv1.MongodbKubernetesStackInput) *Locals {
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

func initializeLocals(ctx *pulumi.Context, stackInput *mongodbkubernetesv1.MongodbKubernetesStackInput) *Locals {
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
  description = "The external hostname for MongoDB if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal hostname for MongoDB if ingress is enabled."
  value       = local.ingress_internal_hostname
}
```

**After**:
```hcl
output "external_hostname" {
  description = "The external hostname for MongoDB if ingress is enabled."
  value       = local.ingress_external_hostname
}

# internal_hostname output removed
```

### 6. Consistent Field Naming

**Before**:
```protobuf
message MongodbKubernetesContainer {
  bool is_persistence_enabled = 3;
}
```

**After**:
```protobuf
message MongodbKubernetesContainer {
  bool persistence_enabled = 3;
}
```

This aligns with modern Go/protobuf conventions and matches patterns used across other Planton resources.

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/mongodbkubernetes/v1/spec.proto`

**Changes Made**:
1. **Added Custom Message**: New `MongodbKubernetesIngress` message with CEL validation
2. **Updated Field Type**: Line 41 changed from `dev.planton.shared.kubernetes.IngressSpec` to `MongodbKubernetesIngress`
3. **Renamed Persistence Field**: `is_persistence_enabled` → `persistence_enabled` (line 89)
4. **Updated Default Values**: Changed default container options to use `persistence_enabled`
5. **Updated Validation**: CEL expression now references `persistence_enabled`

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Terraform Module Updates

**Files Modified**:
1. **`iac/tf/locals.tf`**: 
   - Simplified ingress variables (lines 46-48)
   - Removed hostname construction logic
   - Removed internal hostname variable

2. **`iac/tf/outputs.tf`**:
   - Removed `internal_hostname` output

3. **`iac/tf/variables.tf`**:
   - Renamed `is_persistence_enabled` to `persistence_enabled`
   - Updated ingress structure to use `enabled` and `hostname` fields

### Pulumi Module Updates

**Files Modified**:
1. **`iac/pulumi/module/locals.go`**:
   - Removed `IngressInternalHostname` field from `Locals` struct
   - Simplified ingress hostname logic
   - Removed internal hostname export

2. **`iac/pulumi/module/outputs.go`**:
   - Removed `OpInternalHostname` constant

3. **`iac/pulumi/module/mongodb.go`**:
   - Updated field reference from `IsPersistenceEnabled` to `PersistenceEnabled`

### Documentation Updates

All documentation files updated with correct ingress and field naming syntax:

1. **`v1/examples.md`**: 
   - Updated all persistence field references
   - Updated ingress example with new format

2. **`v1/iac/pulumi/examples.md`**:
   - Updated all persistence field references
   - Updated ingress examples

3. **`v1/iac/tf/hack/manifest.yaml`**:
   - Updated with new field names and ingress structure

4. **`v1/README.md`**:
   - Updated descriptions to reflect new field names

5. **`v1/api_test.go`**:
   - Updated test to use new field names and ingress structure

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for existing MongodbKubernetes resources with ingress enabled or using the old persistence field name.

**Affected Users**: Users who have deployed MongoDB with ingress enabled or references to `isPersistenceEnabled`.

### Migration Steps

#### Step 1: Identify Affected Resources

Find all MongodbKubernetes manifests with ingress configuration:

```bash
# Search for manifests with ingress enabled
grep -r "ingress:" -A 2 *.yaml | grep -E "(enabled|dns_domain)"
```

#### Step 2: Update Manifest Syntax

**Before Migration**:
```yaml
spec:
  container:
    replicas: 1
    isPersistenceEnabled: true
    diskSize: 10Gi
  ingress:
    enabled: true
    dns_domain: "planton.live"
    # System creates: {resource-id}.planton.live
```

**After Migration**:
```yaml
spec:
  container:
    replicas: 1
    persistenceEnabled: true
    diskSize: 10Gi
  ingress:
    enabled: true
    hostname: "my-mongodb.planton.live"
    # User controls exact hostname
```

**Field Name Changes**:
| Old Field              | New Field           | Notes |
|------------------------|---------------------|-------|
| `enabled`              | `enabled`           | ✅ No change |
| `dns_domain`           | `hostname`          | ⚠️ Changed: now specify full hostname |
| `isPersistenceEnabled` | `persistenceEnabled`| ⚠️ Changed: naming consistency |

#### Step 3: Determine Your Hostname

The old system constructed hostnames as `{resource-id}.{dns-domain}`. You need to replicate this or choose a new hostname:

**Option A - Keep Existing Hostname** (recommended for minimal disruption):
```yaml
# If your manifest had:
metadata:
  name: prod-mongodb
spec:
  ingress:
    dns_domain: "example.com"

# The old system created: prod-mongodb.example.com
# So use:
spec:
  ingress:
    hostname: "prod-mongodb.example.com"
```

**Option B - Choose New Hostname** (take advantage of flexibility):
```yaml
spec:
  ingress:
    hostname: "mongodb.example.com"  # Any hostname you want!
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
planton pulumi preview --manifest mongodb.yaml

# Apply
planton pulumi up --manifest mongodb.yaml
```

### Automated Migration Script

For users with many manifests:

```bash
#!/bin/bash
# migrate-mongodb-ingress.sh

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
    local has_ingress=$(grep -q "dns_domain:" "$file" && echo "yes" || echo "no")
    
    # Extract resource ID and DNS domain
    local resource_id=$(get_resource_id "$file")
    local dns_domain=""
    
    if [[ "$has_ingress" == "yes" ]]; then
        dns_domain=$(yq eval '.spec.ingress.dns_domain' "$file")
        
        if [[ "$dns_domain" != "null" ]]; then
            # Construct hostname
            local hostname="${resource_id}.${dns_domain}"
            
            echo "  Resource ID: $resource_id"
            echo "  DNS Domain: $dns_domain"
            echo "  New Hostname: $hostname"
            
            # Replace dns_domain with hostname
            yq eval -i ".spec.ingress.hostname = \"$hostname\" | del(.spec.ingress.dns_domain)" "$file"
        fi
    fi
    
    # Update persistence field name (always do this)
    yq eval -i '.spec.container.persistenceEnabled = .spec.container.isPersistenceEnabled | del(.spec.container.isPersistenceEnabled)' "$file"
    
    echo "  ✅ Migrated successfully"
}

# Find all MongodbKubernetes manifests
find . -name "*.yaml" -type f | while read file; do
    # Check if it's a MongodbKubernetes resource
    if grep -q "kind: MongodbKubernetes" "$file"; then
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
chmod +x migrate-mongodb-ingress.sh
./migrate-mongodb-ingress.sh
```

## Examples

### Basic Ingress Configuration

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: MongodbKubernetes
metadata:
  name: basic-mongodb
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  container:
    replicas: 1
    persistenceEnabled: true
    diskSize: 10Gi
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 1000m
        memory: 1Gi
  ingress:
    enabled: true
    hostname: mongodb.example.com
```

### Production with Custom Hostname

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: MongodbKubernetes
metadata:
  name: prod-mongodb
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  container:
    replicas: 3
    persistenceEnabled: true
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
    hostname: mongodb-prod.company.com
```

### Using with External-DNS

The `hostname` field works seamlessly with external-dns annotation:

```yaml
# MongoDB manifest
spec:
  ingress:
    enabled: true
    hostname: mongodb.example.com
```

This creates a LoadBalancer service with:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-external-lb
  annotations:
    external-dns.alpha.kubernetes.io/hostname: mongodb.example.com
spec:
  type: LoadBalancer
  # ... service configuration
```

External-DNS then automatically:
1. Detects the annotation
2. Waits for LoadBalancer IP assignment
3. Creates DNS A record: `mongodb.example.com` → `<LoadBalancer-IP>`
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
  hostname: "mongodb.example.com"

# System uses: mongodb.example.com (exact match)
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
- Subdomains: `mongodb.analytics.example.com`
- Environments: `mongodb-prod.example.com`, `mongodb-staging.example.com`
- Descriptive: `analytics-db.example.com`, `mongo-primary.example.com`
- No pattern: Full freedom to match organizational DNS conventions

### 5. Removed Unused Features

**Internal Hostname**: Previously generated but never used in:
- LoadBalancer services
- Documentation examples
- Any operational workflows

Removing it simplifies the codebase and API surface.

### 6. Consistent Naming

Renamed `is_persistence_enabled` to `persistence_enabled`:
- More idiomatic
- Matches modern Go/protobuf conventions
- Cleaner generated code in all languages
- Consistent with other Planton resources

## Validation

### CEL Validation Rules

The new `MongodbKubernetesIngress` message includes built-in validation:

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
  hostname: "mongodb.example.com"
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
cat > mongodb-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: MongodbKubernetes
metadata:
  name: test-mongodb
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  container:
    replicas: 1
    persistenceEnabled: true
    diskSize: 20Gi
  ingress:
    enabled: true
    hostname: test-mongodb.example.com
EOF

# Deploy
planton pulumi up --manifest mongodb-test.yaml

# Verify LoadBalancer service created with correct annotation
kubectl get svc -n test-mongodb ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show: test-mongodb.example.com
```

**Scenario 2: Update Existing Deployment**
```bash
# Update existing manifest from old to new format
# Before: ingress.dns_domain = "example.com"
# After: ingress.hostname = "existing-mongodb.example.com"

# Apply update
planton pulumi up --manifest mongodb-existing.yaml

# Verify hostname annotation updated
kubectl get svc -n existing-mongodb ingress-external-lb -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show new hostname
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > mongodb-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: MongodbKubernetes
metadata:
  name: invalid-mongodb
spec:
  container:
    replicas: 1
  ingress:
    enabled: true
    # Missing hostname - should fail validation
EOF

# Attempt deploy
planton pulumi up --manifest mongodb-invalid.yaml
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

- **MongoDB Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/mongodbkubernetes/v1/`
- **External-DNS Integration**: See `ExternalDnsKubernetes` resource for DNS automation
- **Percona MongoDB Operator**: https://docs.percona.com/percona-operator-for-mongodb/
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
- [x] Field naming consistency improved

## Deployment Status

✅ **Protobuf Contract**: Updated with custom ingress message and CEL validation  
✅ **Terraform Module**: Hostname construction removed, direct usage implemented  
✅ **Pulumi Module**: Hostname construction removed, internal hostname eliminated  
✅ **Documentation**: All examples and READMEs updated  
✅ **Migration Script**: Automated migration script provided  
✅ **Validation**: CEL validation ensures hostname required when enabled  
✅ **Outputs**: Internal hostname output removed from both Terraform and Pulumi  
✅ **Field Naming**: Consistent `persistence_enabled` naming throughout

**Ready for**: Protobuf regeneration and user migration

## Future Enhancements

1. **Multiple Hostnames**: Support array of hostnames for multi-domain access
   ```yaml
   ingress:
     enabled: true
     hostnames:
       - mongodb.example.com
       - mongo.example.com
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
     hostname: mongodb.example.com
     tls:
       enabled: true
       secretName: mongodb-tls
   ```

4. **Load Balancer Type**: Support internal vs external LoadBalancer
   ```yaml
   ingress:
     enabled: true
     hostname: mongodb.internal.example.com
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

**Impact**: This change improves the MongoDB Kubernetes API by providing user control over ingress hostnames, simplifying implementation, removing unused features, and improving field naming consistency. The migration path is straightforward with clear documentation and automation tools.

