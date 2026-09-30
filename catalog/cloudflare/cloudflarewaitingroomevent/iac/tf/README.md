# CloudflareWaitingRoomEvent Terraform Module

Terraform IaC module for a scheduled waiting-room event -- a time window whose optional overrides temporarily replace the room's rates, HTML, and Turnstile.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareWaitingRoomEventSpec
main.tf       — cloudflare_waiting_room_event
outputs.tf    — event_id, waiting_room_id, zone_id
```

## Behavior

Times are RFC3339 strings. Start must be at least one minute before end; a prequeue must be at least five minutes before start. Shuffle requires a prequeue. The two user-rate overrides must be set together or not at all. Unset overrides inherit the room. Destroy is a real delete of the event; the room stays.

## Outputs

| Name | Description |
|------|-------------|
| `event_id` | The created event's ID |
| `waiting_room_id` | The room the event runs on |
| `zone_id` | The zone the room belongs to |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
