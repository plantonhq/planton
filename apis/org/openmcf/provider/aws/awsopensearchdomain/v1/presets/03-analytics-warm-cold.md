# Analytics Domain with Warm + Cold Storage

This preset creates an analytics-optimized OpenSearch domain with 3 data nodes, 3 UltraWarm nodes, and cold storage enabled. Designed for log analytics, time-series data, SIEM, and any workload where older data should be retained at progressively lower cost while remaining queryable.

## When to Use

- Centralized log analytics (application logs, infrastructure logs, access logs)
- SIEM and security event management with long retention periods
- Time-series data (metrics, IoT telemetry) with hot/warm/cold lifecycle
- Workloads retaining 30-365 days of data where most queries target recent data

## Key Configuration Choices

- **3 data nodes (r6g.xlarge.search)** — 32 GiB RAM per node for caching hot data; handles active indexing and recent queries
- **3 UltraWarm nodes (ultrawarm1.medium.search)** — S3-backed storage at ~80% lower cost per GB; for data accessed infrequently (>7-14 days old)
- **Cold storage enabled** — Lowest-cost tier for archival data; detached from cluster compute; attach to warm when querying
- **gp3 200 GB with 6000 IOPS, 250 MiB/s** — High-throughput storage for heavy indexing (log ingestion rates of GB/hour)
- **3 dedicated masters** — Cluster stability under heavy indexing and migration operations
- **Log publishing** — Index slow logs, search slow logs, and application logs to CloudWatch for monitoring
- **FGAC enabled** — Role-based access control for multi-team environments
- **Auto-Tune enabled** — Automatic performance optimization critical for analytics workloads with variable patterns

## Data Lifecycle Pattern

Configure Index State Management (ISM) policies in OpenSearch Dashboards after deployment:

| Age | Tier | Action |
| --- | --- | --- |
| 0-14 days | Hot (data nodes) | Active indexing and frequent queries |
| 14-90 days | Warm (UltraWarm nodes) | Read-only, occasional queries |
| 90+ days | Cold (S3) | Archival, query on demand by attaching to warm |

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `analytics-search` | Domain name (3-28 chars, lowercase, hyphens) | Your naming convention |
| `<vpc-name>` | Name of the AwsVpc resource providing subnets | Your OpenMCF VPC manifest |
| `<security-group-name>` | Name of the AwsSecurityGroup allowing HTTPS (443) | Your OpenMCF security group manifest |
| `<master-password>` | Master user password (min 8 chars, mixed case, digit, special) | Generate a strong password |
| `<index-slow-logs-log-group-arn>` | CloudWatch Logs log group ARN for index slow logs | AWS CloudWatch console or pre-created log group |
| `<search-slow-logs-log-group-arn>` | CloudWatch Logs log group ARN for search slow logs | AWS CloudWatch console or pre-created log group |
| `<application-logs-log-group-arn>` | CloudWatch Logs log group ARN for application logs | AWS CloudWatch console or pre-created log group |

## Related Presets

- **01-single-node-dev** — Use for development and testing without warm/cold tiers
- **02-production-vpc** — Use for production search without tiered storage
