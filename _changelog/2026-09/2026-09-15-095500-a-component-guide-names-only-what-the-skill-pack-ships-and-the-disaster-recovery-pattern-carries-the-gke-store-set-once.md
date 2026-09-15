# A component guide names only what the skill pack ships, and the disaster-recovery pattern carries the GKE store set once

**Date**: September 15, 2026
**Type**: Documentation
**Components**: The stateful-kind disaster-recovery pattern; the KubernetesOpenBao, KubernetesPostgres, and KubernetesMongodb guides; the guide and pattern authoring and audit rules; the multi-cloud-catalog skill's references and its eval bank

## Summary

An agent reading the packaged multi-cloud-catalog skill was sent, by three guides, to files the pack does not carry: the E2E fixtures and scenarios and the real-cluster proof batch that each guide named as "the validated manifests for this set". The pack ships a kind's reference page, guide, and fact sheets and the patterns library — never `e2e/`, `presets/`, or `iac/` — so every one of those pointers landed nowhere, and the skill forbids filling the gap from memory. The pointers are gone. The three GCP nodes every keyless GKE backup stands on (the identity, the bucket with its two roles, the Workload Identity binding) are embedded once, validated, in the disaster-recovery pattern; the OpenBao guide embeds the pieces that are the vault's alone — the Cloud KMS seal set with both grants, and the complete bad-day restore target — and the PostgreSQL and MongoDB guides point at the pattern and name their presets by slug. The authoring rules now state the law, and the skill's eval bank pins the OpenBao disaster-recovery facts so a future edit that drops one fails CI.

## What Changed

### The pattern carries the shared set

`catalog/_patterns/stateful-kind-disaster-recovery.md` gains "The GKE keyless store set, complete": a `GcpServiceAccount` with no key, a `GcpGcsBucket` granting it `roles/storage.objectAdmin` and `roles/storage.legacyBucketReader` by reference, and a `GcpGkeWorkloadIdentityBinding` — with the three rules that ride with it (two roles, not one; the binding names the ServiceAccount the kind renders, and each kind renders a different one; MongoDB takes the identity keyed, not bound). The frontmatter declares the three GCP kinds. The R2 presets of the database and the replica set are named by slug instead of by path.

### The OpenBao guide embeds what is the vault's alone

The GKE resource-set section embeds the seal set — the server's identity, its binding on the release-named ServiceAccount, the permanent key ring, the key with `deletionPolicy: PREVENT`, and the two `GcpKmsKeyIamMember` grants (`cryptoKeyEncrypterDecrypter` and `viewer`, with the crash-loop the missing one produces) — and points at the pattern for the store set and at the preset by slug for the vault. "Restore on the bad day" now holds the complete restore target instead of a five-line fragment, and teaches two things a fresh reader asked for: keep the source's name and namespace (the bindings, the login role, and the prefix all carry over; a different name is the rehearsal shape and needs its own binding pair and a re-run of the login recipe), and declare `createNamespace` because the lost cluster's namespace is gone too. The init command's `1`/`1` recovery shares are named as the proof lane's value, with the production shape and what recovery keys are for under auto-unseal. The R2 section points at the pattern's embedded pair and vault.

### The sibling guides stop pointing at the lab

`kubernetespostgres/GUIDE.md` and `kubernetesmongodb/GUIDE.md` replace their four "validated manifests are the lane's own" paragraphs with the pattern's set (PostgreSQL's binding named after the Cluster; MongoDB's identity keyed) and their presets by slug; each restore row's table entry is named as the whole delta from its preset.

### The rules state the law

`write-planton-component-guide.mdc` and `write-planton-architecture-pattern.mdc`: name only what the packaged skill carries — a preset by slug with the sentence that presets ship outside the pack, never a path under `e2e/`, `iac/`, or `presets/`; a manifest the reader must copy is embedded or lives in the pattern that owns the shape. `audit-planton-component-guide.mdc` and `audit-planton-architecture-pattern.mdc` gain the pack-reachability check with the grep that proves it.

### The skill knows, and the bank remembers

`pack-layout.md` says a guide's path resolves inside the pack; the research recipe's pre-restore checklist starts from the guide's embedded restore target. `pkg/explain/refgen/eval_questions.yaml` gains three wisdom questions — the GKE set and its grants, the bad-day restore and what never to do afterwards, the constraints and proof status an agent must state — each pinned by checks on the reference page, the guide, and the pattern.

## Verification

`go test ./pkg/explain/refgen/` (placement, every embedded manifest validates, pattern kinds resolve, links resolve, reference drift, the eval bank fresh); every one of the thirteen embedded manifests validated with the tree-built CLI; `go run ./pkg/skills/defspack -embed-catalog-pack` builds the pack with zero `e2e/`, `aa_e2e`, or `presets/<file>` pointers in the four documents; `go test ./pkg/skills/defspack/`. A fresh agent given only the built pack composed the eleven-manifest GKE set and the restore target and named forty-eight hazards, every one cited to a pack line.
