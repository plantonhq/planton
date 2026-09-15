# DigitalOcean Monitor Alert

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_monitor_alert` resource at the pinned provider version.

## What this component models

An alert policy on DigitalOcean's built-in metrics for Droplets, load balancers, and managed database clusters, with email and Slack notification channels. DigitalOcean's API targets a policy through one untyped id list plus a tag list; this spec replaces the untyped list with one TYPED reference list per resource family, so an id can never be paired with the wrong metric family and resources are wired by reference.

The component covers the provider's full argument surface:

- `description` -- the policy's display handle (DigitalOcean has no separate name)
- `metric_type` -- one of the 28 metric paths across three families: 12 `v1/insights/droplet/*`, 12 `v1/insights/lbaas/*`, 4 `v1/dbaas/alerts/*`
- `compare` -- `GreaterThan` or `LessThan` (this API's CamelCase spelling)
- `value` -- the threshold (stored as a 32-bit float upstream; more than 7 significant digits truncate)
- `window` -- `5m`, `10m`, `30m`, or `1h`
- `enabled` -- optional; unset defers to DigitalOcean's default (enabled)
- `droplet_ids` / `load_balancer_ids` / `database_cluster_ids` -- typed reference lists, validated against the metric family
- `tags` -- tag-targeted droplet alerts (droplet metrics only)
- `alerts` -- the notification channels: `emails` and/or `slack` rows (`channel` + secret `url`); at least one channel required

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanMonitorAlert
metadata:
  name: web-cpu-alert
spec:
  description: CPU of web droplets above 90 percent
  metricType: v1/insights/droplet/cpu
  compare: GreaterThan
  value: 90
  window: 5m
  tags:
    - web
  alerts:
    emails:
      - ops@example.com
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `alert_id` | UUID of the alert policy (the API identity, and the import id) |

## Behavior worth knowing

- **Metric families gate the targets.** Droplet metrics accept droplet references and/or tags; load-balancer and database metrics accept only their own reference lists -- validation enforces the pairing before DigitalOcean ever sees the request.
- **Everything updates in place.** No field on this resource forces a replacement.
- **The provider's `uuid` attribute is dead at the pin** (declared but never populated); the `alert_id` output carries the policy UUID from the resource id.
- **Slack webhook URLs are credentials.** DigitalOcean does not mark them sensitive; this spec does -- the platform accepts only `$secret/<name>` references, and the Pulumi module encrypts the value in stack state (Terraform state stores every value in plain text; protect the backend).
- **Alert email goes only to verified team members.** DigitalOcean rejects any other address at create time (`email is not verified`); invite and verify the recipient first.
- **A targeted tag need not exist.** The policy stores tags as selectors and neither checks nor creates them -- declare the alert before the fleet if you like. Firewalls behave the opposite way.
- **Metric names are DigitalOcean's own API paths**, inconsistencies included: droplet CPU is bare `cpu`, and the database family lives under `v1/dbaas/alerts/` with `_alerts` suffixes. They are never "corrected" here.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
