# AzureContainerApp Pulumi Module

This directory contains the Pulumi IaC implementation for the `AzureContainerApp` component.

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
    ├── main.go      # Resource creation (containerapp.App) + builder functions
    ├── locals.go    # Local variable initialization
    └── outputs.go   # Output key constants
```

## Resources Created

| Resource | Pulumi Type | Condition |
|----------|-------------|-----------|
| Container App | `containerapp.App` | Always |

This is a single-resource module. The complexity lives in the builder functions that translate the rich protobuf spec (21 message types) into Pulumi resource arguments: containers, init containers, probes, scale rules, ingress, secrets, registries, Dapr, identity, and volumes.

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

## Builder Functions

The `module/main.go` file contains builder functions that map spec messages to Pulumi args:

| Function | Spec Message | Pulumi Type |
|----------|-------------|-------------|
| `buildContainers` | `AzureContainerAppContainer` | `AppTemplateContainerArgs` |
| `buildInitContainers` | `AzureContainerAppInitContainer` | `AppTemplateInitContainerArgs` |
| `buildEnvVars` | `AzureContainerAppEnvVar` | `AppTemplateContainerEnvArgs` |
| `buildLivenessProbe` | `AzureContainerAppProbe` | `AppTemplateContainerLivenessProbeArgs` |
| `buildReadinessProbe` | `AzureContainerAppProbe` | `AppTemplateContainerReadinessProbeArgs` |
| `buildStartupProbe` | `AzureContainerAppProbe` | `AppTemplateContainerStartupProbeArgs` |
| `buildVolumes` | `AzureContainerAppVolume` | `AppTemplateVolumeArgs` |
| `buildHttpScaleRules` | `AzureContainerAppHttpScaleRule` | `AppTemplateHttpScaleRuleArgs` |
| `buildTcpScaleRules` | `AzureContainerAppTcpScaleRule` | `AppTemplateTcpScaleRuleArgs` |
| `buildAzureQueueScaleRules` | `AzureContainerAppAzureQueueScaleRule` | `AppTemplateAzureQueueScaleRuleArgs` |
| `buildCustomScaleRules` | `AzureContainerAppCustomScaleRule` | `AppTemplateCustomScaleRuleArgs` |
| `buildSecrets` | `AzureContainerAppSecret` | `AppSecretArgs` |
| `buildRegistries` | `AzureContainerAppRegistry` | `AppRegistryArgs` |
| `buildIngress` | `AzureContainerAppIngress` | `AppIngressArgs` |
| `buildDapr` | `AzureContainerAppDapr` | `AppDaprArgs` |
| `buildIdentity` | `AzureContainerAppIdentity` | `AppIdentityArgs` |
