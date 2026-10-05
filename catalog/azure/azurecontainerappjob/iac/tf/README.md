# AzureContainerAppJob - Terraform Module

Terraform implementation for the AzureContainerAppJob
kind.

## Resources Created

- `azurerm_container_app_job.main` -- the finite-run workload on a
  Container Apps environment, with its full template (containers, init
  containers, volumes), trigger, secret, registry, and identity surface
  realized as dynamic blocks

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.container_app_environment_id` | The hosting environment (ForceNew) |
| `spec.replica_timeout_in_seconds` | The per-execution budget (required) |
| `spec.containers` | At least one; probes, env, and volume mounts nest inside |
| Triggers | Exactly ONE of `manual_trigger` / `schedule_trigger` / `event_trigger` (spec-enforced); switching trigger types is ForceNew |
| `spec.secrets` / `spec.registries` / `spec.identity` | Registry credentials can reference secrets or managed identity |
| `spec.workload_profile_name` | Absent means the Consumption profile |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- Probe defaults are materialized explicitly (liveness initial delay 1s,
  readiness/startup 0s; success threshold is readiness-only) because the
  platform never sends proto defaults -- both engines emit identical
  bodies.
- A volume without a storage type deploys `EmptyDir`.
- Trigger parallelism/completion defaults are applied in the variables
  layer, same reason.
- User tags merge over metadata-derived tags (user wins).

## Usage

```hcl
module "nightly_job" {
  source = "./path/to/module"

  metadata = { name = "nightly-report" }
  spec = {
    region                       = "eastus"
    resource_group               = "apps-rg"
    job_name                     = "nightly-report"
    container_app_environment_id = "/subscriptions/.../managedEnvironments/apps-env"
    replica_timeout_in_seconds   = 1800
    containers = [{
      name  = "report"
      image = "ghcr.io/acme/report:1.4.2"
    }]
    schedule_trigger = { cron_expression = "0 2 * * *" }
  }
}
```
