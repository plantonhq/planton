# CloudflareWaitingRoom Terraform Module

Terraform IaC module for a waiting room -- a virtual queue on a host+path -- plus the room's folded bypass-rules list.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareWaitingRoomSpec
main.tf       — cloudflare_waiting_room + cloudflare_waiting_room_rules
outputs.tf    — waiting_room_id, zone_id
```

## Behavior

`new_users_per_minute` and `total_active_users` have a floor of 200. Advanced-subscription fields fail at the API on plans without the add-on. `bypass_rules` is the room's entire rule set -- the module supplies `action = bypass_waiting_room`. Destroy PUTs an empty rules list and then deletes the room.

## Outputs

| Name | Description |
|------|-------------|
| `waiting_room_id` | The created room's ID -- what events and the import recipe reference |
| `zone_id` | The zone the waiting room belongs to |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
