# Terraform Module to Deploy Apache Kafka on Kubernetes

This Terraform module deploys Apache Kafka on Kubernetes using the Strimzi operator, with support for Schema Registry and Kafka UI (Kowl) components.

## Namespace Management

The module provides flexible namespace management through the `create_namespace` variable:

- **Automatic Creation** (`create_namespace = true`): The module creates the namespace with appropriate labels and manages its lifecycle. This is the default and recommended approach.
- **External Management** (`create_namespace = false`): Use an existing namespace managed outside this module (e.g., via KubernetesNamespace component or manual creation). Resources will be deployed into the specified namespace without creating it.

**When to use `create_namespace = true`:**
- New deployments where namespace doesn't exist
- Full lifecycle management by the module
- Simplified deployment workflow

**When to use `create_namespace = false`:**
- Namespace is managed by a centralized governance system
- Pre-existing namespace with specific RBAC or policies
- Multi-tenant environments with namespace-level controls

## Usage

```shell
planton tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-kafka-cluster.tfstate"
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
