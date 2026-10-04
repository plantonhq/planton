# CloudflareLogpushJob

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareLogpushJobSpec creates a Logpush job: a continuous delivery
pipeline that ships one Cloudflare log dataset (HTTP requests, firewall
events, Zero Trust sessions, audit logs, ...) to a destination you own --
R2, S3-compatible storage, Google Cloud Storage, Splunk, an HTTP endpoint,
and kin. Jobs live either on the account or on a single zone -- exactly one
scope must be set (zone datasets like http_requests need zone scope;
account datasets like audit_logs and gateway_* need account scope).

Two provider truths this spec teaches:
  - `dataset` is IMMUTABLE (changing it forces replacement), and so is the
    scope (account_id/zone_id force replacement).
  - `kind` is IMMUTABLE at the API but not in the plan: the provider plans
    a clean in-place update and Cloudflare rejects it with HTTP 400 at
    apply time (the provider's own immutable-fields test). Pick the job
    variant at creation. destination_conf, by contrast, updates in place
    fine -- the provider's own tests change it three times over.

Most destinations must prove ownership before a job may write to them:
Cloudflare drops a challenge file into the destination, and its token is
passed back as `ownership_challenge`. The optional
`generate_ownership_challenge` arm performs that first step from this same
kind. Destinations the account already owns (same-account R2) skip the
challenge entirely.

## Example

```yaml
# Complete example manifest for CloudflareLogpushJob.
# Ships HTTP request logs for one zone to an R2 bucket, issuing the
# destination ownership challenge in the same deploy (two resources: the
# job and the one-shot challenge).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareLogpushJob
metadata:
  name: http-logs-to-r2
spec:
  zone_id:
    value: "REPLACE_WITH_ZONE_ID"
  dataset: http_requests
  destination_conf:
    value: "r2://REPLACE_WITH_BUCKET/logs/{DATE}?account-id=REPLACE_WITH_ACCOUNT_ID&access-key-id=REPLACE_WITH_R2_ACCESS_KEY_ID&secret-access-key=REPLACE_WITH_R2_SECRET_ACCESS_KEY"
  name: http-logs-to-r2
  filter: '{"where":{"and":[{"key":"ClientRequestHost","operator":"eq","value":"www.example.com"}]}}'
  max_upload_bytes: 5000000
  max_upload_interval_seconds: 60
  output_options:
    output_type: ndjson
    timestamp_format: rfc3339
    field_names:
      - ClientIP
      - ClientRequestHost
      - ClientRequestURI
      - EdgeResponseStatus
      - RayID
      - EdgeStartTimestamp
  generate_ownership_challenge: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` |  |  |  |
| `spec.zoneId` | `string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.dataset` | `string` | yes |  |  |
| `spec.destinationConf` | `string \| valueFrom` (sensitive) | yes |  |  |
| `spec.name` | `string` |  |  |  |
| `spec.enabled` | `bool` |  |  |  |
| `spec.filter` | `string` |  |  |  |
| `spec.kind` | `string` |  |  |  |
| `spec.maxUploadBytes` | `int64` |  |  |  |
| `spec.maxUploadIntervalSeconds` | `int64` |  |  |  |
| `spec.maxUploadRecords` | `int64` |  |  |  |
| `spec.outputOptions` | `CloudflareLogpushJobOutputOptions` |  |  |  |
| `spec.outputOptions.outputType` | `string` |  |  |  |
| `spec.outputOptions.fieldNames` | `[]string` |  |  |  |
| `spec.outputOptions.timestampFormat` | `string` |  |  |  |
| `spec.outputOptions.sampleRate` | `double` |  |  |  |
| `spec.outputOptions.batchPrefix` | `string` |  |  |  |
| `spec.outputOptions.batchSuffix` | `string` |  |  |  |
| `spec.outputOptions.recordPrefix` | `string` |  |  |  |
| `spec.outputOptions.recordSuffix` | `string` |  |  |  |
| `spec.outputOptions.recordDelimiter` | `string` |  |  |  |
| `spec.outputOptions.recordTemplate` | `string` |  |  |  |
| `spec.outputOptions.fieldDelimiter` | `string` |  |  |  |
| `spec.outputOptions.mergeSubrequests` | `bool` |  |  |  |
| `spec.outputOptions.cve202144228` | `bool` |  |  |  |
| `spec.ownershipChallenge` | `string \| valueFrom` (sensitive) |  |  |  |
| `spec.generateOwnershipChallenge` | `bool` |  |  |  |

## Field Details

### spec.accountId

`string`

The Cloudflare account to create an account-scoped job in. Set this OR
zone_id, never both. 32-character hex account ID.

- rule: account_id must be a 32-character hex string

### spec.zoneId

`string | valueFrom`

The zone to create a zone-scoped job in. Set this OR account_id, never
both.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.dataset

`string` · required

The log dataset this job ships. Required here although the provider
defaults it (to http_requests) -- the dataset IS the job's identity, and
it is immutable: changing it replaces the job. The 35 values are the
provider's own list at v5.23.0. Zone datasets (http_requests,
firewall_events, dns_logs, nel_reports, spectrum_events,
page_shield_events, zaraz_events) need zone scope; the rest are
account-scoped. Several datasets are gated by plan or product
subscription (most notably the Enterprise Logpush surface) -- the gate is
Cloudflare's, surfaced at create time.

