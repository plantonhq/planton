# Terraform Module: DigitalOcean Monitor Alert

Provisions an alert policy on DigitalOcean's built-in metrics -- the complete `digitalocean_monitor_alert` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_monitor_alert.alert` | The policy: metric, comparison, threshold, window, targets, notification channels |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanMonitorAlertSpec` proto: `description`, `metric_type`, `compare`, `value`, `window`, optional `enabled`, the three typed reference lists (`droplet_ids`, `load_balancer_ids`, `database_cluster_ids` -- resolved references arrive as literal id strings), `tags`, and the `alerts` channels. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanMonitorAlertOutputs` contract: `alert_id` (the policy UUID from the resource id -- the provider's own `uuid` attribute is never populated at the pin).

## Behavior notes

- The three typed lists merge into the provider's single `entities` argument (`locals.tf`); spec validation guarantees only the metric family's own list is populated.
- `enabled` left null defers to the provider's default (enabled); an explicit false is transmitted (the provider sends a pointer).
- The policy's tags SELECT tagged droplets as targets -- they are alert targeting, not resource labels, so no Planton labels are merged in.
- Import: `terraform import ... <alert_id>` (see `iac/import-map.yaml`).
