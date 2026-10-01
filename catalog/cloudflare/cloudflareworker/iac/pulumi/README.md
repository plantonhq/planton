# CloudflareWorker Pulumi module

Provisions a Cloudflare Worker script and the companions that hang off it: workers.dev subdomain, custom domains, routes, and cron triggers.

## What this module creates

- `cloudflare.WorkersScript` — script, bindings (flattened from typed lists), assets, migrations, observability, placement, limits
- `cloudflare.WorkersScriptSubdomain` — when `workersDev.enabled`
- `cloudflare.WorkersCustomDomain` — one per hostname (`environment` is deprecated and omitted)
- `cloudflare.WorkersRoute` — one per pattern
- `cloudflare.WorkersCronTrigger` — when `schedules` is set
- R2 fetch via the AWS S3 provider — only when `r2Bundle` is set. The provider signs with the connection's R2 key pair (`provider_config.r2`, falling back to the ambient `AWS_*` chain when absent) and reads the object's raw bytes whatever its Content-Type; an empty bundle stops the program with an error naming it

## Outputs

`script_id`, `script_name`, `custom_domain_hostnames`, `route_patterns`, plus keyed maps `custom_domain_ids` (by hostname), `route_ids` and `route_zone_ids` (by list index) so import can reassemble `{account_id}/{domain_id}` and `{zone_id}/{route_id}`.

## PARITY-EXCEPTION

Pulumi's Cloudflare SDK v6.17.0 has no inputs for `cacheOptions`, `exports`, or `packageDependencies`. Tofu honors them; this engine logs a warning and skips them. Rate-limit `mitigationTimeout` and traces `propagationPolicy` have the same gap. The proto keeps the fields (never `reserved`); the next SDK bump dissolves the exception.

## Usage

```bash
planton apply -f manifest.yaml --iac pulumi
```
