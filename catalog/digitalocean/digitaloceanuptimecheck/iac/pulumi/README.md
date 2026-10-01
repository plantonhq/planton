# Pulumi Module: DigitalOcean Uptime Check

Provisions an uptime probe on an external endpoint plus its alert rules -- the complete `digitalocean_uptime_check` surface with one `digitalocean_uptime_alert` per spec alert row, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the stack input and calls the module
- `module/main.go` -- orchestration: locals, provider, resources
- `module/uptime_check.go` -- the `UptimeCheck` resource, the per-row `UptimeAlert` children, and output exports
- `module/locals.go` -- target handle (a check has no tag surface, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanUptimeCheckStackOutputs` contract)

## Behavior notes

- Alert rows parent the check (`pulumi.Parent`) and take its id as `CheckId` -- the upstream mutable-parent corruption class is unrepresentable here.
- The SDK keeps the provider's unbounded notifications list, but the provider reads only the first element -- exactly one is sent per row.
- The Slack webhook URL is not secret-flagged by the SDK, so the module wraps it in `pulumi.ToSecret`.
- `period` is always sent (spec-required): DigitalOcean rejects any alert without one, whatever the SDK's schema says.
- `threshold`/`comparison` are sent as DigitalOcean will store them: authored for latency; the authored threshold plus `less_than` for ssl_expiry; the API's fixed `1` / `less_than` for down and down_global. The spec forbids authoring the fixed values, and sending them is the only shape that reads back unchanged.
- Each alert row's resource name is `<idx>-<alert_name>` -- the same key the Terraform module uses as its `for_each` key and both engines use for the `alert_ids` output, so a blind import can find every row's id from state alone. The index lets two rows share a display name without colliding.
