# Preset: Multi-Authentication with Full Logging

An MSK cluster demonstrating all three authentication methods and all three log destinations
enabled simultaneously. Useful for organizations with diverse client populations and
comprehensive audit requirements.

## When to Use

- Environments with mixed client types (AWS services, external applications, IoT devices)
- Compliance scenarios requiring multiple log archives (real-time + long-term)
- Teams migrating from on-premises Kafka and needing mTLS alongside IAM

## Configuration Highlights

- **Instance type**: `kafka.m5.large` (balanced compute for multi-protocol overhead)
- **Brokers**: 3 across 3 AZs
- **Authentication**: All three methods enabled simultaneously:
  - SASL/IAM (port 9098) for AWS-native services (Lambda, ECS, EKS)
  - SASL/SCRAM (port 9096) for external applications via Secrets Manager
  - mTLS (port 9094) for certificate-based authentication from legacy systems
- **Encryption**: TLS client-broker, in-cluster TLS
- **Logging**: All three destinations enabled:
  - CloudWatch Logs for real-time monitoring and alerting
  - Firehose for streaming to analytics pipelines
  - S3 for long-term audit retention

## Cost Estimate

Approximately $0.50/hr for 3 x kafka.m5.large brokers (~$360/month) plus EBS storage
and nominal CloudWatch/Firehose/S3 costs.

## Customization

- Remove unused authentication methods to reduce broker listener overhead
- Disable logging destinations not required by your compliance framework
- Scale to 6 or 9 brokers for higher throughput
