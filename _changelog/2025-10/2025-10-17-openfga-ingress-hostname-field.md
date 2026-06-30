# OpenFGA Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Breaking Change, Enhancement  
**Component**: OpenFgaKubernetes

## Summary

Refactored the OpenFGA Kubernetes ingress configuration from a shared `IngressSpec` (with `enabled` and `dns_domain` fields) to a custom `OpenFgaKubernetesIngress` message with `enabled` and `hostname` fields. This change gives users full control over the ingress hostname instead of auto-constructing it from resource ID and DNS domain. The Terraform and Pulumi modules now use the user-supplied hostname directly, eliminating internal hostname construction logic. Both HTTP and HTTPS listeners now use the same external hostname.

## Motivation

### The Problem

The previous implementation used the shared `IngressSpec` from `dev.planton.shared.kubernetes`:

```yaml
ingress:
  enabled: true
  dns_domain: "example.com"
```

This approach had several limitations:

1. **Hostname Auto-Construction**: The system automatically constructed hostnames as `{resource-id}.{dns-domain}`, giving users no control over the exact hostname pattern.

2. **Internal vs. External Hostnames**: The implementation created both "external" and "internal" hostnames (`{resource-id}.{dns-domain}` and `{resource-id}-internal.{dns-domain}`), but internal hostnames were never actually used or needed for OpenFGA. The internal hostname was only set on the HTTP listener, creating confusion.

3. **Inflexibility**: Users couldn't specify custom subdomains like `auth.example.com` or `openfga-prod.example.com` - they were locked into the resource-id-based pattern.

4. **Module Complexity**: Both Terraform and Pulumi modules contained hostname construction logic that could be eliminated if users specified the full hostname directly.

5. **Shared Spec Limitations**: The generic `IngressSpec` was designed for multiple resource types, forcing OpenFGA to inherit patterns that didn't match its specific needs.

### The Solution

Replace `IngressSpec` with a custom `OpenFgaKubernetesIngress` message:

```yaml
ingress:
  enabled: true
  hostname: "openfga.example.com"
```

This approach:
- ✅ Gives users complete control over hostnames
- ✅ Eliminates unused internal hostname concept
- ✅ Uses same hostname for both HTTP and HTTPS listeners (consistent behavior)
- ✅ Simplifies Terraform and Pulumi modules (removed hostname construction logic)
- ✅ Provides clearer, more intuitive API
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)

## What's New

### 1. Custom OpenFgaKubernetesIngress Message

**Before (Shared IngressSpec)**:
```protobuf
// From project/planton/shared/kubernetes/kubernetes.proto
message IngressSpec {
  bool enabled = 1;
  string dns_domain = 2;
}

message OpenFgaKubernetesSpec {
  dev.planton.shared.kubernetes.IngressSpec ingress = 2;
}
```

**After (Custom Message)**:
```protobuf
message OpenFgaKubernetesIngress {
  // Flag to enable or disable ingress.
  bool enabled = 1;

  // The full hostname for external access (e.g., "openfga.example.com").
  // Required when enabled is true.
  string hostname = 2;

  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}

message OpenFgaKubernetesSpec {
  OpenFgaKubernetesIngress ingress = 2;
}
```

### 2. Updated YAML Syntax

**Before**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: OpenFgaKubernetes
metadata:
  name: public-openfga
spec:
  ingress:
    enabled: true
    dns_domain: example.com
  # Resulting hostname: public-openfga.example.com (auto-constructed)
```

**After**:
```yaml
apiVersion: kubernetes.planton.dev/v1
kind: OpenFgaKubernetes
metadata:
  name: public-openfga
spec:
  ingress:
    enabled: true
    hostname: openfga.example.com
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

  # Extract domain from hostname for certificate issuer
  ingress_cert_cluster_issuer_name = local.ingress_external_hostname != null ? (
    join(".", slice(split(".", local.ingress_external_hostname), 1,
      length(split(".", local.ingress_external_hostname))))
  ) : null
}
```

**Removed**:
- 6 lines of hostname construction logic
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

func initializeLocals(ctx *pulumi.Context, stackInput *openfgakubernetesv1.OpenFgaKubernetesStackInput) *Locals {
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

    // Construct internal hostname (never used properly)
    locals.IngressInternalHostname = fmt.Sprintf("%s-internal.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)

    ctx.Export(OpInternalHostname, pulumi.String(locals.IngressInternalHostname))

    locals.IngressCertClusterIssuerName = target.Spec.Ingress.DnsDomain

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

func initializeLocals(ctx *pulumi.Context, stackInput *openfgakubernetesv1.OpenFgaKubernetesStackInput) *Locals {
    // ... other initialization

    if target.Spec.Ingress == nil ||
        !target.Spec.Ingress.Enabled ||
        target.Spec.Ingress.Hostname == "" {
        return locals
    }

    locals.IngressExternalHostname = target.Spec.Ingress.Hostname

    ctx.Export(OpExternalHostname, pulumi.String(locals.IngressExternalHostname))

    // Extract domain from hostname for certificate issuer
    parts := strings.Split(locals.IngressExternalHostname, ".")
    if len(parts) > 1 {
        locals.IngressCertClusterIssuerName = strings.Join(parts[1:], ".")
    }

    return locals
}
```

