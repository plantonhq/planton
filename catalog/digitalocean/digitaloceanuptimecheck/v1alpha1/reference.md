# DigitalOceanUptimeCheck

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanUptimeCheckSpec models the digitalocean_uptime_check resource
surface plus its digitalocean_uptime_alert rows: an availability /
latency probe on an EXTERNAL endpoint, run from DigitalOcean's global
vantage regions, with alert rules delivered by email or Slack.

Alert rules cannot exist without their check and DigitalOcean's
standalone alert resource leaves the parent relationship mutable (a
corruption class where re-pointing an alert orphans it on the old
check), so the rules are composed HERE as rows and the modules create
one alert resource per row. Destroying the check destroys its alerts
with it.

## Example

```yaml
# Reference manifests for DigitalOceanUptimeCheck -- protovalidate-valid,
# embedded as the reference page's Example block, and the documents the
# offline tofu plans render. Two documents: a bare https check, and a
# check with latency + down alerts fanning out.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanUptimeCheck
metadata:
  name: homepage-check
spec:
  checkName: homepage
  target: https://www.example.com
  regions:
    - us_east
    - eu_west
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanUptimeCheck
metadata:
  name: api-check
spec:
  checkName: api
  target: https://api.example.com
  type: https
  regions:
    - us_east
    - us_west
    - eu_west
    - se_asia
  alerts:
    - alertName: api-slow
      type: latency
      threshold: 800
      comparison: greater_than
      period: 5m
      notifications:
        emails:
          - ops@example.com
    - alertName: api-down
      type: down_global
      period: 2m
      notifications:
        emails:
          - ops@example.com
        slack:
          - channel: "#alerts"
            url: $secret/slack-alerts-webhook
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.checkName` | `string` | yes |  |  |
| `spec.target` | `string` | yes |  |  |
| `spec.type` | `string` |  | `https` |  |
| `spec.regions` | `[]string` | yes |  |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.alerts` | `[]DigitalOceanUptimeCheckAlert` |  |  |  |
| `spec.alerts[].alertName` | `string` | yes |  |  |
| `spec.alerts[].type` | `string` | yes |  |  |
| `spec.alerts[].threshold` | `int32` |  |  |  |
| `spec.alerts[].comparison` | `string` |  |  |  |
| `spec.alerts[].period` | `string` | yes |  |  |
| `spec.alerts[].notifications` | `DigitalOceanUptimeCheckNotifications` | yes |  |  |
| `spec.alerts[].notifications.emails` | `[]string` |  |  |  |
| `spec.alerts[].notifications.slack` | `[]DigitalOceanUptimeCheckSlack` |  |  |  |
| `spec.alerts[].notifications.slack[].channel` | `string` | yes |  |  |
| `spec.alerts[].notifications.slack[].url` | `string` (sensitive) | yes |  |  |

## Field Details

### spec.checkName

`string` · required

Human-friendly name of the check, shown in the DigitalOcean console.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.target

`string` · required

The endpoint to probe. A URL for http/https checks (for example
"https://www.example.com") and a hostname or IP address for ping
checks; DigitalOcean enforces the pairing at request time.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.type

`string`

(Optional) The probe protocol. Unset defers to DigitalOcean's default,
https.

- default: `https`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ping","http","https"]}}

### spec.regions

`[]string` · required

The DigitalOcean vantage regions the probe runs from. Required here
although DigitalOcean can default it: the provider never reconciles a
defaulted value (omitting regions leaves every subsequent plan trying
to remove what the API chose), so the regions are always declared.
The value list is DigitalOcean's documented set -- the provider itself
does not validate it.

- rule: {"repeated":{"minItems":"1","items":{"string":{"in":["us_east","us_west","eu_west","se_asia"]}}}}

### spec.enabled

`bool` · optional (explicit presence)

(Optional) Whether the check is active. Unset defers to DigitalOcean's
default, which is enabled.

- default: `true`

### spec.alerts

`[]DigitalOceanUptimeCheckAlert`

(Optional) Alert rules on this check. Each row becomes its own alert
object on the check, addressable for notifications independently.

- rule: a latency alert requires threshold (the response-time bar in milliseconds) and comparison (greater_than or less_than)
- rule: an ssl_expiry alert requires threshold (days before the certificate expires; DigitalOcean accepts 0, which fires only on the day it expires)
- rule: an ssl_expiry alert must not set comparison -- DigitalOcean always evaluates it as less_than
- rule: down and down_global alerts must not set threshold or comparison -- DigitalOcean fixes them at 1 / less_than

### spec.alerts[].alertName

`string` · required

Human-friendly name of the alert rule.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.alerts[].type

`string` · required

What the alert watches. latency fires on response time (threshold in
milliseconds); down fires when the target is unreachable from a
vantage region; down_global fires when it is unreachable from ALL
regions; ssl_expiry fires when the certificate is within threshold
DAYS of expiring.

- rule: {"required":true,"string":{"in":["latency","down","down_global","ssl_expiry"]}}

### spec.alerts[].threshold

`int32` · optional (explicit presence)

The threshold the alert compares against: milliseconds for latency
(required), days before expiry for ssl_expiry (required). Must be left
unset for down and down_global -- DigitalOcean fixes theirs at 1 and
the modules send that value.

- rule: {"int32":{"gte":0}}

### spec.alerts[].comparison

`string`

How the measured value is compared against the threshold. Required for
latency (the only type where DigitalOcean honors it); must be left
unset for ssl_expiry, down, and down_global, where DigitalOcean always
evaluates less_than and the modules send that value. snake_case is
this API's spelling; monitor alerts spell the same concept CamelCase
(GreaterThan) -- the two are different DigitalOcean APIs and are
deliberately not unified.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["greater_than","less_than"]}}

### spec.alerts[].period

`string` · required

How long the condition must hold before the alert fires. Required for
every alert type: the provider's schema calls it optional, but
DigitalOcean's API rejects any alert without it ("missing required
field 'period'"), so the omission is rejected here rather than at apply.

- rule: {"required":true,"string":{"in":["2m","3m","5m","10m","15m","30m","1h"]}}

### spec.alerts[].notifications

`DigitalOceanUptimeCheckNotifications` · required

Where this alert's notifications are delivered. At least one channel
is required -- DigitalOcean rejects an alert that notifies nobody.

- rule: {"required":true}
- rule: at least one notification channel (emails or slack) must be specified

### spec.alerts[].notifications.emails

`[]string`

(Optional) Email addresses notifications are sent to. Every address
MUST belong to a verified member of the DigitalOcean team that owns the
check: the API rejects any other address at create time ("invalid
email"), so a shared inbox or an external on-call address has to be
invited to the team first. Monitor alerts enforce the same rule.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.alerts[].notifications.slack

`[]DigitalOceanUptimeCheckSlack`

(Optional) Slack channels notifications are posted to.

### spec.alerts[].notifications.slack[].channel

`string` · required

The Slack channel to post to (for example "#alerts").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.alerts[].notifications.slack[].url

`string` · required · sensitive

The Slack incoming-webhook URL. A credential: DigitalOcean's API does
not mark it sensitive, so it is marked sensitive here -- the platform
accepts only a managed-secret reference ($secret/<name>) for it, never
a literal URL, and the Pulumi module additionally encrypts it in stack
state. Terraform state stores every value in plain text, so on that
engine the protection is the state backend's own encryption.

- rule: {"required":true,"string":{"minLen":"1"}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanUptimeCheck, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.check_id` | `string` | UUID of the uptime check (the API identity, and the import id). |
| `status.outputs.alert_ids` | `map<string, string>` | UUIDs of the composed alert rows, keyed by "<row index>-<alert name>" -- the same key both provisioners address each alert resource by. Each row imports as "{check_id},{alert_id}", and this map is where the second half comes from, so a blind import needs no lookup. |

## See Also

- [Overview](../README.md)
