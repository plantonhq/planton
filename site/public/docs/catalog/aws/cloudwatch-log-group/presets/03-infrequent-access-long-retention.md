---
title: "Preset: Infrequent Access Long Retention"
description: "**Use case:** High-volume logs with long retention at reduced cost."
type: "preset"
rank: "03"
presetSlug: "03-infrequent-access-long-retention"
componentSlug: "cloudwatch-log-group"
componentTitle: "CloudWatch Log Group"
provider: "aws"
icon: "package"
order: 3
---

# Preset: Infrequent Access Long Retention

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

- **Ingestion**: $0.25/GB (50% cheaper than STANDARD)
- **Storage**: $0.0125/GB/month (~58% cheaper than STANDARD)
- **KMS**: $1.00/month per key + $0.03/10,000 API calls
