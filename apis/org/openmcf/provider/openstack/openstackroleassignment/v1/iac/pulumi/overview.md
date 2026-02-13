# OpenStackRoleAssignment Pulumi Module Overview

## Architecture

Single-resource module for OpenStack Identity role assignments.

## StringValueOrRef Fields

- `project_id` uses `StringValueOrRef` -- the platform middleware resolves `valueFrom`
  references before IaC modules run, so the module calls `.GetValue()` to extract the
  resolved literal string
- Other fields (`domain_id`, `user_id`, `group_id`, `role_id`) are plain strings
