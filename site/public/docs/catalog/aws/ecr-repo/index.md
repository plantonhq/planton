---
title: "ECR Repo"
description: "ECR Repo deployment documentation"
icon: "package"
order: 100
componentName: "awsecrrepo"
---

# AWS ECR Repo

Deploys an AWS Elastic Container Registry repository with configurable tag immutability, image scanning, encryption, and optional lifecycle policies for automated image expiration. The component applies OpenMCF resource tags to the repository for traceability.

## What Gets Created

When you deploy an AwsEcrRepo resource, OpenMCF provisions:

- **ECR Repository** — an `ecr.Repository` with the specified name, tag mutability setting, image scanning configuration, encryption configuration, force-delete behavior, and OpenMCF resource tags
- **Lifecycle Policy** (optional) — an `ecr.LifecyclePolicy` attached to the repository, containing up to two rules: one to expire untagged images after a specified number of days, and one to retain only the most recent N images

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A KMS key ARN** if using `KMS` encryption (optional; the default `AES256` encryption requires no additional setup)

## Quick Start

Create a file `ecr-repo.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: my-service
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcrRepo.my-service
spec:
  region: us-east-1
  repositoryName: my-org/my-service
```

Deploy:

```shell
openmcf apply -f ecr-repo.yaml
```

This creates an ECR repository with mutable tags, AES256 encryption, and scan-on-push enabled (the defaults).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the ECR repository will be created. | Valid AWS region |
| `repositoryName` | `string` | Name of the ECR repository. Must be unique within the AWS account and region. Commonly includes the organization or project prefix, e.g., `team-blue/my-microservice`. | 2–256 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `imageImmutable` | `bool` | `false` | When `true`, image tags cannot be overwritten (tag mutability set to `IMMUTABLE`). When `false`, tags are `MUTABLE` and can be overwritten by subsequent pushes. |
| `encryptionType` | `string` | `"AES256"` | How images are encrypted at rest. Valid values: `AES256` (AWS-managed keys), `KMS` (customer-managed key). |
| `kmsKeyId` | `string` | `""` | ARN or ID of a KMS key for encryption. Only used when `encryptionType` is `KMS`. Ignored when `encryptionType` is `AES256`. |
| `forceDelete` | `bool` | `false` | When `true`, allows deleting the repository even when it contains images. All images are removed on delete. When `false`, deletion fails if images exist. |
| `scanOnPush` | `bool` | `true` | Enables automatic vulnerability scanning when images are pushed. Recommended for production environments. |
| `lifecyclePolicy` | `object` | — | Lifecycle rules for automated image expiration. If omitted, no lifecycle policy is created. |
| `lifecyclePolicy.expireUntaggedAfterDays` | `int32` | `14` | Removes untagged images after the specified number of days. Untagged images are typically intermediate build layers or failed builds. Range: 1–365. |
| `lifecyclePolicy.maxImageCount` | `int32` | `30` | Keeps only the most recent N images, expiring all older ones. Prevents unbounded storage growth from CI/CD pipelines. Range: 1–1000. |

## Examples

### Immutable Tags for Production

A repository where tags cannot be overwritten, preventing accidental overwrites of released images:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: prod-api
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcrRepo.prod-api
spec:
  region: us-east-1
  repositoryName: my-org/api-server
  imageImmutable: true
```

### KMS Encryption

A repository using a customer-managed KMS key for compliance requirements:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: compliant-repo
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcrRepo.compliant-repo
spec:
  region: us-east-1
  repositoryName: my-org/compliant-service
  encryptionType: KMS
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcd-1234-efgh-5678
  imageImmutable: true
```

### Lifecycle Policy for Cost Control

A repository with lifecycle rules to expire untagged images after 7 days and keep only the last 50 tagged images:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: ci-images
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcrRepo.ci-images
spec:
  region: us-east-1
  repositoryName: my-org/ci-runner
  lifecyclePolicy:
    expireUntaggedAfterDays: 7
    maxImageCount: 50
```

### Full Production Configuration

A repository with immutable tags, KMS encryption, scan-on-push, lifecycle management, and force-delete disabled:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: prod-frontend
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsEcrRepo.prod-frontend
spec:
  region: us-east-1
  repositoryName: my-org/frontend
  imageImmutable: true
  encryptionType: KMS
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/prod-key-id
  scanOnPush: true
  forceDelete: false
  lifecyclePolicy:
    expireUntaggedAfterDays: 3
    maxImageCount: 100
```

### Development Repository with Force Delete

A disposable development repository that can be torn down even with images present:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsEcrRepo
metadata:
  name: dev-scratch
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsEcrRepo.dev-scratch
spec:
  region: us-west-2
  repositoryName: my-org/scratch
  forceDelete: true
  lifecyclePolicy:
    expireUntaggedAfterDays: 1
    maxImageCount: 10
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `repository_name` | `string` | The repository name, matching `spec.repositoryName` |
| `repository_url` | `string` | The repository URL for docker push/pull (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com/my-repo`) |
| `repository_arn` | `string` | The repository ARN (e.g., `arn:aws:ecr:us-east-1:123456789012:repository/my-repo`) |
| `registry_id` | `string` | The registry ID (AWS account ID) associated with the repository |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides a customer-managed KMS key for repository encryption
- [AwsIamRole](/docs/catalog/aws/iam-role) — grants push/pull permissions to CI/CD pipelines or services
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — Kubernetes cluster that pulls images from ECR repositories
