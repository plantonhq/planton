# The OpenBao guide teaches which storage engine to choose and what each means on the bad day

**Date**: September 15, 2026
**Type**: Documentation
**Components**: `catalog/kubernetes/kubernetesopenbao/GUIDE.md`, `catalog/_patterns/stateful-kind-disaster-recovery.md`, `skills/multi-cloud-catalog/references/research-recipes.md`, `pkg/explain/refgen/eval_questions.yaml`, `pkg/explain/refgen/commons.md` (and the regenerated `catalog/_docs/reference-commons.md`), `catalog/_patterns/README.md`

## Summary

`KubernetesOpenBao` declares a storage engine -- integrated Raft, or PostgreSQL by reference to a `KubernetesPostgres` -- and the reference page says what each engine is, but nothing in the pack an installed agent reads said which to choose or what each means for backups and for the day the cluster is lost. The guide now carries that judgment in a section placed before "Backups and restore", so the Raft-only backup rule lands as a consequence of a choice the reader has already made: Raft when the vault should own its own recovery (snapshots to a store, the restore runbook, quorum arithmetic), PostgreSQL when a team already runs and backs up a database and wants one backup to cover the vault (its disaster recovery becomes the database's; the seal key is still required on restore because the barrier key wraps the data on either engine; availability without quorum; a database outage is a vault outage; `maxParallel` against a shared database's headroom), dev never for real secrets. It names what breaks when the choice is wrong, the database side of the PostgreSQL choice (its own database, declared at bootstrap or created once by hand on a running cluster; the loud failure when it is missing; the tables OpenBao creates itself; every pod unsealed on either engine), the bad day on PostgreSQL storage (the database's restore, then an unseal, never a re-init), and the diagram consequence (PostgreSQL draws edges to the database node and its credential Secret; Raft draws a volume).

The disaster-recovery pattern's "Choices and consequences" table gains the engine row, linking the guide. The skill's stateful-kind recipe widens its guide-heading grep so an installed agent finds the section, and its restore checklist gains the precondition it lacked: the source must be a Raft vault. Eval q19 pins what a reader relies on across the guide, the reference page, and the pattern.

## Proven with a reader who never saw the work

The pack was built from the tree and handed to a fresh agent restricted to it, with a customer's question (a team on a nightly-backed-up `KubernetesPostgres` adding OpenBao: which engine, what each means for backups and the bad day, the manifest, the hand steps). It chose PostgreSQL storage for the guide's own reason, surfaced the guide's own condition back to the customer, cited every consequence, composed a valid manifest with every kept default explained, and listed what the pack did not tell it. Six of those gaps were the guide's to close and were closed in the same change (the database-side steps, the missing-database failure, the tables, the per-pod unseal, the PostgreSQL bad day, the no-share rule); two stale `patterns/` pointers the reader tripped on (`_patterns/README.md` and the generated commons page's template) now say `_patterns/`.

## Verification

`go test ./pkg/explain/refgen/` (guide placement, embedded manifests, links, the eval bank including q19, the reference drift gate after `make generate-reference`); `go test ./pkg/skills/defspack/` and `go run ./pkg/skills/defspack` (the skills lint gate's pair; the pack assembles with the new section in it); the guide's pack-reachability and banned-word greps clean.
