# OpenStackProject Pulumi Module

This Pulumi module provisions an OpenStack Identity (Keystone) project.

## Resources Created

- `openstack_identity_project_v3` -- Keystone project with description, domain, tags, and enabled state

## Usage

This module is invoked by the OpenMCF CLI. For local development:

```bash
make build
make test
```

## Debug

```bash
# Export stack input from manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)

# Run Pulumi preview
pulumi preview --stack test --non-interactive
```
