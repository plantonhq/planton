# AWS KMS Key Deployment: From Manual Clicks to Production-Ready IaC

## The State of Encryption Key Management in AWS

Encryption keys are invisible until they're not. A misconfigured key policy locks an entire team out of critical data. A forgotten rotation schedule turns a compliance audit into a nightmare. A key deletion—accidental or intentional—makes terabytes of data permanently inaccessible.

AWS Key Management Service (KMS) has matured from a simple "encrypt my S3 bucket" checkbox into a sophisticated cryptographic service that integrates with dozens of AWS services. The question isn't whether to use KMS—it's **how to manage KMS keys at scale without creating security gaps or operational landmines**.

This document maps the landscape of AWS KMS deployment methods, from the pitfalls of manual console work to production-ready Infrastructure-as-Code approaches. It explains what Planton supports and why, grounded in patterns that teams actually use in production.

## Understanding Customer-Managed vs AWS-Managed Keys

Before diving into deployment methods, it's crucial to understand when you actually need a customer-managed KMS key.

**AWS-managed keys** (those with aliases like `aws/s3`, `aws/ebs`) are convenient: AWS creates them automatically when you enable encryption on a service, they have no monthly fee, and you don't manage policies. They work well for single-account scenarios where convenience trumps control.

**Customer-managed keys** are necessary when you need:
- **Cross-account access**: Share encrypted snapshots, allow external services to decrypt data
- **Custom key policies**: Enforce least-privilege access, separate admin vs usage permissions
- **Audit and compliance**: Track exactly who uses the key and when via CloudTrail
- **Key rotation control**: Enable automatic annual rotation or implement custom rotation schedules
- **Multi-region encryption**: Use the same key material across regions for disaster recovery

The cost difference is modest—$1/month per customer-managed key plus API usage fees—but the control difference is significant. AWS-managed keys cannot be shared across accounts, their policies cannot be modified, and they're tied to a single service.

## The Deployment Maturity Spectrum

### Level 0: Manual Console Creation (The Risk Factory)

Using the AWS console to click through KMS key creation is straightforward: navigate to KMS, click "Create Key", fill in description and alias, maybe enable rotation if you remember.

**What breaks:**
- **No reproducibility**: Six months later, you have no record of what policy was set or why
- **Configuration drift**: Different keys in different environments have subtly different settings
- **Human error**: Forgetting to enable rotation, setting a 7-day deletion window in production, granting overly broad permissions

**The pitfall:** A team creates a key for RDS encryption in the console. The default policy allows the entire AWS account full access. A developer accidentally schedules deletion while cleaning up test resources. The key enters a 30-day waiting period. No one notices until backups fail.

**Verdict:** Acceptable only for learning or one-off experiments in non-critical environments.

### Level 1: CLI Scripts and SDK Automation (The Half-Step)

AWS CLI (`aws kms create-key`) and SDKs (Boto3, AWS SDK) make key creation scriptable. You can check commands into source control and run them in pipelines.

**What this solves:**
- Repeatability: The same script creates the same key configuration
- Version control: Commands are documented in Git
- Integration: Scripts can be embedded in custom deployment workflows

**What it doesn't:**
- **Idempotency**: Running the script twice creates two keys unless you manually check for existence
- **State management**: No automatic tracking of what was created where
- **Dependency handling**: Coordinating key creation with dependent resources (S3 buckets, RDS instances) requires manual orchestration

**The pitfall:** A deployment script calls `aws kms create-key` without first checking if the alias already exists. Over time, the account accumulates orphaned keys—each costing $1/month—because the script creates a new one on every run.

**Verdict:** Better than manual, but requires careful design to avoid drift and duplication. Often used for specialized tasks like importing external key material (BYOK scenarios) that IaC tools don't fully support.

### Level 2: Configuration Management (Ansible)

Ansible's `amazon.aws.kms_key` module brings declarative intent to KMS management. You specify the desired state (key should exist with these properties) and Ansible ensures it.

**What this solves:**
- **Idempotence**: The module checks current state and only makes changes if needed
- **Policy as code**: Key policies and settings are version-controlled playbooks
- **Integration with server configuration**: Create a key and configure the EC2 instances that use it in one playbook

