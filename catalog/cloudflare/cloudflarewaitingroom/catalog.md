# Cloudflare Waiting Room

Deploys a Cloudflare waiting room: a virtual queue in front of a host and path that admits visitors at a controlled rate and parks the overflow on a queue page. Waiting Room is a Business+ product — a Free or Pro zone rejects every room with "Zone not entitled" (code 1034, measured live). The room's bypass rules ride along in the same manifest and replace the room's entire rule list on every apply; the rules themselves need the Waiting Rooms Advanced add-on, as do scheduled events — add-on fields fail at the Cloudflare API on plans without it. The add-on-free shape is host, path, thresholds at the 200 floor, fifo admission, and default Turnstile.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Waiting Room** — one waiting room on the zone, scoped to `host` + `path`, queueing when `newUsersPerMinute` or `totalActiveUsers` is exceeded
- **Waiting Room Rules** — created only when `bypassRules` is non-empty; one rules list holding the room's ENTIRE bypass-rule set (the action is fixed to `bypass_waiting_room` — the module supplies it), replaced whole on every apply

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** — an active connection in the Connect module with an API token holding Zone → Waiting Room → Edit on the target zone. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** — required when using Runner-based credential delivery. Not needed for inline credential authentication.

### Cloudflare Account

- **A zone on the account** — `zoneId` names the zone the room protects; reference a CloudflareDnsZone Infra Component or pass the 32-character zone ID.
- **The Waiting Rooms Advanced add-on** (only for Advanced fields) — `additionalRoutes`, `customPageHtml`, `disableSessionRenewal`, `jsonResponseEnabled`, a non-fifo `queueingMethod`, `turnstileAction: infinite_queue`, and any `turnstileMode` other than `off` all fail at apply on a plan without it.

## Deploy

### Console

Open the deployment store, find **Cloudflare Waiting Room**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the zone and protected host/path, the admission thresholds, and the queueing-experience fields. Start from the **Basic checkout queue** preset in the [Presets](#presets) tab to pre-populate a free-plan-safe configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareWaitingRoom
metadata:
  name: checkout-queue
  org: acme-corp
  env: prod
spec:
  zoneId:
    value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: checkout-queue
  host: shop.acme.com
  path: /checkout
  newUsersPerMinute: 200
  totalActiveUsers: 200
```

```shell
planton apply -f waiting-room.yaml
```

This creates a queue on `shop.acme.com/checkout` at Cloudflare's 200/200 floors with fifo admission and the default queue page — no Advanced fields, so it works on any plan. An Infra Job tracks the provisioning in real time.

### InfraChart

When the zone is deployed in the same InfraPipeline, wire `zoneId` with ValueFromRef:

```yaml
spec:
  zoneId:
    valueFrom:
      kind: CloudflareDnsZone
      name: acme-com
      fieldPath: status.outputs.zone_id
  name: checkout-queue
  host: shop.acme.com
  newUsersPerMinute: 200
  totalActiveUsers: 200
```

The InfraPipeline resolves the dependency graph, deploys the zone first, then provisions the waiting room with the resolved zone ID.

## Key Configuration

These are the most important decisions when configuring a waiting room. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Thresholds have a floor of 200** — Cloudflare will not create a room with `newUsersPerMinute` or `totalActiveUsers` below 200, and the spec validates that before apply. Size them for what the origin actually survives: `totalActiveUsers` is the concurrency cap on the route, `newUsersPerMinute` is the admission rate once slots free up. A scheduled launch window that temporarily overrides these rates is a separate CloudflareWaitingRoomEvent, not a field of the room.

**Bypass rules are the whole table** — Cloudflare stores waiting-room rules as a per-room list replaced whole on every update, and this kind folds that list into the room. Every apply sets the room's rule set to exactly `bypassRules`; rules added in the dashboard disappear on the next apply. To stop bypassing without losing a rule's place, set its `enabled: false` instead of deleting it.

**Destroy semantics** — destroy clears the rules list (wiping rules added outside the manifest too) and deletes the room; any CloudflareWaitingRoomEvent referencing its `waiting_room_id` starts failing. When you want the room's ID to survive, pause instead: `suspended: true` stops queueing while keeping the configuration, and `queueAll: true` is the "close the shop" switch that queues everyone for a maintenance window.

**Advanced fields fail at apply, not at validation** — the entitlement wall is Cloudflare's. On a plan without the Waiting Rooms Advanced add-on, keep to the defaults: `queueingMethod: fifo`, no `additionalRoutes`, no `customPageHtml`, Turnstile untouched.

**Queueing method** — `fifo` admits in strict arrival order; `random` draws each admission cycle and is fairer under reload storms; `passthrough` queues nobody (useful to test the wiring); `reject` turns overflow away instead of queueing. Everything except `fifo` needs the Advanced add-on.

**Session duration and renewal** — `sessionDuration` (1–30 minutes, Cloudflare default 5) is how long a visitor's slot survives after they leave the route. By default the session renews on every request; `disableSessionRenewal` (Advanced) expires it from entry instead, forcing periodic re-queueing under sustained load.

**Cookie collisions across rooms** — a visitor's place is tracked by the `__cfwaitingroom` cookie. When one host runs several rooms, set a distinct `cookieSuffix` per room so the cookies do not collide.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareDnsZone** | `zoneId` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `waiting_room_id` | The created room's ID | `waitingRoomId` on CloudflareWaitingRoomEvent |
| `zone_id` | The zone the room belongs to | Zone-scoped siblings deployed in the same pipeline |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Checkout or ticket-path queue** — a room on the one path that must not melt the origin, at the 200/200 floors, no Advanced fields. Start from the **Basic checkout queue** preset.

**Office bypass** — the same room with `bypassRules` matching the office network (`ip.src in {203.0.113.0/24}`) so internal testers skip the queue. Declare the rules in this manifest — the list is authoritative.

**Launch window** — keep the room's steady-state rates and pair it with a CloudflareWaitingRoomEvent that raises (or hard-caps) them for a scheduled on-sale, created and deleted per launch while the room persists.

## Works With

- [**Cloudflare Waiting Room Event**](/infra-catalog/cloudflare-waiting-room-event) — a scheduled window that temporarily overrides this room's rates, page, or Turnstile settings; wires `waitingRoomId` from this room's output
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) — the zone the room protects; `zoneId` references its `zone_id` output
