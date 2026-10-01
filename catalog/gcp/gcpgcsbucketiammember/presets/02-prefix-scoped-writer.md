# Prefix-Scoped Object Writer

This preset grants `roles/storage.objectCreator` on one bucket to a service account, limited by an IAM condition to object names under one prefix. Several uploaders can share a bucket while each writes only into its own folder.

## When to Use

- Several producers share one landing bucket and must not write into each other's prefixes
- An uploader that should never be able to read, and should only write where it is told
- The grantee is created from the bucket, so the bucket's own `iam_members` cannot hold the grant (otherwise prefer `iam_members` with the same condition)

## Key Configuration Choices

- **Object-name condition** — `resource.name.startsWith(...)` scopes the grant to one prefix; the bucket's cloud name appears literally in the expression
- **Uniform bucket-level access required** — Cloud Storage accepts conditional role bindings only on buckets with uniform bucket-level access
- **Condition is part of the grant's identity** — changing the prefix replaces the grant

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<bucket-resource-name>` | The Planton resource name of the GcpGcsBucket | Your GcpGcsBucket manifest's `metadata.name` |
| `<service-account-resource-name>` | The uploader's GcpServiceAccount | Your GcpServiceAccount manifest's `metadata.name` |
| `<bucket-name>` | The bucket's cloud name, used inside the condition expression | The GcpGcsBucket's `bucket_name` output |
| `<object-prefix>` | The folder the uploader may write under, e.g. `ingest/team-a` | Your data layout |

## Related Presets

- **01-logging-sink-writer** — The write-only grant a logging sink needs on its destination bucket
