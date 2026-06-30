# AwsKmsKey

AWS Key Management Service (KMS) customer-managed keys provide cryptographic operations for data encryption, signing, and verification with fine-grained access control, automatic key rotation, and comprehensive audit logging. This resource creates and manages KMS keys that integrate with other AWS services or can be used directly for application-level encryption.

## Spec fields (summary)
- key_spec: Type of KMS key (symmetric, RSA_2048, RSA_4096, ECC_NIST_P256) - defaults to symmetric
- description: Human-readable description of the key's purpose (max 250 characters)
- disable_key_rotation: Set to true to disable automatic annual rotation (default: false, rotation enabled)
- deletion_window_days: Waiting period before deletion (7-30 days, recommended: 30)
- alias_name: Optional alias starting with "alias/" for easy key reference (1-250 characters)

## Stack outputs
- key_id: Unique identifier for the KMS key
- key_arn: Amazon Resource Name (ARN) of the KMS key
- alias_name: Alias assigned to the key (if provided)
- rotation_enabled: Whether automatic key rotation is enabled

## How it works
This resource is orchestrated by the Planton CLI as part of a stack-update. The CLI validates your manifest, generates stack inputs, and invokes IaC backends in this repo:
- Pulumi (Go modules under iac/pulumi)
- Terraform (modules under iac/tf)

Customer-managed keys provide control over key policies, cross-account access, rotation schedules, and audit trails that AWS-managed keys cannot offer.

## Common use cases
- **RDS/Aurora encryption**: Encrypt databases with custom keys for cross-account snapshot sharing
- **S3 bucket encryption**: Server-side encryption with customer-managed keys (SSE-KMS)
- **EBS volume encryption**: Encrypt EC2 volumes with auditable, rotatable keys
- **Secrets Manager integration**: Encrypt secrets with your own KMS keys
- **Cross-account data sharing**: Share encrypted snapshots or AMIs across AWS accounts
- **Application-level encryption**: Encrypt/decrypt data directly using GenerateDataKey API
- **Envelope encryption**: Generate data encryption keys (DEKs) for local encryption
- **Digital signatures**: Sign and verify data using asymmetric keys (RSA/ECC)

## Key types explained
- **symmetric** (default): AES-256-GCM encryption, fastest, supports envelope encryption, cannot leave AWS
- **rsa_2048**: Asymmetric encryption and signing, 2048-bit key, moderate security
- **rsa_4096**: Asymmetric encryption and signing, 4096-bit key, higher security, slower
- **ecc_nist_p256**: Elliptic curve signing, faster than RSA, smaller key size, signing only (no encryption)

## Security best practices
- **Enable automatic rotation**: Leave `disable_key_rotation` as false (default) for symmetric keys
- **Set maximum deletion window**: Use 30 days (maximum) to protect against accidental deletion
- **Use aliases**: Never distribute key IDs directly; use aliases like `alias/prod/app-data`
- **Least privilege policies**: Grant only required KMS permissions (Encrypt, Decrypt, GenerateDataKey)
- **Encryption context**: Use encryption context in policies to prevent key misuse across different contexts
- **Monitor with CloudTrail**: Track all KMS API calls for audit and compliance
- **Multi-region for DR**: Use multi-region keys for disaster recovery scenarios (separate resource)
- **Separate admin/usage**: Different IAM roles for key administration vs usage

## Cost optimization
- **$1/month per key**: Each customer-managed key costs $1/month
- **API request charges**: $0.03 per 10,000 requests (encrypt, decrypt, generate data key)
- **Free tier**: 20,000 requests/month free
- **Reuse keys**: Use one key per environment/application, not one per resource
- **Consolidate policies**: Update key policies instead of creating new keys
- **Delete unused keys**: Schedule deletion for keys no longer needed (30-day window)

## When NOT to use customer-managed keys
- **Single-account, simple encryption**: AWS-managed keys (aws/s3, aws/ebs) are free and sufficient
- **No audit requirements**: If you don't need detailed KMS API logging
- **No cross-account sharing**: AWS-managed keys work fine within one account
- **No custom policies needed**: Default AWS-managed key policies cover most use cases
- **Cost-sensitive**: $1/month per key adds up with hundreds of keys

## References
- AWS KMS: https://docs.aws.amazon.com/kms/latest/developerguide/overview.html
- Key types: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#master_keys
- Key policies: https://docs.aws.amazon.com/kms/latest/developerguide/key-policies.html
- Key rotation: https://docs.aws.amazon.com/kms/latest/developerguide/rotate-keys.html
- Encryption context: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#encrypt_context
- Multi-region keys: https://docs.aws.amazon.com/kms/latest/developerguide/multi-region-keys-overview.html
- Research documentation: [docs/README.md](docs/README.md)
- Examples: [examples.md](examples.md)

