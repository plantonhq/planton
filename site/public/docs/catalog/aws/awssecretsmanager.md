---
title: "Secrets Manager"
description: "Secrets Manager deployment documentation"
icon: "package"
order: 100
componentName: "awssecretsmanager"
---

# AWS Secrets Manager

Deploys a set of secrets in AWS Secrets Manager from a list of logical secret names. Each secret is created with a unique identifier derived from the resource metadata and seeded with a placeholder value, ready for out-of-band population via the AWS SDK or console.

## What Gets Created

When you deploy an AwsSecretsManager resource, OpenMCF provisions:

- **Secrets Manager Secret** — one `secretsmanager.Secret` resource per entry in `secretNames`, named with the pattern `{metadata.id}-{secretName}` for uniqueness within the AWS account
- **Placeholder Secret Version** — one `secretsmanager.SecretVersion` per secret, seeded with a placeholder string value; subsequent updates to the secret value outside of OpenMCF are preserved (the `secretString` field is set to ignore changes)

All resources are tagged with OpenMCF metadata (organization, environment, resource kind, resource ID).

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **Unique secret names** — each name in `secretNames` must be unique within the manifest (enforced by validation)

## Quick Start

Create a file `secrets.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: my-app-secrets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSecretsManager.my-app-secrets
spec:
  secretNames:
    - DB_PASSWORD
```

Deploy:

```shell
openmcf apply -f secrets.yaml
```

This creates a single secret in AWS Secrets Manager. After deployment, populate the actual secret value using the AWS SDK or console.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `secretNames` | `string[]` | List of logical secret names to create. Each name becomes a separate secret in AWS Secrets Manager, stored with a unique ID of `{metadata.id}-{name}`. | Minimum 1 item, all items unique, each item minimum length 1 |

### Optional Fields

This component has no optional fields. All behavior is determined by the `secretNames` list and the resource metadata.

## Examples

### Single Secret

Create one secret for a database password:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: db-creds
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSecretsManager.db-creds
spec:
  secretNames:
    - DB_PASSWORD
```

### Multiple Application Secrets

Create several secrets for a microservice that needs database, cache, and API credentials:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: payment-svc-secrets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsSecretsManager.payment-svc-secrets
spec:
  secretNames:
    - DB_PASSWORD
    - REDIS_AUTH_TOKEN
    - STRIPE_API_KEY
    - STRIPE_WEBHOOK_SECRET
```

### Production Environment Secrets

A production deployment with secrets grouped by purpose:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsSecretsManager
metadata:
  name: prod-platform-secrets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsSecretsManager.prod-platform-secrets
spec:
  secretNames:
    - RDS_MASTER_PASSWORD
    - ELASTICACHE_AUTH_TOKEN
    - JWT_SIGNING_KEY
    - OAUTH_CLIENT_SECRET
    - SMTP_PASSWORD
    - DATADOG_API_KEY
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `secret_arn_map` | `map<string, string>` | Map of logical secret names to their full AWS ARNs. Key is the secret name as specified in `secretNames` (e.g., `DB_PASSWORD`). Value is the full ARN (e.g., `arn:aws:secretsmanager:us-east-1:123456789012:secret:myapp-prod-secrets-DB_PASSWORD-XyZ789`). |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/awskmskey) — provides a customer-managed KMS key for encrypting secrets at rest (configured at the AWS account level or via resource policy)
- [AwsIamRole](/docs/catalog/aws/awsiamrole) — creates IAM roles with policies granting `secretsmanager:GetSecretValue` access
- [AwsLambda](/docs/catalog/aws/awslambda) — can consume secrets at runtime via environment variable references or SDK calls
- [AwsEcsService](/docs/catalog/aws/awsecsservice) — can reference secret ARNs for container secret injection
- [AwsRdsInstance](/docs/catalog/aws/awsrdsinstance) — often paired to store database master passwords
