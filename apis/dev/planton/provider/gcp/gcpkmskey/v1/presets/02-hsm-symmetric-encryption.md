# Preset: HSM-Protected Symmetric Encryption Key

## When to Use

Use this preset when you need a customer-managed encryption key that is protected
by hardware security modules (HSM). This is required for compliance scenarios
that mandate FIPS 140-2 Level 3 certified key protection.

Common compliance frameworks that require or recommend HSM:
- **PCI DSS** -- payment card data encryption
- **HIPAA** -- healthcare data protection
- **FedRAMP** -- US federal government workloads
- **SOC 2 Type II** -- with enhanced security controls

## What It Creates

- A KMS key with purpose `ENCRYPT_DECRYPT`
- Algorithm: `GOOGLE_SYMMETRIC_ENCRYPTION`
- Protection level: `HSM` (Cloud HSM, FIPS 140-2 Level 3)
- Automatic rotation every 90 days

## Configuration

| Field | Value | Notes |
|-------|-------|-------|
| Purpose | ENCRYPT_DECRYPT | Default |
| Algorithm | GOOGLE_SYMMETRIC_ENCRYPTION | Explicit for clarity |
| Protection Level | HSM | Hardware security module |
| Rotation | 90 days | Creates new primary version automatically |

## Cost Considerations

HSM-protected keys are significantly more expensive than software-protected keys:
- HSM key versions have per-use and per-version costs
- Use HSM only when compliance requirements mandate it
- Software-protected keys are sufficient for most workloads

## How to Use

1. Ensure the key ring is in a region that supports Cloud HSM
2. Replace `<key-ring-id>` with your fully qualified key ring path
3. Replace `<your-key-name>` with a descriptive name (e.g., `compliance-cmek`)
4. Protection level is **immutable** -- choose carefully at creation time
