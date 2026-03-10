# Terraform Module to Deploy a CronJob on Kubernetes

## Features

- **Conditional Namespace Management**: Control whether to create a new namespace or use an existing one via the `create_namespace` flag
- **Automated Resource Creation**: Creates CronJob, ServiceAccount, and Secrets based on the specification
- **Private Registry Support**: Automatic image pull secret creation when Docker credentials are provided
- **Flexible Scheduling**: Support for all standard cron expressions and concurrency policies
- **Resource Management**: Configure CPU/memory requests and limits for optimal cluster utilization

## Namespace Management

The module supports two modes for namespace management:

### Create New Namespace (create_namespace = true)

When `create_namespace` is set to `true`, the module will:
- Create a new Kubernetes namespace with the specified name
- Apply appropriate labels to the namespace for resource tracking
- Manage the namespace lifecycle as part of the Terraform state

**Use this when:**
- You want dedicated isolation for the CronJob
- The namespace doesn't exist yet
- You want Terraform to manage the namespace lifecycle

### Use Existing Namespace (create_namespace = false)

When `create_namespace` is set to `false`, the module will:
- Reference an existing Kubernetes namespace
- Deploy the CronJob resources into that namespace
- Not manage the namespace lifecycle (namespace must exist beforehand)

**Use this when:**
- Multiple CronJobs share the same namespace
- The namespace is managed by a separate Terraform module or process
- You're following a GitOps pattern where namespaces are pre-created

## Usage

```shell
openmcf tofu init --manifest hack/manifest.yaml --backend-type s3 \
  --backend-config="bucket=planton-tf-state-backend" \
  --backend-config="dynamodb_table=planton-tf-state-backend-lock" \
  --backend-config="region=ap-south-2" \
  --backend-config="key=kubernetes-stacks/test-microservice-on-kuberentes.tfstate"
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
