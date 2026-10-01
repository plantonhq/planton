# Terraform Module: DigitalOcean Uptime Check

Provisions an uptime probe on an external endpoint plus its alert rules -- the complete `digitalocean_uptime_check` surface with one `digitalocean_uptime_alert` per spec alert row.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_uptime_check.check` | The probe: target, protocol, vantage regions |
| `digitalocean_uptime_alert.alerts` | One per spec alert row (`for_each` keyed `"<index>-<alert_name>"`) |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanUptimeCheckSpec` proto: `check_name`, `target`, optional `type`, required `regions`, optional `enabled`, and the `alerts` rows (each with `alert_name`, `type`, required `period`, per-type `threshold`/`comparison`, and `notifications` channels). Authentication uses `digitalocean_token` (falls back to the provider's `DIGITALOCEAN_TOKEN` when null).

## Outputs

Exactly the `DigitalOceanUptimeCheckStackOutputs` contract: `check_id` and `alert_ids` (each row's UUID keyed by the `for_each` key `<index>-<alert_name>` -- the second half of the row's `{check_id},{alert_id}` import id, so a blind import derives it from state).

## Behavior notes

- `check_id` on each alert row is wired to the check this module creates -- the upstream mutable-parent corruption class is unrepresentable here.
- `regions` is always sent (spec-required): the provider never reconciles a DigitalOcean-defaulted region set, so omission would leave a perpetual removal diff.
- `period` is always sent (spec-required): DigitalOcean rejects any alert without one, whatever the provider's schema says.
- `threshold`/`comparison` are sent as DigitalOcean will store them (`locals.tf`): authored for latency; the authored threshold plus `less_than` for ssl_expiry; the API's fixed `1` / `less_than` for down and down_global. The spec forbids authoring the fixed values, and sending them is the only shape that reads back unchanged -- an omitted value reads back as `1` / `less_than` against a null and re-plans forever.
- The `for_each` key `<index>-<alert_name>` is also the Pulumi module's resource name and the `alert_ids` output key -- one key, three roles, so both engines address and export every row identically.
- Import: check by `<check_id>`, alert rows by `<check_id>,<alert_id>` (see `iac/import-map.yaml`).
