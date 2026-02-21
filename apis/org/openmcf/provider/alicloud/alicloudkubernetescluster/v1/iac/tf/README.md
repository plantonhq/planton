# Terraform Module to Deploy AliCloudKubernetesCluster

This module provisions an Alibaba Cloud ACK Managed Kubernetes cluster with configurable networking (Flannel or Terway CNI), addons, control plane logging, maintenance windows, and automatic version upgrades.

Generated `variables.tf` reflects the proto schema for `AliCloudKubernetesCluster`.

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
