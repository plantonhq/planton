# AzureServicePlan Pulumi Module

This directory contains the Pulumi IaC implementation for the `AzureServicePlan` kind.

## Structure

```
pulumi/
├── main.go          # Entrypoint (loads IaC input, calls module)
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build/test targets
├── README.md        # This file
└── module/
    ├── main.go      # Resource creation (appservice.ServicePlan)
    ├── locals.go    # Locals + enum-to-wire-value maps
    └── outputs.go   # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Condition |
|----------|-------------|-----------|
| Service Plan | `appservice.ServicePlan` | Always |

## Build

```bash
make build    # Compile module and entrypoint
make test     # Run module tests
make deps     # Tidy Go modules
```

## Debug

```bash
./debug.sh                           # Uses default manifest
./debug.sh path/to/manifest.yaml     # Uses custom manifest
```
