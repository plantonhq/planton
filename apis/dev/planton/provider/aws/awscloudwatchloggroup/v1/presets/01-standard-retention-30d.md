# Preset: Standard 30-Day Retention

**Use case:** General-purpose application logging with a sensible retention period.

This is the most common pattern — a STANDARD class log group that retains log events for 30 days before automatic deletion. Suitable for development, staging, and most production workloads where you don't need long-term log retention.

## What You Get

- A STANDARD class CloudWatch Log Group
- 30-day retention (log events deleted after 30 days)
- Default AWS encryption (SSE-CWL)
- Outputs: `log_group_arn`, `log_group_name`

## When to Use

- Application logs in dev/staging environments
- Service logs where you primarily use real-time monitoring
- Logs where you have a separate archival system (e.g., S3 via subscription filter)
- General operational logging

## Cost

- **Ingestion**: $0.50/GB
- **Storage**: $0.03/GB/month (first 30 days — then automatically deleted)
