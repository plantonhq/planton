# Elasticsearch Kubernetes Ingress Hostname Field

**Date**: October 17, 2025  
**Type**: Breaking Change, Enhancement  
**Component**: ElasticsearchKubernetes

## Summary

Refactored the Elasticsearch Kubernetes spec to provide separate elasticsearch and kibana configurations (`spec.elasticsearch` and `spec.kibana`), each with their own container and ingress settings. Replaced the shared `IngressSpec` (with `enabled` and `dns_domain` fields) with custom `ElasticsearchKubernetesIngress` messages featuring `enabled` and `hostname` fields. This change gives users full control over both Elasticsearch and Kibana ingress hostnames, eliminates hostname auto-construction logic, and removes unused internal hostname features. Additionally, moved kibana enabled flag from container to spec level and renamed all `is_enabled` fields to `enabled` for consistency.

## Motivation

### The Problem

The previous implementation had several structural and functional limitations:

**1. Flat Spec Structure**: 
```yaml
spec:
  elasticsearchContainer:
    replicas: 1
    resources: {...}
  kibanaContainer:
    isEnabled: true
    replicas: 1
    resources: {...}
  ingress:
    enabled: true
    dns_domain: "example.com"
```

This flat structure:
- Mixed container and ingress concerns at the same level
- Made it unclear that ingress applied to both services
- Separated the kibana enabled flag from its configuration
- Lacked hierarchical organization for related settings

**2. Shared IngressSpec for Both Services**:

The shared ingress configuration created a single hostname pattern for both Elasticsearch and Kibana:
- Elasticsearch: `{resource-id}.{dns-domain}`
- Kibana: `{resource-id}-kb.{dns-domain}`

This gave users no control over individual service hostnames.

**3. Internal Hostnames Never Used**:

The system generated internal hostnames for both services:
- Elasticsearch internal: `{resource-id}-internal.{dns-domain}`
- Kibana internal: `{resource-id}-kb-internal.{dns-domain}`

These were never actually used in:
- Gateway configurations
- HTTPRoute resources
- Documentation examples
- Operational workflows

**4. Inflexibility**:

Users couldn't specify custom hostnames like:
- `analytics.example.com` and `analytics-ui.example.com`
- `search-prod.example.com` and `kibana-prod.example.com`
- Different domains for each service

**5. Module Complexity**:

Both Terraform and Pulumi modules contained hostname construction logic that could be eliminated if users specified full hostnames directly.

### The Solution

Restructure the spec with hierarchical organization and separate ingress configurations:

```yaml
spec:
  elasticsearch:
    container:
      replicas: 1
      resources: {...}
    ingress:
      enabled: true
      hostname: "elasticsearch.example.com"
  kibana:
    enabled: true
    container:
      replicas: 1
      resources: {...}
    ingress:
      enabled: true
      hostname: "kibana.example.com"
```

This approach:
- ✅ Provides clear hierarchical organization
- ✅ Gives users complete control over both hostnames independently
- ✅ Eliminates unused internal hostname concept
- ✅ Simplifies Terraform and Pulumi modules (removed hostname construction logic)
- ✅ Moves kibana enabled flag to logical location (spec level)
- ✅ Enables any hostname pattern users need
- ✅ Maintains validation (hostname required when enabled)
- ✅ Consistent naming (renamed `is_enabled` to `enabled`)

## What's New

### 1. Restructured Spec with Elasticsearch and Kibana Groupings

**Before (Flat Structure)**:
```protobuf
message ElasticsearchKubernetesSpec {
  ElasticsearchKubernetesElasticsearchContainer elasticsearch_container = 1;
  ElasticsearchKubernetesKibanaContainer kibana_container = 2;
  dev.planton.shared.kubernetes.IngressSpec ingress = 3;
}

message ElasticsearchKubernetesKibanaContainer {
  bool is_enabled = 1;  // Separated from other kibana config
  int32 replicas = 2;
  ContainerResources resources = 3;
}
```

**After (Hierarchical Structure)**:
```protobuf
message ElasticsearchKubernetesSpec {
  ElasticsearchKubernetesElasticsearchSpec elasticsearch = 1;
  ElasticsearchKubernetesKibanaSpec kibana = 2;
}

message ElasticsearchKubernetesElasticsearchSpec {
  ElasticsearchKubernetesElasticsearchContainer container = 1;
  ElasticsearchKubernetesIngress ingress = 2;
}

message ElasticsearchKubernetesKibanaSpec {
  bool enabled = 1;  // Now at spec level, logically grouped
  ElasticsearchKubernetesKibanaContainer container = 2;
  ElasticsearchKubernetesIngress ingress = 3;
}
```

