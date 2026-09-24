---
title: Authoring Configuration as a Git-Owned Kustomize Tree
description: Moving a service's per-environment configuration between the record and its repository — eject (the tree becomes the writer), checkout (the record materialized into files, authorship unchanged), init (a fresh skeleton), and patch-schema (the merge schema that keeps name-keyed lists merging). Read when someone wants their service config in git, wants files from what the platform has, asks what the _kustomize layout means, or wants authorship back.
---

# Authoring Configuration as a Git-Owned Kustomize Tree

Every service's per-environment configuration lives in ONE place: the record's `deploy.environments`. What varies is who WRITES it. A manually-declared service is written by whoever applies it (console, agent, `service.yaml`). A git-maintained service declares `deploy.kustomize`, and from then on the platform's own build lane writes the record from the repository's kustomize tree — a push on a branch that drives an environment syncs exactly that environment's entries, provenance stamped. These verbs move between the two worlds.

## The four verbs

- **`planton service kustomize eject <service>`** — hand authorship to the repository. Writes the record's declared configuration into a tree (one overlay per environment, one file per resource, the merge schema included), PROVES the tree renders back identical to the record, then — after an explicit confirmation — declares the tree as the writer. Run it from the repository's project root; commit and push the tree it writes.
- **`planton service kustomize checkout <service>`** — the record as files, authorship unchanged. Use it to inspect configuration, seed a repository before ejecting, or rebuild a tree from what the record currently carries.
- **`planton service kustomize init --env <slug> [--env <slug>...]`** — a fresh skeleton for hand-authoring: empty overlays plus the schema. Touches no record.
- **`planton service kustomize patch-schema --dir <tree>`** — regenerate `planton-schema.json` in an existing tree. Idempotent; run after platform schema upgrades.

## Why ejecting is safe to recommend

Two facts make eject a calm act, not a migration:

1. **The record's entries STAY through the flip.** Ejecting never empties anything — the tree was verified to render byte-identical to the record before the flip, so the first authoritative push converges onto the existing entries and writes nothing. Nothing redeploys because of an eject.
2. **It is reversible.** Removing the `deploy.kustomize` declaration (edit `service.yaml`, or apply the record without it) is the deliberate authorship takeover: the configuration entries stand, sync provenance is stripped, and the caller writes again.

While a service is git-maintained, manual edits to `deploy.environments` are preserved-from-record rather than accepted — the tree is the writer, so the honest way to change configuration is a commit. Relay that rather than helping someone fight the guard. The one door beside the commit is the tree itself: `planton service deploy <service> --env <env> --image <ref> --from-tree` renders the overlay from the folder the person is standing in and deploys it through the control plane, writing that one environment's entry as theirs (who, which commit, uncommitted or not); the next authoritative push with the same content hands the entry back to git. Use it for a freshly ejected or never-pushed service, or to see an overlay run before committing it — see the delivery-verbs reference for what it accepts and refuses.

## The tree layout

```
_kustomize/
  planton-schema.json          # the merge schema (generated; commit it)
  overlays/<env>/              # one per deploy environment — the overlay SET
    kustomization.yaml         #   defines which environments the service deploys to
    <resource>.yaml            # full cloud-resource manifests
  dev/<flavor>/                # local-development flavors: NEVER deployed, never synced
    kustomization.yaml         #   (compose an overlay, patch laptop deltas)
  previews/<env>/              # pull-request preview deltas for that environment: rendered
    kustomization.yaml         #   per PR run, never synced (compose the env's overlay,
                               #   patch what previews change — see preview-environments.md)
```

Ejected trees are written self-contained: each overlay carries complete manifests, no computed base/patches. Factoring shared configuration into a `base/` with strategic-merge patches is the human's own refactor — the platform renders whatever the tree says, so any refactor that renders the same is equally valid. When someone refactors, the schema matters (next section).

## The merge schema, and why losing it hurts

`planton-schema.json` tells kustomize that name-keyed lists (env vars, ports) MERGE across base and patches instead of replacing wholesale. Without it, a patch that sets one env var silently DROPS every other env var in the base's list — a semantics change nobody sees until a deploy is missing half its configuration. The schema is generated, committed, and referenced by every overlay's `kustomization.yaml` (`openapi: path:`). If a tree predates it or a kustomization lost the reference, `patch-schema` repairs both.

## Walking the refusals

- **"already git-maintained"** (eject): the tree already writes this service — `checkout` materializes files, or take authorship back first if the goal is manual control.
- **"declares no per-environment configuration"** (eject/checkout): the record is empty — declare environments first (console, agent, or `service.yaml`), or for a git-maintained service push a driving branch so the sync populates the record.
- **"is git-maintained but no push has synced its tree onto the record yet"** (checkout): the repository itself is the source of truth — clone it instead of checking out.
- **"already exists and is not empty"**: the target directory holds files — generated output never silently mixes into existing work; `--force` overrides deliberately.
