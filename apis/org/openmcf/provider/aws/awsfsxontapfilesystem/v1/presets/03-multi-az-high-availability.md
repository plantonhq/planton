# Multi-AZ High Availability FSx ONTAP

MULTI_AZ_2 deployment with automatic failover across two availability zones. 2 TiB SSD, 512 MB/s throughput. 7-day backups, customer-managed KMS encryption. Mission-critical configuration for workloads requiring high availability.

## When to Use

- Mission-critical workloads requiring automatic failover
- Production databases that cannot tolerate AZ downtime
- VMware Cloud on AWS datastores with HA requirements
- Compliance-sensitive workloads requiring multi-AZ resilience

## What It Configures

- **MULTI_AZ_2** — Latest generation multi-AZ deployment with automatic failover
- **2048 GiB SSD** — 2 TiB storage. Sub-millisecond latency
- **512 MB/s throughput** — Production-grade throughput tier
- **1 HA pair** — Fixed for multi-AZ (active in preferred subnet, standby in second)
- **Two subnets** — One per AZ. Must be in different availability zones
- **Preferred subnet** — Active file server placement
- **Endpoint IP range** — CIDR for floating IPs (seamless failover)
- **Customer-managed KMS** — Encryption at rest
- **7-day backups** — Daily automatic backups at 05:00 UTC

## What to Customize

- Replace placeholders: `name`, `id`, `org`, `env`, both subnet IDs, `endpoint_ip_address_range`, `sg-0123456789abcdef0`, and KMS key ARN
- Ensure `endpoint_ip_address_range` is a CIDR within your VPC that does not overlap with existing subnets
- Add `route_table_ids` if route tables need explicit routes to the file system
- Use `valueFrom` references to wire AwsVpc, AwsSecurityGroup, and AwsKmsKey
- Increase `storage_capacity_gib` or `throughput_capacity_per_ha_pair` as needed
