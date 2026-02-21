---
title: "Private Encrypted"
description: "This preset creates a production-grade OCI Stream Pool with a private endpoint (VCN-only access), customer-managed KMS encryption, auto-create disabled, maximum 7-day retention, and three streams:..."
type: "preset"
rank: "02"
presetSlug: "02-private-encrypted"
componentSlug: "stream-pool"
componentTitle: "Stream Pool"
provider: "oci"
icon: "package"
order: 2
---

# Private Encrypted

This preset creates a production-grade OCI Stream Pool with a private endpoint (VCN-only access), customer-managed KMS encryption, auto-create disabled, maximum 7-day retention, and three streams: events, commands, and dead-letter. The pool is accessible only from within the specified subnet and protected by a Network Security Group, ensuring data never traverses the public internet.

## When to Use

- Production event pipelines where streaming data must not leave the VCN
- Compliance environments requiring customer-managed encryption keys for data at rest and in transit within the pool
- Mission-critical architectures where accidental topic creation must be prevented and a dead-letter stream is needed for failed message handling
- High-throughput systems requiring higher partition counts for consumer parallelism

## Key Configuration Choices

- **Private endpoint** (`privateEndpointSettings`) -- the stream pool is accessible only from within the configured subnet. Kafka clients must be in the same subnet, a peered subnet, or connected via VPN/FastConnect. This eliminates public internet exposure entirely. The entire `privateEndpointSettings` block is immutable after creation (ForceNew).
- **NSG protection** (`nsgIds`) -- restricts which resources can connect to the stream pool's private endpoint. Configure ingress rules allowing TCP ports 9092 (Kafka) and 9093 (Kafka TLS) from application subnets.
- **KMS encryption** (`kmsKeyId`) -- encrypts all stream data with a customer-managed AES key. This provides key rotation control, access audit logging, and the ability to revoke access by disabling the key. Use an `OciKmsKey` created from the 01-aes-256-hsm-auto-rotation preset.
- **Auto-create disabled** (`autoCreateTopicsEnable: false`) -- prevents accidental topic creation from misconfigured producers. All streams must be explicitly defined in the `streams` list. This is the recommended setting for production where topic governance matters.
- **7-day retention** (`logRetentionHours: 168`) -- maximum allowed retention period. Provides a full week for consumers to process messages, handling extended outages, holiday weekends, and disaster recovery replay scenarios.
- **Events stream** (10 partitions, 168h) -- high partition count for maximum consumer parallelism in high-throughput event pipelines.
- **Commands stream** (5 partitions, 72h) -- moderate partition count with 3-day retention for command processing pipelines.
- **Dead-letter stream** (3 partitions, 168h) -- captures messages that failed processing after retry exhaustion. Maximum retention ensures failed messages are preserved for investigation and replay. Partition count is low since dead-letter throughput should be a small fraction of the main streams.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the stream pool will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<kms-key-ocid>` | OCID of the KMS encryption key for stream data | `OciKmsKey` status outputs (`keyId`), or OCI Console > Identity & Security > Vault > Keys |
| `<private-subnet-ocid>` | OCID of the private subnet for the stream pool endpoint | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<streaming-nsg-ocid>` | OCID of the NSG allowing Kafka traffic (ports 9092, 9093) | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |

## Related Presets

- **01-public-kafka-compatible** -- Use instead for development or staging where private networking and customer-managed encryption are not required
