# OpenStackProject Research Documentation

## Terraform Resource: `openstack_identity_project_v3`

### Provider Source

- **Package**: `openstack`
- **Resource**: `resource_openstack_identity_project_v3.go`
- **Provider**: terraform-provider-openstack v3.x

### Schema (Complete)

| Field | Type | Required | Computed | ForceNew | Default | Description |
|-------|------|----------|----------|----------|---------|-------------|
| `region` | string | Optional | Yes | Yes | - | OpenStack region |
| `name` | string | Optional | No | No | - | Project name |
| `description` | string | Optional | No | No | - | Human-readable description |
| `domain_id` | string | Optional | Yes | Yes | - | Keystone domain UUID |
| `enabled` | bool | Optional | No | No | `true` | Whether project is active |
| `is_domain` | bool | Optional | No | Yes | `false` | Whether this is a domain (not a project) |
| `parent_id` | string | Optional | Yes | Yes | - | Parent project UUID |
| `tags` | set(string) | Optional | No | No | - | Tags for filtering |

### Fields Excluded (80/20 Analysis)

| Field | Reason |
|-------|--------|
| `is_domain` | Creating Keystone domains via the Project API is extremely rare and confusing. 99.9% of usage creates projects (tenants), not domains. |
| `name` | Derived from `metadata.name` per Planton convention |

### Pulumi SDK

- **Package**: `github.com/pulumi/pulumi-openstack/sdk/v5/go/openstack/identity`
- **Function**: `identity.NewProject()`
- **Args**: `identity.ProjectArgs`

### Key Behaviors

1. **Admin-only**: Requires admin role or equivalent permissions
2. **ForceNew fields**: `region`, `domain_id`, `is_domain`, `parent_id` -- changing these recreates the project
3. **Computed fields**: `region`, `domain_id`, `parent_id` -- OpenStack provides defaults
4. **Tags**: Stored as a set (deduplicated), queryable via API filters
5. **enabled=false**: Disables access but does NOT delete resources in the project

### Common Use Cases

1. **Tenant provisioning**: Create isolated projects for development teams
2. **Landing zone**: Automated project creation with baseline configuration
3. **Organizational hierarchy**: Nested projects for department/team structure
4. **Decommissioning**: Disable projects (enabled=false) before cleanup
