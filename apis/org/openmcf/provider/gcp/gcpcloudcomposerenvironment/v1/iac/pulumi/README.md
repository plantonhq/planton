# GcpCloudComposerEnvironment Pulumi Module

This Pulumi module provisions a Google Cloud Composer environment — a managed Apache Airflow service for authoring, scheduling, and monitoring data pipelines.

## Architecture

The module creates a single GCP resource:

1. **composer.Environment** — The Composer environment with networking, software configuration, workloads, CMEK encryption, maintenance windows, recovery, and access control

The environment resource includes all configuration inline (Composer's API bundles all settings within the environment resource). The module uses the `pulumi-gcp` (Google) provider with a **local backend** for state storage.

## Prerequisites

- GCP project with Cloud Composer API enabled
- Cloud KMS keys (if using CMEK encryption) with appropriate IAM permissions for the Composer service account
- VPC network and subnetwork (if using VPC peering networking)
- PSC network attachment (if using Composer 3 PSC networking)

## Structure

```
iac/pulumi/
├── main/
│   ├── main.go              # Entry point: loads stack input, calls module.Resources
│   └── Pulumi.yaml          # Project definition
└── module/
    ├── main.go              # Resources(): orchestrates provider and composerEnvironment
    ├── locals.go            # Label construction, context extraction from stack input
    ├── composer_environment.go # composer.NewEnvironment with all configuration
    └── outputs.go           # Export constants (environment_id, environment_name)
```

## Outputs

| Name | Description |
|------|-------------|
| `environment_id` | Fully qualified environment resource name (`projects/{project}/locations/{region}/environments/{name}`) |
| `environment_name` | Short environment name (same as `environmentName` input or `metadata.name`) |

## Running the Module

```bash
cd iac/pulumi/main
pulumi up
```

## Debugging

See the [overview.md](overview.md) for architecture details and the debug script reference in the module directory.
