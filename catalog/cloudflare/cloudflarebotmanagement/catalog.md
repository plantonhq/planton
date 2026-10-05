# Cloudflare Bot Management

Manages a zone's Bot Management configuration: the singleton switchboard deciding how Cloudflare treats automated traffic, from the free Bot Fight Mode toggle through Super Bot Fight Mode's per-category actions up to the Enterprise Bot Management knobs and the AI-crawler controls. A field left unset is not managed — the zone keeps whatever value it already carries. The surface has no delete at Cloudflare: destroying this resource abandons the last-applied values rather than reverting them, so every field you set becomes a value you own until you set it to something else.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bot Management Configuration** — one `cloudflare_bot_management` on the zone. The surface is a zone singleton: Cloudflare's identity for it is the zone ID, create adopts whatever configuration the zone already carries, and only the fields you set produce an API write.

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module whose API token carries **Zone → Bot Management → Edit**. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone** for `zoneId` — typically a CloudflareDnsZone whose `zone_id` output this resource references.
- **The right zone plan for gated fields** — `fightMode` is free; the `sbfm*` fields need Pro/Business (`sbfmLikelyAutomated` needs Business+); `autoUpdateModel`, `suppressSessionScore`, and `bmCookieEnabled` need Enterprise Bot Management. `enableJs` is writable on every plan — and required on-or-with `fightMode` when the zone's JavaScript detections are off (Cloudflare rejects Fight Mode alone with a 400 otherwise). Setting a field the plan does not include fails at the API, and non-entitled zones omit those fields from responses, which surfaces as perpetual refresh drift.

## Deploy

### Console

Open the deployment store, find **Cloudflare Bot Management**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the target zone, and the bot-management fields grouped by plan family — Fight Mode, Super Bot Fight Mode actions, Enterprise knobs, and the AI-crawler controls. Start from the **Bot Fight Mode** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareBotManagement
metadata:
  name: fight-mode
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  fightMode: true
  enableJs: true
```

```shell
planton apply -f bot-management.yaml
```

This turns on Bot Fight Mode and touches nothing else — every other Bot Management field stays exactly as the zone had it. The `enableJs: true` pair is required when enabling Fight Mode on a zone whose JavaScript detections are off — Cloudflare rejects Fight Mode alone with "cannot enable Fight_Mode while EnableJS is disabled". An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire the reference with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-zone
      fieldPath: status.outputs.zone_id
  sbfmDefinitelyAutomated: managed_challenge
  sbfmVerifiedBots: allow
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then applies the Bot Management configuration to the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring Bot Management. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Destroy restores nothing** — Cloudflare's Bot Management API has no delete. If you set `fightMode: true` and later destroy the resource, Fight Mode stays on indefinitely with nothing in your infrastructure declaring it. The discipline is revert-before-retire: set every managed field to the value the zone should keep, apply, and only then destroy.

**Unset means unmanaged, by design** — every field is presence-tracked and the module never sends defaults; managing one field never overwrites the other fourteen. The floor is enforced: a spec with only `zoneId` is rejected, because a resource that manages nothing would deploy nothing.

**Fight Mode or SBFM, not both** — `fightMode` (free) is mutually exclusive with Super Bot Fight Mode at Cloudflare. Zones on SBFM plans manage the `sbfm*` fields instead.

**Verified bots are search engines** — `sbfmVerifiedBots: block` blocks search indexing along with every other bot Cloudflare has verified. `allow` is the safe default posture; handle exceptions with skip rules in a Cloudflare Ruleset.

**Static-resource protection catches real users** — `sbfmStaticResourceProtection` extends SBFM to images, CSS, and JS, and can block legitimate hotlinked-asset traffic. Cloudflare's own documentation carries the same caution.

**Plan walls are Cloudflare's, and they move** — the spec validates values, not entitlements. A field the zone's plan does not include fails at apply, and on non-entitled zones the API omits the field from responses, so the manifest refreshes as a perpetual diff. Manage only what the plan includes.

**AI-crawler posture** — `aiBotsProtection: block` blocks AI scrapers zone-wide (`only_on_ad_pages` limits it to ad-carrying pages), `crawlerProtection: enabled` punishes them with a link maze instead of a plain block, and `contentBotsProtection: block` blocks low-scoring automated traffic while excluding Cloudflare's safe verified-bot categories.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

This component's `status.outputs` only echoes the managed zone's ID back (`zone_id`) — the Bot Management configuration is a zone singleton whose identity is the zone itself, so there is nothing new for downstream resources to consume.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Fight Mode on a free zone** — the one Bot Management field every plan carries, with every other knob left unmanaged. Start from the **Bot Fight Mode** preset.

**SBFM production posture** — `sbfmDefinitelyAutomated: managed_challenge` with `sbfmVerifiedBots: allow`: challenge what is certainly automated, never touch search crawlers, and leave likely-automated unmanaged until you have watched the analytics.

**AI-crawler lockdown** — `aiBotsProtection: block` with `isRobotsTxtManaged: true`, so the policy is both enforced at the edge and published in a Cloudflare-managed robots.txt prepended to the origin's.

## Works With

- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone this manages; its plan gates the SBFM and Enterprise fields
- [**Cloudflare Ruleset**](/infra-catalog/cloudflare-ruleset) — skip rules for verified-bot exceptions under `contentBotsProtection`
- [**Cloudflare IP Access Rule**](/infra-catalog/cloudflare-ip-access-rule) — a static IP/ASN/country decision when scored bot traffic is the wrong tool
