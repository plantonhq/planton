# Infrequent Access Long Retention

**Use case:** High-volume logs with long retention at reduced cost.

This pattern uses the INFREQUENT_ACCESS class (~50% cheaper storage) with 1-year retention and KMS encryption. Ideal for VPC flow logs, CDN access logs, compliance archives, and any high-volume log data that is written frequently but queried rarely.

## What You Get

- An INFREQUENT_ACCESS class CloudWatch Log Group
- 365-day retention (1 year)
- Customer-managed KMS encryption (requires AwsKmsKey resource)
- Outputs: `log_group_arn`, `log_group_name`

## When to Use

- VPC flow logs (high volume, rarely queried)
- CDN or load balancer access logs
- Compliance archives requiring 1-year retention
- Security event logs for forensic analysis
- Any log data with high write volume and low read frequency

## Trade-offs

INFREQUENT_ACCESS class does **not** support:
- Metric filters
- Subscription filters
- Contributor Insights
- Live Tail

It **does** support:
- Logs Insights queries
- Managed ingestion to S3

## Prerequisites

- An AwsKmsKey resource deployed in the same environment

## Cost

- **Ingestion**: billed per GB written, at roughly half the STANDARD class rate
- **Storage**: billed per GB-month, at well under half the STANDARD class rate — the class choice is the cost cliff here
- **KMS**: the customer-managed key adds a flat monthly per-key charge plus per-API-call usage

The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awscloudwatchloggroup.yaml` — computed from the pinned price book, never hand-typed here.
