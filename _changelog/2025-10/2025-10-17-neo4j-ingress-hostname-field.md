# Neo4j Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Breaking Change, Enhancement  
**Component**: Neo4jKubernetes

## Summary

Refactored the Neo4j Kubernetes ingress configuration from a shared `IngressSpec` (with `enabled` and `dns_domain` fields) to a custom `Neo4jKubernetesIngress` message with `enabled` and `hostname` fields. This change gives users full control over the ingress hostname instead of auto-constructing it from resource ID and DNS domain. The Pulumi module now uses the user-supplied hostname directly, eliminating internal hostname construction logic. Additionally, renamed `is_persistence_enabled` to `persistence_enabled` for consistency.

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

2. **Internal vs. External Hostnames**: The implementation created both "external" and "internal" hostnames (`{resource-id}.{dns-domain}` and `{resource-id}-internal.{dns-domain}`), but internal hostnames were never actually used or needed for Neo4j.

3. **Inflexibility**: Users couldn't specify custom subdomains like `graph.example.com` or `neo4j-prod.example.com` - they were locked into the resource-id-based pattern.

4. **Module Complexity**: The Pulumi module contained hostname construction logic that could be eliminated if users specified the full hostname directly.

5. **Shared Spec Limitations**: The generic `IngressSpec` was designed for multiple resource types, forcing Neo4j to inherit patterns that didn't match its specific needs.

6. **Inconsistent Naming**: The field `is_persistence_enabled` didn't match modern Go/protobuf conventions.

### The Solution

Replace `IngressSpec` with a custom `Neo4jKubernetesIngress` message:

```yaml
ingress:
  enabled: true
  hostname: "neo4j.example.com"
```

This approach:
- ✅ Gives users complete control over hostnames
- ✅ Eliminates unused internal hostname concept
- ✅ Simplifies Pulumi module (removed hostname construction logic)
- ✅ Provides clearer, more intuitive API
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)
- ✅ Consistent naming convention (`persistence_enabled` instead of `is_persistence_enabled`)

## What's New

### 1. Custom Neo4jKubernetesIngress Message

**Before (Shared IngressSpec)**:
```protobuf
// From project/planton/shared/kubernetes/kubernetes.proto
message IngressSpec {
  bool enabled = 1;
  string dns_domain = 2;
}

message Neo4jKubernetesSpec {
  dev.planton.shared.kubernetes.IngressSpec ingress = 4;
}
```

**After (Custom Message)**:
```protobuf
message Neo4jKubernetesIngress {
  // Flag to enable or disable ingress.
  // When enabled, creates a LoadBalancer service with external-dns annotations.
  bool enabled = 1;

  // The full hostname for external access (e.g., "neo4j.example.com").
  // This hostname will be configured automatically via external-dns.
  // Required when enabled is true.
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}

message Neo4jKubernetesSpec {
  Neo4jKubernetesIngress ingress = 4;
}
```

### 2. Updated YAML Syntax

**Before**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: public-neo4j
spec:
  container:
    isPersistenceEnabled: true
  ingress:
    enabled: true
    dns_domain: example.com
  # Resulting hostname: public-neo4j.example.com (auto-constructed)
```

**After**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: public-neo4j
spec:
  container:
    persistenceEnabled: true
  ingress:
    enabled: true
    hostname: neo4j.example.com
  # Hostname: exactly as specified by user
```

### 3. Simplified Pulumi Module

**Before** (Hostname Construction):
```go
// locals.go
type Locals struct {
    IngressExternalHostname string
    IngressInternalHostname string  // Never used
    // ... other fields
}

func initializeLocals(ctx *pulumi.Context, stackInput *neo4jkubernetesv1.Neo4JKubernetesStackInput) *Locals {
    // ... other initialization

    if target.Spec.Ingress != nil &&
        target.Spec.Ingress.Enabled &&
        target.Spec.Ingress.DnsDomain != "" {
        locals.IngressExternalHostname = fmt.Sprintf("%s.%s", locals.Namespace,
            target.Spec.Ingress.DnsDomain)
        locals.IngressInternalHostname = fmt.Sprintf("%s-internal.%s", locals.Namespace,
            target.Spec.Ingress.DnsDomain)

        ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))
        ctx.Export(OpInternalHostname, pulumi.String(locals.IngressInternalHostname))
    }
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

func initializeLocals(ctx *pulumi.Context, stackInput *neo4jkubernetesv1.Neo4JKubernetesStackInput) *Locals {
    // ... other initialization

    if target.Spec.Ingress != nil &&
        target.Spec.Ingress.Enabled &&
        target.Spec.Ingress.Hostname != "" {
        locals.IngressExternalHostname = target.Spec.Ingress.Hostname
        ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))
    }
}
```

