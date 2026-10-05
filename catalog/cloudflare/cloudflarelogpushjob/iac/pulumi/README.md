# CloudflareLogpushJob Pulumi Module

Pulumi (Go) IaC module for Logpush jobs.

## Architecture

```
main.go                 — iac-input loading + module entry
module/main.go          — provider setup + resource orchestration
module/locals.go        — metadata/credential references
module/logpush_job.go   — LogpushJob + conditional LogpushOwnershipChallenge
module/outputs.go       — job_id, scope ids, ownership-challenge trio
```

## Behavior

A plain CRUD resource: real create/update/delete. Dual scope -- exactly one of `account_id` or `zone_id` is set (spec validation enforces it). `dataset` and the scope are immutable (the provider replaces the job); `kind` is immutable at the API without a plan guard (an in-place update 400s at apply). `destination_conf` updates in place.

`enabled` defaults to TRUE here even though Cloudflare's own default is FALSE: a declared log job is meant to ship logs.

The ownership-challenge resource is created only when `generate_ownership_challenge` is set. It is one-shot at Cloudflare -- no read, no update, no delete, no import -- so destroying it only drops it from state.

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

## SDK Version

Uses `pulumi-cloudflare` v6.19.0.
