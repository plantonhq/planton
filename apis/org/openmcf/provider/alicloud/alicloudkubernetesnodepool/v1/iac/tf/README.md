# Terraform Module to Deploy AlicloudKubernetesNodePool

This module provisions an ACK Kubernetes node pool with configurable instance types, disk configuration, auto-scaling, managed lifecycle, spot instances, and Kubernetes scheduling properties.

Generated `variables.tf` reflects the proto schema for `AlicloudKubernetesNodePool`.

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