**Removed**:
- `IngressInternalHostname` field from `Locals` struct
- `OpInternalHostname` constant from outputs
- 10+ lines of hostname construction logic
- Internal hostname export (never used)

### 4. Pulumi Outputs Simplified

**Before**:
```go
const (
    OpExternalHostname = "external_hostname"
    OpInternalHostname = "internal_hostname"
    // ... other outputs
)
```

**After**:
```go
const (
    OpExternalHostname = "external_hostname"
    // internal_hostname output removed
    // ... other outputs
)
```

### 5. Renamed Field for Consistency

**Before**:
```protobuf
message Neo4jKubernetesContainer {
  bool is_persistence_enabled = 2;
}
```

**After**:
```protobuf
message Neo4jKubernetesContainer {
  bool persistence_enabled = 2;
}
```

**YAML Syntax**:
```yaml
# Before
container:
  isPersistenceEnabled: true

# After  
container:
  persistenceEnabled: true
```

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/neo4jkubernetes/v1/spec.proto`

**Changes Made**:
1. **Updated Field Type**: Line 36 changed from `dev.planton.shared.kubernetes.IngressSpec` to `Neo4jKubernetesIngress`
2. **Added Message**: New `Neo4jKubernetesIngress` message with CEL validation (lines 66-82)
3. **Renamed Field**: Changed `is_persistence_enabled` to `persistence_enabled` in `Neo4jKubernetesContainer` (line 48)

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Pulumi Module Updates

**Files Modified**:

1. **`iac/pulumi/module/locals.go`**:
   - Removed `IngressInternalHostname` field from `Locals` struct
   - Simplified ingress hostname logic to use hostname directly
   - Removed internal hostname export

2. **`iac/pulumi/module/helm_chart.go`**:
   - Updated ingress enabled check to use `Hostname` instead of `DnsDomain`

3. **`iac/pulumi/module/outputs.go`**:
   - Removed `OpInternalHostname` constant

### Terraform Module Updates

**Files Modified**:

1. **`iac/tf/variables.tf`**:
   - Updated ingress object structure from `is_enabled`/`dns_domain` to `enabled`/`hostname`

### Documentation Updates

All documentation files updated with correct ingress syntax:

1. **`v1/examples.md`**: 
   - Updated "Example w/ Ingress Enabled" section with new `enabled`/`hostname` format

2. **`v1/iac/pulumi/examples.md`**:
   - Updated "Example w/ Ingress Enabled" section with new format

3. **`v1/hack/manifest.yaml`**:
   - Changed `isPersistenceEnabled` to `persistenceEnabled`
   - Changed `isEnabled` to `enabled`

4. **`v1/api_test.go`**:
   - Updated test input to use `Neo4jKubernetesIngress` with `Enabled` and `Hostname` fields

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for existing Neo4jKubernetes resources with ingress enabled.

**Affected Users**: Users who have deployed Neo4j with ingress enabled (likely a small subset given the resource's recent introduction).

### Migration Steps

#### Step 1: Identify Affected Resources

Find all Neo4jKubernetes manifests with ingress configuration:

```bash
# Search for manifests with ingress enabled
grep -r "ingress:" -A 2 *.yaml | grep -E "(enabled|dns_domain)"
```

#### Step 2: Update Manifest Syntax

**Before Migration**:
```yaml
spec:
  container:
    isPersistenceEnabled: true
  ingress:
    enabled: true
    dns_domain: "planton.live"
    # System creates: {resource-id}.planton.live
```

**After Migration**:
```yaml
spec:
  container:
    persistenceEnabled: true
  ingress:
    enabled: true
    hostname: "my-neo4j.planton.live"
    # User controls exact hostname
