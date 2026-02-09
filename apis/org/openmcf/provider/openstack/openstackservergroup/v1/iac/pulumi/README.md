# OpenStackServerGroup Pulumi Module

Provisions an OpenStack Compute server group using the Pulumi OpenStack provider.

## Resources Created

- `openstack_compute_servergroup_v2` -- A server group with the specified placement policy

## Usage

```bash
# Build
make build

# Preview (requires OpenStack credentials)
make test
```

## Module Structure

```
module/
├── main.go           # Entry point: Resources()
├── locals.go         # Local variable initialization
├── outputs.go        # Output constant definitions
└── server_group.go   # Server group resource creation
```
