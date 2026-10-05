# Cloudflare Bot Management

## Overview

`CloudflareBotManagement` manages a zone's Bot Management configuration -- the singleton switchboard deciding how Cloudflare treats automated traffic, from the free Bot Fight Mode toggle through Super Bot Fight Mode's per-category actions up to the Enterprise Bot Management knobs and the AI-crawler controls.

A field left unset is NOT MANAGED: the module never sends it and the zone keeps whatever value it already carries. The configuration is a zone singleton with NO DELETE at Cloudflare -- destroying this resource abandons the last-applied values rather than reverting them. To retire a setting, set it to its off value and apply BEFORE destroying.

## Key Features

- **Presence-based management** -- unset fields are never sent; the zone keeps its current value
- **Four plan families** -- `fight_mode` (free); `sbfm_*` and `optimize_wordpress` (Pro/Business); Enterprise knobs (`auto_update_model`, `suppress_session_score`, `enable_js`, `bm_cookie_enabled`); AI and crawler controls
- **NO-OP destroy** -- Cloudflare has no delete; the last-applied values stay on the zone
- **Refresh-drift caution** -- non-entitled zones omit gated fields from API responses, which reads back as plan drift if the manifest set them

## Use Cases

**Ideal for:**

- Turning on Bot Fight Mode on a free zone
- Declaring Super Bot Fight Mode actions on a Pro or Business zone
- Blocking AI scrapers without touching the rest of Bot Management

**Not ideal for:**

- A static IP/ASN/country allow or block -- that is `CloudflareIpAccessRule`
- Expression-based skip rules for verified-bot exceptions -- that is `CloudflareRuleset`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone whose Bot Management configuration is managed. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |

At least one setting must be configured -- a resource that manages nothing would deploy nothing, and validation rejects it.

### Optional Fields

Every other field is optional, and unset means not managed. Fields are gated by the zone's plan; setting a field the plan does not include fails at the API.

| Field | Type | Description |
|-------|------|-------------|
| `fight_mode` | bool | Bot Fight Mode (free). Mutually exclusive with Super Bot Fight Mode at Cloudflare. |
| `sbfm_definitely_automated` | string | SBFM action on definitely-automated traffic: `allow`, `block`, `managed_challenge`. Pro+. |
| `sbfm_likely_automated` | string | SBFM action on likely-automated traffic. Business+. |
| `sbfm_verified_bots` | string | SBFM action on verified bots: `allow` or `block`. Blocking verified bots blocks search indexing. |
| `sbfm_static_resource_protection` | bool | Extend SBFM to static assets. Can catch legitimate hotlinked traffic. |
| `optimize_wordpress` | bool | Tune SBFM for WordPress (loopback, wp-cron). |
| `auto_update_model` | bool | Enterprise: adopt new bot-detection ML models automatically. |
| `suppress_session_score` | bool | Enterprise: stop tracking the session's highest bot score. |
| `enable_js` | bool | Enterprise: run invisible JavaScript detections. |
| `bm_cookie_enabled` | bool | Enterprise: allow the Bot Management cookie (Cloudflare defaults this to true). |
| `ai_bots_protection` | string | `block`, `disabled`, or `only_on_ad_pages`. |
| `crawler_protection` | string | `enabled` or `disabled` -- punish AI scrapers with a link maze. |
| `content_bots_protection` | string | `block` or `disabled`. |
| `cf_robots_variant` | string | `off` or `policy_only`. |
| `is_robots_txt_managed` | bool | Serve a Cloudflare-managed robots.txt prepended to the origin's. |

### Outputs

| Field | Description |
|-------|-------------|
| `zone_id` | The zone whose Bot Management configuration is managed (the singleton's identity) |

## Example Manifests

Bot Fight Mode on a free zone:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareBotManagement
metadata:
  name: fight-mode
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  fight_mode: true
  # Required alongside fight_mode when the zone's JS detections are off --
  # Cloudflare rejects Fight Mode on its own with "cannot enable Fight_Mode
  # while EnableJS is disabled".
  enable_js: true
```

## Destroy Semantics

Destroy is a NO-OP at Cloudflare. The state entry disappears; the live configuration stays exactly as last applied. To retire a setting, apply its off value first, then destroy. Leaving `fight_mode: true` and deleting the resource leaves Fight Mode on indefinitely.

## Related Resources

- **CloudflareIpAccessRule** -- a static IP/ASN/country decision; this kind is scored bot traffic
- **CloudflareRuleset** -- skip rules for verified-bot exceptions under `content_bots_protection`
- **CloudflareDnsZone** -- `zone_id` foreign key; the zone plan is what gates SBFM and Enterprise fields

## Further Reading

For operational judgment -- no-op destroy, unset-means-unmanaged, plan families, and refresh drift -- see GUIDE.md.

## References

- [Cloudflare Bot Management](https://developers.cloudflare.com/bots/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
