---
title: "OpenStackRoleAssignment Research Documentation"
description: "OpenStackRoleAssignment Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackroleassignment"
---

# OpenStackRoleAssignment Research Documentation

## Terraform Resource: `openstack_identity_role_assignment_v3`

### Schema

All fields are ForceNew. No computed fields except `region`.

| Field | Type | ConflictsWith |
|-------|------|---------------|
| `role_id` | Required | - |
| `project_id` | Optional | `domain_id` |
| `domain_id` | Optional | `project_id` |
| `user_id` | Optional | `group_id` |
| `group_id` | Optional | `user_id` |
| `region` | Optional | - |

### Pulumi SDK

- **Function**: `identity.NewRoleAssignment()`
- **Args**: `identity.RoleAssignmentArgs`
