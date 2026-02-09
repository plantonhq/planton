# OpenStackInstance Pulumi Module

Provisions an OpenStack Compute instance with full networking, storage, and placement support.

## Resources Created

- `openstack_compute_instance_v2` -- A compute instance with networks, security groups, block devices, and placement hints

## Module Structure

```
module/
├── main.go       # Entry point: Resources()
├── locals.go     # FK resolution and local variable initialization
├── outputs.go    # Output constant definitions
└── instance.go   # Instance resource creation
```

## FK Resolution

The module resolves 5 types of StringValueOrRef foreign keys:
- `key_pair` -> keypair name
- `networks[].uuid` -> network UUID
- `networks[].port` -> port UUID
- `security_groups[]` -> security group names
- `server_group_id` -> server group UUID (mapped to scheduler_hints.group)
