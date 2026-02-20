# Enterprise Encrypted RocketMQ

A mission-critical RocketMQ 5.x instance using the ultimate edition with encryption at rest, public internet access, and subscription billing. Designed for compliance-sensitive environments that handle confidential data and need the highest throughput ceiling. Includes TRANSACTION and DELAY topic types with a dead-letter queue for failed message investigation.

## When to Use

- Environments subject to compliance requirements (PCI-DSS, SOC 2, MLPS) that mandate encryption at rest
- Systems that require two-phase commit transaction messaging for distributed consistency
- Applications where external clients need internet access to produce or consume messages
- Long-running production deployments where subscription billing reduces cost over PayAsYouGo
- Workloads that need auto-scaling throughput to handle traffic spikes without manual intervention

## Key Configuration Choices

- **Ultimate edition** (`seriesCode: ultimate`) -- highest throughput ceiling and feature set; required for mission-critical workloads with strict latency and durability SLAs
- **Cluster HA** (`subSeriesCode: cluster_ha`) -- multi-node replication for zero message loss
- **Encryption at rest** (`storageEncryption: true`) -- encrypts all stored messages using a customer-managed KMS key; ForceNew, so plan this decision before deployment
- **Internet access enabled** (`internetInfo.enabled: true`) -- provides a public endpoint for external producers/consumers; traffic billed by usage (`payByTraffic`)
- **Subscription billing** (`paymentType: Subscription`, 12 months) -- significant cost savings over PayAsYouGo for long-running instances; auto-renews every 3 months
- **14-day message retention** (`messageRetentionTime: 336`) -- extended retention for audit trails, compliance replay, and incident investigation
- **Auto-scaling** (`autoScaling: true`) -- throughput scales automatically during traffic spikes, preventing message backlog
- **Message tracing** (`traceOn: true`) -- full message lifecycle visibility for debugging and compliance auditing
- **Transaction topics** (`messageType: TRANSACTION`) -- two-phase commit semantics for distributed transaction consistency across microservices
- **Delay topics** (`messageType: DELAY`) -- scheduled message delivery for time-sensitive workflows (e.g., order timeout, retry scheduling)
- **Fixed retry with dead-letter** -- transaction consumer uses fixed-interval retries (10 attempts) with explicit dead-letter routing for manual investigation of failed messages

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Alibaba Cloud region (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vpc-id>` | VPC ID for the instance | `AlicloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID within the VPC | `AlicloudVswitch` stack outputs |
| `<your-security-group-id>` | Security group for network access control | `AlicloudSecurityGroup` stack outputs |
| `<msg-process-spec>` | Throughput tier (e.g., `rmq.u2.4xlarge`) | [RocketMQ pricing page](https://www.alibabacloud.com/product/mq) |
| `<your-kms-key-id>` | KMS key ID for encryption at rest | `AlicloudKmsKey` stack outputs |
| `<your-vpc-cidr>` | VPC CIDR for IP whitelist (e.g., `172.16.0.0/12`) | Your VPC configuration |
| `<your-external-access-cidr>` | External IP range allowed to access the public endpoint | Your network security policy |
| `<your-compliance-standard>` | Compliance tag (e.g., `pci-dss`, `soc2`) | Your compliance requirement |
| `<your-transaction-topic>` | Topic name for transaction messages (e.g., `payment-transactions`) | Your application design |
| `<your-delay-topic>` | Topic name for delayed messages (e.g., `scheduled-notifications`) | Your application design |
| `<your-transaction-consumer-group>` | Consumer group for transaction topic (e.g., `GID_transaction_handler`) | Your application design |
| `<your-delay-consumer-group>` | Consumer group for delay topic (e.g., `GID_notification_sender`) | Your application design |
| `<your-dead-letter-topic>` | Dead-letter topic for failed transaction messages | Your application design |

## Related Presets

- **01-development-single-node** -- Minimal single-node instance for development and testing
- **02-production-ha** -- Professional edition with HA clustering for standard production workloads without encryption requirements
