# Azure Branch Reconciled with main: Framework Unification After the AWS 90/10 Merges

**Date**: July 15, 2026
**Type**: Refactor / Framework
**Components**: E2E Framework, InfraChart Validation, Outputs Population, Chart Authoring Rules, Azure Provider, Documentation Conventions

## Summary

The Azure 90/10 branch is reconciled with `main` after the AWS 90/10 project's
two squash merges (#460, #465). Beyond the merge itself, the reconciliation
unified three framework capabilities that the two provider projects had built
independently on diverged branches — the scenario-declared E2E fixture
mechanism, the offline InfraChart validation gate, and output map
population — so each now exists exactly once, carrying both projects'
capability sets. The Azure documentation surface also adopts main's
platform-key convention (`planton.dev/provisioner` and backend keys in
`metadata.annotations`, never `labels`).

## Problem Statement / Motivation

The Azure branch was deliberately cut from the AWS project's branch tip so it
would inherit in-flight framework progress, with the reconciliation deferred
to merge time. Both AWS PRs landed as squash merges, so git saw the same AWS
content on both sides as unrelated edits — a large but mechanical conflict
surface — and three genuine duplications had accumulated: each project had
solved the same framework problems separately, in half-overlapping ways.

### Pain Points

- Two scenario-fixture annotations (`planton.dev/e2e-prerequisites` on main,
  `planton.dev/e2e-extra-prerequisites` on the Azure branch) doing 70% of the
  same job with disjoint capabilities
- Two offline chart gates: a Python render+validate harness (venv, Jinja2
  compatibility shims, schema drift risk) and the Go-native
  `planton chart validate` with compiled-in schemas
- Two chart authoring rules teaching overlapping contracts
- Two `pkg/outputs` map-population implementations
- Azure catalog pages and site docs teaching the retired labels-based
  provisioner convention, which the CLI no longer reads

## Solution / What's New

### One E2E fixture mechanism

`planton.dev/e2e-prerequisites` is the single annotation. The resolver keeps
main's implementation (prerequisite-graph expansion with cycle detection,
install-manifest transitive edges) and folds in the Azure branch's
manifest-path entries — a scenario can declare an EXTRA INSTANCE of a kind by
repo-relative manifest path (a virtual-network peering's second network), each
preceded by its own not-yet-scheduled prerequisites and never substituting for
the kind's install profile. Unknown kind names and unregistered path-manifest
kinds fail loudly at resolution time. Both projects' test suites merged; all
66+ Azure scenario manifests renamed; zero references to the retired
annotation remain outside historical changelogs.

### One chart gate, strengthened

The Python harness (`hack/validate_charts_offline.py`) and its venv are
deleted; `planton chart validate` (pkg/infrachart) is the offline gate
everywhere, and it gains the one check only the Python harness had:
**intra-chart target resolution** — every `valueFrom` must reference a
kind+name the same render variant defines, so a toggle that removes a resource
but leaves references standing fails that variant. The ref walker was
refactored to a visitor so per-document validation and the variant-wide check
share one traversal, proven by a fourth defect class in the broken-chart
fixture.

Two renderer defects fixed along the way: gonja v2.8.0's
`strip`/`lstrip`/`rstrip` builtins empty any string without trailing cut
characters (`"a.com".strip()` == `""`) — replaced with `strings.Trim`-based
implementations plus regression tests; and the three Azure chart icon URLs
pointed at a retired asset path (the validator cannot see Chart.yaml — icon
liveness is a curl check in the authoring rule).

### One authoring rule

`_rules/charts/forge-planton-infra-chart.mdc` absorbs the Azure-side rule's
substance (one-chart-one-architecture, secure-by-default posture, chart
self-containment, the seam-gap-is-a-kind-gap principle and its honest-
literal boundary, parameter-earns-its-place, the naming convention);
`author-planton-infra-charts.mdc` is deleted and the repair rule repointed.

### One outputs implementation

`pkg/outputs` map population keeps the union: name-keyed map outputs
(load-balancer pools/probes), dot-flattened keys with per-segment hyphen
normalization so data-bearing keys survive verbatim, empty-map handling — with
both projects' conformance suites green against the single implementation.

### Documentation conventions

209 Azure files moved their platform-key example blocks from
`metadata.labels` to `metadata.annotations`, matching the convention main
established (labels derive into cloud tags; platform keys there would leak
onto users' real resources).

## Impact

- **Cross-project**: every future provider rebuild inherits ONE fixture
  mechanism, ONE chart gate, ONE authoring contract — the duplication cost of
  parallel provider rebuilds was paid once, here.
- **Merge topology**: the Azure branch now contains current main; the eventual
  Azure PR will show only Azure work. A final `merge origin/main` runs before
  the release cut.
- **No user-facing behavior change**: modules, specs, and outputs are
  unchanged except where the union strengthened validation.

## Validation

- Full offline gate on the merged tree: `make build-go`, spec tests for the
  retrofitted kinds, `validate-refs`, `secret-coverage`, `pkg/outputs`
  conformance (both suites), `e2e/framework/runner` + `pkg/infrachart` tests,
  `planton chart validate` green on all five Azure charts (defaults + every
  toggle variant), chart structure guard, icon URLs live.
- Live dual-engine E2E re-runs for the retrofitted kinds (session-037
  scenarios exercising the renamed annotation's kind-name AND manifest-path
  forms through composed fixture chains).