**What's tricky:**
- **State is ephemeral**: Unlike Terraform, Ansible doesn't store state; it checks live resources each run
- **Complex dependencies**: Managing a key, its alias, and 20 dependent resources requires careful playbook ordering
- **Limited multi-resource orchestration**: Better suited for procedural deployments than complex infrastructure graphs

**The pitfall:** The Ansible module requires an alias for every key to prevent "floating" keys without identifiers. If you try to reuse an alias that's already attached to another key, the playbook fails—you must manually clean up the old alias first.

**Verdict:** Excellent for teams already standardized on Ansible, especially when bundling key creation with server provisioning. For pure infrastructure provisioning at scale, dedicated IaC tools offer stronger guarantees.

## Level 3: Production-Ready IaC

### Terraform/OpenTofu: The Industry Standard

Terraform's `aws_kms_key` resource (and OpenTofu's identical implementation) represents the most widely deployed approach to KMS management in production.

**Core strengths:**
- **Declarative state**: Describe the desired key configuration; Terraform handles creation, updates, deletion
- **Plan-before-apply**: See exactly what will change before it happens
- **Dependency resolution**: Reference the key ARN in other resources; Terraform ensures correct creation order
- **Mature ecosystem**: Thousands of modules, extensive documentation, large community

**Production patterns:**

```hcl
resource "aws_kms_key" "prod_data" {
  description             = "Production customer data encryption"
  enable_key_rotation     = true
  deletion_window_in_days = 30

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AdminAccess"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::123456789012:role/SecurityAdminRole"
        }
        Action   = "kms:*"
        Resource = "*"
      },
      {
        Sid    = "ApplicationUsage"
        Effect = "Allow"
        Principal = {
          AWS = "arn:aws:iam::123456789012:role/AppServerRole"
        }
        Action = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = "*"
        Condition = {
          StringEquals = {
            "kms:EncryptionContext:App" = "CustomerPortal"
          }
        }
      }
    ]
  })
}

resource "aws_kms_alias" "prod_data" {
  name          = "alias/prod/customer-data"
  target_key_id = aws_kms_key.prod_data.key_id
}
```

**Critical safeguards:**
- Set `deletion_window_in_days = 30` (maximum) to prevent accidental data loss
- Always create an `aws_kms_alias` resource—never distribute raw key IDs
- Use specific IAM role ARNs in policies, not wildcards
- Require encryption context in key policies to prevent key misuse across contexts

**The gotcha:** Changing certain properties (like `key_spec` from symmetric to asymmetric) forces resource replacement. Terraform will schedule deletion of the old key (30-day wait) and create a new one. All data encrypted under the old key becomes inaccessible after deletion unless you plan the migration.

**Multi-region keys:** Terraform supports multi-region keys with the `multi_region` flag and a separate `aws_kms_replica_key` resource for replicas:

```hcl
# Primary in us-east-1
resource "aws_kms_key" "global" {
  description  = "Global app key"
  multi_region = true
}

# Replica in eu-west-1 (requires separate provider)
resource "aws_kms_replica_key" "global_eu" {
  provider              = aws.eu_west_1
  primary_key_arn       = aws_kms_key.global.arn
  deletion_window_in_days = 30
}
```

**Verdict:** The de facto standard for teams managing AWS infrastructure at scale. Requires discipline around state management (typically S3 backend with locking) and careful planning for immutable property changes.

### Pulumi: Code as Infrastructure

Pulumi uses general-purpose programming languages (TypeScript, Python, Go) to define infrastructure, offering the same AWS APIs as Terraform but with programmatic power.

**When it shines:**
- **Complex logic**: Generate multiple keys dynamically based on data structures
- **Type safety**: Catch configuration errors at compile time (in TypeScript/Go)
- **Abstraction**: Create reusable components that bundle a key, alias, and policy into a single logical unit

**Example in TypeScript:**

