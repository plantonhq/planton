# Terraform Module to Deploy AliCloudSaeApplication

This module provisions an Alibaba Cloud SAE application with configurable compute tiers, VPC networking, health checks (liveness and readiness), rolling update strategy, environment variables, custom host aliases, and SLS log collection.

Generated `variables.tf` reflects the proto schema for `AliCloudSaeApplication`.

## Usage

Use the OpenMCF CLI (tofu) with the default local backend:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).
