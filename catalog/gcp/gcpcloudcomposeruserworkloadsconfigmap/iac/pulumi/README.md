# GcpCloudComposerUserWorkloadsConfigMap - Pulumi Module

This Pulumi (Go) module provisions a Kubernetes ConfigMap in a Cloud Composer environment's workloads namespace (`composer.UserWorkloadsConfigMap`). It is the Pulumi-side implementation of the Planton `GcpCloudComposerUserWorkloadsConfigMap` resource kind and has feature parity with the Terraform module.

## Overview

The ConfigMap's data updates in place; name, environment, region, and project are immutable. Data is plain configuration — use the user workloads Secret kind for secret material. No API enablement here: the Composer API is enabled by the environment this ConfigMap is delivered into. An empty `project_id` falls back to the provider's default project.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpcloudcomposeruserworkloadsconfigmap/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — resolved IaC input values
- `module/user_workloads_config_map.go` — the ConfigMap resource
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `name` | Fully qualified resource name (`projects/{p}/locations/{r}/environments/{e}/userWorkloadsConfigMaps/{n}`) |
| `config_map_name` | The Kubernetes ConfigMap name — what DAGs reference |
