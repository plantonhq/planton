# OpenStackRoleAssignment Pulumi Module Overview

## Architecture

Single-resource module with FK resolution for `project_id`.

## FK Resolution

- `project_id` uses `StringValueOrRef` -- resolved by middleware before IaC runs
- `resolveStringValueOrRef()` extracts the literal value from the oneof wrapper
- Other fields (`domain_id`, `user_id`, `group_id`, `role_id`) are plain strings
