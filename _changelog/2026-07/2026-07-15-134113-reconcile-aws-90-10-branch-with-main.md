# Reconcile the AWS 90/10 Catalog Branch with Main

**Date**: July 15, 2026
**Type**: Refactoring
**Components**: API Definitions, Resource Management, AWS Provider, Infra Charts, Catalog Documentation, Testing Framework

## Summary

Merged `main` into `refactor/aws/bring-kinds-to-90-10-coverage-contd-2`,
reconciling the completed AWS catalog rebuild with everything that landed on
main since the branches diverged — the AwsPlantonRunner appliance, the
import-map/eval surface, the platform-keys annotations convention, the Helm
self-hosting charts, and the `planton explain` engine. Forty-two conflicts
were resolved; the branch now carries the union of both lines and is ready
for the pull request and the release cut.

## Problem Statement / Motivation

The AWS 90/10 branch accumulated 49 commits (the ECR/Route 53 rebuilds
through the 17-chart infra-chart catalog) while main moved 30 commits of its
own. Two lines of development independently claimed the same territory in a
few places, and the branch predated a repo-wide breaking convention change on
main. A straight merge left 42 conflicted paths that needed deliberate calls,
not mechanical resolution.

### Pain Points

- Both lines assigned kind enum slot **354**: main to `AwsPlantonRunner`
  (already shipped in v0.3.8/v0.3.9), the branch to the unreleased
  `AwsRoute53HealthCheck`.
- Main rebuilt `charts/aws/eks-environment` after the branch's charts wave
  had deleted the legacy chart catalog.
- Main moved every platform-behavior key (`planton.dev/provisioner`,
  `pulumi.planton.dev/*`, backend keys) from `metadata.labels` to
  `metadata.annotations` with **no label fallback** — but its sweep could
  not see manifest examples that exist only on the branch, which would have
  silently stopped working.
- Both lines authored an ECR live-E2E surface (verifier, profile, scenario,
  test entrypoints) independently, producing duplicate Go map keys and
  duplicate test function declarations that failed to compile.

## Resolution Calls

- **Kind enum**: `AwsPlantonRunner` keeps 354 — it is released and backs a
  live deployment. `AwsRoute53HealthCheck` (never released) moves to **376**,
  the next free AWS slot. Generated stubs and the catalogkindreflect kind map were
  regenerated, and the full enum was verified duplicate-free (450 kinds).
- **`charts/aws/eks-environment`**: deleted. The clean-slate 17-chart AWS
  catalog is the canonical surface; main's chart targeted the pre-rebuild
  shapes. Stale references in the `chart validate` CLI examples and the
  PR/chart workflow rules were repointed to living charts.
- **Annotations convention completed**: main's labels-to-annotations change
  was applied to every manifest example the branch added after the fork —
  56 blocks across 43 files (catalog pages, presets, hack manifests),
  including a handful of Hetzner/OCI stragglers main's own sweep missed. A
  scripted scan confirms zero platform keys remain under `metadata.labels`
  anywhere in `apis/`, `charts/`, `_rules/`, or `architecture/`.
- **ECR surface**: the branch's rebuilt module and E2E artifacts win — they
  are live-proven against the restructured spec (structured lifecycle rules,
  exclusion-filtered tag mutability, folded repository policy). Main's
  lifecycle-policy tuple fix targeted the retired knob-based composition and
  is moot against the rebuilt module, whose per-rule `for` expression was
  proven live with heterogeneous rule shapes. Main's duplicate ECR verifier,
  its duplicate `awsecrrepo` registry entry, and its duplicate
  `TestAwsEcrRepo_*` entrypoints were removed; main's `import-map.yaml` and
  import catalog entries were kept (they reference `spec.repository_name`,
  which the rebuilt spec retains).
- **Generated artifacts regenerated, never hand-merged**: proto stubs
  (`make protos`), the catalogkindreflect kind map, `go.sum` (`go mod tidy`), and
  the site catalog mirror (copy-docs + structure + stats).

## Verification

All offline gates ran green on the merged tree:

- `aa_e2e` verify package and `-tags=e2e` harness compile cleanly.
- Spec tests: `awsroute53healthcheck`, `awsecrrepo`, `awsplantonrunner`,
  `awscognitouserpool`; framework tests: `pkg/catalogkindreflect`, `pkg/outputs`,
  `pkg/refcheck`.
- Working-tree CLI guards: `validate-refs --check` (all foreign keys
  resolve) and `secret-coverage --check` (gate passed).
- `planton chart validate --all charts/aws`: **17/17 charts pass** across
  every bool-toggle variant. (Other providers' chart failures under
  `charts/` are their in-flight rebuilds on main, untouched here.)
- `tofu init && tofu validate` on the conflicted ECR Terraform module.
- `make build-go` exits clean.

## Impact

The branch is the single reconciled line carrying the complete AWS 90/10
catalog plus everything main shipped for the self-hosting preview. The
release boundary (pull request, release cut, platform upgrade) can proceed
from commit `068acd30e` with no known divergence left between the lines.

One coordination note for the import-mapping work: the eval suites and
import recipes on main were authored against pre-rebuild specs for the kinds
this branch later restructured (ECR was verified compatible; others should
be re-checked against the rebuilt shapes when that work resumes).

---

**Status**: ✅ Production Ready
