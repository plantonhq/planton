# CloudflareZeroTrustGatewaySettings

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustGatewaySettingsSpec configures the account's Secure Web
Gateway: the settings panel behind every Gateway policy (TLS inspection,
antivirus scanning, block-page branding, browser isolation, sandboxing),
the activity-logging controls, and the account's proxy auto-config (PAC)
files.

The spec folds three Cloudflare surfaces with different lifecycles into
one component:
  - settings: the account configuration SINGLETON. Create and update are
    the same PUT; destroy is a NO-OP that abandons the live configuration
    exactly as last applied. An UNSET sub-object is NOT MANAGED -- it is
    never sent, and the live value (dashboard-set or default) stays
    untouched.
  - logging: the logging SINGLETON, same upsert/no-op-destroy behavior.
    Unlike settings, the modules always send the COMPLETE logging tree
    when this field is set (unset leaves logging unmanaged): Cloudflare
    reports drift on omitted logging fields, so partial sends would never
    converge.
  - pac_files: a real per-file collection -- each row is created,
    updated, and DELETED like a normal resource.

## Example

```yaml
# Complete example manifest for CloudflareZeroTrustGatewaySettings.
# Manages the account's Secure Web Gateway configuration (unset sub-objects
# stay unmanaged), the logging controls (the full tree is sent when
# declared), and PAC files (each row a real resource).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustGatewaySettings
metadata:
  name: acme-gateway
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  settings:
    activity_log:
      enabled: true
    protocol_detection:
      enabled: true
    block_page:
      enabled: true
      name: Blocked by Acme IT
      header_text: This site is blocked
      footer_text: Questions? Contact IT.
      background_color: "#0f172a"
    max_ttl_secs: 300
  logging:
    redact_pii: true
    settings_by_rule_type:
      dns:
        log_all: false
        log_blocks: true
      http:
        log_all: false
        log_blocks: true
      l4:
        log_all: false
        log_blocks: false
  pac_files:
    - name: default-proxy
      contents: 'function FindProxyForURL(url, host) { return "DIRECT"; }'
      description: Default direct PAC file
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.settings` | `CloudflareZeroTrustGatewayConfig` |  |  |  |
| `spec.settings.activityLog` | `CloudflareZeroTrustGatewayToggle` |  |  |  |
| `spec.settings.activityLog.enabled` | `bool` |  |  |  |
| `spec.settings.antivirus` | `CloudflareZeroTrustGatewayAntivirus` |  |  |  |
| `spec.settings.antivirus.enabledDownloadPhase` | `bool` |  |  |  |
| `spec.settings.antivirus.enabledUploadPhase` | `bool` |  |  |  |
| `spec.settings.antivirus.failClosed` | `bool` |  |  |  |
| `spec.settings.antivirus.notificationSettings` | `CloudflareZeroTrustGatewayNotificationSettings` |  |  |  |
| `spec.settings.antivirus.notificationSettings.enabled` | `bool` |  |  |  |
| `spec.settings.antivirus.notificationSettings.msg` | `string` |  |  |  |
| `spec.settings.antivirus.notificationSettings.supportUrl` | `string` |  |  |  |
| `spec.settings.antivirus.notificationSettings.includeContext` | `bool` |  |  |  |
| `spec.settings.blockPage` | `CloudflareZeroTrustGatewayBlockPage` |  |  |  |
| `spec.settings.blockPage.enabled` | `bool` |  |  |  |
| `spec.settings.blockPage.mode` | `string` |  |  |  |
| `spec.settings.blockPage.targetUri` | `string` |  |  |  |
| `spec.settings.blockPage.includeContext` | `bool` |  |  |  |
| `spec.settings.blockPage.name` | `string` |  |  |  |
| `spec.settings.blockPage.headerText` | `string` |  |  |  |
| `spec.settings.blockPage.footerText` | `string` |  |  |  |
| `spec.settings.blockPage.suppressFooter` | `bool` |  |  |  |
| `spec.settings.blockPage.backgroundColor` | `string` |  |  |  |
| `spec.settings.blockPage.logoPath` | `string` |  |  |  |
| `spec.settings.blockPage.mailtoAddress` | `string` |  |  |  |
| `spec.settings.blockPage.mailtoSubject` | `string` |  |  |  |
| `spec.settings.bodyScanning` | `CloudflareZeroTrustGatewayBodyScanning` |  |  |  |
| `spec.settings.bodyScanning.inspectionMode` | `string` |  |  |  |
| `spec.settings.browserIsolation` | `CloudflareZeroTrustGatewayBrowserIsolation` |  |  |  |
| `spec.settings.browserIsolation.nonIdentityEnabled` | `bool` |  |  |  |
| `spec.settings.browserIsolation.urlBrowserIsolationEnabled` | `bool` |  |  |  |
| `spec.settings.certificate` | `CloudflareZeroTrustGatewayCertificate` |  |  |  |
| `spec.settings.certificate.id` | `string \| valueFrom` | yes |  |  |
| `spec.settings.extendedEmailMatching` | `CloudflareZeroTrustGatewayToggle` |  |  |  |
| `spec.settings.extendedEmailMatching.enabled` | `bool` |  |  |  |
| `spec.settings.fips` | `CloudflareZeroTrustGatewayFips` |  |  |  |
| `spec.settings.fips.tls` | `bool` |  |  |  |
| `spec.settings.hostSelector` | `CloudflareZeroTrustGatewayToggle` |  |  |  |
| `spec.settings.hostSelector.enabled` | `bool` |  |  |  |
| `spec.settings.inspection` | `CloudflareZeroTrustGatewayInspection` |  |  |  |
| `spec.settings.inspection.mode` | `string` |  |  |  |
| `spec.settings.maxTtlSecs` | `int64` |  |  |  |
| `spec.settings.protocolDetection` | `CloudflareZeroTrustGatewayToggle` |  |  |  |
| `spec.settings.protocolDetection.enabled` | `bool` |  |  |  |
| `spec.settings.sandbox` | `CloudflareZeroTrustGatewaySandbox` |  |  |  |
| `spec.settings.sandbox.enabled` | `bool` |  |  |  |
| `spec.settings.sandbox.fallbackAction` | `string` |  |  |  |
| `spec.settings.tlsDecrypt` | `CloudflareZeroTrustGatewayToggle` |  |  |  |
| `spec.settings.tlsDecrypt.enabled` | `bool` |  |  |  |
| `spec.logging` | `CloudflareZeroTrustGatewayLogging` |  |  |  |
| `spec.logging.redactPii` | `bool` |  |  |  |
| `spec.logging.settingsByRuleType` | `CloudflareZeroTrustGatewayLoggingByRuleType` |  |  |  |
| `spec.logging.settingsByRuleType.dns` | `CloudflareZeroTrustGatewayLoggingRule` |  |  |  |
| `spec.logging.settingsByRuleType.dns.logAll` | `bool` |  |  |  |
| `spec.logging.settingsByRuleType.dns.logBlocks` | `bool` |  |  |  |
| `spec.logging.settingsByRuleType.http` | `CloudflareZeroTrustGatewayLoggingRule` |  |  |  |
| `spec.logging.settingsByRuleType.http.logAll` | `bool` |  |  |  |
| `spec.logging.settingsByRuleType.http.logBlocks` | `bool` |  |  |  |
| `spec.logging.settingsByRuleType.l4` | `CloudflareZeroTrustGatewayLoggingRule` |  |  |  |
| `spec.logging.settingsByRuleType.l4.logAll` | `bool` |  |  |  |
| `spec.logging.settingsByRuleType.l4.logBlocks` | `bool` |  |  |  |
| `spec.pacFiles` | `[]CloudflareZeroTrustGatewayPacFile` |  |  |  |
| `spec.pacFiles[].name` | `string` | yes |  |  |
| `spec.pacFiles[].contents` | `string` | yes |  |  |
| `spec.pacFiles[].slug` | `string` |  |  |  |
| `spec.pacFiles[].description` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account whose Gateway is configured.

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.settings

