# Logs in Cloudflare R2

Chunks, index and ruler state in a Cloudflare R2 bucket, so logs survive
the cluster that wrote them: a rebuilt cluster, a moved region, or an
incident review weeks later still finds them. The bucket is referenced from
a `CloudflareR2Bucket` in the same project, and the key pair comes from
Planton's secrets.

**When to use:** a monitoring hub whose logs must outlive the cluster, or
any Loki that should not pay egress to read its own logs.

**Prerequisites:** the `CloudflareR2Bucket`, and an R2 API token scoped to
that one bucket with Object Read & Write, its key pair stored as two
secrets.

**Retention:** give the bucket a lifecycle expiry longer than
`retention_period`, never shorter, or chunks vanish under Loki's index.
