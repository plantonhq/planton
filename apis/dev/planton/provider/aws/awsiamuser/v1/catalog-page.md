# AWS IAM User

Deploys an AWS IAM User with optional managed policy attachments, inline policy documents, and access key creation. The component creates the user, attaches policies, optionally generates an access key pair, and exports credentials and identifiers for use by other components.

## What Gets Created

When you deploy an AwsIamUser resource, Planton provisions:

- **IAM User** — an `iam.User` resource with the specified username and tags
- **Managed Policy Attachments** — one `iam.UserPolicyAttachment` per entry in `managedPolicyArns`, linking the user to existing AWS-managed or customer-managed policies
- **Inline Policies** — one `iam.UserPolicy` per entry in `inlinePolicies`, embedding policy documents directly on the user
- **Access Key** — an `iam.AccessKey` resource created by default, providing an access key ID and base64-encoded secret key; skipped when `disableAccessKeys` is `true`

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **Policy ARNs** for any managed policies you want to attach (ARNs must start with `arn:aws:iam::`)

## Quick Start

Create a file `iam-user.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamUser
metadata:
  name: my-ci-user
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsIamUser.my-ci-user
spec:
  region: us-east-1
  userName: my-ci-user
```

Deploy:

```shell
planton apply -f iam-user.yaml
```

This creates an IAM user named `my-ci-user` with one active access key pair and no policies attached.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | The AWS region where the resource will be created. | Required |
| `userName` | `string` | IAM user name. Must be 1-64 characters. | Pattern: `^[a-zA-Z0-9+=,.@_-]{1,64}$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `managedPolicyArns` | `string[]` | `[]` | ARNs of AWS-managed or customer-managed IAM policies to attach. Must be unique. Each ARN must match `^arn:aws:iam::`. |
| `inlinePolicies` | `map<string, object>` | `{}` | Map of inline policy names to IAM policy documents. Keys are policy names (max 128 characters); values are `google.protobuf.Struct` policy documents. |
| `disableAccessKeys` | `bool` | `false` | When `true`, prevents creation of access keys for this user. When `false`, one active access key pair is created. |

## Examples

### CI/CD Pipeline User with S3 Access

A user for a CI pipeline that needs to push artifacts to an S3 bucket:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamUser
metadata:
  name: ci-deploy-user
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsIamUser.ci-deploy-user
spec:
  region: us-east-1
  userName: ci-deploy-user
  managedPolicyArns:
    - arn:aws:iam::aws:policy/AmazonS3FullAccess
```

### Service Account with Inline Policy

A user for a third-party integration with a scoped inline policy granting read access to a specific DynamoDB table:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamUser
metadata:
  name: analytics-service
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsIamUser.analytics-service
spec:
  region: us-east-1
  userName: analytics-service
  inlinePolicies:
    dynamodb-read:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - dynamodb:GetItem
            - dynamodb:Query
            - dynamodb:Scan
          Resource: arn:aws:dynamodb:us-east-1:123456789012:table/events
```

### User Without Access Keys

A user intended for AWS Management Console access only, with no programmatic access keys:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamUser
metadata:
  name: audit-viewer
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsIamUser.audit-viewer
spec:
  region: us-east-1
  userName: audit-viewer
  disableAccessKeys: true
  managedPolicyArns:
    - arn:aws:iam::aws:policy/ReadOnlyAccess
```

### Full-Featured User with Multiple Policies

A production service account with both managed and inline policies for SQS and CloudWatch access:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsIamUser
metadata:
  name: worker-service
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsIamUser.worker-service
spec:
  region: us-east-1
  userName: worker-service
  managedPolicyArns:
    - arn:aws:iam::aws:policy/CloudWatchLogsFullAccess
  inlinePolicies:
    sqs-consumer:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - sqs:ReceiveMessage
            - sqs:DeleteMessage
            - sqs:GetQueueAttributes
          Resource: arn:aws:sqs:us-east-1:123456789012:task-queue
    s3-results:
      Version: "2012-10-17"
      Statement:
        - Effect: Allow
          Action:
            - s3:PutObject
          Resource: arn:aws:s3:::results-bucket/*
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `user_arn` | `string` | ARN of the created IAM user |
| `user_name` | `string` | Friendly name of the IAM user |
| `user_id` | `string` | Stable unique ID of the IAM user |
| `access_key_id` | `string` | Access key ID for the user (present only if access keys are enabled) |
| `secret_access_key` | `string` | Base64-encoded secret key associated with the access key (sensitive; present only if access keys are enabled) |
| `console_url` | `string` | AWS Management Console sign-in URL (`https://signin.aws.amazon.com/console`) |

## Related Components

- [AwsIamRole](/docs/catalog/aws/awsiamrole) — creates IAM roles for service-level permission delegation via temporary credentials
- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — bucket policies can reference IAM user ARNs for access control
- [AwsEksCluster](/docs/catalog/aws/awsekscluster) — IAM user credentials are sometimes used for programmatic cluster access
