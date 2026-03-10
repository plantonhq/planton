# Terraform Module to Deploy Elasticsearch on Kubernetes

## Usage

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-elasticsearch-cluster.tfstate"
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

## Namespace Management

The module provides flexible namespace management through the `spec.create_namespace` variable:

- **`create_namespace: true`** (recommended): The module will create the namespace with proper labels and configuration. Use this when you want the component to manage the full namespace lifecycle.

- **`create_namespace: false`**: The module will use an existing namespace. The namespace must already exist in the cluster. Use this when:
  - The namespace is managed separately (e.g., by another component or tool)
  - You're deploying multiple resources into a shared namespace
  - Namespace policies are managed centrally

**Important**: When `create_namespace: false`, ensure the namespace exists before deploying this component, or the deployment will fail.
