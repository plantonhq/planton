# AWS ECR Repository

Deploys an Elastic Container Registry repository — the private Docker/OCI image registry that container workloads (ECS, EKS, Lambda container images, App Runner) pull from. The spec covers the full repository surface: tag mutability including the exclusion-filter modes (freeze releases while `latest` floats), encryption at rest (AES-256, KMS, or dual-layer KMS_DSSE), push-time vulnerability scanning, structured lifecycle rules for storage cost control, and a folded repository policy for cross-account or cross-service pull access.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ECR Repository** -- a Docker image registry with the specified name, tag mutability setting (and exclusion filters), encryption configuration, and scan-on-push behavior
- **Lifecycle Policy** -- created only when `lifecycleRules` are configured; each rule selects images by tag state (untagged, tagged by prefix or pattern, or every image) and either expires them (keep-last-N count, age since push, or days since last pull) or transitions them to the cheaper archive storage class; archived images get their own retention clock (`sinceImageTransitioned`)
- **Repository Policy** -- created only when `repositoryPolicy` is configured; the resource-based IAM policy for cross-account pulls and service access (e.g. Lambda container images)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) -- required only when using KMS encryption (`encryptionType: KMS`). Provide the key ARN directly or reference an AwsKmsKey Infra Component via ValueFromRef. If omitted, ECR uses AWS-managed AES-256 encryption.

## Deploy

### Console

Open the deployment store, find **AWS ECR Repository**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production Immutable ECR Repository** preset in the [Presets](#presets) tab to pre-populate a secure default configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcrRepo
metadata:
  name: api-service
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  repositoryName: "acme-corp/api-service"
  imageTagMutability: IMMUTABLE
  scanOnPush: true
  forceDelete: false
  lifecycleRules:
    - rulePriority: 10
      description: Expire untagged images after 7 days
      tagStatus: untagged
      countType: sinceImagePushed
      countNumber: 7
    - rulePriority: 20
      description: Keep the newest 100 images
      tagStatus: any
      countType: imageCountMoreThan
      countNumber: 100
```

```shell
planton apply -f ecr-repo.yaml
```

This creates an ECR repository with immutable tags, scan-on-push enabled, AWS-default AES-256 encryption, and lifecycle rules that expire untagged images after 7 days and retain the 100 most recent images. An Infra Job tracks the provisioning in real time.

### InfraChart

When using KMS encryption, use ValueFromRef to wire the repository to a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  encryptionType: KMS
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: ecr-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key first, then provisions the ECR repository with KMS encryption using the resolved key ARN.

## Key Configuration

These are the most important decisions when configuring an ECR repository. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Image tag mutability** -- Set `imageTagMutability: IMMUTABLE` for production registries to guarantee that a tag like `v1.2.3` always refers to the same image. The exclusion modes give you both worlds: `IMMUTABLE_WITH_EXCLUSION` freezes everything except tags matching `imageTagMutabilityExclusionFilters` (float `latest`, freeze releases); `MUTABLE_WITH_EXCLUSION` inverts it. Unset keeps the AWS default (mutable).

**Encryption type** -- Unset keeps the AWS default (`AES256`). Set `encryptionType: KMS` with a `kmsKeyId` when you need CloudTrail audit logging of key usage or customer-controlled rotation; `KMS_DSSE` adds a second independent envelope layer for DoD CC SRG-class requirements. Encryption is a create-time choice — changing it replaces the repository, and images are not migrated.

**Lifecycle rules** -- Structured `lifecycleRules[]` evaluated daily by ascending priority; an image is acted on by the FIRST rule that selects it. The starter pair: expire untagged images by age, and keep the newest N per release prefix. The cost-tiering pattern: `actionType: transition` + `targetStorageClass: archive` moves images nobody pulls to the cheaper archive tier (`sinceImagePulled`), and a `storageClass: archive` + `sinceImageTransitioned` rule expires them after their time in archive. At most one `tagStatus: any` rule per storage-class tier.

**Repository policy** -- A standard IAM policy document for access beyond this account: cross-account pulls (grant `ecr:BatchGetImage` / `ecr:GetDownloadUrlForLayer`) and AWS services such as Lambda pulling container images. Leave unset for same-account access.

**Scan on push** -- The AWS default is on. Automatic vulnerability scanning via Amazon Inspector runs each time an image is pushed. Keep enabled in all environments for shift-left security.

**Force delete** -- Defaults to `false`, preventing repository deletion while it contains images. Set to `true` only for ephemeral or CI/CD test registries.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `repository_name` | ECR repository name | CodeBuild environment variables, CodePipeline ECR source configuration |
| `repository_url` | Full ECR repository URL | Docker push/pull commands, Kubernetes image references, Lambda container image URI |
| `repository_arn` | Repository Amazon Resource Name | IAM policies granting push/pull access |
| `registry_id` | AWS account ID hosting the registry | Docker login authentication, cross-account access configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production immutable registry** -- Immutable tags, scan-on-push, 7-day untagged expiration, 100-image retention. Ensures tag integrity for production deployments and provides rollback capability across recent releases. Start from the **Production Immutable ECR Repository** preset.

**Development registry** -- Mutable tags, scan-on-push, 3-day untagged expiration, 20-image retention. Allows rapid iteration with tag overwriting and aggressive cleanup to minimize storage costs. Start from the **Development ECR Repository** preset.

**Archive cost tiering** -- transition rules move images nobody pulls to the archive storage class, and an archive-scoped rule expires them after their time there — storage cost control for registries that accumulate history. Start from the **Archive Cost-Tiering ECR Repository** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for KMS encryption of stored images
- [**AWS ECS Task Definition**](/infra-catalog/aws-ecs-task-definition) -- container images reference the repository URL textually (`<repository_url>:<tag>`)
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- container-image functions pull from ECR (grant `lambda.amazonaws.com` in the repository policy)
- [**AWS App Runner Service**](/infra-catalog/aws-app-runner-service) -- ECR-sourced services deploy images pushed here