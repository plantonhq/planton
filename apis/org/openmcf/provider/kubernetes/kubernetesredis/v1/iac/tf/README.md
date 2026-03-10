# Terraform Module to Deploy Redis on Kubernetes

## Namespace Management

This module provides flexible namespace management through the `create_namespace` variable:

- **`create_namespace = true`**: Creates a new namespace with resource labels for tracking
  - Ideal for new deployments and isolated environments
  - Namespace is automatically created before Redis resources
  
- **`create_namespace = false`**: Uses an existing namespace
  - Namespace must exist before applying this module
  - Suitable for environments with pre-configured policies, quotas, or RBAC

## Usage

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-redis-database.tfstate"
```

```shell
openmcf tofu plan --manifest hack/manifest.yaml
```

```shell
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

```shell
openmcf tofu destroy --manifest hack/manifest.yaml --auto-approve
```
