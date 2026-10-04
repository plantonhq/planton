# CloudflareWaitingRoom

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareWaitingRoomSpec defines one waiting room: a virtual queue in front
of a host+path that admits visitors at a controlled rate and parks the
overflow on a branded queue page. The room's thresholds (new users per
minute, total active users) decide when queueing kicks in; everything else
shapes the queueing experience.

The room's BYPASS RULES ride along in this spec (bypass_rules): Cloudflare
models them as a separate per-room rules list with full-replacement updates,
and this kind manages that list as part of the room -- every apply replaces
the room's whole rules list with exactly what the manifest declares.

Several fields need the WAITING ROOMS ADVANCED add-on (marked "Advanced" in
their comments): additional_routes, custom_page_html, disable_session_renewal,
json_response_enabled, a non-default queueing_method, the infinite_queue
turnstile action, and non-off turnstile modes. Cloudflare enforces the
entitlement at the API -- setting an Advanced field on a plan without the
add-on fails at apply, not here.

## Example

```yaml
# A complete, protovalidate-valid CloudflareWaitingRoom example: a basic room
# on a shop's checkout path with one bypass rule. Free-plan-safe shape --
# Advanced-subscription fields (additional_routes, custom_page_html, non-fifo
# queueing, non-off turnstile) appear in the presets, gated per field.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWaitingRoom
metadata:
  name: checkout-queue
spec:
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: checkout-queue
  host: shop.example.com
  path: /checkout
  new_users_per_minute: 300
  total_active_users: 1000
  session_duration: 5
  cookie_attributes:
    samesite: lax
    secure: auto
  bypass_rules:
    - expression: 'ip.src in {203.0.113.0/24}'
      description: "Office network skips the queue"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.host` | `string` | yes |  |  |
| `spec.path` | `string` |  |  |  |
| `spec.newUsersPerMinute` | `int32` |  |  |  |
| `spec.totalActiveUsers` | `int32` |  |  |  |
| `spec.sessionDuration` | `int32` |  |  |  |
| `spec.suspended` | `bool` |  |  |  |
| `spec.queueAll` | `bool` |  |  |  |
| `spec.queueingMethod` | `string` |  | `fifo` |  |
| `spec.queueingStatusCode` | `int32` |  |  |  |
| `spec.cookieAttributes` | `CloudflareWaitingRoomCookieAttributes` |  |  |  |
| `spec.cookieAttributes.samesite` | `string` |  | `auto` |  |
| `spec.cookieAttributes.secure` | `string` |  | `auto` |  |
| `spec.cookieSuffix` | `string` |  |  |  |
| `spec.customPageHtml` | `string` |  |  |  |
| `spec.defaultTemplateLanguage` | `string` |  | `en-US` |  |
| `spec.description` | `string` |  |  |  |
| `spec.disableSessionRenewal` | `bool` |  |  |  |
| `spec.jsonResponseEnabled` | `bool` |  |  |  |
| `spec.additionalRoutes` | `[]CloudflareWaitingRoomRoute` |  |  |  |
| `spec.additionalRoutes[].host` | `string` | yes |  |  |
| `spec.additionalRoutes[].path` | `string` |  |  |  |
| `spec.enabledOriginCommands` | `[]string` |  |  |  |
| `spec.turnstileAction` | `string` |  | `log` |  |
| `spec.turnstileMode` | `string` |  | `invisible` |  |
| `spec.bypassRules` | `[]CloudflareWaitingRoomBypassRule` |  |  |  |
| `spec.bypassRules[].expression` | `string` | yes |  |  |
| `spec.bypassRules[].description` | `string` |  |  |  |
| `spec.bypassRules[].enabled` | `bool` |  | `true` |  |

## Field Details

### spec.zoneId

`string | valueFrom` · required

The zone the waiting room belongs to.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

A name for the waiting room (shown in the dashboard).

- rule: {"string":{"minLen":"1"}}

### spec.host

`string` · required

The host the room protects (e.g. shop.example.com). Together with path this
is WHERE the queue sits.

- rule: {"required":true}

### spec.path

`string` · optional (explicit presence)

