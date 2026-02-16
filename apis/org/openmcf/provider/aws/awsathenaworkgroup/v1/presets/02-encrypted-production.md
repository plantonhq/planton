# Preset: Encrypted Production Workgroup

A production-grade Athena workgroup with SSE_KMS encryption, cost controls,
and strict configuration enforcement.

## What This Configures

- SSE_KMS encryption for all query results (customer-managed key).
- 10 GB per-query scan limit to prevent runaway costs.
- Configuration enforcement enabled — queries cannot override settings.
- CloudWatch metrics enabled for operational visibility.
- Minimum encryption configuration enforced — all results at least SSE_S3.
- Pinned to Athena engine version 3 for consistent query behavior.

## When to Use

- Production analytics workloads with compliance requirements.
- Regulated industries (HIPAA, SOC 2, PCI DSS) requiring customer-managed
  encryption keys.
- Teams that need cost governance and consistent encryption.

## Before Deploying

1. **Replace the KMS key ARN** with your actual customer-managed key ARN.
2. **Adjust the scan limit** based on your dataset sizes. 10 GB is a reasonable
   starting point; increase for ETL workloads that scan large tables.
3. **Create the S3 bucket** at `prod-athena-results` with appropriate bucket
   policy and lifecycle rules.
