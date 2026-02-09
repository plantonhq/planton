# OpenStackServerGroup Pulumi Module -- Architecture Overview

## Resource Graph

This is the simplest OpenStack module -- a single resource with no dependencies.

```
OpenStackServerGroup
└── compute.ServerGroup (1 resource)
```

## Data Flow

1. `main.go` loads the `StackInput` from the Pulumi config
2. `module/main.go` initializes locals and sets up the OpenStack provider
3. `module/server_group.go` creates the `compute.ServerGroup` resource
4. Outputs (`server_group_id`, `name`, `members`, `region`) are exported

## Policy Wrapping

The proto spec uses a singular `policy` string, but both the Pulumi SDK and OpenStack API expect a list. The module wraps the singular value:

```go
Policies: pulumi.StringArray{pulumi.String(spec.Policy)}
```

## Outputs

| Output | Source |
|--------|--------|
| `server_group_id` | `createdServerGroup.ID()` |
| `name` | `createdServerGroup.Name` |
| `members` | `createdServerGroup.Members` |
| `region` | `createdServerGroup.Region` |
