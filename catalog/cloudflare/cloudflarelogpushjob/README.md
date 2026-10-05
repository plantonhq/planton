# Cloudflare Logpush Job

## Overview

`CloudflareLogpushJob` creates a Logpush job: a continuous delivery pipeline that ships one Cloudflare log dataset (HTTP requests, firewall events, Zero Trust sessions, audit logs, and 30 more) to a destination you own -- R2, S3-compatible storage, Google Cloud Storage, Splunk, Datadog, or any HTTPS endpoint. A plain CRUD object -- real create, update, delete.

## Key Features

- **35 datasets** -- the provider's full list at the pinned version, enum-walled; zone datasets (http_requests, firewall_events, dns_logs, nel_reports, spectrum_events, page_shield_events, zaraz_events) and account datasets (audit logs, all the gateway and Zero Trust families, network analytics, Workers traces) alike
- **Dual scope** -- one job lives on the account or on a single zone, matching its dataset's scope; exactly one is set
- **Record shaping** -- output format (ndjson or csv), explicit field lists, timestamp encoding, sampling, delimiters, prefixes/suffixes, and log4j-exploit redaction
- **Upload tuning** -- batch size, interval, and record-count ceilings, each either Cloudflare-chosen (0) or inside the provider's accepted range
- **The ownership handshake, modeled** -- an optional arm makes Cloudflare drop its challenge file into the destination and surfaces the file's name, so the one manual step (reading the token out of your own bucket) is the only thing left to do

## Use Cases

**Ideal for:**

- Shipping HTTP request logs into a SIEM or data lake for security analytics
- Exporting Zero Trust session and Gateway logs for audit retention
- Streaming audit logs off-platform so account changes are recorded independently

**Not ideal for:**

- Interactive log tailing -- set `kind: edge` for Instant Logs, or use the dashboard's live view
- Metrics or analytics rollups -- Logpush ships raw log records, not aggregates

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `dataset` | string | Yes | Which log dataset ships (35 provider values, enum-walled). Immutable -- changing it replaces the job. |
| `destination_conf` | string reference | Yes | The destination URI with credentials embedded, exactly as Cloudflare's API takes it. Sensitive. |
| `account_id` or `zone_id` | string / reference | Exactly one | Account scope for account datasets, zone scope for zone datasets. |

### Key Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Display name in the dashboard's Logpush list. |
| `enabled` | bool | Whether logs ship. Defaults to **true** here (Cloudflare's own default is disabled). |
| `filter` | string | Cloudflare JSON filter expression narrowing which records ship. |
| `kind` | string | Empty for a normal batching job, `edge` for Instant Logs. Locked at creation: the API rejects changing it (HTTP 400) even though the plan looks clean. |
| `max_upload_bytes` | int64 | 0 (Cloudflare picks) or 5 MB–1 GB. |
| `max_upload_interval_seconds` | int64 | 0 (Cloudflare picks) or 30–300. |
| `max_upload_records` | int64 | 0 (Cloudflare picks) or 1,000–1,000,000. |
| `output_options` | object | Record format, field list, timestamp encoding, sampling, delimiters. |
| `ownership_challenge` | string reference | The token proving destination ownership. Sensitive, write-only. |
| `generate_ownership_challenge` | bool | Also perform the challenge-issuing step (one-shot; see GUIDE.md). |

### Outputs

| Field | Description |
|-------|-------------|
| `job_id` | The Cloudflare-assigned numeric job id, in string form |
| `account_id` / `zone_id` | The scope the job lives in |
| `ownership_challenge_filename` | Name of the challenge file dropped into the destination (challenge arm only) |
| `ownership_challenge_message` | Message accompanying the issued challenge |
| `ownership_challenge_valid` | Whether Cloudflare found the destination valid when issuing |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLogpushJob
metadata:
  name: http-logs-to-r2
spec:
  zone_id:
    valueFrom:
      kind: CloudflareDnsZone
      name: example-com
      fieldPath: status.outputs.zone_id
  dataset: http_requests
  destination_conf:
    valueFrom:
      kind: SecretsManagerSecret
      name: logpush-r2-destination
      fieldPath: status.outputs.secret_value
  name: http-logs-to-r2
  max_upload_interval_seconds: 60
  output_options:
    output_type: ndjson
    timestamp_format: rfc3339
    field_names:
      - ClientIP
      - ClientRequestHost
      - EdgeResponseStatus
      - RayID
```

## Prerequisites

- **A Logpush entitlement** -- most datasets require an Enterprise plan (see cost.yaml). Below Enterprise the create fails with 403 code 1004 "exceeded max jobs allowed" even on a zone with zero jobs -- the plan's job quota is zero; the message is not a real quota overrun.
- **A destination you control**, with its write credentials
- **A Cloudflare API token** with the Logs Write permission for the chosen scope

## Destroy Semantics

Real delete. Deleting the job stops log delivery immediately and silently -- already-shipped objects stay in the destination. The folded ownership challenge is never deleted at Cloudflare (see GUIDE.md).

## Related Kinds

- [Cloudflare DNS Zone](/docs/catalog/cloudflare/cloudflarednszone) -- the scope for zone datasets
- [Cloudflare R2 Bucket](/docs/catalog/cloudflare/cloudflarer2bucket) -- the usual same-account destination
- [Cloudflare Notification Policy](/docs/catalog/cloudflare/cloudflarenotificationpolicy) -- alerts when a Logpush job starts failing

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