### 2. Separate Ingress Configurations

**Before (Shared Ingress)**:
```yaml
spec:
  ingress:
    enabled: true
    dns_domain: "example.com"
  # System creates:
  # - elasticsearch.example.com (auto-constructed)
  # - elasticsearch-kb.example.com (auto-constructed)
```

**After (Independent Ingress)**:
```yaml
spec:
  elasticsearch:
    ingress:
      enabled: true
      hostname: "search.example.com"
  kibana:
    ingress:
      enabled: true
      hostname: "analytics.example.com"
  # Hostnames: exactly as specified by user
```

### 3. Custom ElasticsearchKubernetesIngress Message

**New Message**:
```protobuf
message ElasticsearchKubernetesIngress {
  bool enabled = 1;
  string hostname = 2;
  
  option (buf.validate.message).cel = {
    id: "spec.ingress.hostname.required"
    expression: "!this.enabled || size(this.hostname) > 0"
    message: "hostname is required when ingress is enabled"
  };
}
```

**Validation Behavior**:

✅ **Valid** - Ingress disabled:
```yaml
ingress:
  enabled: false
```

✅ **Valid** - Ingress enabled with hostname:
```yaml
ingress:
  enabled: true
  hostname: "elasticsearch.example.com"
```

❌ **Invalid** - Ingress enabled without hostname:
```yaml
ingress:
  enabled: true
  # Error: hostname is required when ingress is enabled
```

### 4. Renamed Fields for Consistency

**Field Naming Changes**:
| Old Field                | New Field              | Notes |
|--------------------------|------------------------|-------|
| `is_enabled` (kibana)    | `enabled`              | ✅ More idiomatic |
| `is_persistence_enabled` | `persistence_enabled`  | ✅ Consistent pattern |

**YAML Syntax**:
```yaml
# Before
kibanaContainer:
  isEnabled: true
  
elasticsearchContainer:
  isPersistenceEnabled: true

# After  
kibana:
  enabled: true
  
elasticsearch:
  container:
    persistenceEnabled: true
```

### 5. Simplified Pulumi Module

**Before** (Hostname Construction):
```go
// locals.go
type Locals struct {
    ElasticsearchIngressExternalHostname string
    ElasticsearchIngressInternalHostname string  // Never used
    KibanaIngressExternalHostname        string
    KibanaIngressInternalHostname        string  // Never used
}

func initializeLocals(ctx *pulumi.Context, stackInput *elasticsearchkubernetesv1.ElasticsearchKubernetesStackInput) *Locals {
    if target.Spec.Ingress == nil ||
        !target.Spec.Ingress.Enabled ||
        target.Spec.Ingress.DnsDomain == "" {
        return locals
    }

    // Construct hostnames
    locals.ElasticsearchIngressExternalHostname = fmt.Sprintf("%s.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)
    locals.ElasticsearchIngressInternalHostname = fmt.Sprintf("%s-internal.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)
    locals.KibanaIngressExternalHostname = fmt.Sprintf("%s-kb.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)
    locals.KibanaIngressInternalHostname = fmt.Sprintf("%s-kb-internal.%s", locals.Namespace,
        target.Spec.Ingress.DnsDomain)

    locals.IngressHostnames = []string{
        locals.ElasticsearchIngressExternalHostname,
        locals.ElasticsearchIngressInternalHostname,
        locals.KibanaIngressExternalHostname,
        locals.KibanaIngressInternalHostname,
    }
}
```

