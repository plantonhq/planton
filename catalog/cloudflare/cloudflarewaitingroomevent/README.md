# Cloudflare Waiting Room Event

## Overview

`CloudflareWaitingRoomEvent` is one scheduled event on a waiting room: a time window (a product launch, a ticket on-sale) during which the event's override values temporarily replace the room's own settings, with an optional prequeue that gathers early arrivals before the doors open.

Events live on their own cadence -- created and deleted per launch while the room persists -- which is why they are their own kind rather than a field of the room. Every override field is optional and null-means-inherit: an unset override leaves the room's value in charge for the event window.

## Key Features

- **Own cadence** -- create and delete per launch; the room stays
- **Foreign keys** -- `waiting_room_id` (defaults to `CloudflareWaitingRoom` / `status.outputs.waiting_room_id`) and `zone_id`
- **Null-means-inherit** -- unset overrides are never sent; the room's value stays in charge
- **RFC3339 times** -- start at least one minute before end; prequeue at least five minutes before start; shuffle requires a prequeue; the users pair is both-or-neither
- **Real delete** -- destroy removes the event; the room is untouched

## Use Cases

**Ideal for:**

- A product-launch window that temporarily raises the room's admission rate
- A prequeue that gathers early arrivals and shuffles them at the door
- A one-off on-sale you will delete after the date

**Not ideal for:**

- The room itself -- that is `CloudflareWaitingRoom`
- A permanent change to the room's thresholds -- set those on the room, not as an event with no end

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `waiting_room_id` | StringValueOrRef | Yes | The room the event runs on. Can reference a `CloudflareWaitingRoom` via `value_from` (defaults to `status.outputs.waiting_room_id`). |
| `zone_id` | StringValueOrRef | Yes | The zone the room belongs to. Can reference a `CloudflareDnsZone` via `value_from` (defaults to `status.outputs.zone_id`). |
| `name` | string | Yes | Display name. |
| `event_start_time` | string | Yes | RFC3339. At least one minute before `event_end_time`. |
| `event_end_time` | string | Yes | RFC3339. |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `prequeue_start_time` | string | RFC3339. At least five minutes before start. Unset means no prequeue. |
| `shuffle_at_event_start` | bool | Randomize prequeued users when the event starts. Requires a prequeue. |
| `description` | string | Dashboard note. |
| `suspended` | bool | The window comes and goes with no effect while suspended. |
| `custom_page_html` | string | Override the room's queue page (Advanced entitlement applies). |
| `disable_session_renewal` | bool | Override the room's session-renewal setting. |
| `new_users_per_minute` | int32 | Override the room's rate. Floor 200. Must be set together with `total_active_users`. |
| `total_active_users` | int32 | Override the room's cap. Floor 200. Must be set together with `new_users_per_minute`. |
| `queueing_method` | string | Override: `fifo`, `random`, `passthrough`, `reject`. |
| `session_duration` | int32 | Override: 1–30 minutes. |
| `turnstile_action` | string | Override: `log` or `infinite_queue`. |
| `turnstile_mode` | string | Override: `off`, `invisible`, `visible_non_interactive`, `visible_managed`. |

### Outputs

| Field | Description |
|-------|-------------|
| `event_id` | The created event's ID |
| `waiting_room_id` | The room the event runs on |
| `zone_id` | The zone the waiting room belongs to |

## Example Manifests

A launch window with no overrides -- the room's settings stay in charge:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWaitingRoomEvent
metadata:
  name: launch
spec:
  waiting_room_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: launch
  event_start_time: "2026-09-01T10:00:00Z" # replace with your start
  event_end_time: "2026-09-01T14:00:00Z"   # replace with your end
```

## Destroy Semantics

Destroy is a real delete. The event is removed; the room is untouched and returns to its own settings. Suspending (`suspended: true`) is the reversible alternative when you want the ID to stay through a cancelled window.

## Related Resources

- **CloudflareWaitingRoom** -- the room this event overrides; create it first
- **CloudflareDnsZone** -- `zone_id` foreign key

## Further Reading

For operational judgment -- null-means-inherit, the time rules, shuffle-needs-prequeue, and the users pair -- see GUIDE.md.

## References

- [Cloudflare Waiting Room events](https://developers.cloudflare.com/waiting-room/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
