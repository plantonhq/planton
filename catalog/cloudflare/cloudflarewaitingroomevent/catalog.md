# Cloudflare Waiting Room Event

Deploys a scheduled event on a Cloudflare waiting room: a time window (a product launch, a ticket on-sale) during which the event's override values temporarily replace the room's own settings, with an optional prequeue that gathers early arrivals before the doors open. Scheduled events are a Waiting Rooms Advanced add-on feature per Cloudflare's plan matrix (the room itself is Business+). Every override is null-means-inherit — an unset field leaves the room's value in charge for the window. Events live on their own cadence, created and deleted per launch while the room persists, which is why they are their own kind rather than a field of the room.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Waiting Room Event** — one event on the named room and zone, active between `eventStartTime` and `eventEndTime`, carrying only the override fields the spec sets

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Zone → Waiting Room → Edit on the target zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A waiting room** — `waitingRoomId` names the room the event runs on; deploy a CloudflareWaitingRoom first and reference its output.
- **The room's zone** — `zoneId` is required alongside the room ID, and the two must agree; an event on room A in zone B fails at the API.
- **The Waiting Rooms Advanced add-on** (only for Advanced overrides) — an event's `customPageHtml` fails at apply if the zone lacks the add-on, same as on the room.

## Deploy

### Console

Open the deployment store, find **Cloudflare Waiting Room Event**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the room and zone references, the event window, and the optional overrides. Start from the **Launch window** preset in the [Presets](#presets) tab to pre-populate a no-override window.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWaitingRoomEvent
metadata:
  name: launch
  org: acme-corp
  env: prod
spec:
  waitingRoomId:
    value: "699d98642c564d2e855e9661899b7252"
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: launch
  eventStartTime: "2026-09-01T10:00:00Z"
  eventEndTime: "2026-09-01T14:00:00Z"
```

```shell
planton apply -f waiting-room-event.yaml
```

This creates a four-hour window with no overrides — the room queues with its own thresholds throughout. An Infra Job tracks the provisioning in real time.

### InfraChart

When the room and zone are deployed in the same InfraPipeline, wire both references with ValueFromRef:

```yaml
spec:
  waitingRoomId:
    valueFrom:
      kind: CloudflareWaitingRoom
      name: checkout-queue
      fieldPath: status.outputs.waiting_room_id
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  name: launch
  eventStartTime: "2026-09-01T10:00:00Z"
  eventEndTime: "2026-09-01T14:00:00Z"
```

The InfraPipeline resolves the dependency graph, deploys the zone and the room first, then provisions the event with the resolved IDs.

## Key Configuration

These are the most important decisions when configuring a waiting room event. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Null means inherit** — unset override fields are never sent to Cloudflare, so the room's value stays in charge during the window. Do not set a field to "confirm" the room's value; set only what the launch genuinely changes. The inherit rule covers `customPageHtml`, `disableSessionRenewal`, `queueingMethod`, `sessionDuration`, and both Turnstile fields.

**The users pair travels together** — `newUsersPerMinute` and `totalActiveUsers` override each other's math, so Cloudflare requires setting both or neither, and each has the same 200 floor as the room. The spec validates the pairing at manifest time.

**Time rules are validated up front** — times are RFC3339 (`2026-09-01T10:00:00Z`); `eventStartTime` must be at least one minute before `eventEndTime`, and `prequeueStartTime`, when set, at least five minutes before start. The spec enforces all three so a bad window fails at manifest time instead of as an opaque API error.

**Prequeue and shuffle** — `prequeueStartTime` opens an early line before the doors open; `shuffleAtEventStart` randomizes that line at the door instead of first-come-first-served, the fairness tool for on-sales where arrival time is a bot advantage. Shuffle requires a prequeue (there is nothing to shuffle otherwise) and only makes sense when the queueing method respects order (`fifo`).

**Cancelling versus deleting** — destroy is a real delete of the event; the room is untouched and returns to its own settings. For a cancelled launch you may reschedule, set `suspended: true` instead — the window comes and goes with no effect while the event's ID survives.

**Delete order matters** — destroying the room first leaves events pointing at a deleted object. Delete the events before the room, or accept the lookup failure.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareWaitingRoom** | `waitingRoomId` | `status.outputs.waiting_room_id` |
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `event_id` | The created event's ID | Dashboard lookup and import recipes |
| `waiting_room_id` | The room the event runs on | Correlating events to their room in a pipeline |
| `zone_id` | The zone the room belongs to | Zone-scoped siblings deployed in the same pipeline |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Launch window, no overrides** — a named window that keeps the room's thresholds in charge; useful when the launch just needs visibility in the dashboard's event list. Start from the **Launch window** preset.

**On-sale with a shuffled prequeue** — open a prequeue thirty minutes early, shuffle at the door, and raise the users pair together (`newUsersPerMinute` + `totalActiveUsers`) for the sale window. Delete the event after the date; the room persists.

**Maintenance clamp** — an event that overrides `queueingMethod: reject` for a window where the origin cannot take traffic at all, without touching the room's steady-state configuration.

## Works With

- [**Cloudflare Waiting Room**](/infra-catalog/cloudflare-waiting-room) — the room this event overrides; create it first and wire `waitingRoomId` from its output
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone the room belongs to; `zoneId` references its `zone_id` output
