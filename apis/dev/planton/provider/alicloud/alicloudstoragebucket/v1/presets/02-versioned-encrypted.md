# Versioned Encrypted Bucket

This preset creates a production-grade OSS bucket with zone-redundant storage (ZRS), object versioning, and AES256 server-side encryption at rest. Designed for workloads where data durability, recoverability, and compliance are priorities.

## When to Use

- Production data stores (application databases backups, user-uploaded content, platform artifacts)
- Environments subject to compliance requirements (data must be encrypted at rest)
- Workloads where accidental deletion or overwrite recovery is critical
- Multi-AZ deployments where surviving a full availability zone failure is required

## Key Configuration Choices

- **ZRS redundancy** (`redundancyType: ZRS`) -- replicates data across three availability zones within the region. Provides 99.995% availability (vs. 99.99% for LRS) and survives a full AZ outage. Costs approximately 1.5x more than LRS. This is immutable after creation.
- **Versioning enabled** (`versioningEnabled: true`) -- OSS preserves every version of every object. Deletes create a "delete marker" rather than removing data. Pair with lifecycle rules (preset 03) to control version retention costs.
- **AES256 encryption** (`sseAlgorithm: AES256`) -- all objects are encrypted at rest using OSS-managed keys. Zero additional configuration or key management overhead. For customer-managed keys, switch to `KMS` and provide a `kmsMasterKeyId`.
- **Tags** (`team`, `costCenter`) -- organizational metadata for cost attribution and operational ownership.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`) | Your deployment region strategy |
| `<globally-unique-bucket-name>` | Bucket name (3-63 chars, globally unique) | Choose a name with your org prefix (e.g., `myorg-prod-platform-data`) |
| `<your-org>` | Organization identifier | Your Planton organization name |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-cost-center>` | Cost center code | Your finance team |

## Related Presets

- **01-private-standard** -- use instead for development or non-critical workloads
- **03-archive-lifecycle** -- extend this preset with lifecycle rules for long-term data retention
