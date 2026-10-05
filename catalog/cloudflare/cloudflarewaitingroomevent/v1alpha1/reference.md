# CloudflareWaitingRoomEvent

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `cloudflare.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

CloudflareWaitingRoomEventSpec defines one scheduled event on a waiting room:
a time window (a product launch, a ticket on-sale) during which the event's
override values temporarily replace the room's own settings, with an optional
prequeue that gathers early arrivals before the doors open.

Events live on their own cadence -- created and deleted per launch while the
room persists -- which is why they are their own kind rather than a field of
the room.

Every override field is OPTIONAL and null-means-inherit: an unset override
leaves the room's value in charge for the event window; a set one replaces it
while the event is active. Cloudflare's API enforces the time rules (start at
least one minute before end; prequeue at least five minutes before start) --
this spec validates them up front so a bad window fails at manifest time, not
at apply.

## Example

```yaml
# A complete, protovalidate-valid CloudflareWaitingRoomEvent example: a
# prequeued, shuffled on-sale window with rate overrides. All time rules
# (start >= 1 min before end, prequeue >= 5 min before start) validate at
# manifest time.
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWaitingRoomEvent
metadata:
  name: ticket-onsale
spec:
  waiting_room_id:
    value: "699d98642c564d2e855e9661899b7252"
  zone_id:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: ticket-onsale
  event_start_time: "2030-09-01T10:00:00Z"
  event_end_time: "2030-09-01T14:00:00Z"
  prequeue_start_time: "2030-09-01T09:30:00Z"
  shuffle_at_event_start: true
  description: "Season ticket on-sale window"
  new_users_per_minute: 300
  total_active_users: 1500
  session_duration: 5
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.waitingRoomId` | `string \| valueFrom` | yes |  | CloudflareWaitingRoom (`status.outputs.waiting_room_id`) |
| `spec.zoneId` | `string \| valueFrom` | yes |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.eventStartTime` | `string` | yes |  |  |
| `spec.eventEndTime` | `string` | yes |  |  |
| `spec.prequeueStartTime` | `string` |  |  |  |
| `spec.shuffleAtEventStart` | `bool` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.suspended` | `bool` |  |  |  |
| `spec.customPageHtml` | `string` |  |  |  |
| `spec.disableSessionRenewal` | `bool` |  |  |  |
| `spec.newUsersPerMinute` | `int32` |  |  |  |
| `spec.totalActiveUsers` | `int32` |  |  |  |
| `spec.queueingMethod` | `string` |  |  |  |
| `spec.sessionDuration` | `int32` |  |  |  |
| `spec.turnstileAction` | `string` |  |  |  |
| `spec.turnstileMode` | `string` |  |  |  |

## Field Details

### spec.waitingRoomId

`string | valueFrom` · required

The waiting room the event runs on.
When using value_from, defaults to CloudflareWaitingRoom kind and
status.outputs.waiting_room_id field path.

- references: CloudflareWaitingRoom (`status.outputs.waiting_room_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareWaitingRoom, name: <that resource's name>, fieldPath: status.outputs.waiting_room_id}} -- a bare string does not parse

### spec.zoneId

`string | valueFrom` · required

The zone the waiting room belongs to.
When using value_from, defaults to CloudflareDnsZone kind and status.outputs.zone_id field path.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.name

`string` · required

A name for the event (shown in the dashboard).

- rule: {"string":{"minLen":"1"}}

### spec.eventStartTime

`string` · required

When the event window opens, as RFC3339 (e.g. 2026-09-01T10:00:00Z). Must
be at least one minute before event_end_time.

- rule: event_start_time must be an RFC3339 timestamp (e.g. 2026-09-01T10:00:00Z)
- rule: {"required":true}

### spec.eventEndTime

`string` · required

When the event window closes, as RFC3339.

- rule: event_end_time must be an RFC3339 timestamp (e.g. 2026-09-01T14:00:00Z)
- rule: {"required":true}

### spec.prequeueStartTime

`string`

When early arrivals may start gathering in the prequeue, as RFC3339. Must
be at least five minutes before event_start_time. Unset means no prequeue.

- rule: prequeue_start_time must be an RFC3339 timestamp (e.g. 2026-09-01T09:30:00Z) or empty

### spec.shuffleAtEventStart

`bool` · optional (explicit presence)

Randomize the order of prequeued users when the event starts, instead of
first-come-first-served -- fairness for on-sales where arrival time is a
bot advantage. Requires a prequeue.

### spec.description

`string`

Description shown in the dashboard's event list.

### spec.suspended

`bool` · optional (explicit presence)

Suspend the event without deleting it -- the window comes and goes with no
effect while suspended.

### spec.customPageHtml

`string`

Override: custom queue-page HTML while the event is active (the room's
Advanced entitlement applies). Unset inherits the room's page.

### spec.disableSessionRenewal

`bool` · optional (explicit presence)

Override: freeze session renewal during the event (see the room field of
the same name). Unset inherits the room's setting.

### spec.newUsersPerMinute

`int32` · optional (explicit presence)

Override: new-users-per-minute admission rate during the event. Cloudflare
requires setting this together with total_active_users (both or neither);
the floor is 200. Unset inherits the room's rate.

- rule: {"int32":{"gte":200}}

### spec.totalActiveUsers

`int32` · optional (explicit presence)

Override: total concurrent active users during the event. Set together
with new_users_per_minute (both or neither); the floor is 200. Unset
inherits the room's cap.

- rule: {"int32":{"gte":200}}

### spec.queueingMethod

`string` · optional (explicit presence)

Override: queueing method during the event. Cloudflare's own schema leaves
this unvalidated on events -- the wall here mirrors the room's documented
set (a deliberate tightening).

- rule: {"string":{"in":["fifo","random","passthrough","reject"]}}

### spec.sessionDuration

`int32` · optional (explicit presence)

Override: session duration in minutes during the event (1-30). Unset
inherits the room's duration.

- rule: {"int32":{"lte":30,"gte":1}}

### spec.turnstileAction

`string` · optional (explicit presence)

Override: Turnstile action during the event (see the room field). Unset
inherits the room's action.

- rule: {"string":{"in":["log","infinite_queue"]}}

### spec.turnstileMode

`string` · optional (explicit presence)

Override: Turnstile mode during the event (see the room field). Unset
inherits the room's mode.

- rule: {"string":{"in":["off","invisible","visible_non_interactive","visible_managed"]}}

## Validation Rules

- `spec.start_before_end`: event_start_time must be at least one minute before event_end_time
- `spec.prequeue_five_minutes_before`: prequeue_start_time must be at least five minutes before event_start_time
- `spec.shuffle_requires_prequeue`: shuffle_at_event_start needs a prequeue to shuffle -- set prequeue_start_time
- `spec.users_pair_both_or_neither`: new_users_per_minute and total_active_users override each other's math -- Cloudflare requires setting both together or neither

## Outputs

Reference an output from another manifest as `valueFrom: {kind: CloudflareWaitingRoomEvent, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.event_id` | `string` | The ID of the created event. |
| `status.outputs.waiting_room_id` | `string` | The waiting room the event runs on. |
| `status.outputs.zone_id` | `string` | The zone the waiting room belongs to. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.waitingRoomId` | CloudflareWaitingRoom | `status.outputs.waiting_room_id` |
| `spec.zoneId` | CloudflareDnsZone | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
