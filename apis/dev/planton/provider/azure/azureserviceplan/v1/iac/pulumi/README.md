# AzureServicePlan Pulumi Module

This directory contains the Pulumi IaC implementation for the `AzureServicePlan` component.

## Structure

```
pulumi/
├── main.go          # Entrypoint (loads stack input, calls module)
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build/test targets
├── debug.sh         # Debug build script
├── README.md        # This file
├── overview.md      # Architecture overview
└── module/
    ├── main.go      # Resource creation (azurerm_service_plan)
    ├── locals.go    # Local variable initialization
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
