# Terraform Module to Deploy Jenkins on Kubernetes

## Namespace Management

This module provides flexible namespace management through the `create_namespace` variable in the spec:

- **`create_namespace: true`** (default): The module creates a dedicated Kubernetes namespace with appropriate labels
- **`create_namespace: false`**: The module uses an existing namespace without creating it. This is useful when:
  - The namespace already exists in the cluster
  - Multiple deployments share the same namespace
  - Namespaces are managed centrally by cluster administrators
  - Using GitOps workflows where namespaces are managed separately

**Important**: When `create_namespace: false`, ensure the namespace exists before applying this module, otherwise the deployment will fail.

## Usage

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-jenkins-server.tfstate"
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
