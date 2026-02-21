# Development Log Project

This preset creates a minimal SLS project with a single log store using 7-day retention and one shard. This is the lowest-cost configuration for development and testing environments where log volume is low and long-term retention is unnecessary.

## When to Use

- Development and testing environments
- Quick prototyping where you need basic logging without production overhead
- Cost-sensitive workloads that do not require separate log streams or long retention
- Sandbox environments for validating SLS integrations before production deployment

## Key Configuration Choices

- **Single log store** (`app-logs`) -- One stream is sufficient for development. Add more stores by customizing this preset if needed.
- **7-day retention** (`retentionDays: 7`) -- Enough time to debug recent issues while keeping storage costs near zero. The default is 30 days; this preset reduces it to minimize cost.
- **1 shard** (`shardCount: 1`) -- Provides ~4 MB/s write throughput, more than sufficient for development traffic. Auto-split (enabled by default) handles unexpected bursts.
- **Full-text indexing enabled** (`enableIndex: true`) -- Keeps logs searchable even in development. The cost overhead of indexing is negligible at low volume.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`, `cn-shanghai`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-project-name>` | Globally unique SLS project name (3-63 chars, lowercase letters, digits, hyphens) | Choose a name following your organization's naming convention |

## Related Presets

- **01-production-multi-store** -- Use instead for production environments requiring separated log streams with distinct retention policies