```

**Field Name Changes**:
| Old Field                      | New Field                 | Notes |
|--------------------------------|---------------------------|-------|
| `container.isPersistenceEnabled` | `container.persistenceEnabled` | ✅ More idiomatic |
| `ingress.enabled`              | `ingress.enabled`         | ✅ No change |
| `ingress.dns_domain`           | `ingress.hostname`        | ⚠️ Changed: now specify full hostname |

#### Step 3: Determine Your Hostname

The old system constructed hostnames as `{resource-id}.{dns-domain}`. You need to replicate this or choose a new hostname:

**Option A - Keep Existing Hostname** (recommended for minimal disruption):
```yaml
# If your manifest had:
metadata:
  name: prod-neo4j
spec:
  ingress:
    dns_domain: "example.com"

# The old system created: prod-neo4j.example.com
# So use:
spec:
  ingress:
    hostname: "prod-neo4j.example.com"
```

**Option B - Choose New Hostname** (take advantage of flexibility):
```yaml
spec:
  ingress:
    hostname: "graph.example.com"  # Any hostname you want!
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
planton pulumi preview --manifest neo4j.yaml

# Apply
planton pulumi up --manifest neo4j.yaml
```

### Automated Migration Script

For users with many manifests:

```bash
#!/bin/bash
# migrate-neo4j-ingress.sh

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
        echo "  No ingress configuration found, updating field names only"
        # Update field names even without ingress
        yq eval -i '
          .spec.container.persistenceEnabled = .spec.container.isPersistenceEnabled |
          del(.spec.container.isPersistenceEnabled) |
          .spec.ingress.enabled = .spec.ingress.isEnabled |
          del(.spec.ingress.isEnabled)
        ' "$file"
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
    
    # Replace dns_domain with hostname and update field names
    yq eval -i "
      .spec.container.persistenceEnabled = .spec.container.isPersistenceEnabled |
      del(.spec.container.isPersistenceEnabled) |
      .spec.ingress.hostname = \"$hostname\" |
      del(.spec.ingress.dns_domain) |
      .spec.ingress.enabled = .spec.ingress.isEnabled |
      del(.spec.ingress.isEnabled)
    " "$file"
    
    echo "  ✅ Migrated successfully"
}

# Find all Neo4jKubernetes manifests
find . -name "*.yaml" -type f | while read file; do
    # Check if it's a Neo4jKubernetes resource
    if grep -q "kind: Neo4jKubernetes" "$file"; then
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
chmod +x migrate-neo4j-ingress.sh
./migrate-neo4j-ingress.sh
```

## Examples

### Basic Ingress Configuration

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: graph-db
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  container:
    persistenceEnabled: true
    diskSize: 10Gi
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: 2000m
        memory: 8Gi
  memoryConfig:
    heapMax: "1Gi"
    pageCache: "512m"
  ingress:
    enabled: true
    hostname: neo4j.example.com
```

### Production with Custom Hostname

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: prod-graph
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  container:
    persistenceEnabled: true
    diskSize: 100Gi
    resources:
      requests:
        cpu: 2000m
        memory: 8Gi
      limits:
        cpu: 8000m
        memory: 32Gi
  memoryConfig:
    heapMax: "4Gi"
    pageCache: "2Gi"
  ingress:
    enabled: true
    hostname: graph-prod.company.com
```

### Neo4j Without Ingress

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: internal-graph
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  container:
    persistenceEnabled: true
    diskSize: 50Gi
    resources:
      requests:
        cpu: 1000m
        memory: 4Gi
      limits:
        cpu: 4000m
        memory: 16Gi
  ingress:
    enabled: false
```

### Using with External-DNS

The `hostname` field works seamlessly with external-dns annotation:

```yaml
# Neo4j manifest
spec:
  ingress:
    enabled: true
    hostname: neo4j.example.com
```

This creates a LoadBalancer service with:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: neo4j-external-lb
  annotations:
    external-dns.alpha.kubernetes.io/hostname: neo4j.example.com
spec:
  type: LoadBalancer
  # ... service configuration
