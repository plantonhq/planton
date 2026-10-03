# CloudflareWaitingRoomEvent Pulumi Module

Pulumi (Go) IaC module for a scheduled waiting-room event -- a time window whose optional overrides temporarily replace the room's rates, HTML, and Turnstile.

## Architecture

```
main.go                       — Entrypoint loading the IaC input
module/main.go                — Resources(): provider setup, resource, outputs
module/locals.go              — Locals initialization
module/waiting_room_event.go  — cloudflare.WaitingRoomEvent
module/outputs.go             — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: RFC3339 times, duration walls, shuffle-needs-prequeue, users pair both-or-neither, `event_id` / `waiting_room_id` / `zone_id` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `event_id` | The created event's ID |
| `waiting_room_id` | The room the event runs on |
| `zone_id` | The zone the room belongs to |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