**After** (Direct Usage):
```go
// locals.go
type Locals struct {
    ElasticsearchIngressExternalHostname string
    KibanaIngressExternalHostname        string
    // Internal hostname fields removed
}

func initializeLocals(ctx *pulumi.Context, stackInput *elasticsearchkubernetesv1.ElasticsearchKubernetesStackInput) *Locals {
    // Elasticsearch ingress
    if target.Spec.Elasticsearch.Ingress != nil &&
        target.Spec.Elasticsearch.Ingress.Enabled &&
        target.Spec.Elasticsearch.Ingress.Hostname != "" {
        
        locals.ElasticsearchIngressExternalHostname = target.Spec.Elasticsearch.Ingress.Hostname
        ctx.Export(OpElasticsearchExternalHostname, pulumi.String(locals.ElasticsearchIngressExternalHostname))
        locals.IngressHostnames = append(locals.IngressHostnames, locals.ElasticsearchIngressExternalHostname)
    }

    // Kibana ingress
    if target.Spec.Kibana != nil && target.Spec.Kibana.Enabled &&
        target.Spec.Kibana.Ingress != nil &&
        target.Spec.Kibana.Ingress.Enabled &&
        target.Spec.Kibana.Ingress.Hostname != "" {
        
        locals.KibanaIngressExternalHostname = target.Spec.Kibana.Ingress.Hostname
        ctx.Export(OpKibanaExternalHostname, pulumi.String(locals.KibanaIngressExternalHostname))
        locals.IngressHostnames = append(locals.IngressHostnames, locals.KibanaIngressExternalHostname)
    }
}
```

**Removed**:
- `ElasticsearchIngressInternalHostname` field
- `KibanaIngressInternalHostname` field
- `OpElasticsearchInternalHostname` constant
- `OpKibanaInternalHostname` constant
- 30+ lines of hostname construction logic
- Internal hostname exports (never used)

### 6. Simplified Terraform Module

**Before** (Hostname Construction):
```hcl
# locals.tf
locals {
  ingress_is_enabled = try(var.spec.ingress.is_enabled, false)
  ingress_dns_domain = try(var.spec.ingress.dns_domain, "")

  elasticsearch_ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}.${local.ingress_dns_domain}" : null

  elasticsearch_ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-internal.${local.ingress_dns_domain}" : null

  kibana_ingress_external_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-kb.${local.ingress_dns_domain}" : null

  kibana_ingress_internal_hostname = (
    local.ingress_is_enabled && local.ingress_dns_domain != ""
  ) ? "${local.resource_id}-kb-internal.${local.ingress_dns_domain}" : null

  ingress_hostnames = compact([
    local.elasticsearch_ingress_external_hostname,
    local.elasticsearch_ingress_internal_hostname,
    local.kibana_ingress_external_hostname,
    local.kibana_ingress_internal_hostname,
  ])
}
```

**After** (Direct Usage):
```hcl
# locals.tf
locals {
  # Elasticsearch ingress
  elasticsearch_ingress_is_enabled = try(var.spec.elasticsearch.ingress.enabled, false)
  elasticsearch_ingress_external_hostname = try(var.spec.elasticsearch.ingress.hostname, null)

  # Kibana ingress
  kibana_is_enabled = try(var.spec.kibana.enabled, false)
  kibana_ingress_is_enabled = local.kibana_is_enabled && try(var.spec.kibana.ingress.enabled, false)
  kibana_ingress_external_hostname = local.kibana_is_enabled ? try(var.spec.kibana.ingress.hostname, null) : null

  # Combine hostnames for certificate
  ingress_hostnames = compact([
    local.elasticsearch_ingress_external_hostname,
    local.kibana_ingress_external_hostname,
  ])

  # Certificate issuer: extract domain from first hostname
  ingress_cert_cluster_issuer_name = length(local.ingress_hostnames) > 0 ? (
    join(".", slice(split(".", local.ingress_hostnames[0]), 1, length(split(".", local.ingress_hostnames[0]))))
  ) : ""
}
```

**Removed**:
- 20+ lines of hostname construction logic
- Internal hostname variables (never used)
- DNS domain parsing and validation
- Redundant ingress conditional logic

## Implementation Details

### Protobuf Changes

**File**: `apis/project/planton/provider/kubernetes/workload/elasticsearchkubernetes/v1/spec.proto`

**Changes Made**:
1. **Restructured ElasticsearchKubernetesSpec**: Replaced flat container fields with hierarchical `elasticsearch` and `kibana` groupings
2. **Updated Field Options**: Changed from container-level defaults to spec-level defaults
3. **Removed Shared IngressSpec**: No longer imports shared kubernetes.proto for IngressSpec
4. **Added Custom Ingress Message**: New `ElasticsearchKubernetesIngress` message with CEL validation (similar to ClickHouse)
5. **Moved Kibana Enabled Flag**: From `kibana_container.is_enabled` to `kibana.enabled`
6. **Renamed Fields**: `is_enabled` → `enabled`, `is_persistence_enabled` → `persistence_enabled`

