# OpenStackNetworkPort Pulumi Module

Provisions an OpenStack Neutron port with fixed IPs, security groups, and optional MAC address using the Pulumi OpenStack provider.

## Architecture

The module creates a single `networking.Port` resource with:
- Required network attachment (FK to OpenStackNetwork)
- Optional fixed IP allocations from subnets (FK to OpenStackSubnet inside nested FixedIp)
- Optional security group assignments (repeated FK to OpenStackSecurityGroup)
- Optional explicit security group bypass (no_security_groups)

## Local Development

```bash
# Build the binary
make build

# Install required Pulumi plugins
make install-pulumi-plugins

# Run preview with test manifest
make test
```

## Debug

```bash
# Export the test manifest as base64
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)

# Run Pulumi preview
pulumi preview --stack test --non-interactive
```
