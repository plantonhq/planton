---
title: "AWS Provider Setup"
description: "Configure AWS credentials for Planton deployments — IAM users, roles, environment variables, and provider config files"
icon: "cloud"
order: 70
---

# AWS Provider Setup

This guide covers everything you need to authenticate Planton with AWS. It applies to all AWS deployment components: `AwsRdsInstance`, `AwsS3Bucket`, `AwsEksCluster`, `AwsVpc`, and others.

For a quick reference of all provider credentials, see [Credentials](./credentials).

## Prerequisites

- An AWS account with permissions to create IAM users or roles
- The [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) installed (recommended, not required)

## Authentication Methods

### Method 1: Environment Variables

Set the standard AWS environment variables. Both Planton and the underlying IaC engines (Pulumi, Terraform, OpenTofu) read them automatically:

```bash
export AWS_ACCESS_KEY_ID="AKIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_REGION="us-west-2"

planton pulumi up -f ops/aws/database.yaml
```

| Variable | Required | Description |
|----------|----------|-------------|
| `AWS_ACCESS_KEY_ID` | Yes | 20-character key starting with `AKIA` (long-term) or `ASIA` (temporary) |
| `AWS_SECRET_ACCESS_KEY` | Yes | 40-character secret key |
| `AWS_REGION` | Yes | AWS region (e.g., `us-west-2`, `eu-central-1`) |
| `AWS_SESSION_TOKEN` | No | Session token for temporary credentials from STS |
| `AWS_PROFILE` | No | Named profile from `~/.aws/credentials` |

### Method 2: AWS CLI Profiles

If you manage multiple AWS accounts, use named profiles:

```bash
# Configure a named profile
aws configure --profile production
# Enter access key, secret key, and region when prompted

# Use the profile with Planton
export AWS_PROFILE=production
planton pulumi up -f ops/aws/database.yaml
```

### Method 3: Provider Config File (`-p`)

Pass credentials explicitly using the `-p` flag with a YAML file. The file format matches the `AwsProviderConfig` Protocol Buffer definition:

```yaml
# aws-credential.yaml
account_id: "123456789012"
access_key_id: "AKIAIOSFODNN7EXAMPLE"
secret_access_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
region: "us-west-2"
```

Deploy using the `-p` flag:

```bash
planton pulumi up -f ops/aws/database.yaml -p aws-credential.yaml
```

The CLI validates the config file against the proto schema, then converts the fields to environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`) for the IaC engine subprocess.

**Fields in the provider config file:**

| Field | Required | Description |
|-------|----------|-------------|
| `account_id` | Yes | AWS account ID (numeric string) |
| `access_key_id` | Yes | 20-character access key |
| `secret_access_key` | Yes | 40-character secret key |
| `region` | Yes | AWS region |
| `session_token` | No | STS session token for temporary credentials |

All field names use `snake_case`, matching the protobuf definition at `apis/dev/planton/provider/aws/provider.proto`.

### Method 4: IAM Roles (EC2, ECS, Lambda)

When running on AWS compute, use IAM roles instead of access keys. The AWS SDK automatically retrieves credentials from the instance metadata service:

```bash
# No credential configuration needed
# Ensure the EC2 instance, ECS task, or Lambda function has an IAM role attached
planton pulumi up -f ops/aws/database.yaml
```

This is the most secure method for production workloads running on AWS.

## Creating IAM Credentials

### IAM User for Local Development

```bash
# Create an IAM user
aws iam create-user --user-name planton-deployer

# Create access keys
aws iam create-access-key --user-name planton-deployer

# Output includes AccessKeyId and SecretAccessKey
# Store these securely — the secret key is shown only once
```

### IAM User for CI/CD

For CI/CD pipelines, create a dedicated IAM user with programmatic access:

```bash
# Create user
aws iam create-user --user-name planton-ci

# Create access keys
aws iam create-access-key --user-name planton-ci

# Attach a policy (see Least-Privilege Policies below)
aws iam attach-user-policy \
  --user-name planton-ci \
  --policy-arn arn:aws:iam::123456789012:policy/PlantonDeployerPolicy
