# Terraform Module to Deploy a Microservice on Kubernetes

## Namespace Management

This module supports flexible namespace management through the `create_namespace` variable:

- **`create_namespace = true`**: The module creates the namespace with appropriate labels. Use this for new deployments.
- **`create_namespace = false`**: The module uses an existing namespace without creating it. The namespace must already exist in the cluster. Use this when:
  - Multiple deployments share the same namespace
  - Namespaces are managed centrally
  - Using GitOps workflows where namespaces are managed separately

## Usage Commands

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-microservice-on-kuberentes.tfstate"
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
