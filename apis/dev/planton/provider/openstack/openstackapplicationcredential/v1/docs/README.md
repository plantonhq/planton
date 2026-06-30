# OpenStackApplicationCredential Research Documentation

## Terraform Resource: `openstack_identity_application_credential_v3`

### Key Behaviors

1. **Entirely immutable**: Every field is ForceNew. Any change destroys and recreates.
2. **Secret handling**: If user provides a secret, it's used. Otherwise, OpenStack generates one. The secret is returned once and stored in TF state (sensitive).
3. **Project scoping**: `project_id` is Computed from the auth scope, NOT user-configurable.
4. **Access rules**: Fine-grained API restriction. Each rule has path, method, service.
5. **Roles**: If omitted, inherits all roles of the creating user on the current project.

### Fields Excluded (80/20)

| Field | Reason |
|-------|--------|
| `project_id` | Computed from auth scope, not user-configurable |
| `name` | Derived from `metadata.name` |

### Pulumi SDK

- **Package**: `github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/identity`
- **Function**: `identity.NewApplicationCredential()`
