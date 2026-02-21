# Alibaba Cloud KMS Key: From Console Creation to Control Plane Automation

## Introduction

Alibaba Cloud Key Management Service (KMS) provides centralized key management for cryptographic operations across the cloud platform. A customer-managed key (CMK) is the fundamental unit of KMS -- it is the cryptographic material that encrypts, decrypts, signs, or verifies data on behalf of other Alibaba Cloud services.

Nearly every data-at-rest encryption feature on Alibaba Cloud delegates to KMS. When you enable Transparent Data Encryption (TDE) on an RDS instance, Server-Side Encryption (SSE-KMS) on an OSS bucket, or disk encryption on an ECS instance, the service creates an envelope encryption relationship with a KMS key. The data key that encrypts your data is itself encrypted by the CMK, and only KMS can unwrap it. This means the CMK is the root of trust -- lose the key, lose the data.

This document covers the full spectrum of KMS key deployment methods on Alibaba Cloud, explains the design decisions behind the OpenMCF `AliCloudKmsKey` component, and provides production best practices for key lifecycle management.

## Evolution and Historical Context

### Default Service Keys vs. Customer-Managed Keys

Every Alibaba Cloud account has a set of service-managed keys (default keys) that are automatically created when you first enable encryption on a service. These keys are free, require zero configuration, and are managed entirely by Alibaba Cloud. For many development and staging workloads, they are sufficient.

Customer-managed keys (CMKs) exist for organizations that need:

1. **Key lifecycle control** -- the ability to rotate, disable, or schedule deletion of keys
2. **Access policy granularity** -- RAM policies that control which principals can use which keys
3. **Audit trail** -- ActionTrail logs for every cryptographic operation
4. **Compliance requirements** -- regulatory frameworks (PCI-DSS, HIPAA, SOC 2) that mandate customer-controlled encryption keys
5. **Cross-service key sharing** -- using the same key across RDS, OSS, and ECS for unified key management

### Symmetric vs. Asymmetric Keys

KMS supports two categories of keys:

**Symmetric keys** (`ENCRYPT/DECRYPT` usage) are the workhorse of data encryption. They use the same key material for both encryption and decryption. Alibaba Cloud supports AES (128, 192, 256-bit) and SM4 (Chinese national standard) algorithms. These are the keys used for envelope encryption by RDS, OSS, ECS, and other services.

**Asymmetric keys** (`SIGN/VERIFY` usage) use a key pair: a private key held exclusively within KMS and a public key that can be exported. They are used for digital signatures -- signing JWTs, verifying API payloads, code signing, and certificate operations. Supported algorithms include RSA (2048, 3072-bit), NIST P-256, secp256k1, and SM2.

### Protection Levels

KMS offers two protection levels:

- **SOFTWARE**: Key material is protected by software-based cryptographic modules within KMS. This is the default and covers the vast majority of use cases.
- **HSM**: Key material is generated, stored, and used exclusively within FIPS 140-2 Level 3 validated Hardware Security Modules. Required by some regulatory frameworks and recommended for the highest-value keys.

### Dedicated KMS

For organizations requiring complete isolation, Alibaba Cloud offers Dedicated KMS -- a single-tenant KMS instance with dedicated HSMs. Dedicated KMS supports additional key specs (AES-128, AES-192) and provides network isolation via VPC endpoints. The OpenMCF component does not expose `dkms_instance_id` in v1, keeping the scope focused on the shared KMS service that covers 99% of use cases.

## Deployment Methods

### Console

The KMS console at `kms.console.aliyun.com` allows manual key creation. You select the region, key spec, usage, and protection level, then click Create. The console is adequate for one-off keys but provides no version control, no reproducibility, and no integration with infrastructure-as-code workflows.

### Alibaba Cloud CLI

```bash
aliyun kms CreateKey \
  --KeySpec Aliyun_AES_256 \
  --KeyUsage ENCRYPT/DECRYPT \
  --ProtectionLevel SOFTWARE \
  --Description "Production encryption key"
```

The CLI provides scriptability but no state management. You must track key IDs externally and manually handle rotation, deletion protection, and tagging.

### Terraform

