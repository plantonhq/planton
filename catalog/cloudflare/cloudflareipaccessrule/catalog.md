# Cloudflare IP Access Rule

Deploys one IP Access rule: an allow, block, or challenge decision applied to traffic matching an IP address, IP range, ASN, or country before it reaches the zone's other security products. A rule lives either on the whole account (every zone) or on a single zone — exactly one scope must be set, and a zone rule can override an account rule for that zone. Only `mode` and `notes` update in place: Cloudflare's API accepts an edit to what a rule matches and then silently ignores it, so changing the selector means recreating the rule.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **IP Access Rule** — one `cloudflare_access_rule` on the account or the zone, carrying the action (`mode`), a single `{target, value}` selector, and the optional note. The module never sends both scopes — the provider would silently prefer the account if both arrived, so the spec requires exactly one.

Destroy is a real delete: the rule stops matching immediately, with no abandon-in-place behavior.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Account → Firewall Access Rules → Edit** for account-wide rules, or the zone equivalent for zone-scoped rules. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **An account ID** (32-character hex) for an account-wide rule, **or a zone** for `zoneId` on a zone-scoped rule — exactly one, never both.
- **A selector you are willing to recreate** — changing `configuration.target` or `configuration.value` later requires delete-and-recreate, not an in-place edit.

## Deploy

### Console

Open the deployment store, find **Cloudflare IP Access Rule**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the scope choice (account-wide or single zone), the action, and the selector target and value. Start from the **Block an IP** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareIpAccessRule
metadata:
  name: block-scanner
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  mode: block
  configuration:
    target: ip
    value: "192.0.2.10"
  notes: Block a known scanner
```

```shell
planton apply -f ip-access-rule.yaml
```

This blocks a single IPv4 address on every zone in the account. Do not set `zoneId` on the same manifest — the spec requires exactly one scope. An Infra Job tracks the provisioning in real time.

### InfraChart

For a zone-scoped rule with the zone deployed in the same InfraPipeline, wire the reference with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  mode: managed_challenge
  configuration:
    target: country
    value: US
  notes: Challenge visitors from the United States
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then creates the rule on the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring an IP Access rule. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Scope is exactly one** — `accountId` applies the rule to every zone; `zoneId` applies it to one zone and can override an account rule there. The spec rejects manifests that set both or neither, because the provider would silently prefer the account and the manifest should state its intent.

**Selector changes do not stick** — only `mode` and `notes` update in place. An edit to `configuration.target` or `configuration.value` plans an in-place update that applies "successfully" while Cloudflare keeps serving the old match — the provider's own tests document this. To change what a rule matches: create a new rule with the new selector, then destroy the old one.

**Pick the right challenge** — `block` refuses outright; `managed_challenge` lets Cloudflare pick the least intrusive check that confirms a human and is the recommended challenge mode; `challenge` forces the interactive page and `js_challenge` the non-interactive one. `whitelist` is Cloudflare's legacy name for allow — and it bypasses other security features, so use it deliberately.

**IPv6 is long form** — `ip6` values must be fully expanded (eight colon-separated groups, e.g. `2001:0db8:0000:0000:0000:0000:0000:0001`); Cloudflare rejects compressed `::` notation here even though it is valid IPv6 everywhere else.

**CIDR prefixes are a short list** — `ip_range` accepts only IPv4 `/16` or `/24` and IPv6 `/32`, `/48`, or `/64`. A single host belongs in `ip` or `ip6`, never in `ip_range`. Country values are two characters (`US`; `T1` matches Tor exit nodes), and ASN values look like `AS13335` — all validated before they reach the API.

**Write the note** — `notes` shows in the dashboard's rule list and updates in place. A block rule without a "why" is the one someone deletes during an incident.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** (zone-scoped rules) | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `rule_id` | The created rule's ID | Import identity and audit tooling |
| `zone_id` | The zone the rule was created on (empty for account-wide rules) | Deriving the rule's scope without re-reading the manifest |
| `account_id` | The account the rule was created on (empty for zone-scoped rules) | Deriving the rule's scope without re-reading the manifest |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Account-wide block** — a known-bad address refused on every zone in the account. Start from the **Block an IP** preset.

**Zone-scoped country challenge** — `managed_challenge` for visitors from one country on one zone: friction for suspect geography without refusing real users. Start from the **Challenge a country** preset. Set `value: T1` to target Tor exit nodes instead.

**Selector rotation** — when an attacker moves networks, deploy the new rule first, confirm it matches, then destroy the old one — never edit the live rule's selector, because the edit will not take.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone a zone-scoped rule applies to; wire `zoneId` via ValueFromRef
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) — expression-based WAF and custom rules when a single static selector is not enough
- [**Cloudflare Bot Management**](/infra-catalog/cloudflare-bot-management) — zone-wide bot scoring; reach for this kind when the decision is a static IP/ASN/country selector
