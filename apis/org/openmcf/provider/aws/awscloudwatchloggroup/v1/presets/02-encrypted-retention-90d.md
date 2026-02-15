# Preset: Encrypted 90-Day Retention

**Use case:** Production application logging with KMS encryption and 90-day retention for compliance.

This pattern provides customer-managed encryption and a 90-day retention window — suitable for production workloads subject to SOC2, HIPAA, or PCI-DSS requirements where log data must be encrypted at rest and retained for a meaningful period.

## What You Get

- A STANDARD class CloudWatch Log Group
- 90-day retention (quarterly log retention)
- Customer-managed KMS encryption (requires AwsKmsKey resource)
- Outputs: `log_group_arn`, `log_group_name`

## When to Use

- Production application logs requiring encryption at rest
- Compliance environments (SOC2, HIPAA, PCI-DSS)
- Logs that need cross-account access control via KMS key policy
- Audit-sensitive workloads

## Prerequisites

- An AwsKmsKey resource deployed in the same environment
- The KMS key policy must grant `logs.amazonaws.com` the required permissions

## Cost

- **Ingestion**: $0.50/GB
- **Storage**: $0.03/GB/month (for up to 90 days)
- **KMS**: $1.00/month per key + $0.03/10,000 API calls
