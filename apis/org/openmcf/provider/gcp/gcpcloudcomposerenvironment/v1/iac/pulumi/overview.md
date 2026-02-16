# GcpCloudComposerEnvironment Pulumi Module Architecture

## Overview

This module provisions a Google Cloud Composer environment using the `pulumi-gcp` provider. It creates a `composer.Environment` resource with all configuration including networking, software config, workloads, CMEK, maintenance windows, recovery, and access control.

## File Organization

```
iac/pulumi/
├── main/
│   ├── main.go              # Entry point: loads stack input, calls module.Resources
│   └── Pulumi.yaml          # Project definition
└── module/
    ├── main.go              # Resources(): creates provider, calls composerEnvironment()
    ├── locals.go            # Label construction, context extraction from stack input
    ├── composer_environment.go # composer.NewEnvironment with all configuration
    └── outputs.go           # Export constants (environment_id, environment_name)
```

## Control Flow

### main.go (entry point)

- Loads `GcpCloudComposerEnvironmentStackInput` from the Pulumi context
- Invokes `module.Resources(ctx, stackInput)` to provision resources

### module/main.go

- `initializeLocals()` — Builds `Locals` with GCP labels and target resource reference
- `pulumigoogleprovider.Get()` — Configures the Google provider
- `composerEnvironment()` — Creates the Composer environment with all configuration

### locals.go

- **Label construction** — Derives GCP labels from metadata: `openmcf-resource`, `openmcf-resource-name`, `openmcf-resource-kind`, plus optional `openmcf-organization`, `openmcf-environment`, `openmcf-resource-id`
- **Context extraction** — Extracts `GcpCloudComposerEnvironment` target and provider config from stack input

### composer_environment.go

Creates `composer.NewEnvironment` with:

- **Basic config** — Project, region, environment name
- **Node config** — VPC network/subnetwork (Composer 2.x VPC peering) or PSC network attachment (Composer 3), service account, tags, internal IP CIDR block
- **Software config** — Image version, PyPI packages, Airflow config overrides, environment variables, web server plugins mode
- **Private environment config** — Private endpoint, connection type (VPC_PEERING or PRIVATE_SERVICE_CONNECT), IP ranges for Composer 2.x
- **Workloads config** — Scheduler, web server, worker (with autoscaling), triggerer, DAG processor (Composer 3) resource allocation
- **Environment size** — SMALL, MEDIUM, or LARGE
- **Resilience mode** — STANDARD_RESILIENCE or HIGH_RESILIENCE
- **CMEK** — Customer-managed encryption key
- **Maintenance window** — Start/end time, recurrence pattern
- **Recovery config** — Snapshot location, schedule, time zone
- **Web server access control** — Allowed IP ranges
- **Composer 3 flags** — `enablePrivateEnvironment`, `enablePrivateBuildsOnly`
- **Labels** — GCP labels from locals for resource organization and cost allocation

### outputs.go

Exports the following stack outputs:

- `environment_id` — Fully qualified environment resource name
- `environment_name` — Short environment name

## Resource Creation Flow

1. **Load stack input** — Read `GcpCloudComposerEnvironmentStackInput` from Pulumi context
2. **Initialize locals** — Build labels and extract context
3. **Get provider** — Configure Google provider with credentials
4. **Create environment** — Build `composer.EnvironmentArgs` from stack input:
   - Map `nodeConfig` to `composer.EnvironmentConfigNodeConfigArgs`
   - Map `softwareConfig` to `composer.EnvironmentConfigSoftwareConfigArgs`
   - Map `privateEnvironmentConfig` to `composer.EnvironmentConfigPrivateEnvironmentConfigArgs` (Composer 2.x)
   - Map `workloadsConfig` to `composer.EnvironmentConfigWorkloadsConfigArgs`
   - Map `maintenanceWindow` to `composer.EnvironmentConfigMaintenanceWindowArgs`
   - Map `recoveryConfig` to `composer.EnvironmentConfigRecoveryConfigArgs`
   - Map `webServerNetworkAccessControl` to `composer.EnvironmentConfigWebServerNetworkAccessControlArgs`
   - Set Composer 3 flags (`enablePrivateEnvironment`, `enablePrivateBuildsOnly`)
5. **Export outputs** — Export `environment_id` and `environment_name`

## Output Mapping

| Stack Output | Source | Description |
|--------------|--------|-------------|
| `environment_id` | `composer.Environment.Name` | Fully qualified resource name |
| `environment_name` | `spec.environmentName` or `metadata.name` | Short environment name |

## Networking Model Mapping

**Composer 2.x VPC Peering:**
- `nodeConfig.network` → `composer.EnvironmentConfigNodeConfigArgs.Network`
- `nodeConfig.subnetwork` → `composer.EnvironmentConfigNodeConfigArgs.Subnetwork`
- `privateEnvironmentConfig.connectionType: VPC_PEERING` → `composer.EnvironmentConfigPrivateEnvironmentConfigArgs.ConnectionType: "VPC_PEERING"`

**Composer 2.x Private Service Connect:**
- `nodeConfig.network` → `composer.EnvironmentConfigNodeConfigArgs.Network`
- `nodeConfig.subnetwork` → `composer.EnvironmentConfigNodeConfigArgs.Subnetwork`
- `privateEnvironmentConfig.connectionType: PRIVATE_SERVICE_CONNECT` → `composer.EnvironmentConfigPrivateEnvironmentConfigArgs.ConnectionType: "PRIVATE_SERVICE_CONNECT"`
- `privateEnvironmentConfig.cloudComposerConnectionSubnetwork` → `composer.EnvironmentConfigPrivateEnvironmentConfigArgs.CloudComposerConnectionSubnetwork`

**Composer 3 Private Service Connect:**
- `nodeConfig.composerNetworkAttachment` → `composer.EnvironmentConfigNodeConfigArgs.ComposerNetworkAttachment`
- `nodeConfig.composerInternalIpv4CidrBlock` → `composer.EnvironmentConfigNodeConfigArgs.ComposerInternalIpv4CidrBlock`
- `enablePrivateEnvironment` → `composer.EnvironmentConfig.EnablePrivateEnvironment`
- `enablePrivateBuildsOnly` → `composer.EnvironmentConfig.EnablePrivateBuildsOnly`

## Workload Mapping

| Component | Source Field | Pulumi Field |
|-----------|--------------|-------------|
| Scheduler | `workloadsConfig.scheduler` | `composer.EnvironmentConfigWorkloadsConfigArgs.Scheduler` |
| Web Server | `workloadsConfig.webServer` | `composer.EnvironmentConfigWorkloadsConfigArgs.WebServer` |
| Worker | `workloadsConfig.worker` | `composer.EnvironmentConfigWorkloadsConfigArgs.Worker` |
| Triggerer | `workloadsConfig.triggerer` | `composer.EnvironmentConfigWorkloadsConfigArgs.Triggerer` |
| DAG Processor | `workloadsConfig.dagProcessor` | `composer.EnvironmentConfigWorkloadsConfigArgs.DagProcessor` |

Worker autoscaling is mapped via `minCount` and `maxCount` fields.