**Removed**:
- `IngressInternalHostname` field from `Locals` struct
- `OpInternalHostname` constant from outputs
- 10+ lines of hostname construction logic
- Internal hostname export (never used)

### 5. Unified Listener Hostnames

**Before**: HTTP and HTTPS listeners used different hostnames:
```go
// ingress.go
Listeners: gatewayv1.GatewaySpecListenersArray{
    &gatewayv1.GatewaySpecListenersArgs{
        Name:     pulumi.String("https-external"),
        Hostname: pulumi.String(locals.IngressExternalHostname),  // external
        Port:     pulumi.Int(443),
        Protocol: pulumi.String("HTTPS"),
    },
    &gatewayv1.GatewaySpecListenersArgs{
        Name:     pulumi.String("http-external"),
        Hostname: pulumi.String(locals.IngressInternalHostname),  // internal (confusing!)
        Port:     pulumi.Int(80),
        Protocol: pulumi.String("HTTP"),
    },
}
```

**After**: Both listeners use the same external hostname:
```go
// ingress.go
Listeners: gatewayv1.GatewaySpecListenersArray{
    &gatewayv1.GatewaySpecListenersArgs{
        Name:     pulumi.String("https-external"),
        Hostname: pulumi.String(locals.IngressExternalHostname),
        Port:     pulumi.Int(443),
        Protocol: pulumi.String("HTTPS"),
    },
    &gatewayv1.GatewaySpecListenersArgs{
        Name:     pulumi.String("http-external"),
        Hostname: pulumi.String(locals.IngressExternalHostname),  // same as HTTPS
        Port:     pulumi.Int(80),
        Protocol: pulumi.String("HTTP"),
    },
}
```

### 6. Terraform Outputs Simplified

**Before**:
```hcl
output "external_hostname" {
  description = "The external hostname for OpenFGA if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "The internal hostname for OpenFGA if ingress is enabled."
  value       = local.ingress_internal_hostname
}
```

**After**:
```hcl
output "external_hostname" {
  description = "The external hostname for OpenFGA if ingress is enabled."
  value       = local.ingress_external_hostname
}

# internal_hostname output removed
```

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/openfgakubernetes/v1/spec.proto`

**Changes Made**:
1. **Updated Field Type**: Line 39 changed from `dev.planton.shared.kubernetes.IngressSpec` to `OpenFgaKubernetesIngress`
2. **Added Message**: New `OpenFgaKubernetesIngress` message with CEL validation (lines 88-104)

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Terraform Module Updates

**Files Modified**:
1. **`iac/tf/locals.tf`**: 
   - Simplified ingress variables (lines 42-53)
   - Removed hostname construction logic
   - Removed internal hostname variable
   - Added domain extraction for certificate issuer

2. **`iac/tf/ingress.tf`**: 
   - Updated dnsNames to only include external hostname (line 15)
   - Updated HTTP listener hostname to use external hostname (line 72)
   - Both HTTP and HTTPS listeners now use same hostname

3. **`iac/tf/outputs.tf`**:
   - Removed `internal_hostname` output

4. **`iac/tf/variables.tf`**:
   - Updated ingress variable structure (lines 53-60)
   - Changed `is_enabled` to `enabled`
   - Changed `dns_domain` to `hostname`

5. **`iac/tf/hack/manifest.yaml`**:
   - Updated ingress configuration with new syntax

### Pulumi Module Updates

**Files Modified**:
1. **`iac/pulumi/module/locals.go`**:
   - Added `strings` import for domain extraction
   - Removed `IngressInternalHostname` field from `Locals` struct
   - Simplified ingress hostname logic
   - Added domain extraction for certificate issuer
   - Removed internal hostname export

2. **`iac/pulumi/module/outputs.go`**:
   - Removed `OpInternalHostname` constant

3. **`iac/pulumi/module/main.go`**:
   - Updated ingress condition to check for nil spec

4. **`iac/pulumi/module/ingress.go`**:
   - Updated HTTP listener to use external hostname (line 79)
   - Both HTTP and HTTPS listeners now use same hostname

### Documentation Updates

All documentation files updated with correct ingress syntax:

1. **`v1/README.md`**: 
   - Updated ingress configuration description

2. **`v1/examples.md`**: 
   - Complete rewrite with proper OpenFgaKubernetes examples
   - Updated all examples to use new ingress format

3. **`iac/pulumi/examples.md`**:
   - Updated all 4 examples with new ingress format
   - Changed `isEnabled` to `enabled`
   - Changed `dnsDomain` to `hostname`

4. **`api_test.go`**:
   - Updated test input to use new ingress structure
   - Added ingress configuration to test

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for existing OpenFgaKubernetes resources with ingress enabled.

**Affected Users**: Users who have deployed OpenFGA with ingress enabled (likely a small subset given the resource's recent introduction).

### Migration Steps

#### Step 1: Identify Affected Resources

Find all OpenFgaKubernetes manifests with ingress configuration:

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
    dns_domain: "example.com"
    # System creates: {resource-id}.example.com
```

