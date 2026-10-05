# CloudflareZeroTrustDevicePostureRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustDevicePostureRuleSpec creates a device posture rule: a
health check (disk encrypted? OS current? EDR agent healthy?) that WARP
evaluates on enrolled devices and that Access and Gateway policies can
then require before admitting a device. A plain CRUD object -- real
create, update, delete.

The rule's `type` selects WHICH check runs, and `input` carries that
check's parameters. Which input fields a given type reads is owned by
Cloudflare's API (23 types x ~36 input fields is an API-side pairing);
each input field's comment names the check families that use it.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustDevicePostureRule.
# Creates an OS-version check requiring macOS 14.4.1 or newer, polled every
# five minutes.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustDevicePostureRule
metadata:
  name: macos-current
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: macos-current
  type: os_version
  description: laptops must run macOS 14.4.1 or newer
  schedule: 5m
  expiration: 1h
  match:
    - platform: mac
  input:
    version: 14.4.1
    operator: ">="
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.type` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.expiration` | `string` |  |  |  |
| `spec.schedule` | `string` |  |  |  |
| `spec.match` | `[]CloudflareZeroTrustDevicePostureRuleMatch` |  |  |  |
| `spec.match[].platform` | `string` | yes |  |  |
| `spec.input` | `CloudflareZeroTrustDevicePostureRuleInput` |  |  |  |
| `spec.input.operatingSystem` | `string` |  |  |  |
| `spec.input.path` | `string` |  |  |  |
| `spec.input.exists` | `bool` |  |  |  |
| `spec.input.sha256` | `string` |  |  |  |
| `spec.input.thumbprint` | `string` |  |  |  |
| `spec.input.id` | `string` |  |  |  |
| `spec.input.domain` | `string` |  |  |  |
| `spec.input.operator` | `string` |  |  |  |
| `spec.input.version` | `string` |  |  |  |
| `spec.input.osDistroName` | `string` |  |  |  |
| `spec.input.osDistroRevision` | `string` |  |  |  |
| `spec.input.osVersionExtra` | `string` |  |  |  |
| `spec.input.enabled` | `bool` |  |  |  |
| `spec.input.checkDisks` | `[]string` |  |  |  |
| `spec.input.requireAll` | `bool` |  |  |  |
| `spec.input.certificateId` | `string` |  |  |  |
| `spec.input.cn` | `string` |  |  |  |
| `spec.input.checkPrivateKey` | `bool` |  |  |  |
| `spec.input.extendedKeyUsage` | `[]string` |  |  |  |
| `spec.input.locations` | `CloudflareZeroTrustDevicePostureRuleCertificateLocations` |  |  |  |
| `spec.input.locations.paths` | `[]string` |  |  |  |
| `spec.input.locations.trustStores` | `[]string` |  |  |  |
| `spec.input.subjectAlternativeNames` | `[]string` |  |  |  |
| `spec.input.updateWindowDays` | `int64` |  |  |  |
| `spec.input.complianceStatus` | `string` |  |  |  |
| `spec.input.connectionId` | `string` |  |  |  |
| `spec.input.lastSeen` | `string` |  |  |  |
| `spec.input.os` | `string` |  |  |  |
| `spec.input.overall` | `string` |  |  |  |
| `spec.input.sensorConfig` | `string` |  |  |  |
| `spec.input.state` | `string` |  |  |  |
| `spec.input.versionOperator` | `string` |  |  |  |
| `spec.input.authState` | `[]string` |  |  |  |
| `spec.input.countOperator` | `string` |  |  |  |
| `spec.input.issueCount` | `string` |  |  |  |
| `spec.input.eidLastSeen` | `string` |  |  |  |
| `spec.input.riskLevel` | `string` |  |  |  |
| `spec.input.scoreOperator` | `string` |  |  |  |
| `spec.input.totalScore` | `double` |  |  |  |
| `spec.input.activeThreats` | `int64` |  |  |  |
| `spec.input.infected` | `bool` |  |  |  |
| `spec.input.isActive` | `bool` |  |  |  |
| `spec.input.networkStatus` | `string` |  |  |  |
| `spec.input.operationalState` | `string` |  |  |  |
| `spec.input.score` | `double` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account the rule belongs to.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The rule's name, shown in the dashboard and in policy builders.
Required here although the provider marks it optional -- an unnamed
posture rule is indistinguishable in every surface that lists rules.

