# Multi-Region Production Accelerator

This preset creates a production-grade Global Accelerator that distributes TCP traffic across two AWS regions — `us-east-1` (60% traffic) and `eu-west-1` (40% traffic). It configures HTTP health checks on `/health` with aggressive 10-second intervals, enables flow logs to S3 for traffic analysis, and preserves client IP addresses at both ALB endpoints. This is the recommended starting point for production multi-region deployments.

## When to Use

- You operate a global application with users in North America and Europe
- You need automatic regional failover — if the US region becomes unhealthy, all traffic shifts to the EU region within seconds
- You want to gradually shift traffic between regions for blue/green or canary deployments by adjusting `trafficDialPercentage`
- You require traffic visibility through S3 flow logs for auditing, debugging, or compliance

## Key Configuration Choices

- **Two regional endpoint groups** (us-east-1 at 60%, eu-west-1 at 40%) — distributes traffic proportionally; adjust percentages to match your capacity and user distribution
- **HTTP health checks on `/health`** — validates application-level health, not just port reachability; ensure your application exposes this endpoint
- **10-second health check interval** with **threshold 5** — detects failures within 50 seconds; use 30-second intervals with threshold 3 (90 seconds) if cost or health check load is a concern
- **Flow logs enabled** — delivers traffic records to S3 for analysis with Athena or other tools; adds minimal cost but provides invaluable debugging data
- **Client IP preservation** — ALB endpoints see the original client IP in requests, enabling accurate access logging and geo-based application logic

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<flow-logs-s3-bucket-name>` | S3 bucket for flow log delivery | AWS S3 Console or `AwsS3Bucket` status outputs |
| `<us-east-1-alb-arn>` | ARN of the ALB in us-east-1 | AWS EC2 Console or `AwsAlb` status outputs |
| `<eu-west-1-alb-arn>` | ARN of the ALB in eu-west-1 | AWS EC2 Console or `AwsAlb` status outputs |

## Related Presets

- **01-basic-tcp-accelerator** — Use for single-region deployments or quick evaluations
- **03-gaming-udp-accelerator** — Use for UDP-based workloads requiring source IP affinity