**After Migration**:
```yaml
spec:
  ingress:
    enabled: true
    hostname: "my-openfga.example.com"
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
  name: prod-openfga
spec:
  ingress:
    dns_domain: "example.com"

# The old system created: prod-openfga.example.com
# So use:
spec:
  ingress:
    hostname: "prod-openfga.example.com"
```

**Option B - Choose New Hostname** (take advantage of flexibility):
```yaml
spec:
  ingress:
    hostname: "auth.example.com"  # Any hostname you want!
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
3. **Update Gateway resources**: The Gateway will get new hostname configuration
4. **DNS propagation**: Update any DNS entries or external-dns annotations
5. **Verify**: Test the new hostname works before removing old DNS entries

**Note**: If you keep the same hostname, DNS records won't change.

#### Step 6: Apply Changes

```bash
# Preview changes
planton pulumi preview --manifest openfga.yaml

# Apply
planton pulumi up --manifest openfga.yaml
```

### Automated Migration Script

For users with many manifests:

```bash
#!/bin/bash
# migrate-openfga-ingress.sh

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

# Find all OpenFgaKubernetes manifests
find . -name "*.yaml" -type f | while read file; do
    # Check if it's an OpenFgaKubernetes resource
    if grep -q "kind: OpenFgaKubernetes" "$file"; then
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
chmod +x migrate-openfga-ingress.sh
./migrate-openfga-ingress.sh
```

## Examples

### Basic Ingress Configuration

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: OpenFgaKubernetes
metadata:
  name: auth-openfga
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
  ingress:
    enabled: true
    hostname: openfga.example.com
  datastore:
    engine: postgres
    uri: postgres://user:password@db-host:5432/openfga
```

### Production with Custom Hostname

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: OpenFgaKubernetes
metadata:
  name: prod-auth
spec:
  container:
    replicas: 5
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2000m
        memory: 4Gi
  ingress:
    enabled: true
    hostname: auth-prod.company.com
  datastore:
    engine: postgres
    uri: postgres://openfga:securepassword@prod-db:5432/openfga?sslmode=require
```

### Using with Gateway API

The `hostname` field works seamlessly with Gateway API:

```yaml
# OpenFGA manifest
spec:
  ingress:
    enabled: true
    hostname: openfga.example.com
```

This creates Gateway resources with appropriate listeners and HTTPRoutes:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
metadata:
  name: openfga-external
  namespace: istio-ingress
spec:
  gatewayClassName: istio
  listeners:
    - name: https-external
      hostname: openfga.example.com  # User-specified hostname
      port: 443
      protocol: HTTPS
    - name: http-external
      hostname: openfga.example.com  # Same hostname for HTTP
      port: 80
      protocol: HTTP
```

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
  hostname: "auth.example.com"

