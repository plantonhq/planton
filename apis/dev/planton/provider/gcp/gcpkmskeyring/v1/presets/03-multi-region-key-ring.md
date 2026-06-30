# Multi-Region Key Ring

This preset creates a KMS key ring in a multi-region location (e.g., `us`, `europe`, `asia`), providing high availability with automatic replication across all regions within the specified geography while maintaining data residency within that continental boundary.

## When to Use

- Workloads spread across multiple regions within the same continent
- Compliance requirements that mandate data stays within a geography (e.g., GDPR for `europe`)
- High availability requirements — keys remain accessible even if individual regions experience outages
- Disaster recovery scenarios where key availability must not depend on a single region

## Key Configuration Choices

- **Multi-region `us`** — keys replicated across all US regions. Change to `europe` or `asia` based on your compliance requirements.
- **Continental data residency** — unlike `global`, multi-region locations keep key material within the specified geography.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the key ring will be created | GCP Console or `GcpProject` outputs |
| `<your-key-ring-name>` | Permanent name for this key ring (1-63 chars, letters/digits/hyphens/underscores) | Choose a descriptive name (e.g., `us-compliance-keys`) |

## Multi-Region Options

| Location | Geography | Use Case |
|----------|-----------|----------|
| `us` | United States | US-only data residency |
| `europe` | European Union | GDPR compliance |
| `asia` | Asia Pacific | APAC data residency |

## Important

Key rings **cannot be deleted** from GCP. The name you choose is permanent within the project and location.

## Related Presets

- **01-regional-key-ring** — Key ring in a single region (strictest data residency)
- **02-global-key-ring** — Key ring accessible from all regions (no data residency)
