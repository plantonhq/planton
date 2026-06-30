---
title: "Deploy Your First AWS Resource"
description: "Deploy an S3 bucket to AWS with Planton — from manifest to deployed resource in minutes"
icon: "tutorial"
order: 20
---

# Deploy Your First AWS Resource

In this tutorial, you will deploy an S3 bucket to AWS using Planton. You will write a manifest, preview the deployment plan, apply it, modify the bucket configuration, and tear it down — experiencing the full lifecycle of an Planton-managed resource.

By the end, you will have a working understanding of how Planton deploys cloud resources and how the manifest-driven workflow operates end to end.

## What You Will Build

An S3 bucket with:

- Server-side encryption (SSE-S3)
- Versioning enabled for object protection
- Tags for resource governance
- A lifecycle rule to transition old objects to cheaper storage

## Prerequisites

Before starting, ensure you have:

- **Planton CLI** installed (`planton version` should print a version). See [Getting Started](../getting-started) for installation.
- **AWS credentials** configured. Planton needs permission to create S3 buckets in your AWS account. See [AWS Provider Setup](../guides/aws-provider-setup) for detailed instructions.
- **Pulumi CLI** installed (`brew install pulumi`) with a backend configured (`pulumi login --local` for local state), **or** **OpenTofu CLI** installed (`brew install opentofu`). This tutorial uses Pulumi, but you can substitute OpenTofu by changing the provisioner label.

## Step 1: Write the Manifest

Create a file named `s3-bucket.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3Bucket
metadata:
  name: my-first-bucket
  labels:
    planton.dev/provisioner: pulumi
spec:
  awsRegion: us-east-1
  versioningEnabled: true
  encryptionType: ENCRYPTION_TYPE_SSE_S3
  tags:
    environment: tutorial
    managed-by: planton
```

Every Planton manifest follows the [Kubernetes Resource Model](../concepts/manifests): `apiVersion`, `kind`, `metadata`, and `spec`. The `spec` fields are defined by the component's Protocol Buffer schema — in this case, `AwsS3BucketSpec`.

Here is what each field does:

| Field | Purpose |
|-------|---------|
| `apiVersion` | Identifies the provider and API version (`aws.planton.dev/v1`) |
| `kind` | The deployment component type (`AwsS3Bucket`) |
| `metadata.name` | A unique name for this resource instance |
| `metadata.labels` | The `planton.dev/provisioner` label tells Planton which IaC engine to use |
| `spec.awsRegion` | AWS region where the bucket will be created (required) |
| `spec.versioningEnabled` | Keeps all versions of objects, protecting against accidental deletes |
| `spec.encryptionType` | Server-side encryption method. `ENCRYPTION_TYPE_SSE_S3` uses AWS-managed AES-256 keys |
| `spec.tags` | Key-value pairs applied to the AWS resource for cost tracking and governance |

## Step 2: Preview the Deployment

Before deploying, preview what Planton will create:

```bash
planton plan -f s3-bucket.yaml
```

Planton reads the manifest, resolves the `AwsS3Bucket` deployment component module, and delegates to Pulumi to generate an execution plan. You will see output describing the resources that will be created — an S3 bucket with the configuration you specified.

Review the plan to confirm it matches your expectations before proceeding.

## Step 3: Deploy

Apply the manifest to create the bucket:

```bash
planton apply -f s3-bucket.yaml
```

Planton performs the same steps as `plan`, then executes the deployment. Pulumi provisions the S3 bucket in your AWS account with versioning, encryption, and tags configured.

The deployment outputs include:

| Output | Description |
|--------|-------------|
| `bucket_id` | The name of the S3 bucket created on AWS |
| `bucket_arn` | The ARN, used in IAM policies and cross-account access |
| `region` | The AWS region where the bucket was created |
| `bucket_regional_domain_name` | The regional endpoint for accessing the bucket |

## Step 4: Verify

Confirm the bucket exists using the AWS CLI:

```bash
aws s3 ls | grep my-first-bucket
```

Check that versioning is enabled:

```bash
aws s3api get-bucket-versioning --bucket <bucket_id from outputs>
```

You should see `"Status": "Enabled"`.

Check encryption:

```bash
aws s3api get-bucket-encryption --bucket <bucket_id from outputs>
```

The output should show `AES256` as the SSE algorithm.

## Step 5: Modify the Resource

One of Planton's strengths is idempotent updates. You can modify your manifest and re-apply — Planton will compute the diff and apply only the changes.

Add a lifecycle rule that transitions objects older than 30 days to Infrequent Access storage. Update `s3-bucket.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3Bucket
metadata:
  name: my-first-bucket
  labels:
    planton.dev/provisioner: pulumi
spec:
  awsRegion: us-east-1
  versioningEnabled: true
  encryptionType: ENCRYPTION_TYPE_SSE_S3
  tags:
    environment: tutorial
    managed-by: planton
  lifecycleRules:
    - id: move-to-ia
      enabled: true
      prefix: ""
      transitionDays: 30
      transitionStorageClass: STORAGE_CLASS_STANDARD_IA
      abortIncompleteMultipartUploadDays: 7
```

Preview the change:

```bash
planton plan -f s3-bucket.yaml
```

The plan will show that the bucket is being updated (not replaced) — only the lifecycle rule is being added. Apply it:

```bash
planton apply -f s3-bucket.yaml
```

This demonstrates the declarative workflow: you describe the desired state, and Planton computes and applies the delta.

## Step 6: Clean Up

Destroy the resource when you are done:

```bash
planton destroy -f s3-bucket.yaml
```

Planton reads the manifest, identifies the managed resources, and removes them from AWS. The bucket and its configuration are deleted.

## What You Learned

- How to write an Planton manifest for an AWS resource, with fields defined by the component's Protocol Buffer schema
- The `plan` -> `apply` -> `destroy` lifecycle that applies to every Planton deployment
- How to modify a deployed resource by updating the manifest and re-applying
- How manifest labels (`planton.dev/provisioner`) control which IaC engine Planton uses

## What's Next

- [Deploy Your First Kubernetes Resource](./first-kubernetes-resource) — deploy PostgreSQL on Kubernetes with custom databases and users
- [Writing Manifests](../guides/manifests) — practical guide to writing manifests for any component
- [Deployment Components](../concepts/deployment-components) — understand the anatomy of the component you just deployed
- [CLI Reference](../cli/cli-reference) — full reference for all flags available on `apply`, `plan`, and `destroy`
