# AzureLinuxWebApp Pulumi Module

This directory contains the Pulumi IaC implementation for the `AzureLinuxWebApp` component.

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
    ├── main.go      # Resource creation (azurerm_linux_web_app)
    ├── locals.go    # Local variable initialization
    └── outputs.go   # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Condition |
|----------|-------------|-----------|
| Linux Web App | `appservice.LinuxWebApp` | Always |

## Prerequisites

- **Azure credentials**: `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, `ARM_TENANT_ID` (or `az login`)
- **App Service Plan**: An existing `AzureServicePlan` (the plan ARM ID is a required input)
- **Go 1.21+**: Required to build the module

## Usage with openmcf CLI

```bash
openmcf apply -f manifest.yaml
```

## Build

```bash
make build    # Compile module and entrypoint
make test     # Run module tests
make deps     # Tidy Go modules
```

## Local Development

```bash
./debug.sh                           # Uses default manifest (iac/hack/manifest.yaml)
./debug.sh path/to/manifest.yaml     # Uses custom manifest
```
