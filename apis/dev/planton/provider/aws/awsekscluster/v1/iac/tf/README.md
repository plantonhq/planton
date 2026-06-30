# Terraform Module to Deploy AwsEksCluster

This module provisions an AWS EKS (Elastic Kubernetes Service) cluster with support for high availability, security, and monitoring.
It includes optional KMS encryption, control plane logging, and configurable network access settings.

Generated `variables.tf` reflects the proto schema for `AwsEksCluster`.

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
