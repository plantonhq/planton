# Terraform Module to Deploy a HelmChart

This Terraform module deploys Helm charts to a Kubernetes cluster using the KubernetesHelmRelease resource. It supports both creating new namespaces and deploying to existing namespaces based on the `create_namespace` flag.

## Usage

### Initialize Terraform

```shell
planton tofu init --manifest ../hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-helm-release.tfstate"
```

### Plan Deployment

```shell
planton tofu plan --manifest ../hack/manifest.yaml
```

### Apply Changes

```shell
planton tofu apply --manifest ../hack/manifest.yaml --auto-approve
```

### Destroy Resources

```shell
planton tofu destroy --manifest ../hack/manifest.yaml --auto-approve
```

## Module Inputs

See `variables.tf` for all available input variables.

## Module Outputs

- `namespace` - The Kubernetes namespace where the Helm release is deployed

## Examples

For comprehensive examples, see [examples.md](./examples.md).