```hcl
resource "alicloud_kms_key" "main" {
  key_spec             = "Aliyun_AES_256"
  key_usage            = "ENCRYPT/DECRYPT"
  protection_level     = "SOFTWARE"
  automatic_rotation   = "Enabled"
  rotation_interval    = "365d"
  pending_window_in_days = 30
  deletion_protection  = "Enabled"
  description          = "Production encryption key"
}
```

Terraform provides declarative state management and drift detection. The `alicloud_kms_key` resource supports all key configuration options. Note that `automatic_rotation` and `deletion_protection` are string fields (`"Enabled"`/`"Disabled"`), not booleans -- a provider convention that the OpenMCF component abstracts into proper booleans.

### Pulumi (Go SDK)

```go
key, err := kms.NewKey(ctx, "main", &kms.KeyArgs{
    KeySpec:            pulumi.String("Aliyun_AES_256"),
    KeyUsage:           pulumi.String("ENCRYPT/DECRYPT"),
    ProtectionLevel:    pulumi.String("SOFTWARE"),
    AutomaticRotation:  pulumi.String("Enabled"),
    RotationInterval:   pulumi.String("365d"),
    PendingWindowInDays: pulumi.Int(30),
    DeletionProtection: pulumi.String("Enabled"),
    Description:        pulumi.String("Production encryption key"),
})
```

Pulumi provides the same declarative model as Terraform with the added benefit of general-purpose programming language constructs.

### OpenMCF

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudKmsKey
metadata:
  name: prod-encryption-key
  org: my-org
  env: production
spec:
  region: cn-hangzhou
  description: Production master encryption key
  automaticRotation: true
  rotationInterval: "365d"
  deletionProtection: true
  pendingWindowInDays: 30
```

OpenMCF wraps both Terraform and Pulumi implementations behind a unified KRM manifest. The component validates inputs at the proto level (CEL expressions, range constraints), applies standard tags, and exports outputs that downstream components can reference via `StringValueOrRef`.

## Design Decisions

### Bool vs. String for Enabled/Disabled Fields

The Alibaba Cloud provider represents `automatic_rotation` and `deletion_protection` as string fields with values `"Enabled"` and `"Disabled"`. The OpenMCF component uses `bool` instead, which is more natural for YAML manifests (`automaticRotation: true` vs. `automaticRotation: Enabled`). The IaC modules handle the conversion.

### 80/20 Field Scoping

The component exposes the fields that cover the vast majority of production use cases. Fields intentionally excluded from v1:

- `dkms_instance_id`: Dedicated KMS is an enterprise feature with separate pricing and provisioning
- `origin`: External key material (BYOK) is a niche use case for organizations migrating existing key material
- `policy`: Key-level IAM policies add complexity; most users manage access through RAM roles and policies
- `status`: Operational state management (enable/disable) is a lifecycle operation better handled outside IaC

### Deletion Protection Default

Deletion protection defaults to `false` to keep the minimal configuration simple (especially for development/testing). The documentation and presets strongly recommend enabling it for production keys, where accidental deletion would cause irrecoverable data loss.

### Pending Window Default

The default pending window of 30 days provides a generous recovery period. AWS KMS allows 7-30 days; Alibaba Cloud allows 7-366 days. The 30-day default balances safety (enough time to catch accidental deletions) with practicality (not waiting a year for key cleanup in development).

## Production Best Practices

### Always Enable Deletion Protection for Production Keys

A deleted KMS key means permanently encrypted data. There is no recovery. Enable `deletionProtection: true` and provide a clear `deletionProtectionDescription` for every production key.

### Enable Automatic Rotation

For symmetric encryption keys used in long-lived production systems, enable annual rotation (`rotationInterval: "365d"`). Rotation creates new key material while preserving the ability to decrypt data encrypted with previous versions. This limits the blast radius of a compromised key version.

### Use Separate Keys per Service

Rather than sharing a single key across RDS, OSS, and ECS, create dedicated keys for each service. This provides granular access control (each service only has access to its own key) and limits the impact of a key compromise.

### Tag Consistently

Use tags to track key ownership, purpose, and compliance requirements. The OpenMCF component automatically applies standard resource tags (name, kind, organization, environment) and merges user-specified tags.

## Provider Resource Reference

- **Terraform**: [`alicloud_kms_key`](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/kms_key)
- **Pulumi**: [`alicloud.kms.Key`](https://www.pulumi.com/registry/packages/alicloud/api-docs/kms/key/)
