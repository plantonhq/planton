# CloudflareZeroTrustGatewayPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

CloudflareZeroTrustGatewayPolicySpec defines one Gateway policy (Cloudflare
calls them rules): a filter expression over employee traffic (DNS queries,
HTTP requests, or network connections) plus the action to take when it
matches -- block, allow, isolate, redirect, override, resolve, and kin.

Two provider behaviors deserve loud warnings:
  - `enabled` DEFAULTS TO FALSE at Cloudflare. A policy authored without
    enabled: true deploys DISABLED and filters nothing.
  - The wirefilter expressions (traffic, identity, device_posture) are
    reformatted by the API before storing. An expression that round-trips
    differently from how it was written shows as permanent plan drift; use
    the API-formatted form when that happens.

## Example

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareZeroTrustGatewayPolicy
metadata:
  name: test-gateway-policy
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: block-example-domains
  action: block
  filter: dns
  enabled: true
  precedence: 1000
  traffic: any(dns.domains[*] == "gambling.example.com")
  rule_settings:
    block_page_enabled: true
    block_reason: Blocked by company policy
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.accountId` | `string` | yes |  |  |
| `spec.name` | `string` | yes |  |  |
| `spec.action` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.filter` | `string` |  |  |  |
| `spec.enabled` | `bool` |  |  |  |
| `spec.precedence` | `int64` |  |  |  |
| `spec.traffic` | `string` |  |  |  |
| `spec.identity` | `string` |  |  |  |
| `spec.devicePosture` | `string` |  |  |  |
| `spec.expiration` | `CloudflareZeroTrustGatewayPolicyExpiration` |  |  |  |
| `spec.expiration.expiresAt` | `string` | yes |  |  |
| `spec.expiration.duration` | `int32` |  |  |  |
| `spec.schedule` | `CloudflareZeroTrustGatewayPolicySchedule` |  |  |  |
| `spec.schedule.mon` | `string` |  |  |  |
| `spec.schedule.tue` | `string` |  |  |  |
| `spec.schedule.wed` | `string` |  |  |  |
| `spec.schedule.thu` | `string` |  |  |  |
| `spec.schedule.fri` | `string` |  |  |  |
| `spec.schedule.sat` | `string` |  |  |  |
| `spec.schedule.sun` | `string` |  |  |  |
| `spec.schedule.timeZone` | `string` |  |  |  |
| `spec.ruleSettings` | `CloudflareZeroTrustGatewayPolicyRuleSettings` |  |  |  |
| `spec.ruleSettings.addHeaders` | `map<string, CloudflareZeroTrustGatewayPolicyStringList>` |  |  |  |
| `spec.ruleSettings.addHeaders.*.values` | `[]string` | yes |  |  |
| `spec.ruleSettings.allowChildBypass` | `bool` |  |  |  |
| `spec.ruleSettings.auditSsh` | `CloudflareZeroTrustGatewayPolicyAuditSsh` |  |  |  |
| `spec.ruleSettings.auditSsh.commandLogging` | `bool` | yes |  |  |
| `spec.ruleSettings.bisoAdminControls` | `CloudflareZeroTrustGatewayPolicyBisoAdminControls` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.version` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.copy` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.download` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.paste` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.keyboard` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.printing` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.upload` | `string` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.dcp` | `bool` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.dd` | `bool` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.dk` | `bool` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.dp` | `bool` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.du` | `bool` |  |  |  |
| `spec.ruleSettings.bisoAdminControls.wmId` | `string` |  |  |  |
| `spec.ruleSettings.blockPage` | `CloudflareZeroTrustGatewayPolicyBlockPage` |  |  |  |
| `spec.ruleSettings.blockPage.targetUri` | `string` | yes |  |  |
| `spec.ruleSettings.blockPage.includeContext` | `bool` |  |  |  |
| `spec.ruleSettings.blockPageEnabled` | `bool` |  |  |  |
| `spec.ruleSettings.blockReason` | `string` |  |  |  |
| `spec.ruleSettings.bypassParentRule` | `bool` |  |  |  |
| `spec.ruleSettings.checkSession` | `CloudflareZeroTrustGatewayPolicyCheckSession` |  |  |  |
| `spec.ruleSettings.checkSession.duration` | `string` |  |  |  |
| `spec.ruleSettings.checkSession.enforce` | `bool` |  |  |  |
| `spec.ruleSettings.deleteHeaders` | `[]string` |  |  |  |
| `spec.ruleSettings.dnsResolvers` | `CloudflareZeroTrustGatewayPolicyDnsResolvers` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv4` | `[]CloudflareZeroTrustGatewayPolicyDnsResolverV4` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv4[].ip` | `string` | yes |  |  |
| `spec.ruleSettings.dnsResolvers.ipv4[].port` | `int32` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv4[].routeThroughPrivateNetwork` | `bool` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv4[].vnetId` | `string \| valueFrom` |  |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.ruleSettings.dnsResolvers.ipv6` | `[]CloudflareZeroTrustGatewayPolicyDnsResolverV6` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv6[].ip` | `string` | yes |  |  |
| `spec.ruleSettings.dnsResolvers.ipv6[].port` | `int32` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv6[].routeThroughPrivateNetwork` | `bool` |  |  |  |
| `spec.ruleSettings.dnsResolvers.ipv6[].vnetId` | `string \| valueFrom` |  |  | CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.ruleSettings.egress` | `CloudflareZeroTrustGatewayPolicyEgress` |  |  |  |
| `spec.ruleSettings.egress.ipv4` | `string` |  |  |  |
| `spec.ruleSettings.egress.ipv4Fallback` | `string` |  |  |  |
| `spec.ruleSettings.egress.ipv6` | `string` |  |  |  |
| `spec.ruleSettings.forensicCopy` | `CloudflareZeroTrustGatewayPolicyForensicCopy` |  |  |  |
| `spec.ruleSettings.forensicCopy.enabled` | `bool` | yes |  |  |
| `spec.ruleSettings.ignoreCnameCategoryMatches` | `bool` |  |  |  |
| `spec.ruleSettings.insecureDisableDnssecValidation` | `bool` |  |  |  |
| `spec.ruleSettings.ipCategories` | `bool` |  |  |  |
| `spec.ruleSettings.ipIndicatorFeeds` | `bool` |  |  |  |
| `spec.ruleSettings.l4override` | `CloudflareZeroTrustGatewayPolicyL4Override` |  |  |  |
| `spec.ruleSettings.l4override.ip` | `string` |  |  |  |
| `spec.ruleSettings.l4override.port` | `int32` |  |  |  |
| `spec.ruleSettings.notificationSettings` | `CloudflareZeroTrustGatewayPolicyNotificationSettings` |  |  |  |
| `spec.ruleSettings.notificationSettings.enabled` | `bool` |  |  |  |
| `spec.ruleSettings.notificationSettings.includeContext` | `bool` |  |  |  |
| `spec.ruleSettings.notificationSettings.msg` | `string` |  |  |  |
| `spec.ruleSettings.notificationSettings.supportUrl` | `string` |  |  |  |
| `spec.ruleSettings.overrideHost` | `string` |  |  |  |
| `spec.ruleSettings.overrideIps` | `[]string` |  |  |  |
| `spec.ruleSettings.payloadLog` | `CloudflareZeroTrustGatewayPolicyPayloadLog` |  |  |  |
| `spec.ruleSettings.payloadLog.enabled` | `bool` | yes |  |  |
| `spec.ruleSettings.quarantine` | `CloudflareZeroTrustGatewayPolicyQuarantine` |  |  |  |
| `spec.ruleSettings.quarantine.fileTypes` | `[]string` |  |  |  |
| `spec.ruleSettings.redirect` | `CloudflareZeroTrustGatewayPolicyRedirect` |  |  |  |
| `spec.ruleSettings.redirect.targetUri` | `string` | yes |  |  |
| `spec.ruleSettings.redirect.includeContext` | `bool` |  |  |  |
| `spec.ruleSettings.redirect.preservePathAndQuery` | `bool` |  |  |  |
| `spec.ruleSettings.resolveDnsInternally` | `CloudflareZeroTrustGatewayPolicyResolveDnsInternally` |  |  |  |
| `spec.ruleSettings.resolveDnsInternally.fallback` | `string` |  |  |  |
| `spec.ruleSettings.resolveDnsInternally.viewId` | `string` |  |  |  |
| `spec.ruleSettings.resolveDnsThroughCloudflare` | `bool` |  |  |  |
| `spec.ruleSettings.setHeaders` | `map<string, CloudflareZeroTrustGatewayPolicyStringList>` |  |  |  |
| `spec.ruleSettings.setHeaders.*.values` | `[]string` | yes |  |  |
| `spec.ruleSettings.untrustedCert` | `CloudflareZeroTrustGatewayPolicyUntrustedCert` |  |  |  |
| `spec.ruleSettings.untrustedCert.action` | `string` |  |  |  |