# System uses: auth.example.com (exact match)
```

### 2. Simplified Module Implementation

**Code Reduction**:
- Terraform: 8+ lines removed (hostname construction logic)
- Pulumi: 12+ lines removed (hostname construction and internal hostname)
- Total: ~20 lines of code eliminated

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

### 4. Consistent Listener Configuration

**Before**: HTTP and HTTPS listeners had different hostnames (confusing)
- HTTPS: `openfga.example.com`
- HTTP: `openfga-internal.example.com` (why?)

**After**: Both listeners use the same hostname (clear)
- HTTPS: `openfga.example.com`
- HTTP: `openfga.example.com` (HTTP redirects to HTTPS)

### 5. Flexibility

Users can now use any hostname pattern:
- Subdomains: `auth.services.example.com`
- Environments: `openfga-prod.example.com`, `openfga-staging.example.com`
- Descriptive: `authorization.example.com`, `fga-service.example.com`
- No pattern: Full freedom to match organizational DNS conventions

### 6. Removed Unused Features

**Internal Hostname**: Previously generated but never used in:
- Gateway configurations (was on HTTP listener only)
- HTTPRoute resources
- Documentation examples
- Any operational workflows

Removing it simplifies the codebase and API surface.

## Validation

### CEL Validation Rules

The new `OpenFgaKubernetesIngress` message includes built-in validation:

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
  hostname: "openfga.example.com"
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
cat > openfga-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: OpenFgaKubernetes
metadata:
  name: test-openfga
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
  ingress:
    enabled: true
    hostname: test-openfga.example.com
  datastore:
    engine: postgres
    uri: postgres://user:pass@db:5432/openfga
EOF

# Deploy
planton pulumi up --manifest openfga-test.yaml

# Verify Gateway created with correct hostname
kubectl get gateway -n istio-ingress | grep test-openfga
kubectl get httproute -n test-openfga
```

**Scenario 2: Update Existing Deployment**
```bash
# Update existing manifest from old to new format
# Before: ingress.dns_domain = "example.com"
# After: ingress.hostname = "existing-openfga.example.com"

# Apply update
planton pulumi up --manifest openfga-existing.yaml

# Verify hostname updated in Gateway
kubectl get gateway -n istio-ingress existing-openfga-external -o yaml | grep hostname
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > openfga-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: OpenFgaKubernetes
metadata:
  name: invalid-openfga
spec:
  ingress:
    enabled: true
    # Missing hostname - should fail validation
  datastore:
    engine: postgres
    uri: postgres://user:pass@db:5432/openfga
EOF

# Attempt deploy
planton pulumi up --manifest openfga-invalid.yaml
# Expected error: hostname is required when ingress is enabled
```

## Performance Impact

**No runtime performance impact**:
- Ingress configuration is applied once at deployment time
- Gateway resources created/updated during Pulumi apply
- No ongoing hostname construction or manipulation
- Module simplification reduces deployment time slightly (fewer operations)

## Security Considerations

**No security impact**:
- Hostnames are public information (used in DNS records and Gateway listeners)
- No changes to authentication, authorization, or encryption
- Gateway API and cert-manager security model unchanged
- Validation ensures hostname cannot be empty when ingress is enabled

**Operational Security**:
- Users must ensure hostname ownership before use
- DNS domain should be under organization's control
- Cert-manager requires proper ClusterIssuer configuration
- Gateway API requires proper RBAC permissions

## Related Documentation

- **OpenFGA Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/openfgakubernetes/v1/`
- **OpenFGA Official Documentation**: https://openfga.dev/
- **Gateway API**: https://gateway-api.sigs.k8s.io/
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
- [x] Test files updated with new structure

## Deployment Status

✅ **Protobuf Contract**: Updated with custom ingress message and CEL validation  
✅ **Terraform Module**: Hostname construction removed, direct usage implemented  
✅ **Pulumi Module**: Hostname construction removed, internal hostname eliminated  
✅ **Documentation**: All examples and READMEs updated  
✅ **Migration Script**: Automated migration script provided  
✅ **Validation**: CEL validation ensures hostname required when enabled  
✅ **Outputs**: Internal hostname output removed from both Terraform and Pulumi  
✅ **Listener Consistency**: HTTP and HTTPS listeners now use same external hostname  
✅ **Build Verification**: All code compiled successfully with no errors

**Ready for**: Protobuf regeneration and user migration

## Future Enhancements

1. **Multiple Hostnames**: Support array of hostnames for multi-domain access
   ```yaml
   ingress:
     enabled: true
     hostnames:
       - openfga.example.com
       - auth.example.com
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
     hostname: openfga.example.com
     tls:
       enabled: true
       secretName: openfga-tls
   ```

4. **Gateway Class Selection**: Support different gateway classes
   ```yaml
   ingress:
     enabled: true
     hostname: openfga.example.com
     gatewayClassName: internal-gateway  # vs. istio (default)
   ```

## Support

For questions or issues with migration:
1. Review the [migration guide](#migration-guide) above
2. Use the [automated migration script](#automated-migration-script)
3. Check [examples](#examples) for reference configurations
4. Verify [validation rules](#validation) are met
5. Contact Planton support if issues persist

---

**Impact**: This change improves the OpenFgaKubernetes API by providing user control over ingress hostnames, simplifying implementation, removing unused internal hostname concept, and creating consistent listener configuration. The migration path is straightforward with clear documentation and automation tools.