```typescript
import * as aws from "@pulumi/aws";

const createEncryptionKey = (name: string, appRole: string) => {
  const key = new aws.kms.Key(`${name}-key`, {
    description: `Encryption key for ${name}`,
    enableKeyRotation: true,
    deletionWindowInDays: 30,
    policy: JSON.stringify({
      Version: "2012-10-17",
      Statement: [
        {
          Sid: "AdminAccess",
          Effect: "Allow",
          Principal: { AWS: "arn:aws:iam::123456789012:role/SecurityAdmin" },
          Action: "kms:*",
          Resource: "*"
        },
        {
          Sid: "AppUsage",
          Effect: "Allow",
          Principal: { AWS: appRole },
          Action: ["kms:Decrypt", "kms:GenerateDataKey"],
          Resource: "*"
        }
      ]
    })
  });

  const alias = new aws.kms.Alias(`${name}-alias`, {
    name: `alias/${name}`,
    targetKeyId: key.id
  });

  return { key, alias };
};

// Create keys for multiple apps programmatically
["api", "worker", "analytics"].forEach(app => {
  createEncryptionKey(app, `arn:aws:iam::123456789012:role/${app}Role`);
});
```

**Trade-offs:**
- Steeper learning curve for teams unfamiliar with programming-based IaC
- Smaller ecosystem compared to Terraform (though growing rapidly)
- Same replacement behavior for immutable properties as Terraform

**Verdict:** Powerful for complex scenarios requiring loops, conditionals, or tight integration with application code. Planton leverages Pulumi's automation libraries under the hood for programmatic deployment.

### AWS CloudFormation/CDK: Native AWS Integration

**CloudFormation** (`AWS::KMS::Key`) is AWS's native IaC service. **CDK** provides a higher-level programming interface that synthesizes to CloudFormation templates.

**CloudFormation strengths:**
- Fully managed by AWS—no separate state storage
- Rollback on failure built-in
- Native integration with AWS Service Catalog, StackSets

