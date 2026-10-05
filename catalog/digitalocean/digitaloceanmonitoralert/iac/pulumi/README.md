# Pulumi Module: DigitalOcean Monitor Alert

Provisions an alert policy on DigitalOcean's built-in metrics -- the complete `digitalocean_monitor_alert` resource surface, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the IaC input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/monitor_alert.go` -- the `MonitorAlert` resource and output exports
- `module/locals.go` -- target handle (the policy's tags are alert targeting, not resource labels, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanMonitorAlertOutputs` contract)

## Behavior notes

- The SDK flattens the provider's one-element `alerts` block to a single object and pluralizes the channel lists (`Emails`/`Slacks`).
- The Slack webhook URL is not secret-flagged by the SDK, so the module wraps it in `pulumi.ToSecret` -- a credential never ships as plaintext state.
- The three typed reference lists merge into the single `Entities` array; spec validation guarantees only the metric family's own list is populated.
- `alert_id` is exported from the resource id (the provider's `Uuid` output is never populated at the pin).
