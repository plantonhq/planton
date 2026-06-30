# KubernetesPostgres: Add Support for Creating Multiple Databases

**Date**: December 23, 2025
**Type**: Feature
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added support for specifying database names and their owner roles when deploying PostgreSQL on Kubernetes. The new `databases` field in `KubernetesPostgresSpec` enables users to declaratively define multiple databases to be created during cluster initialization, eliminating the need for manual database creation after deployment.

## Problem Statement / Motivation

Previously, deploying PostgreSQL via KubernetesPostgres only created the default `postgres` database. Users who needed multiple databases for their applications (e.g., separate databases for different microservices, analytics, or reporting) had to manually create them after deployment.

### Pain Points

- Manual database creation required post-deployment SQL commands
- No declarative way to define database structure in infrastructure-as-code
- Inconsistent database setup across environments (dev, staging, prod)
- Additional operational burden for platform teams

## Solution / What's New

Introduced a `databases` field as a `map<string, string>` in `KubernetesPostgresSpec` where:
- **Key**: Database name (e.g., `app_database`, `analytics_db`)
- **Value**: Owner role name (e.g., `app_user`, `analytics_role`)

The Zalando PostgreSQL operator automatically creates both the databases and their owner roles during cluster initialization.

### Configuration Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: multi-db-server
spec:
  target_cluster:
    cluster_name: "my-gke-cluster"
  namespace:
    value: "multi-db-server"
  create_namespace: true
  container:
    replicas: 1
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
    diskSize: "20Gi"
  
  # Create multiple databases with their owner roles
  databases:
    app_database: app_user
    analytics_db: analytics_role
    reporting: reporting_user
```

## Implementation Details

### Proto Schema Change

Added to `spec.proto`:

```protobuf
// Map of database names to their owner roles.
// Key: database name (e.g., "app_database", "analytics_db")
// Value: owner role name (e.g., "app_user", "analytics_role")
// The operator will create these databases during cluster initialization.
// If not specified, only the default "postgres" database will be available.
map<string, string> databases = 7;
```

### Pulumi Module Update

**File**: `iac/pulumi/module/main.go`

```go
// Convert databases map if specified
var databasesMap pulumi.StringMapInput
if len(locals.KubernetesPostgres.Spec.Databases) > 0 {
    databasesMap = pulumi.ToStringMap(locals.KubernetesPostgres.Spec.Databases)
}

// Added to PostgresqlSpecArgs
Databases: databasesMap,
```

### Terraform Module Update

**File**: `iac/tf/variables.tf`

```hcl
# Map of database names to their owner roles.
databases = optional(map(string), {})
```

**File**: `iac/tf/database.tf`

Uses Terraform's `merge()` function to conditionally include the databases field only when specified, maintaining backward compatibility.

### Files Changed

| File | Change |
|------|--------|
| `apis/.../kubernetespostgres/v1/spec.proto` | Added `databases` field |
| `apis/.../kubernetespostgres/v1/spec.pb.go` | Auto-generated from proto |
| `apis/.../kubernetespostgres/v1/iac/pulumi/module/main.go` | Pass databases to operator |
| `apis/.../kubernetespostgres/v1/iac/tf/variables.tf` | Added databases variable |
| `apis/.../kubernetespostgres/v1/iac/tf/database.tf` | Include databases in manifest |
| `apis/.../kubernetespostgres/v1/examples.md` | Added example with databases |
| `apis/.../kubernetespostgres/v1/README.md` | Documented databases feature |
| `apis/.../kubernetespostgres/v1/iac/pulumi/README.md` | Updated config parameters |

## Benefits

- **Declarative Database Management**: Define all databases in infrastructure-as-code
- **Environment Consistency**: Same database structure across dev/staging/prod
- **Reduced Operational Burden**: No manual post-deployment SQL required
- **Automatic Role Creation**: Owner roles created with appropriate permissions
- **Backward Compatible**: Existing deployments unaffected (field is optional)

## Impact

### Users
- Can now declare multiple databases in their YAML manifests
- Automatic database and role provisioning during cluster bootstrap
- Improved GitOps workflow for database infrastructure

### Developers
- Both Pulumi and Terraform modules updated for feature parity
- Documentation updated across all relevant files
- Examples provided for immediate adoption

### Operations
- Reduced manual intervention for database setup
- Consistent naming and ownership across environments

## Validation

- ✅ `make protos` - Proto stubs regenerated successfully
- ✅ `go test ./apis/dev/planton/provider/kubernetes/kubernetespostgres/v1/...` - Tests pass
- ✅ `make build` - Full build passes including Bazel, Go, and frontend builds

## Related Work

This feature leverages the Zalando PostgreSQL operator's native `spec.databases` field, which has been supported by the operator but was not exposed through our API. See the research document at `v1/docs/README.md` for operator comparison and design rationale.

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