```

External-DNS then automatically:
1. Detects the annotation
2. Waits for LoadBalancer IP assignment
3. Creates DNS A record: `neo4j.example.com` → `<LoadBalancer-IP>`
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
  hostname: "graph.example.com"

# System uses: graph.example.com (exact match)
```

### 2. Simplified Module Implementation

**Code Reduction**:
- Pulumi: 15+ lines removed (hostname construction and internal hostname)
- Total: ~15 lines of code eliminated

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
- Subdomains: `neo4j.analytics.example.com`
- Environments: `neo4j-prod.example.com`, `neo4j-staging.example.com`
- Descriptive: `graph-db.example.com`, `analytics-neo4j.example.com`
- No pattern: Full freedom to match organizational DNS conventions

### 5. Removed Unused Features

**Internal Hostname**: Previously generated but never used in:
- LoadBalancer services
- Helm chart configurations
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

The new `Neo4jKubernetesIngress` message includes built-in validation:

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
  hostname: "neo4j.example.com"
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
cat > neo4j-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: test-neo4j
spec:
  container:
    persistenceEnabled: true
    diskSize: 10Gi
  ingress:
    enabled: true
    hostname: test-neo4j.example.com
EOF

# Deploy
planton pulumi up --manifest neo4j-test.yaml

# Verify LoadBalancer service created with correct annotation
kubectl get svc -n test-neo4j -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show: test-neo4j.example.com
```

**Scenario 2: Update Existing Deployment**
```bash
# Update existing manifest from old to new format
# Before: ingress.dns_domain = "example.com"
# After: ingress.hostname = "existing-neo4j.example.com"

# Apply update
planton pulumi up --manifest neo4j-existing.yaml

# Verify hostname annotation updated
kubectl get svc -n existing-neo4j -o yaml | \
  grep "external-dns.alpha.kubernetes.io/hostname"
# Should show new hostname
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > neo4j-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: Neo4jKubernetes
metadata:
  name: invalid-neo4j
spec:
  ingress:
    enabled: true
    # Missing hostname - should fail validation
EOF

# Attempt deploy
planton pulumi up --manifest neo4j-invalid.yaml
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

- **Neo4j Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/neo4jkubernetes/v1/`
- **External-DNS Integration**: See `ExternalDnsKubernetes` resource for DNS automation
- **Neo4j Helm Chart**: https://github.com/neo4j/helm-charts
- **CEL Validation**: https://github.com/bufbuild/protovalidate

## Breaking Change Checklist

- [x] Migration guide provided with step-by-step instructions
- [x] Automated migration script included
- [x] Documentation updated (examples, README, API docs)
- [x] Validation added to catch invalid configurations
- [x] Clear error messages for validation failures
- [x] Before/after comparison provided
- [x] Testing instructions included
- [x] Pulumi module updated
- [x] All examples updated to new syntax
- [x] Field naming consistency improved

## Deployment Status

✅ **Protobuf Contract**: Updated with custom ingress message and CEL validation  
✅ **Pulumi Module**: Hostname construction removed, direct usage implemented  
✅ **Terraform Variables**: Updated with new ingress structure  
✅ **Documentation**: All examples and READMEs updated  
✅ **Migration Script**: Automated migration script provided  
✅ **Validation**: CEL validation ensures hostname required when enabled  
✅ **Outputs**: Internal hostname output removed from Pulumi  
✅ **Build Verification**: All code compiled successfully with no errors  
✅ **Field Naming**: Consistent `persistence_enabled` naming

**Ready for**: Protobuf regeneration and user migration

## Future Enhancements

1. **Multiple Hostnames**: Support array of hostnames for multi-domain access
   ```yaml
   ingress:
     enabled: true
     hostnames:
       - neo4j.example.com
       - graph.example.com
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
     hostname: neo4j.example.com
     tls:
       enabled: true
       secretName: neo4j-tls
   ```

4. **Internal LoadBalancer**: Support internal vs external LoadBalancer
   ```yaml
   ingress:
     enabled: true
     hostname: neo4j.internal.example.com
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

**Impact**: This change improves the Neo4j Kubernetes API by providing user control over ingress hostnames, simplifying implementation, and removing unused features. The migration path is straightforward with clear documentation and automation tools.