**Validation Strategy**: Uses CEL (Common Expression Language) to validate that `hostname` is required when `enabled` is true, providing clear error messages and type safety.

### Pulumi Module Updates

**Files Modified**:

1. **`iac/pulumi/module/locals.go`**:
   - Added `strings` import for domain extraction
   - Removed `ElasticsearchIngressInternalHostname` and `KibanaIngressInternalHostname` fields
   - Replaced shared ingress logic with separate elasticsearch and kibana ingress checks
   - Added domain extraction logic for certificate issuer
   - Updated all field references to use new nested structure

2. **`iac/pulumi/module/outputs.go`**:
   - Removed `OpElasticsearchInternalHostname` constant
   - Removed `OpKibanaInternalHostname` constant

3. **`iac/pulumi/module/elasticsearch.go`**:
   - Updated all references from `Spec.ElasticsearchContainer` to `Spec.Elasticsearch.Container`
   - Updated all references from `Spec.KibanaContainer` to `Spec.Kibana.Container`
   - Changed `KibanaContainer.IsEnabled` to `Kibana.Enabled`
   - Updated field names: `IsPersistenceEnabled` → `PersistenceEnabled`

4. **`iac/pulumi/module/main.go`**:
   - Updated ingress condition to check both elasticsearch and kibana ingress configs independently
   - Added kibana enabled check to ingress condition

5. **`iac/pulumi/module/ingress.go`**:
   - Wrapped Elasticsearch gateway/routes in conditional block checking `ElasticsearchIngressExternalHostname != ""`
   - Wrapped Kibana gateway/routes in conditional block checking `KibanaIngressExternalHostname != ""`
   - Already correctly references local hostname variables (no changes to resource definitions)

### Terraform Module Updates

**Files Modified**:

1. **`iac/tf/locals.tf`**:
   - Simplified ingress variables to separate elasticsearch and kibana
   - Removed hostname construction logic
   - Removed internal hostname variables
   - Added domain extraction for certificate issuer from first available hostname
   - Combined only external hostnames for certificate (removed internal hostnames)

2. **`iac/tf/elasticsearch.tf`**:
   - Updated all references from `var.spec.elasticsearch_container` to `var.spec.elasticsearch.container`
   - Updated all references from `var.spec.kibana_container` to `var.spec.kibana.container`
   - Changed `kibana_container.is_enabled` to `kibana.enabled`
   - Updated field names: `is_persistence_enabled` → `persistence_enabled`

3. **`iac/tf/ingress.tf`**:
   - Updated certificate count condition to use `length(local.ingress_hostnames) > 0`
   - Updated Elasticsearch gateway/routes count to use `elasticsearch_ingress_is_enabled && elasticsearch_ingress_external_hostname != null`
   - Updated Kibana gateway/routes count to use `kibana_ingress_is_enabled && kibana_ingress_external_hostname != null`
   - Already correctly references local hostname variables (no changes to resource definitions)

4. **`iac/tf/outputs.tf`**:
   - Removed `elasticsearch_internal_hostname` output
   - Removed `kibana_internal_hostname` output

5. **`iac/tf/variables.tf`**:
   - Restructured spec variable to match new hierarchical protobuf structure
   - Elasticsearch and kibana as top-level objects
   - Each with nested container and ingress objects
   - Updated field names to match new naming convention

### Documentation Updates

All documentation files updated with correct syntax:

1. **`v1/examples.md`**:
   - Updated Example 2 with new ingress format
   - Changed to hierarchical structure with separate elasticsearch and kibana configs

2. **`v1/iac/pulumi/examples.md`**:
   - Updated Example 2 with new ingress format
   - Matched top-level examples.md structure

3. **`v1/iac/tf/hack/manifest.yaml`**:
   - Complete restructure to new hierarchical format
   - Separate ingress configurations for elasticsearch and kibana

4. **`v1/api_test.go`**:
   - Updated test input to use new nested structure
   - Updated field names: `IsPersistenceEnabled` → `PersistenceEnabled`, `IsEnabled` → `Enabled`
   - Added ingress configurations for both elasticsearch and kibana

## Migration Guide

### Breaking Change Impact

This is a **breaking change** for all existing ElasticsearchKubernetes resources.

**Affected Users**: All users with deployed Elasticsearch instances (the resource type was recently introduced, so impact is limited).

### Migration Steps

#### Step 1: Identify Affected Resources

Find all ElasticsearchKubernetes manifests:

```bash
# Search for manifests
grep -r "kind: ElasticsearchKubernetes" *.yaml
```

#### Step 2: Update Manifest Structure

**Before Migration**:
```yaml
spec:
  elasticsearchContainer:
    replicas: 1
    isPersistenceEnabled: true
    diskSize: "10Gi"
    resources:
      requests:
        cpu: 500m
        memory: 1Gi
      limits:
        cpu: 1000m
        memory: 2Gi
  kibanaContainer:
    isEnabled: true
    replicas: 1
    resources:
      requests:
        cpu: 200m
        memory: 512Mi
      limits:
        cpu: 500m
        memory: 1Gi
  ingress:
    enabled: true
    dns_domain: "example.com"
    # System creates:
    # - {resource-id}.example.com (Elasticsearch)
    # - {resource-id}-kb.example.com (Kibana)
```

**After Migration**:
```yaml
spec:
  elasticsearch:
    container:
      replicas: 1
      persistenceEnabled: true
      diskSize: "10Gi"
      resources:
        requests:
          cpu: 500m
          memory: 1Gi
        limits:
          cpu: 1000m
          memory: 2Gi
    ingress:
      enabled: true
      hostname: "my-elasticsearch.example.com"
  kibana:
    enabled: true
    container:
      replicas: 1
      resources:
        requests:
          cpu: 200m
          memory: 512Mi
        limits:
          cpu: 500m
          memory: 1Gi
    ingress:
      enabled: true
      hostname: "my-kibana.example.com"
```

**Key Changes**:
| Old Structure                        | New Structure                          |
|--------------------------------------|----------------------------------------|
| `elasticsearchContainer`             | `elasticsearch.container`              |
| `elasticsearchContainer.isPersistenceEnabled` | `elasticsearch.container.persistenceEnabled` |
| `kibanaContainer`                    | `kibana.container`                     |
| `kibanaContainer.isEnabled`          | `kibana.enabled`                       |
| `ingress.enabled`                    | `elasticsearch.ingress.enabled` + `kibana.ingress.enabled` |
| `ingress.dns_domain`                 | `elasticsearch.ingress.hostname` + `kibana.ingress.hostname` |

#### Step 3: Determine Your Hostnames

The old system constructed hostnames as:
- Elasticsearch: `{resource-id}.{dns-domain}`
- Kibana: `{resource-id}-kb.{dns-domain}`

You need to replicate this or choose new hostnames:

**Option A - Keep Existing Hostnames** (recommended for minimal disruption):
```yaml
# If your manifest had:
metadata:
  name: prod-search
spec:
  ingress:
    dns_domain: "example.com"

# The old system created:
# - prod-search.example.com (Elasticsearch)
# - prod-search-kb.example.com (Kibana)

# So use:
spec:
  elasticsearch:
    ingress:
      hostname: "prod-search.example.com"
  kibana:
    ingress:
      hostname: "prod-search-kb.example.com"
```

**Option B - Choose New Hostnames** (take advantage of flexibility):
```yaml
spec:
  elasticsearch:
    ingress:
      hostname: "search.example.com"
  kibana:
    ingress:
      hostname: "analytics.example.com"
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

#### Step 5: Update DNS Records (if changing hostnames)

If you chose different hostnames than the auto-constructed ones:

1. **Before applying**: Note the current hostnames
2. **Update manifest**: Apply new configuration
3. **Verify Gateway/Ingress**: The resources will get new hostname configurations
4. **DNS propagation**: Update any external-dns annotations or manual DNS entries
5. **Verify**: Test the new hostnames work before removing old DNS entries

**Note**: If you keep the same hostnames, DNS records won't change.

#### Step 6: Apply Changes

```bash
# Preview changes
planton pulumi preview --manifest elasticsearch.yaml

# Apply
planton pulumi up --manifest elasticsearch.yaml
```

### Automated Migration Script

For users with many manifests:

```bash
#!/bin/bash
# migrate-elasticsearch-ingress.sh

get_resource_id() {
    local file=$1
    local id=$(yq eval '.metadata.id // .metadata.name' "$file")
    echo "$id"
}

