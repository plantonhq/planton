# AWS IAM Policy

Deploys a customer-managed IAM policy: a standalone, versioned permission
document with its own ARN. Attach it to many roles and users at once through
their `managedPolicyArns` fields, or use it as a permissions boundary -- one
definition, referenced everywhere it is needed, updated in one place.

## What Gets Created

When you deploy an AwsIamPolicy resource, Planton provisions:

- **Managed policy** — an `aws_iam_policy` / `iam.Policy` holding the permission
  document, with the name taken from `metadata.name` and an optional IAM path
  and description. Document updates create a new policy version and mark it
  default; the oldest non-default version is pruned automatically so updates
  never hit AWS's 5-version cap.

Role and user attachments are **not** created here — reference this policy's
`policy_arn` output from an `AwsIamRole` or `AwsIamUser` component's
`managedPolicyArns` (or `permissionsBoundary`) field.

## Prerequisites

- **AWS credentials** configured via the Planton provider config (keyless SSO/OIDC).
- **A permission document**: an IAM policy JSON with `Version` and `Statement`.

## Quick Start

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamPolicy
metadata:
  name: s3-read-only
spec:
  region: us-west-2
  description: Read-only access to the analytics bucket
  policyDocument:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Action:
          - s3:GetObject
          - s3:ListBucket
        Resource:
          - arn:aws:s3:::analytics-bucket
          - arn:aws:s3:::analytics-bucket/*
```

```shell
planton apply -f policy.yaml
```

## Configuration Reference

| Field | Type | Default | Description |
| --- | --- | --- | --- |
| `region` | `string` | — | AWS region for the provider's API calls. IAM is global — the policy is attachable in every region. Required. |
| `policyDocument` | `object` | — | The permission document (IAM policy JSON with `Version` and `Statement`). The only field updatable after creation. Required. |
| `description` | `string` | — | Human-readable purpose, shown in the IAM console. Immutable — changing it replaces the policy. Max 1000 characters. |
| `path` | `string` | `/` | IAM path for organizing and wildcard-matching policies (e.g. `/service-boundaries/`). Must begin and end with `/`. Immutable. |

## Examples

### Permissions boundary for CI principals

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamPolicy
metadata:
  name: ci-permissions-boundary
spec:
  region: us-west-2
  description: Maximum permissions any CI-created principal can hold
  path: /boundaries/
  policyDocument:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Action:
          - s3:*
          - dynamodb:*
          - logs:*
        Resource: "*"
      - Effect: Deny
        Action:
          - iam:*
          - organizations:*
        Resource: "*"
```

### Attaching to a role by reference

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: analytics-reader
spec:
  region: us-west-2
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: lambda.amazonaws.com
        Action: sts:AssumeRole
  managedPolicyArns:
    - valueFrom:
        kind: AwsIamPolicy
        name: s3-read-only
        fieldPath: status.outputs.policy_arn
```

## Stack Outputs

| Output | Description |
| --- | --- |
| `policy_arn` | ARN of the managed policy — what attachments and permissions boundaries reference |
| `policy_id` | Stable unique ID AWS assigns to the policy (`ANPA...`) |
| `policy_name` | Friendly name of the policy |

## Related Components

- [AwsIamRole](/docs/catalog/aws/iam-role) — attaches this policy via `managedPolicyArns` or uses it as a permissions boundary
- [AwsIamUser](/docs/catalog/aws/iam-user) — attaches this policy via `managedPolicyArns`
- [AwsIamInstanceProfile](/docs/catalog/aws/iam-instance-profile) — wraps a role whose permissions come from policies like this one
