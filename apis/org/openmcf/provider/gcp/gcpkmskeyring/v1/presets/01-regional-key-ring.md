# Regional Key Ring

This preset creates a KMS key ring in a specific GCP region. It is the most common configuration — co-locating encryption keys with the workloads they protect for lowest latency and data residency compliance.

## When to Use

- Production workloads that require encryption keys in the same region as data
- GDPR, HIPAA, or other regulatory requirements mandating data residency
- BigQuery datasets, Cloud SQL instances, or Spanner databases that need CMEK in a specific region
- Standard key management setup for any regional GCP deployment

## Key Configuration Choices

- **Regional location** (`us-central1`) — change to match your workload region. Keys are stored exclusively in this region.
- **No multi-region replication** — keys exist in one region only. Use the multi-region preset if you need continental availability.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the key ring will be created | GCP Console or `GcpProject` outputs |
| `<your-key-ring-name>` | Permanent name for this key ring (1-63 chars, letters/digits/hyphens/underscores) | Choose a descriptive name (e.g., `prod-encryption`) |

## Important

Key rings **cannot be deleted** from GCP. The name you choose is permanent within the project and location.

## Related Presets

- **02-global-key-ring** — Key ring accessible from all regions (no data residency)
- **03-multi-region-key-ring** — Key ring replicated across a continent (high availability + data residency)