migrate_file() {
    local file=$1
    echo "Processing $file..."
    
    if ! grep -q "kind: ElasticsearchKubernetes" "$file"; then
        echo "  Not an ElasticsearchKubernetes resource, skipping"
        return
    fi
    
    # Extract values
    local resource_id=$(get_resource_id "$file")
    local dns_domain=$(yq eval '.spec.ingress.dns_domain' "$file")
    
    if [[ "$dns_domain" == "null" ]]; then
        echo "  No ingress configuration, performing structure update only"
        # Restructure without ingress
        yq eval -i '
          .spec.elasticsearch.container = .spec.elasticsearchContainer |
          del(.spec.elasticsearchContainer) |
          .spec.elasticsearch.container.persistenceEnabled = .spec.elasticsearch.container.isPersistenceEnabled |
          del(.spec.elasticsearch.container.isPersistenceEnabled) |
          .spec.kibana.enabled = .spec.kibanaContainer.isEnabled |
          .spec.kibana.container = (.spec.kibanaContainer | del(.isEnabled)) |
          del(.spec.kibanaContainer)
        ' "$file"
    else
        # Construct hostnames
        local es_hostname="${resource_id}.${dns_domain}"
        local kibana_hostname="${resource_id}-kb.${dns_domain}"
        
        echo "  Elasticsearch hostname: $es_hostname"
        echo "  Kibana hostname: $kibana_hostname"
        
        # Restructure with ingress
        yq eval -i "
          .spec.elasticsearch.container = .spec.elasticsearchContainer |
          del(.spec.elasticsearchContainer) |
          .spec.elasticsearch.container.persistenceEnabled = .spec.elasticsearch.container.isPersistenceEnabled |
          del(.spec.elasticsearch.container.isPersistenceEnabled) |
          .spec.elasticsearch.ingress.enabled = .spec.ingress.enabled |
          .spec.elasticsearch.ingress.hostname = \"$es_hostname\" |
          .spec.kibana.enabled = .spec.kibanaContainer.isEnabled |
          .spec.kibana.container = (.spec.kibanaContainer | del(.isEnabled)) |
          del(.spec.kibanaContainer) |
          .spec.kibana.ingress.enabled = .spec.ingress.enabled |
          .spec.kibana.ingress.hostname = \"$kibana_hostname\" |
          del(.spec.ingress)
        " "$file"
    fi
    
    echo "  ✅ Migrated successfully"
}

# Find and migrate all ElasticsearchKubernetes manifests
find . -name "*.yaml" -type f | while read file; do
    if grep -q "kind: ElasticsearchKubernetes" "$file"; then
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
chmod +x migrate-elasticsearch-ingress.sh
./migrate-elasticsearch-ingress.sh
```

## Examples

### Basic Configuration Without Ingress

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ElasticsearchKubernetes
metadata:
  name: internal-search
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  elasticsearch:
    container:
      replicas: 1
      persistenceEnabled: true
      diskSize: "50Gi"
      resources:
        requests:
          cpu: 500m
          memory: 2Gi
        limits:
          cpu: 2000m
          memory: 8Gi
  kibana:
    enabled: true
    container:
      replicas: 1
      resources:
        requests:
          cpu: 200m
          memory: 512Mi
        limits:
          cpu: 500m
          memory: 1Gi
```

### Production with Separate Ingress Hostnames

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ElasticsearchKubernetes
metadata:
  name: production-search
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  elasticsearch:
    container:
      replicas: 3
      persistenceEnabled: true
      diskSize: "200Gi"
      resources:
        requests:
          cpu: 1000m
          memory: 4Gi
        limits:
          cpu: 4000m
          memory: 16Gi
    ingress:
      enabled: true
      hostname: search-prod.company.com
  kibana:
    enabled: true
    container:
      replicas: 2
      resources:
        requests:
          cpu: 500m
          memory: 1Gi
        limits:
          cpu: 2000m
          memory: 4Gi
    ingress:
      enabled: true
      hostname: analytics-prod.company.com
```

### Elasticsearch Only (No Kibana)

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: ElasticsearchKubernetes
metadata:
  name: logs-cluster
spec:
  kubernetesProviderConfigId: k8s-cluster-01
  elasticsearch:
    container:
      replicas: 3
      persistenceEnabled: true
      diskSize: "100Gi"
      resources:
        requests:
          cpu: 1000m
          memory: 4Gi
        limits:
          cpu: 4000m
          memory: 16Gi
    ingress:
      enabled: true
      hostname: logs.example.com
  kibana:
    enabled: false  # Kibana disabled
```

### Using with Istio Gateway API

The `hostname` fields work seamlessly with Gateway API and external-dns:

