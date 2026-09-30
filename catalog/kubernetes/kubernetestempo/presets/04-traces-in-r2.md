# Traces in Cloudflare R2

Trace blocks in a Cloudflare R2 bucket, so traces survive the cluster that
wrote them: a rebuilt cluster, a moved region, or an incident review weeks
later still finds them. The bucket is referenced from a `CloudflareR2Bucket`
in the same project, and the key pair comes from Planton's secrets.

**When to use:** a monitoring hub whose data must outlive the cluster, or
any Tempo that should not pay egress to read its own traces.

**Prerequisites:** the `CloudflareR2Bucket`, and an R2 API token scoped to
that one bucket with Object Read & Write, its key pair stored as two
secrets.

**Retention:** give the bucket a lifecycle expiry longer than `retention`,
never shorter, or blocks vanish under Tempo's index.
