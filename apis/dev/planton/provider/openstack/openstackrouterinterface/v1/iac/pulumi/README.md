# OpenStackRouterInterface Pulumi Module

This directory contains the Pulumi Go module for attaching OpenStack Neutron routers to subnets.

## Structure

```
iac/pulumi/
├── main.go           # Entrypoint (Pulumi program)
├── Pulumi.yaml       # Pulumi project config
├── Makefile          # Build and test targets
├── module/
│   ├── main.go             # Resources() entry point
│   ├── locals.go           # Input extraction and dual FK resolution
│   ├── router_interface.go # Router interface resource creation
│   └── outputs.go          # Output constants
└── README.md         # This file
```

## How It Works

1. The CLI loads the `OpenStackRouterInterfaceStackInput` from the stack config (base64-encoded YAML manifest)
2. `module.Resources()` is called with the stack input
3. `initializeLocals()` extracts the spec fields, resolving both `router_id` and `subnet_id` from `StringValueOrRef`
4. `routerInterface()` creates the `networking.RouterInterface` resource
5. Stack outputs are exported matching `stack_outputs.proto` fields

## Local Development

```bash
# Build the binary
make build

# Run a preview with the test manifest
make test
```

## Foreign Key Resolution

Both `router_id` and `subnet_id` use the required `StringValueOrRef` pattern. At runtime:
- **Literal value**: `.GetValue()` returns the UUID directly
- **Reference**: The FK resolver middleware resolves `value_from` to the actual UUID before the module runs

In both cases, `locals.RouterId` and `locals.SubnetId` contain the resolved string values.