```yaml
# Elasticsearch manifest
spec:
  elasticsearch:
    ingress:
      enabled: true
      hostname: search.example.com
  kibana:
    ingress:
      enabled: true
      hostname: kibana.example.com
```

This creates Gateway resources with appropriate listeners and HTTPRoutes that integrate with external-dns for automatic DNS record creation.

## Benefits

### 1. Hierarchical Organization

**Before**: Flat structure mixing concerns
```yaml
spec:
  elasticsearchContainer: {...}
  kibanaContainer: {...}
  ingress: {...}
```

**After**: Clear grouping of related settings
```yaml
spec:
  elasticsearch:
    container: {...}
    ingress: {...}
  kibana:
    container: {...}
    ingress: {...}
```

**Benefits**:
- Clear separation of Elasticsearch and Kibana concerns
- Easier to understand and maintain
- Logical grouping of related fields
- Kibana enabled flag now co-located with kibana configuration

### 2. Independent Hostname Control

**Before**: System decides hostname pattern for both services
```yaml
ingress:
  dns_domain: "example.com"
# System creates:
# - my-search.example.com (Elasticsearch)
# - my-search-kb.example.com (Kibana)
```

**After**: User decides exact hostname for each service
```yaml
elasticsearch:
  ingress:
    hostname: "search.example.com"
kibana:
  ingress:
    hostname: "analytics.example.com"
# System uses exact hostnames provided
```

### 3. Simplified Module Implementation

**Code Reduction**:
- Pulumi: 30+ lines removed (hostname construction and internal hostname logic)
- Terraform: 25+ lines removed (hostname construction logic)
- Total: ~55 lines of code eliminated

**Maintenance Benefits**:
- Fewer edge cases to handle
- No string manipulation or formatting logic
- Direct pass-through from manifest to Kubernetes
- Clearer code flow

### 4. Clearer API

**Before** (Multi-step mental model):
1. User provides DNS domain
2. System constructs Elasticsearch hostname from resource ID + DNS domain
3. System constructs Kibana hostname from resource ID + "-kb" + DNS domain
4. System constructs internal hostnames (never used)

**After** (Direct mental model):
1. User provides exact Elasticsearch hostname
2. User provides exact Kibana hostname

### 5. Flexibility

Users can now use any hostname patterns:
- Different domains: `search.company.com`, `kibana.analytics.com`
- Subdomain patterns: `prod.search.example.com`, `prod.kibana.example.com`
- Environment-specific: `search-staging.example.com`, `kibana-staging.example.com`
- Descriptive: `elastic-logs.example.com`, `kibana-dashboard.example.com`

### 6. Removed Unused Features

**Internal Hostnames**: Previously generated but never used in:
- Gateway configurations
- HTTPRoute resources
- Service definitions
- Documentation examples
- Any operational workflows

Removing them simplifies the codebase and API surface.

### 7. Consistent Naming

Renamed all `is_enabled` and `is_persistence_enabled` to `enabled` and `persistence_enabled`:
- More idiomatic
- Matches modern Go/protobuf conventions
- Cleaner generated code in all languages
- Consistent with other Planton resources

## Validation

### CEL Validation Rules

The new `ElasticsearchKubernetesIngress` message includes built-in validation:

```protobuf
option (buf.validate.message).cel = {
  id: "spec.ingress.hostname.required"
  expression: "!this.enabled || size(this.hostname) > 0"
  message: "hostname is required when ingress is enabled"
};
```

**Validation Behavior**:

✅ **Valid** - Both ingresses disabled:
```yaml
elasticsearch:
  ingress:
    enabled: false
kibana:
  ingress:
    enabled: false
```

✅ **Valid** - One ingress enabled with hostname:
```yaml
elasticsearch:
  ingress:
    enabled: true
    hostname: "search.example.com"
kibana:
  ingress:
    enabled: false
```

❌ **Invalid** - Ingress enabled without hostname:
```yaml
elasticsearch:
  ingress:
    enabled: true
    # Error: hostname is required when ingress is enabled
```

## Testing

### Test Scenarios

