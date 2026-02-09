# OpenStackProject

An OpenStack Identity (Keystone) project -- the fundamental organizational unit in OpenStack.

## Overview

A project (historically called a "tenant") provides resource isolation, quota boundaries, and access control scoping. All cloud resources in OpenStack (instances, volumes, networks, security groups) belong to a project.

## When to Use

- **Tenant provisioning**: Automatically create projects for development teams
- **Landing zone automation**: Provision baseline project structure with networking and access controls
- **Multi-tenant environments**: Manage isolated project hierarchies

## Important Notes

- **Admin-only operation**: Creating projects requires admin credentials
- **All fields are optional except metadata.name**: Projects can be created with minimal configuration
- **enabled defaults to true**: Projects are active by default; set to false to disable access without deleting resources

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Human-readable project description |
| `domain_id` | string | Keystone domain (ForceNew, defaults to 'default') |
| `enabled` | bool | Whether the project is active (default: true) |
| `parent_id` | string | Parent project UUID for nested hierarchies (ForceNew) |
| `tags` | list | Tags for filtering and organization |
| `region` | string | Region override (rarely needed for Keystone) |

## Outputs

| Output | Description |
|--------|-------------|
| `project_id` | UUID of the created project (primary FK target) |
| `name` | Project name (from metadata.name) |
| `domain_id` | Domain the project belongs to |
| `enabled` | Whether the project is active |
| `region` | OpenStack region |

## Examples

See [examples.md](examples.md) for YAML manifests.

## Terraform Resource

`openstack_identity_project_v3`

## Pulumi Resource

`openstack.identity.Project`
