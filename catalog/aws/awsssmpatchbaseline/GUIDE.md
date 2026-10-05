# AwsSsmPatchBaseline — Kind Guide

Authored operational judgment for the patch baseline kind: the
design decisions behind the spec's shape, and what to know before
operating baselines in production.

## Design decisions

- **Patch groups fold into the baseline.** A group registration is
  `{group name, baseline id}` and nothing else — a pure edge that
  cannot outlive the baseline. Each spec entry is its own registration
  (the for_each key and half of the `{patch_group},{baseline_id}`
  import composite).
- **The default designation folds as `setAsDefaultBaseline`** — and
  unlike the App Runner account-default claim, this one REVERTS:
  destroying it (or the whole component) restores AWS's own predefined
  default baseline for the OS, which the provider looks up and
  re-registers. The standalone provider resource can fail on an
  OS-mismatched baseline; the fold makes that unrepresentable (the
  designation always names this baseline's own resolved OS).
- **The approval arms are mutually exclusive by CEL** —
  `approveAfterDays` XOR `approveUntilDate` — a rule AWS enforces
  server-side but the provider never pre-checks. A rule may carry
  NEITHER arm: on Debian and Ubuntu, days-based auto-approval is not
  supported by AWS, and filter-only rules are the honest shape there.
- **`approveAfterDays` is presence-typed** because 0 is meaningful
  (approve immediately on release) and distinct from unset.

## Operating baselines in production

- **One group, one baseline per OS.** A patch group can be registered
  with only ONE baseline per operating system account-wide — AWS
  state, not validation, referees, and the second registration fails
  at apply. Coordinate group ownership across components.
- **The designation displaces silently**: claiming the default marks
  the previous holder non-default. At most one baseline per OS should
  set `setAsDefaultBaseline` — components cannot see each other, so
  this is an operating rule, not a validation.
- **Deleting a baseline that holds the designation works**: the
  provider restores AWS's predefined default for the OS and retries
  the delete.
- **Approval-rule soak periods are the risk lever**: `approveAfterDays:
  7` means a bad vendor patch has a week to be pulled before your
  fleet takes it; `0` takes patches on release day. `approveUntilDate`
  freezes approval at a cutoff — advance it deliberately per change
  window.
- **`rejectedPatchesAction: BLOCK` reports noncompliance** when a
  rejected patch is an approved patch's dependency;
  `ALLOW_AS_DEPENDENCY` (the default) installs it quietly.
- **Alternative sources are Linux-only** and stored verbatim — AWS
  does not validate reachability; a wrong repo shows up as install
  failures at patch time, not at deploy.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
