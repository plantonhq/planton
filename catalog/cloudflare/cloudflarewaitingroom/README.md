# Cloudflare Waiting Room

## Overview

`CloudflareWaitingRoom` is a virtual queue in front of a host+path that admits visitors at a controlled rate and parks the overflow on a branded queue page. The room's thresholds (new users per minute, total active users) decide when queueing kicks in; everything else shapes the queueing experience.

Waiting Room is a **Business+ product**: a Free or Pro zone rejects every room create with "Zone not entitled to this functionality" (code 1034, measured live). The `bypass_rules` list and scheduled events additionally require the Enterprise advanced add-on.

The room's bypass rules ride along in this spec (`bypass_rules`). Cloudflare models them as a separate per-room rules list with full-replacement updates, and this kind manages that list as part of the room -- every apply replaces the room's whole rules list with exactly what the manifest declares.

## Key Features

- **Host + path queue** -- floors of 200 new users per minute and 200 total active users; session duration 1–30 minutes
- **Folded bypass rules** -- `bypass_rules` is the room's entire rule set; destroy PUTs `[]` and wipes rules added outside the manifest
- **Advanced add-on fields** -- `additional_routes`, `custom_page_html`, `disable_session_renewal`, `json_response_enabled`, a non-default `queueing_method`, `infinite_queue`, and non-off Turnstile modes fail at the API without the add-on
- **Real delete** -- destroy removes the room; events that referenced it start failing their lookup

## Use Cases

**Ideal for:**

- A checkout or ticket-sale path that must not melt the origin
- Letting an office CIDR skip the queue via a bypass rule
- A maintenance window (`queue_all: true`) that parks everyone without deleting the room

**Not ideal for:**

- A scheduled launch window that overrides the room's rates -- that is `CloudflareWaitingRoomEvent`
- Bot scoring for the rest of the zone -- that is `CloudflareBotManagement`

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `zone_id` | StringValueOrRef | Yes | The zone the waiting room belongs to. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `name` | string | Yes | Display name shown in the dashboard. |
| `host` | string | Yes | The host the room protects (e.g. `shop.example.com`). |
| `new_users_per_minute` | int32 | Yes | New admissions per minute before queueing. Cloudflare's floor is 200. |
| `total_active_users` | int32 | Yes | Concurrent users on the route before queueing. Cloudflare's floor is 200. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `path` | string | Path the room protects (Cloudflare default: `/`). |
| `session_duration` | int32 | Minutes a slot stays valid (1–30; Cloudflare default: 5). |
| `suspended` | bool | Pause queueing without deleting the room. |
| `queue_all` | bool | Send everyone to the queue regardless of thresholds. |
| `queueing_method` | string | `fifo` (default), `random`, `passthrough`, `reject`. Non-default methods need Advanced. |
| `queueing_status_code` | int32 | `200` (default), `202`, or `429`. |
| `cookie_attributes` | object | `{samesite, secure}` of the `__cfwaitingroom` cookie. |
| `cookie_suffix` | string | Suffix so several rooms on one host do not collide cookies. |
| `custom_page_html` | string | Custom queue page (Advanced). |
| `default_template_language` | string | Language of Cloudflare's default page (default `en-US`; 38-language wall). |
| `description` | string | Shown in the dashboard. |
| `disable_session_renewal` | bool | Expire the session from entry, not last request (Advanced). |
| `json_response_enabled` | bool | JSON queue response for API clients (Advanced). |
| `additional_routes` | list | Extra host+path covered by the same quota (Advanced). |
| `enabled_origin_commands` | list | Currently only `revoke`. |
| `turnstile_action` | string | `log` (default) or `infinite_queue` (Advanced). |
| `turnstile_mode` | string | `off`, `invisible` (default), `visible_non_interactive`, `visible_managed`. Non-off needs Advanced. |
| `bypass_rules` | list | The room's entire bypass-rule table. Action is fixed to `bypass_waiting_room`. |

### Outputs

| Field | Description |
|-------|-------------|
| `waiting_room_id` | The created room's ID -- what events and the import recipe reference |
| `zone_id` | The zone the waiting room belongs to |

## Example Manifests

A basic checkout queue with one office bypass:

```yaml
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
  new_users_per_minute: 200
  total_active_users: 200
  bypass_rules:
    - expression: 'ip.src in {203.0.113.0/24}'
      description: Office network skips the queue
```

## Destroy Semantics

Destroy is a real delete of the room. The folded bypass-rules resource PUTs `[]` first, which wipes rules added outside this manifest too. Events that referenced `waiting_room_id` start failing their lookup -- delete or retarget them first. Suspending (`suspended: true`) or `queue_all: true` is the reversible alternative when you want the ID to stay.

## Related Resources

- **CloudflareWaitingRoomEvent** -- a scheduled window that temporarily overrides this room's rates
- **CloudflareDnsZone** -- `zone_id` foreign key
- **CloudflareBotManagement** -- zone-wide bot scoring; this kind is a queue, not a bot score

## Further Reading

For operational judgment -- the 200 floors, Advanced entitlement, and bypass-rules full-replacement -- see GUIDE.md.

## References

- [Cloudflare Waiting Room](https://developers.cloudflare.com/waiting-room/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
