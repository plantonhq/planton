# E2E framework: a pending_proof profile status separates authoring from proving, and repo-global locks coordinate concurrent agent sessions

**Date**: 2026-07-24
**Scope**: `apis/dev/planton/qa/catalogkinde2eprofile/v1/spec.proto` (+ regenerated Go stub), `pkg/e2e/profile/discover.go`, `cmd/planton/root/e2e/discover.go`, `internal/cli/ui/e2ediscover/` (table + interactive), `e2e/kubernetes_test.go` (profile skip switch), `_locks/` (new), forge/update workflow rules. No kind behavior change.

## What changed

1. **`pending_proof` status in `CatalogKindE2EProfileSpec.Status`.** The
   lifecycle previously had no honest state for a kind that is fully
   authored and offline-validated but whose live lanes have not yet run:
   `deferred` records a KNOWN failure or blocker, and `green` would put an
   unproven kind into CI matrices (which are built from green profiles
   only). `pending_proof` names that state precisely. The provider test
   entrypoints skip it exactly like `deferred` — a proving session flips the
   profile to green immediately before executing the lanes (the existing
   flip mechanic) — and `planton e2e discover --status pending_proof`
   enumerates everything awaiting its first proof straight from the tree.
   While a profile carries this status, `validated_provisioners` stays empty.
2. **Discover CLI/UI carry the full status vocabulary.** The `--status`
   filter now accepts `real_cluster` (previously missing from the filter
   switch despite existing in the enum) and `pending_proof`; the table
   summary line and the interactive view render the new status.
3. **`_locks/` at the repo root** — cooperative lock files coordinating
   multiple coding agents on the same branch: `git-commit.lock.md` (a
   session's single wrap-up commit + push) and `proto-build.lock.md`
   (whole-tree generation — protos/stubs, kind map, gazelle, proto-docs,
   e2e matrix, site regen — and brief edits to shared choke-point files).
   Committed README + `.gitignore`; the lock files themselves are transient
   and ignored. Per-kind work never needs a lock.
4. **Forge/update rules teach the split execution mode.** The forge rule's
   profile-honesty section now names the two phases — authoring (through the
   offline gates, profile at `pending_proof`) and proving (live lanes +
   round-trips, the profile flip, `validated_provisioners`, live-caught
   lessons landed on reader surfaces, CI matrix regen) — and states they may
   be executed by different sessions, with the proving session owning the
   kind completely while proving. The update rule carries the same contract
   for updates that change deployable behavior.

## Validation

- `buf lint` + `buf format --diff` clean on the edited proto dir; chunked
  stub regeneration verified (`pending_proof` present in the generated Go).
- `go build` green on all touched packages; `go test ./pkg/e2e/profile/...`
  green; `go vet -tags=e2e ./e2e/` green.
- No profile currently carries the new status, so the generated CI matrix is
  byte-identical — verified by the green-only filter in the matrix builder.
