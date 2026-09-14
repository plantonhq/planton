# The OpenBao disaster-recovery story reaches coding agents whole: four module truths become spec facts

**Date**: September 14, 2026
**Type**: Documentation
**Components**: KubernetesOpenBao (spec comments, reference page, guide), the stateful-kind disaster-recovery pattern, the multi-cloud-catalog skill (research recipes, pack layout, contributing guide), component update rule

## Summary

An agent reading the packaged multi-cloud-catalog skill can now answer "back up our OpenBao and tell me what can go wrong" without opening module code. Four truths that only the Pulumi and Terraform modules knew — the name budget once `backup` is declared, the pod state a restore Job shows while it waits for the operator's token, what happens when a snapshot install fails part-way and how to recover, and the DNS name the jobs need in a TLS certificate — are now facts on the spec fields they belong to, rendered into the reference page. The guide gains the day-2 judgment around them, the pattern gains what "restore again" means per kind, and the skill's references stop pointing agents at files the pack does not carry.

## What Changed

### Facts on the fields

The spec's header states the three name budgets (54 characters; 44 with the injector; 45 with `backup`, because Kubernetes caps CronJob names at 52 and the backup CronJob is `<name>-backup`) and that a longer name fails the deploy before anything is created. `restore.rootToken` says the restore pod shows `CreateContainerConfigError` until the token Secret exists, and that this is the designed wait. `restore` says a finished Job must never be deleted by hand, and that OpenBao seals itself when an install fails part-way — with the recovery loop the Job's log prints. `tls.certSecretName` says a certificate serving a vault with `backup` declared must include the active-leader Service, `<name>-active.<namespace>.svc`, or every job run fails with an x509 error naming it. The reference page and the proto docs are regenerated; no field, rule, or default moved.

### Judgment in the guide

A "Day-2 operations" section: confirm a snapshot from the run's own `Uploaded …` line rather than a listing made with your own credentials; rehearse a restore beside a live source with `snapshotKey` and give the clone its own prefix in the same apply that removes `restore` (otherwise the clone's schedule resumes into the source's prefix and prunes it); "restore again" is a changed declaration, never a deleted Job. One honest paragraph states which arms are proven live (S3 with keys and the transit restore on kind; keyless GCS and R2 with a KMS restore on GKE) and which render and validate but have no lane yet (Azure Blob, keyless S3 on IRSA).

### The pattern, per kind

The stateful-kind disaster-recovery pattern now says that a waiting restore is a designed state each kind's field doc names, and that "restore again" differs: OpenBao's Job and MongoDB's Restore object are hash-named and a deleted one is recreated and restores again over live data, while PostgreSQL's recovery is the cluster's bootstrap and runs once.

### The skill knows what it ships

`pack-layout.md` gains "What the pack does not carry" (presets, README, catalog page, IaC modules, E2E trees) and where those live instead; `contributing-wisdom.md` routes operational truths to the proto comment and spells the pattern folder as `_patterns/`; the "Can I back this up, and get it back?" recipe names OpenBao, points at the guide and the pattern's embedded manifests instead of `presets/`, and carries a six-line checklist an agent runs before writing a `restore`.

### The update rule

A new teaching in the Update IaC list: an operational truth the module enforces or prints is a spec fact and lives in the proto comment of its field, because the packaged skill ships the reference page, the guide, and the patterns — never the module's comments or the preset explainers.

## Verification

Spec tests (80/80), `go test ./pkg/explain/refgen/ ./pkg/protodocs/` (no drift after regeneration), `go run ./pkg/skills/defspack` (validated), `buf lint` and `buf format` on the spec; the generated stub diff is comment-only.
