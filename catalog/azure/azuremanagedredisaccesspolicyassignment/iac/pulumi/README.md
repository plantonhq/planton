# AzureManagedRedisAccessPolicyAssignment - Pulumi Module

Pulumi implementation for the AzureManagedRedisAccessPolicyAssignment
kind.

## Architecture

```
managedredis.AccessPolicyAssignment (single resource)
```

## Key Design Decisions

- **The instance is addressed by ARM id** (`managed_redis_id`) -- the
  provider derives the default database's path from it; the assignment
  is a pure child resource.
- **The object id defaults to an identity's PRINCIPAL id** -- the
  workload-identity grant is the common case, and granting the client
  id instead is the classic mistake that fails at connect time, not
  deploy time (called out in the spec).
- **The policy is a constant, not a field** -- Managed Redis exposes
  exactly one built-in policy ("default") and the provider hardcodes
  it; a policy reference lands when Azure ships custom policies.
- **Everything is ForceNew** -- replacing a grant momentarily revokes
  and re-grants, which is safe for the grant class.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