```

Store the access key and secret key in your CI/CD platform's secret management (GitHub Actions secrets, GitLab CI variables, etc.).

## Least-Privilege Policies

Grant only the permissions required for the resources you deploy. Here are starting points for common components.

### S3 Bucket (`AwsS3Bucket`)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:CreateBucket",
        "s3:DeleteBucket",
        "s3:GetBucketPolicy",
        "s3:PutBucketPolicy",
        "s3:GetBucketVersioning",
        "s3:PutBucketVersioning",
        "s3:GetEncryptionConfiguration",
        "s3:PutEncryptionConfiguration",
        "s3:GetBucketTagging",
        "s3:PutBucketTagging",
        "s3:ListBucket"
      ],
      "Resource": "arn:aws:s3:::*"
    }
  ]
}
```

### RDS Instance (`AwsRdsInstance`)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "rds:CreateDBInstance",
        "rds:DeleteDBInstance",
        "rds:DescribeDBInstances",
        "rds:ModifyDBInstance",
        "rds:CreateDBSubnetGroup",
        "rds:DeleteDBSubnetGroup",
        "rds:DescribeDBSubnetGroups",
        "rds:AddTagsToResource",
        "rds:ListTagsForResource"
      ],
      "Resource": "*"
    }
  ]
}
```

### VPC (`AwsVpc`)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateVpc",
        "ec2:DeleteVpc",
        "ec2:DescribeVpcs",
        "ec2:ModifyVpcAttribute",
        "ec2:CreateSubnet",
        "ec2:DeleteSubnet",
        "ec2:DescribeSubnets",
        "ec2:CreateInternetGateway",
        "ec2:DeleteInternetGateway",
        "ec2:AttachInternetGateway",
        "ec2:DetachInternetGateway",
        "ec2:DescribeInternetGateways",
        "ec2:CreateRouteTable",
        "ec2:DeleteRouteTable",
        "ec2:CreateRoute",
        "ec2:DescribeRouteTables",
        "ec2:AssociateRouteTable",
        "ec2:CreateTags",
        "ec2:DeleteTags",
        "ec2:DescribeTags"
      ],
      "Resource": "*"
    }
  ]
}
```

For broader deployments across multiple component types, start with the `PowerUserAccess` managed policy and narrow permissions as you identify the exact resources being created.

## Verifying Credentials

```bash
# Check if credentials are configured
aws sts get-caller-identity

# Expected output:
# {
#   "UserId": "AIDAIOSFODNN7EXAMPLE",
#   "Account": "123456789012",
#   "Arn": "arn:aws:iam::123456789012:user/planton-deployer"
# }
```

If this command succeeds, Planton can use the same credentials.

## Troubleshooting

### "Unable to locate credentials"

The AWS SDK cannot find credentials. Check:

```bash
# Are environment variables set?
env | grep AWS

# Is there a credentials file?
cat ~/.aws/credentials

# Is the profile correct?
echo $AWS_PROFILE
```

### "Access Denied" or "UnauthorizedAccess"

The credentials are valid but lack the required permissions:

```bash
# Check which user/role the credentials belong to
aws sts get-caller-identity

# List attached policies
aws iam list-attached-user-policies --user-name <username>

# Check inline policies
aws iam list-user-policies --user-name <username>
```

Add the necessary IAM permissions for the resources you are deploying.

### "ExpiredToken" or "ExpiredTokenException"

Temporary credentials (from `aws sts assume-role` or SSO) have expired:

```bash
# Re-authenticate
aws sso login --profile <profile-name>

# Or re-assume the role
aws sts assume-role \
  --role-arn arn:aws:iam::123456789012:role/DeployerRole \
  --role-session-name planton-session
```

### Provider Config File Validation Error

The `-p` YAML file does not match the expected format:

- Field names must use `snake_case`: `access_key_id`, not `accessKeyId`
- `account_id` must contain only digits
- `access_key_id` must be exactly 20 characters starting with `AKIA` or `ASIA`
- `secret_access_key` must be exactly 40 characters

## What's Next

- [Credentials](./credentials) — Quick reference for all providers
- [GCP Provider Setup](./gcp-provider-setup) — Configure GCP credentials
- [Azure Provider Setup](./azure-provider-setup) — Configure Azure credentials
- [CI/CD Integration](./cicd-integration) — Use AWS credentials in pipelines