**Scenario 1: New Deployment with Both Ingresses**
```bash
# Create manifest with new syntax
cat > elasticsearch-test.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: ElasticsearchKubernetes
metadata:
  name: test-search
spec:
  elasticsearch:
    container:
      replicas: 1
      persistenceEnabled: true
      diskSize: "20Gi"
    ingress:
      enabled: true
      hostname: test-elasticsearch.example.com
  kibana:
    enabled: true
    container:
      replicas: 1
    ingress:
      enabled: true
      hostname: test-kibana.example.com
EOF

# Deploy
planton pulumi up --manifest elasticsearch-test.yaml

# Verify Gateway resources created with correct hostnames
kubectl get gateway -n istio-ingress
kubectl get httproute -n test-search
```

**Scenario 2: Elasticsearch Only (No Kibana)**
```bash
# Kibana disabled
cat > elasticsearch-no-kibana.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: ElasticsearchKubernetes
metadata:
  name: logs-only
spec:
  elasticsearch:
    container:
      replicas: 1
      persistenceEnabled: true
      diskSize: "50Gi"
    ingress:
      enabled: true
      hostname: logs.example.com
  kibana:
    enabled: false
EOF

# Deploy
planton pulumi up --manifest elasticsearch-no-kibana.yaml

# Verify only Elasticsearch Gateway created
kubectl get gateway -n istio-ingress | grep logs-only
```

**Scenario 3: Validation Error**
```bash
# Try invalid configuration
cat > elasticsearch-invalid.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: ElasticsearchKubernetes
metadata:
  name: invalid-search
spec:
  elasticsearch:
    ingress:
      enabled: true
      # Missing hostname - should fail validation
EOF

# Attempt deploy
planton pulumi up --manifest elasticsearch-invalid.yaml
# Expected error: hostname is required when ingress is enabled
```

## Performance Impact

**No runtime performance impact**:
- Ingress configuration applied once at deployment time
- Gateway and HTTPRoute resources created/updated during Pulumi apply
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
- DNS domains should be under organization's control
- Cert-manager requires proper ClusterIssuer configuration
- Gateway API requires proper RBAC permissions

## Related Documentation

- **Elasticsearch Kubernetes API**: `apis/project/planton/provider/kubernetes/workload/elasticsearchkubernetes/v1/`
- **Elastic Cloud on Kubernetes**: https://www.elastic.co/guide/en/cloud-on-k8s/current/index.html
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

✅ **Protobuf Contract**: Updated with hierarchical spec and custom ingress messages  
✅ **Terraform Module**: Hostname construction removed, direct usage implemented  
✅ **Pulumi Module**: Hostname construction removed, internal hostnames eliminated  
✅ **Documentation**: All examples and READMEs updated  
✅ **Migration Script**: Automated migration script provided  
✅ **Validation**: CEL validation ensures hostname required when enabled  
✅ **Outputs**: Internal hostname outputs removed from both Terraform and Pulumi  
✅ **Build Verification**: All code compiled successfully with no errors  
✅ **Field Naming**: Consistent `enabled` and `persistence_enabled` naming throughout

**Ready for**: User migration and deployment

## Future Enhancements

1. **External Secrets Integration**: Reference existing Kubernetes Secrets for Elasticsearch credentials
   ```yaml
   elasticsearch:
     credentialsSecretRef:
       name: elasticsearch-admin-credentials
   ```

2. **Multiple Hostnames**: Support array of hostnames for multi-domain access
   ```yaml
   elasticsearch:
     ingress:
       enabled: true
       hostnames:
         - search.example.com
         - search.example.org
   ```

3. **Hostname Validation**: Add regex validation for DNS compliance
   ```protobuf
   string hostname = 2 [
     (buf.validate.field).string.pattern = "^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"
   ];
   ```

4. **TLS Configuration**: Add optional TLS cert configuration
   ```yaml
   elasticsearch:
     ingress:
       enabled: true
       hostname: search.example.com
       tls:
         secretName: search-tls
   ```

5. **Path-Based Routing**: Support different URL paths for services
   ```yaml
   elasticsearch:
     ingress:
       hostname: analytics.example.com
       path: /search
   kibana:
     ingress:
       hostname: analytics.example.com
       path: /dashboard
   ```

## Support

For questions or issues with migration:
1. Review the [migration guide](#migration-guide) above
2. Use the [automated migration script](#automated-migration-script)
3. Check [examples](#examples) for reference configurations
4. Verify [validation rules](#validation) are met
5. Contact Planton support if issues persist

---

**Impact**: This change improves the ElasticsearchKubernetes API by providing hierarchical organization, independent hostname control for Elasticsearch and Kibana, simplified implementation, and removing unused features. The migration path is straightforward with clear documentation and automation tools.

