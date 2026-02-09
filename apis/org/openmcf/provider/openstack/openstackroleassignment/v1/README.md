# OpenStackRoleAssignment

An OpenStack Identity (Keystone) role assignment -- binds a role to a user or group on a project or domain.

## Overview

Role assignments are the fundamental authorization mechanism in OpenStack. They determine what actions a user or group can perform on a specific project or domain.

## Constraints

- Exactly one of `project_id` or `domain_id` must be set (scope)
- Exactly one of `user_id` or `group_id` must be set (principal)
- All fields are ForceNew (immutable)
- Admin-level operation

## Key Fields

| Field | Type | Description |
|-------|------|-------------|
| `role_id` | string | UUID of the role to assign (required) |
| `project_id` | StringValueOrRef | Project scope (FK to OpenStackProject) |
| `domain_id` | string | Domain scope (plain UUID) |
| `user_id` | string | User principal (plain UUID) |
| `group_id` | string | Group principal (plain UUID) |

## Terraform Resource

`openstack_identity_role_assignment_v3`

## Pulumi Resource

`openstack.identity.RoleAssignment`
