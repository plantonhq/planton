# AWS KMS Key

Deploys a customer-managed AWS KMS encryption key with configurable key type, automatic rotation, and an optional alias. OpenMCF creates the key, applies organization and environment tags, and optionally registers an alias for human-readable key identification.

## What Gets Created

When you deploy an AwsKmsKey resource, OpenMCF provisions:

- **KMS Key** — a `kms.Key` resource with the specified cryptographic key type, description, rotation setting, and deletion window
- **KMS Alias** (conditional) — a `kms.Alias` resource created only when `aliasName` is provided, mapping a friendly name (e.g., `alias/my-app-key`) to the key ID

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **IAM permissions** to create and manage KMS keys (`kms:CreateKey`, `kms:CreateAlias`, `kms:EnableKeyRotation`, `kms:TagResource`)

## Quick Start

Create a file `kms-key.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: my-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsKmsKey.my-key
spec: {}
```

Deploy:

```shell
openmcf apply -f kms-key.yaml
```

This creates a symmetric KMS key with automatic annual rotation enabled and a 30-day deletion window.

## Configuration Reference

### Required Fields

All spec fields are optional. An empty `spec: {}` creates a symmetric key with sensible defaults.

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `keySpec` | `enum` | `symmetric` | Cryptographic key type. Valid values: `symmetric`, `rsa_2048`, `rsa_4096`, `ecc_nist_p256`. Maps to AWS key specs `SYMMETRIC_DEFAULT`, `RSA_2048`, `RSA_4096`, `ECC_NIST_P256`. |
| `description` | `string` | `""` | Human-readable description for the KMS key. Maximum 250 characters. |
| `disableKeyRotation` | `bool` | `false` | When `true`, disables automatic annual key rotation. When `false` (default), rotation is enabled. |
| `deletionWindowDays` | `int32` | `30` | Waiting period in days before the key is permanently deleted after scheduling deletion. Must be between 7 and 30 inclusive. |
| `aliasName` | `string` | — | Friendly name for the key. Must begin with `alias/` and contain only letters, numbers, underscores, or hyphens (up to 250 characters after the prefix). Example: `alias/my-app-encryption`. |

## Examples

### Default Symmetric Key

A symmetric encryption key with all defaults — rotation enabled, 30-day deletion window:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: default-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsKmsKey.default-key
spec: {}
```

### Symmetric Key with Alias and Description

A named key for application-level encryption:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: app-encryption-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsKmsKey.app-encryption-key
spec:
  description: "Encryption key for application secrets"
  aliasName: alias/app-encryption
```

### RSA Key for Asymmetric Operations

An RSA 4096-bit key for signing or encryption workflows that require asymmetric cryptography:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: rsa-signing-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsKmsKey.rsa-signing-key
spec:
  keySpec: rsa_4096
  description: "RSA key for JWT signing"
  aliasName: alias/jwt-signing
  deletionWindowDays: 14
```

### ECC Key with Short Deletion Window

An elliptic curve key for ECDSA signing with a minimal deletion safety window:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: ecc-signing-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.AwsKmsKey.ecc-signing-key
spec:
  keySpec: ecc_nist_p256
  description: "ECDSA P-256 signing key"
  aliasName: alias/ecdsa-signing
  deletionWindowDays: 7
  disableKeyRotation: true
```

### Full Configuration

A production symmetric key using every available field:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: prod-master-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsKmsKey.prod-master-key
spec:
  keySpec: symmetric
  description: "Master encryption key for production data"
  disableKeyRotation: false
  deletionWindowDays: 30
  aliasName: alias/prod-master
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `key_id` | `string` | Unique identifier (UUID) of the KMS key |
| `key_arn` | `string` | Full Amazon Resource Name of the KMS key, used in IAM policies and resource configurations |
| `alias_name` | `string` | The alias assigned to the key, or an empty string if no alias was specified |
| `rotation_enabled` | `bool` | Whether automatic annual key rotation is enabled (`true`) or disabled (`false`) |

## Related Components

- [AwsS3Bucket](/docs/catalog/aws/awss3bucket) — can use the KMS key ARN for server-side encryption of bucket objects
- [AwsRdsInstance](/docs/catalog/aws/awsrdsinstance) — can use the KMS key ARN for encrypting database storage
- [AwsSecretsManager](/docs/catalog/aws/awssecretsmanager) — can use the KMS key ARN to encrypt stored secrets
- [AwsEksCluster](/docs/catalog/aws/awsekscluster) — can use the KMS key ARN for envelope encryption of Kubernetes secrets