- rule: dataset must be one of the provider's logpush datasets (access_requests, audit_logs, audit_logs_v2, biso_user_actions, casb_findings, device_posture_results, dex_application_tests, dex_device_state_events, dlp_forensic_copies, dns_firewall_logs, dns_logs, email_security_alerts, email_security_post_delivery_events, firewall_events, gateway_dns, gateway_http, gateway_network, http_requests, ipsec_logs, magic_ids_detections, mcp_portal_logs, mnm_flow_logs, nel_reports, network_analytics_logs, page_shield_events, sinkhole_http_logs, spectrum_events, ssh_logs, turnstile_events, warp_config_changes, warp_toggle_changes, websocket_analytics, workers_trace_events, zaraz_events, zero_trust_network_sessions)
- rule: {"required":true}

### spec.destinationConf

`string | valueFrom` · required · sensitive

Where the logs go, as one opaque destination URI with any credentials
embedded -- exactly the string Cloudflare's API takes (for example
"r2://bucket/{DATE}?account-id=...&access-key-id=...&secret-access-key=..."
or "https://splunk.example.com/services/collector/raw?channel=...").
Deliberately NOT decomposed into typed parts: the URI grammar differs per
destination product and is Cloudflare's contract to evolve. SENSITIVE:
the string routinely embeds access keys -- provide a managed-secret
reference and the platform resolves it just-in-time at deploy. Updates
in place; a NEW destination must re-prove ownership (see
ownership_challenge below).

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.name

`string`

Display name for the job, shown in the dashboard's Logpush list.

### spec.enabled

`bool` · optional (explicit presence)

Whether the job ships logs. Cloudflare's own default is DISABLED (a
job created without this flag sits idle) -- a declared job is meant to
ship, so the modules send true when unset. Set false explicitly to
pause delivery without deleting the job.

### spec.filter

`string`

Filter narrowing which log lines are shipped, as Cloudflare's JSON
filter expression (for example
{"where":{"and":[{"key":"ClientRequestHost","operator":"eq","value":"example.com"}]}}).
Empty ships every line of the dataset.

### spec.kind

`string`

The job variant. Empty is a normal batching Logpush job; "edge" is an
Instant Logs job (live tail over WebSocket, http_requests dataset only).
IMMUTABLE at the API without a plan-time guard: changing it plans a
clean in-place update that Cloudflare rejects with HTTP 400 -- pick the
variant at creation (see the message comment).

- rule: kind must be empty (normal logpush) or edge (instant logs)

### spec.maxUploadBytes

`int64` · optional (explicit presence)

Upload batch size ceiling in bytes. 0 (or unset) lets Cloudflare pick;
otherwise 5 MB to 1 GB. The API may still deliver smaller batches.

- rule: max_upload_bytes must be 0 or between 5000000 (5 MB) and 1000000000 (1 GB)

### spec.maxUploadIntervalSeconds

`int64` · optional (explicit presence)

Upload interval ceiling in seconds. 0 (or unset) lets Cloudflare pick;
otherwise 30 to 300 seconds.

