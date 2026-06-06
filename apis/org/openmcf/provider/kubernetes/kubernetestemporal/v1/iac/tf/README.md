# Terraform Module to Deploy Temporal on Kubernetes

## Namespace Management

This module supports two namespace management modes:

### 1. Use Existing Namespace (Default)

```hcl
spec = {
  namespace = "existing-namespace"
  # create_namespace defaults to false  # Namespace must already exist
  # ...
}
```

### 2. Create New Namespace

```hcl
spec = {
  namespace = "temporal-prod"
  create_namespace = true  # Module creates namespace
  # ...
}
```

**Important**: `create_namespace` defaults to `false` (matching the proto3 zero value), so the namespace must exist before running `terraform apply` unless you set `create_namespace = true`.

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
