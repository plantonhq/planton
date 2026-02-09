# OpenStackProject Pulumi Module Overview

## Architecture

Single-resource module: creates one `identity.Project` from the spec.

```
OpenStackProjectStackInput
  ├── target (OpenStackProject)
  │   ├── metadata.name → project name
  │   └── spec
  │       ├── description → project description
  │       ├── domain_id → Keystone domain
  │       ├── enabled → active state (default: true)
  │       ├── parent_id → parent project for hierarchy
  │       ├── tags → project tags
  │       └── region → region override
  └── provider_config → OpenStack credentials
```

## Outputs

| Output | Source |
|--------|--------|
| `project_id` | `createdProject.ID()` |
| `name` | `createdProject.Name` |
| `domain_id` | `createdProject.DomainId` |
| `enabled` | `createdProject.Enabled` |
| `region` | `createdProject.Region` |

## Notes

- `enabled` uses `GetEnabled()` getter since it's an `optional bool` with default
- Tags are converted from `[]string` to `pulumi.StringArray`
- All fields except `metadata.name` are optional
