# AWS IAM Role

Deploys an IAM role: an assumable identity with temporary credentials, the
backbone of every service-to-service permission on AWS. The trust policy
controls who may assume the role; managed-policy references and inline
documents control what it can do; an optional permissions boundary caps the
maximum it can ever do.

## What Gets Created

When you deploy an AwsIamRole resource, Planton provisions:

- **IAM Role** — an `aws_iam_role` / `iam.Role` named from `metadata.name`,
  with the trust (assume-role) policy, optional description, IAM path, session
  duration ceiling, and permissions boundary
- **Managed Policy Attachments** — one `aws_iam_role_policy_attachment` /
  `iam.RolePolicyAttachment` per entry in `managedPolicyArns`, each reconciled
  individually so adding or removing an entry never touches the role
- **Inline Policies** — one `aws_iam_role_policy` / `iam.RolePolicy` per entry
  in `inlinePolicies`, embedding role-specific documents directly on the role

An instance profile is **not** created here — EC2 needs one to carry a role,
so wrap this role in an `AwsIamInstanceProfile` component when EC2 instances
will use it. Every other AWS service assumes the role directly.

## Prerequisites

- **AWS credentials** configured via the Planton provider config (keyless SSO/OIDC).
- **A trust policy document** defining which principals (services, accounts,
  or federated identities) may assume this role.

## Quick Start

Create a file `iam-role.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: my-ecs-task-role
spec:
  region: us-east-1
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: ecs-tasks.amazonaws.com
        Action: sts:AssumeRole
```

Deploy:

```shell
planton apply -f iam-role.yaml
```

This creates an IAM role that can be assumed by ECS tasks, with no additional
policies attached.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region for the provider's API calls. IAM is global — the role is assumable in every region. | Required |
| `trustPolicy` | `object` | JSON trust policy document defining which principals may assume this role. Updatable in place. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable purpose, shown in the IAM console. Updatable in place. Max 1000 characters. |
| `path` | `string` | `"/"` | IAM path for organizing and wildcard-matching roles (e.g. `/service-roles/`). Immutable. |
| `managedPolicyArns` | `StringValueOrRef[]` | `[]` | Managed policies to attach — `valueFrom` references to `AwsIamPolicy` resources, or literal ARNs for AWS-managed policies. |
| `inlinePolicies` | `map<string, object>` | `{}` | Map of policy name to JSON document for permissions unique to this role. |
| `maxSessionDuration` | `int` | `3600` | Ceiling for assumed-session duration, in seconds (3600–43200). |
| `permissionsBoundary` | `StringValueOrRef` | — | Managed policy whose grants cap this role's maximum permissions (intersection semantics). Reference an `AwsIamPolicy` or pass a literal ARN. |
| `forceDetachPolicies` | `bool` | `false` | Force-detach remaining policy attachments (including out-of-band ones) on deletion instead of failing. |

## Examples

### EC2 instance role (wrapped in an instance profile)

The role EC2 instances receive through an `AwsIamInstanceProfile`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: ec2-instance-role
spec:
  region: us-east-1
  description: Allows EC2 instances to call AWS services
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: ec2.amazonaws.com
        Action: sts:AssumeRole
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
---
apiVersion: aws.planton.dev/v1
kind: AwsIamInstanceProfile
metadata:
  name: ec2-instance-profile
spec:
  region: us-east-1
  role:
    valueFrom:
      kind: AwsIamRole
      name: ec2-instance-role
      fieldPath: status.outputs.role_name
```

### Lambda execution role attaching a Planton-defined policy

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: lambda-exec-role
spec:
  region: us-east-1
  description: Execution role for order-processing Lambda
  path: /service-roles/
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: lambda.amazonaws.com
        Action: sts:AssumeRole
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
    - valueFrom:
        kind: AwsIamPolicy
        name: orders-table-access
        fieldPath: status.outputs.policy_arn
```

### Cross-account role with a boundary and a longer session

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: cross-account-reader
spec:
  region: us-east-1
  description: Allows account 111111111111 to read S3 buckets in this account
  maxSessionDuration: 14400
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          AWS: arn:aws:iam::111111111111:root
        Action: sts:AssumeRole
        Condition:
          StringEquals:
            sts:ExternalId: unique-external-id-abc123
  managedPolicyArns:
    - value: arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
  permissionsBoundary:
    valueFrom:
      kind: AwsIamPolicy
      name: external-access-boundary
      fieldPath: status.outputs.policy_arn
```

### Worker role mixing shared and role-specific permissions

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: ecs-worker-role
spec:
  region: us-east-1
  description: Worker task role with SQS and S3 permissions
  trustPolicy:
    Version: "2012-10-17"
    Statement:
      - Effect: Allow
        Principal:
          Service: ecs-tasks.amazonaws.com
        Action: sts:AssumeRole
  managedPolicyArns:
    - valueFrom:
        kind: AwsIamPolicy
        name: worker-shared-logging
        fieldPath: status.outputs.policy_arn
  inlinePolicies:
    sqs-access:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - sqs:ReceiveMessage
            - sqs:DeleteMessage
            - sqs:GetQueueAttributes
          Resource: arn:aws:sqs:us-east-1:123456789012:worker-queue
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `role_arn` | `string` | ARN of the IAM role — what service integrations reference |
| `role_name` | `string` | Name of the role — what an `AwsIamInstanceProfile`'s `role` field references |
| `role_id` | `string` | Stable unique ID AWS assigns to the role (`AROA...`) |

## Related Components

- [AwsIamPolicy](/docs/catalog/aws/iam-policy) — reusable managed policies attached via `managedPolicyArns` or used as the permissions boundary
- [AwsIamInstanceProfile](/docs/catalog/aws/iam-instance-profile) — wraps this role to deliver it to EC2 instances
- [AwsIamUser](/docs/catalog/aws/iam-user) — long-lived IAM users for programmatic access
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — EKS clusters use IAM roles for the control plane and node groups