`CloudflareZeroTrustGatewayConfig`

The Gateway configuration surfaces to manage. Each unset sub-object is
left untouched at Cloudflare.

### spec.settings.activityLog

`CloudflareZeroTrustGatewayToggle`

Activity logging of Gateway traffic (the master switch; per-rule-type
detail lives in the spec's logging field).

### spec.settings.activityLog.enabled

`bool` · optional (explicit presence)

Whether the surface is enabled.

### spec.settings.antivirus

`CloudflareZeroTrustGatewayAntivirus`

Anti-virus scanning of proxied traffic.

### spec.settings.antivirus.enabledDownloadPhase

`bool` · optional (explicit presence)

Scan files being downloaded through Gateway.

### spec.settings.antivirus.enabledUploadPhase

`bool` · optional (explicit presence)

Scan files being uploaded through Gateway.

### spec.settings.antivirus.failClosed

`bool` · optional (explicit presence)

When a file cannot be scanned (encrypted archives, oversize), block it
instead of letting it through.

### spec.settings.antivirus.notificationSettings

`CloudflareZeroTrustGatewayNotificationSettings`

What users see when a scan blocks their transfer.

### spec.settings.antivirus.notificationSettings.enabled

`bool` · optional (explicit presence)

Whether to show a notification at all.

### spec.settings.antivirus.notificationSettings.msg

`string`

The notification message text.

### spec.settings.antivirus.notificationSettings.supportUrl

`string`

A support URL included in the notification.

### spec.settings.antivirus.notificationSettings.includeContext

`bool` · optional (explicit presence)

Include context (rule and category detail) in the notification.

### spec.settings.blockPage

`CloudflareZeroTrustGatewayBlockPage`

The branded page Gateway shows on blocked requests.

- rule: mode redirect_uri needs target_uri set to the page to redirect to

### spec.settings.blockPage.enabled

`bool` · optional (explicit presence)

Whether the custom block page is enabled (false serves Cloudflare's
plain default).

### spec.settings.blockPage.mode

`string`

How blocks are presented: customized_block_page (Cloudflare hosts the
branded page below) or redirect_uri (redirect to target_uri). Unset
keeps Cloudflare's default presentation.

- rule: mode must be customized_block_page or redirect_uri (or unset for the default presentation)

### spec.settings.blockPage.targetUri

`string`

The URL to redirect blocked requests to (redirect_uri mode).

### spec.settings.blockPage.includeContext

`bool` · optional (explicit presence)

Include the block context (rule, category) as query parameters on the
redirect / page.

### spec.settings.blockPage.name

`string`

The page name shown in the dashboard.

### spec.settings.blockPage.headerText

`string`

The page heading text.

### spec.settings.blockPage.footerText

`string`

The page footer text.

### spec.settings.blockPage.suppressFooter

`bool` · optional (explicit presence)

Hide the default Gateway footer entirely.

### spec.settings.blockPage.backgroundColor

`string`

The page background color, as a CSS color (e.g. "#1e293b").

### spec.settings.blockPage.logoPath

`string`

The path to the logo shown on the page.

### spec.settings.blockPage.mailtoAddress

`string`

A mailto: address users can report false positives to.

### spec.settings.blockPage.mailtoSubject

`string`

The subject line pre-filled on the report email.

### spec.settings.bodyScanning

`CloudflareZeroTrustGatewayBodyScanning`

Deep inspection of uploaded/downloaded file bodies.

### spec.settings.bodyScanning.inspectionMode

`string`

deep waits for the whole file before verdict (thorough, slower);
shallow scans the stream as it passes (faster, first bytes only).

- rule: inspection_mode must be deep or shallow

### spec.settings.browserIsolation

`CloudflareZeroTrustGatewayBrowserIsolation`

Remote browser isolation behavior.

### spec.settings.browserIsolation.nonIdentityEnabled

`bool` · optional (explicit presence)

Isolate browsing for traffic without a user identity (e.g. anonymous
DNS-over-HTTPS locations).

### spec.settings.browserIsolation.urlBrowserIsolationEnabled

`bool` · optional (explicit presence)

Enable URL-triggered isolation (policies with the isolate action).

### spec.settings.certificate

`CloudflareZeroTrustGatewayCertificate`

The certificate Gateway presents when inspecting TLS.

### spec.settings.certificate.id

`string | valueFrom` · required

The Gateway certificate to present: a certificate UUID from the
account's Zero Trust certificates (the nil UUID selects the Cloudflare
Root CA). The certificate must be ACTIVATED on the account before TLS
decryption will turn on. Accepts a literal UUID or a reference; no
catalog kind manages Gateway certificates yet, so a literal UUID is the
common form today.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.settings.extendedEmailMatching

