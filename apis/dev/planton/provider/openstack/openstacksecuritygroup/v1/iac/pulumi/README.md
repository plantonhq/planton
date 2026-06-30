# OpenStackSecurityGroup Pulumi Module

Pulumi Go module for provisioning an OpenStack Neutron security group with optional inline rules.

## Usage

This module is invoked by the Planton CLI. The stack input is loaded from a base64-encoded YAML manifest.

```bash
# Preview changes
planton pulumi preview --manifest security-group.yaml --module-dir ./iac/pulumi

# Apply changes
planton pulumi up --manifest security-group.yaml --module-dir ./iac/pulumi --yes
```

## Local Development

```bash
# Build the module
make build

# Install required Pulumi plugins
make install-pulumi-plugins

# Run preview with test manifest
make test
```

## Resources Created

- `openstack:networking/secGroup:SecGroup` -- The security group itself
- `openstack:networking/secGroupRule:SecGroupRule` -- One per inline rule (keyed by `rule.key`)

## Stack Outputs

| Output | Description |
|--------|-------------|
| `security_group_id` | UUID of the security group |
| `name` | Name of the security group |
| `region` | OpenStack region |
