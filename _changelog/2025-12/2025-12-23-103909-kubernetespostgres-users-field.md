# KubernetesPostgres: Add Users Field for Declarative Role Management

**Date**: December 23, 2025
**Type**: Feature
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added support for declaring PostgreSQL users/roles via the `users` field in `KubernetesPostgresSpec`. This complements the existing `databases` field and allows the Zalando operator to create users before assigning them as database owners—a requirement that was previously causing database creation to silently fail.

## Problem Statement / Motivation

The Zalando PostgreSQL operator requires users to be declared in `spec.users` before they can be referenced as database owners. Without explicit user declarations, the operator would skip database creation with a log message:

```
skipping creation of the "openfga" database, user "postgres" does not exist
```

### Pain Points

- Users couldn't create databases with custom owners without manual intervention
- The `databases` field alone was insufficient—owner roles had to exist first
- No declarative way to define users with specific permissions (e.g., `createdb`, `superuser`)
- Silent failures when referencing non-existent users as database owners

## Solution / What's New

Introduced a `users` field as a `repeated KubernetesPostgresUser` in `KubernetesPostgresSpec` where each user has:
- **name**: The username/role name (required)
- **flags**: Optional array of permission flags (e.g., `["createdb"]`, `["superuser"]`)

The Pulumi/Terraform modules convert this list to the Zalando operator's expected `map[string][]string` format.

### Configuration Example

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: PostgresKubernetes
metadata:
  name: my-postgres
spec:
  # Step 1: Declare users first
  users:
    - name: openfga
      flags: []              # Standard user with login
    - name: analytics
      flags:
        - createdb           # Can create databases
  
  # Step 2: Then assign as database owners
  databases:
    openfga: openfga
    analytics_db: analytics
```

### Common User Flags

| Flag | Description |
|------|-------------|
| `createdb` | User can create new databases |
| `superuser` | Full superuser privileges (use with caution) |
| `createrole` | User can create other roles |
| `inherit` | Inherits privileges of roles it belongs to |
| `login` | Can log in (default for users) |
| `replication` | Can initiate streaming replication |

## Implementation Details

### Proto Schema

Added new message type and field to `spec.proto`:

```protobuf
// PostgreSQL user/role definition
message KubernetesPostgresUser {
  // Username/role name (e.g., "app_user", "openfga")
  string name = 1 [(buf.validate.field).required = true];

  // User permission flags
  repeated string flags = 2;
}

// In KubernetesPostgresSpec:
repeated KubernetesPostgresUser users = 8;
```

### Pulumi Module Update

**File**: `iac/pulumi/module/main.go`

Converts the repeated `KubernetesPostgresUser` to Zalando's `map[string][]string`:

```go
// Convert users list to map[string][]string for Zalando operator
var usersMap pulumi.StringArrayMapInput
if len(locals.KubernetesPostgres.Spec.Users) > 0 {
    usersMapData := make(map[string][]string)
    for _, user := range locals.KubernetesPostgres.Spec.Users {
        usersMapData[user.Name] = user.Flags
    }
    usersMap = pulumi.ToStringArrayMap(usersMapData)
}
```

### Terraform Module Update

**File**: `iac/tf/variables.tf`

```hcl
users = optional(list(object({
  name  = string
  flags = optional(list(string), [])
})), [])
```

**File**: `iac/tf/database.tf`

Uses Terraform's `for` expression to convert list to map:

```hcl
users = { for user in var.spec.users : user.name => user.flags }
```

### Files Changed

| File | Change |
|------|--------|
| `apis/.../kubernetespostgres/v1/spec.proto` | Added `KubernetesPostgresUser` message and `users` field |
| `apis/.../kubernetespostgres/v1/spec.pb.go` | Auto-generated from proto |
| `apis/.../kubernetespostgres/v1/iac/pulumi/module/main.go` | Convert users list to map, pass to operator |
| `apis/.../kubernetespostgres/v1/iac/tf/variables.tf` | Added users variable as list of objects |
| `apis/.../kubernetespostgres/v1/iac/tf/database.tf` | Convert users list to map in manifest |
| `apis/.../kubernetespostgres/v1/examples.md` | Updated example with users + databases pattern |
| `apis/.../kubernetespostgres/v1/README.md` | Added User Configuration section |
| `apis/.../kubernetespostgres/v1/iac/pulumi/README.md` | Updated config parameters |

## Benefits

- **Declarative User Management**: Define users and their permissions in infrastructure-as-code
- **Complete Database Provisioning**: Users + databases work together without silent failures
- **Granular Permissions**: Assign specific flags to each user based on application needs
- **Operator Compliance**: Follows Zalando operator's expected pattern for user/database creation
- **Backward Compatible**: Existing deployments without users continue to work

## Impact

### Users
- Can now declare PostgreSQL users/roles before assigning them as database owners
- No more silent failures when creating databases with custom owners
- Fine-grained control over user permissions via flags

### Developers
- Both Pulumi and Terraform modules updated for feature parity
- Documentation updated across all relevant files
- Clear examples showing the users-before-databases pattern

### Operations
- Reduced debugging time—clear documentation on user declaration requirement
- Consistent user/role setup across environments

## Related Work

This feature builds on the `databases` field added earlier in this session. Together, they provide complete declarative control over PostgreSQL users and databases:

1. **databases field** (added earlier): Map of database names to owner roles
2. **users field** (this change): List of users with names and permission flags

The Zalando operator creates users first, then databases with the specified owners.

## Validation

- ✅ `make protos` - Proto stubs regenerated successfully
- ✅ `go test ./apis/dev/planton/provider/kubernetes/kubernetespostgres/v1/...` - Tests pass
- ✅ `make build` - Full build passes including Bazel, Go, and frontend builds

---

**Status**: ✅ Production Ready
**Timeline**: Single session implementation

