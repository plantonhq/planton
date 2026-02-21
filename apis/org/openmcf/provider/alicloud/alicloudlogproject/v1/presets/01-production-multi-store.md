# Production Multi-Store Log Project

This preset creates a production-ready Alibaba Cloud SLS project with three purpose-specific log stores: application logs (90-day retention), audit logs (365-day retention), and access logs (30-day retention). Each store is indexed for full-text search. The shard configuration is tuned for production throughput on the application log stream while keeping audit and access stores cost-efficient.

## When to Use

- Production environments that need separated log streams with distinct retention policies
- Workloads requiring audit trail preservation for compliance (365-day retention on audit-logs)
- Applications generating significant log volume that benefit from higher shard throughput on the primary store
- Foundation logging project referenced by ACK clusters, FC functions, or SAE applications

## Key Configuration Choices

- **Three log stores** (`app-logs`, `audit-logs`, `access-logs`) -- Separating log streams by purpose allows independent retention policies, access controls, and query scoping. This is the standard production logging pattern on Alibaba Cloud.
- **90-day app-log retention** (`retentionDays: 90`) -- Balances debugging needs with storage cost. Most application issues surface within 90 days.
- **365-day audit-log retention** (`retentionDays: 365`) -- Meets common compliance requirements for audit trail preservation. Increase to 3650 for permanent retention if required.
- **4 shards on app-logs** (`shardCount: 4`) -- Application logs are typically the highest-volume stream. Four shards provide ~16 MB/s write throughput. Auto-split (enabled by default) handles spikes beyond this baseline.
- **2 shards on audit/access stores** (`shardCount: 2`) -- These stores see lower write volume. Two shards provide ~8 MB/s throughput, sufficient for structured audit events and access records.
- **Full-text indexing enabled** (`enableIndex: true`) -- Logs without indexes cannot be searched in the SLS console or via API. Enabled on all stores to ensure immediate searchability.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-project-name>` | Globally unique SLS project name (3-63 chars, lowercase letters, digits, hyphens) | Choose a name following your organization's naming convention |

## Related Presets

- **02-development** -- Use instead for development and testing environments where cost and simplicity are prioritized over log separation and long retention
