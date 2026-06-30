# OpenStackLoadBalancerPool Pulumi Module

This directory contains the Pulumi Go module for provisioning OpenStack Octavia backend pools.

## Structure

```
iac/pulumi/
├── main.go           # Entrypoint (Pulumi program)
├── Pulumi.yaml       # Pulumi project config
├── Makefile          # Build and test targets
├── module/
│   ├── main.go       # Resources() entry point
│   ├── locals.go     # Input extraction and FK resolution
│   ├── pool.go       # Pool resource creation
│   └── outputs.go    # Output constants
└── README.md         # This file
```

## How It Works

1. The CLI loads the `OpenStackLoadBalancerPoolStackInput` from the stack config (base64-encoded YAML manifest)
2. `module.Resources()` is called with the stack input
3. `initializeLocals()` extracts the spec fields, including resolving `listener_id` from `StringValueOrRef`
4. `pool()` creates the `loadbalancer.Pool` resource with all spec fields mapped to Pulumi args
5. Stack outputs are exported matching `stack_outputs.proto` fields

## Local Development

```bash
# Build the binary
make build

# Run a preview with the test manifest
make test
```

## Foreign Key Resolution

The `listener_id` field uses the `StringValueOrRef` pattern. At runtime:
- **Literal value**: `listener_id.value` is passed directly to the Pulumi resource
- **Reference**: The FK resolver middleware resolves `listener_id.value_from` to the actual UUID before the module runs

In both cases, `locals.ListenerId` contains the resolved string value.