## Field Details

### spec.accountId

`string` · required

The Cloudflare account ID that owns this policy (Gateway policies are
account-scoped).

- rule: account_id must be a 32-character hex string
- rule: {"required":true}

### spec.name

`string` · required

The display name of the policy.

- rule: {"string":{"minLen":"1"}}

### spec.action

`string` · required

The action performed when the traffic, identity, and device-posture
expressions match (absent expressions count as matching). Which actions are
valid depends on the filter: dns takes allow/block/override/safesearch/
ytrestricted, http takes allow/block/isolate/redirect/quarantine/scan,
l4 takes allow/block/audit_ssh (via "on")/l4_override, egress rules take
egress, dns_resolver rules take resolve.

- rule: {"required":true,"string":{"in":["on","off","allow","block","scan","noscan","safesearch","ytrestricted","isolate","noisolate","override","l4_override","egress","resolve","quarantine","redirect"]}}

### spec.description

`string`

Optional description of what the policy does.

### spec.filter

`string`

The protocol / layer this policy evaluates: dns (DNS queries), http (HTTP
requests), l4 (network connections), egress (dedicated egress IP
selection), or dns_resolver (custom DNS resolver policies). Cloudflare
models this as a list that can only contain a single value, so the module
sends [filter]. Leave empty to let Cloudflare infer it from the action.

