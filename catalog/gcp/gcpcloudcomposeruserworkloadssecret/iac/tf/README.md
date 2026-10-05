# GcpCloudComposerUserWorkloadsSecret - Terraform Module

This Terraform module provisions a Kubernetes Secret in a Cloud Composer environment's workloads namespace (`google_composer_user_workloads_secret`). It is the Terraform-side implementation of the Planton `GcpCloudComposerUserWorkloadsSecret` resource kind and has feature parity with the Pulumi module.

## Overview

The Secret's data updates in place; name, environment, region, and project are immutable (ForceNew). Values are base64-encoded secret material (the Kubernetes Secret contract). The provider marks the `data` attribute sensitive — plans redact it — and it is never surfaced in outputs; IaC state is the engine's secret boundary. No API enablement here: the Composer API is enabled by the environment this Secret is delivered into. An empty `project_id` falls back to the provider's default project.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpcloudcomposeruserworkloadssecret/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpCloudComposerUserWorkloadsSecret spec | — |

The `spec` object includes: `project_id` (optional — provider default when empty), `region`, `environment` (resolved ref to the environment name), `secret_name`, and `data` (min one entry, base64 values).

## Outputs

| Name | Description |
|------|-------------|
| `name` | Fully qualified resource name (`projects/{p}/locations/{r}/environments/{e}/userWorkloadsSecrets/{n}`) |
| `secret_name` | The Kubernetes Secret name — what DAGs reference |

The Secret's data is deliberately never exported.
