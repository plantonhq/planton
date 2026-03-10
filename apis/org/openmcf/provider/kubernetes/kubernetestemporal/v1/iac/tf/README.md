# Terraform Module to Deploy Temporal on Kubernetes

## Namespace Management

This module supports two namespace management modes:

### 1. Create New Namespace (Default)

```hcl
spec = {
  namespace = "temporal-prod"
  create_namespace = true  # Module creates namespace
  # ...
}
```

### 2. Use Existing Namespace

```hcl
spec = {
  namespace = "existing-namespace"  
  create_namespace = false  # Namespace must exist
  # ...
}
```

**Important**: When `create_namespace = false`, ensure the namespace exists before running `terraform apply`.

## Usage

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-temporal.tfstate"
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
