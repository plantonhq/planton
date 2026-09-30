# CloudflareLogpushJob Terraform Module

Terraform IaC module for Logpush jobs.

## Architecture

```
provider.tf   — Cloudflare provider configuration (~> 5.23)
variables.tf  — Input variables mirroring CloudflareLogpushJobSpec (generated)
main.tf       — cloudflare_logpush_job + conditional cloudflare_logpush_ownership_challenge
outputs.tf    — job_id, scope ids, ownership-challenge trio
```

## Behavior

A plain CRUD resource: real create/update/delete. Dual scope -- exactly one of `account_id` or `zone_id` is set (spec validation enforces it). `dataset` and the scope are immutable (the provider replaces the job); `kind` is immutable at the API without a plan guard (an in-place update 400s at apply). `destination_conf` updates in place.

`enabled` defaults to TRUE here even though Cloudflare's own default is FALSE: a declared log job is meant to ship logs.

The ownership-challenge companion is deployed only when `generate_ownership_challenge` is set. It is one-shot at Cloudflare -- no read, no update, no delete, no import -- so destroying it only drops it from state.

Import the job as `{accounts|zones}/{scope_id}/{job_id}`.

## Outputs

| Name | Description |
|------|-------------|
| `job_id` | The Cloudflare-assigned numeric job id, in string form |
| `account_id` | The account scope, when account-scoped |
| `zone_id` | The zone scope, when zone-scoped |
| `ownership_challenge_filename` | Challenge file name (challenge arm only) |
| `ownership_challenge_message` | Message accompanying the challenge |
| `ownership_challenge_valid` | Whether Cloudflare found the destination valid |

## Provider Version

Uses `cloudflare/cloudflare ~> 5.23`.