`CloudflareZeroTrustGatewayToggle`

Matching of email aliases (e.g. user+tag@domain) to the canonical
identity in policies.

### spec.settings.extendedEmailMatching.enabled

`bool` · optional (explicit presence)

Whether the surface is enabled.

### spec.settings.fips

`CloudflareZeroTrustGatewayFips`

FIPS-compliant TLS enforcement.

### spec.settings.fips.tls

`bool` · optional (explicit presence)

Restrict TLS to FIPS-approved ciphers and versions.

### spec.settings.hostSelector

`CloudflareZeroTrustGatewayToggle`

Host selectors in HTTP policies (matching on host across both DNS and
HTTP filtering).

### spec.settings.hostSelector.enabled

`bool` · optional (explicit presence)

Whether the surface is enabled.

### spec.settings.inspection

`CloudflareZeroTrustGatewayInspection`

How Gateway decides whether to inspect traffic: static (by
configuration) or dynamic (by learned signals).

### spec.settings.inspection.mode

`string`

static inspects what configuration says to inspect; dynamic learns from
traffic signals.

- rule: mode must be static or dynamic

### spec.settings.maxTtlSecs

`int64` · optional (explicit presence)

The maximum TTL, in seconds (60-36000), Gateway's DNS resolver returns
for cached responses. Unset leaves responses uncapped (or managed
elsewhere).

