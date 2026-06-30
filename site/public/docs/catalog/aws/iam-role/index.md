---
title: "IAM Role"
description: "IAM Role deployment documentation"
icon: "package"
order: 100
componentName: "awsiamrole"
---

# AWS IAM Role

Deploys an AWS IAM Role with a configurable trust policy, optional managed policy attachments, and optional inline policy documents. The component creates the role, attaches any specified policies, and exports the role ARN and name for use by other components.

## What Gets Created

When you deploy an AwsIamRole resource, Planton provisions:

- **IAM Role** — an `iam.Role` resource with the specified name, trust (assume-role) policy, optional description, and IAM path
- **Managed Policy Attachments** — one `iam.RolePolicyAttachment` per entry in `managedPolicyArns`, linking the role to existing AWS-managed or customer-managed policies
- **Inline Policies** — one `iam.RolePolicy` per entry in `inlinePolicies`, embedding policy documents directly on the role

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **A trust policy document** defining which principals (services, accounts, or users) may assume this role
- **Policy ARNs** for any managed policies you want to attach (AWS-managed or customer-managed)

## Quick Start

Create a file `iam-role.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: my-ecs-task-role
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsIamRole.my-ecs-task-role
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

This creates an IAM role that can be assumed by ECS tasks, with no additional policies attached.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the resource will be created (e.g., `us-east-1`). | Required |
| `trustPolicy` | `object` | JSON trust policy document defining which principals may assume this role. Serialized as a `google.protobuf.Struct`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | `""` | Human-readable description of the IAM role's purpose. |
| `path` | `string` | `"/"` | IAM path for the role, used for organizational grouping (e.g., `/service-roles/`). |
| `managedPolicyArns` | `string[]` | `[]` | ARNs of AWS-managed or customer-managed IAM policies to attach. Must be unique. |
| `inlinePolicies` | `map<string, object>` | `{}` | Map of inline policy names to IAM policy documents. Keys are policy names; values are `google.protobuf.Struct` policy documents. |

## Examples

### EC2 Instance Role with Managed Policies

A role that EC2 instances can assume, with the SSM managed policy attached:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: ec2-instance-role
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsIamRole.ec2-instance-role
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
    - arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore
```

### Lambda Execution Role with Inline Policy

A role for Lambda functions with an inline policy granting DynamoDB access:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: lambda-exec-role
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsIamRole.lambda-exec-role
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
    - arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
  inlinePolicies:
    dynamodb-access:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - dynamodb:GetItem
            - dynamodb:PutItem
            - dynamodb:Query
          Resource: arn:aws:dynamodb:us-east-1:123456789012:table/orders
```

### Cross-Account Assume Role

A role that allows a different AWS account to assume it, useful for cross-account resource access:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: cross-account-reader
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsIamRole.cross-account-reader
spec:
  region: us-east-1
  description: Allows account 111111111111 to read S3 buckets in this account
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
    - arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess
```

### ECS Task Role with Multiple Inline Policies

A role combining managed and inline policies for fine-grained ECS task permissions:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamRole
metadata:
  name: ecs-worker-role
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsIamRole.ecs-worker-role
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
    - arn:aws:iam::aws:policy/CloudWatchLogsFullAccess
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
    s3-upload:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - s3:PutObject
          Resource: arn:aws:s3:::output-bucket/*
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `role_arn` | `string` | ARN of the created IAM role |
| `role_name` | `string` | Name of the IAM role in AWS |

## Related Components

- [AwsIamUser](/docs/catalog/aws/iam-user) — creates long-lived IAM users for programmatic access
- [AwsEksCluster](/docs/catalog/aws/eks-cluster) — EKS clusters frequently use IAM roles for node groups and service accounts
- [AwsS3Bucket](/docs/catalog/aws/s3-bucket) — bucket policies often reference IAM role ARNs for access control
