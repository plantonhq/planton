# Terraform Module to Deploy Apache Solr on Kubernetes

## Quick Start

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-solr-cloud.tfstate"
```

```shell
planton tofu plan --manifest hack/manifest.yaml
```

```shell
planton tofu apply --manifest hack/manifest.yaml --auto-approve
```

## Namespace Management

This module provides flexible namespace management through the `spec.create_namespace` variable:

### create_namespace = true (default)

When set to `true`, the module creates the namespace and manages its lifecycle. The namespace will be destroyed when running `terraform destroy`.

### create_namespace = false

When set to `false`, the module uses an existing namespace. This is useful for:
- Shared namespaces across multiple components
- Centralized namespace management
- GitOps-based namespace provisioning

**Important**: Ensure the namespace exists before running `terraform apply` when using `create_namespace = false`.

## Examples

For detailed configuration examples and best practices, see [examples.md](./examples.md).
