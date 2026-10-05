# Cloudflare Logpush Job

Deploys a Cloudflare Logpush job that continuously ships one log dataset — HTTP requests, firewall events, Zero Trust sessions, audit logs, and 30+ more — to a destination you own, such as R2, S3-compatible storage, Google Cloud Storage, Splunk, or an HTTPS endpoint. Jobs live either on the account or on a single zone, and the component can also run the destination ownership handshake that most destinations require before Cloudflare will write to them. Delivery starts as soon as the job exists: unlike Cloudflare's own default, a job declared here ships logs unless you explicitly pause it.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Logpush Job** -- one `cloudflare_logpush_job` on the account or zone scope, shipping the chosen dataset to the configured destination
- **Logpush Ownership Challenge** -- created only when `generateOwnershipChallenge` is `true`; makes Cloudflare drop a challenge file into the destination so you can prove you control it. The challenge is one-shot at Cloudflare: it cannot be read back, updated, or imported, and destroying it only removes it from state

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Logs Write on the target account or zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **A Logpush entitlement** -- most datasets are gated by plan or product subscription (notably the Enterprise Logpush surface); the gate is Cloudflare's and surfaces at create time. Measured: below Enterprise the create fails with 403 code 1004 "exceeded max jobs allowed" even on a zone with zero jobs -- the plan's job quota is zero, not a real quota overrun.
- **A destination you control** with write credentials -- the credentials ride inside `destinationConf` as one opaque URI. Same-account R2 buckets skip the ownership handshake entirely; every other destination must prove ownership via `ownershipChallenge`.
- **The matching scope** -- zone datasets (`http_requests`, `firewall_events`, `dns_logs`, `nel_reports`, `spectrum_events`, `page_shield_events`, `zaraz_events`) need `zoneId`; everything else needs `accountId`. A mismatch fails at the Cloudflare API.

## Deploy

### Console

Open the deployment store, find **Cloudflare Logpush Job**, and click **Deploy**. The creation wizard walks you through the scope (account or zone), the dataset, the destination URI, and the delivery-shaping options (filter, batching ceilings, output format). Start from the **HTTP requests to R2** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLogpushJob
metadata:
  name: http-logs-to-r2
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "023e105f4ecef8ad9ca31a8372d0c353"
  dataset: http_requests
  destinationConf:
    value: "r2://web-logs/http/{DATE}?account-id=0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d&access-key-id=R2_ACCESS_KEY_ID&secret-access-key=R2_SECRET_ACCESS_KEY"
  name: http-logs-to-r2
  maxUploadIntervalSeconds: 60
```

```shell
planton apply -f logpush-job.yaml
```

This creates an enabled zone-scoped job shipping the zone's HTTP request logs to a same-account R2 bucket in one-minute batches — same-account R2 needs no ownership proof, so logs flow after this single apply. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the zone scope to a zone managed in the same InfraPipeline:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  dataset: http_requests
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then provisions the Logpush job against it.

## Key Configuration

These are the most important decisions when configuring a Logpush job. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Dataset and scope are the job's identity — and immutable.** Changing `dataset`, `accountId`, or `zoneId` replaces the job. Pick the dataset deliberately, and put it on the right scope: zone datasets need `zoneId`, account datasets (audit logs, all the `gateway_*` datasets) need `accountId`. The spec enforces exactly one scope, precisely because the provider silently prefers the account when both are set.

**The job variant is locked at creation, and the plan will not warn you.** `kind` is empty for a normal batching job or `edge` for Instant Logs (live tail, `http_requests` only). The provider carries no replacement guard for it: a change plans as a clean in-place update, then Cloudflare rejects the apply with HTTP 400. Changing the variant means delete and recreate.

**`destinationConf` is one opaque, sensitive URI.** It carries the destination and its credentials exactly as Cloudflare's API takes them, so provide it as a managed-secret reference and the platform resolves it just-in-time at deploy. It updates in place fine — but a new destination must re-prove ownership, and rotating credentials inside the URI counts as an update, so rotate deliberately rather than as a side effect of an unrelated apply.

**Ownership proof passes through your storage, on purpose.** Deploy once with `generateOwnershipChallenge: true`; Cloudflare drops a challenge file into the destination and the outputs report its name. Read the token out of that file yourself — no API can do it for you, because reading the file is what proves control — then set `ownershipChallenge` and apply again. Same-account R2 skips the handshake entirely.

**A declared job ships logs.** Cloudflare's own default creates jobs disabled; this component sends `enabled: true` when you say nothing, because a declared-but-idle job is the easy accident. Set `enabled: false` explicitly to keep a paused job on file.

**Bound the volume before you turn it on.** `http_requests` on a busy zone is a firehose, and the bill lands on your destination. `filter` (Cloudflare's JSON filter expression) narrows which lines ship, and `outputOptions.sampleRate` ships a fraction (0.1 is a tenth). Both are cheaper to set now than to discover on next month's storage invoice. The `maxUploadBytes` / `maxUploadIntervalSeconds` / `maxUploadRecords` ceilings shape batch cadence, not volume.

**Deleting the job is silent.** Destroy stops delivery immediately, but the objects already shipped remain — a dashboard glance looks fine while nothing new arrives. If a job matters for audit or compliance, pair it with a Cloudflare Notification Policy on the failing-Logpush-job alert so Cloudflare tells you when delivery stops.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `job_id` | The Cloudflare-assigned numeric job ID, in string form | Dashboard lookup, API calls against the job |
| `ownership_challenge_filename` | Name of the challenge file Cloudflare dropped into the destination (challenge arm only) | Locating the file to fetch the token for `ownershipChallenge` |
| `ownership_challenge_valid` | Whether Cloudflare reported the destination valid when issuing the challenge | Confirming the destination URI before the second apply |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Traffic analytics to R2** -- One zone's HTTP request logs into a same-account R2 bucket with an explicit `fieldNames` list instead of the dataset's default. Same-account R2 skips the ownership handshake, so this shape deploys in one pass; trim the field list to what your queries actually read — fewer fields is less storage and faster scans. Start from the **HTTP requests to R2** preset.

**Audit logs to a SIEM** -- Account-scoped `audit_logs_v2` streamed to an external HTTPS collector every 30 seconds, so the record of who changed what survives independently of the Cloudflare account itself. An HTTPS destination must prove ownership: deploy once with `generateOwnershipChallenge: true`, fetch the token from the posted file, then apply with `ownershipChallenge` filled in. Start from the **Audit logs to a SIEM** preset.

**Sampled firehose** -- A high-volume dataset shipped with `outputOptions.sampleRate: 0.1` and a `filter` restricted to the hostnames or status classes you actually investigate — the shape that keeps a busy zone's logs affordable while preserving enough signal for debugging.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the scope for every zone dataset; wire `zoneId` via ValueFromRef.
- [**Cloudflare R2 Bucket**](/infra-catalog/cloudflare-r2-bucket) -- the usual same-account destination, and the only kind that skips the ownership handshake.
- [**Cloudflare Notification Policy**](/infra-catalog/cloudflare-notification-policy) -- alerts when a job starts failing or gets disabled, closing the silent-delete gap.
