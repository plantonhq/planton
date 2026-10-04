---
title: Deleting a Service — The Cleanup Cascade
description: Service deletion is a narrated teardown that stops in-flight runs, DESTROYS deployed Infra Components environment by environment (reverse promotion order, reverse dependency order), then removes history and the record — with a retain-the-resources hand-over arm and a protected-environment refusal. Read when someone wants to retire a service, a deletion refused or failed partway (the retry is re-issuing the same delete), or the record should go while the infrastructure stays.
---

# Deleting a Service — The Cleanup Cascade

Deleting a service deletes it genuinely: the platform stops the service's in-flight runs, DESTROYS its deployed Infra Components environment by environment, then removes the run history and the record itself — narrated live so the person watches the teardown happen. Read this when someone wants to retire a service, when a deletion refused and you need to explain why, when a deletion failed partway, or when someone wants the record gone but the infrastructure kept.

## What delete actually does

`planton service delete <service>` confirms, starts the cascade, and follows the narrated timeline by default (`--no-wait` starts it and returns; `--force` skips the confirmation). The order is the order a careful operator would use:

1. **Stop in-flight runs** — every queued, running, or approval-parked run of the service is cancelled through the same stage-aware cancel a person uses, and the cascade waits until none remain.
2. **Destroy resources, per environment** — in REVERSE promotion order (production-like environments first, the reverse of how releases roll out), each environment's resources are destroyed in the reverse of their dependency order (a consumer is gone before the producer it depends on), through real destroy jobs whose completion the cascade awaits one by one.
3. **Remove the records** — the destroyed resources' records, then the deployment history and the run records, then the service record itself.

The delete API returns as soon as the cascade STARTS. The record shows `spec.is_delete_in_progress: true` while it runs; when the service answers NOT_FOUND, the deletion finished. The `delete_service` agent tool works the same way — it requires `confirm: true`, and completion is observed by polling `get_service` until NOT_FOUND.

## While a deletion runs, the service takes no new work

Every door refuses with the deletion named: applies and manifest pushes, manual runs and reruns, deploys, promotions, rollbacks, and webhook pushes (which simply skip the service). This is not a lock to work around — a write landing mid-cascade would be destroyed moments later. The refusal ends when the deletion completes (record gone) or is answered by the retry path below when it failed.

## When a deletion fails partway

A destroy that errors stops the cascade LOUDLY: the narrated timeline shows exactly which resource in which environment failed and names the destroy job carrying the provider's own error. The service stays delete-in-progress — visibly mid-deletion, refusing new work — and nothing is silently forgotten. The retry is simple and safe: **re-issue the same delete**. Already-destroyed resources skip, already-deleted records count as done, and the cascade picks up where the failure left it. Explain it exactly that way: fix the provider-side cause (a deletion protection, a dependency outside the platform), then delete again.

## The two refusals worth explaining well

- **"has deployments in protected environment X"**: protection exists to put ceremony in front of destructive acts, and deleting a service destroys its production resources. The working paths are in the refusal: remove the environment's protection first (whoever holds that right performs the deliberate act), or delete with the resources retained.
- **"service is being deleted"**: a cascade is already running (or failed and awaits a retry). Wait for it, or re-issue the delete to retry a failed one.

## Keeping the infrastructure: the hand-over arm

`planton service delete <service> --retain-infra-components` removes the service from the platform — history and record — while deliberately leaving every deployed Infra Component running. From that moment the resources are unmanaged: no record, no rollback, no verification. This is the hand-over case (another team, another tool, another platform takes ownership), and the narration says so plainly. It also passes the protection refusal, because nothing is destroyed.

## Watching a deletion

The CLI follows by default: a phase timeline (stop runs → each environment → history → record) with per-resource sub-steps and a terminal line that tells success from failure. A deletion someone started elsewhere can be followed while it runs through the service's cleanup progress stream; once the record is gone, the stream answers NOT_FOUND — which IS the success answer, not an error.
