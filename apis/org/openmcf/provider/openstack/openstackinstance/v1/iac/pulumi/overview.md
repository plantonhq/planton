# OpenStackInstance Pulumi Module -- Architecture Overview

## Resource Graph

A single resource with the most complex argument set in the OpenStack component suite.

```
OpenStackInstance
└── compute.Instance (1 resource)
    ├── Networks: 1+ network attachments (uuid or port mode)
    ├── SecurityGroups: 0+ security group name references
    ├── BlockDevices: 0+ block device mappings
    ├── SchedulerHints: 0-1 server group reference
    └── Metadata, UserData, Tags: optional configuration
```

## Data Flow

1. `main.go` loads the `StackInput` from the Pulumi config
2. `module/locals.go` resolves all StringValueOrRef FKs:
   - key_pair -> keypair name
   - server_group_id -> server group UUID
   - security_groups[] -> list of SG names
   - networks[].uuid/port -> list of network/port UUIDs
3. `module/instance.go` builds the `compute.InstanceArgs` from spec + resolved locals
4. `server_group_id` is mapped to `SchedulerHints[].Group`
5. Outputs (instance_id, name, access_ip_v4, access_ip_v6, region) are exported

## FK Resolution Pattern

```
StringValueOrRef.GetValue() -> resolved string
  ├── Literal mode: returns the value directly
  └── value_from mode: FK resolver middleware resolves before IaC runs
```

All FKs are resolved in `initializeLocals()` before resource creation begins.
