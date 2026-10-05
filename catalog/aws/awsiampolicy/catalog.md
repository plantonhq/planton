# AWS IAM Policy

Deploys a customer-managed IAM policy — the reusable unit of AWS permissions: a standalone, versioned permission document with its own ARN that can be attached to many roles and users at once. Define a grant once ("read-only access to the analytics bucket", "the permissions boundary for CI jobs"), attach it everywhere it is needed, and update it in exactly one place. Roles and users attach it through their `managedPolicyArns` fields, and a `permissionsBoundary` also takes a policy ARN, which makes this kind a leaf that much of an AWS architecture composes onto. Its `policy_arn` output is what every attachment references via ValueFromRef.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Managed Policy** -- the named permission document under its IAM path. The policy name comes from `metadata.name`; name, path, and description are create-only in AWS (changing any of them replaces the policy)
- **Policy Versions** -- each document update becomes a new version and is promoted to default. AWS keeps at most 5 versions, and the module prunes the oldest non-default version before saving a new one, so updates keep working indefinitely without manual version cleanup

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **The ARNs the policy will govern** -- statements name exact resources (bucket ARNs, table ARNs, queue ARNs). Gather them before authoring; wildcards that stand in for unknown ARNs are how over-broad policies are born.
- **No pre-existing resources required** -- the policy is a leaf: it references nothing and grants nothing until attached.

## Deploy

### Console

Open the deployment store, find **AWS IAM Policy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the document editor opens seeded with the current policy language version so you author statements directly. Start from the **S3 Read-Only Access** preset in the [Presets](#presets) tab to pre-populate the most common shared grant.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsIamPolicy
metadata:
  name: s3-read-only
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  description: Read-only access to the analytics bucket
  policyDocument:
    Version: "2012-10-17"
    Statement:
      - Sid: AnalyticsReadOnly
        Effect: Allow
        Action:
          - s3:GetObject
          - s3:ListBucket
        Resource:
          - arn:aws:s3:::analytics
          - arn:aws:s3:::analytics/*
```

```shell
planton apply -f iam-policy.yaml
```

This publishes the shared read-only grant; every role that attaches it inherits exactly these permissions, and widening the grant later is a one-place edit. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring an IAM policy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The document is the policy** -- `policyDocument` is the IAM policy language: each statement pairs an Effect (Allow or Deny) with the Actions it covers and the Resources it applies to. Prefer exact actions and exact ARNs — `"Action": "s3:*"` on `"Resource": "*"` works today and becomes the finding in next year's audit. An explicit Deny anywhere overrides every Allow any attached policy grants, which is what makes Deny statements the backbone of permission boundaries.

**Only the document updates in place** -- Each document edit becomes a new policy version, promoted to default, with attachments following immediately. Everything else — name, path, description — is create-only: changing any of them REPLACES the policy, and the IaC engine re-creates its attachments. Write the description as if it is permanent, because it is.

**The path is a policy-administration handle** -- The IAM path (`/service-boundaries/`) is an ARN infix that IAM conditions can match on: delegated administrators can be restricted to attaching only the policies published under a given path. Most teams leave it as `/`; organizations that delegate IAM administration choose paths deliberately.

**Region is a management detail** -- IAM is global: the policy is visible and attachable in every region. The region only picks the API endpoint the provider calls while managing it.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The policy is a leaf — it references no other Infra Components; the ARNs its statements govern travel as plain strings inside the policy document.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `policy_arn` | Amazon Resource Name of the managed policy | AwsIamRole / AwsIamUser `managedPolicyArns[]` and `permissionsBoundary` |
| `policy_id` | The stable unique ID AWS assigns (ANPA…) | Audit trails — it never encodes the name or path |
| `policy_name` | The friendly name (mirrors `metadata.name`) | IAM console URLs and CLI commands |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Shared read-only grant** -- One managed policy granting read access to a bucket or table, attached by every consumer role — the most common shared permission set in an AWS estate. Start from the **S3 Read-Only Access** preset.

**Permissions boundary** -- An allow-list of workload services plus an explicit Deny on identity escalation, applied through roles' `permissionsBoundary` field: the ceiling no attached policy can escalate past, and the mechanism that makes delegated role creation safe. Start from the **Workload Permissions Boundary** preset.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- attaches this policy through `managedPolicyArns` and can carry it as a `permissionsBoundary`
- [**AWS IAM Instance Profile**](/infra-catalog/aws-iam-instance-profile) -- delivers a role (and the policies attached to it) to EC2 instances
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- the most common Resource target of shared read/write grants
- [**AWS Lambda**](/infra-catalog/aws-lambda) -- its execution role attaches managed policies for the function's data access