The path the room protects (Cloudflare's default: /, the whole host). A
more specific path scopes the queue to that subtree.

### spec.newUsersPerMinute

`int32`

How many NEW users per minute may enter the protected route before new
arrivals queue. Cloudflare's floor is 200.

- rule: {"int32":{"gte":200}}

### spec.totalActiveUsers

`int32`

How many users may be on the protected route AT ONCE before new arrivals
queue. Cloudflare's floor is 200.

- rule: {"int32":{"gte":200}}

### spec.sessionDuration

`int32` · optional (explicit presence)

Minutes a user's slot stays valid after leaving the route before they must
re-queue (Cloudflare's default: 5; the ceiling is 30).

- rule: {"int32":{"lte":30,"gte":1}}

### spec.suspended

`bool` · optional (explicit presence)

Pause the room's queueing without deleting its configuration.

### spec.queueAll

`bool` · optional (explicit presence)

Send everyone to the queue regardless of thresholds -- the "close the shop"
switch for maintenance windows.

### spec.queueingMethod

`string` · optional (explicit presence)

How queued users are admitted (Cloudflare's default: fifo).
  - fifo: strict arrival order.
  - random: random draw each admission cycle (fairer under reload-storms).
  - passthrough: nobody queues; useful for testing the room's wiring.
  - reject: turn overflow away instead of queueing.
Non-default methods need the Advanced add-on.

- default: `fifo`
- rule: {"string":{"in":["fifo","random","passthrough","reject"]}}

### spec.queueingStatusCode

`int32` · optional (explicit presence)

HTTP status the queue page answers with (Cloudflare's default: 200; 202 and
429 are the alternatives some monitoring stacks prefer).

- rule: {"int32":{"in":[200,202,429]}}

### spec.cookieAttributes

`CloudflareWaitingRoomCookieAttributes`

Attributes of the __cfwaitingroom cookie that tracks a visitor's place.

### spec.cookieAttributes.samesite

`string` · optional (explicit presence)

SameSite attribute (Cloudflare's default: auto -- None when Always Use
HTTPS is on, Lax otherwise). Note: none requires the cookie be Secure.

- default: `auto`
- rule: {"string":{"in":["auto","lax","none","strict"]}}

### spec.cookieAttributes.secure

`string` · optional (explicit presence)

Secure attribute (Cloudflare's default: auto -- Secure when Always Use
HTTPS is on).

- default: `auto`
- rule: {"string":{"in":["auto","always","never"]}}

### spec.cookieSuffix

`string`

Suffix appended to the __cfwaitingroom cookie name, for hosts running
several rooms whose cookies must not collide.

### spec.customPageHtml

`string`

Custom HTML for the queue page (Advanced). Supports Cloudflare's waiting
room template variables. Unset serves Cloudflare's default queue page in
default_template_language.

### spec.defaultTemplateLanguage

`string` · optional (explicit presence)

Language of Cloudflare's default queue page (Cloudflare's default: en-US).
Ignored when custom_page_html is set.

- default: `en-US`
- rule: {"string":{"in":["en-US","es-ES","de-DE","fr-FR","it-IT","ja-JP","ko-KR","pt-BR","zh-CN","zh-TW","nl-NL","pl-PL","id-ID","tr-TR","ar-EG","ru-RU","fa-IR","bg-BG","hr-HR","cs-CZ","da-DK","fi-FI","lt-LT","ms-MY","nb-NO","ro-RO","el-GR","he-IL","hi-IN","hu-HU","sr-BA","sk-SK","sl-SI","sv-SE","tl-PH","th-TH","uk-UA","vi-VN"]}}

### spec.description

`string`

Description shown in the dashboard's room list.

### spec.disableSessionRenewal

`bool` · optional (explicit presence)

Do not renew a user's session while they stay on the route (Advanced): the
session expires session_duration minutes after ENTRY instead of after their
last request, forcing periodic re-queueing under sustained load.

### spec.jsonResponseEnabled

`bool` · optional (explicit presence)

Answer queued API clients with JSON instead of the HTML queue page
(Advanced) -- for native apps and XHR callers that cannot render HTML.

### spec.additionalRoutes

`[]CloudflareWaitingRoomRoute`

Additional host+path routes covered by the SAME room and quota (Advanced) --
e.g. queue www and apex together so a visitor holds one place across both.

### spec.additionalRoutes[].host

`string` · required

The additional host (e.g. www.example.com).

- rule: {"required":true}

### spec.additionalRoutes[].path

`string` · optional (explicit presence)

The path on that host (Cloudflare's default: /).

### spec.enabledOriginCommands

`[]string`

Origin commands the room accepts from the protected origin.
Cloudflare currently supports only "revoke" (revoke a visitor's session
from the origin side).

- rule: {"repeated":{"items":{"string":{"in":["revoke"]}}}}

### spec.turnstileAction

`string` · optional (explicit presence)

What happens when Turnstile (the room's bot check) flags a visitor
(Cloudflare's default: log).
  - log: record it and let them through.
  - infinite_queue: park suspected bots in a queue that never admits
    (Advanced).

- default: `log`
- rule: {"string":{"in":["log","infinite_queue"]}}

### spec.turnstileMode

`string` · optional (explicit presence)

How Turnstile challenges queued visitors (Cloudflare's default: invisible;
off disables it). Non-off modes need the Advanced add-on.

- default: `invisible`
- rule: {"string":{"in":["off","invisible","visible_non_interactive","visible_managed"]}}

### spec.bypassRules

`[]CloudflareWaitingRoomBypassRule`

The room's bypass rules: requests matching an expression skip the queue
entirely (the action is fixed to bypass_waiting_room -- the module supplies
it). Cloudflare stores these as a per-room list replaced WHOLE on every
apply -- this list is the entire rule set, and destroying the room's rules
(or the room) clears rules added outside the manifest too.

### spec.bypassRules[].expression

`string` · required

The match expression in Cloudflare's Rules language (wirefilter), e.g.
`ip.src in {203.0.113.0/24}` to let an office network skip the queue.

- rule: {"required":true}

### spec.bypassRules[].description

`string`

Optional description shown in the dashboard's rule list.

### spec.bypassRules[].enabled

`bool` · optional (explicit presence)

Whether the rule is active (Cloudflare's default for waiting-room rules:
true -- unlike snippet rules, a declared bypass rule runs unless disabled).

- default: `true`

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareWaitingRoom, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.waiting_room_id` | `string` | The ID of the created waiting room -- what events and the import recipe reference. |
| `status.outputs.zone_id` | `string` | The zone the waiting room belongs to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| CloudflareWaitingRoomEvent | `spec.waitingRoomId` | `status.outputs.waiting_room_id` |

## See Also

- [Overview](../README.md)
