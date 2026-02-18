# AWS KMS Key Examples

Below are several examples demonstrating how to define an AWS KMS Key component in
OpenMCF. After creating one of these YAML manifests, apply it with Terraform using the OpenMCF CLI:

```shell
openmcf tofu apply --manifest <yaml-path> --stack <stack-name>
```

---

## Basic KMS Key

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: basic-kms-key
spec:
  region: us-west-2
  description: "Basic symmetric KMS key for general encryption"
  keySpec: "symmetric"
  disableKeyRotation: false
  deletionWindowDays: 30
```

This example creates a basic KMS key:
• Symmetric key for general encryption operations.
• Automatic key rotation enabled (annual).
• 30-day deletion window for safety.
• Suitable for most encryption use cases.

---

## KMS Key for EKS Secrets

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: eks-secrets-kms-key
spec:
  region: us-west-2
  description: "KMS key for EKS cluster secrets encryption"
  keySpec: "symmetric"
  disableKeyRotation: false
  deletionWindowDays: 30
  aliasName: "alias/eks-secrets-key"
```

This example creates a KMS key for EKS:
• Symmetric key for Kubernetes secrets encryption.
• Automatic key rotation for security compliance.
• Custom alias for easy identification.
• 30-day deletion window for cluster safety.

---

## KMS Key for Application Encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: app-encryption-kms-key
spec:
  region: us-west-2
  description: "KMS key for application data encryption"
  keySpec: "symmetric"
  disableKeyRotation: false
  deletionWindowDays: 7
  aliasName: "alias/app-encryption-key"
```

This example creates an application encryption key:
• Symmetric key for application data encryption.
• Automatic key rotation enabled.
• Shorter deletion window for faster cleanup.
• Custom alias for application identification.

---

## RSA KMS Key for Asymmetric Encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: rsa-kms-key
spec:
  region: us-west-2
  description: "RSA 2048 KMS key for asymmetric encryption"
  keySpec: "rsa_2048"
  disableKeyRotation: false
  deletionWindowDays: 30
  aliasName: "alias/rsa-encryption-key"
```

This example creates an RSA key:
• RSA 2048-bit key for asymmetric encryption.
• Automatic key rotation enabled.
• Suitable for digital signatures and asymmetric encryption.
• Custom alias for key identification.

---

## ECC KMS Key for High Performance

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: ecc-kms-key
spec:
  region: us-west-2
  description: "ECC NIST P-256 KMS key for high-performance encryption"
  keySpec: "ecc_nist_p256"
  disableKeyRotation: false
  deletionWindowDays: 30
  aliasName: "alias/ecc-encryption-key"
```

This example creates an ECC key:
• ECC NIST P-256 key for high-performance encryption.
• Automatic key rotation enabled.
• Suitable for high-throughput applications.
• Custom alias for performance identification.

---

## KMS Key with Disabled Rotation

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: no-rotation-kms-key
spec:
  region: us-west-2
  description: "KMS key with disabled rotation for compliance"
  keySpec: "symmetric"
  disableKeyRotation: true
  deletionWindowDays: 30
  aliasName: "alias/compliance-key"
```

This example creates a key with disabled rotation:
• Symmetric key with manual rotation control.
• Automatic rotation disabled for compliance requirements.
• Custom alias for compliance identification.
• Suitable for regulatory compliance scenarios.

---

## KMS Key with Short Deletion Window

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: short-deletion-kms-key
spec:
  region: us-west-2
  description: "KMS key with short deletion window for development"
  keySpec: "symmetric"
  disableKeyRotation: false
  deletionWindowDays: 7
  aliasName: "alias/dev-key"
```

This example creates a key with short deletion window:
• Symmetric key for development environments.
• 7-day deletion window for faster cleanup.
• Automatic rotation enabled.
• Custom alias for development identification.

---

## Production KMS Key

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: production-kms-key
spec:
  region: us-west-2
  description: "Production KMS key for critical data encryption"
  keySpec: "symmetric"
  disableKeyRotation: false
  deletionWindowDays: 30
  aliasName: "alias/production-encryption-key"
```

This example creates a production key:
• Symmetric key for production data encryption.
• Automatic rotation for security compliance.
• 30-day deletion window for safety.
• Custom alias for production identification.
• Suitable for critical production workloads.

---

## KMS Key with Minimal Configuration

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: minimal-kms-key
spec:
  region: us-west-2
  description: "Minimal KMS key configuration"
```

This example creates a minimal KMS key:
• Only required description specified.
• Uses default symmetric key type.
• Automatic rotation enabled by default.
• 30-day deletion window by default.
• No custom alias.
• Suitable as a starting point for custom configurations.

---

## After Deploying

Once you've applied your manifest with OpenMCF tofu, you can confirm that the KMS key is active via the AWS console or by
using the AWS CLI:

```shell
aws kms describe-key --key-id <your-key-id>
```

For detailed key information including rotation status:

```shell
aws kms get-key-rotation-status --key-id <your-key-id>
```

To list key aliases:

```shell
aws kms list-aliases --key-id <your-key-id>
```

This will show the KMS key details including key type, rotation status, alias, and description.

