# Logging Sink Writer

This preset grants `roles/storage.objectCreator` on one bucket to a logging sink's writer identity — the grant every bucket log export needs before Cloud Logging can write its first hourly batch. Both sides arrive by reference, so the grant deploys after the bucket and the sink without a dependency cycle.

## When to Use

- A `GcpLoggingSink` exports to a `GcpGcsBucket` (an audit archive, a compliance log store)
- Any bucket log export deployed in one chart, where the first batch must not be lost
- The grantee depends on the bucket, so the bucket's own `iam_members` cannot hold the grant

## Key Configuration Choices

- **`valueFrom` member** — the sink's `writer_identity` output is minted by Google; the reference keeps the grant on the current identity even if the sink is recreated
- **Write-only role** — `objectCreator` lets the sink add objects and never read or overwrite them
- **Not in `iam_members`** — the bucket cannot reference the sink that references it; never declare this pair in both places

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<bucket-resource-name>` | The Planton resource name of the destination GcpGcsBucket | Your GcpGcsBucket manifest's `metadata.name` |
| `<logging-sink-resource-name>` | The Planton resource name of the GcpLoggingSink exporting to the bucket | Your GcpLoggingSink manifest's `metadata.name` |

## Related Presets

- **02-prefix-scoped-writer** — Write access limited to one object prefix by an IAM condition
