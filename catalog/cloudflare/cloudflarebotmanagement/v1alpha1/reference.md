# CloudflareBotManagement

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareBotManagementSpec manages a zone's Bot Management configuration --
the singleton switchboard deciding how Cloudflare treats automated traffic,
from the free Bot Fight Mode toggle through Super Bot Fight Mode's per-category
actions up to the Enterprise Bot Management knobs and the AI-crawler controls.

A field left unset is NOT MANAGED: the module never sends it and the zone keeps
whatever value it already carries. The configuration is a zone singleton with
NO DELETE at Cloudflare -- destroying this resource abandons the last-applied
values rather than reverting them (Cloudflare's own tooling warns the same).
To retire a setting, set it to its off value and apply BEFORE destroying.

Fields are gated by the zone's plan, in four families:
  - Free: fight_mode.
  - Pro/Business (Super Bot Fight Mode): sbfm_definitely_automated,
    sbfm_likely_automated (Business+), sbfm_verified_bots,
    sbfm_static_resource_protection, optimize_wordpress.
  - Enterprise Bot Management: auto_update_model, suppress_session_score,
    enable_js, bm_cookie_enabled.
  - AI & crawler controls (plan-independent rollout): ai_bots_protection,
    crawler_protection, content_bots_protection, cf_robots_variant,
    is_robots_txt_managed.
Setting a field the zone's plan does not include fails at the API, not here --
the plan walls are Cloudflare's, and they move; the families above are the
provider-documented grouping.

## Example

```yaml
# A complete, protovalidate-valid CloudflareBotManagement example: the free
# Bot Fight Mode toggle plus the AI-crawler controls. SBFM and Enterprise
# fields are omitted -- set them only on zones whose plan includes them (the
# API rejects fields the plan does not carry).
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareBotManagement
metadata:
  name: bot-management-posture
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  fight_mode: true
  # Cloudflare requires the zone's JS detections on before Fight Mode enables.
  enable_js: true
  ai_bots_protection: block
  crawler_protection: enabled
  is_robots_txt_managed: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.fightMode` | `bool` |  |  |  |
| `spec.sbfmDefinitelyAutomated` | `string` |  |  |  |
| `spec.sbfmLikelyAutomated` | `string` |  |  |  |
| `spec.sbfmVerifiedBots` | `string` |  |  |  |
| `spec.sbfmStaticResourceProtection` | `bool` |  |  |  |
| `spec.optimizeWordpress` | `bool` |  |  |  |
| `spec.autoUpdateModel` | `bool` |  |  |  |
| `spec.suppressSessionScore` | `bool` |  |  |  |
| `spec.enableJs` | `bool` |  |  |  |
| `spec.bmCookieEnabled` | `bool` |  |  |  |
| `spec.aiBotsProtection` | `string` |  |  |  |
| `spec.crawlerProtection` | `string` |  |  |  |
| `spec.contentBotsProtection` | `string` |  |  |  |
| `spec.cfRobotsVariant` | `string` |  |  |  |
| `spec.isRobotsTxtManaged` | `bool` |  |  |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone whose Bot Management configuration is managed.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.fightMode

`bool` · optional (explicit presence)

Bot Fight Mode (free plans): challenge requests matching known bot patterns.
Mutually exclusive with Super Bot Fight Mode at Cloudflare -- zones on SBFM
plans manage the sbfm_* fields instead. Cloudflare refuses to enable Fight
Mode while the zone's JavaScript detections are off (400 code 10400 "cannot
enable Fight_Mode while EnableJS is disabled", measured live 2026-08-27):
when enabling from scratch, declare enable_js: true alongside. A zone whose
JS detections are already on accepts fight_mode alone -- the constraint is
zone-state-dependent, which is why it is not a validation rule here.

### spec.sbfmDefinitelyAutomated

`string` · optional (explicit presence)

Super Bot Fight Mode: action on requests Cloudflare scores as DEFINITELY
automated. Pro plans and above.

- rule: {"string":{"in":["allow","block","managed_challenge"]}}

### spec.sbfmLikelyAutomated

`string` · optional (explicit presence)

Super Bot Fight Mode: action on requests Cloudflare scores as LIKELY
automated. Business plans and above.

- rule: {"string":{"in":["allow","block","managed_challenge"]}}

### spec.sbfmVerifiedBots

`string` · optional (explicit presence)

Super Bot Fight Mode: action on verified bots (search-engine crawlers and
other bots Cloudflare has verified). Blocking verified bots blocks search
indexing -- allow is the safe default posture.

- rule: {"string":{"in":["allow","block"]}}

### spec.sbfmStaticResourceProtection

`bool` · optional (explicit presence)

Super Bot Fight Mode: extend protection to static resources (images, CSS,
JS). Can block legitimate hotlinked-asset traffic -- Cloudflare's own docs
caution that static protection may catch real users.

### spec.optimizeWordpress

`bool` · optional (explicit presence)

Super Bot Fight Mode: tune protections for WordPress sites (loopback
requests, wp-cron and kin).

### spec.autoUpdateModel

`bool` · optional (explicit presence)

Enterprise Bot Management: automatically adopt new bot-detection ML models
as Cloudflare releases them.

### spec.suppressSessionScore

`bool` · optional (explicit presence)

Enterprise Bot Management: stop tracking the session's highest bot score in
the Bot Management cookie.

### spec.enableJs

`bool` · optional (explicit presence)

Run Cloudflare's lightweight invisible JavaScript detections to sharpen bot
scoring. Writable on every plan (measured live on a free zone 2026-08-27,
despite older docs labeling it Enterprise-only) and REQUIRED to be on
before or with fight_mode -- see that field's pair rule.

### spec.bmCookieEnabled

`bool` · optional (explicit presence)

Enterprise Bot Management: allow the Bot Management cookie to be placed on
end-user devices (Cloudflare defaults this to true).

### spec.aiBotsProtection

`string` · optional (explicit presence)

AI crawler control: block AI scrapers and crawlers. only_on_ad_pages limits
the block to pages carrying ads.

- rule: {"string":{"in":["block","disabled","only_on_ad_pages"]}}

### spec.crawlerProtection

`string` · optional (explicit presence)

AI crawler control: punish AI scrapers with a link maze instead of a plain
block.

- rule: {"string":{"in":["enabled","disabled"]}}

### spec.contentBotsProtection

`string` · optional (explicit presence)

Content-bot control: block automated traffic with low bot scores, excluding
Cloudflare's safe verified-bot categories. Manage exceptions via skip rules
in a CloudflareRuleset.

- rule: {"string":{"in":["block","disabled"]}}

### spec.cfRobotsVariant

`string` · optional (explicit presence)

Robots Access Control License variant: policy_only publishes the license
policy without enforcement; off disables it.

- rule: {"string":{"in":["off","policy_only"]}}

### spec.isRobotsTxtManaged

`bool` · optional (explicit presence)

Serve a Cloudflare-managed robots.txt (prepended to any existing robots.txt
the origin serves).

## Validation Rules

- `spec.at_least_one_setting`: configure at least one bot-management setting -- a CloudflareBotManagement resource that manages nothing would deploy nothing

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareBotManagement, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The zone whose Bot Management configuration is managed. The configuration is a zone singleton -- the zone ID is its identity (Cloudflare's rule ID for this surface IS the zone ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
