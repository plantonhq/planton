# Encrypted Queue with Large Messages and Consumer Groups

This preset creates an enterprise-grade OCI Queue with customer-managed KMS encryption, large message support (up to 512 KB), and consumer groups for partitioned consumption. The longer visibility timeout accommodates heavier processing workloads, and the consumer group configuration enables multiple independent consumer applications to process messages from the same queue without interfering with each other.

## When to Use

- Messaging workloads where message payloads exceed 128 KB (documents, serialized objects, batch records)
- Regulated environments requiring customer-managed encryption keys for data at rest
- Multi-consumer architectures where different services need independent consumption positions on the same queue
- Enterprise event buses where compliance mandates envelope encryption with keys you control and can rotate

## Key Configuration Choices

- **Customer-managed encryption** (`customEncryptionKeyId`) -- message content is encrypted using a KMS key you own. This enables key rotation policies, audit trails via OCI Vault, and compliance with regulations requiring customer-managed keys (PCI-DSS, HIPAA).
- **Large messages enabled** (`isLargeMessagesEnabled: true`) -- raises the per-message size limit from 128 KB to 512 KB. This maps to the `LARGE_MESSAGES` capability in the OCI API. Enable this when payloads include serialized documents, encoded images, or verbose structured data.
- **120-second visibility timeout** (`visibilityInSeconds: 120`) -- consumers have 2 minutes to process each message before it becomes visible again. Appropriate for workloads that involve downstream API calls, file processing, or database writes.
- **Consumer groups** (`consumerGroupConfig`) -- enables the `CONSUMER_GROUPS` capability, creating a primary consumer group with its own DLQ. Consumer groups allow multiple independent applications to consume from the queue with separate read positions, enabling fan-out patterns without message duplication.
- **3 delivery attempts** (`deadLetterQueueDeliveryCount: 3`) -- tighter retry budget than the standard preset. With larger messages and heavier processing, failures are more likely to be deterministic (malformed payload, schema mismatch), so fewer retries reduce wasted processing time before routing to the DLQ.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the queue will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<kms-key-ocid>` | OCID of the KMS key for message encryption | OCI Console > Security > Vault > Keys, or `OciKmsKey` status outputs (`keyId`) |

## Related Presets

- **01-standard-with-dlq** -- use instead for standard messaging workloads that do not require large messages, customer-managed encryption, or consumer groups