- rule: {"required":true}

### spec.type

`string` · required

The check this rule runs. The 23 values are the provider's own list at
v5.23.0. WARP-client checks (file, application, os_version,
disk_encryption, firewall, domain_joined, serial_number,
unique_client_id, client_certificate, client_certificate_v2, gateway,
warp, antivirus, carbonblack, sentinelone, tanium, kolide) run on the
device; *_s2s and vendor types (crowdstrike_s2s, intune, workspace_one,
sentinelone_s2s, tanium_s2s, custom_s2s) read a service-to-service
posture integration.

- rule: type must be one of the provider's posture rule types (file, application, tanium, gateway, warp, disk_encryption, serial_number, sentinelone, carbonblack, firewall, os_version, domain_joined, client_certificate, client_certificate_v2, antivirus, unique_client_id, kolide, tanium_s2s, crowdstrike_s2s, intune, workspace_one, sentinelone_s2s, custom_s2s)
- rule: {"required":true}

### spec.description

`string`

A description of what the rule checks.

### spec.expiration

`string`

How long a posture result stays valid before it must be refreshed
(e.g. "1h", "30m"). Empty keeps results valid until overwritten by new
data from the client.

### spec.schedule

`string`

How often the WARP client re-runs the check (e.g. "5m", "1h").
Cloudflare's default is 5m; the minimum is 1m.

### spec.match

`[]CloudflareZeroTrustDevicePostureRuleMatch`

Which platforms the rule runs on. Empty runs wherever the check type
applies.

### spec.match[].platform

`string` · required

The platform the rule runs on.

- rule: platform must be one of windows, mac, linux, android, ios, chromeos
- rule: {"required":true}

### spec.input

`CloudflareZeroTrustDevicePostureRuleInput`

The check's parameters. Which fields apply depends on `type` -- see
each field's comment.

### spec.input.operatingSystem

`string`

Operating system the check targets (os_version and file-family
checks).

- rule: operating_system must be one of windows, linux, mac, android, ios, chromeos

### spec.input.path

`string`

File path to check (file and application checks).

### spec.input.exists

`bool` · optional (explicit presence)

Whether the file must exist (file check).

### spec.input.sha256

`string`

Expected SHA-256 of the file (file check).

### spec.input.thumbprint

`string`

Signing certificate thumbprint (application check).

### spec.input.id

`string`

Zero Trust list ID holding allowed values (serial_number and
unique_client_id checks read a list of device identifiers).

### spec.input.domain

`string`

The Windows domain the device must be joined to (domain_joined check).

### spec.input.operator

`string`

Version comparison operator (os_version check).

- rule: operator must be one of <, <=, >, >=, ==

### spec.input.version

`string`

OS version compared against (os_version check).

### spec.input.osDistroName

`string`

Linux distribution name (os_version check on linux).

### spec.input.osDistroRevision

`string`

Linux distribution version (os_version check on linux).

### spec.input.osVersionExtra

`string`

Extra OS version detail: Windows UBR, macOS/iOS Product Version Extra,
or the Linux distro name+version (os_version check).

### spec.input.enabled

`bool` · optional (explicit presence)

Whether the checked feature must be enabled (firewall and
disk_encryption checks).

### spec.input.checkDisks

`[]string`

Volume names to check for encryption (disk_encryption check). Empty
with require_all checks every disk.

### spec.input.requireAll

`bool` · optional (explicit presence)

Whether every disk must be encrypted (disk_encryption check).

### spec.input.certificateId

`string`

UUID of the Cloudflare-managed certificate (client_certificate and
client_certificate_v2 checks). The WARP client certificate is minted by
Cloudflare's zero-trust certificate surface, which has no catalog kind
-- a literal UUID is the only form.

### spec.input.cn

`string`

Common Name the certificate must protect (client_certificate checks).

### spec.input.checkPrivateKey

`bool` · optional (explicit presence)

Confirm the certificate was not imported from another device
(client_certificate_v2 check). Keep enabled unless the certificate was
deployed without a private key.

### spec.input.extendedKeyUsage

`[]string`

Allowed certificate public key purposes (client_certificate_v2 check).

