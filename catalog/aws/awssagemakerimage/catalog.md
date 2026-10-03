# AWS SageMaker Image

Deploys an Amazon SageMaker image — the named registry entry that makes your own container images selectable as custom kernels and training environments in SageMaker Studio — together with its folded versions, each pointing at a concrete ECR image. AWS numbers versions sequentially and never reuses a number, so each entry's position in `versions` is its stable identity: append at the end, never reorder. Aliases (`latest`, `stable`) move freely between versions, and the compatibility metadata updates in place, while a changed base image retires the old number and mints a new one.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **SageMaker Image** — the registry entry named from `metadata.name`, carrying the ECR-pull role and the in-place Studio display metadata (`displayName`, `description`)
- **SageMaker Image Versions** — one per `versions` entry, each registering an ECR image under an AWS-assigned sequential number, with movable aliases and compatibility metadata (`jobType`, `mlFramework`, `processor`, `programmingLang`, `vendorGuidance`)

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with SageMaker control-plane permissions (`sagemaker:CreateImage`, `sagemaker:CreateImageVersion`, and their siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- An IAM role trusting `sagemaker.amazonaws.com` with pull access to the version images, wired via `roleArn`.
- For each version: the container image pushed to ECR in the SAME account and region as the image — cross-account registries need a replication step first.

## Deploy

### Console

Open the deployment store, find **AWS SageMaker Image**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, the region and ECR-pull role, display metadata, and the version list. Start from the **Custom Kernel Image** preset in the [Presets](#presets) tab for a Studio kernel registry entry, or the **Versioned Training Image** preset for an annotated GPU training environment.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerImage
metadata:
  name: team-kernels
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sagemaker-image-role
      fieldPath: status.outputs.role_arn
  displayName: Team Custom Kernels
  description: Custom kernel images for Studio notebooks
  versions:
    - baseImage: 123456789012.dkr.ecr.us-east-1.amazonaws.com/team-kernels:pytorch-2.4
      aliases:
        - latest
      jobType: NOTEBOOK_KERNEL
      mlFramework: PyTorch 2.4
      processor: CPU
      programmingLang: python 3.12
```

```shell
planton apply -f sagemaker-image.yaml
```

This creates the registry entry and registers the ECR image as version 1, aliased `latest` and annotated as a CPU PyTorch notebook kernel. An Infra Job tracks the provisioning in real time.

### InfraChart

When the image deploys alongside its ECR-pull role in one chart, wire the role reference via ValueFromRef:

```yaml
spec:
  region: us-east-1
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: sagemaker-image-role
      fieldPath: status.outputs.role_arn
  displayName: Team Custom Kernels
```

The InfraPipeline resolves the dependency graph, creates the role first, then the image that assumes it.

## Key Configuration

These are the most important decisions when configuring a SageMaker image. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Append-only versions, keyed by position** — AWS assigns version numbers sequentially and never reuses them, so no spec field can be a version's stable identity; the entry's POSITION in `versions` is. Add new versions at the end and never reorder existing entries — a reorder makes the modules see every shifted entry as a replacement.

**`baseImage` is the version's identity** — Changing it replaces the version under a NEW AWS-assigned number; the old number stays retired forever. The compatibility metadata (`jobType`, `mlFramework`, `processor`, `programmingLang`, `vendorGuidance`) updates in place, so annotate freely after the fact.

**Let aliases carry meaning, not numbers** — Aliases move freely between versions and update in place. Point consumers at `stable` and promote a new version by moving the alias, instead of chasing AWS-assigned numbers through job configurations.

**The role is a one-way door** — Changing `roleArn` replaces the image (and with it, every version). Get the ECR-pull role right before registering versions consumers depend on.

**Budget a minute per create, and serialize version pushes** — The provider waits about a minute before CreateImage for IAM propagation, and AWS serializes version creation per image — ten new versions in one apply land one after another, not in parallel.

**Versions carry no tags upstream** — By provider design, only the image itself is taggable. Don't plan cost allocation per version.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `image_name` | The image's AWS identity | Attaching the image to a SageMaker domain or space so Studio users can select it |
| `image_arn` | Amazon Resource Name of the image | Domain and space custom-image configuration; IAM policies scoping image access |

`version_numbers` is also exported — a map from each `versions` entry's position ("0", "1", …) to its AWS-assigned number. It is bookkeeping for the modules' satellite keying, not a composition input: consumers should reference versions by alias instead.

## Common Patterns

**Custom Studio kernel** — the registry entry plus one `NOTEBOOK_KERNEL` version per kernel image, aliased `latest`. Attach the image to a SageMaker domain to make it selectable in Studio; data scientists pick it by display name. Start from the **Custom Kernel Image** preset.

**Versioned training environment** — a `TRAINING` version per release of a GPU training image, fully annotated with framework, processor, and language so consumers can verify compatibility before a job runs. Promote releases by moving the `stable` alias. Start from the **Versioned Training Image** preset.

**Registry entry first, versions later** — `versions` is optional, so the image can exist before any container ships. Useful when the platform team owns the entry and product teams append versions as they publish images — the append-only contract keeps their entries from colliding.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the ECR-pull role SageMaker assumes, wired via `roleArn`
- [**AWS ECR Repository**](/infra-catalog/aws-ecr-repo) — where the version images live, referenced by registry path in `baseImage`
- [**AWS SageMaker Domain**](/infra-catalog/aws-sagemaker-domain) — the Studio domain the image attaches to for kernel selection