- rule: filter must be http, dns, l4, egress, or dns_resolver

### spec.enabled

`bool` · optional (explicit presence)

Whether the policy is active. CLOUDFLARE DEFAULTS THIS TO FALSE: a policy
authored without enabled: true deploys disabled and filters nothing. Set it
explicitly.

### spec.precedence

`int64` · optional (explicit presence)

Evaluation order among the account's policies: LOWER values run EARLIER.
Leave unset to let Cloudflare assign one -- but for any account with more
than a few policies, manage precedence explicitly so rule order is
deliberate instead of accidental.

- rule: {"int64":{"gte":"0"}}

### spec.traffic

`string`

The wirefilter expression matching the traffic (e.g.
'any(dns.domains[*] == "example.com")'). The API reformats expressions
before storing -- if the plan shows drift, adopt the API-returned form.

### spec.identity

`string`

The wirefilter expression matching the user's identity (e.g.
'identity.email == "jane@example.com"'). Same API-reformat caveat as
traffic.

### spec.devicePosture

`string`

The wirefilter expression over device-posture check results. Same
API-reformat caveat as traffic.

### spec.expiration

`CloudflareZeroTrustGatewayPolicyExpiration`

Expiration for DNS policies: a timestamp after which the policy stops
applying (takes precedence over `schedule`). Settable only for dns rules.

### spec.expiration.expiresAt

`string` · required

When the policy expires and stops applying, as RFC3339 with a UTC offset
(non-zero offsets are converted to UTC by the API and returned with a
trailing Z).

- rule: expires_at must be an RFC3339 timestamp (e.g. 2026-09-01T00:00:00Z)
- rule: {"required":true}

