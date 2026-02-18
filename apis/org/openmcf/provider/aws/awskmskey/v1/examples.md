# AwsKmsKey Examples

## Minimal manifest: Basic symmetric key with rotation

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: app-data-key
  org: my-org
spec:
  region: us-east-1
  description: "Encryption key for application data"
  aliasName: "alias/app/data"
```

## Production RDS encryption key

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: rds-prod-key
  org: my-org
  tags:
    environment: production
    purpose: rds-encryption
spec:
  region: us-east-1
  description: "Production RDS database encryption key"
  aliasName: "alias/prod/rds/main"
  deletionWindowDays: 30
  # disable_key_rotation: false (default - rotation enabled)
```

## S3 bucket encryption key with custom deletion window

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: s3-bucket-key
  org: my-org
  tags:
    app: document-storage
    team: backend
spec:
  region: us-east-1
  description: "S3 bucket server-side encryption key"
  aliasName: "alias/s3/documents"
  deletionWindowDays: 30
```

## EBS volume encryption key

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: ebs-encryption-key
  org: my-org
  tags:
    purpose: ebs-encryption
    environment: production
spec:
  region: us-east-1
  description: "EBS volume encryption for EC2 instances"
  aliasName: "alias/ec2/ebs-volumes"
  deletionWindowDays: 30
```

## Secrets Manager encryption key

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: secrets-manager-key
  org: my-org
spec:
  region: us-east-1
  description: "Encryption key for AWS Secrets Manager secrets"
  aliasName: "alias/secrets/app-config"
  deletionWindowDays: 30
```

## RSA key for asymmetric encryption

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: rsa-encryption-key
  org: my-org
spec:
  region: us-east-1
  keySpec: rsa_2048
  description: "RSA key for asymmetric encryption operations"
  aliasName: "alias/rsa/app-encryption"
  disableKeyRotation: true  # Rotation not supported for asymmetric keys
  deletionWindowDays: 30
```

## ECC key for digital signatures

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: ecc-signing-key
  org: my-org
  tags:
    purpose: digital-signatures
spec:
  region: us-east-1
  keySpec: ecc_nist_p256
  description: "ECC key for signing and verification"
  aliasName: "alias/signing/app-tokens"
  disableKeyRotation: true  # Rotation not supported for asymmetric keys
  deletionWindowDays: 30
```

## Key with minimum deletion window (not recommended for production)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: dev-test-key
  org: my-org
  tags:
    environment: development
spec:
  region: us-east-1
  description: "Development environment encryption key"
  aliasName: "alias/dev/test"
  deletionWindowDays: 7  # Minimum allowed, use for dev/test only
```

## Key without automatic rotation (asymmetric keys)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: rsa-4096-key
  org: my-org
spec:
  region: us-east-1
  keySpec: rsa_4096
  description: "RSA-4096 key for high-security encryption"
  aliasName: "alias/rsa/high-security"
  disableKeyRotation: true  # Required for asymmetric keys
  deletionWindowDays: 30
```

## Multi-environment pattern with consistent naming

```yaml
# Production
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: app-data-prod
  org: my-org
  tags:
    environment: production
    app: customer-portal
spec:
  region: us-east-1
  description: "Production encryption key for customer data"
  aliasName: "alias/prod/customer-data"
  deletionWindowDays: 30
---
# Staging
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: app-data-staging
  org: my-org
  tags:
    environment: staging
    app: customer-portal
spec:
  region: us-east-1
  description: "Staging encryption key for customer data"
  aliasName: "alias/staging/customer-data"
  deletionWindowDays: 30
```

## CLI flows

Validate manifest:
```bash
openmcf validate --manifest ./kms-key.yaml
```

Pulumi deploy:
```bash
openmcf pulumi update --manifest ./kms-key.yaml --stack my-org/project/prod --module-dir apis/org/openmcf/provider/aws/awskmskey/v1/iac/pulumi
```

Terraform deploy:
```bash
openmcf tofu apply --manifest ./kms-key.yaml --auto-approve
```

Get outputs:
```bash
# Get key ID
openmcf pulumi stack output key_id --stack my-org/project/prod

# Get key ARN (for cross-account policies)
openmcf pulumi stack output key_arn --stack my-org/project/prod

# Get alias name
openmcf pulumi stack output alias_name --stack my-org/project/prod
```

Use KMS key with AWS CLI:
```bash
# Encrypt data
aws kms encrypt \
  --key-id alias/prod/customer-data \
  --plaintext "sensitive data" \
  --output text \
  --query CiphertextBlob

# Decrypt data
aws kms decrypt \
  --ciphertext-blob fileb://encrypted-data.bin \
  --output text \
  --query Plaintext | base64 -d

# Generate data key for envelope encryption
aws kms generate-data-key \
  --key-id alias/prod/customer-data \
  --key-spec AES_256
```

Note: Provider credentials (AWS access key, secret, region) are supplied via stack input, not in the spec.

## Best practices demonstrated

1. **Always use aliases**: Reference keys by alias (`alias/prod/app-data`) not key IDs
2. **Descriptive names**: Use clear, consistent naming conventions across environments
3. **Maximum deletion window**: Set to 30 days for production keys (protects against accidents)
4. **Enable rotation for symmetric keys**: Leave `disableKeyRotation` unset (defaults to rotation enabled)
5. **Disable rotation for asymmetric keys**: RSA and ECC keys don't support automatic rotation
6. **Tag appropriately**: Use metadata tags for cost tracking and resource management
7. **Environment separation**: Use separate keys per environment (prod, staging, dev)
8. **Purpose-specific keys**: Create dedicated keys for different use cases (RDS, S3, Secrets Manager)

## Common patterns

### Envelope encryption pattern
```yaml
# Application uses this key to generate data encryption keys (DEKs)
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: envelope-master-key
  org: my-org
spec:
  region: us-east-1
  description: "Master key for envelope encryption (generates DEKs)"
  aliasName: "alias/app/envelope-master"
  deletionWindowDays: 30
```

Application code:
```python
# Generate data encryption key
response = kms_client.generate_data_key(
    KeyId='alias/app/envelope-master',
    KeySpec='AES_256'
)
plaintext_dek = response['Plaintext']
encrypted_dek = response['CiphertextBlob']

# Encrypt data locally with DEK
# Store encrypted DEK with encrypted data
# KMS never sees the actual data
```

### Cross-account sharing pattern
```yaml
# Create key with policy allowing another account
apiVersion: aws.openmcf.org/v1
kind: AwsKmsKey
metadata:
  name: shared-snapshots-key
  org: my-org
spec:
  region: us-east-1
  description: "Encryption key for cross-account RDS snapshot sharing"
  aliasName: "alias/shared/rds-snapshots"
  deletionWindowDays: 30
  # Note: Key policy for cross-account access must be configured
  # separately via AWS console or separate policy resource
```

## References

For more detailed patterns and architecture guidance, see:
- [Research documentation](docs/README.md) - Deep dive into KMS deployment approaches
- [Terraform examples](iac/tf/examples.md) - Additional Terraform-specific examples
- [Pulumi examples](iac/pulumi/examples.md) - Pulumi-specific examples

