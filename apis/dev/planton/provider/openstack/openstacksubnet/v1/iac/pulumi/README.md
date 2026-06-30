# OpenStackSubnet Pulumi Module

This directory contains the Pulumi Go module for provisioning OpenStack Neutron subnets.

## Structure

```
iac/pulumi/
├── main.go           # Entrypoint (Pulumi program)
├── Pulumi.yaml       # Pulumi project config
├── Makefile          # Build and test targets
├── module/
│   ├── main.go       # Resources() entry point
│   ├── locals.go     # Input extraction and FK resolution
│   ├── subnet.go     # Subnet resource creation
│   └── outputs.go    # Output constants
└── README.md         # This file
```

## How It Works

1. The CLI loads the `OpenStackSubnetStackInput` from the stack config (base64-encoded YAML manifest)
2. `module.Resources()` is called with the stack input
3. `initializeLocals()` extracts the spec fields, including resolving `network_id` from `StringValueOrRef`
4. `subnet()` creates the `networking.Subnet` resource with all spec fields mapped to Pulumi args
5. Stack outputs are exported matching `stack_outputs.proto` fields

## Local Development

```bash
# Build the binary
make build

# Run a preview with the test manifest
make test
```

## Foreign Key Resolution

The `network_id` field uses the `StringValueOrRef` pattern. At runtime:
- **Literal value**: `network_id.value` is passed directly to the Pulumi resource
- **Reference**: The FK resolver middleware resolves `network_id.value_from` to the actual UUID before the module runs

In both cases, `locals.NetworkId` contains the resolved string value.
