# Terraform Module to Deploy AliCloudSaeApplication

This module provisions an Alibaba Cloud SAE application with configurable compute tiers, VPC networking, health checks (liveness and readiness), rolling update strategy, environment variables, custom host aliases, and SLS log collection.

Generated `variables.tf` reflects the proto schema for `AliCloudSaeApplication`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via stack input (CLI), not in the manifest `spec`.

For more examples, see [`examples.md`](./examples.md) and [`hack/manifest.yaml`](../hack/manifest.yaml).