**CDK improvements:**
- Sensible defaults: Keys are created with `RemovalPolicy: Retain` by default (won't delete on stack destroy)
- Higher-level constructs: `key.grantEncryptDecrypt(role)` automatically updates the key policy
- Programming abstractions without leaving the AWS ecosystem

**Example CDK (TypeScript):**

```typescript
import * as kms from 'aws-cdk-lib/aws-kms';
import * as iam from 'aws-cdk-lib/aws-iam';

const key = new kms.Key(this, 'ProdKey', {
  alias: 'alias/prod/app-data',
  description: 'Production application data encryption',
  enableKeyRotation: true,
  removalPolicy: cdk.RemovalPolicy.RETAIN
});

// Grant access programmatically
const appRole = iam.Role.fromRoleArn(this, 'AppRole', 'arn:aws:iam::123456789012:role/App');
key.grantEncryptDecrypt(appRole);
```

**Limitations:**
- YAML/JSON verbosity for raw CloudFormation
- Multi-region keys require workarounds (no native `AWS::KMS::ReplicaKey` as of this writing)
- Update constraints: Many properties can't be changed without replacement

**The safeguard:** Always set `DeletionPolicy: Retain` on `AWS::KMS::Key` resources. This prevents CloudFormation from scheduling key deletion when the stack is destroyed or the resource is removed—a critical protection against accidental data loss.

**Verdict:** Excellent for teams fully invested in the AWS ecosystem, especially when using CDK for type-safe infrastructure code. Be mindful of deletion protection and plan carefully for multi-region scenarios.

## Production Essentials

### Automatic Key Rotation

**Enable it for every production symmetric key.** AWS KMS will rotate the key material every 365 days (configurable between 90 and 2,560 days) without changing the key ID or ARN. Applications continue to work transparently; KMS retains old key material to decrypt older data.

Cost impact: Negligible—AWS charges up to $3/month for a key with multiple rotated versions (not an additional $1 per rotation).

**What can't be rotated automatically:**
- Asymmetric keys (RSA, ECC)
- HMAC keys
- Keys with imported material (BYOK)

For these, you must manually create new keys and update references.

### Key Policies: Least Privilege by Default

A production key policy has two sections:

1. **Administration**: Who can manage the key (update policies, schedule deletion, enable/disable)
2. **Usage**: Who can use the key for cryptographic operations (encrypt, decrypt, generate data keys)

**Never use wildcards in principals** unless you have a specific reason. Instead of:

```json
{
  "Principal": { "AWS": "arn:aws:iam::123456789012:root" },
  "Action": "kms:*"
}
```

Use specific roles:

```json
{
  "Principal": { 
    "AWS": [
      "arn:aws:iam::123456789012:role/SecurityAdminRole",
      "arn:aws:iam::123456789012:role/BackupRole"
    ]
  },
  "Action": ["kms:DescribeKey", "kms:PutKeyPolicy"]
}
```

**Encryption context** adds a cryptographic binding between the key and the use case:

```json
{
  "Action": "kms:Decrypt",
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "kms:EncryptionContext:App": "PaymentProcessor",
      "kms:EncryptionContext:Environment": "Production"
    }
  }
}
```

If an application calls KMS without providing the exact context, decryption fails—even if the IAM permissions are correct. This prevents key misuse across different applications or environments.

### Aliases: The Stable Reference

Always create an alias (`alias/prod/app-data`) for every key. Aliases provide:
- **Human-readable names** in CloudTrail logs and console views
- **Stable references**: Update the alias to point to a new key without changing application code
- **Easier rotation**: Point the alias to a new key during manual rotation scenarios

In IaC, pair every `aws_kms_key` with an `aws_kms_alias` (Terraform) or create keys with the `alias` property (CDK).

### Deletion Protection

**The safeguard every team forgets:** Set the longest deletion window (30 days in Terraform, CloudFormation) and use retention policies (CDK's `RemovalPolicy.RETAIN`, Terraform's lifecycle rules) to prevent IaC from auto-deleting keys on stack destroy.

A deleted key makes all encrypted data permanently inaccessible. There is no recovery.

### Multi-Region Keys: When and Why

Multi-region keys share the same key ID and material across AWS regions. Use them when:
- You replicate data across regions (DynamoDB Global Tables, S3 Cross-Region Replication)
- You need disaster recovery: decrypt data in a secondary region without re-encryption
- You run active-active systems that encrypt in Region A and decrypt in Region B

**Cost:** Each replica is billed as a separate key ($1/month per region). Two regions = $2/month total.

**What doesn't work:** Cross-account replicas. Multi-region keys must stay within the same AWS account.

## The 80/20 Configuration Rule

Most KMS keys use a small subset of available properties:

**Essential (covers 80% of use cases):**
- **Alias**: Friendly name (e.g., `alias/prod/db-encryption`)
- **Description**: Human-readable purpose
- **Enable rotation**: `true` for symmetric keys in production
- **Key spec**: `SYMMETRIC_DEFAULT` (AES-256) for 95% of scenarios
- **Deletion window**: 30 days (maximum safety)
- **Policy**: Admin and usage permissions

**Rare/Advanced:**
- Asymmetric keys (RSA, ECC) for digital signing or public-key encryption
- BYOK (bring your own key material) for compliance scenarios
- Custom key stores (CloudHSM integration)
- HMAC keys for message authentication codes

### Example: Production Database Encryption Key

```yaml
alias: "alias/prod/rds-primary"
description: "Production RDS cluster encryption key"
enableRotation: true
keySpec: SYMMETRIC_DEFAULT
deletionWindowDays: 30
policy:
  adminRole: "arn:aws:iam::123456789012:role/SecurityAdmin"
  usageRoles:
    - "arn:aws:iam::123456789012:role/RDSServiceRole"
  encryptionContext:
    App: "CustomerDatabase"
    Environment: "Production"
```

This covers 95% of production needs: rotation enabled, restrictive policy, encryption context for audit trails, maximum deletion protection.

## Integration Patterns

### S3 Bucket Encryption (SSE-KMS)

When using a customer-managed key for S3:
1. The key policy must allow the S3 service principal (`s3.amazonaws.com`) or the IAM roles writing objects
2. The bucket policy should enforce the specific key: `s3:x-amz-server-side-encryption-aws-kms-key-id`
3. For cross-account access, both the key policy and bucket policy must allow the external account

### RDS Encrypted Databases

RDS uses the key to encrypt the underlying storage. The key policy must allow:
- The RDS service principal (`rds.amazonaws.com`)
- The AWS account's RDS service-linked role (`AWSServiceRoleForRDS`)

For cross-account snapshot sharing, the key policy must explicitly allow the target account to use the key.

### Lambda Environment Variables

Lambda encrypts environment variables with a KMS key (AWS-managed by default). For a customer-managed key:
- The Lambda service principal needs `kms:Decrypt` permission
- Use encryption context to bind the key to a specific function: `"kms:EncryptionContext:LambdaFunctionName": "my-function"`

### CI/CD Pipelines

Store secrets encrypted with KMS in Parameter Store or Secrets Manager. The pipeline's IAM role needs:
- `secretsmanager:GetSecretValue` or `ssm:GetParameter`
- `kms:Decrypt` on the KMS key

This keeps secrets out of plaintext config files and Git repositories.

## Cost Optimization

**Key costs:** $1/month per key (flat fee)
**Usage costs:** ~$0.03 per 10,000 symmetric encrypt/decrypt requests

**Optimization strategies:**
1. **Consolidate when security allows**: Use one key for similar data types (e.g., all dev environment data) rather than creating hundreds of single-use keys
2. **Envelope encryption**: Call KMS once to generate a data key, then use that key locally to encrypt thousands of objects—1 KMS call instead of 1,000
3. **Data key caching**: AWS Encryption SDK caches data keys in memory for a short period, reducing KMS API calls
4. **Use AWS-managed keys for non-sensitive data**: Save $1/month per key if you don't need policy control
5. **Monitor usage**: Delete truly unused keys (carefully—ensure no encrypted data depends on them)

**Cost example:** An application with 5 million KMS operations per month:
- 5M / 10,000 × $0.03 = $15/month in API costs
- Plus $1 for the key = $16/month total

At 500 million operations (high-scale):
- 500M / 10,000 × $0.03 = $1,500/month

At this scale, envelope encryption and data key caching become critical optimizations.

## What Planton Supports

Planton's `AwsKmsKey` API focuses on the 80%: the configuration that production teams actually need.

**Supported:**
- Symmetric encryption keys (the default and most common use case)
- Asymmetric keys (RSA, ECC) for signing and verification
- Automatic key rotation with configurable periods
- Custom key policies with encryption context requirements
- Aliases for stable references
- Deletion windows (7-30 days)
- Multi-region keys (primary and replicas)
- Tags for cost allocation and management

**Design philosophy:**
- **Sane defaults**: Rotation enabled, 30-day deletion window, least-privilege policies
- **Protobuf-defined schemas**: Type-safe configuration that generates Terraform or Pulumi code
- **Integration with other resources**: Reference KMS keys in RDS, S3, DynamoDB specs; Planton handles dependency ordering

**What's intentionally out of scope (for now):**
- BYOK (bring your own key material)—complex import workflows better handled manually
- Custom key stores (CloudHSM integration)—advanced use case requiring dedicated setup
- Grants—managed via key policies instead for simplicity

## The Paradigm Shift

KMS key management has evolved from "click a checkbox in the console" to a critical security practice requiring:
- **Version-controlled policies** that define exactly who can decrypt what data
- **Automated rotation** to limit the blast radius of a potential key compromise
- **Encryption context** to prevent keys from being misused across different applications
- **Deletion protection** to avoid catastrophic data loss from accidental operations

The teams that treat KMS keys as infrastructure—defined in code, reviewed in pull requests, deployed through CI/CD—build systems that are both more secure and more maintainable than those that manage keys manually.

Planton codifies these patterns, making it straightforward to create production-ready KMS keys with the right defaults, while still allowing customization when needed. The goal is simple: make doing it the right way easier than doing it the wrong way.

---

**Next Steps:**
- See [IAC modules](../iac/) for Terraform and Pulumi implementation details
- Review [examples.md](../iac/pulumi/examples.md) for common configuration patterns
- Explore [stack_outputs.proto](../stack_outputs.proto) for available output values (key ID, ARN, alias ARN)