### spec.expiration.duration

`int32` · optional (explicit presence)

Default active duration in minutes (minimum 5). Required to use the
API's reset_expiration endpoint on this policy.

- rule: {"int32":{"gte":5}}

### spec.schedule

`CloudflareZeroTrustGatewayPolicySchedule`

Activation schedule for DNS policies: per-weekday time intervals and an
optional time zone. Settable only for dns and dns_resolver rules.

### spec.schedule.mon

`string`

Active intervals on Mondays (e.g. "08:00-17:00").

### spec.schedule.tue

`string`

Active intervals on Tuesdays.

### spec.schedule.wed

`string`

Active intervals on Wednesdays.

### spec.schedule.thu

`string`

Active intervals on Thursdays.

### spec.schedule.fri

`string`

Active intervals on Fridays.

### spec.schedule.sat

`string`

Active intervals on Saturdays.

### spec.schedule.sun

`string`

Active intervals on Sundays.

### spec.schedule.timeZone

`string`

The tz-database city name used to evaluate the intervals (e.g.
"America/New_York"). Omit to use the time zone inferred from each user's
IP address (falling back to the Cloudflare data center's zone).

### spec.ruleSettings

`CloudflareZeroTrustGatewayPolicyRuleSettings`

Action-specific settings (block page, browser-isolation controls, custom
resolvers, egress IPs, header rewrites, and kin). Each setting applies only
to specific filter/action combinations -- the field comments carry the
pairing.

- rule: set at most one of dns_resolvers, resolve_dns_internally, or resolve_dns_through_cloudflare

### spec.ruleSettings.addHeaders

`map<string, CloudflareZeroTrustGatewayPolicyStringList>`

Add custom headers to allowed requests: header name -> list of values.
At most 20 header operations (add + set + delete) per policy; names up to
256 bytes. Only for http rules with action allow. KNOWN UPSTREAM DRIFT:
the v5 provider shows computed-field drift on policies carrying
add_headers even on first apply.

### spec.ruleSettings.addHeaders.*.values

`[]string` · required

The values for this key.

- rule: {"repeated":{"minItems":"1"}}

### spec.ruleSettings.allowChildBypass

`bool` · optional (explicit presence)

Let MSP child accounts bypass this rule (settable only by parent MSP
accounts). Applies to all rule types.

### spec.ruleSettings.auditSsh

`CloudflareZeroTrustGatewayPolicyAuditSsh`

Audit SSH settings. Only for l4 rules with the audit_ssh action.

### spec.ruleSettings.auditSsh.commandLogging

`bool` · required · optional (explicit presence)

Log the SSH commands executed in the session. Required inside audit_ssh.

- rule: {"required":true}

### spec.ruleSettings.bisoAdminControls

`CloudflareZeroTrustGatewayPolicyBisoAdminControls`

Browser-isolation behavior controls. Only for http rules with action
isolate.

### spec.ruleSettings.bisoAdminControls.version

`string`

Which control generation applies. Leave empty for Cloudflare's default
(v1). The v2 string controls below require "v2"; the v1 booleans require
"v1".

- rule: version must be v1 or v2

### spec.ruleSettings.bisoAdminControls.copy

`string`

v2: copy behavior. remote_only blocks copying isolated content to the
local clipboard; absent leaves copying enabled.

- rule: copy must be enabled, disabled, or remote_only

### spec.ruleSettings.bisoAdminControls.download

`string`

v2: download behavior. remote_only lets users view but not save downloads;
absent leaves downloading enabled.

- rule: download must be enabled, disabled, or remote_only

### spec.ruleSettings.bisoAdminControls.paste

`string`

v2: paste behavior. remote_only blocks pasting local-clipboard content
into isolated pages; absent leaves pasting enabled.

- rule: paste must be enabled, disabled, or remote_only

### spec.ruleSettings.bisoAdminControls.keyboard

`string`

v2: keyboard behavior; absent leaves keyboard usage enabled.

- rule: keyboard must be enabled or disabled

### spec.ruleSettings.bisoAdminControls.printing

`string`

v2: print behavior; absent leaves printing enabled.

- rule: printing must be enabled or disabled

### spec.ruleSettings.bisoAdminControls.upload

`string`

v2: upload behavior; absent leaves uploading enabled.

- rule: upload must be enabled or disabled

### spec.ruleSettings.bisoAdminControls.dcp

`bool` · optional (explicit presence)

v1: set false to ENABLE copy-paste (the double negative is Cloudflare's).

### spec.ruleSettings.bisoAdminControls.dd

`bool` · optional (explicit presence)

v1: set false to ENABLE downloading.

### spec.ruleSettings.bisoAdminControls.dk

`bool` · optional (explicit presence)

v1: set false to ENABLE keyboard usage.

### spec.ruleSettings.bisoAdminControls.dp

`bool` · optional (explicit presence)

v1: set false to ENABLE printing.

### spec.ruleSettings.bisoAdminControls.du

`bool` · optional (explicit presence)

v1: set false to ENABLE uploading.

### spec.ruleSettings.bisoAdminControls.wmId

`string`

Watermark ID (UUID) rendered over the isolated browser session, when set.

### spec.ruleSettings.blockPage

`CloudflareZeroTrustGatewayPolicyBlockPage`

Custom block page (overriding the account-level page). Only for http
rules with action block.

### spec.ruleSettings.blockPage.targetUri

`string` · required

The URI the blocked user is redirected to.

- rule: {"required":true}

### spec.ruleSettings.blockPage.includeContext

`bool`

Pass the block context (rule, user, categories) as query parameters.

### spec.ruleSettings.blockPageEnabled

`bool` · optional (explicit presence)

Show the custom block page for blocked DNS queries. Only for dns rules
with action block.

### spec.ruleSettings.blockReason

`string`

Why the rule blocks the request; shown on the block page when enabled.
For dns, l4, and http rules with action block.

### spec.ruleSettings.bypassParentRule

`bool` · optional (explicit presence)

Let this MSP child account bypass its parent's rules (settable only by
child accounts). Applies to all rule types.

### spec.ruleSettings.checkSession

`CloudflareZeroTrustGatewayPolicyCheckSession`

Session-freshness enforcement. Only for l4 and http rules with action
allow. The API normalizes duration strings (e.g. "24h" becomes "24h0m0s").

### spec.ruleSettings.checkSession.duration

`string`

The required session freshness threshold (e.g. "24h"). The API returns a
normalized form ("24h0m0s") -- use it if the plan shows drift.

### spec.ruleSettings.checkSession.enforce

`bool`

Enforce the session check.

### spec.ruleSettings.deleteHeaders

`[]string`

Remove headers from allowed requests, by name. Counts toward the 20
header-operation budget. Only for http rules with action allow.

### spec.ruleSettings.dnsResolvers

`CloudflareZeroTrustGatewayPolicyDnsResolvers`

Custom upstream resolvers for matched queries (routed to the resolver
closest to the query's origin). Only for dns_resolver rules with action
resolve; mutually exclusive with the other resolution mechanisms.

### spec.ruleSettings.dnsResolvers.ipv4

`[]CloudflareZeroTrustGatewayPolicyDnsResolverV4`

IPv4 upstream resolvers.

- rule: route_through_private_network must be true when vnet_id is set

### spec.ruleSettings.dnsResolvers.ipv4[].ip

`string` · required

The resolver's IPv4 address.

- rule: {"required":true}

### spec.ruleSettings.dnsResolvers.ipv4[].port

`int32` · optional (explicit presence)

The resolver's port. Omit for 53.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.ruleSettings.dnsResolvers.ipv4[].routeThroughPrivateNetwork

`bool`

Connect to this resolver over a private network (a Zero Trust tunnel).
Must be true when vnet_id is set.

### spec.ruleSettings.dnsResolvers.ipv4[].vnetId

`string | valueFrom`

The virtual network the private resolver lives in: a literal virtual
network UUID, or a reference to a CloudflareZeroTrustTunnelVirtualNetwork.
Omit to use the account's default virtual network.

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.ruleSettings.dnsResolvers.ipv6

`[]CloudflareZeroTrustGatewayPolicyDnsResolverV6`

IPv6 upstream resolvers.

- rule: route_through_private_network must be true when vnet_id is set

### spec.ruleSettings.dnsResolvers.ipv6[].ip

`string` · required

The resolver's IPv6 address.

- rule: {"required":true}

### spec.ruleSettings.dnsResolvers.ipv6[].port

`int32` · optional (explicit presence)

The resolver's port. Omit for 53.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.ruleSettings.dnsResolvers.ipv6[].routeThroughPrivateNetwork

`bool`

Connect to this resolver over a private network (a Zero Trust tunnel).
Must be true when vnet_id is set.

### spec.ruleSettings.dnsResolvers.ipv6[].vnetId

`string | valueFrom`

The virtual network the private resolver lives in: a literal virtual
network UUID, or a reference to a CloudflareZeroTrustTunnelVirtualNetwork.
Omit to use the account's default virtual network.

- references: CloudflareZeroTrustTunnelVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareZeroTrustTunnelVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.ruleSettings.egress

`CloudflareZeroTrustGatewayPolicyEgress`

Dedicated egress IPs for matched proxy traffic; omit for local egress via
WARP IPs. Only for egress rules.

### spec.ruleSettings.egress.ipv4

`string`

The dedicated IPv4 address to egress through.

### spec.ruleSettings.egress.ipv4Fallback

`string`

The fallback IPv4 address when the primary fails. "0.0.0.0" means fall
back to local egress via WARP IPs.

### spec.ruleSettings.egress.ipv6

`string`

The IPv6 range to egress through.

### spec.ruleSettings.forensicCopy

`CloudflareZeroTrustGatewayPolicyForensicCopy`

Send a copy of the matched HTTP request to storage. Only for http rules.

### spec.ruleSettings.forensicCopy.enabled

`bool` · required · optional (explicit presence)

Enable sending the copy to storage. Required inside forensic_copy.

- rule: {"required":true}

### spec.ruleSettings.ignoreCnameCategoryMatches

`bool` · optional (explicit presence)

Ignore category matches on CNAME domains in the response (evaluate only
the queried domain's categories). Only for dns and dns_resolver rules.

### spec.ruleSettings.insecureDisableDnssecValidation

`bool` · optional (explicit presence)

Disable DNSSEC validation for allowed queries. INSECURE. Only for dns
rules.

### spec.ruleSettings.ipCategories

`bool` · optional (explicit presence)

Block IPs in DNS category blocks, not just domain names. Only for dns and
dns_resolver rules.

### spec.ruleSettings.ipIndicatorFeeds

`bool` · optional (explicit presence)

Include IPs in DNS indicator-feed blocks, not just domain names. Only for
dns and dns_resolver rules.

### spec.ruleSettings.l4override

`CloudflareZeroTrustGatewayPolicyL4Override`

Destination override for matched traffic. Only for l4 rules with action
l4_override.

### spec.ruleSettings.l4override.ip

`string`

The destination IPv4 or IPv6 address.

### spec.ruleSettings.l4override.port

`int32` · optional (explicit presence)

The destination port for TCP/UDP overrides.

- rule: {"int32":{"lte":65535,"gte":1}}

### spec.ruleSettings.notificationSettings

`CloudflareZeroTrustGatewayPolicyNotificationSettings`

Device notification shown when the rule matches. For any rule type with
action block.

### spec.ruleSettings.notificationSettings.enabled

`bool`

Enable the notification.

### spec.ruleSettings.notificationSettings.includeContext

`bool`

Pass the block context as query parameters on the support URL.

### spec.ruleSettings.notificationSettings.msg

`string`

The message shown in the notification.

### spec.ruleSettings.notificationSettings.supportUrl

`string`

A URL directing users to more information. Omit to open the block page.

### spec.ruleSettings.overrideHost

`string`

Hostname answered for matched DNS queries. Only for dns rules with action
override.

### spec.ruleSettings.overrideIps

`[]string`

IP(s) answered for matched DNS queries. Only for dns rules with action
override. KNOWN UPSTREAM DRIFT: the v5 provider shows computed-field
drift on policies carrying override_ips even on first apply.

### spec.ruleSettings.payloadLog

`CloudflareZeroTrustGatewayPolicyPayloadLog`

DLP payload logging. Only for http rules.

### spec.ruleSettings.payloadLog.enabled

`bool` · required · optional (explicit presence)

Enable DLP payload logging for this rule. Required inside payload_log.

- rule: {"required":true}

### spec.ruleSettings.quarantine

`CloudflareZeroTrustGatewayPolicyQuarantine`

Quarantine (sandbox) settings. Only for http rules with action quarantine.

### spec.ruleSettings.quarantine.fileTypes

`[]string`

The file types to sandbox.

- rule: {"repeated":{"items":{"string":{"in":["exe","pdf","doc","docm","docx","rtf","ppt","pptx","xls","xlsm","xlsx","zip","rar"]}}}}

### spec.ruleSettings.redirect

`CloudflareZeroTrustGatewayPolicyRedirect`

Redirect settings. Only for http rules with action redirect.

### spec.ruleSettings.redirect.targetUri

`string` · required

The URI the user is redirected to.

- rule: {"required":true}

### spec.ruleSettings.redirect.includeContext

`bool`

Pass the block context as query parameters.

### spec.ruleSettings.redirect.preservePathAndQuery

`bool`

Append the original request's path and query to target_uri.

### spec.ruleSettings.resolveDnsInternally

`CloudflareZeroTrustGatewayPolicyResolveDnsInternally`

Forward matched queries to the internal DNS service. Only for
dns_resolver rules with action resolve; mutually exclusive with the other
resolution mechanisms.

### spec.ruleSettings.resolveDnsInternally.fallback

`string`

Fallback when the internal response is not NOERROR, or when an A/AAAA
answer carries only CNAMEs: none (return as-is) or public_dns.

- rule: fallback must be none or public_dns

### spec.ruleSettings.resolveDnsInternally.viewId

`string`

The internal DNS view ID passed to the internal DNS service.

### spec.ruleSettings.resolveDnsThroughCloudflare

`bool` · optional (explicit presence)

Send matched queries to Cloudflare's public 1.1.1.1 resolver. Only for
dns_resolver rules with action resolve; mutually exclusive with the other
resolution mechanisms.

### spec.ruleSettings.setHeaders

`map<string, CloudflareZeroTrustGatewayPolicyStringList>`

Replace (or add) headers on allowed requests: header name -> list of
values. Values may interpolate `@{selector.name}` variables at the edge
(escape a literal `@{` as `@@{`). Counts toward the 20 header-operation
budget; values up to 4 KB. Only for http rules with action allow.

### spec.ruleSettings.setHeaders.*.values

`[]string` · required

The values for this key.

- rule: {"repeated":{"minItems":"1"}}

### spec.ruleSettings.untrustedCert

`CloudflareZeroTrustGatewayPolicyUntrustedCert`

Behavior when the upstream certificate is invalid or an SSL error occurs.
Only for http rules with action allow.

### spec.ruleSettings.untrustedCert.action

`string`

The action taken: pass_through, block, or error (the default, an HTTP 526).

- rule: action must be pass_through, block, or error

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareZeroTrustGatewayPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.policy_id` | `string` | The UUID of the created policy. |
| `status.outputs.precedence` | `string` | The policy's evaluation precedence (Cloudflare assigns one when the spec leaves it unset; lower runs earlier). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.ruleSettings.dnsResolvers.ipv4[].vnetId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |
| `spec.ruleSettings.dnsResolvers.ipv6[].vnetId` | CloudflareZeroTrustTunnelVirtualNetwork | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
