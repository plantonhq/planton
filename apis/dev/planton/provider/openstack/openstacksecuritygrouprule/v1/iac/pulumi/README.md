# OpenStackSecurityGroupRule Pulumi Module

This Pulumi module provisions a standalone OpenStack Neutron security group rule.

## Resources Created

- `openstack.networking.SecGroupRule` -- A single security group rule attached to the specified security group

## Usage

```bash
# Build the module
make build

# Run a preview with the test manifest
make test

# Debug with a custom manifest
./debug.sh path/to/manifest.yaml
```

## Required Plugins

- `openstack` v5.4.0+

## Stack Outputs

| Output | Description |
|--------|-------------|
| `rule_id` | UUID of the created security group rule |
| `security_group_id` | UUID of the parent security group |
| `direction` | Direction of the rule (ingress/egress) |
| `protocol` | IP protocol of the rule |
| `port_range_min` | Lower bound of port range |
| `port_range_max` | Upper bound of port range |
| `region` | OpenStack region |
