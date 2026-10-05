# CloudflareWaitingRoom Pulumi Module

Pulumi (Go) IaC module for a waiting room -- a virtual queue on a host+path -- plus the room's folded bypass-rules list.

## Architecture

```
main.go                    — Entrypoint loading the IaC input
module/main.go             — Resources(): provider setup, resource, outputs
module/locals.go           — Locals initialization
module/waiting_room.go     — cloudflare.WaitingRoom + cloudflare.WaitingRoomRules
module/outputs.go          — output keys
```

## Behavior

Mirrors the Terraform module's contract exactly: 200 floors, Advanced fields fail at the API, folded bypass rules with fixed `bypass_waiting_room` action, `waiting_room_id` / `zone_id` outputs.

## Outputs

| Name | Description |
|------|-------------|
| `waiting_room_id` | The created room's ID -- what events and the import recipe reference |
| `zone_id` | The zone the waiting room belongs to |

## Provider Version

Uses `pulumi-cloudflare` SDK v6 (Cloudflare Terraform provider v5 line).
