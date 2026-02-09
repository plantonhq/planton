# OpenStackRoleAssignment Terraform Module

This Terraform module provisions an OpenStack Identity role assignment.

## Resources Created

- `openstack_identity_role_assignment_v3` -- Keystone role assignment

## Important Notes

- All fields are ForceNew (immutable)
- Exactly one of `project_id` or `domain_id` must be set
- Exactly one of `user_id` or `group_id` must be set