- rule: {"int64":{"lte":"36000","gte":"60"}}

### spec.settings.protocolDetection

`CloudflareZeroTrustGatewayToggle`

Detection of protocols tunneled over non-standard ports.

### spec.settings.protocolDetection.enabled

`bool` · optional (explicit presence)

Whether the surface is enabled.

### spec.settings.sandbox

`CloudflareZeroTrustGatewaySandbox`

Sandboxed execution of downloaded files before delivery.

### spec.settings.sandbox.enabled

`bool` · optional (explicit presence)

Whether sandboxing is enabled.

### spec.settings.sandbox.fallbackAction

`string`

What happens to a file while its sandbox verdict is pending: allow or
block.

- rule: fallback_action must be allow or block

### spec.settings.tlsDecrypt

`CloudflareZeroTrustGatewayToggle`

TLS decryption -- the switch that lets Gateway inspect HTTPS traffic.
Cloudflare REJECTS enabling this (error 2211) until a Gateway
certificate exists and is activated on the account.

### spec.settings.tlsDecrypt.enabled

`bool` · optional (explicit presence)

Whether the surface is enabled.

### spec.logging

`CloudflareZeroTrustGatewayLogging`

The Gateway activity-logging configuration. Unset leaves logging
unmanaged.

### spec.logging.redactPii

`bool` · optional (explicit presence)

Redact personally identifiable information from activity logs.

### spec.logging.settingsByRuleType

`CloudflareZeroTrustGatewayLoggingByRuleType`

Per-rule-type logging detail.

### spec.logging.settingsByRuleType.dns

`CloudflareZeroTrustGatewayLoggingRule`

Logging for DNS firewall rules.

### spec.logging.settingsByRuleType.dns.logAll

`bool` · optional (explicit presence)

Log every request evaluated by this rule type.

### spec.logging.settingsByRuleType.dns.logBlocks

`bool` · optional (explicit presence)

Log only requests that a rule blocked.

### spec.logging.settingsByRuleType.http

`CloudflareZeroTrustGatewayLoggingRule`

Logging for HTTP/HTTPS firewall rules.

### spec.logging.settingsByRuleType.http.logAll

`bool` · optional (explicit presence)

Log every request evaluated by this rule type.

### spec.logging.settingsByRuleType.http.logBlocks

`bool` · optional (explicit presence)

Log only requests that a rule blocked.

### spec.logging.settingsByRuleType.l4

`CloudflareZeroTrustGatewayLoggingRule`

Logging for network (L4) firewall rules.

### spec.logging.settingsByRuleType.l4.logAll

`bool` · optional (explicit presence)

Log every request evaluated by this rule type.

### spec.logging.settingsByRuleType.l4.logBlocks

`bool` · optional (explicit presence)

Log only requests that a rule blocked.

### spec.pacFiles

`[]CloudflareZeroTrustGatewayPacFile`

Proxy auto-config (PAC) files served by Gateway. Each row is one file
with its own lifecycle (removing a row deletes the file).

### spec.pacFiles[].name

`string` · required

The file's name, shown in the dashboard.

- rule: {"required":true}

### spec.pacFiles[].contents

`string` · required

The PAC file body (JavaScript with a FindProxyForURL function).

- rule: {"required":true}

### spec.pacFiles[].slug

`string`

The URL-friendly slug the file is served under. IMMUTABLE: the slug is
baked into the file's public URL, so changing it replaces the file
(clients configured with the old URL break). When omitted, the modules
derive a deterministic slug from the file's name (lowercased,
non-alphanumerics become hyphens) -- NOT the server's choice, because a
server-generated slug is RANDOM and, being replace-forcing, would make
every subsequent plan propose recreating the file and breaking its URL
(live-measured at provider v5.23.0).

### spec.pacFiles[].description

`string`

A free-form description of the file.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustGatewaySettings, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.account_id` | `string` | The Cloudflare account the configuration was applied to (the singleton's identity -- the harness and import recipes key on it). |
| `status.outputs.pacfile_ids` | `map<string, string>` | Cloudflare-assigned ids of the PAC files, keyed by each file's name (the same key the modules' per-row resources use). PAC-file ids are server-assigned at creation, so import recipes derive per-file import IDs ("{account_id}/{pacfile_id}") from this map instead of sending adopters to list the API collection. |

## See Also

- [Overview](../README.md)