- rule: max_upload_interval_seconds must be 0 or between 30 and 300

### spec.maxUploadRecords

`int64` · optional (explicit presence)

Upload batch record-count ceiling. 0 (or unset) lets Cloudflare pick;
otherwise 1,000 to 1,000,000 records.

- rule: max_upload_records must be 0 or between 1000 and 1000000

### spec.outputOptions

`CloudflareLogpushJobOutputOptions`

How each shipped record is structured: format, field list, delimiters,
and sampling. When set, it fully replaces the deprecated logpull-era
options.

### spec.outputOptions.outputType

`string`

Record format: ndjson (newline-delimited JSON, the default) or csv.

- rule: output_type must be ndjson or csv

### spec.outputOptions.fieldNames

`[]string`

Which log fields each record carries, in order. Empty ships the
dataset's default field set.

### spec.outputOptions.timestampFormat

`string`

Timestamp encoding used in records.

- rule: timestamp_format must be one of unixnano, unix, rfc3339, rfc3339ms, rfc3339ns

### spec.outputOptions.sampleRate

`double` · optional (explicit presence)

Fraction of records shipped, 0 to 1 (1 ships everything; 0.1 ships a
10% sample).

- rule: {"double":{"lte":1,"gte":0}}

### spec.outputOptions.batchPrefix

`string`

String prepended to each upload batch.

### spec.outputOptions.batchSuffix

`string`

String appended to each upload batch.

### spec.outputOptions.recordPrefix

`string`

String prepended to each record.

### spec.outputOptions.recordSuffix

`string`

String appended to each record.

### spec.outputOptions.recordDelimiter

`string`

Delimiter inserted between records (mutually exclusive with
record_suffix at the API).

### spec.outputOptions.recordTemplate

`string`

Go text/template applied per record, overriding the plain field list
(advanced formatting).

### spec.outputOptions.fieldDelimiter

`string`

Delimiter between fields of a csv record.

### spec.outputOptions.mergeSubrequests

`bool` · optional (explicit presence)

Whether Workers subrequest logs are merged into their parent request's
record (workers_trace_events dataset).

### spec.outputOptions.cve202144228

`bool` · optional (explicit presence)

Redact CVE-2021-44228 (log4j JNDI) exploit strings from shipped logs
before delivery.

### spec.ownershipChallenge

`string | valueFrom` · sensitive

The ownership-challenge token proving you control the destination:
Cloudflare drops a challenge file into the destination and the token is
its content. Fetch the token from that file and set it here (or start
with the generate_ownership_challenge arm below to make Cloudflare drop
the file). Same-account R2 destinations skip the challenge. SENSITIVE
and write-only: the API never returns it.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.generateOwnershipChallenge

`bool`

When true, the module also performs the challenge-issuing step for this
job's destination_conf: Cloudflare drops the challenge file into the
destination, and the outputs surface the file's name
(ownership_challenge_filename). MANUAL STEP BY DESIGN: fetch the token
from that file at the destination, then set ownership_challenge above --
Cloudflare deliberately makes the proof pass through storage you
control, so no API (and no IaC engine) can read the token for you. The
issuing step is one-shot at Cloudflare: it cannot be read back, updated,
deleted, or imported; destroying the resource only forgets it.

## Validation Rules

- `spec.scope_exactly_one`: set exactly one of account_id (account-scoped dataset) or zone_id (zone-scoped dataset)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareLogpushJob, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.job_id` | `string` | The Cloudflare-assigned numeric job ID, in string form. |
| `status.outputs.account_id` | `string` | The account the job lives in (account-scoped jobs). |
| `status.outputs.zone_id` | `string` | The zone the job lives in (zone-scoped jobs). |
| `status.outputs.ownership_challenge_filename` | `string` | Name of the challenge file Cloudflare dropped into the destination (only when the generate_ownership_challenge arm ran). Fetch the token from this file and set spec.ownership_challenge. |
| `status.outputs.ownership_challenge_message` | `string` | Message accompanying the issued challenge, when Cloudflare returns one. |
| `status.outputs.ownership_challenge_valid` | `bool` | Whether Cloudflare reported the destination valid when issuing the challenge. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
