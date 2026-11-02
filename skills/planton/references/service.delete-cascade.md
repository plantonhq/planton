---
title: Deleting a Service — The Cleanup Cascade
description: Service deletion is a narrated teardown that is refused up front when a resource is marked prevent-destroy, has deletion protection on, or is still read by something else; otherwise it stops in-flight runs, DESTROYS deployed Infra Components environment by environment (each Promotion Path's stages last to first, then the environments on no path; dependents first), releases keep-on-provider resources, then removes history and the record — with a Keep Resources hand-over arm and a protected-environment refusal. Read when someone wants to retire a service, a deletion refused or failed partway (the retry is re-issuing the same delete), or the record should go while the infrastructure stays.
---

# Deleting a Service — The Cleanup Cascade

Deleting a service deletes it genuinely: the platform stops the service's in-flight runs, DESTROYS its deployed Infra Components environment by environment, then removes the run history and the record itself — narrated live so the person watches the teardown happen. Read this when someone wants to retire a service, when a deletion refused and you need to explain why, when a deletion failed partway, or when someone wants the record gone but the infrastructure kept.

## What delete actually does

`planton service delete <service>` confirms, starts the cascade, and follows the narrated timeline by default (`--no-wait` starts it and returns; `--yes` answers the question ahead, and without a terminal the delete refuses and prints itself with `--yes` added). Before anything starts, the delete asks of every resource it would destroy whether it can go (next section), and refuses the whole delete if one cannot. Then the order is the order a careful operator would use:

1. **Stop in-flight runs** — every queued, running, or approval-parked run of the service is cancelled through the same cancel a person uses, and the cascade waits for each run to really end (bounded at fifteen minutes) — destroying while a deploy still applies would race the two.
2. **Destroy resources, per environment** — the reverse of how releases roll out: the organization's Promotion Paths last to first, each path's stages last to first (production-like environments first), then the environments on no path. Within an environment a resource is destroyed only after everything that depends on it (a consumer is gone before the producer it depends on), and independent resources are destroyed side by side, each through a real destroy job whose real ending the cascade waits for. Each resource's record settles as its own destroy ends.
3. **Remove the records** — the deployment history and the run records, then the service record itself.

At each resource the same decision a stack's teardown makes is taken again, at the moment of destroy:

- **`planton.dev/keep-on-provider: "true"`** — never destroyed. The resource is **released**: the service lets go of it, its record stays, and it keeps running on its provider, findable and adoptable into something else. The mark is an annotation under the resource manifest's `metadata.annotations`, set in the service's configuration for each environment whose copy should stay (prod's overlay keeps prod's database; dev's copy is still destroyed), then deployed before the delete. A kept resource is never destroyed, so the up-front checks that guard a destroy (deletion protection, readers) do not ask about it; the protected-environment refusal still applies while the service has other deployments there.
- **`planton.dev/prevent-destroy: "true"`** — never destroyed (the up-front check already refused the delete; this guards a mark added since).
- **Unmanaged** (`planton.dev/infra-job.disable-on-lifecycle-events: "true"`) — never destroyed; it stays on its provider and its record goes with the service.
- **Owned Elsewhere** — a resource another live stack or service now owns, or one an earlier run already let go of (a removal it kept on its provider), is left exactly as it is. A removal a run left in place, or one whose destroy failed, is still the service's and is destroyed here like the rest.
- **Already gone** — a record that no longer exists counts as done.

The delete API returns as soon as the cascade STARTS. The record shows `spec.is_delete_in_progress: true` while it runs; when the service answers NOT_FOUND, the deletion finished. The `delete_service` agent tool works the same way — it requires `confirm: true`, and completion is observed by polling `get_service` until NOT_FOUND. A lost report from a destroy job never leaves a deletion hanging: the platform re-sends any ending the cascade missed.

## The refusals worth explaining well

Each names the resource, its environment, and the way out. Relay it as written:

- **A resource is marked prevent-destroy**: "checkout-api can't be deleted: orders-db in prod is marked prevent-destroy. Remove the mark, or delete with Keep Resources."
- **A resource's own deletion protection is on** (the switch its kind names, like `deletion_protection`): the provider would refuse the destroy halfway through, so the delete refuses first. Turn the switch off through the service's configuration and deploy, or delete with Keep Resources.
- **Something else still reads a resource** — another stack, another service's resource, or a resource deployed on its own (the service's own resources never count as readers of each other). Destroying it would break that reader; the refusal names the reader. Remove the reference there first, or delete with Keep Resources.
- **A kept resource depends on one the delete would destroy**: a resource that stays on its provider keeps what it references or lives inside. Mark those too, or remove the mark.
- **"has deployments in protected environment X"**: protection exists to put ceremony in front of destructive acts, and deleting a service destroys its production resources. The working paths are in the refusal: remove the environment's protection first (whoever holds that right performs the deliberate act), or delete with Keep Resources.
- **"service is being deleted"**: a cascade is already running (or failed and awaits a retry). Wait for it, or re-issue the delete to retry a failed one.

A refusal costs nothing: no run was stopped and nothing was destroyed.

## While a deletion runs, the service takes no new work

Every door refuses with the deletion named: applies and manifest pushes, manual runs and reruns, deploys, promotions, rollbacks, redeploys, and webhook pushes (which simply skip the service). This is not a lock to work around — a write landing mid-cascade would be destroyed moments later. The refusal ends when the deletion completes (record gone) or is answered by the retry path below when it failed.

## When a deletion fails partway

A destroy that errors stops the cascade LOUDLY: the narrated timeline shows exactly which resource in which environment failed and names the destroy job carrying the provider's own error. The service stays delete-in-progress — visibly mid-deletion, refusing new work — and nothing is silently forgotten. The retry is simple and safe: **re-issue the same delete**. Already-destroyed resources skip, already-deleted records count as done, and the cascade picks up where the failure left it. Explain it exactly that way: fix the provider-side cause (a dependency outside the platform), then delete again.

## Keeping the infrastructure: Keep Resources

`planton service delete <service> --retain-infra-components` (Keep Resources in the console) removes the service from the platform — history and record — while deliberately leaving every deployed Infra Component running. Each resource is released, exactly as Keep does on a stack: the service lets go, the resource's record stays, and it can be adopted by a stack or managed on its own. This is the hand-over case (another team, another tool, another stack takes ownership), and the narration says so plainly. Because nothing is destroyed, nothing that guards a destroy stands in its way: it passes the prevent-destroy, deletion-protection, reader and protected-environment refusals.

## Watching a deletion

The CLI follows by default: a phase timeline (stop runs → each environment → history → record) with per-resource sub-steps and a terminal line that tells success from failure. A deletion someone started elsewhere can be followed while it runs through the service's cleanup progress stream; once the record is gone, the stream answers NOT_FOUND — which IS the success answer, not an error.
