# The self-hosted platform's documents teach the vault's whole disaster story

**Date**: September 16, 2026
**Type**: Documentation
**Components**: The platform kind's guide (`catalog/kubernetes/kubernetesplantonplatform/GUIDE.md`), the disaster-recovery pattern (`catalog/_patterns/stateful-kind-disaster-recovery.md`) and the patterns index, the multi-cloud-catalog skill's research recipes and eval bank, the public self-hosting pages

## Summary

The operator stores the bundled vault in the platform's database, seals it with the adopter's cloud key or a Secret the adopter owns, and brings it back with the records on a restore. The documents an adopter or a coding agent reads now say so end to end, in the shapes the catalog's other stateful kinds already teach. The platform kind's guide gains three sections in the standalone vault guide's own heading forms: the GKE disaster-recovery resource set (the seal identity, ring, key with `PREVENT`, both grants, the vault's Workload Identity binding, and the platform, embedded whole and validated by the catalog's guardrails, with the store set linked to the pattern), the Cloudflare R2 shape by preset slug with what the keys Secret holds and how to keep it, and the bad-day runbook per seal written from the operator's own Kind lanes -- the steps, the sentences the status shows at each, the refusal a person meets when they skip the first step, what a restore can never bring back, and the embedded bad-day declaration. The disaster-recovery pattern gains the bundled platform as its fourth subject: one archive for records and secrets, the seal or the Secret as the only continuity, a row in the choices table, the platform's step in the restore list. The skill's research recipe gains the self-hosted question and its checklist; the eval bank gains three questions that fail CI if any of it rots. The public backup page is rewritten whole -- the vault is in the backup, the keys are the exception, how to protect them, the restore's step per seal -- and no longer tells a team to re-enter their cloud credentials after a restore.

Proven the way the catalog proves its teaching: a fresh agent given only the built skill pack and the adopter's question composed the whole GKE set with every document cited, wrote both runbooks, and listed fifty cited hazards; its four real misses (no embedded restore declaration, whether the restored platform keeps the source's version and hostname, what cert-manager's absence does) were fixed in the guide before this change landed, and the eval questions were written after the run so they guard what the agent relied on.

## What Changed

### The platform kind's guide

`## Disaster recovery on GKE: the resource set` -- the numbered table (two identities on purpose: the seal key and the archive bucket are different blast radii; the ring's permanence; the key's `PREVENT`; both KMS roles and why the second exists; the bucket's two storage roles; the two bindings and the two ServiceAccount names, `<platform>-openbao` and `<platform>-postgres`; the platform's `gcs` backup arm and `gcpKms` seal arm), the embedded set of seven validated documents, the apply order and why the platform goes last, and the honesty that the database's backup identity is a literal email while the vault's seal identity is by reference. `## Disaster recovery with Cloudflare R2: the resource set` -- preset `06-backups-to-r2` by slug, what the keys Secret holds, the copy taken the day the platform is born. `## Restore on the bad day` -- the same version and hostname as the source, cert-manager before the platform, the cloud-seal steps (grants and binding for the new cluster; the vault opens itself; the Secret recreated after, with the hazard sentence quoted), the built-in-seal steps (the Secret before the declaration; the refusal sentence if forgotten, quoted), the embedded bad-day declaration, what a restore cannot bring back, and how to read `status.backup.vault` and `status.backup.restoredFrom`. Spec-field mentions across the guide converged to camelCase per the guide rule.

### The pattern

Frontmatter gains `KubernetesPlantonPlatform`; the introduction names the fourth subject and its grain; `### The bundled platform: one archive for records and secrets` states the composition, the two answers to "what opens the restored vault", and the two bindings; a row in the choices table; the platform's entry in "The restore is a declaration, and one step stays yours"; two "when not to use" bullets (an in-cluster store is a lab shape for the platform too; never propose a second archive for the vault); the platform guide in "See also". `catalog/_patterns/README.md` names the fourth subject.

### The skill

`skills/multi-cloud-catalog/references/research-recipes.md`: `### Self-hosted Planton: one archive for records and secrets` -- three reads on the platform's page and guide and a seven-line checklist, every line a rule the pack states. `pkg/explain/refgen/eval_questions.yaml`: q20 (the GKE set and its two bindings), q21 (the bare-metal posture and the refusal), q22 (the bad-day runbook per seal and what a restore never brings back), each pinned to the guide, the reference page, and the pattern by durable text.

### The public pages

`site/public/docs/self-hosting/backup-and-recovery.md` rewritten whole: the vault in the backup and keyless connections surviving a restore; the keys as the deliberate exception; `## Protect the vault's keys` with the cloud-key fragment and the Secret fragment and the copy command; the R2 declaration carrying `vault.init_secret_name` (a backup without it is refused); `status.backup.vault` in "Read the archive's state"; the restore with the same `vault` block, the vault's step per seal, and "the seal is decided once"; the false "re-enter your credentials" sentence gone; requirements naming the operator chart that carries the vault in the archive, its version to be named when it ships. `index.md`: the secrets manager stores in the platform's own database so one backup carries records and secrets together.

## Why

A promise about the bad day is worth exactly what the documents say on the bad day. The code has kept this promise since the operator learned to store the vault in the database, seal it from the adopter's key, and sign into it as itself; until now the public page still told a team to re-enter their credentials, the guide had no set an agent could compose from, and the pattern did not know the platform existed. Every document now describes one reality, and a fresh reader proved it teaches the story whole.