- rule: extended_key_usage entries must be clientAuth or emailProtection

### spec.input.locations

`CloudflareZeroTrustDevicePostureRuleCertificateLocations`

Where to look for the client certificate (client_certificate_v2
check).

### spec.input.locations.paths

`[]string`

Filesystem paths to check on linux.

### spec.input.locations.trustStores

`[]string`

OS trust stores to check.

- rule: trust_stores entries must be system or user

### spec.input.subjectAlternativeNames

`[]string`

Certificate Subject Alternative Names that must be present
(client_certificate_v2 check).

### spec.input.updateWindowDays

`int64` · optional (explicit presence)

Days within which the antivirus must have updated (antivirus check).

- rule: {"int64":{"gte":"0"}}

### spec.input.complianceStatus

`string`

Required compliance state reported by the MDM (intune and
workspace_one checks).

- rule: compliance_status must be one of compliant, noncompliant, unknown, notapplicable, ingraceperiod, error

### spec.input.connectionId

`string`

The posture integration this service-to-service check reads (*_s2s,
intune, workspace_one types). Posture integrations are managed outside
this catalog today, so the value is the integration's UUID.

### spec.input.lastSeen

`string`

How recently the vendor must have seen the device, per the vendor's
own format (crowdstrike_s2s check).

### spec.input.os

`string`

Vendor-reported OS version constraint (crowdstrike_s2s check).

### spec.input.overall

`string`

Vendor-reported overall posture score constraint (crowdstrike_s2s
check).

### spec.input.sensorConfig

`string`

Vendor-reported sensor configuration constraint (crowdstrike_s2s
check).

### spec.input.state

`string`

Required device state at the vendor (crowdstrike_s2s check).

- rule: state must be one of online, offline, unknown

### spec.input.versionOperator

`string`

Operator applied to the vendor version comparison (crowdstrike_s2s
check).

- rule: version_operator must be one of <, <=, >, >=, ==

### spec.input.authState

`[]string`

Kolide device authentication states that pass the check -- the device
must be in one of them (kolide check). Values are Kolide's own,
capitalized exactly as its API reports them.

- rule: auth_state entries must be one of Good, Notified, Will Block, Blocked

### spec.input.countOperator

`string`

Operator applied to the Kolide issue count (kolide check).

- rule: count_operator must be one of <, <=, >, >=, ==

### spec.input.issueCount

`string`

The Kolide issue count compared against (kolide check). Cloudflare's
API takes this as a string.

### spec.input.eidLastSeen

`string`

How recently Tanium must have seen the device, per Tanium's format
(tanium and tanium_s2s checks).

### spec.input.riskLevel

`string`

Maximum Tanium risk level that passes (tanium and tanium_s2s checks).

- rule: risk_level must be one of low, medium, high, critical

### spec.input.scoreOperator

`string`

Operator applied to the Tanium total score (tanium and tanium_s2s
checks).

- rule: score_operator must be one of <, <=, >, >=, ==

### spec.input.totalScore

`double` · optional (explicit presence)

The Tanium total score compared against (tanium and tanium_s2s
checks).

### spec.input.activeThreats

`int64` · optional (explicit presence)

Maximum number of active threats that passes (sentinelone_s2s check).

- rule: {"int64":{"gte":"0"}}

### spec.input.infected

`bool` · optional (explicit presence)

Whether the device may be flagged infected (sentinelone_s2s check).

### spec.input.isActive

`bool` · optional (explicit presence)

Whether the vendor must report the device active (sentinelone_s2s
check).

### spec.input.networkStatus

`string`

Required network status at the vendor (sentinelone_s2s check).

- rule: network_status must be one of connected, disconnected, disconnecting, connecting

### spec.input.operationalState

`string`

Required agent operational state (sentinelone_s2s check).

- rule: operational_state must be one of na, partially_disabled, auto_fully_disabled, fully_disabled, auto_partially_disabled, disabled_error, db_corruption

### spec.input.score

`double` · optional (explicit presence)

Posture score 0-100 assigned by the third-party provider (custom_s2s
check).

- rule: {"double":{"lte":100,"gte":0}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustDevicePostureRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.rule_id` | `string` | The Cloudflare-assigned UUID of the rule (what Access and Gateway policies reference). |

## See Also

- [Overview](../README.md)
