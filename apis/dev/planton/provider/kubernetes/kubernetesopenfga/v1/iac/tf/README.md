# Terraform Module to Deploy OpenFGA on Kubernetes

## Overview

This Terraform module deploys OpenFGA on a Kubernetes cluster with configurable namespace management, container resources, ingress settings, and datastore configuration.

## Namespace Management

This Terraform module supports flexible namespace management through the `create_namespace` variable:

- **When `create_namespace` is `true`**: The module creates a new Kubernetes namespace with the specified name and labels
- **When `create_namespace` is `false`**: The module expects the namespace to already exist and will reference it using a data source

This flexibility allows you to:
- Manage namespaces independently for better organization
- Share namespaces across multiple components
- Comply with organizational policies that require centralized namespace management
- Support both greenfield and brownfield deployments

## Usage Examples

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-open-fga-server.tfstate"
```

```shell
planton tofu plan --manifest hack/manifest.yaml
```

```shell
planton tofu apply --manifest hack/manifest.yaml --auto-approve
```

```shell
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```
